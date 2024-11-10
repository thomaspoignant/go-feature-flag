//go:build docker
// +build docker

package azureexporter_test

import (
	"context"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/azurite"
	"github.com/thomaspoignant/go-feature-flag/exporter"
	"github.com/thomaspoignant/go-feature-flag/exporter/azureexporter"
	"github.com/thomaspoignant/go-feature-flag/utils/fflog"
	"log/slog"
	"os"
	"testing"
)

var containerName = "testcontainer"

func TestAzureBlobStorage_Export(t *testing.T) {
	hostname, _ := os.Hostname()
	tests := []struct {
		name         string
		exporter     azureexporter.Exporter
		events       []exporter.FeatureEvent
		wantErr      assert.ErrorAssertionFunc
		wantBlobName string
	}{
		{
			name: "Should insert 1 file in the root of the container",
			exporter: azureexporter.Exporter{
				Container:   containerName,
				AccountName: azurite.AccountName,
				AccountKey:  azurite.AccountKey,
			},
			events: []exporter.FeatureEvent{
				{
					Kind: "feature", ContextKind: "anonymousUser", UserKey: "ABCD", CreationDate: 1617970547, Key: "random-key",
					Variation: "Default", Value: "YO", Default: false,
				},
			},
			wantErr:      assert.NoError,
			wantBlobName: "^flag-variation-" + hostname + "-[0-9]*\\.json$",
		},
		{
			name: "Should insert 1 file with a path in the container",
			exporter: azureexporter.Exporter{
				Path:        "random/path",
				Container:   containerName,
				AccountName: azurite.AccountName,
				AccountKey:  azurite.AccountKey,
			},
			events: []exporter.FeatureEvent{
				{
					Kind: "feature", ContextKind: "anonymousUser", UserKey: "ABCD", CreationDate: 1617970547, Key: "random-key",
					Variation: "Default", Value: "YO", Default: false,
				},
			},
			wantErr:      assert.NoError,
			wantBlobName: "^random/path/flag-variation-" + hostname + "-[0-9]*\\.json$",
		},
		{
			name: "Should insert 1 file in the root of the container as CSV",
			exporter: azureexporter.Exporter{
				Format:      "csv",
				Container:   containerName,
				AccountName: azurite.AccountName,
				AccountKey:  azurite.AccountKey,
			},
			events: []exporter.FeatureEvent{
				{
					Kind: "feature", ContextKind: "anonymousUser", UserKey: "ABCD", CreationDate: 1617970547, Key: "random-key",
					Variation: "Default", Value: "YO", Default: false,
				},
			},
			wantErr:      assert.NoError,
			wantBlobName: "^flag-variation-" + hostname + "-[0-9]*\\.csv",
		},
		{
			name: "Should insert 1 file with a custom filename",
			exporter: azureexporter.Exporter{
				Filename:    "test-json-{{ .Timestamp}}-{{ .Hostname}}.{{ .Format}}",
				Container:   containerName,
				AccountName: azurite.AccountName,
				AccountKey:  azurite.AccountKey,
			},
			events: []exporter.FeatureEvent{
				{
					Kind: "feature", ContextKind: "anonymousUser", UserKey: "ABCD", CreationDate: 1617970547, Key: "random-key",
					Variation: "Default", Value: "YO", Default: false,
				},
			},
			wantErr:      assert.NoError,
			wantBlobName: "^test-json-[0-9]*-" + hostname + "\\.json$",
		},
		{
			name: "Should error with invalid file name",
			exporter: azureexporter.Exporter{
				Filename:    "{{ .InvalidField}}",
				Container:   containerName,
				AccountName: azurite.AccountName,
				AccountKey:  azurite.AccountKey,
			},
			events: []exporter.FeatureEvent{
				{
					Kind: "feature", ContextKind: "anonymousUser", UserKey: "ABCD", CreationDate: 1617970547, Key: "random-key",
					Variation: "Default", Value: "YO", Default: false,
				},
			},
			wantErr: assert.Error,
		},
		{
			name: "Should error with invalid csv formatter",
			exporter: azureexporter.Exporter{
				Format:      "csv",
				CsvTemplate: "{{ .Foo}}",
				Container:   containerName,
				AccountName: azurite.AccountName,
				AccountKey:  azurite.AccountKey,
			},
			events: []exporter.FeatureEvent{
				{
					Kind: "feature", ContextKind: "anonymousUser", UserKey: "ABCD", CreationDate: 1617970547, Key: "random-key",
					Variation: "Default", Value: "YO", Default: false,
				},
			},
			wantErr: assert.Error,
		},
		{
			name: "Should error with empty container",
			exporter: azureexporter.Exporter{
				Container:   "",
				AccountName: azurite.AccountName,
				AccountKey:  azurite.AccountKey,
			},
			events: []exporter.FeatureEvent{
				{
					Kind: "feature", ContextKind: "anonymousUser", UserKey: "ABCD", CreationDate: 1617970547, Key: "random-key",
					Variation: "Default", Value: "YO", Default: false,
				},
			},
			wantErr: assert.Error,
		},
		{
			name:     "Should error with nil container",
			exporter: azureexporter.Exporter{},
			events: []exporter.FeatureEvent{
				{
					Kind: "feature", ContextKind: "anonymousUser", UserKey: "ABCD", CreationDate: 1617970547, Key: "random-key",
					Variation: "Default", Value: "YO", Default: false,
				},
			},
			wantErr: assert.Error,
		},
		{
			name: "Should error if no account name provided",
			exporter: azureexporter.Exporter{
				AccountName: "",
			},
			events: []exporter.FeatureEvent{
				{
					Kind: "feature", ContextKind: "anonymousUser", UserKey: "ABCD", CreationDate: 1617970547, Key: "random-key",
					Variation: "Default", Value: "YO", Default: false,
				},
			},
			wantErr: assert.Error,
		},
		{
			name: "Should error if cred creation fails",
			exporter: azureexporter.Exporter{
				AccountName: "wrong name",
				AccountKey:  azurite.AccountKey,
				Container:   containerName,
			},
			events: []exporter.FeatureEvent{
				{
					Kind: "feature", ContextKind: "anonymousUser", UserKey: "ABCD", CreationDate: 1617970547, Key: "random-key",
					Variation: "Default", Value: "YO", Default: false,
				},
			},
			wantErr: assert.Error,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			container, client := setupTest(t)
			defer tearDown(t, container)
			tt.exporter.ServiceURL = fmt.Sprintf("%s/%s", container.MustServiceURL(context.Background(), azurite.BlobService), azurite.AccountName)
			err := tt.exporter.Export(context.Background(), &fflog.FFLogger{LeveledLogger: slog.Default()}, tt.events)
			tt.wantErr(t, err, "Export() error")
			if err == nil {
				files := make([]string, 0)
				pager := client.NewListBlobsFlatPager(containerName, nil)
				for pager.More() {
					page, err := pager.NextPage(context.Background())
					require.NoError(t, err)
					for _, blob := range page.Segment.BlobItems {
						files = append(files, *blob.Name)
					}
				}
				assert.Len(t, files, 1, "should have one file")
				assert.Regexp(t, tt.wantBlobName, files[0], "filename should match")
			}
		})
	}
}

func TestAzureBlobStorage_IsBulk(t *testing.T) {
	e := azureexporter.Exporter{}
	assert.True(t, e.IsBulk(), "exporter is a bulk exporter")
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

	return azuriteContainer, client
}

func tearDown(t *testing.T, container *azurite.AzuriteContainer) {
	err := testcontainers.TerminateContainer(container)
	require.NoError(t, err)
}
