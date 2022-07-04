package ffclient

import (
	"bytes"
	"context"
	"crypto/md5" //nolint: gosec
	"fmt"
	"io"
	"io/ioutil"

	"cloud.google.com/go/storage"

	"google.golang.org/api/option"
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

	// Internal field used to cache the file data.
	cache []byte

	// Internal MD5 hash of cache
	md5 []byte

	// Internal field used to fetch metadata of the file.
	obj object

	// Internal field used to read from the file.
	rC io.ReadCloser
}

type object interface {
	Attrs(ctx context.Context) (*storage.ObjectAttrs, error)
	NewReader(ctx context.Context) (*storage.Reader, error)
}

func (retriever *GCStorageRetriever) Retrieve(ctx context.Context) (content []byte, err error) {
	if retriever.obj == nil {
		// Create GC Storage Client.
		client, err := storage.NewClient(ctx, retriever.Options...)
		if err != nil {
			return nil, err
		}

		// Construct Object.
		retriever.obj = client.Bucket(retriever.Bucket).Object(retriever.Object)
	}

	// Fetch the metadata of the remote file.
	attrs, err := retriever.obj.Attrs(ctx)
	if err != nil {
		return nil, err
	}

	// When local and remote hashes match, return cached data.
	if bytes.Equal(attrs.MD5, retriever.md5) {
		return retriever.cache, nil
	}

	// When cache is outdated download file and create Reader to read from it.
	if retriever.rC == nil {
		retriever.rC, err = retriever.obj.NewReader(ctx)
		if err != nil {
			return nil, err
		}
	}
	defer retriever.rC.Close()

	// Read all contents from the Reader.
	content, err = ioutil.ReadAll(retriever.rC)
	if err != nil {
		return nil, fmt.Errorf(
			"unable to read from GCP Object %s in Bucket %s, error: %s", retriever.Bucket, retriever.Object, err,
		)
	}

	// Update Cache along with its hash.
	retriever.cache = content
	md5Hash := md5.Sum(content) //nolint: gosec
	retriever.md5 = md5Hash[:]

	return content, nil
}
