package ffclient

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thomaspoignant/go-feature-flag/modules/core/ffcontext"
	"github.com/thomaspoignant/go-feature-flag/modules/core/flag"
	"github.com/thomaspoignant/go-feature-flag/retriever/fileretriever"
)

func TestGoFeatureFlag_FlagDependencies(t *testing.T) {
	flagFile := filepath.Join(t.TempDir(), "flags.yaml")
	content := `
parent:
  variations: {on: true, off: false}
  defaultRule: {variation: off}
child:
  variations: {on: true, off: false}
  defaultRule: {variation: on}
  dependsOn:
    - flag: parent
      variation: "on"
      appliesTo: ["on"]
orphan:
  variations: {on: true, off: false}
  defaultRule: {variation: on}
  dependsOn:
    - flag: unknown
      notVariation: "off"
expiring-child:
  variations: {on: true, off: false}
  defaultRule: {variation: on}
  dependsOn:
    - flag: scheduled-parent
      notVariation: "off"
      expiresWithParent: true
scheduled-parent:
  variations: {on: true, off: false}
  defaultRule: {variation: on}
  scheduledRollout:
    - date: 2020-01-02T00:00:00Z
      disable: true
`
	require.NoError(t, os.WriteFile(flagFile, []byte(content), 0o600))
	g, err := New(Config{PollingInterval: time.Minute, Retriever: &fileretriever.Retriever{Path: flagFile}})
	require.NoError(t, err)
	defer g.Close()
	ctx := ffcontext.NewEvaluationContext("user")

	res, _ := g.BoolVariationDetails("child", ctx, false)
	assert.False(t, res.Value)
	assert.Equal(t, flag.ErrorCodeFlagDependency, res.ErrorCode)
	assert.False(t, res.Cacheable)

	res, _ = g.BoolVariationDetails("orphan", ctx, false)
	assert.Equal(t, flag.ErrorCodeFlagDependency, res.ErrorCode)

	// the scheduled parent is disabled, the expiring child is disabled with it, without error.
	res, _ = g.BoolVariationDetails("expiring-child", ctx, false)
	assert.False(t, res.Value)
	assert.Equal(t, flag.ReasonDisabled, res.Reason)
	assert.Empty(t, res.ErrorCode)

	states := g.AllFlagsState(ctx)
	assert.Equal(t, flag.ErrorCodeFlagDependency, states.GetFlags()["child"].ErrorCode)
	assert.Equal(t, flag.ReasonDisabled, states.GetFlags()["expiring-child"].Reason)
}

func TestGoFeatureFlag_GetFlagStatesSelection(t *testing.T) {
	flagFile := filepath.Join(t.TempDir(), "flags.yaml")
	require.NoError(t, os.WriteFile(flagFile, []byte(`
parent:
  variations: {on: true, off: false}
  defaultRule: {variation: on}
`), 0o600))
	g, err := New(Config{
		PollingInterval: 0,
		Retriever:       &fileretriever.Retriever{Path: flagFile},
	})
	require.NoError(t, err)
	ctx := ffcontext.NewEvaluationContext("user")

	states := g.GetFlagStates(ctx, []string{"parent", "unknown"})
	assert.Contains(t, states.GetFlags(), "parent")
	assert.NotContains(t, states.GetFlags(), "unknown")

	g.Close()
	closed := g.GetFlagStates(ctx, []string{"parent"})
	assert.Empty(t, closed.GetFlags())
}
