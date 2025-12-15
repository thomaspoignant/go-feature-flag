package ffclient_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	ffclient "github.com/thomaspoignant/go-feature-flag"
	"github.com/thomaspoignant/go-feature-flag/ffcontext"
	"github.com/thomaspoignant/go-feature-flag/retriever/fileretriever"
	"github.com/thomaspoignant/go-feature-flag/testutils/mock"
)

func TestValidTrackingEvent(t *testing.T) {
	exp := mock.TrackingEventExporter{Bulk: false}
	goff, err := ffclient.New(ffclient.Config{
		PollingInterval: 500 * time.Millisecond,
		Retriever:       &fileretriever.Retriever{Path: "./testdata/flag-config.yaml"},
		DataExporters: []ffclient.DataExporter{
			{
				FlushInterval:     100 * time.Millisecond,
				MaxEventInMemory:  100,
				Exporter:          &exp,
				ExporterEventType: ffclient.TrackingEventExporter,
			},
		},
	})
	assert.NoError(t, err)

	goff.Track(
		"my-feature-flag",
		ffcontext.NewEvaluationContextBuilder("1668d845-051d-4dd9-907a-7ebe6aa2c9da").
			AddCustom("admin", true).
			AddCustom("anonymous", true).
			Build(),
		map[string]any{"additional data": "value"},
	)

	assert.Equal(t, 1, len(exp.ExportedEvents))
	assert.Equal(t, "1668d845-051d-4dd9-907a-7ebe6aa2c9da", exp.ExportedEvents[0].UserKey)
	assert.Equal(t, "my-feature-flag", exp.ExportedEvents[0].Key)
	assert.Equal(
		t,
		map[string]any{
			"targetingKey": "1668d845-051d-4dd9-907a-7ebe6aa2c9da",
			"admin":        true,
			"anonymous":    true,
		},
		exp.ExportedEvents[0].EvaluationContext,
	)
	assert.Equal(
		t,
		map[string]any{"additional data": "value"},
		exp.ExportedEvents[0].TrackingDetails,
	)
}
