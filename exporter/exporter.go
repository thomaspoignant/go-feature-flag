package exporter

import (
	"context"
	"log"

	"github.com/thomaspoignant/go-feature-flag/utils/fflog"
)

// DeprecatedExporterV1 is an interface to describe how an exporter looks like.
//
// Deprecated: use Exporter instead.
type DeprecatedExporterV1 interface {
	CommonExporter
	// Export will send the data to the exporter.
	Export(context.Context, *log.Logger, []FeatureEvent) error
}

type DeprecatedExporterV2 interface {
	CommonExporter
	Export(context.Context, *fflog.FFLogger, []FeatureEvent) error
}

type Exporter interface {
	CommonExporter
	Export(context.Context, *fflog.FFLogger, []ExportableEvent) error
}
type CommonExporter interface {
	// IsBulk return false if we should directly send the data as soon as it is produce
	// and true if we collect the data to send them in bulk.
	IsBulk() bool
}
