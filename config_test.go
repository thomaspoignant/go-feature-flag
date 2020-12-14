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
		PollInterval  int
		LocalFile     string
		HTTPRetriever *ffClient.HTTPRetriever
		S3Retriever   *ffClient.S3Retriever
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "File retriever",
			fields: fields{
				PollInterval: 3,
				LocalFile:    "file-example.yaml",
			},
			want: "*retriever.localRetriever",
		},
		{
			name: "S3 retriever",
			fields: fields{
				PollInterval: 3,
				S3Retriever: &ffClient.S3Retriever{
					Bucket: "tpoi-test",
					Item:   "test.yaml",
					AwsConfig: aws.Config{
						Region: aws.String("eu-west-1"),
					},
				},
			},
			want: "*retriever.s3Retriever",
		},
		{
			name: "HTTP retriever",
			fields: fields{
				PollInterval: 3,
				HTTPRetriever: &ffClient.HTTPRetriever{
					URL:    "http://example.com/test.yaml",
					Method: http.MethodGet,
					Body:   "",
				},
			},
			want: "*retriever.httpRetriever",
		},
		{
			name: "Priority to S3",
			fields: fields{
				PollInterval: 3,
				HTTPRetriever: &ffClient.HTTPRetriever{
					URL:    "http://example.com/test.yaml",
					Method: http.MethodGet,
					Body:   "",
				},
				S3Retriever: &ffClient.S3Retriever{
					Bucket: "tpoi-test",
					Item:   "test.yaml",
					AwsConfig: aws.Config{
						Region: aws.String("eu-west-1"),
					},
				},
				LocalFile: "file-example.yaml",
			},
			want: "*retriever.s3Retriever",
		},
		{
			name: "Priority to HTTP",
			fields: fields{
				PollInterval: 3,
				HTTPRetriever: &ffClient.HTTPRetriever{
					URL:    "http://example.com/test.yaml",
					Method: http.MethodGet,
					Body:   "",
				},
				LocalFile: "file-example.yaml",
			},
			want: "*retriever.httpRetriever",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &ffClient.Config{
				PollInterval:  tt.fields.PollInterval,
				LocalFile:     tt.fields.LocalFile,
				HTTPRetriever: tt.fields.HTTPRetriever,
				S3Retriever:   tt.fields.S3Retriever,
			}
			got := c.GetRetriever()
			assert.Equal(t, tt.want, reflect.ValueOf(got).Type().String())
		})
	}
}
