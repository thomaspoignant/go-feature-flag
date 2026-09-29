package stream_test

import (
	"github.com/thomaspoignant/go-feature-flag/modules/core/flag"
	"github.com/thomaspoignant/go-feature-flag/modules/core/testutils/testconvert"
	"github.com/thomaspoignant/go-feature-flag/notifier"
)

func sensitiveDiffCache() notifier.DiffCache {
	beforeQuery := `user.email eq "before@example.com"`
	afterQuery := `user.email eq "after@example.com"`
	enabled := "enabled"
	return notifier.DiffCache{
		Deleted: map[string]flag.Flag{
			"deleted-flag": &flag.InternalFlag{
				Variations: &map[string]*any{"enabled": testconvert.Interface(true)},
				Rules: &[]flag.Rule{
					{Query: &beforeQuery, VariationResult: &enabled},
				},
			},
		},
		Added: map[string]flag.Flag{
			"added-flag": &flag.InternalFlag{
				Variations: &map[string]*any{"enabled": testconvert.Interface(true)},
				Rules: &[]flag.Rule{
					{Query: &afterQuery, VariationResult: &enabled},
				},
			},
		},
		Updated: map[string]notifier.DiffUpdated{
			"updated-flag": {
				Before: &flag.InternalFlag{Rules: &[]flag.Rule{{Query: &beforeQuery}}},
				After:  &flag.InternalFlag{Rules: &[]flag.Rule{{Query: &afterQuery}}},
			},
		},
	}
}
