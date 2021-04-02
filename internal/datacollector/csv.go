package datacollector

import (
	"fmt"
	"strings"

	"github.com/thomaspoignant/go-feature-flag/internal/model"
)

type CsvCollector struct{}

func (c *CsvCollector) Collect(userKey string, flag string, value interface{}, cohort model.VariationType) {
	separator := ";"
	var b strings.Builder
	b.WriteString(userKey)
	b.WriteString(separator)
	b.WriteString(flag)
	b.WriteString(separator)
	b.WriteString(fmt.Sprintf("%v", value))
	b.WriteString(separator)
	b.WriteString(string(cohort))
	fmt.Println(b.String())
}
