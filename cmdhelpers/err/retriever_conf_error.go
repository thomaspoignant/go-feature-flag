package err

import (
	"fmt"
	"regexp"
	"strings"
)

var camelToKebabRegex = regexp.MustCompile("([a-z0-9])([A-Z])")

type RetrieverConfError struct {
	property string
	kind     string
}

func NewRetrieverConfError(property string, kind string) *RetrieverConfError {
	return &RetrieverConfError{
		property: property,
		kind:     kind,
	}
}

func (e *RetrieverConfError) Error() string {
	return fmt.Sprintf("invalid retriever: no \"%s\" property found for kind \"%s\"", e.property, e.kind)
}

func (e *RetrieverConfError) CliErrorMessage() string {
	kebab := camelToKebabRegex.ReplaceAllString(e.property, "${1}-${2}")
	return fmt.Sprintf("invalid retriever: no \"%s\" property found for kind \"%s\"", strings.ToLower(kebab), e.kind)
}
