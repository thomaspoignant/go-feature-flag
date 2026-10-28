// Package s3folder holds the validation shared by the places configuring the folder mode of the
// S3 retriever (the retriever itself and the relay proxy / CLI configuration).
// It has no dependency so it can be imported from anywhere.
package s3folder

import (
	"errors"
	"fmt"
	"path"
	"strings"
)

// Validate checks the settings of the S3 retriever. Item is the path of a single file, an empty
// Item means folder mode, the only mode where prefix, pattern and recursive can be used.
func Validate(item, prefix, pattern string, recursive bool) error {
	switch {
	case item != "" && prefix != "":
		return errors.New("item and prefix are mutually exclusive")
	case item != "" && (pattern != "" || recursive):
		return errors.New("pattern and recursive can only be used without item")
	}
	return validatePattern(pattern)
}

// validatePattern checks that the pattern is a valid glob and only targets the file name.
func validatePattern(pattern string) error {
	if pattern == "" {
		return nil
	}
	if _, err := path.Match(pattern, ""); err != nil {
		return fmt.Errorf("invalid pattern %q: %v", pattern, err)
	}
	if strings.Contains(pattern, "/") {
		return fmt.Errorf(
			"invalid pattern %q: it is matched against the file name only and can't contain \"/\"",
			pattern)
	}
	return nil
}
