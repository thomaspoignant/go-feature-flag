package s3retrieverv2_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"log/slog"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thomaspoignant/go-feature-flag/retriever"
	"github.com/thomaspoignant/go-feature-flag/retriever/s3retrieverv2"
	"github.com/thomaspoignant/go-feature-flag/testutils"
	"github.com/thomaspoignant/go-feature-flag/utils/fflog"
)

const (
	yamlA = "flagA:\n  variations:\n    on: true\n    off: false\n  defaultRule:\n    variation: on\n"
	yamlB = "flagB:\n  variations:\n    on: true\n    off: false\n  defaultRule:\n    variation: off\n"
	jsonC = `{"flagC": {"variations": {"on": true, "off": false}, "defaultRule": {"variation": "on"}}}`
	tomlD = "[flagD]\n[flagD.variations]\non = true\noff = false\n[flagD.defaultRule]\nvariation = \"on\"\n"
)

// cfg holds the settings of a retriever (Retriever contains a mutex, so it can't be copied).
type cfg struct {
	Item      string
	Prefix    string
	Pattern   string
	Recursive bool
	// noInit skips Init, noLister skips SetLister.
	noInit   bool
	noLister bool
}

func (c cfg) build() *s3retrieverv2.Retriever {
	return &s3retrieverv2.Retriever{
		Bucket: "b", Item: c.Item, Prefix: c.Prefix, Pattern: c.Pattern, Recursive: c.Recursive,
	}
}

// step is one Retrieve call (or several concurrent ones), preceded by a mutation of the bucket.
type step struct {
	mutate        func(m *testutils.S3PrefixMock)
	nilCtx        bool
	concurrent    int      // number of concurrent Retrieve calls, 0 means one call
	wantFlags     []string // flag names (prefix mode: parsed from JSON; item mode: substrings of the raw file)
	wantContains  []string // substrings expected in the raw output
	wantLog       string   // substring expected in the logs
	wantErr       string
	wantDownloads map[string]int // cumulative number of downloads per key
}

func TestRetriever_Retrieve(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		setup func(m *testutils.S3PrefixMock)
		r     cfg
		steps []step
	}{
		{
			name:  "item mode unchanged",
			setup: func(m *testutils.S3PrefixMock) { m.Put("flags/a.yaml", yamlA) },
			r:     cfg{Item: "flags/a.yaml"},
			steps: []step{
				{wantFlags: []string{"flagA"}},
				// item mode has no ETag cache: every call downloads
				{wantFlags: []string{"flagA"}, nilCtx: true, wantDownloads: map[string]int{"flags/a.yaml": 2}},
			},
		},
		{
			name:  "item mode download error",
			setup: func(m *testutils.S3PrefixMock) {},
			r:     cfg{Item: "missing.yaml"},
			steps: []step{{wantErr: "unable to download item from S3 \"missing.yaml\""}},
		},
		{
			name:  "not initialized",
			setup: func(m *testutils.S3PrefixMock) {},
			r:     cfg{Prefix: "f/", noInit: true, noLister: true},
			steps: []step{{wantErr: "lister is not initialized"}},
		},
		{
			name: "prefix non recursive",
			setup: func(m *testutils.S3PrefixMock) {
				m.Put("flags/a.yaml", yamlA)
				m.Put("flags/b.yaml", yamlB)
				m.Put("flags/sub/c.yaml", "flagC: {}\n")
				m.Put("flags/", "") // folder marker
				m.Put("other/d.yaml", "flagD: {}\n")
			},
			r:     cfg{Prefix: "flags/"},
			steps: []step{{wantFlags: []string{"flagA", "flagB"}}},
		},
		{
			name: "prefix without trailing slash non recursive",
			setup: func(m *testutils.S3PrefixMock) {
				m.Put("flags/a.yaml", yamlA)
				m.Put("flags/sub/c.yaml", "flagC: {}\n")
				m.Put("flags-old/z.yaml", "flagZ: {}\n")
			},
			r:     cfg{Prefix: "flags"},
			steps: []step{{wantFlags: []string{"flagA"}}},
		},
		{
			name: "prefix recursive",
			setup: func(m *testutils.S3PrefixMock) {
				m.Put("flags/a.yaml", yamlA)
				m.Put("flags/sub/b.yml", yamlB)
				m.Put("flags/sub/deep/", "")
				m.Put("other/d.yaml", "flagD: {}\n")
			},
			r:     cfg{Prefix: "flags/", Recursive: true},
			steps: []step{{wantFlags: []string{"flagA", "flagB"}}},
		},
		{
			name: "pattern filter",
			setup: func(m *testutils.S3PrefixMock) {
				m.Put("flags/a.yaml", yamlA)
				m.Put("flags/notes.txt", "not a flag file")
				m.Put("flags/sub/b.yaml", yamlB)
			},
			r:     cfg{Prefix: "flags/", Pattern: "*.yaml", Recursive: true},
			steps: []step{{wantFlags: []string{"flagA", "flagB"}}},
		},
		{
			name:  "json files",
			setup: func(m *testutils.S3PrefixMock) { m.Put("f/c.json", jsonC) },
			r:     cfg{Prefix: "f/"},
			steps: []step{{wantFlags: []string{"flagC"}}},
		},
		{
			name:  "toml files",
			setup: func(m *testutils.S3PrefixMock) { m.Put("f/d.toml", tomlD) },
			r:     cfg{Prefix: "f/"},
			steps: []step{{wantFlags: []string{"flagD"}}},
		},
		{
			name: "yaml with non string keys behaves like item mode",
			setup: func(m *testutils.S3PrefixMock) {
				m.Put("f/k.yaml", "flagK:\n  variations:\n    true: x\n    1: y\n  defaultRule:\n    variation: x\n")
			},
			r:     cfg{Prefix: "f/"},
			steps: []step{{wantFlags: []string{"flagK"}}},
		},
		{
			name: "large integers keep their precision",
			setup: func(m *testutils.S3PrefixMock) {
				m.Put("f/big.yaml",
					"flagBig:\n  variations:\n    big: 9007199254740993\n  defaultRule:\n    variation: big\n")
			},
			r: cfg{Prefix: "f/"},
			steps: []step{{
				wantFlags:    []string{"flagBig"},
				wantContains: []string{"9007199254740993"},
			}},
		},
		{
			name: "etag cache hit does not redownload",
			setup: func(m *testutils.S3PrefixMock) {
				m.Put("f/a.yaml", yamlA)
				m.Put("f/b.yaml", yamlB)
			},
			r: cfg{Prefix: "f/"},
			steps: []step{
				{wantFlags: []string{"flagA", "flagB"}},
				{
					wantFlags:     []string{"flagA", "flagB"},
					wantDownloads: map[string]int{"f/a.yaml": 1, "f/b.yaml": 1},
				},
			},
		},
		{
			name: "changed etag and new file are downloaded, others are not",
			setup: func(m *testutils.S3PrefixMock) {
				m.Put("f/a.yaml", yamlA)
				m.Put("f/b.yaml", yamlB)
			},
			r: cfg{Prefix: "f/"},
			steps: []step{
				{wantFlags: []string{"flagA", "flagB"}},
				{
					mutate: func(m *testutils.S3PrefixMock) {
						m.Put("f/a.yaml", "flagA2:\n  defaultRule:\n    variation: on\n")
						m.Put("f/c.yaml", "flagC: {}\n")
					},
					wantFlags: []string{"flagA2", "flagB", "flagC"},
					wantDownloads: map[string]int{
						"f/a.yaml": 2, "f/b.yaml": 1, "f/c.yaml": 1,
					},
				},
			},
		},
		{
			name: "deleted file disappears",
			setup: func(m *testutils.S3PrefixMock) {
				m.Put("f/a.yaml", yamlA)
				m.Put("f/b.yaml", yamlB)
			},
			r: cfg{Prefix: "f/"},
			steps: []step{
				{wantFlags: []string{"flagA", "flagB"}},
				{
					mutate:        func(m *testutils.S3PrefixMock) { m.Delete("f/b.yaml") },
					wantFlags:     []string{"flagA"},
					wantDownloads: map[string]int{"f/a.yaml": 1, "f/b.yaml": 1},
				},
				{
					// file comes back: it must be downloaded again as the cache entry was dropped
					mutate:        func(m *testutils.S3PrefixMock) { m.Put("f/b.yaml", yamlB) },
					wantFlags:     []string{"flagA", "flagB"},
					wantDownloads: map[string]int{"f/a.yaml": 1, "f/b.yaml": 2},
				},
			},
		},
		{
			name: "duplicate flag key",
			setup: func(m *testutils.S3PrefixMock) {
				m.Put("f/a.yaml", yamlA)
				m.Put("f/z.yaml", yamlA)
			},
			r:     cfg{Prefix: "f/"},
			steps: []step{{wantErr: `duplicate flag "flagA" found in files "f/a.yaml" and "f/z.yaml"`}},
		},
		{
			name: "mixed formats",
			setup: func(m *testutils.S3PrefixMock) {
				m.Put("f/a.yaml", yamlA)
				m.Put("f/c.json", jsonC)
			},
			r:     cfg{Prefix: "f/"},
			steps: []step{{wantErr: `"f/a.yaml" is yaml but "f/c.json" is json`}},
		},
		{
			name:  "unsupported extension",
			setup: func(m *testutils.S3PrefixMock) { m.Put("f/readme.txt", "hello") },
			r:     cfg{Prefix: "f/"},
			steps: []step{{wantErr: `unsupported file extension for "f/readme.txt"`}},
		},
		{
			name:  "invalid content",
			setup: func(m *testutils.S3PrefixMock) { m.Put("f/a.json", "{not json") },
			r:     cfg{Prefix: "f/"},
			steps: []step{{wantErr: `unable to parse file "f/a.json"`}},
		},
		{
			name: "empty files are empty flag sets for all formats",
			setup: func(m *testutils.S3PrefixMock) {
				m.Put("j/empty.json", "")
				m.Put("y/empty.yaml", "  \n")
				m.Put("t/empty.toml", "\n\n")
			},
			r: cfg{Prefix: "j/"},
			steps: []step{
				{wantFlags: []string{}},
				{
					mutate: func(m *testutils.S3PrefixMock) {
						m.Delete("j/empty.json")
						m.Put("j/a.yaml", yamlA)
						m.Put("j/empty.yaml", " ")
					},
					wantFlags: []string{"flagA"},
				},
				{
					mutate: func(m *testutils.S3PrefixMock) {
						m.Delete("j/a.yaml")
						m.Delete("j/empty.yaml")
						m.Put("j/empty.toml", "\n")
					},
					wantFlags: []string{},
				},
			},
		},
		{
			name:  "empty result logs a warning",
			setup: func(m *testutils.S3PrefixMock) { m.Put("other/a.yaml", yamlA) },
			r:     cfg{Prefix: "f/"},
			steps: []step{{wantFlags: []string{}, wantLog: "no file matches the prefix"}},
		},
		{
			name: "pagination",
			setup: func(m *testutils.S3PrefixMock) {
				m.PageSize = 2
				m.Put("f/a.yaml", yamlA)
				m.Put("f/b.yaml", yamlB)
				m.Put("f/c.yaml", "flagC: {}\n")
				m.Put("f/d.yaml", "flagD: {}\n")
				m.Put("f/e.yaml", "flagE: {}\n")
			},
			r: cfg{Prefix: "f/"},
			steps: []step{
				{wantFlags: []string{"flagA", "flagB", "flagC", "flagD", "flagE"}},
			},
		},
		{
			name:  "list error",
			setup: func(m *testutils.S3PrefixMock) { m.ListErr = errors.New("boom") },
			r:     cfg{Prefix: "f/"},
			steps: []step{{wantErr: `unable to list objects from S3 with prefix "f/", boom`}},
		},
		{
			name: "download error",
			setup: func(m *testutils.S3PrefixMock) {
				m.Put("f/a.yaml", yamlA)
				m.DownloadErr = errors.New("kaboom")
			},
			r:     cfg{Prefix: "f/"},
			steps: []step{{wantErr: `unable to download item from S3 "f/a.yaml", kaboom`}},
		},
		{
			name: "partial download failure keeps the files already fetched",
			setup: func(m *testutils.S3PrefixMock) {
				m.Put("f/a.yaml", yamlA)
				m.Put("f/b.yaml", yamlB)
				m.SetDownloadErrKey("f/b.yaml", errors.New("kaboom"))
			},
			r: cfg{Prefix: "f/"},
			steps: []step{
				{wantErr: `unable to download item from S3 "f/b.yaml", kaboom`},
				{
					mutate:        func(m *testutils.S3PrefixMock) { m.SetDownloadErrKey("f/b.yaml", nil) },
					wantFlags:     []string{"flagA", "flagB"},
					wantDownloads: map[string]int{"f/a.yaml": 1, "f/b.yaml": 2},
				},
			},
		},
		{
			name:  "failed retrieve does not corrupt the cache",
			setup: func(m *testutils.S3PrefixMock) { m.Put("f/a.yaml", yamlA) },
			r:     cfg{Prefix: "f/"},
			steps: []step{
				{wantFlags: []string{"flagA"}},
				{
					mutate:  func(m *testutils.S3PrefixMock) { m.Put("f/a.yaml", "{ bad yaml") },
					wantErr: `unable to parse file "f/a.yaml"`,
				},
				{
					mutate: func(m *testutils.S3PrefixMock) { m.Put("f/a.yaml", yamlA) },
					// same ETag as the cached good version: no new download is needed
					wantFlags:     []string{"flagA"},
					wantDownloads: map[string]int{"f/a.yaml": 2},
				},
			},
		},
		{
			name: "concurrent calls are serialized",
			setup: func(m *testutils.S3PrefixMock) {
				m.Put("f/a.yaml", yamlA)
				m.Put("f/b.yaml", yamlB)
			},
			r: cfg{Prefix: "f/"},
			steps: []step{{
				concurrent:    20,
				wantFlags:     []string{"flagA", "flagB"},
				wantDownloads: map[string]int{"f/a.yaml": 1, "f/b.yaml": 1},
			}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			m := &testutils.S3PrefixMock{}
			tt.setup(m)
			r := tt.r.build()
			r.SetDownloader(m)
			if !tt.r.noLister {
				r.SetLister(m)
			}
			logs := &bytes.Buffer{}
			logger := &fflog.FFLogger{LeveledLogger: slog.New(slog.NewTextHandler(logs, nil))}
			if !tt.r.noInit {
				require.NoError(t, r.Init(t.Context(), logger))
			}
			t.Cleanup(func() { _ = r.Shutdown(t.Context()) })

			for i, st := range tt.steps {
				if st.mutate != nil {
					st.mutate(m)
				}
				calls := max(st.concurrent, 1)
				var wg sync.WaitGroup
				results := make([][]byte, calls)
				errs := make([]error, calls)
				for c := range calls {
					wg.Go(func() {
						if st.nilCtx {
							results[c], errs[c] = r.Retrieve(nil) //nolint:staticcheck // nil ctx is supported
							return
						}
						results[c], errs[c] = r.Retrieve(t.Context())
					})
				}
				wg.Wait()

				for c := range calls {
					if st.wantErr != "" {
						require.ErrorContains(t, errs[c], st.wantErr, "step %d", i)
						continue
					}
					require.NoError(t, errs[c], "step %d", i)
					checkOutput(t, tt.r.Prefix != "", st, results[c], i)
				}
				for k, n := range st.wantDownloads {
					assert.Equal(t, n, m.DownloadCalls(k), "downloads of %s at step %d", k, i)
				}
				if st.wantLog != "" {
					assert.Contains(t, logs.String(), st.wantLog, "step %d", i)
				}
			}
		})
	}
}

func checkOutput(t *testing.T, prefixMode bool, st step, got []byte, i int) {
	t.Helper()
	if prefixMode {
		var flags map[string]any
		require.NoError(t, json.Unmarshal(got, &flags), "step %d", i)
		keys := make([]string, 0, len(flags))
		for k := range flags {
			keys = append(keys, k)
		}
		assert.ElementsMatch(t, st.wantFlags, keys, "step %d", i)
	} else {
		// item mode returns the raw file
		for _, f := range st.wantFlags {
			assert.Contains(t, string(got), f, "step %d", i)
		}
	}
	for _, c := range st.wantContains {
		assert.Contains(t, string(got), c, "step %d", i)
	}
}

func TestRetriever_Init(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		r    cfg
		// noDownloader skips SetDownloader, so Init must build a real client.
		wantErr string
		// wantAwsConfigNil: item mode with an injected downloader must not load the AWS config.
		wantAwsConfigNil bool
	}{
		{name: "item only", r: cfg{Item: "a.yaml"}, wantAwsConfigNil: true},
		{name: "prefix only", r: cfg{Prefix: "f/"}, wantAwsConfigNil: true},
		{name: "prefix with pattern", r: cfg{Prefix: "f/", Pattern: "*.yaml"}, wantAwsConfigNil: true},
		{
			name:    "item and prefix",
			r:       cfg{Item: "a.yaml", Prefix: "f/"},
			wantErr: "mutually exclusive",
		},
		{
			name:    "neither item nor prefix",
			r:       cfg{},
			wantErr: "one of Item or Prefix is required",
		},
		{
			name:    "item with pattern",
			r:       cfg{Item: "a.yaml", Pattern: "*.yaml"},
			wantErr: "can only be used with Prefix",
		},
		{
			name:    "item with recursive",
			r:       cfg{Item: "a.yaml", Recursive: true},
			wantErr: "can only be used with Prefix",
		},
		{
			name:    "invalid pattern",
			r:       cfg{Prefix: "f/", Pattern: "[a-"},
			wantErr: "invalid Pattern",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			r := tt.r.build()
			m := &testutils.S3PrefixMock{}
			r.SetDownloader(m)
			r.SetLister(m)
			err := r.Init(t.Context(), nil)
			if tt.wantErr != "" {
				require.ErrorContains(t, err, tt.wantErr)
				assert.Equal(t, retriever.RetrieverError, r.Status())
				return
			}
			require.NoError(t, err)
			assert.Equal(t, retriever.RetrieverReady, r.Status())
			if tt.wantAwsConfigNil {
				assert.Nil(t, r.AwsConfig)
			}
		})
	}

	t.Run("item mode with only a downloader does not need a lister", func(t *testing.T) {
		t.Parallel()
		r := cfg{Item: "a.yaml"}.build()
		r.SetDownloader(&testutils.S3PrefixMock{})
		require.NoError(t, r.Init(t.Context(), nil))
		assert.Nil(t, r.AwsConfig)
	})
}

func TestRetriever_Shutdown(t *testing.T) {
	t.Parallel()
	m := &testutils.S3PrefixMock{}
	m.Put("f/a.yaml", yamlA)
	r := cfg{Prefix: "f/"}.build()
	r.SetDownloader(m)
	r.SetLister(m)
	require.NoError(t, r.Init(t.Context(), nil))
	_, err := r.Retrieve(t.Context())
	require.NoError(t, err)

	require.NoError(t, r.Shutdown(t.Context()))
	assert.Equal(t, retriever.RetrieverNotReady, r.Status())
	_, err = r.Retrieve(t.Context())
	require.ErrorContains(t, err, "downloader is not initialized")
	assert.Equal(t, retriever.RetrieverError, r.Status())
}

func TestRetriever_SetLister(t *testing.T) {
	t.Parallel()
	m := &testutils.S3PrefixMock{}
	m.Put("f/a.yaml", yamlA)
	r := cfg{Prefix: "f/"}.build()
	r.SetDownloader(m)
	r.SetLister(m)
	got, err := r.Retrieve(t.Context())
	require.NoError(t, err)
	assert.Contains(t, string(got), "flagA")
	assert.Equal(t, 1, m.ListCalls())
}

func TestRetriever_SetDownloader(t *testing.T) {
	t.Parallel()
	m := &testutils.S3PrefixMock{}
	m.Put("a.yaml", yamlA)
	r := cfg{Item: "a.yaml"}.build()
	r.SetDownloader(m)
	got, err := r.Retrieve(t.Context())
	require.NoError(t, err)
	assert.Equal(t, yamlA, string(got))
	assert.Equal(t, 1, m.DownloadCalls("a.yaml"))
}

func TestRetriever_OutputFormat(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		r    cfg
		want string
	}{
		{name: "item mode keeps the global format", r: cfg{Item: "a.yaml"}, want: ""},
		{name: "prefix mode is json", r: cfg{Prefix: "f/"}, want: "json"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, tt.r.build().OutputFormat())
		})
	}
}
