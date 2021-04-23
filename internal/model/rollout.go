package model

import (
	"fmt"
	"strings"
	"time"
)

type Rollout struct {
	// Experimentation is your object to configure an experimentation, it will allow you to configure a start date and
	// an end date for your flag.
	// When the experimentation is not running, the flag will serve the default value.
	Experimentation *Experimentation `json:"experimentation,omitempty" yaml:"experimentation,omitempty" toml:"experimentation,omitempty" slack_short:"false"` // nolint: lll
}

func (e Rollout) String() string {
	if e.Experimentation != nil {
		return "experimentation: " + e.Experimentation.String()
	}
	return ""
}

type Experimentation struct {
	// Start is the starting time of the experimentation
	Start *time.Time `json:"start,omitempty" yaml:"start,omitempty" toml:"start,omitempty"`

	// End is the ending time of the experimentation
	End *time.Time `json:"end,omitempty" yaml:"end,omitempty" toml:"end,omitempty"`
}

func (e Experimentation) String() string {
	buf := make([]string, 0)
	lo, _ := time.LoadLocation("UTC")

	if e.Start != nil {
		buf = append(buf, fmt.Sprintf("start:[%v]", e.Start.In(lo).Format(time.RFC3339)))
	}
	if e.End != nil {
		buf = append(buf, fmt.Sprintf("end:[%v]", e.End.In(lo).Format(time.RFC3339)))
	}
	return strings.Join(buf, " ")
}
