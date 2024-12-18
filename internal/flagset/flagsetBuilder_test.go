package flagset_test

import (
	"github.com/stretchr/testify/assert"
	"github.com/thomaspoignant/go-feature-flag/exporter"
	"github.com/thomaspoignant/go-feature-flag/exporter/logsexporter"
	"github.com/thomaspoignant/go-feature-flag/internal/flagset"
	"github.com/thomaspoignant/go-feature-flag/notifier"
	"github.com/thomaspoignant/go-feature-flag/notifier/logsnotifier"
	"github.com/thomaspoignant/go-feature-flag/retriever"
	"github.com/thomaspoignant/go-feature-flag/retriever/fileretriever"
	"github.com/thomaspoignant/go-feature-flag/utils/fflog"
	"log/slog"
	"testing"
)

func TestFlagSetBuilder(t *testing.T) {
	tests := []struct {
		name          string
		flagsetName   string
		flagsetType   flagset.Type
		wantErr       assert.ErrorAssertionFunc
		notifiers     []notifier.Notifier
		retrievers    []retriever.Retriever
		exporter      exporter.CommonExporter
		expected      *flagset.FlagSet
		expectedError string
	}{
		{
			name:          "Should return an error if no name for flag set",
			flagsetName:   "",
			flagsetType:   flagset.FlagSetTypeDynamic,
			wantErr:       assert.Error,
			expectedError: "flagset name is required",
		},
		{
			name:          "Should return an error if no flag set type provided",
			flagsetName:   "test",
			wantErr:       assert.Error,
			expectedError: "invalid flagset type",
		},
		{
			name:          "Should return an error if wrong flag set type provided",
			flagsetName:   "test",
			flagsetType:   "foobar",
			wantErr:       assert.Error,
			expectedError: "invalid flagset type",
		},
		{
			name:        "Should return a flag set will everything set",
			flagsetName: "test",
			wantErr:     assert.NoError,
			notifiers: []notifier.Notifier{
				&logsnotifier.Notifier{
					Logger: &fflog.FFLogger{LeveledLogger: slog.Default()},
				},
			},
			retrievers: []retriever.Retriever{
				&fileretriever.Retriever{Path: "../../testdata/flag1.json"},
			},
			exporter: &logsexporter.Exporter{},
			expected: &flagset.FlagSet{
				Name: "test",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fsBuilder := flagset.NewBuilder(tt.flagsetName, tt.flagsetType)
			fsBuilder.Notifiers(tt.notifiers)
			fsBuilder.Retrievers(tt.retrievers)
			fsBuilder.Exporter(tt.exporter)
			got, err := fsBuilder.Build()
			tt.wantErr(t, err)

			if tt.expectedError != "" {
				assert.Equal(t, tt.expectedError, err.Error())
				return
			}
			assert.Equal(t, tt.expected, got)
		})
	}
}
