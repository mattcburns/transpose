package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"

	cenats "github.com/cloudevents/sdk-go/protocol/nats/v2"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/google/uuid"
)

func main() {

	wg := new(sync.WaitGroup)

	wg.Add(1)

	go func() {
		// Basic HTTP Target
		http.HandleFunc("/message", func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(w, "Hello, %q", r.URL.Path)
		})
		fmt.Println("Starting the receiver on port 8085...")
		http.ListenAndServe(":8085", nil)
		wg.Done()
	}()

	for i := 0; i < 10; i++ {
		event := cloudevents.NewEvent()
		event.SetSource("example/uri")
		event.SetType("example.type")
		event.SetData(cloudevents.ApplicationJSON, map[string]string{"hello": "world"})
		event.SetID(uuid.UUID{}.String())

		sender, err := cenats.NewSender("localhost:4222", "local.test", cenats.NatsOptions())
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
}
