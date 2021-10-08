package flagv2

import "time"

type ProgressiveRollout struct {
	// TODO: add comments here
	Percentage  ProgressivePercentage  `json:"percentage,omitempty" yaml:"percentage,omitempty" toml:"percentage,omitempty"`
	ReleaseRamp ProgressiveReleaseRamp `json:"releaseRamp,omitempty" yaml:"releaseRamp,omitempty" toml:"releaseRamp,omitempty"` // nolint: lll
	Variation   ProgressiveVariation   `json:"variation,omitempty" yaml:"variation,omitempty" toml:"variation,omitempty"`       // nolint: lll
}

type ProgressivePercentage struct {
	// Initial is the initial percentage before the rollout start date.
	// This field is optional
	// Default: 0.0
	Initial float64 `json:"initial,omitempty" yaml:"initial,omitempty" toml:"initial,omitempty"`

	// End is the target percentage we want to reach at the end of the rollout phase.
	// This field is optional
	// Default: 100.0
	End float64 `json:"end,omitempty" yaml:"end,omitempty" toml:"end,omitempty"`
}

type ProgressiveReleaseRamp struct {
	// Start is the starting time of the ramp
	Start *time.Time `json:"start,omitempty" yaml:"start,omitempty" toml:"start,omitempty"`

	// End is the ending time of the ramp
	End *time.Time `json:"end,omitempty" yaml:"end,omitempty" toml:"end,omitempty"`
}

type ProgressiveVariation struct {
	// Initial is the initial variation for the rollout.
	Initial *string `json:"initial,omitempty" yaml:"initial,omitempty" toml:"initial,omitempty"`

	// End is the targeted variation.
	End *string `json:"end,omitempty" yaml:"end,omitempty" toml:"end,omitempty"`
}
