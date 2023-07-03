package logger

import (
	"os"

	"github.com/charmbracelet/log"
)

var logger *log.Logger

// DefaultLogger returns a default logger
func DefaultLogger() *log.Logger {
	if logger == nil {
		logger = log.New(os.Stderr)
	}

	return logger
}
