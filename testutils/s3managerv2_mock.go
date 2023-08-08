package testutils

import (
	"context"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"io"
	"strings"
)

type S3ManagerV2Mock struct {
	S3ManagerMockFileSystem map[string]string
}

func (s *S3ManagerV2Mock) Upload(ctx context.Context, uploadInput *s3.PutObjectInput, opts ...func(uploader *manager.Uploader)) (*manager.UploadOutput, error) {
	if uploadInput.Bucket == nil || *uploadInput.Bucket == "" {
		return nil, errors.New("invalid bucket")
	}

	if s.S3ManagerMockFileSystem == nil {
		s.S3ManagerMockFileSystem = make(map[string]string)
	}

	buf := new(strings.Builder)
	_, err := io.Copy(buf, uploadInput.Body)
	if err != nil {
		fmt.Println(err)
	}
	s.S3ManagerMockFileSystem[*uploadInput.Key] = buf.String()

	return &manager.UploadOutput{
		Location: *uploadInput.Key,
	}, nil
}
