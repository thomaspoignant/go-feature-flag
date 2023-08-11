package s3retrieverv2

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/stretchr/testify/assert"
	"github.com/thomaspoignant/go-feature-flag/testutils"
	"os"
	"testing"
)

func Test_s3Retriever_Retrieve(t *testing.T) {
	type fields struct {
		downloader DownloaderAPI
		bucket     string
		item       string
		context    context.Context
	}
	tests := []struct {
		name    string
		fields  fields
		want    string
		wantErr bool
	}{
		{
			name: "File on S3",
			fields: fields{
				downloader: &testutils.S3ManagerV2Mock{
					TestDataLocation: "./testdata",
				},
				bucket: "Bucket",
				item:   "valid",
			},
			want:    "./testdata/flag-config.yaml",
			wantErr: false,
		},
		{
			name: "File not present S3",
			fields: fields{
				downloader: &testutils.S3ManagerV2Mock{
					TestDataLocation: "./testdata",
				},
				bucket: "Bucket",
				item:   "no-file",
			},
			wantErr: true,
		},
		{
			name: "File on S3 with context",
			fields: fields{
				downloader: &testutils.S3ManagerV2Mock{
					TestDataLocation: "./testdata",
				},
				bucket:  "Bucket",
				item:    "valid",
				context: context.Background(),
			},
			want:    "./testdata/flag-config.yaml",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			awsConf, _ := config.LoadDefaultConfig(context.TODO())
			s := Retriever{
				Bucket:     tt.fields.bucket,
				Item:       tt.fields.item,
				AwsConfig:  &awsConf,
				downloader: tt.fields.downloader,
			}
			got, err := s.Retrieve(tt.fields.context)
			assert.Equal(t, tt.wantErr, err != nil, "Retrieve() error = %v, wantErr %v", err, tt.wantErr)
			if err == nil {
				want, err := os.ReadFile(tt.want)
				assert.NoError(t, err)
				assert.Equal(t, string(want), string(got), "Retrieve() got = %v, want %v", string(want), string(got))
			}
		})
	}
}
