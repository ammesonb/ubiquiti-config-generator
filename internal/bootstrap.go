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
	"github.com/charmbracelet/log"
)

var serviceRegistrationError = "failed to register service %s"

// serviceList is a list of services that must be registered in a fixed order, since some services rely on others
var serviceList = []string{
	"filesystem",
	"configuration",
	"database",
}

var serviceFuncs = map[string]func() error{
	"configuration": configuration.RegisterService,
	"filesystem":    filesystem.RegisterService,
	"database":      db.RegisterService,
}

// RegisterServices adds all services defined in serviceList to the registry/cache
func RegisterServices(logger *log.Logger, ctx context.Context, serviceGroup *sync.WaitGroup) []error {
	var errs []error

	for _, service := range serviceList {
		logger.Debugf("Registering service %s", service)
		if err := serviceFuncs[service](); err != nil {
			errs = append(errs, errors.ErrWithCtxParent(serviceRegistrationError, service, err))
			fmt.Println(err)
			continue
		}
	}

	if len(errs) == 0 {
		for _, service := range serviceList {
			serviceGroup.Add(1)
			go func() {
				<-ctx.Done()
				service, err := services.GetService(services.ServiceIndex(service))
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
