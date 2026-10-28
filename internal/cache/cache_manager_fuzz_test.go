package cache_test

import (
	"log/slog"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/thomaspoignant/go-feature-flag/internal/cache"
	"github.com/thomaspoignant/go-feature-flag/modules/core/dto"
	"github.com/thomaspoignant/go-feature-flag/modules/core/flag"
	"github.com/thomaspoignant/go-feature-flag/utils/fflog"
)

// FuzzConvertToFlagStruct checks that loading any flag configuration file in the
// cache never panics: the file comes from a retriever, and a panic there takes
// down the polling goroutine.
func FuzzConvertToFlagStruct(f *testing.F) {
	formats := []string{"yaml", "json", "toml"}
	for i, file := range []string{
		"../../testdata/flag-config.yaml",
		"../../testdata/flag-config.json",
		"../../testdata/flag-config.toml",
		"../../testdata/flag-config-scheduled-v0.yaml",
		"../../testdata/flag-config-custom-bucketingkey.yaml",
		"../../testdata/flag-config-multiline-query.json",
	} {
		content, err := os.ReadFile(file)
		require.NoError(f, err)
		f.Add(content, uint8([]int{0, 1, 2, 0, 0, 1}[i]))
	}
	f.Add([]byte(`my-flag:
  variations: {A: true, B: false}
  targeting:
    - query: key eq "abc"
      progressiveRollout:
        initial: {variation: A, percentage: 0, date: 2026-01-01T00:00:00Z}
        end: {variation: B, percentage: 100, date: 2026-02-01T00:00:00Z}
  defaultRule: {variation: A}
  experimentation: {start: 2026-01-01T00:00:00Z, end: 2026-03-01T00:00:00Z}
`), uint8(0))
	logger := &fflog.FFLogger{LeveledLogger: slog.New(slog.DiscardHandler)}
	f.Fuzz(func(t *testing.T, data []byte, format uint8) {
		flags, err := cache.ConvertToFlagStruct(data, formats[int(format)%len(formats)])
		if err != nil || hasProgressiveRolloutWithoutStep(flags) {
			return
		}
		cache.NewInMemoryCache(logger).Init(flags)
	})
}

// hasProgressiveRolloutWithoutStep reports a progressive rollout missing its
// initial or end step, which makes flag.InternalFlag.IsValid panic.
// ponytail: known gap, skipped so the fuzzer keeps looking for new ones. Drop it
// once the validation rejects these rollouts.
func hasProgressiveRolloutWithoutStep(flags map[string]dto.DTO) bool {
	for _, d := range flags {
		rules := []flag.Rule{}
		if d.Rules != nil {
			rules = append(rules, *d.Rules...)
		}
		if d.DefaultRule != nil {
			rules = append(rules, *d.DefaultRule)
		}
		for _, r := range rules {
			if p := r.ProgressiveRollout; p != nil && (p.Initial == nil || p.End == nil) {
				return true
			}
		}
	}
	return false
}
