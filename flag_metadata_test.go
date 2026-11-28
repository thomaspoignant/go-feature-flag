package ffclient_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	ffclient "github.com/thomaspoignant/go-feature-flag"
	"github.com/thomaspoignant/go-feature-flag/retriever/fileretriever"
)

func TestGetFlagMetadata(t *testing.T) {
	path := filepath.Join(t.TempDir(), "flags.yaml")
	const config = `target:
  disable: true
  variations: {on: true}
  defaultRule: {variation: on}
  metadata:
    owner: team
    groups: [{name: original}]
    weights: {3: [4]}
empty:
  variations: {on: true}
  defaultRule: {variation: on}
`
	require.NoError(t, os.WriteFile(path, []byte(config), 0600))
	require.NoError(t, ffclient.Init(ffclient.Config{Retriever: &fileretriever.Retriever{Path: path}}))
	t.Cleanup(ffclient.Close)

	got, err := ffclient.GetFlagMetadata("target")
	require.NoError(t, err)
	require.Equal(t, "team", got["owner"])
	got["owner"] = "changed"
	got["groups"].([]any)[0].(map[string]any)["name"] = "changed"
	got["weights"].(map[any]any)[3].([]any)[0] = 99

	again, err := ffclient.GetFlagMetadata("target")
	require.NoError(t, err)
	require.Equal(t, "team", again["owner"])
	require.Equal(t, []any{map[string]any{"name": "original"}}, again["groups"])
	require.Equal(t, map[any]any{3: []any{4}}, again["weights"])

	none, err := ffclient.GetFlagMetadata("empty")
	require.NoError(t, err)
	require.Nil(t, none)
	none, err = ffclient.GetFlagMetadata("missing")
	require.Error(t, err)
	require.Nil(t, none)
}

func TestGetFlagMetadataAfterRefresh(t *testing.T) {
	path := filepath.Join(t.TempDir(), "flags.yaml")
	const config = "checkout:\n  variations: {on: true}\n  defaultRule: {variation: on}\n  metadata: {owner: first}\n"
	require.NoError(t, os.WriteFile(path, []byte(config), 0600))
	client, err := ffclient.New(ffclient.Config{Retriever: &fileretriever.Retriever{Path: path}})
	require.NoError(t, err)
	t.Cleanup(client.Close)

	before, err := client.GetFlagMetadata("checkout")
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(path, []byte(strings.Replace(config, "first", "second", 1)), 0600))
	require.True(t, client.ForceRefresh())
	after, err := client.GetFlagMetadata("checkout")
	require.NoError(t, err)
	require.Equal(t, "first", before["owner"])
	require.Equal(t, "second", after["owner"])
}

func TestGetFlagMetadataUnavailable(t *testing.T) {
	offline, err := ffclient.New(ffclient.Config{Offline: true})
	require.NoError(t, err)
	t.Cleanup(offline.Close)
	for _, client := range []*ffclient.GoFeatureFlag{nil, {}, offline} {
		got, err := client.GetFlagMetadata("flag")
		require.Error(t, err)
		require.Nil(t, got)
	}
}
