package ffexporter

import (
	"bytes"
	"context"
	"log"
	"sync"
	"text/template"
	"time"

	"github.com/thomaspoignant/go-feature-flag/internal/exporter"
)

const defaultLoggerFormat = "[{{ .FormattedDate}}] user=\"{{ .UserKey}}\", flag=\"{{ .Key}}\", value=\"{{ .Value}}\""

type Log struct {
	// Format is the template configuration of the output format of your log.
	// You can use all the key from the exporter.FeatureEvent + a key called FormattedDate that represent the date with
	// the RFC 3339 Format
	// Default: [{{ .FormattedDate}}] user="{{ .UserKey}}", flag="{{ .Key}}", value="{{ .Value}}"
	Format string

	logTemplate   *template.Template
	initTemplates sync.Once
}

// Export is saving a collection of events in a file.
func (f *Log) Export(ctx context.Context, logger *log.Logger, featureEvents []exporter.FeatureEvent) error {
	f.initTemplates.Do(func() {
		f.logTemplate = parseTemplate("logFormat", f.Format, defaultLoggerFormat)
	})

	for _, event := range featureEvents {
		var log bytes.Buffer
		err := f.logTemplate.Execute(&log, struct {
			exporter.FeatureEvent
			FormattedDate string
		}{FeatureEvent: event, FormattedDate: time.Unix(event.CreationDate, 0).Format(time.RFC3339)})

		logger.Print(log.String())
		if err != nil {
			return err
		}
	}
	return nil
}

func (f *Log) IsBulk() bool {
	return false
}
