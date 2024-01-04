package main

import (
	"os"
	"os/signal"

	config2 "github.com/ammesonb/ubiquiti-config-generator/config"
	"github.com/ammesonb/ubiquiti-config-generator/logger"
	"github.com/ammesonb/ubiquiti-config-generator/web"
)

/*
* This app will be called when a PR is created for ANOTHER repo
* When a GitHub webhook check suite request is received, this program will do the following:
* - Create a new check run that:
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
* Convert custom YAML files into VyOS equivalents
* Load abstractions
* Merge abstractions into VyOS boot configs

* Web pages for checks and deployments

* GitHub check run/validations
* val_help from node_parser does not get surfaced anywhere
* Validation for VyOS stuff
* Validation for custom YAML nodes
  - address for host in subnet
  - subnets have addresses matching interfaces
  - firewalls referenced in network interfaces actually exist
  - others?

* Run validation command scripts on router when PR checks run

* Get existing configuration from router
* Upload a diff of existing config vs generated config to branch for viewing

* GitHub deployments
* Perform load commands
*/
func main() {
	log := logger.DefaultLogger()

	log.Debug("Reading settings")
	configData, err := config2.ReadConfig("./config.yaml")
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	config, err := config2.LoadConfig(configData)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	log.Debugf("Settings read, found %d configured routers", len(config.Devices))

	shutdownChannel := make(chan os.Signal, 1)
	// We'll accept graceful shutdowns when quit via SIGINT (Ctrl+C)
	// SIGKILL, SIGQUIT or SIGTERM (Ctrl+/) will not be caught.
	signal.Notify(shutdownChannel, os.Interrupt)

	web.StartWebhookServer(log, config, shutdownChannel)
}
