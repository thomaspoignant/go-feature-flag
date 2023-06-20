package main

import (
	"context"
	"fmt"
	"github.com/thomaspoignant/go-feature-flag/ffcontext"
	"github.com/thomaspoignant/go-feature-flag/retriever"
	"log"
	"os"
	"time"

	"github.com/thomaspoignant/go-feature-flag/exporter/fileexporter"
	"github.com/thomaspoignant/go-feature-flag/retriever/fileretriever"

	ffclient "github.com/thomaspoignant/go-feature-flag"
)

func main() {
	// Init ffclient with multiple file retrievers.
	err := ffclient.Init(ffclient.Config{
		PollingInterval: 10 * time.Second,
		Logger:          log.New(os.Stdout, "", 0),
		Context:         context.Background(),
		Retrievers: []retriever.Retriever{
			&fileretriever.Retriever{
				Path: "examples/retriever_multiple_config_files/flags.yaml",
			},
			&fileretriever.Retriever{
				Path: "examples/retriever_multiple_config_files/flags2.yaml",
			},
		},
		DataExporter: ffclient.DataExporter{
			FlushInterval:    1 * time.Second,
			MaxEventInMemory: 100,
			Exporter: &fileexporter.Exporter{
				OutputDir: "./",
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

	// user1
	user1HasAccessToNewAdmin, err := ffclient.BoolVariation("new-admin-access", user1, false)
	if err != nil {
		// we log the error, but we still have a meaningful value in user1HasAccessToNewAdmin (the default value).
		log.Printf("something went wrong when getting the flag: %v", err)
	}
	if !user1HasAccessToNewAdmin {
		fmt.Println("user1 should not have access to the new admin because we are using the configuration from flag2.yaml")
	}
}
