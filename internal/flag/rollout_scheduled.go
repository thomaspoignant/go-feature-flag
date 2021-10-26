package flag

import (
	"time"
)

type ScheduledRollout struct {
	// Steps is the list of updates to do in a specific date.
	Steps []ScheduledStep `json:"steps,omitempty" yaml:"steps,omitempty" toml:"steps,omitempty"`
}

type ScheduledStep struct {
	FlagData `yaml:",inline"`
	Date     *time.Time `json:"date,omitempty" yaml:"date,omitempty" toml:"date,omitempty"`
}
