package internal

import (
	"context"
	"fmt"
	"sync"

	"github.com/ammesonb/ubiquiti-config-generator/internal/errors"
	"github.com/ammesonb/ubiquiti-config-generator/services"
	"github.com/ammesonb/ubiquiti-config-generator/services/configuration"
	"github.com/ammesonb/ubiquiti-config-generator/services/db"
	"github.com/ammesonb/ubiquiti-config-generator/services/filesystem"
	"github.com/ammesonb/ubiquiti-config-generator/services/github"
	"github.com/charmbracelet/log"
)

var serviceRegistrationError = "failed to register service %s"

// serviceList is a list of services that must be registered in a fixed order, since some services rely on others
var serviceList = []services.ServiceIndex{
	services.FilesystemServiceIndex,
	services.ConfigurationServiceIndex,
	services.DatabaseServiceIndex,
	services.GitHubServiceIndex,
}

var serviceFuncs = map[services.ServiceIndex]func(context.Context) error{
	services.ConfigurationServiceIndex: configuration.RegisterService,
	services.FilesystemServiceIndex:    filesystem.RegisterService,
	services.DatabaseServiceIndex:      db.RegisterService,
	services.GitHubServiceIndex:        github.RegisterService,
}

// RegisterServices adds all services defined in serviceList to the registry/cache
func RegisterServices(logger *log.Logger, ctx context.Context, serviceGroup *sync.WaitGroup) []error {
	var errs []error

	for _, service := range serviceList {
		logger.Debugf("Registering service %s", service)
		if err := serviceFuncs[service](ctx); err != nil {
			errs = append(errs, errors.ErrWithCtxParent(serviceRegistrationError, service, err))
			fmt.Println(err)
			continue
		} else {
			// only setup teardown for services that registered successfully
			serviceGroup.Add(1)
			go func() {
				<-ctx.Done()
				service, err := services.GetService(service)
				if err != nil {
					logger.Errorf("Failed getting service %s at shutdown: %v", service, err)
				} else {
					service.StopService(logger)
				}
			}()
		}
	}

	return errs
}

// CheckAuthorizations checks the authorizations for all services
func CheckAuthorizations(logger *log.Logger, ctx context.Context) []error {
	errs := []error{}
	githubService := github.GetService()
	if err := githubService.CheckAuthorization(ctx); err != nil {
		errs = append(errs, err)
	}
	return errs
}
