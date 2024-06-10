package retriever_test

import (
	"github.com/stretchr/testify/assert"
	ffclient "github.com/thomaspoignant/go-feature-flag"
	"github.com/thomaspoignant/go-feature-flag/retriever"
	"github.com/thomaspoignant/go-feature-flag/utils/fflog"
	"golang.org/x/net/context"
	"log"
	"testing"
	"time"
)

func TestMixLegacyTypesOfRetrievers(t *testing.T) {
	sr := &simpleRetriever{}
	ilr := &initializableRetrieverLegacy{}
	il := &initializableRetriever{}
	goff, err := ffclient.New(ffclient.Config{
		PollingInterval: 10 * time.Second,
		Retrievers: []retriever.Retriever{
			sr,
			ilr,
			il,
		},
	})
	assert.NoError(t, err)
	goff.Close()

	assert.True(t, sr.retrieveCalled)

	assert.True(t, ilr.initCalled)
	assert.True(t, ilr.statusCalled)
	assert.True(t, ilr.retrieveCalled)
	assert.True(t, ilr.shutdownCalled)

	assert.True(t, il.initCalled)
	assert.True(t, il.statusCalled)
	assert.True(t, il.retrieveCalled)
	assert.True(t, il.shutdownCalled)
}

type simpleRetriever struct {
	retrieveCalled bool
}

func (s *simpleRetriever) Retrieve(_ context.Context) ([]byte, error) {
	s.retrieveCalled = true
	return []byte{}, nil
}

type initializableRetrieverLegacy struct {
	retrieveCalled bool
	initCalled     bool
	shutdownCalled bool
	statusCalled   bool
}

func (i *initializableRetrieverLegacy) Retrieve(_ context.Context) ([]byte, error) {
	i.retrieveCalled = true
	return []byte{}, nil
}

func (i *initializableRetrieverLegacy) Init(_ context.Context, _ *log.Logger) error {
	i.initCalled = true
	return nil
}

func (i *initializableRetrieverLegacy) Shutdown(_ context.Context) error {
	i.shutdownCalled = true
	return nil
}

func (i *initializableRetrieverLegacy) Status() retriever.Status {
	i.statusCalled = true
	return retriever.RetrieverReady
}

type initializableRetriever struct {
	retrieveCalled bool
	initCalled     bool
	shutdownCalled bool
	statusCalled   bool
}

func (i *initializableRetriever) Retrieve(_ context.Context) ([]byte, error) {
	i.retrieveCalled = true
	return []byte{}, nil
}

func (i *initializableRetriever) Init(_ context.Context, _ *fflog.FFLogger) error {
	i.initCalled = true
	return nil
}

func (i *initializableRetriever) Shutdown(_ context.Context) error {
	i.shutdownCalled = true
	return nil
}

func (i *initializableRetriever) Status() retriever.Status {
	i.statusCalled = true
	return retriever.RetrieverReady
}
