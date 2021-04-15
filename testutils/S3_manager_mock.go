package testutils

import (
	"errors"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"io"
	"io/ioutil"
	"strings"
)

type S3ManagerMock struct {
	S3ManagerMockFileSystem map[string]string
}

func (s *S3ManagerMock) Download(at io.WriterAt, input *s3.GetObjectInput,
	f ...func(*s3manager.Downloader)) (int64, error) {
	if *input.Key == "valid" {
		res, _ := ioutil.ReadFile("../../testdata/flag-config.yaml")
		_, _ = at.WriteAt(res, 0)
		return 1, nil
	} else if *input.Key == "no-file" {
		return 0, errors.New("no file")
	}

	return 1, nil
}

func (s *S3ManagerMock) DownloadWithContext(context aws.Context, at io.WriterAt,
	input *s3.GetObjectInput, f ...func(*s3manager.Downloader)) (int64, error) {
	return s.Download(at, input)
}

func (s *S3ManagerMock) Upload(uploadInput *s3manager.UploadInput,
	funC ...func(*s3manager.Uploader)) (*s3manager.UploadOutput, error) {
	if uploadInput.Bucket == nil || *uploadInput.Bucket == "" {
		return nil, errors.New("invalid bucket")
	}

	if s.S3ManagerMockFileSystem == nil {
		s.S3ManagerMockFileSystem = make(map[string]string)
	}

	buf := new(strings.Builder)
	_, _ = io.Copy(buf, uploadInput.Body)
	s.S3ManagerMockFileSystem[*uploadInput.Key] = buf.String()

	return &s3manager.UploadOutput{
		Location: *uploadInput.Key,
	}, nil
}

func (s *S3ManagerMock) UploadWithContext(context aws.Context, uploadInput *s3manager.UploadInput,
	funC ...func(*s3manager.Uploader)) (*s3manager.UploadOutput, error) {
	return s.Upload(uploadInput, funC...)
}

func (s *S3ManagerMock) GetFile(key string) (string, error) {
	content, ok := s.S3ManagerMockFileSystem[key]
	if !ok {
		return "", errors.New("does not exists")
	}
	return content, nil
}
