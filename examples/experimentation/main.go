package main

import (
	"context"
	"log"
	"os"

	ffclient "github.com/thomaspoignant/go-feature-flag"
	"github.com/thomaspoignant/go-feature-flag/ffexporter"
	"github.com/thomaspoignant/go-feature-flag/ffuser"
)

func main() {
	// Before running this example you should edit the date in the flag file (examples/experimentation/flags.yaml)
	// The important part is the experimentation configuration

	// Init ffclient with a file retriever.
	err := ffclient.Init(ffclient.Config{
		PollInterval: 10,
		Logger:       log.New(os.Stdout, "", 0),
		Context:      context.Background(),
		Retriever: &ffclient.FileRetriever{
			Path: "examples/experimentation/flags.yaml",
		},
		DataExporter: ffclient.DataExporter{
			FlushInterval:    10,
			MaxEventInMemory: 2,
			Exporter: &ffexporter.Log{
				Format: "[{{ .FormattedDate}}] user=\"{{ .UserKey}}\", flag=\"{{ .Key}}\", value=\"{{ .Value}}\", variation=\"{{ .Variation}}\"",
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
	user1 := ffuser.NewAnonymousUser("332460b9-a8aa-4f7a-bc5d-9cc33632df9a")
	user2 := ffuser.NewAnonymousUser("91ff5618-6cbb-4f54-a038-3e99b078f560")
	_, _ = ffclient.StringVariation("experimentation-flag", user1, "error")
	_, _ = ffclient.StringVariation("experimentation-flag", user2, "error")

	// If the current time is in the range of the experimentation you should have an output in your log
	// that looks like this:
	//
	// [2021-04-20T17:11:40+02:00] user="332460b9-a8aa-4f7a-bc5d-9cc33632df9a", flag="experimentation-flag", value="B", variation="True"
	// [2021-04-20T17:11:40+02:00] user="91ff5618-6cbb-4f54-a038-3e99b078f560", flag="experimentation-flag", value="A", variation="False"
	//
	// You can see that the variation is True and False, it means that the flag has been evaluated and user1 is in cohort B
	// while user2 is in cohort A.

	// If you change the date again and the current time is not in the range anymore, your output will looks like:
	//
	// [2021-04-20T17:20:38+02:00] user="332460b9-a8aa-4f7a-bc5d-9cc33632df9a", flag="experimentation-flag", value="A", variation="Default"
	// [2021-04-20T17:20:38+02:00] user="91ff5618-6cbb-4f54-a038-3e99b078f560", flag="experimentation-flag", value="A", variation="Default"
}
