package ffclient

import (
	"cloud.google.com/go/storage"
	"context"
	"fmt"
	"google.golang.org/api/option"
	"io"
	"io/ioutil"
)

// GCStorageRetriever is a configuration struct for a Google Cloud Storage retriever.
type GCStorageRetriever struct {
	// Bucket is the name of your Google Cloud Storage Bucket.
	Bucket string

	// Object is the name of your file in your bucket.
	Object string

	// Options are Google Cloud Api options needed for downloading
	// your feature flag configuration file.
	Options []option.ClientOption

	// Internal field used to read the flag file from.
	rC io.ReadCloser
}

func (retriever *GCStorageRetriever) Retrieve(ctx context.Context) ([]byte, error) {
	if retriever.rC == nil {
		// Create GC Storage Client.
		client, err := storage.NewClient(ctx, retriever.Options...)
		if err != nil {
			return nil, err
		}

		// Downloads the file and creates Reader to read from it.
		retriever.rC, err = client.Bucket(retriever.Bucket).Object(retriever.Object).NewReader(ctx)
		if err != nil {
			return nil, err
		}
	}

	defer retriever.rC.Close()

	// Read all contents from the Reader.
	content, err := ioutil.ReadAll(retriever.rC)
	if err != nil {
		return nil, fmt.Errorf(
			"unable to read from GCP Object %s in Bucket %s", retriever.Bucket, retriever.Object,
		)
	}

	return content, nil
}
