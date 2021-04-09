package ffclient

import (
	"time"

	"github.com/thomaspoignant/go-feature-flag/internal/exporter"
)

type CollectorConfig struct {
	FlushInterval   time.Duration
	MaxEventInCache int64
	Collector       exporter.Exporter
}
