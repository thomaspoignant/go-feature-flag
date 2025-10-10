package gcstorageretriever

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thomaspoignant/go-feature-flag/testutils"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
)

func TestRetriever_Retrieve(t *testing.T) {
	ctx := context.Background()

	bucketFiles := map[string]string{
		"testdata/flag-config.yaml": "flag-config.yaml",
	}

	bucketName := "flags"

	type fields struct {
		Bucket string
		Object string
	}
	type storageConfig struct {
		bucket       string
		files        map[string]string
		updatedFiles map[string]string
	}

	tests := []struct {
		name            string
		fields          fields
		storage         storageConfig
		want            string
		wantWhenUpdated string
		wantFromCache   bool
		wantUpdated     bool
		wantErr         bool
	}{
		{
			name: "first bootstrap of cloud storage retriever",
			storage: storageConfig{
				bucket: bucketName,
				files:  bucketFiles,
			},
			fields: fields{
				Bucket: bucketName,
				Object: "flag-config.yaml",
			},
			want: "testdata/flag-config.yaml",
		},
		{
			name: "get content from cache",
			storage: storageConfig{
				bucket: bucketName,
				files:  bucketFiles,
			},
			fields: fields{
				Bucket: bucketName,
				Object: "flag-config.yaml",
			},
			want:          "testdata/flag-config.yaml",
			wantFromCache: true,
		},
		{
			name: "retriever update file when it changes in bucket",
			storage: storageConfig{
				bucket: bucketName,
				files:  bucketFiles,
				updatedFiles: map[string]string{
					"testdata/flag-config-updated.yaml": "flag-config.yaml",
				},
			},
			fields: fields{
				Bucket: bucketName,
				Object: "flag-config.yaml",
			},
			want:            "testdata/flag-config.yaml",
			wantWhenUpdated: "testdata/flag-config-updated.yaml",
			wantUpdated:     true,
		},
		{
			name: "object not found in bucket",
			storage: storageConfig{
				bucket: bucketName,
				files:  bucketFiles,
			},
			fields: fields{
				Bucket: bucketName,
				Object: "fake-flag-config.yaml",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockedStorage := testutils.NewMockedGCS(t)
			mockedStorage.WithFiles(t, tt.storage.bucket, tt.storage.files)

			retriever := &Retriever{
				Bucket: tt.fields.Bucket,
				Object: tt.fields.Object,
				Options: []option.ClientOption{
					option.WithCredentials(&google.Credentials{}),
					option.WithHTTPClient(mockedStorage.Server.HTTPClient()),
				},
			}

			assertRetrieve := func(want string) {
				gotContent, err := retriever.Retrieve(ctx)
				assert.Equal(
					t,
					tt.wantErr,
					err != nil,
					"retrieve() error = %v, wantErr %v",
					err,
					tt.wantErr,
				)

				if err == nil {
					wantContent, err := os.ReadFile(want)
					assert.NoError(t, err)

					assert.Equal(
						t,
						wantContent,
						gotContent,
						"retrieve() got = %v, want %v",
						gotContent,
						wantContent,
					)
				}
			}

			assertRetrieve(tt.want)

			if tt.wantFromCache {
				assertRetrieve(tt.want)
			}

			if tt.wantUpdated {
				mockedStorage.WithFiles(t, tt.storage.bucket, tt.storage.updatedFiles)
				assertRetrieve(tt.wantWhenUpdated)
			}
		})
	}
}
