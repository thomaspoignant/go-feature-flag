package flag

import (
	"fmt"
	"github.com/thomaspoignant/go-feature-flag/internal/fflog"
	"strings"
	"time"
)

// ScheduledRollout is a way to modify your flag along the time
type ScheduledRollout struct {
	// Steps is the list of updates to do in a specific date.
	Steps []ScheduledStep `json:"steps,omitempty" yaml:"steps,omitempty" toml:"steps,omitempty"`
}

func (s ScheduledRollout) String() string {
	steps := make([]string, len(s.Steps))
	for i, step := range s.Steps {
		steps[i] = step.String()
	}
	return strings.Join(steps, ",")
}

// ScheduledStep is one change of the flag.
type ScheduledStep struct {
	FlagData `yaml:",inline"`
	Date     *time.Time `json:"date,omitempty" yaml:"date,omitempty" toml:"date,omitempty"`
}

func (s ScheduledStep) String() string {
	if s.Date != nil {
		return fmt.Sprintf("[%s: %s]", s.Date.Format(fflog.LogDateFormat), s.FlagData.String())
	}
	return ""
}
