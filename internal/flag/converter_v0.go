package flag

const legacyRuleName = "defaultRule"

var trueVariation = "True"
var falseVariation = "False"
var defaultVariation = "Default"

// ConvertV0DtoToFlag is converting a flag in the config file to the internal format.
// this function convert only the old format of the flag (before v1.0.0), to keep
// backward support of the configurations.
func ConvertV0DtoToFlag(d DtoFlag, isScheduleStep bool) (FlagData, error) {
	// Create variations based on the available definition in the flag v0
	variations := map[string]*interface{}{}
	if d.True != nil {
		variations[trueVariation] = d.True
	}
	if d.False != nil {
		variations[falseVariation] = d.False
	}
	if d.Default != nil {
		variations[defaultVariation] = d.Default
	}

	// Convert the rule to the new format
	var rules = make(map[string]Rule)
	legacyRule := Rule{}

	// Deal with the percentages.
	if !isScheduleStep || d.Percentage != nil {
		percentages := map[string]float64{}
		percentage := float64(0)
		if d.Percentage != nil {
			percentage = *d.Percentage
		}
		percentages[trueVariation] = percentage
		percentages[falseVariation] = 100 - percentage
		legacyRule.Percentages = &percentages
	}

	// If we have a query.
	if d.Rule != nil {
		legacyRule.Query = d.Rule
	}

	// Handle the specific use case of progressive rollout.
	if d.Rollout != nil &&
		d.Rollout.Progressive != nil &&
		d.Rollout.Progressive.ReleaseRamp.Start != nil &&
		d.Rollout.Progressive.ReleaseRamp.End != nil {
		legacyRule.ProgressiveRollout = &ProgressiveRollout{
			Initial: &ProgressiveRolloutStep{
				Variation:  &falseVariation,
				Percentage: d.Rollout.Progressive.Percentage.Initial,
				Date:       d.Rollout.Progressive.ReleaseRamp.Start,
			},
			End: &ProgressiveRolloutStep{
				Variation:  &trueVariation,
				Percentage: d.Rollout.Progressive.Percentage.End,
				Date:       d.Rollout.Progressive.ReleaseRamp.End,
			},
		}
	}

	// Affect the new rule to the collection of Rules.
	rules[legacyRuleName] = legacyRule

	// Create the flag struct.
	return FlagData{
		Variations:  &variations,
		Rules:       &rules,
		DefaultRule: &Rule{VariationResult: &defaultVariation},
		Rollout:     d.Rollout.convertRollout(),
		TrackEvents: d.TrackEvents,
		Disable:     d.Disable,
		Version:     d.Version,
	}, nil
}
