package main

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config contains the runtime settings needed to analyze and deploy configurations
type Config struct {
	TemplatesDir string `yaml:"templatesDir"`
	ConfigFile   string `yaml:"configFile"`
}

// ReadConfig takes a path and returns its contents
func ReadConfig(path string) ([]byte, error) {
	return os.ReadFile(path)
}

// LoadConfig takes YAML config data and loads it into the struce
func LoadConfig(config []byte) (*Config, error) {
	conf := &Config{}
	err := yaml.Unmarshal(config, conf)

	if err != nil {
		return nil, fmt.Errorf("Failed loading config: %+v", err)
	}

	return conf, nil
}
