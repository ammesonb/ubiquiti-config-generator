package internal

import (
	"fmt"

	"github.com/ammesonb/ubiquiti-config-generator/internal/errors"
	"github.com/ammesonb/ubiquiti-config-generator/services/configuration"
	"github.com/ammesonb/ubiquiti-config-generator/services/db"
	"github.com/ammesonb/ubiquiti-config-generator/services/filesystem"
	"github.com/charmbracelet/log"
)

var serviceRegistrationError = "failed to register service %s"

// Ensure fixed initialization order, since some services rely on others
var services = []string{
	"filesystem",
	"configuration",
	"database",
}

var serviceFuncs = map[string]func() error{
	"configuration": configuration.RegisterService,
	"filesystem":    filesystem.RegisterService,
	"database":      db.RegisterService,
}

func RegisterServices(logger *log.Logger) []error {
	var errs []error

	for _, service := range services {
		logger.Debugf("Registering service %s", service)
		if err := serviceFuncs[service](); err != nil {
			errs = append(errs, errors.ErrWithCtxParent(serviceRegistrationError, service, err))
			fmt.Println(err)
		}
	}

	return errs
}
