package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	ffclient "github.com/thomaspoignant/go-feature-flag"
	"github.com/thomaspoignant/go-feature-flag/ffuser"
)

func main() {
	// Before running this code please check the flag.yaml file
	// You can update the dates of the steps in the rollout to see it working.

	err := ffclient.Init(ffclient.Config{
		PollInterval: 10,
		Logger:       log.New(os.Stdout, "", 0),
		Context:      context.Background(),
		Retriever: &ffclient.FileRetriever{
			Path: "examples/rollout_scheduled/flags.yaml",
		},
	})

	// Check init errors.
	if err != nil {
		log.Fatal(err)
	}
	// defer closing ffclient
	defer ffclient.Close()

	// create users
	user := ffuser.NewUserBuilder("785a14bf-d2c5-4caa-9c70-2bbc4e3732a5").
		AddCustom("beta", "true").
		Build()

	// Call multiple time the same flag to see the change in time.
	for true {
		time.Sleep(1 * time.Second)
		fmt.Println(ffclient.BoolVariation("new-admin-access", user, false))
	}
}
