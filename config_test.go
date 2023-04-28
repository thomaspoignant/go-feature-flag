package ffclient_test

import (
	"net/http"
	"reflect"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/thomaspoignant/go-feature-flag/retriever"
	"github.com/thomaspoignant/go-feature-flag/retriever/fileretriever"
	"github.com/thomaspoignant/go-feature-flag/retriever/githubretriever"
	"github.com/thomaspoignant/go-feature-flag/retriever/gitlabretriever"
	"github.com/thomaspoignant/go-feature-flag/retriever/httpretriever"
	"github.com/thomaspoignant/go-feature-flag/retriever/s3retriever"

	"github.com/stretchr/testify/assert"

	ffClient "github.com/thomaspoignant/go-feature-flag"
)

func TestConfig_GetRetrievers(t *testing.T) {
	type fields struct {
		PollingInterval time.Duration
		Retriever       retriever.Retriever
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
				PollingInterval: 3 * time.Second,
				Retriever:       &fileretriever.Retriever{Path: "file-example.yaml"},
			},
			want:    "*fileretriever.Retriever",
			wantErr: false,
		},
		{
			name: "S3 retriever",
			fields: fields{
				PollingInterval: 3 * time.Second,
				Retriever: &s3retriever.Retriever{
					Bucket: "tpoi-test",
					Item:   "flag-config.yaml",
					AwsConfig: aws.Config{
						Region: aws.String("eu-west-1"),
					},
				},
			},
			want:    "*s3retriever.Retriever",
			wantErr: false,
		},
		{
			name: "HTTP retriever",
			fields: fields{
				PollingInterval: 3 * time.Second,
				Retriever: &httpretriever.Retriever{
					URL:    "http://example.com/flag-config.yaml",
					Method: http.MethodGet,
				},
			},
			want:    "*httpretriever.Retriever",
			wantErr: false,
		},
		{
			name: "Github retriever",
			fields: fields{
				PollingInterval: 3 * time.Second,
				Retriever: &githubretriever.Retriever{
					RepositorySlug: "thomaspoignant/go-feature-flag",
					FilePath:       "testdata/flag-config.yaml",
					GithubToken:    "XXX",
				},
			},
			want:    "*githubretriever.Retriever",
			wantErr: false,
		},
		{
			name: "Gitlab retriever",
			fields: fields{
				PollingInterval: 3 * time.Second,
				Retriever: &gitlabretriever.Retriever{
					BaseURL:        "https://gitlab.com/",
					RepositorySlug: "ruairi2/go-feature-flags-config",
					FilePath:       "flag-config.yaml",
					GitlabToken:    "XXX",
				},
			},
			want:    "*gitlabretriever.Retriever",
			wantErr: false,
		},
		{
			name: "No retriever",
			fields: fields{
				PollingInterval: 3 * time.Second,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &ffClient.Config{
				PollingInterval: tt.fields.PollingInterval,
				Retriever:       tt.fields.Retriever,
			}
			got, err := c.GetRetrievers()
			assert.Equal(t, tt.wantErr, err != nil)
			if err == nil {
				assert.Equal(t, tt.want, reflect.ValueOf(got[0]).Type().String())
			}
		})
	}
}
