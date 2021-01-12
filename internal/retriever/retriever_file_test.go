package retriever_test

import (
	"context"
	"github.com/google/go-cmp/cmp"
	"testing"

	"github.com/thomaspoignant/go-feature-flag/internal/retriever"
)

var expectedFile = `test-flag:
  rule: key eq "random-key"
  percentage: 100
  true: true
  false: false
  default: false

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
				path: "../../testdata/test.yaml",
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
			if (err != nil) != tt.wantErr {
				t.Errorf("Retrieve() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !cmp.Equal(got, tt.want) {
				t.Errorf(cmp.Diff(got, tt.want))
			}
		})
	}
}
