package mongodbretriever_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thomaspoignant/go-feature-flag/retriever/mongodbretriever"
)

func Test_MongoDBRetriever_Retrieve_NotInitialized(t *testing.T) {
	tests := []struct {
		name      string
		retriever *mongodbretriever.Retriever
		wantErr   string
	}{
		{
			name:      "should return an error when Init was not called",
			retriever: &mongodbretriever.Retriever{},
			wantErr:   "database connection is not initialized",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.retriever.Retrieve(context.Background())
			assert.Nil(t, got)
			assert.ErrorContains(t, err, tt.wantErr)
		})
	}
}
