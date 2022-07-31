package config

import (
	"os"
	"strings"

	"github.com/coneno/logger"
	"github.com/tekenradar/researcher-backend/pkg/types"
)

const (
	ENV_LOG_LEVEL      = "LOG_LEVEL"
	ENV_GIN_DEBUG_MODE = "GIN_DEBUG_MODE"

	ENV_RESEARCHER_BACKEND_LISTEN_PORT = "RESEARCHER_BACKEND_LISTEN_PORT"
	ENV_CORS_ALLOW_ORIGINS             = "CORS_ALLOW_ORIGINS"
	ENV_USE_DUMMY_LOGIN                = "USE_DUMMY_LOGIN"            // if true, test mode for auth is used
	ENV_LOGIN_SUCCESS_REDIRECT_URL     = "LOGIN_SUCCESS_REDIRECT_URL" // address of the web-application
	ENV_API_KEYS                       = "API_KEYS"

	ENV_SAML_IDP_URL                   = "SAML_IDP_URL"
	ENV_SAML_SERVICE_PROVIDER_ROOT_URL = "SAML_SERVICE_PROVIDER_ROOT_URL"
	ENV_SAML_ENTITY_ID                 = "SAML_ENTITY_ID"
	ENV_SAML_IDP_METADATA_URL          = "SAML_IDP_METADATA_URL"
	ENV_SAML_SESSION_CERT_PATH         = "SAML_SESSION_CERT_PATH"
	ENV_SAML_SESSION_KEY_PATH          = "SAML_SESSION_KEY_PATH"

	ENV_JWT_TOKEN_KEY = "JWT_TOKEN_KEY"
)

// Config is the structure that holds all global configuration data
type Config struct {
	Port                    string
	AllowOrigins            []string
	APIKeys                 []string
	LogLevel                logger.LogLevel
	GinDebugMode            bool
	SAMLConfig              *types.SAMLConfig `yaml:"saml_config"`
	UseDummyLogin           bool
	LoginSuccessRedirectURL string
}

func InitConfig() Config {
	conf := Config{}
	conf.Port = os.Getenv(ENV_RESEARCHER_BACKEND_LISTEN_PORT)
	conf.AllowOrigins = strings.Split(os.Getenv(ENV_CORS_ALLOW_ORIGINS), ",")

	conf.APIKeys = strings.Split(os.Getenv(ENV_API_KEYS), ",")
	conf.LogLevel = getLogLevel()
	conf.GinDebugMode = os.Getenv(ENV_GIN_DEBUG_MODE) == "true"
	conf.UseDummyLogin = os.Getenv(ENV_USE_DUMMY_LOGIN) == "true"
	conf.LoginSuccessRedirectURL = os.Getenv(ENV_LOGIN_SUCCESS_REDIRECT_URL)

	conf.SAMLConfig = &types.SAMLConfig{
		IDPUrl:          os.Getenv(ENV_SAML_IDP_URL),                   // arbitrary name to refer to IDP in the logs
		SPRootUrl:       os.Getenv(ENV_SAML_SERVICE_PROVIDER_ROOT_URL), // url of the management api
		EntityID:        os.Getenv(ENV_SAML_ENTITY_ID),
		MetaDataURL:     os.Getenv(ENV_SAML_IDP_METADATA_URL),
		SessionCertPath: os.Getenv(ENV_SAML_SESSION_CERT_PATH),
		SessionKeyPath:  os.Getenv(ENV_SAML_SESSION_KEY_PATH),
	}

	if len(conf.SAMLConfig.IDPUrl) > 0 {
		conf.AllowOrigins = append(conf.AllowOrigins, conf.SAMLConfig.IDPUrl)
	}

	return conf
}

func getLogLevel() logger.LogLevel {
	switch os.Getenv(ENV_LOG_LEVEL) {
	case "debug":
		return logger.LEVEL_DEBUG
	case "info":
		return logger.LEVEL_INFO
	case "error":
		return logger.LEVEL_ERROR
	case "warning":
		return logger.LEVEL_WARNING
	default:
		return logger.LEVEL_INFO
	}
}