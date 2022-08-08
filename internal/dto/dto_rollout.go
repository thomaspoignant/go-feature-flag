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
	Scheduled *[]flag.ScheduledStep `json:"scheduled,omitempty" yaml:"scheduled,omitempty" toml:"scheduled,omitempty"` // nolint: lll
}

type V0Rollout struct {
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
	_ = json.Unmarshal(data, &v1)
	p.V1Rollout = v1

	var v0 V0Rollout
	// we ignore the unmarshal errors because they are expected since we have multiple format
	_ = json.Unmarshal(data, &v0)
	p.V0Rollout = v0

	return nil
}
