package proxynotifier_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/proxynotifier"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/service"
	"github.com/thomaspoignant/go-feature-flag/modules/core/flag"
	"github.com/thomaspoignant/go-feature-flag/modules/core/testutils/testconvert"
	"github.com/thomaspoignant/go-feature-flag/notifier"
)

func TestNotifierSSE_Notify(t *testing.T) {
	tests := []struct {
		name        string
		flagsetName string
		diff        notifier.DiffCache
	}{
		{
			name:        "empty diff notifies without error",
			flagsetName: "my-flagset",
			diff:        notifier.DiffCache{},
		},
		{
			name:        "diff with added flag notifies without error",
			flagsetName: "my-flagset",
			diff: notifier.DiffCache{
				Added: map[string]flag.Flag{
					"flag-1": &flag.InternalFlag{
						Variations: &map[string]*any{
							"A": testconvert.Interface(true),
						},
						DefaultRule: &flag.Rule{VariationResult: testconvert.String("A")},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sseService := service.NewSSEService()
			defer sseService.Close()

			n := proxynotifier.NewNotifierSSE(sseService, tt.flagsetName)
			assert.NoError(t, n.Notify(tt.diff))
		})
	}
}
