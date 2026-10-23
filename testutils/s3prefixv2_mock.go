package testutils

import (
	"context"
	"crypto/md5" //nolint:gosec // only used to fake an ETag
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/s3/transfermanager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// S3PrefixMock is an in-memory fake of a S3 bucket implementing both the lister and the
// downloader APIs used by the S3 retriever v2. It supports prefix/delimiter semantics,
// pagination, ETags and call counters.
type S3PrefixMock struct {
	// PageSize is the max number of keys returned by one ListObjectsV2 call (default 1000).
	PageSize int
	// ListErr, if set, is returned by every ListObjectsV2 call.
	ListErr error
	// DownloadErr, if set, is returned by every DownloadObject call.
	DownloadErr error
	// DownloadErrKeys makes the download of the given keys fail. Guarded by the mock, use SetDownloadErrKey.
	downloadErrKeys map[string]error

	mu            sync.Mutex
	objects       map[string][]byte
	listCalls     int
	downloadCalls map[string]int
}

// Put creates or replaces an object. The ETag is derived from the content.
func (m *S3PrefixMock) Put(key, content string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.objects == nil {
		m.objects = map[string][]byte{}
	}
	m.objects[key] = []byte(content)
}

// SetDownloadErrKey makes the download of one key fail with err, nil removes the failure.
func (m *S3PrefixMock) SetDownloadErrKey(key string, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.downloadErrKeys == nil {
		m.downloadErrKeys = map[string]error{}
	}
	if err == nil {
		delete(m.downloadErrKeys, key)
		return
	}
	m.downloadErrKeys[key] = err
}

// Delete removes an object.
func (m *S3PrefixMock) Delete(key string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.objects, key)
}

// ListCalls returns the number of ListObjectsV2 calls.
func (m *S3PrefixMock) ListCalls() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.listCalls
}

// DownloadCalls returns the number of times the key was downloaded.
func (m *S3PrefixMock) DownloadCalls(key string) int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.downloadCalls[key]
}

// TotalDownloadCalls returns the number of downloads, all keys included.
func (m *S3PrefixMock) TotalDownloadCalls() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	total := 0
	for _, c := range m.downloadCalls {
		total += c
	}
	return total
}

func etagOf(content []byte) string {
	return fmt.Sprintf("\"%x\"", md5.Sum(content)) //nolint:gosec // fake ETag
}

// ListObjectsV2 mimics the S3 API (prefix, delimiter, continuation token, max keys).
func (m *S3PrefixMock) ListObjectsV2(
	_ context.Context,
	params *s3.ListObjectsV2Input,
	_ ...func(*s3.Options),
) (*s3.ListObjectsV2Output, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.listCalls++
	if m.ListErr != nil {
		return nil, m.ListErr
	}
	prefix := aws.ToString(params.Prefix)
	delimiter := aws.ToString(params.Delimiter)
	after := aws.ToString(params.ContinuationToken)

	pageSize := m.PageSize
	if pageSize <= 0 {
		pageSize = 1000
	}
	out := &s3.ListObjectsV2Output{IsTruncated: aws.Bool(false)}
	seenPrefixes := map[string]bool{}
	count := 0
	for _, k := range m.sortedKeys(prefix) {
		if k <= after {
			continue
		}
		if count == pageSize {
			out.IsTruncated = aws.Bool(true)
			break
		}
		count++
		out.NextContinuationToken = aws.String(k)
		if cp := commonPrefix(k, prefix, delimiter); cp != "" {
			if !seenPrefixes[cp] {
				seenPrefixes[cp] = true
				out.CommonPrefixes = append(out.CommonPrefixes, s3types.CommonPrefix{Prefix: aws.String(cp)})
			}
			continue
		}
		out.Contents = append(out.Contents, s3types.Object{
			Key:  aws.String(k),
			ETag: aws.String(etagOf(m.objects[k])),
		})
	}
	if !aws.ToBool(out.IsTruncated) {
		out.NextContinuationToken = nil
	}
	out.KeyCount = aws.Int32(int32(count)) //nolint:gosec // small test values
	return out, nil
}

// sortedKeys returns the sorted keys starting with prefix.
func (m *S3PrefixMock) sortedKeys(prefix string) []string {
	keys := make([]string, 0, len(m.objects))
	for k := range m.objects {
		if strings.HasPrefix(k, prefix) {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)
	return keys
}

// commonPrefix returns the "folder" the key is rolled up into when a delimiter is used, or "".
func commonPrefix(key, prefix, delimiter string) string {
	if delimiter == "" {
		return ""
	}
	rest := strings.TrimPrefix(key, prefix)
	i := strings.Index(rest, delimiter)
	if i < 0 {
		return ""
	}
	return prefix + rest[:i+len(delimiter)]
}

// DownloadObject writes the content of the object in the WriterAt of the input.
func (m *S3PrefixMock) DownloadObject(
	_ context.Context,
	input *transfermanager.DownloadObjectInput,
	_ ...func(*transfermanager.Options),
) (*transfermanager.DownloadObjectOutput, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.downloadCalls == nil {
		m.downloadCalls = map[string]int{}
	}
	key := aws.ToString(input.Key)
	m.downloadCalls[key]++
	if m.DownloadErr != nil {
		return nil, m.DownloadErr
	}
	if err := m.downloadErrKeys[key]; err != nil {
		return nil, err
	}
	if input.WriterAt == nil {
		return nil, errors.New("WriterAt is required")
	}
	content, ok := m.objects[key]
	if !ok {
		return nil, fmt.Errorf("NoSuchKey: %s", key)
	}
	_, _ = input.WriterAt.WriteAt(content, 0)
	return &transfermanager.DownloadObjectOutput{}, nil
}
