package main

import (
	"os"

	"github.com/charmbracelet/log"

	"github.com/ammesonb/ubiquiti-config-generator/logger"
)

/*
TODO:
* This app will be called when a PR is created for ANOTHER repo
* So then will need to compare the new definitions of that against the existing configuration, which will need to be dumped
* from the router's current config
* Cannot cache it since it could change during the lifetime of a branch, which would result in stale diffs

* Check router connectivity

* ParseBootDefinitions never called in actual code
* Convert custom YAML files into VyOS equivalents
* val_help from node_parser does not get surfaced anywhere
* Merge custom nodes into VyOS templates
* Validation for custom YAML nodes
* Validation for VyOS stuff
* GitHub web hook app
* GitHub check suite/validations
* Run validation command scripts on router when PR checks run
* Get existing configuration from router
* Upload a diff of existing config vs generated config to branch for viewing
* GitHub deployments
* Perform load commands
*/
func main() {
	logger := logger.DefaultLogger()

	logger.Debug("Reading settings")
	configData, err := ReadConfig("./config.yaml")
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	config, err := LoadConfig(configData)
	if err != nil {
		logger.Fatal(err)
		os.Exit(1)
	}

	logger.Debugf("Settings read, found %d configured routers", len(*config))

	/*
		log.Debugf("Parsing templates from %s", config.TemplatesDir)
		node, err := vyos.Parse(config.TemplatesDir)
		if err != nil {
			logger.Fatal(err)
			os.Exit(1)
		}

		// This is just to generate for testing
		res, err := yaml.Marshal(node)
		if err != nil {
			logger.Fatal(err)
			os.Exit(1)
		}

		if err = os.WriteFile("./generated-node-fixtures.yaml", res, 0644); err != nil {
			logger.Fatal(err)
			os.Exit(1)
		}*/
}
