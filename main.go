package main

import (
	"context"
	"log"
	"os"

	"gopkg.in/yaml.v2"

	cenats "github.com/cloudevents/sdk-go/protocol/nats/v2"
	cloudevents "github.com/cloudevents/sdk-go/v2"
)

type NATSConfig struct {
	Host    string      `yaml:"host"`
	Auth    interface{} `yaml:"auth,omitempty"`
	Subject string      `yaml:"subject"`
	Types   []string    `yaml:"types,omitempty"`
}

type TargetConfig struct {
	Host string      `yaml:"host"`
	Auth interface{} `yaml:"auth,omitempty"`
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
		log.Fatalf("Unable to read config file: %v", err)
	}

	var config Config

	err = yaml.Unmarshal(contents, &config)
	if err != nil {
		log.Fatalf("Unable to parse config file: %v", err)
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
	// Start the HTTP receiever with handler attached
	client, err := cloudevents.NewClientHTTP()
	if err != nil {
		log.Fatalf("failed to create client: %v", err)
	}

	err = client.StartReceiver(context.Background(), c.pushToNATS)
	if err != nil {
		log.Fatalf("failed to start receiver: %v", err)
	}
}

func (c *Config) pushToNATS(event cloudevents.Event) {
	sender, err := cenats.NewSender(c.NATS.Host, c.NATS.Subject, cenats.NatsOptions())
	if err != nil {
		log.Fatalf("Failed to create nats protocol: %v", err)
	}
	defer sender.Close(context.Background())

	client, err := cloudevents.NewClient(sender)
	if err != nil {
		log.Fatalf("Failed to create cloudevent client: %v", err)
	}

	client.Send(context.Background(), event)
}

// fromNATS
// Startup the NATS client to listen for incoming events

func (c *Config) target() {
	ctx := context.Background()

	consumer, err := cenats.NewConsumer(c.NATS.Host, c.NATS.Subject, cenats.NatsOptions())
	if err != nil {
		log.Fatalf("failed to create nats protocol: %v", err)
	}

	defer consumer.Close(ctx)

	client, err := cloudevents.NewClient(consumer)
	if err != nil {
		log.Fatalf("failed to create cloudevent client: %v", err)
	}

	for {
		err = client.StartReceiver(ctx, c.pullFromNats)
		if err != nil {
			log.Fatalf("failed to start cloudevent receiver: %v", err)
		}
	}
}

func (c *Config) pullFromNats(_ context.Context, event cloudevents.Event) {
	client, err := cloudevents.NewClientHTTP()
	if err != nil {
		log.Fatalf("failed to create client: %v", err)
	}

	ctx := cloudevents.ContextWithTarget(context.Background(), c.Target.Host)

	client.Send(ctx, event)

}
