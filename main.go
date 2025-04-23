package main

import (
	"context"
	"os/signal"
	"sync"
	"syscall"

	"github.com/ammesonb/ubiquiti-config-generator/internal"

	"github.com/ammesonb/ubiquiti-config-generator/console_logger"
	"github.com/ammesonb/ubiquiti-config-generator/web"
)

/*
* This app will be called when a PR is created for ANOTHER repo
* When a GitHub webhook check suite request is received, this program will do the following:
* - Create a new check run that:
*   - Check which files were modified and determine which devices will need updates
*   - Parses the configuration and abstractions and merges the VyOS equivalents
*   - Validates the new configuration
*   - Gets the live config for affected production routers and diffs it against the new one
*   - Posts a PR comment with the validation results and diff
* - On branch merge/push to the main branch:
*   - Creates a new deployment
*   - Loads the new configuration
*
* There is a small configurable web server set up that reports the status of checks and deployments as well, with logs
* of actions and results.
*
* VyOS Terminology:
* - Nodes are the result of parsing templates, which define the hierarchy the validations for the schema
* - Definitions are the values contained in an actual configuration, which will be tested against node specifications

TODO:
* Determine which files were changed and which devices need updating
* Convert custom YAML files into VyOS equivalents
* Load abstractions
* Merge abstractions into VyOS boot configs

* Web pages for checks and deployments

* GitHub check run/validations
* When loading device config/diffs, set up NAT firewall counters
* val_help from node_parser does not get surfaced anywhere
* Validation for VyOS stuff
* Validation for custom YAML nodes
  - address for host in a subnet's CIDR
  - mac addresses are valid
  - no duplicate MACs per network
  - no duplicate IP addresses per network
  - subnets have addresses matching interfaces
  - subnets do not overlap
  - firewalls referenced in network interfaces actually exist
  - port only forwarded to one host
  - others?

* Run validation command scripts on router when PR checks run

* Get existing configuration from router
* Upload a diff of existing config vs generated config to branch for viewing

* GitHub deployments
* Secrets should be pulled from existing config, not committed in YAML files
*  - Need per-device secret indicator for configuration path/node so we know to retrieve it prior to deploy
*  - Or maybe injected via environment variables?
* Perform load commands
*/
func main() {
	log := console_logger.DefaultLogger()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGABRT, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGKILL)
	defer stop()

	log.Info("Registering services")
	var serviceGroup sync.WaitGroup
	errs := internal.RegisterServices(log, ctx, &serviceGroup)
	if len(errs) > 0 {
		log.Fatal(errs)
	}
	if errs := internal.CheckAuthorizations(log, ctx); len(errs) > 0 {
		log.Fatal(errs)
	}

	log.Debug("Services loaded")

	web.StartWebhookServer(log, ctx)

	<-ctx.Done()

	log.Warn("Received signal, shutting down gracefully")

	stop()
	log.Warn("Gracefully shutting down services...Press CTRL + C (or other signal) again to force")
	serviceGroup.Wait()
}
