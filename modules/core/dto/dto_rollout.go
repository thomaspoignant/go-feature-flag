package dto

import (
	"time"
)

type ExperimentationDto struct {
	// Start is the starting time of the experimentation
	Start *time.Time `json:"start,omitempty" yaml:"start,omitempty" toml:"start,omitempty" jsonschema:"required,title=start,description=Time of the start of the experimentation."` // nolint: lll

	// End is the ending time of the experimentation
	End *time.Time `json:"end,omitempty" yaml:"end,omitempty" toml:"end,omitempty" jsonschema:"required,title=start,description=Time of the end of the experimentation."` // nolint: lll
}
