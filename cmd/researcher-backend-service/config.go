package main

import (
	"os"
	"strings"

	"github.com/coneno/logger"

	v1 "github.com/tekenradar/researcher-backend/pkg/http/v1"
)

const (
	ENV_LOG_LEVEL      = "LOG_LEVEL"
	ENV_GIN_DEBUG_MODE = "GIN_DEBUG_MODE"

	ENV_RESEARCHER_BACKEND_LISTEN_PORT = "RESEARCHER_BACKEND_LISTEN_PORT"
	ENV_CORS_ALLOW_ORIGINS             = "CORS_ALLOW_ORIGINS"

	ENV_SAML_IDP_URL                     = "SAML_IDP_URL"
	ENV_SAML_SERVICE_PROVIDER_ROOT_URL   = "SAML_SERVICE_PROVIDER_ROOT_URL"
	ENV_SAML_ENTITY_ID                   = "SAML_ENTITY_ID"
	ENV_SAML_IDP_METADATA_URL            = "SAML_IDP_METADATA_URL"
	ENV_SAML_SESSION_CERT_PATH           = "SAML_SESSION_CERT_PATH"
	ENV_SAML_SESSION_KEY_PATH            = "SAML_SESSION_KEY_PATH"
	ENV_SAML_TEMPLATE_PATH_LOGIN_SUCCESS = "SAML_TEMPLATE_PATH_LOGIN_SUCCESS"
)

// Config is the structure that holds all global configuration data
type Config struct {
	Port         string
	AllowOrigins []string
	LogLevel     logger.LogLevel
	GinDebugMode bool
	SAMLConfig   *v1.SAMLConfig `yaml:"saml_config"`
}

func InitConfig() Config {
	conf := Config{}
	conf.Port = os.Getenv(ENV_RESEARCHER_BACKEND_LISTEN_PORT)
	conf.AllowOrigins = strings.Split(os.Getenv(ENV_CORS_ALLOW_ORIGINS), ",")

	conf.LogLevel = getLogLevel()
	conf.GinDebugMode = os.Getenv(ENV_GIN_DEBUG_MODE) == "true"

	conf.SAMLConfig = &v1.SAMLConfig{
		IDPUrl:                   os.Getenv(ENV_SAML_IDP_URL),                   // arbitrary name to refer to IDP in the logs
		SPRootUrl:                os.Getenv(ENV_SAML_SERVICE_PROVIDER_ROOT_URL), // url of the management api
		EntityID:                 os.Getenv(ENV_SAML_ENTITY_ID),
		MetaDataURL:              os.Getenv(ENV_SAML_IDP_METADATA_URL),
		SessionCertPath:          os.Getenv(ENV_SAML_SESSION_CERT_PATH),
		SessionKeyPath:           os.Getenv(ENV_SAML_SESSION_KEY_PATH),
		TemplatePathLoginSuccess: os.Getenv(ENV_SAML_TEMPLATE_PATH_LOGIN_SUCCESS),
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
