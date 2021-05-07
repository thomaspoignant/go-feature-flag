package retriever_test

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/thomaspoignant/go-feature-flag/internal/retriever"
	"testing"
)

var expectedFile = `test-flag:
  rule: key eq "random-key"
  percentage: 100
  true: true
  false: false
  default: false
  rollout:
    scheduled:
      steps:
        - date: 2021-04-10T00:00:00.10-05:00
          rule: internal eq true
          percentage: 100
        - date: 2021-04-14T00:00:00.10-05:00
          rule: internal eq true OR beta eq true
          percentage: 100

test-flag2:
  rule: key eq "not-a-key"
  percentage: 100
  true: true
  false: false
  default: false
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
				path: "../../testdata/flag-config.yaml",
			},
			want:    []byte(expectedFile),
			wantErr: false,
		},
		{
			name: "File does not exists",
			fields: fields{
				path: "../../testdata/test-not-exist.yaml",
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := retriever.NewLocalRetriever(tt.fields.path)
			got, err := l.Retrieve(context.Background())
			if tt.wantErr {
				assert.Error(t, err, "Retrieve() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, string(tt.want), string(got))
		})
	}
}
