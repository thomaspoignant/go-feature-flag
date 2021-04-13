package ffexporter

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"text/template"
)

func Test_parseTemplate(t *testing.T) {
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defaultT, _ := template.New("random-name").Parse(tt.args.defaultTemplate)
			assert.NotPanics(t, func() { parseTemplate("random-name", tt.args.template, tt.args.defaultTemplate) })
			got := parseTemplate("random-name", tt.args.template, tt.args.defaultTemplate)

			if tt.wantErr {
				assert.Equal(t, defaultT, got, "If template invalid we should use default template")
				return
			}
			assert.NotEqual(t, defaultT, got, "We should not have the same template")
		})
	}
}
