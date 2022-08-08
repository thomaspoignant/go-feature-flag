package dto

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
