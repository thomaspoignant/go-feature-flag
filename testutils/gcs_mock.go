package testutils

import (
	"crypto/md5" //nolint: gosec
	"encoding/base64"
	"os"
	"testing"

	"github.com/fsouza/fake-gcs-server/fakestorage"
)

type MockedStorage struct {
	Server *fakestorage.Server
}

func NewMockedGCS(t *testing.T) MockedStorage {
	server := fakestorage.NewServer(nil)
	t.Cleanup(func() {
		server.Stop()
	})

	return MockedStorage{
		Server: server,
	}
}

func (m MockedStorage) WithFiles(t *testing.T, bucketName string, files map[string]string) {
	for filename, name := range files {
		content, err := os.ReadFile(filename)
		if err != nil {
			t.Fatalf("could not read testfile: %v", err)
		}

		object := fakestorage.Object{
			Content: content,
			ObjectAttrs: fakestorage.ObjectAttrs{
				BucketName: bucketName,
				Name:       name,
				Md5Hash:    EncodedMd5Hash(content),
			},
		}
		m.Server.CreateObject(object)
	}
}

func EncodedMd5Hash(content []byte) string {
	h := md5.New() //nolint: gosec
	h.Write(content)

	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}
