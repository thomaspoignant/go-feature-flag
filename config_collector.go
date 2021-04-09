package ffclient

import (
	"time"

	"github.com/thomaspoignant/go-feature-flag/internal/exporter"
)

// DataExporter is the configuration of your export target.
type DataExporter struct {
	// FlushInterval is the interval we are waiting to export the data.
	// example: if you set your FlushInterval to 1 minutes, we will send
	// the data every minute unless we reach the max event in cache before.
	FlushInterval time.Duration

	// MaxEventInCache is the maximum number of event you keep in the cache
	// before sending the data to the Exporter.
	// We will send the data when the MaxEventInCache is reach or if we have
	// waited the FlushInterval.
	MaxEventInCache int64

	// Exporter is the configuration of your exporter.
	// You can see all available exporter in the ffexporter package.
	Exporter exporter.Exporter
}
