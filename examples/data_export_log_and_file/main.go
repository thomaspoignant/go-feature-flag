package main

import (
	"context"
	"log"
	"log/slog"
	"time"

	"github.com/thomaspoignant/go-feature-flag/ffcontext"

	"github.com/thomaspoignant/go-feature-flag/exporter/fileexporter"
	"github.com/thomaspoignant/go-feature-flag/exporter/logsexporter"
	"github.com/thomaspoignant/go-feature-flag/retriever/fileretriever"

	ffclient "github.com/thomaspoignant/go-feature-flag"
)

func main() {
	// Init ffclient with multiple exporters
	err := ffclient.Init(ffclient.Config{
		PollingInterval: 10 * time.Second,
		LeveledLogger:   slog.Default(),
		Context:         context.Background(),
		Retriever: &fileretriever.Retriever{
			Path: "examples/data_export_log_and_file/flags.goff.yaml",
		},
		// Main exporter (bulk) - file exporter with small buffer and short interval
		DataExporter: ffclient.DataExporter{
			FlushInterval:    2 * time.Second, // Flush every 2 seconds
			MaxEventInMemory: 3,               // Flush after 3 events
			Exporter: &fileexporter.Exporter{
				Format:    "json",
				OutputDir: "./examples/data_export_log_and_file/variation-events/",
				Filename:  "bulk-main-{{ .Timestamp}}.{{ .Format}}",
			},
		},
		// Multiple additional exporters with different configurations
		DataExporters: []ffclient.DataExporter{
			{
				// Bulk exporter with larger buffer and longer interval
				FlushInterval:    5 * time.Second, // Flush every 5 seconds
				MaxEventInMemory: 5,               // Flush after 5 events
				Exporter: &fileexporter.Exporter{
					Format:    "json",
					OutputDir: "./examples/data_export_log_and_file/variation-events/",
					Filename:  "bulk-secondary-{{ .Timestamp}}.{{ .Format}}",
				},
			},
			{
				// Non-bulk exporter (logs) - should process immediately
				FlushInterval:    1 * time.Second,
				MaxEventInMemory: 1,
				Exporter: &logsexporter.Exporter{
					LogFormat: "IMMEDIATE - user=\"{{ .UserKey}}\", flag=\"{{ .Key}}\", value=\"{{ .Value}}\", variation=\"{{ .Variation}}\"",
				},
			},
			{
				// Another bulk exporter with different settings
				FlushInterval:    3 * time.Second, // Flush every 3 seconds
				MaxEventInMemory: 4,               // Flush after 4 events
				Exporter: &fileexporter.Exporter{
					Format:    "json",
					OutputDir: "./examples/data_export_log_and_file/variation-events/",
					Filename:  "bulk-tertiary-{{ .Timestamp}}.{{ .Format}}",
				},
			},
		},
	})

	if err != nil {
		log.Fatal(err)
	}
	defer ffclient.Close()

	// Create test users
	user1 := ffcontext.NewEvaluationContextBuilder("user1").Build()
	user2 := ffcontext.NewEvaluationContextBuilder("user2").Build()
	user3 := ffcontext.NewEvaluationContextBuilder("user3").Build()

	// Test scenario to trigger different flush conditions

	log.Println("Phase 1: Generate 3 events")
	_, _ = ffclient.BoolVariation("new-admin-access", user1, false)
	_, _ = ffclient.BoolVariation("new-admin-access", user2, false)
	_, _ = ffclient.BoolVariation("new-admin-access", user3, false)

	log.Println("Waiting 1 second")
	time.Sleep(1000 * time.Millisecond)

	log.Println("Phase 2: Generate 2 more events")
	_, _ = ffclient.StringVariation("unknown-flag", user1, "default1")
	_, _ = ffclient.StringVariation("unknown-flag", user2, "default2")

	log.Println("Waiting 2 seconds...")
	time.Sleep(2000 * time.Millisecond)

	log.Println("Phase 3: Generate 2 more events")
	_, _ = ffclient.JSONVariation("json-flag", user1, map[string]interface{}{"test": "value1"})
	_, _ = ffclient.JSONVariation("json-flag", user2, map[string]interface{}{"test": "value2"})

	log.Println("Waiting 3 seconds...")
	time.Sleep(3000 * time.Millisecond)

	log.Println("Phase 4: Generate 1 final event")
	_, _ = ffclient.JSONVariation("json-flag", user3, map[string]interface{}{"test": "value3"})

	log.Println("Waiting 5 seconds...")
	time.Sleep(5000 * time.Millisecond)

	/*
		Expected behavior:

		Phase 1 (3 events):
		- Main exporter: Flushes immediately (hit max 3)
		- Secondary exporter: Holds events (not yet at max 5)
		- Tertiary exporter: Holds events (not yet at max 4)
		- Logger: Processes immediately

		After 1s:
		- No flushes (intervals not reached)

		Phase 2 (+2 events, total 5):
		- Main exporter: Holds 2 events (not yet at max 3)
		- Secondary exporter: Flushes immediately (hit max 5)
		- Tertiary exporter: Flushes immediately at 4 events and then holds 1 event
		- Logger: Processes immediately

		After 2s:
		- Main exporter: Flushes (interval hit)
		- Secondary exporter: Empty after previous flush
		- Tertiary exporter: Holds 1 event (not yet at max 4)

		Phase 3 (+2 events, total 2 since last flush):
		- Main exporter: Holds 2 events (not yet at max 3)
		- Secondary exporter: Holds 2 events (not yet at max 5)
		- Tertiary exporter: Holds 3 events (not yet at max 4)
		- Logger: Processes immediately

		After 3s:
		- Main exporter: Flushes (interval hit)
		- Secondary exporter: Flushes (interval hit)
		- Tertiary exporter: Flushed after only 1 second

		Phase 4 (+1 event, total 3 since last flush):
		- Main exporter: Holds 1 event (not yet at max 3)
		- Secondary exporter: Holds 1 event (not yet at max 5)
		- Tertiary exporter: Holds 1 event (not yet at max 4)
		- Logger: Processes immediately

		After 5s:
		- Main exporter: Flushed after only 3 seconds
		- Secondary exporter: Flushes remaining events (interval hit)
		- Tertiary exporter: Flushed after only 1 second

		Finally:
		- All exporters will flush any remaining events on Close()

		Note:
		- Total we have 8 events
		- Main exporter will generate 4 files containing 3, 2, 2, 1 events respectively
		- Secondary exporter will generate 3 files containing 5, 2, 1 events respectively
		- Tertiary exporter will generate 3 files containing 4, 3, 1 events respectively
		- Logger will generate 8 events in the logs
			(format "IMMEDIATE - user=\"{{ .UserKey}}\", flag=\"{{ .Key}}\", value=\"{{ .Value}}\", variation=\"{{ .Variation}}\"")
	*/
}
