package fflog_test

import (
	"github.com/stretchr/testify/assert"
	"log"
	"os"
	"testing"

	"github.com/thomaspoignant/go-feature-flag/internal/fflog"
)

func TestPrintf(t *testing.T) {
	type args struct {
		logger *log.Logger
		format string
		v      []interface{}
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "no logger",
			args: args{
				logger: nil,
				format: "Toto",
				v:      nil,
			},
		},
		{
			name: "with logger",
			args: args{
				logger: log.New(os.Stdout, "", 0),
				format: "Toto %v",
				v:      []interface{}{"toto"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotPanics(t, func() { fflog.Printf(tt.args.logger, tt.args.format, tt.args.v) })
		})
	}
}
