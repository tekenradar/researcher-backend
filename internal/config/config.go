package config

import (
	"fmt"
	"os"
	"strconv"
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

	ENV_API_KEYS = "API_KEYS"

	ENV_SAML_IDP_URL                   = "SAML_IDP_URL"
	ENV_SAML_SERVICE_PROVIDER_ROOT_URL = "SAML_SERVICE_PROVIDER_ROOT_URL"
	ENV_SAML_ENTITY_ID                 = "SAML_ENTITY_ID"
	ENV_SAML_IDP_METADATA_URL          = "SAML_IDP_METADATA_URL"
	ENV_SAML_SESSION_CERT_PATH         = "SAML_SESSION_CERT_PATH"
	ENV_SAML_SESSION_KEY_PATH          = "SAML_SESSION_KEY_PATH"

	ENV_SAML_LOGIN_FAILED_REDIRECT_URL       = "SAML_LOGIN_FAILED_REDIRECT_URL"
	ENV_SAML_ATTRIBUTE_FOR_TEKENRADAR_ACCESS = "SAML_ATTRIBUTE_FOR_TEKENRADAR_ACCESS"
	ENV_RESEARCHADMIN_EMAILS                 = "RESEARCHADMIN_EMAILS"

	ENV_JWT_TOKEN_KEY = "JWT_TOKEN_KEY"

	ENV_RESEARCHER_DB_CONNECTION_STR    = "RESEARCHER_DB_CONNECTION_STR"
	ENV_RESEARCHER_DB_USERNAME          = "RESEARCHER_DB_USERNAME"
	ENV_RESEARCHER_DB_PASSWORD          = "RESEARCHER_DB_PASSWORD"
	ENV_RESEARCHER_DB_CONNECTION_PREFIX = "RESEARCHER_DB_CONNECTION_PREFIX"

	ENV_DB_TIMEOUT           = "DB_TIMEOUT"
	ENV_DB_IDLE_CONN_TIMEOUT = "DB_IDLE_CONN_TIMEOUT"
	ENV_DB_MAX_POOL_SIZE     = "DB_MAX_POOL_SIZE"
	ENV_DB_NAME_PREFIX       = "DB_DB_NAME_PREFIX"

	ENV_ADDR_STUDY_SERVICE        = "ADDR_STUDY_SERVICE"
	ENV_ADDR_EMAIL_CLIENT_SERVICE = "ADDR_EMAIL_CLIENT_SERVICE"
	ENV_GRPC_MAX_MSG_SIZE         = "GRPC_MAX_MSG_SIZE"
)

const (
	DefaultGRPCMaxMsgSize = 4194304
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
	ResearcherDBConfig      types.DBConfig
	ServiceURLs             struct {
		StudyService string `yaml:"study_service"`
		EmailClient  string `yaml:"email_client_service"`
	}
	MaxMsgSize int
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

	conf.ServiceURLs.StudyService = os.Getenv(ENV_ADDR_STUDY_SERVICE)
	conf.ServiceURLs.EmailClient = os.Getenv(ENV_ADDR_EMAIL_CLIENT_SERVICE)

	conf.SAMLConfig = &types.SAMLConfig{
		IDPUrl:          os.Getenv(ENV_SAML_IDP_URL),                   // arbitrary name to refer to IDP in the logs
		SPRootUrl:       os.Getenv(ENV_SAML_SERVICE_PROVIDER_ROOT_URL), // url of the management api
		EntityID:        os.Getenv(ENV_SAML_ENTITY_ID),
		MetaDataURL:     os.Getenv(ENV_SAML_IDP_METADATA_URL),
		SessionCertPath: os.Getenv(ENV_SAML_SESSION_CERT_PATH),
		SessionKeyPath:  os.Getenv(ENV_SAML_SESSION_KEY_PATH),
	}

	conf.ResearcherDBConfig = getResearcherDBConfig()

	if len(conf.SAMLConfig.IDPUrl) > 0 {
		conf.AllowOrigins = append(conf.AllowOrigins, conf.SAMLConfig.IDPUrl)
	}

	// Max message size for gRPC client
	conf.MaxMsgSize = DefaultGRPCMaxMsgSize
	ms, err := strconv.Atoi(os.Getenv(ENV_GRPC_MAX_MSG_SIZE))
	if err != nil {
		logger.Debug.Printf("using default max gRPC message size: %d", DefaultGRPCMaxMsgSize)
	} else {
		conf.MaxMsgSize = ms
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

func getResearcherDBConfig() types.DBConfig {
	connStr := os.Getenv(ENV_RESEARCHER_DB_CONNECTION_STR)
	username := os.Getenv(ENV_RESEARCHER_DB_USERNAME)
	password := os.Getenv(ENV_RESEARCHER_DB_PASSWORD)
	prefix := os.Getenv(ENV_RESEARCHER_DB_CONNECTION_PREFIX) // Used in test mode
	if connStr == "" || username == "" || password == "" {
		logger.Error.Fatal("Couldn't read DB credentials.")
	}
	URI := fmt.Sprintf(`mongodb%s://%s:%s@%s`, prefix, username, password, connStr)

	var err error
	Timeout, err := strconv.Atoi(os.Getenv(ENV_DB_TIMEOUT))
	if err != nil {
		logger.Error.Fatal("DB_TIMEOUT: " + err.Error())
	}
	IdleConnTimeout, err := strconv.Atoi(os.Getenv(ENV_DB_IDLE_CONN_TIMEOUT))
	if err != nil {
		logger.Error.Fatal("DB_IDLE_CONN_TIMEOUT" + err.Error())
	}
	mps, err := strconv.Atoi(os.Getenv("DB_MAX_POOL_SIZE"))
	MaxPoolSize := uint64(mps)
	if err != nil {
		logger.Error.Fatal("DB_MAX_POOL_SIZE: " + err.Error())
	}

	DBNamePrefix := os.Getenv(ENV_DB_NAME_PREFIX)

	return types.DBConfig{
		URI:             URI,
		Timeout:         Timeout,
		IdleConnTimeout: IdleConnTimeout,
		MaxPoolSize:     MaxPoolSize,
		DBNamePrefix:    DBNamePrefix,
	}
}
