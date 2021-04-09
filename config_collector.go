package ffclient

import (
	"time"

	"github.com/thomaspoignant/go-feature-flag/internal/exporter"
)

// TODO : Add documentation
type DataExporter struct {
	FlushInterval   time.Duration
	MaxEventInCache int64
	Exporter        exporter.Exporter
}
