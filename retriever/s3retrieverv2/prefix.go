package s3retrieverv2

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"path"
	"sort"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/thomaspoignant/go-feature-flag/internal/cache"
	"github.com/thomaspoignant/go-feature-flag/modules/core/dto"
	"github.com/thomaspoignant/go-feature-flag/retriever"
	"gopkg.in/yaml.v3"
)

// cachedObject is the last parsed version of an object.
type cachedObject struct {
	etag  string
	flags map[string]dto.DTO
}

// normalizedPrefix returns the prefix as sent to S3: S3 keys never start with "/", so leading
// slashes are trimmed, and a trailing "/" is added if missing so that "flags" is a folder like
// "flags/". The prefix "/" (or "") is the root of the bucket and gives an empty prefix.
func (s *Retriever) normalizedPrefix() string {
	p := strings.TrimLeft(s.Prefix, "/")
	if p != "" && !strings.HasSuffix(p, "/") {
		p += "/"
	}
	return p
}

// loadFolder lists the matching objects, downloads and parses the ones that changed since the
// last call and merges all of them into a single flag set, encoded in the format of the files
// (see OutputFormat) so that the flag manager decodes it exactly like a single file.
// The caller must hold s.mu.
func (s *Retriever) loadFolder(ctx context.Context) ([]byte, error) {
	if s.lister == nil {
		s.status = retriever.RetrieverError
		return nil, errors.New("lister is not initialized")
	}

	objects, err := s.list(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to list objects from S3 with prefix %q, %v", s.Prefix, err)
	}
	keys := sortedKeys(objects)
	if err := checkFormats(keys); err != nil {
		return nil, err
	}
	if len(keys) == 0 {
		s.logger.Warn("S3 retriever: no supported file matches the prefix, the flag set is empty",
			"bucket", s.Bucket, "prefix", s.normalizedPrefix(), "pattern", s.Pattern)
	}
	if err := s.refreshCache(ctx, keys, objects); err != nil {
		return nil, err
	}

	merged, err := s.mergeFlags(keys)
	if err != nil {
		return nil, err
	}
	format := "json" // an empty flag set is "{}", valid in JSON
	if len(keys) > 0 {
		format = formatOf(keys[0])
	}
	s.format.Store(format)
	return marshalFlags(format, merged)
}

func sortedKeys(objects map[string]string) []string {
	keys := make([]string, 0, len(objects))
	for k := range objects {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// refreshCache downloads and parses the objects whose ETag changed. The cache only keeps the
// listed keys (deleted files are dropped) and is committed even if a download fails, so the files
// already fetched are not downloaded again.
func (s *Retriever) refreshCache(ctx context.Context, keys []string, objects map[string]string) error {
	working := make(map[string]cachedObject, len(keys))
	for _, key := range keys {
		if c, ok := s.cache[key]; ok {
			working[key] = c
		}
	}
	defer func() { s.cache = working }()

	for _, key := range keys {
		// an empty ETag can't tell if the content changed, so it never hits the cache
		if c, ok := working[key]; ok && c.etag != "" && c.etag == objects[key] {
			continue
		}
		content, err := s.download(ctx, key)
		if err != nil {
			return fmt.Errorf("unable to download item from S3 %q, %v", key, err)
		}
		flags, err := parseFile(key, content)
		if err != nil {
			return err
		}
		working[key] = cachedObject{etag: objects[key], flags: flags}
	}
	return nil
}

// mergeFlags merges the flags of all the files, a flag defined twice is an error.
func (s *Retriever) mergeFlags(keys []string) (map[string]dto.DTO, error) {
	merged := make(map[string]dto.DTO)
	origin := make(map[string]string)
	for _, key := range keys {
		for flagKey, flag := range s.cache[key].flags {
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

// list returns the object keys to load with their ETag. The keys matching the pattern but with an
// unsupported extension are skipped with a warning (see warnSkipped).
func (s *Retriever) list(ctx context.Context) (map[string]string, error) {
	input := &s3.ListObjectsV2Input{Bucket: aws.String(s.Bucket)}
	if p := s.normalizedPrefix(); p != "" {
		input.Prefix = aws.String(p)
	}
	if !s.Recursive {
		input.Delimiter = aws.String("/")
	}
	res := make(map[string]string)
	skipped := make(map[string]struct{})
	paginator := s3.NewListObjectsV2Paginator(s.lister, input)
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		s.collect(page.Contents, res, skipped)
	}
	s.warnSkipped(skipped)
	return res, nil
}

// collect sorts the objects of a page between the ones to load (res) and the ones skipped because
// of their extension (skipped).
func (s *Retriever) collect(objects []s3types.Object, res map[string]string, skipped map[string]struct{}) {
	for _, obj := range objects {
		key := aws.ToString(obj.Key)
		switch {
		case !s.isEligible(key):
		case formatOf(key) == "":
			skipped[key] = struct{}{}
		default:
			res[key] = aws.ToString(obj.ETag)
		}
	}
}

// warnSkipped logs the keys newly skipped, and remembers the skipped keys of this poll. A key is
// logged once while it stays skipped, and again if it disappears and comes back.
// The caller must hold s.mu.
func (s *Retriever) warnSkipped(skipped map[string]struct{}) {
	newKeys := make([]string, 0, len(skipped))
	for key := range skipped {
		if _, known := s.skipped[key]; !known {
			newKeys = append(newKeys, key)
		}
	}
	sort.Strings(newKeys)
	for _, key := range newKeys {
		s.logger.Warn("S3 retriever: file skipped because of its unsupported extension "+
			"(supported: .json, .yaml, .yml, .toml)", "bucket", s.Bucket, "key", key)
	}
	s.skipped = skipped
}

// isEligible tells if an object key must be loaded: not a folder marker and matching the pattern.
func (s *Retriever) isEligible(key string) bool {
	if key == "" || strings.HasSuffix(key, "/") {
		return false
	}
	if s.Pattern == "" {
		return true
	}
	ok, _ := path.Match(s.Pattern, path.Base(key))
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
// are identical to the item mode. An empty file is an empty flag set.
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
