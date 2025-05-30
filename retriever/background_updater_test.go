package retriever

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestClose(t *testing.T) {
	tests := []struct {
		name              string
		backgroundUpdater backgroundUpdater
	}{
		{
			name: "Close nil background updater",
			backgroundUpdater: backgroundUpdater{
				ticker:      nil,
				updaterChan: nil,
			},
		},
		{
			name:              "Close new background updater",
			backgroundUpdater: newBackgroundUpdater(500*time.Millisecond, false),
		},
		{
			name:              "Close new background updater with jitter",
			backgroundUpdater: newBackgroundUpdater(500*time.Millisecond, true),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotPanics(t, func() {
				tt.backgroundUpdater.close()
			})
		})
	}
}
