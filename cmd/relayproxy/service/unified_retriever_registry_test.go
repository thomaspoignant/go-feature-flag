package service

import (
	"context"
	"errors"
	"os"
	"strings"
	"sync/atomic"
	"testing"
	"testing/synctest"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/s3/transfermanager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/config"
	"github.com/thomaspoignant/go-feature-flag/cmdhelpers/retrieverconf"
	"github.com/thomaspoignant/go-feature-flag/retriever/s3retrieverv2"
	"go.uber.org/zap"
)

// fakeS3 serves the same flag file for every key and counts the requests.
type fakeS3 struct {
	content []byte
	listErr error
	lists   atomic.Int32
	gets    atomic.Int32
}

func (f *fakeS3) ListObjectsV2(
	_ context.Context, in *s3.ListObjectsV2Input, _ ...func(*s3.Options),
) (*s3.ListObjectsV2Output, error) {
	f.lists.Add(1)
	if f.listErr != nil {
		return nil, f.listErr
	}
	out := &s3.ListObjectsV2Output{}
	for _, k := range []string{"p/a.yaml", "p/b.yaml", "p/c.yaml"} {
		if strings.HasPrefix(k, aws.ToString(in.Prefix)) {
			out.Contents = append(out.Contents, s3types.Object{Key: aws.String(k), ETag: aws.String(`"x"`)})
		}
	}
	return out, nil
}

func (f *fakeS3) DownloadObject(
	_ context.Context, in *transfermanager.DownloadObjectInput, _ ...func(*transfermanager.Options),
) (*transfermanager.DownloadObjectOutput, error) {
	f.gets.Add(1)
	if _, err := in.WriterAt.WriteAt(f.content, 0); err != nil {
		return nil, err
	}
	return &transfermanager.DownloadObjectOutput{ETag: aws.String(`"x"`)}, nil
}

func newFakeRegistry(t *testing.T, f *fakeS3, created *atomic.Int32) *unifiedRetrieverRegistry {
	t.Helper()
	return newUnifiedRetrieverRegistry(func(ctx context.Context, bucket string) (*s3retrieverv2.SharedItems, error) {
		created.Add(1)
		return s3retrieverv2.NewSharedItems(ctx, s3retrieverv2.SharedItemsOptions{
			Bucket: bucket, Lister: f, Downloader: f,
		})
	})
}

func s3FlagSet(name, bucket, item string) config.FlagSet {
	return config.FlagSet{
		Name:    name,
		APIKeys: []string{name + "-key"},
		CommonFlagSet: config.CommonFlagSet{
			Retriever:           &retrieverconf.RetrieverConf{Kind: retrieverconf.S3Retriever, Bucket: bucket, Item: item},
			UseUnifiedRetriever: true,
			PollingInterval:     60000,
		},
	}
}

func Test_unifiedRetrieverRegistry_retrieverFor(t *testing.T) {
	t.Parallel()
	t.Run("same bucket shares a group, other bucket has its own", func(t *testing.T) {
		t.Parallel()
		var created atomic.Int32
		reg := newFakeRegistry(t, &fakeS3{}, &created)
		fs1, fs2, fs3 := s3FlagSet("a", "b1", "p/a.yaml"), s3FlagSet("b", "b1", "p/b.yaml"),
			s3FlagSet("c", "b2", "p/c.yaml")
		r1, err := reg.retrieverFor(&fs1)
		require.NoError(t, err)
		_, err = reg.retrieverFor(&fs2)
		require.NoError(t, err)
		_, err = reg.retrieverFor(&fs3)
		require.NoError(t, err)
		assert.EqualValues(t, 2, created.Load())
		assert.Equal(t, 2, reg.groupCount())

		// the group is dropped when its last retriever is shut down (Shutdown twice is harmless)
		require.NoError(t, r1.(*releasingRetriever).Shutdown(t.Context()))
		require.NoError(t, r1.(*releasingRetriever).Shutdown(t.Context()))
		assert.Equal(t, 2, reg.groupCount())
	})

	t.Run("group is dropped after the last shutdown", func(t *testing.T) {
		t.Parallel()
		var created atomic.Int32
		reg := newFakeRegistry(t, &fakeS3{}, &created)
		fs1, fs2 := s3FlagSet("a", "b1", "p/a.yaml"), s3FlagSet("b", "b1", "p/b.yaml")
		r1, err := reg.retrieverFor(&fs1)
		require.NoError(t, err)
		r2, err := reg.retrieverFor(&fs2)
		require.NoError(t, err)
		require.NoError(t, r1.(*releasingRetriever).Shutdown(t.Context()))
		assert.Equal(t, 1, reg.groupCount())
		require.NoError(t, r2.(*releasingRetriever).Shutdown(t.Context()))
		assert.Equal(t, 0, reg.groupCount())
	})

	t.Run("unsupported configurations", func(t *testing.T) {
		t.Parallel()
		reg := newFakeRegistry(t, &fakeS3{}, &atomic.Int32{})
		folder := s3FlagSet("a", "b1", "")
		other := s3FlagSet("a", "b1", "x")
		other.Retriever.Kind = retrieverconf.HTTPRetriever
		multi := s3FlagSet("a", "b1", "x")
		multi.Retrievers = &[]retrieverconf.RetrieverConf{*multi.Retriever}
		for _, fs := range []config.FlagSet{folder, other, multi} {
			_, err := reg.retrieverFor(&fs)
			assert.Error(t, err)
		}
		assert.Equal(t, 0, reg.groupCount())
	})
}

func TestFlagsetManager_UnifiedRetriever(t *testing.T) {
	t.Parallel()
	synctest.Test(t, func(t *testing.T) {
		testUnifiedManager(t)
	})
}

func testUnifiedManager(t *testing.T) {
	t.Helper()
	content, err := os.ReadFile("../testdata/goff/configuration_flags.yaml")
	require.NoError(t, err)

	f := &fakeS3{content: content}
	var created atomic.Int32
	reg := newFakeRegistry(t, f, &created)
	nonUnified := config.FlagSet{
		Name:    "isolated",
		APIKeys: []string{"isolated-key"},
		CommonFlagSet: config.CommonFlagSet{
			Retriever: &retrieverconf.RetrieverConf{Kind: retrieverconf.FileRetriever, Path: "../testdata/goff/configuration_flags.yaml"},
		},
	}
	cfg := &config.Config{FlagSets: []config.FlagSet{
		s3FlagSet("u1", "flags", "p/a.yaml"),
		s3FlagSet("u2", "flags", "p/b.yaml"),
		nonUnified,
	}}
	mgr, err := newFlagsetManagerWithFlagsetsAndRegistry(cfg, zap.NewNop(), nil, nil, nil, reg)
	require.NoError(t, err)

	// two unified flagsets on the same bucket share one group, the non unified one does not use any
	assert.EqualValues(t, 1, created.Load())
	assert.Equal(t, 1, reg.groupCount())
	// all the unified items are pre-registered, so startup lists the directory once
	assert.EqualValues(t, 1, f.lists.Load())
	assert.EqualValues(t, 2, f.gets.Load())

	// a flagset added at runtime joins the same group
	impl := mgr.(*flagsetManagerImpl)
	impl.addFlagset(s3FlagSet("u3", "flags", "p/c.yaml"))
	assert.EqualValues(t, 1, created.Load())
	assert.Equal(t, 1, reg.groupCount())
	all, err := mgr.AllFlagSets()
	require.NoError(t, err)
	assert.Contains(t, all, "u3")

	// removing flagsets releases the group once the last one is gone
	impl.removeFlagset(s3FlagSet("u1", "flags", "p/a.yaml"))
	impl.removeFlagset(s3FlagSet("u2", "flags", "p/b.yaml"))
	assert.Equal(t, 1, reg.groupCount())
	mgr.Close()
	assert.Equal(t, 0, reg.groupCount())
}

func TestFlagsetManager_UnifiedRetriever_PreRegister(t *testing.T) {
	t.Parallel()
	var created atomic.Int32
	reg := newFakeRegistry(t, &fakeS3{}, &created)
	fs1, fs2 := s3FlagSet("a", "b1", "p/a.yaml"), s3FlagSet("b", "b1", "q/b.yaml")
	require.NoError(t, reg.preRegister(&fs1))
	require.NoError(t, reg.preRegister(&fs2))
	assert.EqualValues(t, 1, created.Load())

	// claimed retrievers are the pre-registered ones, no additional reference is taken
	r1, err := reg.retrieverFor(&fs1)
	require.NoError(t, err)
	require.NoError(t, r1.(*releasingRetriever).Shutdown(t.Context()))
	assert.Equal(t, 1, reg.groupCount(), "fs2 is still pending")

	// the unclaimed ones are released
	reg.releasePending()
	assert.Equal(t, 0, reg.groupCount())

	bad := s3FlagSet("c", "b1", "")
	assert.Error(t, reg.preRegister(&bad))
}

func Test_newGoFeatureFlagClient_ReleasesUnifiedRetrieverOnFailure(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		s3   *fakeS3
		mod  func(fs *config.FlagSet)
	}{
		{
			name: "invalid exporter",
			s3:   &fakeS3{content: []byte("{}")},
			mod: func(fs *config.FlagSet) {
				fs.Exporter = &config.ExporterConf{Kind: "unknown"}
			},
		},
		{
			name: "first retrieve fails",
			s3:   &fakeS3{listErr: errors.New("AccessDenied")},
			mod:  func(*config.FlagSet) {},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			synctest.Test(t, func(t *testing.T) {
				reg := newFakeRegistry(t, tt.s3, &atomic.Int32{})
				fs := s3FlagSet("a", "b1", "p/a.yaml")
				tt.mod(&fs)
				_, err := newGoFeatureFlagClient(&fs, zap.NewNop(), nil, withUnifiedRetrieverRegistry(reg))
				require.Error(t, err)
				assert.Equal(t, 0, reg.groupCount(), "the retriever must have left its group")
			})
		})
	}
}
