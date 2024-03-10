package console_logger

import (
	"github.com/charmbracelet/log"
	"github.com/stretchr/testify/assert"
	"os"
	"strings"
	"testing"
)

func TestDefaultLogger(t *testing.T) {
	startLevel := os.Getenv(logEnv)

	var expected log.Level
	if len(startLevel) == 0 {
		assert.NoError(t, os.Setenv(logEnv, "WARN"))
		expected = log.WarnLevel
	} else {
		expected = levels[strings.ToLower(os.Getenv(logEnv))]
	}

	logger := DefaultLogger()
	assert.Equal(t, logger.GetLevel(), expected)

	assert.NoError(t, os.Setenv(logEnv, ""))
	logger = DefaultLogger()
	assert.Equal(t, logger.GetLevel(), defaultLogLevel)

	assert.NoError(t, os.Setenv(logEnv, startLevel))
}
