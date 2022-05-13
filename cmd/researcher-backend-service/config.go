package main

import (
	"os"
	"strings"

	"github.com/coneno/logger"
)

const (
	ENV_LOG_LEVEL = "LOG_LEVEL"

	ENV_RESEARCHER_BACKEND_LISTEN_PORT = "RESEARCHER_BACKEND_LISTEN_PORT"
	ENV_CORS_ALLOW_ORIGINS             = "CORS_ALLOW_ORIGINS"
)

// Config is the structure that holds all global configuration data
type Config struct {
	Port              string
	AllowOrigins      []string
	APIKeyForRW       []string
	APIKeyForReadOnly []string
	LogLevel          logger.LogLevel
}

func InitConfig() Config {
	conf := Config{}
	conf.Port = os.Getenv(ENV_RESEARCHER_BACKEND_LISTEN_PORT)
	conf.AllowOrigins = strings.Split(os.Getenv(ENV_CORS_ALLOW_ORIGINS), ",")

	conf.LogLevel = getLogLevel()

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
