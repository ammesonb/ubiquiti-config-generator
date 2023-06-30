package main

import (
	"os"

	"github.com/ammesonb/ubiquiti-config-generator/vyos"
	"github.com/charmbracelet/log"
)

/* TODO:
 * Settings parser
 * Templates parser
 * Load config
 * Convert custom YAML files into VyOS equivalents
 * Get existing configuration from router
 * Diff configs
 * Convert GitHub web hook stuff
 * Run validation command scripts on router when PR checks run
 * Perform load commands
 */
func main() {
	log.Debug("Reading settings")
	configData, err := ReadConfig("./config.yaml")
	if err != nil {
		log.Fatal(err.Error())
		os.Exit(1)
	}

	config, err := LoadConfig(configData)
	if err != nil {
		log.Fatal(err.Error())
		os.Exit(1)
	}

	log.Debugf("Parsing templates from %s", config.TemplatesDir)
	_, err = vyos.Parse(config.TemplatesDir)
	if err != nil {
		log.Fatal(err.Error())
		os.Exit(1)
	}
}
