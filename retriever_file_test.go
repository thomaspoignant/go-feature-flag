package ffclient_test

import (
	"context"
	"github.com/stretchr/testify/assert"
	ffclient "github.com/thomaspoignant/go-feature-flag"
	"testing"
)

var expectedFile = `test-flag:
  variations:
    Default: false
    "False": false
    "True": true
  targeting:
    - query: key eq "random-key"
      percentage:
        - "True": 100
        - "False": 0
  defaultRule:
    variation: Default

test-flag2:
  variations:
    Default: false
    "False": false
    "True": true
  targeting:
    - query: key eq "not-a-key"
      percentage:
        - "True": 100
        - "False": 0
  defaultRule:
    variation: Default
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
				path: "./testdata/flag-config-v2.yaml",
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
			l := ffclient.FileRetriever{Path: tt.fields.path}
			got, err := l.Retrieve(context.Background())
			if tt.wantErr {
				assert.Error(t, err, "Retrieve() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, string(tt.want), string(got))
		})
	}
}
