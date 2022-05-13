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
	"strings"

	"github.com/coneno/logger"

	"github.com/crewjam/saml/samlsp"
	"github.com/gin-gonic/gin"

	mw "github.com/tekenradar/researcher-backend/pkg/http/middlewares"
)

func (h *HttpEndpoints) AddAuthAPI(rg *gin.RouterGroup) {
	samlSP, err := h.InitSamlSP()
	if err != nil {
		logger.Error.Panic(err)
	}
	rg.POST("/saml/acs", gin.WrapF(samlSP.ServeACS))

	auth := rg.Group("/auth")

	app := http.HandlerFunc(h.loginWithSAML)
	auth.GET("/login", mw.RequireQueryParams([]string{"role", "instance"}), gin.WrapH(samlSP.RequireAccount(app)))
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

type GroupInfo struct {
	Customer   string
	Prefix     string
	InstanceID string
	Role       string
}

func parseSAMLgroupInfo(groups []string) []GroupInfo {
	sep := "-"
	infos := []GroupInfo{}
	for _, groupText := range groups {
		parts := strings.Split(groupText, sep)
		if len(parts) != 4 {
			logger.Error.Printf("'%s' has %d parts when using '%s' as a separator but 4 are expected.", groupText, len(parts), sep)
			continue
		}

		c := GroupInfo{
			Customer:   parts[0],
			Prefix:     parts[1],
			InstanceID: parts[2],
			Role:       parts[3],
		}
		infos = append(infos, c)
	}
	return infos
}

func checkPermission(samlGroupInfos []GroupInfo, instanceID string, role string) (bool, *GroupInfo) {
	for _, g := range samlGroupInfos {
		if g.InstanceID == instanceID && role == g.Role {
			return true, &g
		}
	}
	return false, nil
}

type SAMLLoginInfo struct {
	Username   string
	Tokens     string
	InstanceID string
	Role       string
}

func (h *HttpEndpoints) loginWithSAML(w http.ResponseWriter, r *http.Request) {
	instanceIDs, ok := r.URL.Query()["instance"]
	if !ok || len(instanceIDs[0]) < 1 {
		http.Error(w, "Url Param 'instance' is missing", http.StatusBadRequest)
		return
	}
	roles, ok := r.URL.Query()["role"]
	if !ok || len(roles[0]) < 1 {
		http.Error(w, "Url Param 'role' is missing", http.StatusBadRequest)
		return
	}

	instanceID := instanceIDs[0]
	role := roles[0]

	s := samlsp.SessionFromContext(r.Context())
	if s == nil {
		logger.Error.Println("session not found")
		return
	}

	jwtSessionClaims, ok := s.(samlsp.JWTSessionClaims)
	if !ok {
		logger.Error.Println("Unable to decode session into JWTSessionClaims")
		return
	}

	email := jwtSessionClaims.Subject

	sa, ok := s.(samlsp.SessionWithAttributes)
	if !ok {
		logger.Error.Println("attributes not found")
		return
	}

	attributes := sa.GetAttributes()
	groups, ok := attributes["http://schemas.xmlsoap.org/claims/Group"]
	if !ok {
		err := fmt.Errorf("group infos not found in the response token for %s", email)
		logger.Error.Println(err.Error())
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}
	groupInfos := parseSAMLgroupInfo(groups)

	hasPermission, usedGroupInfo := checkPermission(groupInfos, instanceID, role)
	if !hasPermission {
		err := fmt.Errorf("'%s' is not authorized to access '%s' with role '%s'.", email, instanceID, role)
		logger.Error.Println(err.Error())
		logger.Debug.Printf("valid group infos are %v", groupInfos)
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	logger.Debug.Print(groupInfos)
	logger.Debug.Print(usedGroupInfo)

	/*
		req := umAPI.LoginWithExternalIDPMsg{
			InstanceId: instanceID,
			Email:      email,
			Role:       strings.ToUpper(role),
			Customer:   usedGroupInfo.Customer,
			GroupInfo:  strings.Join(groups, ";"),
			Idp:        h.samlConfig.IDPUrl,
		}

		resp, err := h.clients.UserManagement.LoginWithExternalIDP(context.Background(), &req)
		if err != nil {
			logger.Error.Println(err.Error())
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}

		loginInfos := SAMLLoginInfo{
			Username:   email,
			InstanceID: instanceID,
			Role:       role,
			Tokens: strings.Join([]string{
				resp.Token.AccessToken,
				resp.Token.RefreshToken,
			}, "<!>"),
		}

		parsedTemplate, _ := template.ParseFiles(h.samlConfig.TemplatePathLoginSuccess)
		err = parsedTemplate.Execute(w, loginInfos)
		if err != nil {
			logger.Error.Println(err.Error())
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}
		if err != nil {
			logger.Error.Println("Error executing template :", err)
			return
		}*/

	// c.Data(http.StatusOK, "text/html; charset=utf-8", tpl.Bytes())
	//fmt.Fprintf(w, "Logged in as: %s, Token contents, %+v!\n\n%v \n\n %s - %s \n\n%s", email, sa.GetAttributes(), groupInfos, instanceID, role, resp.Token.AccessToken)
}
