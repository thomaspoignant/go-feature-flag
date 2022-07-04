package dataexporter

import (
	"context"
	"log"

	"github.com/thomaspoignant/go-feature-flag/exporter"
)

// Exporter is an interface to describe how a exporter looks like.
type Exporter interface {
	// Export will send the data to the exporter.
	Export(context.Context, *log.Logger, []exporter.FeatureEvent) error

	// IsBulk return false if we should directly send the data as soon as it is produce
	// and true if we collect the data to send them in bulk.
	IsBulk() bool
}
