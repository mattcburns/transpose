package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"gopkg.in/yaml.v2"

	cenats "github.com/cloudevents/sdk-go/protocol/nats/v2"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/nats-io/nats.go"
)

type NatsAuth struct {
	SeedPath string `yaml:"seed_path"`
}

type NATSConfig struct {
	Host    string   `yaml:"host"`
	Auth    NatsAuth `yaml:"auth"`
	Subject string   `yaml:"subject"`
	Types   []string `yaml:"types"`
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
	parseConfigurationFile()
}

// Parse out the configuration file
// Determine which mode we're going to be running

func parseConfigurationFile() {
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
		config.notarget()
	} else {
		config.target()
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

	fmt.Println("Starting the receiver...")
	err = client.StartReceiver(context.Background(), c.pushToNATS)
	if err != nil {
		log.Fatalf("failed to start receiver: %v", err)
	}
}

func (c *Config) pushToNATS(event cloudevents.Event) {

	opts := cenats.NatsOptions()

	if c.NATS.Auth.SeedPath != "" {
		opts = append(opts, c.secureNATSOption())
	}

	sender, err := cenats.NewSender(c.NATS.Host, c.NATS.Subject, opts)
	if err != nil {
		log.Fatalf("Failed to create nats protocol: %v", err)
	}
	defer sender.Close(context.Background())

	client, err := cloudevents.NewClient(sender)
	if err != nil {
		log.Fatalf("Failed to create cloudevent client: %v", err)
	}

	result := client.Send(context.Background(), event)
	if cloudevents.IsUndelivered(result) {
		log.Fatalf("Failed to send cloudevent: %v", result)
	}
}

// fromNATS
// Startup the NATS client to listen for incoming events

func (c *Config) target() {
	ctx := context.Background()

	opts := cenats.NatsOptions()

	if c.NATS.Auth.SeedPath != "" {
		opts = append(opts, c.secureNATSOption())
	}

	consumer, err := cenats.NewConsumer(c.NATS.Host, c.NATS.Subject, opts)
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

func (c *Config) secureNATSOption() nats.Option {
	nkey, err := nats.NkeyOptionFromSeed(c.NATS.Auth.SeedPath)
	if err != nil {
		log.Fatalf("Failed to parse nkey seed: %v", err)
	}

	return nkey
}
