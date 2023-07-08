package main

import (
	"os"

	"github.com/ammesonb/ubiquiti-config-generator/logger"
	"github.com/ammesonb/ubiquiti-config-generator/vyos"
	"github.com/charmbracelet/log"
	"gopkg.in/yaml.v3"
)

/*
	 TODO:
	 * Load config
	   - allow for modular configs - config.boot always checked, plu
		   either interfaces.boot or maybe interfaces/<others>.boot?
		 - With recursion support?
	 * Convert custom YAML files into VyOS equivalents
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

	//if err = os.WriteFile("./generated-node-fixtures.yaml", res, 0644); err != nil {
	////logger.Fatal(err)
	//os.Exit(1)
	//}
}
