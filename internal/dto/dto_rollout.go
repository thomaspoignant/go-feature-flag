package dto

import (
	"encoding/json"
	"time"

	"github.com/thomaspoignant/go-feature-flag/internal/flag"
)

type ExperimentationDto struct {
	// Start is the starting time of the experimentation
	Start *time.Time `json:"start,omitempty" yaml:"start,omitempty" toml:"start,omitempty"`

	// End is the ending time of the experimentation
	End *time.Time `json:"end,omitempty" yaml:"end,omitempty" toml:"end,omitempty"`
}

type Rollout struct {
	// CommonRollout is the struct containing the configuration for rollout that applies for
	// all types of flag in your input file.
	CommonRollout `json:",inline" yaml:",inline" toml:",inline"`

	// V1Rollout contains the configuration available only for the flags version v1.X.X
	V1Rollout `json:",inline" yaml:",inline" toml:",inline"`

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

type V1Rollout struct {
	// Scheduled is your struct to configure an update on some fields of your flag over time.
	// You can add several steps that updates the flag, this is typically used if you want to gradually add more user
	// in your flag.
	Scheduled *[]flag.ScheduledStep `json:"scheduled,omitempty" yaml:"scheduled,omitempty" toml:"scheduled,omitempty"` // nolint: lll
}

type V0Rollout struct {
	// Scheduled is your struct to configure an update on some fields of your flag over time.
	// You can add several steps that updates the flag, this is typically used if you want to gradually add more user
	// in your flag.
	Scheduled *ScheduledRolloutV0 `json:"scheduled,omitempty" yaml:"scheduled,omitempty" toml:"scheduled,omitempty"` // nolint: lll
}

// UnmarshalJSON is dealing with the fact that we have 2 different entry called scheduled
// in different format, this can cause an issue when unmarshalling the data.
// This is the reason why we have a custom unmarshalling function for this struct.
func (p *Rollout) UnmarshalJSON(data []byte) error {
	var c CommonRollout
	err := json.Unmarshal(data, &c)
	if err != nil {
		return err
	}
	p.CommonRollout = c

	var v1 V1Rollout
	// we ignore the unmarshal errors because they are expected since we have multiple format
	err = json.Unmarshal(data, &v1)
	if err != nil {
		// TODO: add log in debug only
	}
	if v1.Scheduled != nil && *v1.Scheduled != nil {
		p.V1Rollout = v1
	}

	var v0 V0Rollout
	// we ignore the unmarshal errors because they are expected since we have multiple format
	err = json.Unmarshal(data, &v0)
	if err != nil {
		// TODO: add log in debug only
	}
	p.V0Rollout = v0

	return nil
}

// UnmarshalTOML is used for TOML unmarshalling, the lib is not calling directly UnmarshalJSON,
// so we are calling it after marshaling input in JSON string
func (p *Rollout) UnmarshalTOML(input interface{}) error {
	jsonStr, err := json.Marshal(input)
	if err != nil {
		// TODO: add log in debug only
		return err
	}
	return p.UnmarshalJSON(jsonStr)
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
