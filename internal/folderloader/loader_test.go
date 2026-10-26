package folderloader_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thomaspoignant/go-feature-flag/internal/cache"
	"github.com/thomaspoignant/go-feature-flag/internal/folderloader"
	"github.com/thomaspoignant/go-feature-flag/modules/core/dto"
	"github.com/thomaspoignant/go-feature-flag/utils/fflog"
)

const (
	yamlA = "flagA:\n  variations:\n    on: true\n    off: false\n  defaultRule:\n    variation: on\n"
	yamlB = "flagB:\n  variations:\n    on: true\n    off: false\n  defaultRule:\n    variation: off\n"
	jsonC = `{"flagC": {"variations": {"on": true, "off": false}, "defaultRule": {"variation": "on"}}}`
	tomlD = "[flagD]\n[flagD.variations]\non = true\noff = false\n[flagD.defaultRule]\nvariation = \"on\"\n"
)

// fakeStore is an in-memory Store. The version of a file is derived from its content.
type fakeStore struct {
	mu sync.Mutex
	// files are the objects by key.
	files map[string]string
	// noVersion lists the files with an empty version, and Get returns an empty one too.
	noVersion bool
	// getNoVersion makes Get return an empty version, the listed one is used then.
	getNoVersion bool
	// listVersion / getVersion, if set, replace the version of every file in List / Get.
	listVersion string
	getVersion  string
	listErr     error
	getErr      error
	getErrKeys  map[string]error
	// beforeGet is called at the start of Get, without the lock held.
	beforeGet func(key string)

	getCalls     map[string]int
	listPrefixes []string
	listRecurse  []bool
}

func (f *fakeStore) put(key, content string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.files == nil {
		f.files = map[string]string{}
	}
	f.files[key] = content
}

func (f *fakeStore) del(key string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.files, key)
}

func (f *fakeStore) setGetErrKey(key string, err error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.getErrKeys == nil {
		f.getErrKeys = map[string]error{}
	}
	if err == nil {
		delete(f.getErrKeys, key)
		return
	}
	f.getErrKeys[key] = err
}

func (f *fakeStore) calls(key string) int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.getCalls[key]
}

func (f *fakeStore) version(content string) string {
	if f.noVersion {
		return ""
	}
	return "v:" + content
}

func (f *fakeStore) List(_ context.Context, prefix string, recursive bool) ([]folderloader.Entry, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.listPrefixes = append(f.listPrefixes, prefix)
	f.listRecurse = append(f.listRecurse, recursive)
	if f.listErr != nil {
		return nil, f.listErr
	}
	var res []folderloader.Entry
	for k, c := range f.files {
		rest, ok := strings.CutPrefix(k, prefix)
		if !ok || (!recursive && strings.Contains(strings.TrimSuffix(rest, "/"), "/")) {
			continue
		}
		v := f.version(c)
		if f.listVersion != "" {
			v = f.listVersion
		}
		res = append(res, folderloader.Entry{Key: k, Version: v})
	}
	return res, nil
}

func (f *fakeStore) Get(_ context.Context, key string) ([]byte, string, error) {
	if f.beforeGet != nil {
		f.beforeGet(key)
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.getCalls == nil {
		f.getCalls = map[string]int{}
	}
	f.getCalls[key]++
	if f.getErr != nil {
		return nil, "", f.getErr
	}
	if err := f.getErrKeys[key]; err != nil {
		return nil, "", err
	}
	c, ok := f.files[key]
	if !ok {
		return nil, "", fmt.Errorf("no such key: %s", key)
	}
	if f.getNoVersion {
		return []byte(c), "", nil
	}
	if f.getVersion != "" {
		return []byte(c), f.getVersion, nil
	}
	return []byte(c), f.version(c), nil
}

type step struct {
	mutate         func(s *fakeStore)
	wantFlags      []string // flag names of the merged flag set, nil means not checked
	wantFormat     string   // expected OutputFormat after the step, "" means not checked
	wantLog        string
	wantNoLog      string
	wantLogCount   int
	wantLogCountOf string
	wantErr        string
	wantSameAs     []string // local files whose decoding (merged) must equal the decoded result
	check          func(t *testing.T, flags map[string]dto.DTO)
	wantGets       map[string]int // cumulative number of Get per key
	wantContent    string         // exact content expected for the flag file(s), "" means not checked
}

type loadCase struct {
	name  string
	setup func(s *fakeStore)
	opts  folderloader.Options
	steps []step
}

func TestLoader_Load(t *testing.T) {
	t.Parallel()
	tests := slices.Concat([]loadCase{
		{
			name: "non recursive only loads the direct children and ignores folder markers",
			setup: func(s *fakeStore) {
				s.put("flags/a.yaml", yamlA)
				s.put("flags/b.yaml", yamlB)
				s.put("flags/sub/c.yaml", "flagC: {}\n")
				s.put("flags/", "")
				s.put("other/d.yaml", "flagD: {}\n")
			},
			opts:  folderloader.Options{Prefix: "flags/"},
			steps: []step{{wantFlags: []string{"flagA", "flagB"}, wantFormat: "yaml"}},
		},
		{
			name: "prefix without trailing slash is the folder",
			setup: func(s *fakeStore) {
				s.put("flags/a.yaml", yamlA)
				s.put("flags/sub/c.yaml", "flagC: {}\n")
				s.put("flags-old/z.yaml", "flagZ: {}\n")
			},
			opts:  folderloader.Options{Prefix: "flags"},
			steps: []step{{wantFlags: []string{"flagA"}}},
		},
		{
			name: "leading and trailing slash are the same folder",
			setup: func(s *fakeStore) {
				s.put("flags/a.yaml", yamlA)
				s.put("flags-old/z.yaml", "flagZ: {}\n")
				s.put("a.yaml", yamlB)
			},
			opts:  folderloader.Options{Prefix: "/flags/"},
			steps: []step{{wantFlags: []string{"flagA"}}},
		},
		{
			name: "empty prefix is the root, non recursive",
			setup: func(s *fakeStore) {
				s.put("a.yaml", yamlA)
				s.put("flags/b.yaml", yamlB)
			},
			opts:  folderloader.Options{},
			steps: []step{{wantFlags: []string{"flagA"}}},
		},
		{
			name: "slash prefix is the root, recursive",
			setup: func(s *fakeStore) {
				s.put("a.yaml", yamlA)
				s.put("flags/b.yaml", yamlB)
				s.put("flags/deep/c.yaml", "flagC: {}\n")
			},
			opts:  folderloader.Options{Prefix: "/", Recursive: true},
			steps: []step{{wantFlags: []string{"flagA", "flagB", "flagC"}}},
		},
		{
			name: "recursive ignores nested folder markers",
			setup: func(s *fakeStore) {
				s.put("flags/a.yaml", yamlA)
				s.put("flags/sub/b.yml", yamlB)
				s.put("flags/sub/deep/", "")
				s.put("other/d.yaml", "flagD: {}\n")
			},
			opts:  folderloader.Options{Prefix: "flags/", Recursive: true},
			steps: []step{{wantFlags: []string{"flagA", "flagB"}}},
		},
		{
			name: "pattern is matched on the file name",
			setup: func(s *fakeStore) {
				s.put("flags/a.yaml", yamlA)
				s.put("flags/b.yml", yamlB)
				s.put("flags/notes.txt", "not a flag file")
				s.put("flags/sub/c.yaml", "flagC: {}\n")
			},
			opts:  folderloader.Options{Prefix: "flags/", Pattern: "*.yaml", Recursive: true},
			steps: []step{{wantFlags: []string{"flagA", "flagC"}, wantNoLog: "notes.txt"}},
		},
		{
			name:  "json files",
			setup: func(s *fakeStore) { s.put("f/c.json", jsonC) },
			opts:  folderloader.Options{Prefix: "f/"},
			steps: []step{{wantFlags: []string{"flagC"}, wantFormat: "json"}},
		},
		{
			name:  "toml files",
			setup: func(s *fakeStore) { s.put("f/d.toml", tomlD) },
			opts:  folderloader.Options{Prefix: "f/"},
			steps: []step{{wantFlags: []string{"flagD"}, wantFormat: "toml"}},
		},
		{
			name: "yaml with non string keys behaves like a single file",
			setup: func(s *fakeStore) {
				s.put("f/k.yaml", "flagK:\n  variations:\n    true: x\n    1: y\n  defaultRule:\n    variation: x\n")
			},
			opts:  folderloader.Options{Prefix: "f/"},
			steps: []step{{wantFlags: []string{"flagK"}}},
		},
		{
			name: "large integers keep their precision in yaml",
			setup: func(s *fakeStore) {
				s.put("f/big.yaml",
					"flagBig:\n  variations:\n    big: 9007199254740993\n  defaultRule:\n    variation: big\n")
			},
			opts: folderloader.Options{Prefix: "f/"},
			steps: []step{{
				wantFlags: []string{"flagBig"},
				check: func(t *testing.T, flags map[string]dto.DTO) {
					t.Helper()
					assert.Equal(t, 9007199254740993, variationOf(flags, "flagBig", "big"))
				},
			}},
		},
		{
			name: "large integers keep their precision in toml",
			setup: func(s *fakeStore) {
				s.put("f/big.toml",
					"[flagBig.variations]\nbig = 9007199254740993\n[flagBig.defaultRule]\nvariation = \"big\"\n")
			},
			opts: folderloader.Options{Prefix: "f/"},
			steps: []step{{
				wantFlags: []string{"flagBig"},
				check: func(t *testing.T, flags map[string]dto.DTO) {
					t.Helper()
					assert.Equal(t, int64(9007199254740993), variationOf(flags, "flagBig", "big"))
				},
			}},
		},
		{
			name: "unchanged version does not redownload",
			setup: func(s *fakeStore) {
				s.put("f/a.yaml", yamlA)
				s.put("f/b.yaml", yamlB)
			},
			opts: folderloader.Options{Prefix: "f/"},
			steps: []step{
				{wantFlags: []string{"flagA", "flagB"}},
				{
					wantFlags: []string{"flagA", "flagB"},
					wantGets:  map[string]int{"f/a.yaml": 1, "f/b.yaml": 1},
				},
			},
		},
		{
			name: "changed version and new file are downloaded, others are not",
			setup: func(s *fakeStore) {
				s.put("f/a.yaml", yamlA)
				s.put("f/b.yaml", yamlB)
			},
			opts: folderloader.Options{Prefix: "f/"},
			steps: []step{
				{wantFlags: []string{"flagA", "flagB"}},
				{
					mutate: func(s *fakeStore) {
						s.put("f/a.yaml", "flagA2:\n  defaultRule:\n    variation: on\n")
						s.put("f/c.yaml", "flagC: {}\n")
					},
					wantFlags: []string{"flagA2", "flagB", "flagC"},
					wantGets:  map[string]int{"f/a.yaml": 2, "f/b.yaml": 1, "f/c.yaml": 1},
				},
			},
		},
		{
			name: "empty version never hits the cache",
			setup: func(s *fakeStore) {
				s.noVersion = true
				s.put("f/a.yaml", yamlA)
			},
			opts: folderloader.Options{Prefix: "f/"},
			steps: []step{
				{wantFlags: []string{"flagA"}},
				{
					mutate:    func(s *fakeStore) { s.put("f/a.yaml", yamlB) },
					wantFlags: []string{"flagB"},
					wantGets:  map[string]int{"f/a.yaml": 2},
				},
			},
		},
		{
			name: "listed version is used when Get has none",
			setup: func(s *fakeStore) {
				s.getNoVersion = true
				s.put("f/a.yaml", yamlA)
			},
			opts: folderloader.Options{Prefix: "f/"},
			steps: []step{
				{wantFlags: []string{"flagA"}},
				{wantFlags: []string{"flagA"}, wantGets: map[string]int{"f/a.yaml": 1}},
			},
		},
		{
			name: "quoting differences between listed and served versions are a cache hit",
			setup: func(s *fakeStore) {
				s.put("f/a.yaml", yamlA)
				s.listVersion = `"abc"`
				s.getVersion = "abc"
			},
			opts: folderloader.Options{Prefix: "f/"},
			steps: []step{
				{wantFlags: []string{"flagA"}},
				{wantFlags: []string{"flagA"}, wantGets: map[string]int{"f/a.yaml": 1}},
				{
					mutate:    func(s *fakeStore) { s.listVersion = `W/"abc"` },
					wantFlags: []string{"flagA"},
					wantGets:  map[string]int{"f/a.yaml": 1},
				},
				{
					mutate:    func(s *fakeStore) { s.listVersion = `"other"` },
					wantFlags: []string{"flagA"},
					wantGets:  map[string]int{"f/a.yaml": 2},
				},
			},
		},
		{
			// the file is replaced between the list and the download: the cache must hold the
			// version actually downloaded, so that the next list (A again) triggers a new download
			name: "file replaced between list and download is downloaded again on the next poll",
			setup: func(s *fakeStore) {
				s.put("f/a.yaml", yamlA)
				swapped := false
				s.beforeGet = func(key string) {
					if !swapped {
						swapped = true
						s.put(key, yamlB)
					}
				}
			},
			opts: folderloader.Options{Prefix: "f/"},
			steps: []step{
				// listed A, but B is downloaded
				{wantFlags: []string{"flagB"}},
				{
					// A is restored: it must not be served from the stale cache
					mutate:    func(s *fakeStore) { s.put("f/a.yaml", yamlA) },
					wantFlags: []string{"flagA"},
					wantGets:  map[string]int{"f/a.yaml": 2},
				},
			},
		},
		{
			name: "deleted file disappears and is downloaded again when it comes back",
			setup: func(s *fakeStore) {
				s.put("f/a.yaml", yamlA)
				s.put("f/b.yaml", yamlB)
			},
			opts: folderloader.Options{Prefix: "f/"},
			steps: []step{
				{wantFlags: []string{"flagA", "flagB"}},
				{
					mutate:    func(s *fakeStore) { s.del("f/b.yaml") },
					wantFlags: []string{"flagA"},
					wantGets:  map[string]int{"f/a.yaml": 1, "f/b.yaml": 1},
				},
				{
					mutate:    func(s *fakeStore) { s.put("f/b.yaml", yamlB) },
					wantFlags: []string{"flagA", "flagB"},
					wantGets:  map[string]int{"f/a.yaml": 1, "f/b.yaml": 2},
				},
			},
		},
		{
			name: "duplicate flag key",
			setup: func(s *fakeStore) {
				s.put("f/a.yaml", yamlA)
				s.put("f/z.yaml", yamlA)
			},
			opts:  folderloader.Options{Prefix: "f/"},
			steps: []step{{wantErr: `duplicate flag "flagA" found in files "f/a.yaml" and "f/z.yaml"`}},
		},
		{
			name: "mixed formats",
			setup: func(s *fakeStore) {
				s.put("f/a.yaml", yamlA)
				s.put("f/c.json", jsonC)
			},
			opts:  folderloader.Options{Prefix: "f/"},
			steps: []step{{wantErr: `"f/a.yaml" is yaml but "f/c.json" is json`}},
		},
		{
			name: "unsupported extension is skipped with a warning",
			setup: func(s *fakeStore) {
				s.put("f/a.yaml", yamlA)
				s.put("f/readme.txt", "hello")
			},
			opts:  folderloader.Options{Prefix: "f/", Location: "my-bucket"},
			steps: []step{{wantFlags: []string{"flagA"}, wantLog: `key=f/readme.txt`}},
		},
		{
			name:  "only unsupported extensions gives an empty flag set",
			setup: func(s *fakeStore) { s.put("f/readme.txt", "hello") },
			opts:  folderloader.Options{Prefix: "f/", Location: "my-bucket"},
			steps: []step{{
				wantFlags: []string{}, wantFormat: "json", wantContent: "{}",
				wantLog: "no supported file matches the prefix",
			}},
		},
		{
			name: "unsupported extension warns once while it stays skipped",
			setup: func(s *fakeStore) {
				s.put("f/a.yaml", yamlA)
				s.put("f/readme.txt", "hello")
			},
			opts: folderloader.Options{Prefix: "f/"},
			steps: []step{
				{wantFlags: []string{"flagA"}, wantLogCount: 1, wantLogCountOf: "key=f/readme.txt"},
				{wantFlags: []string{"flagA"}, wantLogCount: 1, wantLogCountOf: "key=f/readme.txt"},
				{
					mutate:       func(s *fakeStore) { s.del("f/readme.txt") },
					wantFlags:    []string{"flagA"},
					wantLogCount: 1, wantLogCountOf: "key=f/readme.txt",
				},
				{
					// the file comes back: it is reported again
					mutate:       func(s *fakeStore) { s.put("f/readme.txt", "hello") },
					wantFlags:    []string{"flagA"},
					wantLogCount: 2, wantLogCountOf: "key=f/readme.txt",
				},
			},
		},
		{
			name: "mixed formats are still an error next to an unsupported file",
			setup: func(s *fakeStore) {
				s.put("f/a.yaml", yamlA)
				s.put("f/c.json", jsonC)
				s.put("f/readme.txt", "hello")
			},
			opts:  folderloader.Options{Prefix: "f/"},
			steps: []step{{wantErr: `"f/a.yaml" is yaml but "f/c.json" is json`}},
		},
		{
			name:  "invalid content",
			setup: func(s *fakeStore) { s.put("f/a.json", "{not json") },
			opts:  folderloader.Options{Prefix: "f/"},
			steps: []step{{wantErr: `unable to parse file "f/a.json"`}},
		},
		{
			name: "empty files are empty flag sets for all formats",
			setup: func(s *fakeStore) {
				s.put("j/empty.json", "")
				s.put("y/empty.yaml", "  \n")
				s.put("t/empty.toml", "\n\n")
			},
			opts: folderloader.Options{Prefix: "j/"},
			steps: []step{
				{wantFlags: []string{}},
				{
					mutate: func(s *fakeStore) {
						s.del("j/empty.json")
						s.put("j/a.yaml", yamlA)
						s.put("j/empty.yaml", " ")
					},
					wantFlags: []string{"flagA"},
				},
				{
					mutate: func(s *fakeStore) {
						s.del("j/a.yaml")
						s.del("j/empty.yaml")
						s.put("j/empty.toml", "\n")
					},
					wantFlags: []string{},
				},
			},
		},
		{
			name:  "empty result logs a warning",
			setup: func(s *fakeStore) { s.put("other/a.yaml", yamlA) },
			opts:  folderloader.Options{Prefix: "f/"},
			steps: []step{{wantFlags: []string{}, wantLog: "no supported file matches the prefix"}},
		},
		{
			name:  "list error",
			setup: func(s *fakeStore) { s.listErr = errors.New("boom") },
			opts:  folderloader.Options{Prefix: "f/"},
			steps: []step{{wantErr: `unable to list objects with prefix "f/", boom`}},
		},
		{
			name: "download error",
			setup: func(s *fakeStore) {
				s.put("f/a.yaml", yamlA)
				s.getErr = errors.New("kaboom")
			},
			opts:  folderloader.Options{Prefix: "f/"},
			steps: []step{{wantErr: `unable to download item "f/a.yaml", kaboom`}},
		},
		{
			name: "partial download failure keeps the files already fetched",
			setup: func(s *fakeStore) {
				s.put("f/a.yaml", yamlA)
				s.put("f/b.yaml", yamlB)
				s.setGetErrKey("f/b.yaml", errors.New("kaboom"))
			},
			opts: folderloader.Options{Prefix: "f/"},
			steps: []step{
				{wantErr: `unable to download item "f/b.yaml", kaboom`},
				{
					mutate:    func(s *fakeStore) { s.setGetErrKey("f/b.yaml", nil) },
					wantFlags: []string{"flagA", "flagB"},
					wantGets:  map[string]int{"f/a.yaml": 1, "f/b.yaml": 2},
				},
			},
		},
		{
			name:  "failed load does not corrupt the cache",
			setup: func(s *fakeStore) { s.put("f/a.yaml", yamlA) },
			opts:  folderloader.Options{Prefix: "f/"},
			steps: []step{
				{wantFlags: []string{"flagA"}},
				{
					mutate:  func(s *fakeStore) { s.put("f/a.yaml", "{ bad yaml") },
					wantErr: `unable to parse file "f/a.yaml"`,
				},
				{
					// same version as the cached good one: no new download is needed
					mutate:    func(s *fakeStore) { s.put("f/a.yaml", yamlA) },
					wantFlags: []string{"flagA"},
					wantGets:  map[string]int{"f/a.yaml": 2},
				},
			},
		},
	}, roundTripCases(t))
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			runLoadCase(t, tt)
		})
	}
}

// roundTripCases checks that the merged output, decoded like the flag manager does, is the same
// as the decoding of the original file.
func roundTripCases(t *testing.T) []loadCase {
	t.Helper()
	files := []string{
		"flag-config.yaml", "flag-config.json", "flag-config.toml",
		"flag-config-multiline-query.json", "flag-config-scheduled-v0.yaml",
		"flag-config-3rd-file.yaml", "flag-config-custom-bucketingkey.yaml",
	}
	cases := make([]loadCase, 0, len(files))
	for _, f := range files {
		local := "../../testdata/" + f
		content, err := os.ReadFile(local)
		require.NoError(t, err)
		cases = append(cases, loadCase{
			name:  "round trip " + f,
			setup: func(s *fakeStore) { s.put("f/"+f, string(content)) },
			opts:  folderloader.Options{Prefix: "f/"},
			steps: []step{{wantSameAs: []string{local}}},
		})
	}
	return cases
}

func runLoadCase(t *testing.T, c loadCase) {
	t.Helper()
	s := &fakeStore{}
	c.setup(s)
	logs := &bytes.Buffer{}
	c.opts.Logger = &fflog.FFLogger{LeveledLogger: slog.New(slog.NewTextHandler(logs, nil))}
	l := folderloader.New(s, c.opts)

	for i, st := range c.steps {
		if st.mutate != nil {
			st.mutate(s)
		}
		got, err := l.Load(t.Context())
		if st.wantErr != "" {
			require.ErrorContains(t, err, st.wantErr, "step %d", i)
		} else {
			require.NoError(t, err, "step %d", i)
			checkOutput(t, l.OutputFormat(), st, got, i)
		}
		if st.wantFormat != "" {
			assert.Equal(t, st.wantFormat, l.OutputFormat(), "step %d", i)
		}
		for k, n := range st.wantGets {
			assert.Equal(t, n, s.calls(k), "Get calls of %s at step %d", k, i)
		}
		checkLogs(t, logs, st, i)
	}
}

func checkLogs(t *testing.T, logs *bytes.Buffer, st step, i int) {
	t.Helper()
	if st.wantLog != "" {
		assert.Contains(t, logs.String(), st.wantLog, "step %d", i)
	}
	if st.wantNoLog != "" {
		assert.NotContains(t, logs.String(), st.wantNoLog, "step %d", i)
	}
	if st.wantLogCountOf != "" {
		assert.Equal(t, st.wantLogCount, strings.Count(logs.String(), st.wantLogCountOf), "step %d", i)
	}
}

func checkOutput(t *testing.T, format string, st step, got []byte, i int) {
	t.Helper()
	if st.wantContent != "" {
		assert.Equal(t, st.wantContent, string(got), "step %d", i)
	}
	flags, err := cache.ConvertToFlagStruct(got, format)
	require.NoError(t, err, "step %d", i)
	if st.wantFlags != nil {
		keys := make([]string, 0, len(flags))
		for k := range flags {
			keys = append(keys, k)
		}
		assert.ElementsMatch(t, st.wantFlags, keys, "step %d", i)
	}
	if len(st.wantSameAs) > 0 {
		want := map[string]dto.DTO{}
		for _, f := range st.wantSameAs {
			content, rErr := os.ReadFile(f)
			require.NoError(t, rErr)
			parsed, pErr := cache.ConvertToFlagStruct(content, strings.TrimPrefix(filepath.Ext(f), "."))
			require.NoError(t, pErr)
			maps.Copy(want, parsed)
		}
		assert.Equal(t, want, flags, "step %d", i)
	}
	if st.check != nil {
		st.check(t, flags)
	}
}

func variationOf(flags map[string]dto.DTO, flagName, variation string) any {
	f, ok := flags[flagName]
	if !ok || f.Variations == nil {
		return nil
	}
	v := (*f.Variations)[variation]
	if v == nil {
		return nil
	}
	return *v
}

func TestLoader_Load_Provider(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		opts        folderloader.Options
		wantList    string
		wantGet     string
		wantLogs    []string
		wantNotLogs []string
	}{
		{
			name:        "neutral wording",
			opts:        folderloader.Options{Prefix: "f/", Location: "loc"},
			wantList:    `unable to list objects with prefix "f/", boom`,
			wantGet:     `unable to download item "f/a.yaml", kaboom`,
			wantLogs:    []string{"folder loader: no supported file", "location=loc"},
			wantNotLogs: []string{"bucket="},
		},
		{
			name: "provider wording",
			opts: folderloader.Options{
				Prefix: "f/", Location: "loc", Provider: "S3", LocationKey: "bucket",
			},
			wantList: `unable to list objects from S3 with prefix "f/", boom`,
			wantGet:  `unable to download item from S3 "f/a.yaml", kaboom`,
			wantLogs: []string{"S3 retriever: no supported file", "bucket=loc"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			logs := &bytes.Buffer{}
			tt.opts.Logger = &fflog.FFLogger{LeveledLogger: slog.New(slog.NewTextHandler(logs, nil))}

			ls := &fakeStore{listErr: errors.New("boom")}
			_, err := folderloader.New(ls, tt.opts).Load(t.Context())
			require.ErrorContains(t, err, tt.wantList)

			gs := &fakeStore{getErr: errors.New("kaboom")}
			gs.put("f/a.yaml", yamlA)
			_, err = folderloader.New(gs, tt.opts).Load(t.Context())
			require.ErrorContains(t, err, tt.wantGet)

			_, err = folderloader.New(&fakeStore{}, tt.opts).Load(t.Context())
			require.NoError(t, err)
			for _, w := range tt.wantLogs {
				assert.Contains(t, logs.String(), w)
			}
			for _, w := range tt.wantNotLogs {
				assert.NotContains(t, logs.String(), w)
			}
		})
	}
}

func TestLoader_Load_StorePrefix(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name          string
		prefix        string
		recursive     bool
		wantPrefix    string
		wantRecursive bool
	}{
		{name: "empty", prefix: "", wantPrefix: ""},
		{name: "slash", prefix: "/", recursive: true, wantPrefix: "", wantRecursive: true},
		{name: "no trailing slash", prefix: "flags", wantPrefix: "flags/"},
		{name: "leading slashes", prefix: "//flags/sub/", wantPrefix: "flags/sub/"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			s := &fakeStore{}
			l := folderloader.New(s, folderloader.Options{Prefix: tt.prefix, Recursive: tt.recursive})
			_, err := l.Load(t.Context())
			require.NoError(t, err)
			assert.Equal(t, []string{tt.wantPrefix}, s.listPrefixes)
			assert.Equal(t, []bool{tt.wantRecursive}, s.listRecurse)
		})
	}
}

func TestLoader_OutputFormat(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		pattern string
		files   map[string]string
		want    string
		// afterLoad is the expected format once Load was called.
		afterLoad string
	}{
		{name: "no pattern is unknown before the first load", afterLoad: "json"},
		{name: "yaml pattern", pattern: "*.yml", want: "yaml", afterLoad: "json"},
		{name: "toml pattern", pattern: "*.toml", want: "toml", afterLoad: "json"},
		{name: "json pattern", pattern: "*.json", want: "json", afterLoad: "json"},
		{name: "wildcard pattern", pattern: "*", afterLoad: "json"},
		{
			name: "files win over the pattern", pattern: "*", files: map[string]string{"f/a.yaml": yamlA},
			afterLoad: "yaml",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			s := &fakeStore{files: tt.files}
			l := folderloader.New(s, folderloader.Options{Prefix: "f/", Pattern: tt.pattern})
			assert.Equal(t, tt.want, l.OutputFormat())
			_, err := l.Load(t.Context())
			require.NoError(t, err)
			assert.Equal(t, tt.afterLoad, l.OutputFormat())
		})
	}
}

func TestLoader_Reset(t *testing.T) {
	t.Parallel()
	s := &fakeStore{}
	s.put("f/a.yaml", yamlA)
	s.put("f/readme.txt", "hello")
	logs := &bytes.Buffer{}
	logger := &fflog.FFLogger{LeveledLogger: slog.New(slog.NewTextHandler(logs, nil))}
	l := folderloader.New(s, folderloader.Options{Prefix: "f/", Logger: logger})

	for range 2 {
		_, err := l.Load(t.Context())
		require.NoError(t, err)
	}
	assert.Equal(t, 1, s.calls("f/a.yaml"))
	assert.Equal(t, 1, strings.Count(logs.String(), "key=f/readme.txt"))

	// Reset drops the cache and the skipped files: all is downloaded and reported again.
	l.Reset()
	_, err := l.Load(t.Context())
	require.NoError(t, err)
	assert.Equal(t, 2, s.calls("f/a.yaml"))
	assert.Equal(t, 2, strings.Count(logs.String(), "key=f/readme.txt"))
}

func TestReuse(t *testing.T) {
	t.Parallel()
	logger := &fflog.FFLogger{}
	base := folderloader.Options{Prefix: "f/", Pattern: "a.*", Recursive: true, Logger: logger, Location: "b"}
	tests := []struct {
		name      string
		nilLoader bool
		opts      func(o folderloader.Options) folderloader.Options
		wantSame  bool
	}{
		{name: "nil loader", nilLoader: true},
		{name: "same options", opts: func(o folderloader.Options) folderloader.Options { return o }, wantSame: true},
		{name: "prefix changed", opts: func(o folderloader.Options) folderloader.Options { o.Prefix = "g/"; return o }},
		{name: "pattern changed", opts: func(o folderloader.Options) folderloader.Options { o.Pattern = "*.yaml"; return o }},
		{name: "recursive changed", opts: func(o folderloader.Options) folderloader.Options { o.Recursive = false; return o }},
		{
			name: "logger changed",
			opts: func(o folderloader.Options) folderloader.Options { o.Logger = &fflog.FFLogger{}; return o },
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			s := &fakeStore{}
			s.put("f/a.yaml", yamlA)
			s.put("g/a.yaml", yamlA)
			old := folderloader.New(s, base)
			_, err := old.Load(t.Context())
			require.NoError(t, err)

			var prev *folderloader.Loader
			if !tt.nilLoader {
				prev = old
			}
			opts := base
			if tt.opts != nil {
				opts = tt.opts(base)
			}
			got := folderloader.Reuse(prev, s, opts)
			require.NotNil(t, got)
			if tt.wantSame {
				assert.Same(t, old, got)
				assert.Equal(t, "yaml", got.OutputFormat())
				return
			}
			assert.NotSame(t, old, got)
			// a new loader starts with an empty cache: the file is downloaded again
			_, err = got.Load(t.Context())
			require.NoError(t, err)
			assert.Equal(t, 2, s.calls("f/a.yaml")+s.calls("g/a.yaml"))
		})
	}
}
