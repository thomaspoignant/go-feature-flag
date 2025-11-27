//go:build docker

package azblobretriever_test

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/azure/azurite"
	azblobretriever "github.com/thomaspoignant/go-feature-flag/retriever/azblobstorageretriever"
	"github.com/thomaspoignant/go-feature-flag/utils/fflog"
)

var containerName = "testcontainer"

func TestAzureBlobStorageRetriever(t *testing.T) {
	tests := []struct {
		name      string
		want      string
		wantErr   bool
		context   context.Context
		retriever azblobretriever.Retriever
	}{
		{
			name:    "File on Container",
			want:    "./testdata/flag-config.yaml",
			wantErr: false,
			retriever: azblobretriever.Retriever{
				Container:   containerName,
				AccountName: azurite.AccountName,
				AccountKey:  azurite.AccountKey,
				Object:      "flag-config.yaml",
			},
		},
		{
			name: "File on Container with Context",
			retriever: azblobretriever.Retriever{
				Container:   containerName,
				AccountName: azurite.AccountName,
				AccountKey:  azurite.AccountKey,
				Object:      "flag-config.yaml",
			},
			want:    "./testdata/flag-config.yaml",
			context: context.Background(),
			wantErr: false,
		},
		{
			name: "File not on Container",
			retriever: azblobretriever.Retriever{
				Container:   containerName,
				AccountName: azurite.AccountName,
				AccountKey:  azurite.AccountKey,
				Object:      "feature-config.csv",
			},
			wantErr: true,
		},
		{
			name: "Should Err on Empty container",
			retriever: azblobretriever.Retriever{
				AccountName: azurite.AccountName,
				AccountKey:  azurite.AccountKey,
				Object:      "feature-config.csv",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			container, _ := setupTest(t)
			defer tearDown(t, container)
			blobURL, err := container.BlobServiceURL(context.Background())
			require.NoError(t, err)

			tt.retriever.ServiceURL = fmt.Sprintf("%s/%s", blobURL, azurite.AccountName)
			err = tt.retriever.Init(context.Background(), &fflog.FFLogger{LeveledLogger: slog.Default()})
			assert.NoError(t, err)
			defer func() {
				err := tt.retriever.Shutdown(context.Background())
				assert.NoError(t, err)
			}()
			got, err := tt.retriever.Retrieve(context.Background())
			assert.Equal(t, tt.wantErr, err != nil, "retrieve() error = %v, wantErr %v", err, tt.wantErr)
			if err == nil {
				want, err := os.ReadFile(tt.want)
				assert.NoError(t, err)
				assert.Equal(t, string(want), string(got), "retrieve() got = %v, want %v", string(want), string(got))
			}
		})
	}
}

func TestInit(t *testing.T) {
	t.Run("Should error when no account name", func(t *testing.T) {
		retriever := azblobretriever.Retriever{
			Container:  containerName,
			AccountKey: azurite.AccountKey,
			Object:     "flag-config.yaml",
		}
		err := retriever.Init(context.Background(), &fflog.FFLogger{LeveledLogger: slog.Default()})
		assert.Error(t, err)
	})

	t.Run("Should error when calling retrieve without init", func(t *testing.T) {
		retriever := azblobretriever.Retriever{
			Container:   containerName,
			AccountName: azurite.AccountName,
			AccountKey:  azurite.AccountKey,
			Object:      "flag-config.yaml",
		}
		_, err := retriever.Retrieve(context.Background())
		assert.Error(t, err)
	})

}

func setupTest(t *testing.T) (*azurite.Container, *azblob.Client) {
	ctx := context.Background()
	azuriteContainer, err := azurite.Run(
		ctx,
		"mcr.microsoft.com/azure-storage/azurite:3.35.0",
	)
	require.NoError(t, err)

	cred, err := azblob.NewSharedKeyCredential(azurite.AccountName, azurite.AccountKey)
	require.NoError(t, err)

	blobURL, err := azuriteContainer.BlobServiceURL(ctx)
	require.NoError(t, err)

	blobServiceURL := fmt.Sprintf("%s/%s", blobURL, azurite.AccountName)
	client, err := azblob.NewClientWithSharedKeyCredential(blobServiceURL, cred, nil)
	require.NoError(t, err)

	_, err = client.CreateContainer(context.TODO(), containerName, nil)
	require.NoError(t, err)

	blobName := "flag-config.yaml"
	blob, err := os.ReadFile(fmt.Sprintf("./testdata/%s", blobName))
	require.NoError(t, err)

	_, err = client.UploadStream(context.TODO(), containerName, blobName, strings.NewReader(string(blob)), nil)
	require.NoError(t, err)

	return azuriteContainer, client
}

func tearDown(t *testing.T, container *azurite.Container) {
	err := testcontainers.TerminateContainer(container)
	require.NoError(t, err)
}
