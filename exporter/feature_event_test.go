package exporter_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/thomaspoignant/go-feature-flag/exporter"
	"github.com/thomaspoignant/go-feature-flag/ffuser"
)

func TestNewFeatureEvent(t *testing.T) {
	type args struct {
		user      ffuser.User
		flagKey   string
		value     interface{}
		variation string
		failed    bool
		version   float64
	}
	tests := []struct {
		name string
		args args
		want exporter.FeatureEvent
	}{
		{
			name: "anonymous user",
			args: args{
				user:      ffuser.NewAnonymousUser("ABCD"),
				flagKey:   "random-key",
				value:     "YO",
				variation: "Default",
				failed:    false,
				version:   0,
			},
			want: exporter.FeatureEvent{
				Kind: "feature", ContextKind: "anonymousUser", UserKey: "ABCD", CreationDate: time.Now().Unix(), Key: "random-key",
				Variation: "Default", Value: "YO", Default: false,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, exporter.NewFeatureEvent(tt.args.user, tt.args.flagKey, tt.args.value, tt.args.variation, tt.args.failed, tt.args.version), "NewFeatureEvent(%v, %v, %v, %v, %v, %v)", tt.args.user, tt.args.flagKey, tt.args.value, tt.args.variation, tt.args.failed, tt.args.version)
		})
	}
}
