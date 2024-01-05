package mongodbretriever

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thomaspoignant/go-feature-flag/testutils"
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"
)

func Test_MongoDBRetriever_Retrieve(t *testing.T) {
	mtDB := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	tests := []struct {
		name    string
		want    []byte
		mocker  *func(t *mtest.T)
		wantErr bool
	}{
		{
			name:    "Returns well formed flag definition document",
			mocker:  &testutils.MockSuccessFind,
			want:    []byte(testutils.QueryResult),
			wantErr: false,
		},
		{
			name:    "Flag definition document does not have 'flag' key/value",
			mocker:  &testutils.MockNoFlags,
			want:    nil,
			wantErr: true,
		},
		{
			name:    "Flag definition document 'flag' key does not have 'string' value",
			mocker:  &testutils.MockFlagNotStrResult,
			want:    nil,
			wantErr: true,
		},
		{
			name:    "No flags found on DB",
			mocker:  &testutils.MockNoFlags,
			want:    nil,
			wantErr: true,
		},
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

			var gotUnm, wantUn interface{}
			json.Unmarshal(tt.want, &wantUn)
			json.Unmarshal(got, &gotUnm)

			assert.Equal(t, wantUn, gotUnm)
		})
	}
}
