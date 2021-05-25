package ffclient

import (
	"context"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3/s3manager/s3manageriface"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"testing"

	"github.com/thomaspoignant/go-feature-flag/testutils"
)

func Test_s3Retriever_Retrieve(t *testing.T) {
	type fields struct {
		downloader s3manageriface.DownloaderAPI
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
				downloader: &testutils.S3ManagerMock{},
				bucket:     "Bucket",
				item:       "valid",
			},
			want:    "./testdata/flag-config.yaml",
			wantErr: false,
		},
		{
			name: "File not present S3",
			fields: fields{
				downloader: &testutils.S3ManagerMock{},
				bucket:     "Bucket",
				item:       "no-file",
			},
			wantErr: true,
		},
		{
			name: "File on S3 with context",
			fields: fields{
				downloader: &testutils.S3ManagerMock{},
				bucket:     "Bucket",
				item:       "valid",
				context:    context.Background(),
			},
			want:    "./testdata/flag-config.yaml",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := S3Retriever{
				Bucket:     tt.fields.bucket,
				Item:       tt.fields.item,
				AwsConfig:  aws.Config{},
				downloader: tt.fields.downloader,
			}
			got, err := s.Retrieve(tt.fields.context)
			assert.Equal(t, tt.wantErr, err != nil, "Retrieve() error = %v, wantErr %v", err, tt.wantErr)
			if err == nil {
				want, err := ioutil.ReadFile(tt.want)
				assert.NoError(t, err)
				assert.Equal(t, want, got, "Retrieve() got = %v, want %v", string(got), string(want))
			}
		})
	}
}
