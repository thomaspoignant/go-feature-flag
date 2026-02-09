package testutils

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/feature/s3/transfermanager"
)

type S3ManagerV2Mock struct {
	S3ManagerMockFileSystem map[string]string
	TestDataLocation        string
}

func (s *S3ManagerV2Mock) UploadObject(
	ctx context.Context,
	uploadInput *transfermanager.UploadObjectInput,
	opts ...func(*transfermanager.Options),
) (*transfermanager.UploadObjectOutput, error) {
	if ctx == nil {
		return nil, errors.New("cannot create context from nil parent")
	}
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

	return &transfermanager.UploadObjectOutput{
		Location: uploadInput.Key,
	}, nil
}

func (s *S3ManagerV2Mock) DownloadObject(
	ctx context.Context,
	input *transfermanager.DownloadObjectInput,
	opts ...func(*transfermanager.Options),
) (*transfermanager.DownloadObjectOutput, error) {
	if ctx == nil {
		return nil, errors.New("cannot create context from nil parent")
	}
	if input.WriterAt == nil {
		return nil, errors.New("WriterAt is required")
	}

	if input.Key != nil && *input.Key == "valid" {
		res, _ := os.ReadFile(s.TestDataLocation + "/flag-config.yaml")
		_, _ = input.WriterAt.WriteAt(res, 0)
		return &transfermanager.DownloadObjectOutput{}, nil
	}
	if input.Key != nil && *input.Key == "no-file" {
		return nil, errors.New("no file")
	}

	return &transfermanager.DownloadObjectOutput{}, nil
}
