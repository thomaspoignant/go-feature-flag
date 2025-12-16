package main

import (
	"context"
	"log"
	"log/slog"
	"time"

	"github.com/thomaspoignant/go-feature-flag/ffcontext"

	"github.com/thomaspoignant/go-feature-flag/exporter/fileexporter"
	"github.com/thomaspoignant/go-feature-flag/retriever/fileretriever"

	ffclient "github.com/thomaspoignant/go-feature-flag"
)

func main() {
	// Init ffclient with a file retriever.
	err := ffclient.Init(ffclient.Config{
		PollingInterval: 10 * time.Second,
		LeveledLogger:   slog.Default(),
		Context:         context.Background(),
		Retriever: &fileretriever.Retriever{
			Path: "examples/data_export_file/flags.goff.yaml",
		},
		DataExporter: ffclient.DataExporter{
			FlushInterval:    1 * time.Second,
			MaxEventInMemory: 100,
			Exporter: &fileexporter.Exporter{
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
	user1 := ffcontext.
		NewEvaluationContextBuilder("aea2fdc1-b9a0-417a-b707-0c9083de68e3").
		AddCustom("anonymous", true).
		Build()
	user2 := ffcontext.NewEvaluationContext("332460b9-a8aa-4f7a-bc5d-9cc33632df9a")

	_, _ = ffclient.BoolVariation("new-admin-access", user1, false)
	_, _ = ffclient.BoolVariation("new-admin-access", user2, false)
	_, _ = ffclient.StringVariation("unknown-flag", user1, "defaultValue")
	_, _ = ffclient.JSONVariation("unknown-flag-2", user1, map[string]any{"test": "toto"})

	// Wait 2 seconds to have a second file
	time.Sleep(2 * time.Second)
	_, _ = ffclient.BoolVariation("new-admin-access", user1, false)
	_, _ = ffclient.BoolVariation("new-admin-access", user2, false)

	/*
		The output will be something like that:

		flag-variation-EXAMPLE-<timestamp>.json:
			{"kind":"feature","contextKind":"anonymousUser","userKey":"aea2fdc1-b9a0-417a-b707-0c9083de68e3","creationDate":1618234129,"key":"new-admin-access","variation":"True","value":true,"default":false,"source":"SERVER"}
			{"kind":"feature","contextKind":"user","userKey":"332460b9-a8aa-4f7a-bc5d-9cc33632df9a","creationDate":1618234129,"key":"new-admin-access","variation":"False","value":false,"default":false,"source":"SERVER"}
			{"kind":"feature","contextKind":"anonymousUser","userKey":"aea2fdc1-b9a0-417a-b707-0c9083de68e3","creationDate":1618234129,"key":"unknown-flag","variation":"SdkDefault","value":"defaultValue","default":true,"source":"SERVER"}
			{"kind":"feature","contextKind":"anonymousUser","userKey":"aea2fdc1-b9a0-417a-b707-0c9083de68e3","creationDate":1618234129,"key":"unknown-flag-2","variation":"SdkDefault","value":{"test":"toto"},"default":true,"source":"SERVER"}
		----
		flag-variation-EXAMPLE-<timestamp>.json:
			{"kind":"feature","contextKind":"anonymousUser","userKey":"aea2fdc1-b9a0-417a-b707-0c9083de68e3","creationDate":1618234131,"key":"new-admin-access","variation":"True","value":true,"default":false,"source":"SERVER"}
			{"kind":"feature","contextKind":"user","userKey":"332460b9-a8aa-4f7a-bc5d-9cc33632df9a","creationDate":1618234131,"key":"new-admin-access","variation":"False","value":false,"default":false,"source":"SERVER"}
	*/
}
