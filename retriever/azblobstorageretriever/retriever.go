package azblobretriever

import (
	"context"
	"fmt"
	"io"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/thomaspoignant/go-feature-flag/retriever"
	"github.com/thomaspoignant/go-feature-flag/utils/fflog"
)

// Retriever is the interface to fetch the flags from Azure Blob Storage.
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

	// client is a pointer to an Azure Blob Storage client.
	// It provides access to Azure Blob Storage services for operations like
	// creating, reading, updating, and deleting blobs.
	client *azblob.Client
	status retriever.Status
}

// Init is initializing the retriever to start fetching the flags configuration.
func (r *Retriever) Init(_ context.Context, _ *fflog.FFLogger) error {
	if r.AccountName == "" {
		return fmt.Errorf(
			"unable to connect to Azure Blob Storage, \"AccountName\" cannot be empty",
		)
	}

	url := r.ServiceURL
	if url == "" {
		url = fmt.Sprintf("https://%s.blob.core.windows.net/", r.AccountName)
	}

	var client *azblob.Client
	var err error

	if r.AccountKey == "" {
		var cred *azidentity.DefaultAzureCredential
		cred, err = azidentity.NewDefaultAzureCredential(nil)
		if err != nil {
			r.status = retriever.RetrieverError
			return err
		}
		client, err = azblob.NewClient(url, cred, nil)
	} else {
		var cred *azblob.SharedKeyCredential
		cred, err = azblob.NewSharedKeyCredential(r.AccountName, r.AccountKey)
		if err != nil {
			r.status = retriever.RetrieverError
			return err
		}
		client, err = azblob.NewClientWithSharedKeyCredential(url, cred, nil)
	}

	if err != nil {
		r.status = retriever.RetrieverError
		return err
	}

	r.client = client
	r.status = retriever.RetrieverReady
	return nil
}

// Shutdown gracefully shutdown the provider and set the status as not ready.
func (r *Retriever) Shutdown(_ context.Context) error {
	r.client = nil
	r.status = retriever.RetrieverNotReady
	return nil
}

// Status is the function returning the internal state of the retriever.
func (r *Retriever) Status() retriever.Status {
	return r.status
}

// Retrieve is the function in charge of fetching the flag configuration.
func (r *Retriever) Retrieve(ctx context.Context) ([]byte, error) {
	if r.client == nil {
		r.status = retriever.RetrieverError
		return nil, fmt.Errorf("client is not initialized")
	}

	if r.Object == "" || r.Container == "" {
		return nil, fmt.Errorf(
			"missing mandatory information object=%s, repositorySlug=%s",
			r.Object,
			r.Container,
		)
	}

	fileStream, err := r.client.DownloadStream(ctx, r.Container, r.Object, nil)
	if err != nil {
		return nil, err
	}

	retryReader := fileStream.NewRetryReader(ctx, nil)
	defer func() { _ = retryReader.Close() }()

	body, err := io.ReadAll(retryReader)
	if err != nil {
		return nil,
			fmt.Errorf(
				"unable to read from Azure Blob Storage Object %s in Container %s, error: %s",
				r.Object,
				r.Container,
				err,
			)
	}

	return body, nil
}
