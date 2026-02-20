package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"time"

	"github.com/thomaspoignant/go-feature-flag/ffcontext"

	"github.com/thomaspoignant/go-feature-flag/retriever/fileretriever"

	ffclient "github.com/thomaspoignant/go-feature-flag"
)

func main() {
	// Before running this code please check the flag.yaml file
	// You can update the dates of the steps in the rollout to see it working.

	err := ffclient.Init(ffclient.Config{
		PollingInterval: 10 * time.Second,
		LeveledLogger:   slog.Default(),
		Context:         context.Background(),
		Retriever: &fileretriever.Retriever{
			Path: "examples/rollout_scheduled/flags.goff.yaml",
		},
	})
	// Check init errors.
	if err != nil {
		log.Fatal(err)
	}
	// defer closing ffclient
	defer ffclient.Close()

	// create users
	user := ffcontext.NewEvaluationContextBuilder("785a14bf-d2c5-4caa-9c70-2bbc4e3732a56").
		AddCustom("beta", "true").
		Build()

	user2 := ffcontext.NewEvaluationContextBuilder("785a14bf-d2c5-4caa-9c70-2bbc4e3732a5").
		Build()

	// Call multiple time the same flag to see the change in time.
	for true {
		time.Sleep(1 * time.Second)
		details, _ := ffclient.BoolVariationDetails("new-admin-access", user, false)
		fmt.Println("Value user1: ", details.Value)
		fmt.Println("Reason user1: ", details.Reason)
		details, _ = ffclient.BoolVariationDetails("new-admin-access", user2, false)
		fmt.Println("Value user2: ", details.Value)
		fmt.Println("Reason user2: ", details.Reason)
	}
}
