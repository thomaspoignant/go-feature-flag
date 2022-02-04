package flag

import (
	"fmt"
	"github.com/thomaspoignant/go-feature-flag/internal/fflog"
	"strings"
	"time"
)

type Experimentation struct {
	// Start is the starting time of the experimentation
	Start *time.Time `json:"start,omitempty" yaml:"start,omitempty" toml:"start,omitempty"`

	// End is the ending time of the experimentation
	End *time.Time `json:"end,omitempty" yaml:"end,omitempty" toml:"end,omitempty"`
}

// String is displaying the experimentation rollout.
func (e Experimentation) String() string {
	buf := make([]string, 0)
	lo, _ := time.LoadLocation("UTC")

	if e.Start != nil {
		buf = append(buf, fmt.Sprintf("start:[%v]", e.Start.In(lo).Format(fflog.LogDateFormat)))
	}
	if e.End != nil {
		buf = append(buf, fmt.Sprintf("end:[%v]", e.End.In(lo).Format(fflog.LogDateFormat)))
	}
	return strings.Join(buf, " ")
}
