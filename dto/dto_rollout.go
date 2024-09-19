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

type Rollout struct {
	// CommonRollout is the struct containing the configuration for rollout that applies for
	// all types of flag in your input file.
	CommonRollout `json:",inline" yaml:",inline" toml:",inline"`

	// V0Rollout contains the configuration available only for the flags version v0.X.X
	V0Rollout `json:",inline" yaml:",inline" toml:",inline"` // nolint: govet
}

type CommonRollout struct {
	// Experimentation is your struct to configure an experimentation, it will allow you to configure a start date and
	// an end date for your flag.
	// When the experimentation is not running, the flag will serve the default value.
	Experimentation *ExperimentationDto `json:"experimentation,omitempty" yaml:"experimentation,omitempty" toml:"experimentation,omitempty"` // nolint: lll

	// Progressive (only for v0) is your struct to configure a progressive rollout deployment of your flag.
	// It will allow you to ramp up the percentage of your flag over time.
	// You can decide at which percentage you start and at what percentage you ends in your release ramp.
	// Before the start date we will serve the initial percentage and, after we will serve the end percentage.
	Progressive *ProgressiveV0 `json:"progressive,omitempty" yaml:"progressive,omitempty" toml:"progressive,omitempty"` // nolint: lll
}

type V0Rollout struct {
	// Scheduled is your struct to configure an update on some fields of your flag over time.
	// You can add several steps that updates the flag, this is typically used if you want to gradually add more user
	// in your flag.
	Scheduled *ScheduledRolloutV0 `json:"scheduled,omitempty" yaml:"scheduled,omitempty" toml:"scheduled,omitempty"` // nolint: lll
}

type ScheduledRolloutV0 struct {
	// Steps is the list of updates to do in a specific date.
	Steps []ScheduledStepV0 `json:"steps,omitempty" yaml:"steps,omitempty" toml:"steps,omitempty"`
}

type ScheduledStepV0 struct {
	DTO  `yaml:",inline"`
	Date *time.Time `json:"date,omitempty" yaml:"date,omitempty" toml:"date,omitempty"`
}

// ProgressiveV0 is the configuration struct to define a progressive rollout.
type ProgressiveV0 struct {
	// Percentage is where you can configure at what percentage your progressive rollout start
	// and at what percentage it ends.
	// This field is optional
	Percentage ProgressivePercentageV0 `json:"percentage,omitempty" yaml:"percentage,omitempty" toml:"percentage,omitempty"` // nolint: lll

	// ReleaseRamp is the defining when the progressive rollout starts and ends.
	// This field is mandatory if you want to use a progressive rollout.
	// If any field missing we ignore the progressive rollout.
	ReleaseRamp ProgressiveReleaseRampV0 `json:"releaseRamp,omitempty" yaml:"releaseRamp,omitempty" toml:"releaseRamp,omitempty"` // nolint: lll
}

type ProgressivePercentageV0 struct {
	// Initial is the initial percentage before the rollout start date.
	// This field is optional
	// Default: 0.0
	Initial float64 `json:"initial,omitempty" yaml:"initial,omitempty" toml:"initial,omitempty"`

	// End is the target percentage we want to reach at the end of the rollout phase.
	// This field is optional
	// Default: 100.0
	End float64 `json:"end,omitempty" yaml:"end,omitempty" toml:"end,omitempty"`
}

type ProgressiveReleaseRampV0 struct {
	// Start is the starting time of the ramp
	Start *time.Time `json:"start,omitempty" yaml:"start,omitempty" toml:"start,omitempty"`

	// End is the ending time of the ramp
	End *time.Time `json:"end,omitempty" yaml:"end,omitempty" toml:"end,omitempty"`
}
