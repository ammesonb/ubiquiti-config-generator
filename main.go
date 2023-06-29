package main

import (
	"fmt"

	"github.com/ammesonb/ubiquiti-config-generator/vyos"
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
	fmt.Println("vim-go")

	configData, err := ReadConfig("./config.yaml")
	if err != nil {
		panic(err)
	}

	config, err := LoadConfig(configData)
	if err != nil {
		panic(err)
	}

	_, err = vyos.Parse(config.TemplatesDir)
	if err != nil {
		panic(err)
	}
}
