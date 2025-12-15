package main

import (
	"context"
	"log"
	"log/slog"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"

	kex "github.com/thomaspoignant/go-feature-flag/exporter/kinesisexporter"

	ffclient "github.com/thomaspoignant/go-feature-flag"
	"github.com/thomaspoignant/go-feature-flag/ffcontext"
	"github.com/thomaspoignant/go-feature-flag/retriever/fileretriever"
)

func main() {
	/*
		1. Start a kinesis server by running `docker-compose -f examples/data_export_kinesis/docker-compose.yml up`
		2. Create a stream: docker exec $(docker ps | grep cp-kinesis |  awk '{print $1}') kafka-topics --create --topic go-feature-flag-events --bootstrap-server localhost:9092
		3. Run this example
			-> if you check the logs, you should see the events being sent 1 by 1 to kafka.
		4. Read the items in the topic: docker exec $(docker ps | grep cp-kafka |  awk '{print $1}')  kafka-console-consumer --bootstrap-server localhost:9092 --topic go-feature-flag-events --from-beginning
	*/

	config, err := config.LoadDefaultConfig(
		context.Background(),
		config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider("test", "test", ""),
		),
		config.WithRegion("us-east-1"),
		config.WithBaseEndpoint("http://localhost:4566"),
	)

	if err != nil {
		panic("Can't instantiate localstack connection credentials")
	}

	err = ffclient.Init(ffclient.Config{
		PollingInterval: 10 * time.Second,
		LeveledLogger:   slog.Default(),
		Context:         context.Background(),
		Retriever: &fileretriever.Retriever{
			Path: "examples/data_export_kinesis/flags.goff.yaml",
		},
		DataExporter: ffclient.DataExporter{
			FlushInterval:    1 * time.Second,
			MaxEventInMemory: 100,
			Exporter: &kex.Exporter{
				Settings: kex.NewSettings(
					kex.WithStreamName("test-stream"),
					kex.WithExplicitHashKey("0"),
				),
				AwsConfig: &config,
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
