package ffclient

import (
	"context"
)

type Retriever interface {
	Retrieve(ctx context.Context) ([]byte, error)
}
