package ffexporter

import (
	"bytes"
	"os"
	"strconv"
	"strings"
	"text/template"
	"time"
)

const DefaultCsvTemplate = "{{ .Kind}};{{ .ContextKind}};{{ .UserKey}};{{ .CreationDate}};{{ .Key}};{{ .Variation}};" +
	"{{ .Value}};{{ .Default}}\n"
const DefaultFilenameTemplate = "flag-variation-{{ .Hostname}}-{{ .Timestamp}}.{{ .Format}}"

// parseTemplate is parsing the template given by the config or use the default template
func parseTemplate(csvTemplate string, defaultTemplate string) *template.Template {
	if csvTemplate == "" {
		csvTemplate = defaultTemplate
	}
	t, err := template.New("exporter").Parse(csvTemplate)
	if err != nil {
		// TODO: log that we are using default template
		t, _ = template.New("exporter").Parse(defaultTemplate)
	}
	return t
}

// getFileName is computing the filename to use for the export file
func computeFilename(template *template.Template, format string) (string, error) {
	hostname, _ := os.Hostname()
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	format = strings.ToLower(format)

	var buf bytes.Buffer
	// TODO: Handle error
	err := template.Execute(&buf, struct {
		Hostname  string
		Timestamp string
		Format    string
	}{
		Hostname:  hostname,
		Timestamp: timestamp,
		Format:    format,
	})

	return buf.String(), err
}
