package ffclient

import (
	"cloud.google.com/go/storage"
	"context"
	"fmt"
	"google.golang.org/api/option"
	"io"
	"io/ioutil"
)

type GCPRetriever struct {
	Option option.ClientOption
	Bucket string
	Object string
	rC     io.ReadCloser
}

func (retriever *GCPRetriever) Retrieve(ctx context.Context) ([]byte, error) {
	if retriever.rC == nil {
		client, err := storage.NewClient(ctx, retriever.Option)
		if err != nil {
			return nil, err
		}

		retriever.rC, err = client.Bucket(retriever.Bucket).Object(retriever.Object).NewReader(ctx)
		if err != nil {
			return nil, err
		}
	}

	defer retriever.rC.Close()

	content, err := ioutil.ReadAll(retriever.rC)
	if err != nil {
		return nil, fmt.Errorf(
			"unable to read from GCP Object %s in Bucket %s", retriever.Bucket, retriever.Object,
		)
	}

	return content, nil
}
