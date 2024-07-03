package ffclient

import (
	"github.com/thomaspoignant/go-feature-flag/exporter"
	"time"
)

// DataExporter is the configuration of your export target.
type DataExporter struct {
	// FlushInterval is the interval we are waiting to export the data.
	// example: if you set your FlushInterval to 1 minutes, we will send
	// the data every minute unless we reach the max event in cache before.
	FlushInterval time.Duration

	// MaxEventInMemory is the maximum number of event you keep in the cache
	// before sending the data to the Exporter.
	// We will send the data when the MaxEventInMemory is reach or if we have
	// waited the FlushInterval.
	MaxEventInMemory int64

	// Exporter is the configuration of your exporter.
	// You can see all available exporter in the exporter package.
	Exporter exporter.CommonExporter
}
