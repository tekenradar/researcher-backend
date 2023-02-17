package v1

import (
	"context"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/coneno/logger"
	"github.com/tekenradar/researcher-backend/internal/config"
	mw "github.com/tekenradar/researcher-backend/pkg/http/middlewares"
	"github.com/tekenradar/researcher-backend/pkg/http/utils"
	"github.com/tekenradar/researcher-backend/pkg/jwt"

	"github.com/crewjam/saml/samlsp"
	"github.com/gin-gonic/gin"
)

const adminRoleFromADFS = "researchadmin"

func (h *HttpEndpoints) AddAuthAPI(rg *gin.RouterGroup) {
	auth := rg.Group("/auth")

	if h.useDummyLogin {
		auth.GET("/login", h.dummyLogin)
	} else {
		samlSP, err := h.InitSamlSP()
		if err != nil {
			logger.Error.Panic(err)
		}
		rg.POST("/saml/acs", gin.WrapF(samlSP.ServeACS))

		app := http.HandlerFunc(h.loginWithSAML)
		auth.GET("/login", gin.WrapH(samlSP.RequireAccount(app)))
	}

	auth.GET("/init-session", mw.HasValidAPIKey(h.apiKeys), h.initSession)
	auth.GET("/logout", h.logout)
}

func (h *HttpEndpoints) dummyLogin(c *gin.Context) {
	token, err := jwt.GenerateNewToken("testaccount@rivm.nl", utils.InitSessionTokenAge*time.Second,
		[]string{
			jwt.ROLE_ADMIN,
		})

	if err != nil {
		logger.Error.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	url := fmt.Sprintf("%s?id=%s", h.loginSuccessRedirectURL, token)
	c.Redirect(http.StatusFound, url)
}

func (h *HttpEndpoints) initSession(c *gin.Context) {
	token := c.DefaultQuery("token", "")

	logger.Debug.Printf("received token %s", token)

	if token == "" {
		var err error
		token, err = c.Cookie(utils.AuthCookieName)
		if err != nil {
			logger.Error.Println(err)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "no Authorization token found"})
			return
		}
	}

	claims, valid, err := jwt.ValidateToken(token)
	if err != nil || !valid {
		logger.Error.Printf("invalid token with err: %v", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
		return
	}

	accessToken, err := jwt.GenerateNewToken(claims.ID, utils.TokenMaxAge*time.Second, claims.Roles)
	if err != nil {
		logger.Error.Printf("unexpected error when generating token: %v", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unexpected error when generating token"})
		return
	}

	c.SetSameSite(http.SameSiteLaxMode)
	secure := true
	c.SetCookie(
		utils.AuthCookieName,
		accessToken,
		utils.TokenMaxAge,
		"/",
		"",
		secure,
		true,
	)

	c.JSON(http.StatusOK, gin.H{"userID": claims.ID})
}

func (h *HttpEndpoints) logout(c *gin.Context) {
	c.SetCookie(
		utils.AuthCookieName,
		"",
		-1,
		"/",
		"",
		true,
		true,
	)
	c.JSON(http.StatusOK, gin.H{"msg": "logout successful"})
}

func (h HttpEndpoints) InitSamlSP() (*samlsp.Middleware, error) {
	if h.samlConfig == nil {
		return nil, errors.New("SAML config not available")
	}

	keyPair, err := tls.LoadX509KeyPair(h.samlConfig.SessionCertPath, h.samlConfig.SessionKeyPath)
	if err != nil {
		return nil, fmt.Errorf("Problem loading session certificate or key: %v", err)
	}

	keyPair.Leaf, err = x509.ParseCertificate(keyPair.Certificate[0])
	if err != nil {
		return nil, fmt.Errorf("Problem parsing session certificate or key: %v", err)
	}

	idpMetadataURL, err := url.Parse(h.samlConfig.MetaDataURL)
	if err != nil {
		return nil, fmt.Errorf("Can't parse metadata url: %v", err)
	}

	rootURL, err := url.Parse(h.samlConfig.SPRootUrl)
	if err != nil {
		return nil, fmt.Errorf("Can't parse service provider root URL: %v", err)
	}

	metaData, err := samlsp.FetchMetadata(context.TODO(), http.DefaultClient, *idpMetadataURL)
	if err != nil {
		return nil, fmt.Errorf("Error when fetching metadata: %v", err)
	}

	samlSP, err := samlsp.New(samlsp.Options{
		URL:         *rootURL,
		Key:         keyPair.PrivateKey.(*rsa.PrivateKey),
		Certificate: keyPair.Leaf,
		IDPMetadata: metaData,
		EntityID:    h.samlConfig.EntityID,
	})
	if err != nil {
		return nil, err
	}

	acsURL, err := url.Parse(h.samlConfig.SPRootUrl + "/v1/saml/acs")
	if err != nil {
		return nil, fmt.Errorf("Can't parse ACS URL: %v", err)
	}
	samlSP.ServiceProvider.AcsURL = *acsURL

	return samlSP, nil
}

type SAMLLoginInfo struct {
	Username   string
	Tokens     string
	InstanceID string
	Role       string
}

func (h *HttpEndpoints) loginWithSAML(w http.ResponseWriter, r *http.Request) {
	urlForFailedLogin := os.Getenv(config.ENV_SAML_LOGIN_FAILED_REDIRECT_URL)
	s := samlsp.SessionFromContext(r.Context())
	if s == nil {
		logger.Error.Println("session not found")
		http.Redirect(w, r, urlForFailedLogin, http.StatusForbidden)
		return
	}

	jwtSessionClaims, ok := s.(samlsp.JWTSessionClaims)
	if !ok {
		logger.Error.Println("Unable to decode session into JWTSessionClaims")
		http.Redirect(w, r, urlForFailedLogin, http.StatusForbidden)
		return
	}

	email := jwtSessionClaims.Subject

	sa, ok := s.(samlsp.SessionWithAttributes)
	if !ok {
		logger.Error.Println("attributes not found")
		http.Redirect(w, r, urlForFailedLogin, http.StatusForbidden)
		return
	}

	attributes := sa.GetAttributes()

	claimedRoles, ok := attributes[os.Getenv(config.ENV_SAML_ATTRIBUTE_FOR_TEKENRADAR_ACCESS)]
	if !ok {
		err := fmt.Errorf("role infos not found in the response token for %s", email)
		logger.Error.Println(err.Error())
		http.Redirect(w, r, urlForFailedLogin, http.StatusForbidden)
		return
	}

	roles := []string{}
	logger.Debug.Println(email)
	logger.Debug.Println(attributes)

	hasAccessToTekenradar := false
	for _, r := range claimedRoles {
		// convert string to lowercase
		r = strings.ToLower(r)
		if strings.Contains(r, "tekenradar") {
			hasAccessToTekenradar = true
			break
		}
	}
	if !hasAccessToTekenradar {
		err := fmt.Errorf("user %s has no access to tekenradar", email)
		logger.Error.Println(err.Error())
		http.Redirect(w, r, urlForFailedLogin, http.StatusForbidden)
		return
	}

	// check if user is a research admin
	researchAdmins := os.Getenv(config.ENV_RESEARCHADMIN_EMAILS)
	if strings.Contains(researchAdmins, email) {
		roles = []string{
			jwt.ROLE_ADMIN,
		}
	}

	// prepare token
	token, err := jwt.GenerateNewToken(
		email,
		utils.InitSessionTokenAge*time.Second,
		roles,
	)
	if err != nil {
		logger.Error.Println(err)
		http.Redirect(w, r, urlForFailedLogin, http.StatusForbidden)
		return
	}

	url := fmt.Sprintf("%s?id=%s", h.loginSuccessRedirectURL, token)
	http.Redirect(w, r, url, http.StatusFound)
}
