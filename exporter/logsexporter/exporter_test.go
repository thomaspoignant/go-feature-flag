package logsexporter_test

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thomaspoignant/go-feature-flag/exporter"
	"github.com/thomaspoignant/go-feature-flag/exporter/logsexporter"
	"github.com/thomaspoignant/go-feature-flag/testutils"
	"github.com/thomaspoignant/go-feature-flag/testutils/slogutil"
	"github.com/thomaspoignant/go-feature-flag/utils/fflog"
)

func TestLog_Export(t *testing.T) {
	type fields struct {
		LogFormat string
	}
	type args struct {
		featureEvents []exporter.ExportableEvent
	}
	tests := []struct {
		name        string
		fields      fields
		args        args
		expectedLog string
		wantErr     bool
	}{
		{
			name:   "Default format",
			fields: fields{LogFormat: ""},
			args: args{featureEvents: []exporter.ExportableEvent{
				exporter.FeatureEvent{
					Kind: "feature", ContextKind: "anonymousUser", UserKey: "ABCD", CreationDate: 1617970547, Key: "random-key",
					Variation: "Default", Value: "YO", Default: false,
				},
			}},
			expectedLog: "user=\"ABCD\", flag=\"random-key\", value=\"YO\"\n",
		},
		{
			name: "Custom format",
			fields: fields{
				LogFormat: "key=\"{{ .Key}}\"",
			},
			args: args{featureEvents: []exporter.ExportableEvent{
				exporter.FeatureEvent{
					Kind: "feature", ContextKind: "anonymousUser", UserKey: "ABCD", CreationDate: 1617970547, Key: "random-key",
					Variation: "Default", Value: "YO", Default: false,
				},
			}},
			expectedLog: "key=\"random-key\"\n",
		},
		{
			name: "LogFormat error",
			fields: fields{
				LogFormat: "key=\"{{ .Key}\" [{{ .FormattedDate}}]",
			},
			args: args{featureEvents: []exporter.ExportableEvent{
				exporter.FeatureEvent{
					Kind: "feature", ContextKind: "anonymousUser", UserKey: "ABCD", CreationDate: 1617970547, Key: "random-key",
					Variation: "Default", Value: "YO", Default: false,
				},
			}},
			expectedLog: "^\\[" + testutils.RFC3339Regex + "\\] user=\"ABCD\", flag=\"random-key\", value=\"YO\"\n",
		},
		{
			name: "Field does not exist",
			fields: fields{
				LogFormat: "key=\"{{ .UnknownKey}}\" [{{ .FormattedDate}}]",
			},
			args: args{featureEvents: []exporter.ExportableEvent{
				exporter.FeatureEvent{
					Kind: "feature", ContextKind: "anonymousUser", UserKey: "ABCD", CreationDate: 1617970547, Key: "random-key",
					Variation: "Default", Value: "YO", Default: false,
				},
			}},
			expectedLog: "^\\[" + testutils.RFC3339Regex + "\\] user=\"ABCD\", flag=\"random-key\", value=\"YO\"\n",
			wantErr:     true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &logsexporter.Exporter{
				LogFormat: tt.fields.LogFormat,
			}

			logFile, _ := os.CreateTemp("", "")
			textHandler := slogutil.MessageOnlyHandler{Writer: logFile}
			logger := &fflog.FFLogger{LeveledLogger: slog.New(&textHandler)}
			err := f.Export(context.Background(), logger, tt.args.featureEvents)

			if tt.wantErr {
				assert.Error(t, err, "It should return an error")
				return
			}

			assert.NoError(t, err, "DeprecatedExporterV1 should not throw errors")

			logContent, _ := os.ReadFile(logFile.Name())
			assert.Regexp(t, tt.expectedLog, string(logContent))
		})
	}
}

func TestLog_IsBulk(t *testing.T) {
	exporter := logsexporter.Exporter{}
	assert.False(t, exporter.IsBulk(), "File exporter is not a bulk exporter")
}
