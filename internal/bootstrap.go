package internal

import (
	"context"
	"os"

	"github.com/ammesonb/ubiquiti-config-generator/services/configuration"
	"github.com/ammesonb/ubiquiti-config-generator/services/filesystem"
	"github.com/ammesonb/ubiquiti-config-generator/services/github"
	"github.com/charmbracelet/log"
)

var (
	defaultConfigPath = "./config.yaml"
	configEnvVar      = "UBQ_CONFIGURATION_FILE"
)

// CheckAuthorizations checks the authorizations for all services
func CheckAuthorizations(logger *log.Logger, ctx context.Context, configService configuration.Service, fsService filesystem.Service, githubService github.Service) []error {
	errs := []error{}
	if err := githubService.CheckAuthorization(ctx, configService, fsService); err != nil {
		errs = append(errs, err)
	}
	return errs
}

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
