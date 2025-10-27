package flag

import (
	"time"
)

// ScheduledStep is one change of the flag.
type ScheduledStep struct {
	InternalFlag `yaml:",inline"`
	Date         *time.Time `yaml:"date,omitempty" json:"date,omitempty" toml:"date,omitempty"`
}
