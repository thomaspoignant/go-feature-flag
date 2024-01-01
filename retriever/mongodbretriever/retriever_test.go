package mongodbretriever

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"

	"github.com/stretchr/testify/assert"
)

var jsonString = `[{"test-flag":{"variations":{"true_var":true,"false_var":false},"targeting":[{"query":"key eq \"random-key\"","percentage":{"true_var":0,"false_var":100}}],"defaultRule":{"variation":"false_var"},"trackEvents":false},"test-flag2":{"variations":{"true_var":true,"false_var":false},"targeting":[{"query":"key eq \"not-a-key\"","percentage":{"true_var":0,"false_var":100}}],"defaultRule":{"variation":"false_var"},"trackEvents":false}}]`

func mapToBsonD(inputMap []map[string]interface{}) (bson.D, error) {
	var bsonD bson.D

	for _, v := range inputMap {
		for key, value := range v {
			bsonD = append(bsonD, bson.E{Key: key, Value: value})
		}
	}

	return bsonD, nil
}

var mockFunc = func (t *mtest.T) {

	var resultMap []map[string]interface{}

	// Unmarshal the JSON string into the map
	err := json.Unmarshal([]byte(jsonString), &resultMap)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// Convert the map to a BSON document (bson.D)
	bsonDocument, err := mapToBsonD(resultMap)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	t.AddMockResponses(mtest.CreateCursorResponse(1, fmt.Sprintf("#{mdb.Database}.#{mdb.Collection}"),
		mtest.FirstBatch,
		bsonDocument,
	))
}

func Test_MongoDBRetriever_Retrieve(t *testing.T) {
	mtDB := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	// type fields struct {
	// 	dbConnection *mongo.Database
	// }
	tests := []struct {
		name    string
		want    []byte
		mocker  *func(t *mtest.T)
		wantErr bool
	}{
		{
			name:    "File exists",
			mocker:  &mockFunc,
			want:    []byte(jsonString),
			wantErr: false,
		},
		// {
		// 	name: "File does not exists",
		// 	fields: fields{
		// 		path: "./testdata/test-not-exist.yaml",
		// 	},
		// 	want:    nil,
		// 	wantErr: true,
		// },
	}
	for _, tt := range tests {
		mtDB.Run(tt.name, func(t *mtest.T) {
			mdb := Retriever{
				Uri:          "mongouri",
				Collection:   "collection",
				Database:     "database",
				dbConnection: t.DB,
			}

			if tt.mocker != nil {
				(*tt.mocker)(t)
			}

			got, err := mdb.Retrieve(context.Background())

			if tt.wantErr {
				assert.Error(t, err, "Retrieve() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, string(tt.want), string(got))
		})
	}
}
