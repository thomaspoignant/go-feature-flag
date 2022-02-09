package flag

import (
	"time"
)

type DtoRollout struct {
	// Experimentation is your struct to configure an experimentation, it will allow you to configure a start date and
	// an end date for your flag.
	// When the experimentation is not running, the flag will serve the default value.
	Experimentation *Experimentation `json:"experimentation,omitempty" yaml:"experimentation,omitempty" toml:"experimentation,omitempty"` // nolint: lll

	// Progressive is your struct to configure a progressive rollout deployment of your flag.
	// It will allow you to ramp up the percentage of your flag over time.
	// You can decide at which percentage you starts and at what percentage you ends in your release ramp.
	// Before the start date we will serve the initial percentage and after we will serve the end percentage.
	Progressive *DtoProgressiveRollout `json:"progressive,omitempty" yaml:"progressive,omitempty" toml:"progressive,omitempty"` // nolint: lll

	// Scheduled is your struct to configure an update on some fields of your flag over time.
	// You can add several steps that updates the flag, this is typically used if you want to gradually add more user
	// in your flag.
	Scheduled *DtoScheduledRollout `json:"scheduled,omitempty" yaml:"scheduled,omitempty" toml:"scheduled,omitempty"` // nolint: lll
}

type DtoScheduledRollout struct {
	// Steps is the list of updates to do in a specific date.
	Steps []DtoScheduledStep `json:"steps,omitempty" yaml:"steps,omitempty" toml:"steps,omitempty"`
}

type DtoScheduledStep struct {
	DtoFlag `yaml:",inline"`
	Date    *time.Time `json:"date,omitempty" yaml:"date,omitempty" toml:"date,omitempty"`
}

// convertRollout convert the rollout configuration in a Flag_data format.
func (dr *DtoRollout) convertRollout(version int) *Rollout {
	if dr == nil || (dr.Scheduled == nil && dr.Experimentation == nil) {
		return nil
	}

	r := Rollout{
		Experimentation: dr.Experimentation,
	}
	if dr.Scheduled != nil {
		scheduled := ScheduledRollout{Steps: []ScheduledStep{}}
		for _, item := range dr.Scheduled.Steps {
			var convertedDtoFlag FlagData
			if version < 1 {
				convertedDtoFlag = ConvertV0DtoToFlag(item.DtoFlag, true)
			} else {
				convertedDtoFlag = ConvertV1DtoToFlag(item.DtoFlag, true)
			}
			scheduled.Steps = append(scheduled.Steps, ScheduledStep{
				FlagData: convertedDtoFlag,
				Date:     item.Date,
			})
		}
		r.Scheduled = &scheduled
	}
	return &r
}

// DtoProgressiveRollout is the configuration struct to define a progressive rollout.
type DtoProgressiveRollout struct {
	// Percentage is where you can configure at what percentage your progressive rollout start
	// and at what percentage it ends.
	// This field is optional
	Percentage DtoProgressivePercentage `json:"percentage,omitempty" yaml:"percentage,omitempty" toml:"percentage,omitempty"` // nolint: lll

	// ReleaseRamp is the defining when the progressive rollout starts and ends.
	// This field is mandatory if you want to use a progressive rollout.
	// If any field missing we ignore the progressive rollout.
	ReleaseRamp DtoProgressiveReleaseRamp `json:"releaseRamp,omitempty" yaml:"releaseRamp,omitempty" toml:"releaseRamp,omitempty"` // nolint: lll
}

type DtoProgressivePercentage struct {
	// Initial is the initial percentage before the rollout start date.
	// This field is optional
	// Default: 0.0
	Initial float64 `json:"initial,omitempty" yaml:"initial,omitempty" toml:"initial,omitempty"`

	// End is the target percentage we want to reach at the end of the rollout phase.
	// This field is optional
	// Default: 100.0
	End float64 `json:"end,omitempty" yaml:"end,omitempty" toml:"end,omitempty"`
}

type DtoProgressiveReleaseRamp struct {
	// Start is the starting time of the ramp
	Start *time.Time `json:"start,omitempty" yaml:"start,omitempty" toml:"start,omitempty"`

	// End is the ending time of the ramp
	End *time.Time `json:"end,omitempty" yaml:"end,omitempty" toml:"end,omitempty"`
}
