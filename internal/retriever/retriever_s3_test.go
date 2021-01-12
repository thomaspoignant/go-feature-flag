package retriever_test

import (
	"context"
	"errors"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/aws/aws-sdk-go/service/s3/s3manager/s3manageriface"
	"github.com/stretchr/testify/assert"
	"io"
	"io/ioutil"
	"testing"

	"github.com/thomaspoignant/go-feature-flag/internal/retriever"
)

type s3ManagerMock struct {
}

func (s s3ManagerMock) Download(at io.WriterAt, input *s3.GetObjectInput, f ...func(*s3manager.Downloader)) (int64, error) {
	if *input.Key == "valid" {
		res, _ := ioutil.ReadFile("../../testdata/test.yaml")
		_, _ = at.WriteAt(res, 0)
		return 1, nil
	} else if *input.Key == "no-file" {
		return 0, errors.New("no file")
	}

	return 1, nil
}

func (s s3ManagerMock) DownloadWithContext(context aws.Context, at io.WriterAt, input *s3.GetObjectInput, f ...func(*s3manager.Downloader)) (int64, error) {
	return s.Download(at, input)
}

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
				downloader: s3ManagerMock{},
				bucket:     "Bucket",
				item:       "valid",
			},
			want:    "../../testdata/test.yaml",
			wantErr: false,
		},
		{
			name: "File not present S3",
			fields: fields{
				downloader: s3ManagerMock{},
				bucket:     "Bucket",
				item:       "no-file",
			},
			wantErr: true,
		},
		{
			name: "File on S3 with context",
			fields: fields{
				downloader: s3ManagerMock{},
				bucket:     "Bucket",
				item:       "valid",
				context:    context.Background(),
			},
			want:    "../../testdata/test.yaml",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := retriever.NewS3Retriever(tt.fields.downloader, tt.fields.bucket, tt.fields.item)
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
