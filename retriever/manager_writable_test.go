package retriever_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thomaspoignant/go-feature-flag/internal/cache"
	"github.com/thomaspoignant/go-feature-flag/internal/notification"
	"github.com/thomaspoignant/go-feature-flag/notifier"
	"github.com/thomaspoignant/go-feature-flag/retriever"
	"github.com/thomaspoignant/go-feature-flag/testutils/mock/mockretriever"
	"github.com/thomaspoignant/go-feature-flag/utils/fflog"
)

func TestManager_GetWritableRetrievers(t *testing.T) {
	tests := []struct {
		name       string
		retrievers []retriever.Retriever
		wantCount  int
	}{
		{
			name:       "no retriever configured",
			retrievers: []retriever.Retriever{},
			wantCount:  0,
		},
		{
			name: "single non-writable retriever",
			retrievers: []retriever.Retriever{
				mockretriever.NewSimpleRetriever("simple"),
			},
			wantCount: 0,
		},
		{
			name: "single writable retriever",
			retrievers: []retriever.Retriever{
				mockretriever.NewWritableRetriever("pg"),
			},
			wantCount: 1,
		},
		{
			name: "mixed writable and non-writable retrievers",
			retrievers: []retriever.Retriever{
				mockretriever.NewSimpleRetriever("simple"),
				mockretriever.NewWritableRetriever("pg"),
			},
			wantCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := fflog.FFLogger{}
			cacheManager := cache.New(notification.NewService([]notifier.Notifier{}), "", &logger)
			config := retriever.ManagerConfig{FileFormat: "json", PollingInterval: 0}
			manager := retriever.NewManager(config, tt.retrievers, cacheManager, &logger)

			assert.Equal(t, len(tt.retrievers), manager.RetrieverCount())
			assert.Len(t, manager.GetWritableRetrievers(), tt.wantCount)
		})
	}
}
