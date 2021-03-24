package ffclient_test

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/stretchr/testify/assert"
	"net/http"
	"reflect"
	"testing"

	ffClient "github.com/thomaspoignant/go-feature-flag"
)

func TestConfig_GetRetriever(t *testing.T) {
	type fields struct {
		PollInterval int
		Retriever    ffClient.Retriever
	}
	tests := []struct {
		name    string
		fields  fields
		want    string
		wantErr bool
	}{
		{
			name: "File retriever",
			fields: fields{
				PollInterval: 3,
				Retriever:    &ffClient.FileRetriever{Path: "file-example.yaml"},
			},
			want:    "*retriever.localRetriever",
			wantErr: false,
		},
		{
			name: "S3 retriever",
			fields: fields{
				PollInterval: 3,
				Retriever: &ffClient.S3Retriever{
					Bucket: "tpoi-test",
					Item:   "flag-config.yaml",
					AwsConfig: aws.Config{
						Region: aws.String("eu-west-1"),
					},
				},
			},
			want:    "*retriever.s3Retriever",
			wantErr: false,
		},
		{
			name: "HTTP retriever",
			fields: fields{
				PollInterval: 3,
				Retriever: &ffClient.HTTPRetriever{
					URL:    "http://example.com/flag-config.yaml",
					Method: http.MethodGet,
				},
			},
			want:    "*retriever.httpRetriever",
			wantErr: false,
		},
		{
			name: "Github retriever",
			fields: fields{
				PollInterval: 3,
				Retriever: &ffClient.GithubRetriever{
					RepositorySlug: "thomaspoignant/go-feature-flag",
					FilePath:       "testdata/flag-config.yaml",
					GithubToken:    "XXX",
				},
			},
			// we should have a http retriever because Github retriever is using httpRetriever
			want:    "*retriever.httpRetriever",
			wantErr: false,
		},
		{
			name: "No retriever",
			fields: fields{
				PollInterval: 3,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &ffClient.Config{
				PollInterval: tt.fields.PollInterval,
				Retriever:    tt.fields.Retriever,
			}
			got, err := c.GetRetriever()
			assert.Equal(t, tt.wantErr, err != nil)
			if err == nil {
				assert.Equal(t, tt.want, reflect.ValueOf(got).Type().String())
			}
		})
	}
}
