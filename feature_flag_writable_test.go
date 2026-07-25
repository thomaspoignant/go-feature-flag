package ffclient_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	ffclient "github.com/thomaspoignant/go-feature-flag"
	"github.com/thomaspoignant/go-feature-flag/retriever"
	"github.com/thomaspoignant/go-feature-flag/retriever/fileretriever"
	"github.com/thomaspoignant/go-feature-flag/testutils/mock/mockretriever"
)

func TestGoFeatureFlag_GetWritableRetriever(t *testing.T) {
	tests := []struct {
		name       string
		config     ffclient.Config
		wantOK     bool
		wantSource string
	}{
		{
			name: "single writable retriever",
			config: ffclient.Config{
				PollingInterval: 60 * time.Second,
				Retriever:       mockretriever.NewWritableRetriever("pg"),
			},
			wantOK:     true,
			wantSource: "mock:pg",
		},
		{
			name: "single non-writable retriever",
			config: ffclient.Config{
				PollingInterval: 60 * time.Second,
				Retriever:       &fileretriever.Retriever{Path: "testdata/flag-config.yaml"},
			},
			wantOK: false,
		},
		{
			name: "multiple retrievers, one writable",
			config: ffclient.Config{
				PollingInterval: 60 * time.Second,
				Retriever:       mockretriever.NewWritableRetriever("pg"),
				Retrievers: []retriever.Retriever{
					&fileretriever.Retriever{Path: "testdata/flag-config.yaml"},
				},
			},
			wantOK: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := ffclient.New(tt.config)
			assert.NoError(t, err)
			defer client.Close()

			got, ok := client.GetWritableRetriever()
			assert.Equal(t, tt.wantOK, ok)
			if tt.wantOK {
				assert.Equal(t, tt.wantSource, got.Source())
			} else {
				assert.Nil(t, got)
			}
		})
	}
}

func TestGoFeatureFlag_GetWritableRetriever_Offline(t *testing.T) {
	client, err := ffclient.New(ffclient.Config{Offline: true})
	assert.NoError(t, err)
	defer client.Close()

	got, ok := client.GetWritableRetriever()
	assert.False(t, ok)
	assert.Nil(t, got)
}
