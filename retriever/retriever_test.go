package retriever_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	ffclient "github.com/thomaspoignant/go-feature-flag"
	"github.com/thomaspoignant/go-feature-flag/retriever"
	"github.com/thomaspoignant/go-feature-flag/testutils/mock/mockretriever"
)

func TestMixLegacyTypesOfRetrievers(t *testing.T) {
	simple := mockretriever.NewSimpleRetriever("simple")
	initRetrieverLegacy := mockretriever.NewInitializableRetrieverLegacy("legacy")
	initRetriever := mockretriever.NewInitializableRetriever("standard")
	intRetrieverFlagset := mockretriever.NewInitializableRetrieverWithFlagset("flagset")
	goff, err := ffclient.New(ffclient.Config{
		PollingInterval: 10 * time.Second,
		Retrievers: []retriever.Retriever{
			simple,
			initRetrieverLegacy,
			initRetriever,
			intRetrieverFlagset,
		},
	})
	assert.NoError(t, err)
	goff.Close()

	assert.True(t, simple.RetrieveCalled)

	assert.True(t, initRetrieverLegacy.InitCalled)
	assert.True(t, initRetrieverLegacy.StatusCalled)
	assert.True(t, initRetrieverLegacy.RetrieveCalled)
	assert.True(t, initRetrieverLegacy.ShutdownCalled)

	assert.True(t, initRetriever.InitCalled)
	assert.True(t, initRetriever.StatusCalled)
	assert.True(t, initRetriever.RetrieveCalled)
	assert.True(t, initRetriever.ShutdownCalled)

	assert.True(t, intRetrieverFlagset.InitCalled)
	assert.True(t, intRetrieverFlagset.StatusCalled)
	assert.True(t, intRetrieverFlagset.RetrieveCalled)
	assert.True(t, intRetrieverFlagset.ShutdownCalled)
}
