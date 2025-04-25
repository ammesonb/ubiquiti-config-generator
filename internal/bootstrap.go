package internal

import (
	"context"
	"os"

	"github.com/ammesonb/ubiquiti-config-generator/services/configuration"
	"github.com/ammesonb/ubiquiti-config-generator/services/filesystem"
	"github.com/charmbracelet/log"
)

var (
	defaultConfigPath = "./config.yaml"
	configEnvVar      = "UBQ_CONFIGURATION_FILE"
)

// LoadConfiguration
func LoadConfiguration(logger *log.Logger, ctx context.Context, fsService filesystem.Service) (configuration.Service, error) {
	config := configuration.YAMLService{}
	configPath := os.Getenv(configEnvVar)
	if configPath == "" {
		configPath = defaultConfigPath
	}

	logger.Info("Loading configuration from %s", configPath)

	err := config.Load(configPath, fsService)
	if err != nil {
		return nil, err
	}
	return &config, nil
}
