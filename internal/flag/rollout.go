package flag

import "strings"

type Rollout struct {
	// Experimentation is your struct to configure an experimentation, it will allow you to configure a start date and
	// an end date for your flag.
	// When the experimentation is not running, the flag will serve the default value.
	Experimentation *Experimentation `json:"experimentation,omitempty" yaml:"experimentation,omitempty" toml:"experimentation,omitempty"` // nolint: lll

	// Scheduled is your struct to configure an update on some fields of your flag over time.
	// You can add several steps that updates the flag, this is typically used if you want to gradually add more user
	// in your flag.
	Scheduled *ScheduledRollout `json:"scheduled,omitempty" yaml:"scheduled,omitempty" toml:"scheduled,omitempty"` // nolint: lll
}

func (e Rollout) String() string {
	if e.Experimentation == nil && e.Scheduled == nil {
		return ""
	}
	str := make([]string, 0)
	appendIfHasValue(str, "experimentation", e.Experimentation.String())
	appendIfHasValue(str, "scheduled", e.Scheduled.String())
	return strings.Join(str, ",")
}
