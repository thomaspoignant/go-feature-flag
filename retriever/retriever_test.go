package retriever_test

import (
	"log"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	ffclient "github.com/thomaspoignant/go-feature-flag"
	"github.com/thomaspoignant/go-feature-flag/retriever"
	"github.com/thomaspoignant/go-feature-flag/utils/fflog"
	"golang.org/x/net/context"
)

func TestMixLegacyTypesOfRetrievers(t *testing.T) {
	sr := &simpleRetriever{}
	ilr := &initializableRetrieverLegacy{}
	il := &initializableRetriever{}
	ilf := &initializableRetrieverWithFlagset{}
	goff, err := ffclient.New(ffclient.Config{
		PollingInterval: 10 * time.Second,
		Retrievers: []retriever.Retriever{
			sr,
			ilr,
			il,
			ilf,
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

	assert.True(t, ilf.initCalled)
	assert.True(t, ilf.statusCalled)
	assert.True(t, ilf.retrieveCalled)
	assert.True(t, ilf.shutdownCalled)
}

type simpleRetriever struct {
	retrieveCalled bool
}

// retrieve is the function in charge of fetching the flag configuration.
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

// retrieve is the function in charge of fetching the flag configuration.
func (i *initializableRetrieverLegacy) Retrieve(_ context.Context) ([]byte, error) {
	i.retrieveCalled = true
	return []byte{}, nil
}

// Init is initializing the retriever to start fetching the flags configuration.
func (i *initializableRetrieverLegacy) Init(_ context.Context, _ *log.Logger) error {
	i.initCalled = true
	return nil
}

// Shutdown gracefully shutdown the provider and set the status as not ready.
func (i *initializableRetrieverLegacy) Shutdown(_ context.Context) error {
	i.shutdownCalled = true
	return nil
}

// Status is the function returning the internal state of the retriever.
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

// retrieve is the function in charge of fetching the flag configuration.
func (i *initializableRetriever) Retrieve(_ context.Context) ([]byte, error) {
	i.retrieveCalled = true
	return []byte{}, nil
}

// Init is initializing the retriever to start fetching the flags configuration.
func (i *initializableRetriever) Init(_ context.Context, _ *fflog.FFLogger) error {
	i.initCalled = true
	return nil
}

// Shutdown gracefully shutdown the provider and set the status as not ready.
func (i *initializableRetriever) Shutdown(_ context.Context) error {
	i.shutdownCalled = true
	return nil
}

// Status is the function returning the internal state of the retriever.
func (i *initializableRetriever) Status() retriever.Status {
	i.statusCalled = true
	return retriever.RetrieverReady
}

type initializableRetrieverWithFlagset struct {
	retrieveCalled bool
	initCalled     bool
	shutdownCalled bool
	statusCalled   bool
}

// retrieve is the function in charge of fetching the flag configuration.
func (i *initializableRetrieverWithFlagset) Retrieve(_ context.Context) ([]byte, error) {
	i.retrieveCalled = true
	return []byte{}, nil
}

// Init is initializing the retriever to start fetching the flags configuration.
func (i *initializableRetrieverWithFlagset) Init(_ context.Context, _ *fflog.FFLogger, _ *string) error {
	i.initCalled = true
	return nil
}

// Shutdown gracefully shutdown the provider and set the status as not ready.
func (i *initializableRetrieverWithFlagset) Shutdown(_ context.Context) error {
	i.shutdownCalled = true
	return nil
}

// Status is the function returning the internal state of the retriever.
func (i *initializableRetrieverWithFlagset) Status() retriever.Status {
	i.statusCalled = true
	return retriever.RetrieverReady
}
