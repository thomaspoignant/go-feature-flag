package testutils

import (
	"fmt"
	"io"
	"io/ioutil"
)

type GCPStorageReaderMock struct {
	ShouldFail bool
	data       []byte
	readIndex  int64
}

func (r *GCPStorageReaderMock) Close() error {
	return nil
}

func (r *GCPStorageReaderMock) Read(p []byte) (n int, err error) {
	if r.data == nil {
		r.data, err = ioutil.ReadFile("./testdata/flag-config.yaml")
		if err != nil {
			return 0, err
		}
	}

	if r.ShouldFail {
		return 0, fmt.Errorf("failed to read from GCP")
	}

	if r.readIndex >= int64(len(r.data)) {
		err = io.EOF
		return
	}

	n = copy(p, r.data[r.readIndex:])
	r.readIndex += int64(n)
	return
}
