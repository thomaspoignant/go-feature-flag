package exporter_test

import (
	"fmt"
	"os"
	"testing"
	"text/template"

	"github.com/thomaspoignant/go-feature-flag/exporter"

	"github.com/stretchr/testify/assert"
)

func Test_ParseTemplate(t *testing.T) {
	type args struct {
		template        string
		defaultTemplate string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Invalid template",
			args: args{
				template: "{{ .Kind}};{{ .ContextKind}};{{ .UserKey}};{{ .CreationDate}};{{ .Key}};{{ .Variation}};" +
					"{{ .Value}};{{ .Default}",
				defaultTemplate: "{{ .Kind}};{{ .ContextKind}};{{ .UserKey}};{{ .CreationDate}};{{ .Key}};{{ .Variation}};" +
					"{{ .Value}};{{ .Default}}\n",
			},
			wantErr: true,
		},
		{
			name: "Valid template",
			args: args{
				template: "{{ .Kind}};{{ .ContextKind}};{{ .UserKey}};{{ .CreationDate}};{{ .Key}};{{ .Variation}};",
				defaultTemplate: "{{ .Kind}};{{ .ContextKind}};{{ .UserKey}};{{ .CreationDate}};{{ .Key}};{{ .Variation}};" +
					"{{ .Value}};{{ .Default}}\n",
			},
			wantErr: false,
		},
		{
			name: "empty template use default",
			args: args{
				template: "",
				defaultTemplate: "{{ .Kind}};{{ .ContextKind}};{{ .UserKey}};{{ .CreationDate}};{{ .Key}};{{ .Variation}};" +
					"{{ .Value}};{{ .Default}}\n",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defaultT, _ := template.New("random-name").Parse(tt.args.defaultTemplate)
			assert.NotPanics(t, func() { exporter.ParseTemplate("random-name", tt.args.template, tt.args.defaultTemplate) })
			got := exporter.ParseTemplate("random-name", tt.args.template, tt.args.defaultTemplate)

			if tt.wantErr {
				assert.Equal(t, defaultT, got, "If template invalid we should use default template")
				return
			}
			assert.NotEqual(t, defaultT, got, "We should not have the same template")
		})
	}
}

func TestComputeFilename(t *testing.T) {
	hostname, _ := os.Hostname()
	type args struct {
		template *template.Template
		format   string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "Nothing to template",
			args: args{
				template: exporter.ParseTemplate("filenameFormat", "flag-variation", "flag-variation"),
				format:   "json",
			},
			want:    "flag-variation",
			wantErr: assert.NoError,
		},
		{
			name: "With extension",
			args: args{
				template: exporter.ParseTemplate("filenameFormat", "flag-variation.{{ .Format }}", "flag-variation"),
				format:   "json",
			},
			want:    "flag-variation.json",
			wantErr: assert.NoError,
		},
		{
			name: "Multiple templates",
			args: args{
				template: exporter.ParseTemplate("filenameFormat", "flag-variation-{{ .Hostname}}.{{ .Format}}", "flag-variation"),
				format:   "json",
			},
			want:    "flag-variation-" + hostname + ".json",
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := exporter.ComputeFilename(tt.args.template, tt.args.format)
			if !tt.wantErr(t, err, fmt.Sprintf("ComputeFilename(%v, %v)", tt.args.template, tt.args.format)) {
				return
			}
			assert.Equalf(t, tt.want, got, "ComputeFilename(%v, %v)", tt.args.template, tt.args.format)
		})
	}
}

func TestFormatEventInCSV(t *testing.T) {
	type args struct {
		csvTemplate *template.Template
		event       exporter.FeatureEvent
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "valid",
			args: args{
				csvTemplate: exporter.ParseTemplate("exporterExample", exporter.DefaultCsvTemplate, exporter.DefaultCsvTemplate),
				event: exporter.FeatureEvent{
					Kind: "feature", ContextKind: "anonymousUser", UserKey: "ABCD", CreationDate: 1617970547, Key: "random-key",
					Variation: "Default", Value: "YO", Default: false,
				},
			},
			want:    "feature;anonymousUser;ABCD;1617970547;random-key;Default;YO;false\n",
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := exporter.FormatEventInCSV(tt.args.csvTemplate, tt.args.event)
			if !tt.wantErr(t, err, fmt.Sprintf("FormatEventInCSV(%v, %v)", tt.args.csvTemplate, tt.args.event)) {
				return
			}
			assert.Equalf(t, tt.want, string(got), "FormatEventInCSV(%v, %v)", tt.args.csvTemplate, tt.args.event)
		})
	}
}

func TestFormatEventInJSON(t *testing.T) {
	type args struct {
		event exporter.FeatureEvent
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "valid",
			args: args{event: exporter.FeatureEvent{
				Kind: "feature", ContextKind: "anonymousUser", UserKey: "ABCD", CreationDate: 1617970547, Key: "random-key",
				Variation: "Default", Value: "YO", Default: false,
			}},
			want:    "{\"kind\":\"feature\",\"contextKind\":\"anonymousUser\",\"userKey\":\"ABCD\",\"creationDate\":1617970547,\"key\":\"random-key\",\"variation\":\"Default\",\"value\":\"YO\",\"default\":false,\"version\":0}\n",
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := exporter.FormatEventInJSON(tt.args.event)
			if !tt.wantErr(t, err, fmt.Sprintf("FormatEventInJSON(%v)", tt.args.event)) {
				return
			}
			assert.Equalf(t, tt.want, string(got), "FormatEventInJSON(%v)", tt.args.event)
		})
	}
}
