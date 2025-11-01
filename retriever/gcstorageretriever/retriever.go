package gcstorageretriever

import (
	"bytes"
	"context"
	"crypto/md5" //nolint: gosec
	"fmt"
	"io"

	"cloud.google.com/go/storage"
	"google.golang.org/api/option"
)

// Retriever is a configuration struct for a Google Cloud Storage retriever.
type Retriever struct {
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
	obj *storage.ObjectHandle
}

func (retriever *Retriever) SetOptions(options []option.ClientOption) {
	retriever.Options = options
}

// Retrieve is the function in charge of fetching the flag configuration.
func (retriever *Retriever) Retrieve(ctx context.Context) (content []byte, err error) {
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
	reader, err := retriever.obj.NewReader(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = reader.Close() }()

	// Read all contents from the Reader.
	content, err = io.ReadAll(reader)
	if err != nil {
		return nil,
			fmt.Errorf(
				"unable to read from GCP Object %s in Bucket %s, error: %s",
				retriever.Bucket,
				retriever.Object,
				err,
			)
	}

	// Update Cache along with its hash.
	retriever.cache = content
	md5Hash := md5.Sum(content) //nolint: gosec
	retriever.md5 = md5Hash[:]

	return content, nil
}
