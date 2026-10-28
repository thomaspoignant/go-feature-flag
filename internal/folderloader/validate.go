package folderloader

import (
	"fmt"
	"path"
	"strings"
)

// Validate checks the settings of a retriever supporting the folder mode. single is the path of a
// single file, an empty single means folder mode, the only mode where prefix, pattern and
// recursive can be used. itemName is how the provider calls a single file ("item", "object"...),
// it is only used in the error messages.
func Validate(itemName, single, prefix, pattern string, recursive bool) error {
	switch {
	case single != "" && prefix != "":
		return fmt.Errorf("%s and prefix are mutually exclusive", itemName)
	case single != "" && (pattern != "" || recursive):
		return fmt.Errorf("pattern and recursive can only be used without %s", itemName)
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
