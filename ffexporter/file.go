package ffexporter

import (
	"encoding/json"
	"log"
	"os"
	"strings"
	"sync"
	"text/template"
	"time"

	"github.com/thomaspoignant/go-feature-flag/internal/exporter"
)

type File struct {
	// Format should be JSON or CSV depending
	Format string
	// OutputDir is the location where to store the exported files
	OutputDir string
	// Filename is the template used to name the file
	Filename string

	// CsvTemplate is used if you want to personalized your CSV output (only used if the format is CSV)
	CsvTemplate string

	csvTemplate      *template.Template
	filenameTemplate *template.Template

	initTemplate sync.Once
}

func (f *File) Export(logger *log.Logger, featureEvents []exporter.FeatureEvent) error {
	// Parse the template only once
	f.initTemplate.Do(func() {
		f.csvTemplate = parseTemplate(f.CsvTemplate, DefaultCsvTemplate)
		f.filenameTemplate = parseTemplate(f.Filename, DefaultFilenameTemplate)
	})

	// Default format for the output
	if f.Format == "" {
		f.Format = "json"
	}

	// Get the filename
	// TODO: handle error
	filename, _ := computeFilename(f.filenameTemplate, f.Format)
	filePath := f.OutputDir + "/" + filename

	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	for _, event := range featureEvents {
		switch strings.ToLower(f.Format) {
		case "csv":
			err := f.csvTemplate.Execute(file, event)
			if err != nil {
				if logger != nil {
					logger.Printf("[%v] impossible to parse the event in CSV: %v\n", time.Now().Format(time.RFC3339), err)
				}
			}
		case "json":
		default:
			b, err := json.Marshal(event)
			if err != nil {
				if logger != nil {
					logger.Printf("[%v] error while marshal into JSON: %v\n", time.Now().Format(time.RFC3339), err)
				}
			}
			b = append(b, []byte("\n")...)
			_, err = file.Write(b)
			if err != nil {
				if logger != nil {
					logger.Printf("[%v] error while writing the export file: %v\n", time.Now().Format(time.RFC3339), err)
				}
			}
		}
	}
	return nil
}
