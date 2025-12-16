package main

import (
	"context"
	"github.com/thomaspoignant/go-feature-flag/exporter/kafkaexporter"
	"log"
	"log/slog"
	"time"

	ffclient "github.com/thomaspoignant/go-feature-flag"
	"github.com/thomaspoignant/go-feature-flag/ffcontext"
	"github.com/thomaspoignant/go-feature-flag/retriever/fileretriever"
)

func main() {
	/*
		1. Start a kafka server by running `docker-compose -f examples/data_export_kafka/docker-compose.yml up`
		2. Create a topic: docker exec $(docker ps | grep cp-kafka |  awk '{print $1}') kafka-topics --create --topic go-feature-flag-events --bootstrap-server localhost:9092
		3. Run this example
			-> if you check the logs, you should see the events being sent 1 by 1 to kafka.
		4. Read the items in the topic: docker exec $(docker ps | grep cp-kafka |  awk '{print $1}')  kafka-console-consumer --bootstrap-server localhost:9092 --topic go-feature-flag-events --from-beginning
	*/
	err := ffclient.Init(ffclient.Config{
		PollingInterval: 10 * time.Second,
		LeveledLogger:   slog.Default(),
		Context:         context.Background(),
		Retriever: &fileretriever.Retriever{
			Path: "examples/data_export_s3/flags.goff.yaml",
		},
		DataExporter: ffclient.DataExporter{
			FlushInterval:    1 * time.Second,
			MaxEventInMemory: 100,
			Exporter: &kafkaexporter.Exporter{
				Settings: kafkaexporter.Settings{
					Topic:     "go-feature-flag-events",
					Addresses: []string{"localhost:29092"},
				},
			},
		},
	})
	// Check init errors.
	if err != nil {
		log.Fatal(err)
	}
	// defer closing ffclient
	defer ffclient.Close()

	// create users
	user1 := ffcontext.
		NewEvaluationContextBuilder("aea2fdc1-b9a0-417a-b707-0c9083de68e3").
		AddCustom("anonymous", true).
		Build()
	user2 := ffcontext.NewEvaluationContext("332460b9-a8aa-4f7a-bc5d-9cc33632df9a")

	_, _ = ffclient.BoolVariation("new-admin-access", user1, false)
	_, _ = ffclient.BoolVariation("new-admin-access", user2, false)
	_, _ = ffclient.StringVariation("unknown-flag", user1, "defaultValue")
	_, _ = ffclient.JSONVariation("unknown-flag-2", user1, map[string]any{"test": "toto"})
	_, _ = ffclient.BoolVariation("new-admin-access", user1, false)
	_, _ = ffclient.BoolVariation("new-admin-access", user2, false)

}
