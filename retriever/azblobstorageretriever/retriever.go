package azblobretriever

import (
	"context"
	"fmt"
	"io"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
)

type Retriever struct {
	// Container is the name of your Azure Blob Storage Container.
	Container string

	// Storage Account Name and Key
	AccountName string
	AccountKey  string

	// ServiceURL is the URL of the storage account e.g. https://<account>.blob.core.windows.net/
	// It can be overridden by the user to use a custom URL.
	// Default: https://<account>.blob.core.windows.net/
	ServiceURL string

	// Object is the name of your file in your container.
	Object string
}

func (f *Retriever) initializeAzureClient() (*azblob.Client, error) {
	url := fmt.Sprintf("https://%s.blob.core.windows.net/", f.AccountName)
	if f.ServiceURL != "" {
		url = f.ServiceURL
	}
	if f.AccountKey == "" {
		cred, err := azidentity.NewDefaultAzureCredential(nil)
		if err != nil {
			return nil, err
		}
		return azblob.NewClient(url, cred, nil)
	}
	cred, err := azblob.NewSharedKeyCredential(f.AccountName, f.AccountKey)
	if err != nil {
		return nil, err
	}
	return azblob.NewClientWithSharedKeyCredential(url, cred, nil)
}

func (r *Retriever) Retrieve(ctx context.Context) ([]byte, error) {
	if r.Object == "" || r.Container == "" {
		return nil, fmt.Errorf("missing mandatory information filePath=%s, repositorySlug=%s", r.Object, r.Container)
	}

	client, err := r.initializeAzureClient()
	if err != nil {
		return nil, err
	}

	fileStream, err := client.DownloadStream(ctx, r.Container, r.Object, nil)
	if err != nil {
		return nil, err
	}

	retryReader := fileStream.NewRetryReader(ctx, nil)
	defer func() { _ = retryReader.Close() }()

	body, err := io.ReadAll(retryReader)
	if err != nil {
		return nil,
			fmt.Errorf("unable to read from Azure Blob Storage Object %s in Container %s, error: %s", r.Container, r.Object, err)
	}

	return body, nil
}
