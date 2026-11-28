//go:build bench

package ffclient_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	ffclient "github.com/thomaspoignant/go-feature-flag"
	"github.com/thomaspoignant/go-feature-flag/retriever/fileretriever"
)

func BenchmarkGetFlagMetadata(b *testing.B) {
	for _, count := range []int{10, 169, 1000} {
		b.Run(fmt.Sprintf("%d_flags", count), func(b *testing.B) {
			var flags strings.Builder
			for i := range count {
				fmt.Fprintf(&flags, "flag-%d:\n  variations: {on: true}\n  defaultRule: {variation: on}\n  metadata: {owner: team}\n", i)
			}
			path := filepath.Join(b.TempDir(), "flags.yaml")
			require.NoError(b, os.WriteFile(path, []byte(flags.String()), 0600))
			client, err := ffclient.New(ffclient.Config{
				Retriever:       &fileretriever.Retriever{Path: path},
				PollingInterval: time.Hour,
			})
			require.NoError(b, err)
			b.Cleanup(client.Close)

			const key = "flag-0"
			b.Run("all_flags", func(b *testing.B) {
				b.ReportAllocs()
				for b.Loop() {
					flags, err := client.GetFlagsFromCache()
					if err != nil {
						b.Fatal(err)
					}
					if flags[key].GetMetadata()["owner"] != "team" {
						b.Fatal("unexpected metadata")
					}
				}
			})
			b.Run("single_flag", func(b *testing.B) {
				b.ReportAllocs()
				for b.Loop() {
					metadata, err := client.GetFlagMetadata(key)
					if err != nil {
						b.Fatal(err)
					}
					if metadata["owner"] != "team" {
						b.Fatal("unexpected metadata")
					}
				}
			})
		})
	}
}
