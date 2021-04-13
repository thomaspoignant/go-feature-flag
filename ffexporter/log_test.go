package ffexporter_test

import (
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"log"
	"testing"

	"github.com/thomaspoignant/go-feature-flag/ffexporter"
	"github.com/thomaspoignant/go-feature-flag/internal/exporter"
	"github.com/thomaspoignant/go-feature-flag/testutil"
)

func TestLog_Export(t *testing.T) {
	type fields struct {
		Format string
	}
	type args struct {
		featureEvents []exporter.FeatureEvent
	}
	tests := []struct {
		name        string
		fields      fields
		args        args
		expectedLog string
	}{
		{
			name:   "Default format",
			fields: fields{Format: ""},
			args: args{featureEvents: []exporter.FeatureEvent{
				{Kind: "feature", ContextKind: "anonymousUser", UserKey: "ABCD", CreationDate: 1617970547, Key: "random-key",
					Variation: "Default", Value: "YO", Default: false},
			}},
			expectedLog: "\\[" + testutil.RFC3339Regex + "\\] user=\"ABCD\", flag=\"random-key\", value=\"YO\"\n",
		},
		{
			name: "Custom format",
			fields: fields{
				Format: "key=\"{{ .Key}}\" [{{ .FormattedDate}}]",
			},
			args: args{featureEvents: []exporter.FeatureEvent{
				{Kind: "feature", ContextKind: "anonymousUser", UserKey: "ABCD", CreationDate: 1617970547, Key: "random-key",
					Variation: "Default", Value: "YO", Default: false},
			}},
			expectedLog: "key=\"random-key\" \\[" + testutil.RFC3339Regex + "\\]\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &ffexporter.Log{
				Format: tt.fields.Format,
			}

			logFile, _ := ioutil.TempFile("", "")
			logger := log.New(logFile, "", 0)

			err := f.Export(logger, tt.args.featureEvents)
			assert.NoError(t, err, "Log exporter should not throw errors")

			logContent, _ := ioutil.ReadFile(logFile.Name())
			assert.Regexp(t, tt.expectedLog, string(logContent))
		})
	}
}
