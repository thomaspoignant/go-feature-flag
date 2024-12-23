package flagset

import (
	"errors"

	"github.com/thomaspoignant/go-feature-flag/exporter"
	"github.com/thomaspoignant/go-feature-flag/notifier"
	"github.com/thomaspoignant/go-feature-flag/retriever"
)

var (
	ErrInvalidFlagSetType = errors.New("invalid flagset type")
	ErrNameRequired       = errors.New("flagset name is required")
)

type Builder interface {
	Notifiers([]notifier.Notifier) Builder
	Retrievers([]retriever.Retriever) Builder
	Exporter(exporter.CommonExporter) Builder
	Build() (*FlagSet, error)
}

type builderImpl struct {
	name        string
	flagSetType Type
	notifiers   []notifier.Notifier
	retrievers  []retriever.Retriever
	exporter    exporter.CommonExporter
}

func NewBuilder(name string, flagSetType Type) Builder {
	return &builderImpl{
		name:        name,
		flagSetType: flagSetType,
	}
}

func (f *builderImpl) Notifiers(notifiers []notifier.Notifier) Builder {
	f.notifiers = notifiers
	return f
}

func (f *builderImpl) Retrievers(retrievers []retriever.Retriever) Builder {
	f.retrievers = retrievers
	return f
}

func (f *builderImpl) Exporter(exporter exporter.CommonExporter) Builder {
	f.exporter = exporter
	return f
}

func (f *builderImpl) Build() (*FlagSet, error) {
	if f.name == "" {
		return nil, ErrNameRequired
	}
	if f.flagSetType != FlagSetTypeDynamic && f.flagSetType != FlagSetTypeStatic {
		return nil, ErrInvalidFlagSetType
	}

	return &FlagSet{
		Name: f.name,
		Type: f.flagSetType,
	}, nil

}
