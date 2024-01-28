package logger

import (
	"os"
	"strings"

	"github.com/charmbracelet/log"
)

var logger *log.Logger

var logEnv = "LOGLEVEL"
var levels = map[string]log.Level{
	"debug":   log.DebugLevel,
	"info":    log.InfoLevel,
	"warn":    log.WarnLevel,
	"warning": log.WarnLevel,
	"error":   log.ErrorLevel,
	"fatal":   log.FatalLevel,
}
var defaultLogLevel = log.InfoLevel

func getEnvLogLevel() log.Level {
	level := strings.ToLower(os.Getenv(logEnv))

	logLevel, ok := levels[level]
	if !ok {
		return defaultLogLevel
	}

	return logLevel
}

// DefaultLogger returns a default logger
func DefaultLogger() *log.Logger {
	if logger == nil {
		logger = log.New(os.Stderr)
	}

	logger.SetLevel(getEnvLogLevel())

	return logger
}
