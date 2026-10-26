// Package folderloader holds the provider independent logic of the "folder mode" of the
// retrievers (S3, GCS, Azure Blob...): list the files under a prefix, download and parse the ones
// that changed, and merge all of them into a single flag set.
// The providers only have to implement the Store interface.
package folderloader

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"path"
	"sort"
	"strings"
	"sync/atomic"

	"github.com/BurntSushi/toml"
	"github.com/thomaspoignant/go-feature-flag/internal/cache"
	"github.com/thomaspoignant/go-feature-flag/modules/core/dto"
	"github.com/thomaspoignant/go-feature-flag/utils/fflog"
	"gopkg.in/yaml.v3"
)

// Entry is an object found by Store.List.
type Entry struct {
	// Key is the full key (path) of the object.
	Key string
	// Version tells if the content changed (ETag, generation...). It is opaque for the loader,
	// an empty Version never matches the cache.
	Version string
}

// Store is the access to the objects of a provider.
type Store interface {
	// List returns the objects under prefix, which is already normalized (no leading "/", a
	// trailing "/" unless empty). If recursive is false only the direct children of prefix must be
	// returned, so the store does the delimiter work. It may return folder markers (keys ending
	// with "/"), the loader ignores them.
	List(ctx context.Context, prefix string, recursive bool) ([]Entry, error)
	// Get downloads an object, and returns the version of what it actually served. It may be ""
	// if the provider has none, the Version found by List is used then.
	// The version must identify the same revision as the Version of List, so that they can be
	// compared. Differences of quoting are tolerated: a weak prefix (W/) and surrounding double
	// quotes are ignored.
	Get(ctx context.Context, key string) (content []byte, version string, err error)
}

// Options are the settings of a Loader.
type Options struct {
	// Prefix is the folder to load. Leading "/" are trimmed and a trailing "/" is added if missing.
	// "" and "/" are the root.
	Prefix string
	// Pattern is an optional glob (see path.Match) matched against the file name only.
	Pattern string
	// Recursive also loads the "sub folders" of Prefix.
	Recursive bool
	// Logger is used for the warnings, it can be nil.
	Logger *fflog.FFLogger
	// Location is the name of the bucket / container, used in the logs.
	Location string
	// LocationKey is the log field holding Location (ex: "bucket"), "location" if empty.
	LocationKey string
	// Provider is the name of the provider used in the logs and the errors (ex: "S3"). If empty,
	// the wording is provider-neutral.
	Provider string
}

// cachedFile is the last parsed version of a file.
type cachedFile struct {
	version string
	flags   map[string]dto.DTO
}

// Loader loads and merges the flag files of a folder.
// A Loader is not safe for concurrent use: the caller must serialize the calls.
type Loader struct {
	store  Store
	opts   Options
	prefix string
	// format is the format of the last Load, read by OutputFormat.
	format  atomic.Value
	cache   map[string]cachedFile
	skipped map[string]struct{}
}

// New creates a Loader reading the files from store.
func New(store Store, opts Options) *Loader {
	return &Loader{store: store, opts: opts, prefix: normalizePrefix(opts.Prefix)}
}

// normalizePrefix returns the prefix as sent to the store: keys never start with "/", so leading
// slashes are trimmed, and a trailing "/" is added if missing so that "flags" is a folder like
// "flags/". The prefix "/" (or "") is the root and gives an empty prefix.
func normalizePrefix(prefix string) string {
	p := strings.TrimLeft(prefix, "/")
	if p != "" && !strings.HasSuffix(p, "/") {
		p += "/"
	}
	return p
}

// from returns " from <Provider>" for the error messages, or "" if there is no provider.
func (l *Loader) from() string {
	if l.opts.Provider == "" {
		return ""
	}
	return " from " + l.opts.Provider
}

// logPrefix is the beginning of the log messages.
func (l *Loader) logPrefix() string {
	if l.opts.Provider == "" {
		return "folder loader"
	}
	return l.opts.Provider + " retriever"
}

// locationKey is the name of the log field holding the location.
func (l *Loader) locationKey() string {
	if l.opts.LocationKey == "" {
		return "location"
	}
	return l.opts.LocationKey
}

// sameVersion compares two versions ignoring the weak prefix and the surrounding double quotes,
// which providers or SDKs do not always keep (ex: the ETag of a listing vs a download).
func sameVersion(a, b string) bool {
	return normalizeVersion(a) == normalizeVersion(b)
}

func normalizeVersion(v string) string {
	v = strings.TrimPrefix(v, "W/")
	return strings.Trim(v, `"`)
}

// Reset drops the cache and the skipped files.
func (l *Loader) Reset() {
	l.cache = nil
	l.skipped = nil
}

// OutputFormat is the format of the bytes returned by Load: the format of the files found by the
// last Load ("json" for an empty flag set). Before the first Load it is guessed from the pattern.
func (l *Loader) OutputFormat() string {
	if f, ok := l.format.Load().(string); ok {
		return f
	}
	return formatOf(l.opts.Pattern)
}

// Load lists the matching files, downloads and parses the ones that changed since the last call
// and merges all of them into a single flag set, encoded in the format of the files
// (see OutputFormat) so that the flag manager decodes it exactly like a single file.
func (l *Loader) Load(ctx context.Context) ([]byte, error) {
	files, err := l.list(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to list objects%s with prefix %q, %v", l.from(), l.opts.Prefix, err)
	}
	keys := sortedKeys(files)
	if err := checkFormats(keys); err != nil {
		return nil, err
	}
	if len(keys) == 0 {
		l.opts.Logger.Warn(l.logPrefix()+": no supported file matches the prefix, the flag set is empty",
			l.locationKey(), l.opts.Location, "prefix", l.prefix, "pattern", l.opts.Pattern)
	}
	if err := l.refreshCache(ctx, keys, files); err != nil {
		return nil, err
	}

	merged, err := l.mergeFlags(keys)
	if err != nil {
		return nil, err
	}
	format := "json" // an empty flag set is "{}", valid in JSON
	if len(keys) > 0 {
		format = formatOf(keys[0])
	}
	l.format.Store(format)
	return marshalFlags(format, merged)
}

func sortedKeys(files map[string]string) []string {
	keys := make([]string, 0, len(files))
	for k := range files {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// refreshCache downloads and parses the files whose version changed. The cache only keeps the
// listed keys (deleted files are dropped) and is committed even if a download fails, so the files
// already fetched are not downloaded again.
// The version cached is the one of the downloaded content, not the listed one: if the file was
// replaced between the list and the download, the next poll sees a different version and
// downloads it again.
func (l *Loader) refreshCache(ctx context.Context, keys []string, files map[string]string) error {
	working := make(map[string]cachedFile, len(keys))
	for _, key := range keys {
		if c, ok := l.cache[key]; ok {
			working[key] = c
		}
	}
	defer func() { l.cache = working }()

	for _, key := range keys {
		// an empty version can't tell if the content changed, so it never hits the cache
		if c, ok := working[key]; ok && c.version != "" && sameVersion(c.version, files[key]) {
			continue
		}
		content, version, err := l.store.Get(ctx, key)
		if err != nil {
			return fmt.Errorf("unable to download item%s %q, %v", l.from(), key, err)
		}
		flags, err := parseFile(key, content)
		if err != nil {
			return err
		}
		if version == "" {
			version = files[key]
		}
		working[key] = cachedFile{version: version, flags: flags}
	}
	return nil
}

// mergeFlags merges the flags of all the files, a flag defined twice is an error.
func (l *Loader) mergeFlags(keys []string) (map[string]dto.DTO, error) {
	merged := make(map[string]dto.DTO)
	origin := make(map[string]string)
	for _, key := range keys {
		for flagKey, flag := range l.cache[key].flags {
			if prev, dup := origin[flagKey]; dup {
				return nil, fmt.Errorf(
					"duplicate flag %q found in files %q and %q", flagKey, prev, key)
			}
			origin[flagKey] = key
			merged[flagKey] = flag
		}
	}
	return merged, nil
}

// marshalFlags encodes the flags in the given format. The encoding is the same family as the one
// used to decode the files, which keeps the value types (ex: big integers) intact.
func marshalFlags(format string, flags map[string]dto.DTO) ([]byte, error) {
	var out []byte
	var err error
	switch format {
	case "yaml":
		out, err = yaml.Marshal(flags)
	case "toml":
		buf := &bytes.Buffer{}
		err = toml.NewEncoder(buf).Encode(flags)
		out = buf.Bytes()
	default:
		out, err = json.Marshal(flags)
	}
	if err != nil {
		return nil, fmt.Errorf("unable to marshal the merged flags: %v", err)
	}
	return out, nil
}

// list returns the keys to load with their version. The keys matching the pattern but with an
// unsupported extension are skipped with a warning (see warnSkipped).
func (l *Loader) list(ctx context.Context) (map[string]string, error) {
	entries, err := l.store.List(ctx, l.prefix, l.opts.Recursive)
	if err != nil {
		return nil, err
	}
	res := make(map[string]string)
	skipped := make(map[string]struct{})
	for _, e := range entries {
		switch {
		case !l.isEligible(e.Key):
		case formatOf(e.Key) == "":
			skipped[e.Key] = struct{}{}
		default:
			res[e.Key] = e.Version
		}
	}
	l.warnSkipped(skipped)
	return res, nil
}

// warnSkipped logs the keys newly skipped, and remembers the skipped keys of this poll. A key is
// logged once while it stays skipped, and again if it disappears and comes back.
func (l *Loader) warnSkipped(skipped map[string]struct{}) {
	newKeys := make([]string, 0, len(skipped))
	for key := range skipped {
		if _, known := l.skipped[key]; !known {
			newKeys = append(newKeys, key)
		}
	}
	sort.Strings(newKeys)
	for _, key := range newKeys {
		l.opts.Logger.Warn(l.logPrefix()+": file skipped because of its unsupported extension "+
			"(supported: .json, .yaml, .yml, .toml)", l.locationKey(), l.opts.Location, "key", key)
	}
	l.skipped = skipped
}

// isEligible tells if a key must be loaded: not a folder marker and matching the pattern.
func (l *Loader) isEligible(key string) bool {
	if key == "" || strings.HasSuffix(key, "/") {
		return false
	}
	if l.opts.Pattern == "" {
		return true
	}
	ok, _ := path.Match(l.opts.Pattern, path.Base(key))
	return ok
}

func formatOf(key string) string {
	switch strings.ToLower(path.Ext(key)) {
	case ".json":
		return "json"
	case ".yaml", ".yml":
		return "yaml"
	case ".toml":
		return "toml"
	}
	return ""
}

// checkFormats ensures that all the files are of the same format. The keys must have a supported
// extension, the others are skipped by list.
func checkFormats(keys []string) error {
	var firstKey, firstFormat string
	for _, key := range keys {
		f := formatOf(key)
		if firstFormat == "" {
			firstKey, firstFormat = key, f
			continue
		}
		if f != firstFormat {
			return fmt.Errorf(
				"all the files under the prefix must have the same format: %q is %s but %q is %s",
				firstKey, firstFormat, key, f)
		}
	}
	return nil
}

// parseFile parses a file the same way the flag manager does for a single file, so the semantics
// are identical to the single file mode. An empty file is an empty flag set.
func parseFile(key string, content []byte) (map[string]dto.DTO, error) {
	if len(bytes.TrimSpace(content)) == 0 {
		return map[string]dto.DTO{}, nil
	}
	flags, err := cache.ConvertToFlagStruct(content, formatOf(key))
	if err != nil {
		return nil, fmt.Errorf("unable to parse file %q: %v", key, err)
	}
	return flags, nil
}
