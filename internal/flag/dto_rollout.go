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
	Progressive *Progressive `json:"progressive,omitempty" yaml:"progressive,omitempty" toml:"progressive,omitempty"` // nolint: lll

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
func (dr *DtoRollout) convertRollout() *Rollout {
	if dr == nil {
		return nil
	}

	r := Rollout{
		Experimentation: dr.Experimentation,
	}
	if dr.Scheduled != nil {
		scheduled := ScheduledRollout{Steps: []ScheduledStep{}}
		for _, item := range dr.Scheduled.Steps {
			f, _ := item.ConvertToFlagData(true)
			scheduled.Steps = append(scheduled.Steps, ScheduledStep{
				FlagData: f,
				Date:     item.Date,
			})
		}
		r.Scheduled = &scheduled
	}
	return &r
}
