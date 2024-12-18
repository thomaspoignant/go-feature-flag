package flagset

import (
	"github.com/thomaspoignant/go-feature-flag/exporter"
	"github.com/thomaspoignant/go-feature-flag/internal/cache"
	"github.com/thomaspoignant/go-feature-flag/retriever"
)

type Type string

const (
	FlagSetTypeStatic  Type = "static"
	FlagSetTypeDynamic Type = "dynamic"
)

type FlagSet struct {
	Name string
	Type Type

	Cache            cache.Manager
	DataExporter     *exporter.Scheduler
	RetrieverManager *retriever.Manager
}
