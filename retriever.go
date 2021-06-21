package ffclient

import (
	"context"
)

// Retriever is the interface to create a Retriever to load you flags.
type Retriever interface {
	// Retrieve function is supposed to load the file and to return a []byte of your flag configuration file.
	// If you want to specify the format of the file, you can use the ffclient.Config.FileFormat option to
	// specify if it is a YAML, JSON or TOML file.
	Retrieve(ctx context.Context) ([]byte, error)
}
