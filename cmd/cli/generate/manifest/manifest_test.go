package manifest_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thomaspoignant/go-feature-flag/cmd/cli/generate/manifest"
)

func TestNewManifestCmd(t *testing.T) {
	type args struct {
		config         string
		format         string
		setDestination bool
	}
	tests := []struct {
		name        string
		args        args
		errorAssert assert.ErrorAssertionFunc
	}{
		{
			name: "should error if config file does not exist",
			args: args{
				config:         "testdata/invalid.yaml",
				setDestination: true,
			},
			errorAssert: assert.Error,
		},
		{
			name: "should not error if config file exists",
			args: args{
				config:         "testdata/input/flag.goff.yaml",
				setDestination: true,
			},
			errorAssert: assert.NoError,
		},
		{
			name: "should not error no destination provided",
			args: args{
				config:         "testdata/input/flag.goff.yaml",
				setDestination: false,
			},
			errorAssert: assert.Error,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			destination := ""
			if tt.args.setDestination {
				f, err := os.CreateTemp("", "temp")
				assert.NoError(t, err)
				destination = f.Name()
				defer func() { _ = os.Remove(destination) }()
			}
			_, err := manifest.NewManifest(tt.args.config, tt.args.format, destination)
			tt.errorAssert(t, err)
		})
	}
}
