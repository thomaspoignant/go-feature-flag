//go:build docker
// +build docker

package azblobretriever_test

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/azurite"
)

var containerName = "testcontainer"

func TestAzureBlobStorageRetriever(t *testing.T) {
	type fields struct {
		container   string
		accountName string
		accountKey  string
		object      string
		serviceURL  string
	}
	tests := []struct {
		name    string
		fields  fields
		want    string
		wantErr bool
		context context.Context
	}{
		{
			name: "File on Container",
			fields: fields{
				container:   containerName,
				accountName: azurite.AccountName,
				accountKey:  azurite.AccountKey,
				object:      "flag-config.yaml",
			},
			want:    "./testdata/flag-config.yaml",
			wantErr: false,
		},
		{
			name: "File on Container with Context",
			fields: fields{
				container:   containerName,
				accountName: azurite.AccountName,
				accountKey:  azurite.AccountKey,
				object:      "flag-config.yaml",
			},
			want:    "./testdata/flag-config.yaml",
			context: context.Background(),
			wantErr: false,
		},
		{
			name: "File not on Container",
			fields: fields{
				container:   containerName,
				accountName: azurite.AccountName,
				accountKey:  azurite.AccountKey,
				object:      "feature-config.csv",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			container, _ := setupTest(t)
			defer tearDown(t, container)
			serviceURL := fmt.Sprintf("%s/%s", container.MustServiceURL(context.Background(), azurite.BlobService), azurite.AccountName)
			r := &Retriever{
				Container:   tt.fields.container,
				AccountName: tt.fields.accountName,
				AccountKey:  tt.fields.accountKey,
				ServiceURL:  serviceURL,
				Object:      tt.fields.object,
			}
			err := r.Init(context.Background(), nil)
			assert.NoError(t, err)
			defer func() {
				err := r.Shutdown(context.Background())
				assert.NoError(t, err)
			}()
			got, err := r.Retrieve(tt.context)
			assert.Equal(t, tt.wantErr, err != nil, "Retrieve() error = %v, wantErr %v", err, tt.wantErr)
			if err == nil {
				want, err := os.ReadFile(tt.want)
				assert.NoError(t, err)
				assert.Equal(t, string(want), string(got), "Retrieve() got = %v, want %v", string(want), string(got))
			}
		})
	}
}

func setupTest(t *testing.T) (*azurite.AzuriteContainer, *azblob.Client) {
	ctx := context.Background()
	azuriteContainer, err := azurite.Run(
		ctx,
		"mcr.microsoft.com/azure-storage/azurite:3.33.0",
	)
	require.NoError(t, err)

	cred, err := azblob.NewSharedKeyCredential(azurite.AccountName, azurite.AccountKey)
	require.NoError(t, err)

	blobServiceURL := fmt.Sprintf("%s/%s", azuriteContainer.MustServiceURL(ctx, azurite.BlobService), azurite.AccountName)
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

func tearDown(t *testing.T, container *azurite.AzuriteContainer) {
	err := testcontainers.TerminateContainer(container)
	require.NoError(t, err)
}
