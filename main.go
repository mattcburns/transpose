package main

import (
	"log"
	"os"

	"gopkg.in/yaml.v2"
)

type NATSConfig struct {
	Host    string      `yaml:"host"`
	Auth    interface{} `yaml:"auth"`
	Subject string      `yaml:"subject"`
	Types   []string    `yaml:"types,omitempty"`
}

type TargetConfig struct {
	Host string      `yaml:"host"`
	Auth interface{} `yaml:"auth"`
}

type Config struct {
	NATS   NATSConfig   `yaml:"nats"`
	Target TargetConfig `yaml:"target"`
}

func main() {
	config, target := parseConfigurationFile()
	if target {
		config.target()
	} else {
		config.notarget()
	}
}

// Parse out the configuration file
// Determine which mode we're going to be running

func parseConfigurationFile() (Config, bool) {
	contents, err := os.ReadFile("config.yaml")
	if err != nil {
		log.Fatalf("Unable to read config file: %v", err.Error())
	}

	var config Config

	err = yaml.Unmarshal(contents, &config)
	if err != nil {
		log.Fatalf("Unable to parse config file: %v", err.Error())
	}

	if config.Target.Host == "" {
		return config, true
	} else {
		return config, false
	}
}

// toNATS
// Startup the http server to accept incoming cloudevents
// Verify connectivity with NATS
func (c *Config) notarget() {

}

// fromNATS
// Startup the NATS client to listen for incoming events

func (c *Config) target() {

}
