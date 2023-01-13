package fileretriever_test

import (
	"context"
	"testing"

	"github.com/thomaspoignant/go-feature-flag/retriever/fileretriever"

	"github.com/stretchr/testify/assert"
)

var expectedFile = `test-flag:
  variations:
    true_var: true
    false_var: false
  targeting:
    - query: key eq "random-key"
      percentage:
        true_var: 0
        false_var: 100
  defaultRule:
    variation: false_var
  trackEvents: false

test-flag2:
  variations:
    true_var: true
    false_var: false
  targeting:
    - query: key eq "not-a-key"
      percentage:
        true_var: 0
        false_var: 100
  defaultRule:
    variation: false_var
  trackEvents: false
`

func Test_localRetriever_Retrieve(t *testing.T) {
	type fields struct {
		path string
	}
	tests := []struct {
		name    string
		fields  fields
		want    []byte
		wantErr bool
	}{
		{
			name: "File exists",
			fields: fields{
				path: "./testdata/flag-config.yaml",
			},
			want:    []byte(expectedFile),
			wantErr: false,
		},
		{
			name: "File does not exists",
			fields: fields{
				path: "./testdata/test-not-exist.yaml",
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := fileretriever.Retriever{Path: tt.fields.path}
			got, err := l.Retrieve(context.Background())
			if tt.wantErr {
				assert.Error(t, err, "Retrieve() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, string(tt.want), string(got))
		})
	}
}
