package main

import (
	"context"
	"log"
	"os"
	"time"

	ffclient "github.com/thomaspoignant/go-feature-flag"
	"github.com/thomaspoignant/go-feature-flag/ffexporter"
	"github.com/thomaspoignant/go-feature-flag/ffuser"
)

func main() {
	// Init ffclient with a file retriever.
	err := ffclient.Init(ffclient.Config{
		PollingInterval: 10 * time.Second,
		Logger:          log.New(os.Stdout, "", 0),
		Context:         context.Background(),
		Retriever: &ffclient.FileRetriever{
			Path: "examples/data_export_file/flags.yaml",
		},
		DataExporter: ffclient.DataExporter{
			FlushInterval:    1 * time.Second,
			MaxEventInMemory: 100,
			Exporter: &ffexporter.File{
				Format:    "json",
				OutputDir: "./examples/data_export_file/",
				Filename:  " flag-variation-EXAMPLE-{{ .Timestamp}}.{{ .Format}}",
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
	user1 := ffuser.NewAnonymousUser("aea2fdc1-b9a0-417a-b707-0c9083de68e3")
	user2 := ffuser.NewUser("332460b9-a8aa-4f7a-bc5d-9cc33632df9a")

	_, _ = ffclient.BoolVariation("new-admin-access", user1, false)
	_, _ = ffclient.BoolVariation("new-admin-access", user2, false)
	_, _ = ffclient.StringVariation("unknown-flag", user1, "defaultValue")
	_, _ = ffclient.JSONVariation("unknown-flag-2", user1, map[string]interface{}{"test": "toto"})

	// Wait 2 seconds to have a second file
	time.Sleep(2 * time.Second)
	_, _ = ffclient.BoolVariation("new-admin-access", user1, false)
	_, _ = ffclient.BoolVariation("new-admin-access", user2, false)

	/*
		The output will be something like that:

		flag-variation-EXAMPLE-<timestamp>.json:
			{"kind":"feature","contextKind":"anonymousUser","userKey":"aea2fdc1-b9a0-417a-b707-0c9083de68e3","creationDate":1618234129,"key":"new-admin-access","variation":"True","value":true,"default":false}
			{"kind":"feature","contextKind":"user","userKey":"332460b9-a8aa-4f7a-bc5d-9cc33632df9a","creationDate":1618234129,"key":"new-admin-access","variation":"False","value":false,"default":false}
			{"kind":"feature","contextKind":"anonymousUser","userKey":"aea2fdc1-b9a0-417a-b707-0c9083de68e3","creationDate":1618234129,"key":"unknown-flag","variation":"SdkDefault","value":"defaultValue","default":true}
			{"kind":"feature","contextKind":"anonymousUser","userKey":"aea2fdc1-b9a0-417a-b707-0c9083de68e3","creationDate":1618234129,"key":"unknown-flag-2","variation":"SdkDefault","value":{"test":"toto"},"default":true}
		----
		flag-variation-EXAMPLE-<timestamp>.json:
			{"kind":"feature","contextKind":"anonymousUser","userKey":"aea2fdc1-b9a0-417a-b707-0c9083de68e3","creationDate":1618234131,"key":"new-admin-access","variation":"True","value":true,"default":false}
			{"kind":"feature","contextKind":"user","userKey":"332460b9-a8aa-4f7a-bc5d-9cc33632df9a","creationDate":1618234131,"key":"new-admin-access","variation":"False","value":false,"default":false}
	*/
}
