//go:build docker

package mongodbretriever_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/thomaspoignant/go-feature-flag/retriever"
	"github.com/thomaspoignant/go-feature-flag/retriever/mongodbretriever"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go/modules/mongodb"
	"github.com/thomaspoignant/go-feature-flag/testutils"
	"github.com/thomaspoignant/go-feature-flag/utils/fflog"
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"
)

func Test_MongoDBRetriever_Retrieve(t *testing.T) {
	mtDB := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	tests := []struct {
		name    string
		want    []byte
		data    string
		wantErr bool
	}{
		{
			name:    "Returns well formed flag definition document",
			data:    testutils.MongoFindResultString,
			want:    []byte(testutils.QueryResult),
			wantErr: false,
		},
		{
			name:    "One of the Flag definition document does not have 'flag' key/value (ignore this document)",
			data:    testutils.MongoMissingFlagKey,
			want:    []byte(testutils.MissingFlagKeyResult),
			wantErr: false,
		},
		{
			name:    "Flag definition document 'flag' key does not have 'string' value (ignore this document)",
			data:    testutils.MongoFindResultFlagNoStr,
			want:    []byte(testutils.FlagKeyNotStringResult),
			wantErr: false,
		},
		{
			name:    "No flags found on DB",
			want:    []byte("{}"),
			wantErr: true,
		},
	}
	for _, tt := range tests {
		mtDB.Run(tt.name, func(t *mtest.T) {
			mongodbContainer, err := mongodb.Run(context.TODO(), "mongo:6")
			require.NoError(t, err)
			defer func() {
				err := mongodbContainer.Terminate(context.TODO())
				require.NoError(t, err)
			}()

			uri, err := mongodbContainer.ConnectionString(context.Background())

			if tt.data != "" {
				// insert data
				client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(uri))
				coll := client.Database("database").Collection("collection")
				var documents []bson.M
				err = json.Unmarshal([]byte(tt.data), &documents)
				require.NoError(t, err)

				for _, doc := range documents {
					_, err := coll.InsertOne(context.TODO(), doc)
					require.NoError(t, err)
				}
			}

			// retriever
			mdb := mongodbretriever.Retriever{
				URI:        uri,
				Collection: "collection",
				Database:   "database",
			}
			assert.Equal(t, retriever.RetrieverNotReady, mdb.Status())
			err = mdb.Init(context.TODO(), &fflog.FFLogger{})
			assert.NoError(t, err)
			defer func() { _ = mdb.Shutdown(context.TODO()) }()
			assert.Equal(t, retriever.RetrieverReady, mdb.Status())

			got, err := mdb.Retrieve(context.Background())
			if tt.want == nil {
				assert.Nil(t, got)
			} else {
				modifiedGot, err := removeIDFromJSON(string(got))
				require.NoError(t, err)
				assert.JSONEq(t, string(tt.want), modifiedGot)
			}
		})
	}
}

func Test_MongoDBRetriever_InvalidURI(t *testing.T) {
	mdb := mongodbretriever.Retriever{
		URI:        "invalidURI",
		Collection: "collection",
		Database:   "database",
	}
	assert.Equal(t, retriever.RetrieverNotReady, mdb.Status())
	err := mdb.Init(context.TODO(), &fflog.FFLogger{})
	assert.Error(t, err)
	assert.Equal(t, retriever.RetrieverError, mdb.Status())
}

func removeIDFromJSON(jsonStr string) (string, error) {
	var data any
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		return "", err
	}

	removeIDFields(data)

	modifiedJSON, err := json.Marshal(data)
	if err != nil {
		return "", err
	}

	return string(modifiedJSON), nil
}

func removeIDFields(data any) {
	switch v := data.(type) {
	case map[string]any:
		delete(v, "_id")
		for _, value := range v {
			removeIDFields(value)
		}
	case []any:
		for _, item := range v {
			removeIDFields(item)
		}
	}
}
