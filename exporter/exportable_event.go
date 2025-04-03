package exporter

import (
	"text/template"
)

type ExportableEvent interface {
	// GetUserKey returns the unique key for the event.
	GetUserKey() string
	// GetKey returns the unique key for the event.
	GetKey() string
	// GetCreationDate returns the creationDate of the event.
	GetCreationDate() int64
	// FormatInCSV FormatEventInCSV returns the event in CSV format.
	FormatInCSV(csvTemplate *template.Template) ([]byte, error)
	// FormatInJSON FormatEventInJSON returns the event in JSON format.
	FormatInJSON() ([]byte, error)
}
