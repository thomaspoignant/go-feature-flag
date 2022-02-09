package flag

const LegacyRuleName = "legacyRuleV0"

var trueVariation = "True"
var falseVariation = "False"
var defaultVariation = "Default"

// ConvertV0DtoToFlag is converting a flag in the config file to the internal format.
// this function convert only the old format of the flag (before v1.0.0), to keep
// backward support of the configurations.
func ConvertV0DtoToFlag(d DtoFlag, isScheduleStep bool) FlagData {
	// Create variations based on the available definition in the flag v0
	var variations *map[string]*interface{}
	newVariations := createVariationsV0(d, isScheduleStep)
	if newVariations != nil {
		variations = &newVariations
	}

	// Convert the rule to the new format
	var rules *map[string]Rule
	legacyRule := createLegacyRuleV0(d, isScheduleStep)
	if legacyRule != nil {
		rules = &map[string]Rule{LegacyRuleName: *legacyRule}
	}

	var defaultRule *Rule
	if !isScheduleStep {
		defaultRule = &Rule{VariationResult: &defaultVariation}
	}

	r := d.Rollout.convertRollout(0)
	// Create the flag struct.
	return FlagData{
		Variations:  variations,
		DefaultRule: defaultRule,
		Rules:       rules,
		Rollout:     r,
		TrackEvents: d.TrackEvents,
		Disable:     d.Disable,
		Version:     d.Version,
	}
}

// createLegacyRuleV0 will create a rule based on the previous format
func createLegacyRuleV0(d DtoFlag, isScheduleStep bool) *Rule {
	legacyRule := Rule{}

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

	// Deal with the percentages.
	if (!isScheduleStep || d.Percentage != nil) && legacyRule.ProgressiveRollout == nil {
		percentages := map[string]float64{}
		percentage := float64(0)
		if d.Percentage != nil {
			percentage = *d.Percentage
		}
		percentages[trueVariation] = percentage
		percentages[falseVariation] = 100 - percentage
		legacyRule.Percentages = &percentages
	}

	if legacyRule == (Rule{}) {
		return nil
	}

	return &legacyRule
}

// createVariationsV0 will create a set of variations based on the previous format
func createVariationsV0(d DtoFlag, isScheduleStep bool) map[string]*interface{} {
	var variations = make(map[string]*interface{}, 3)
	if d.True != nil {
		variations[trueVariation] = d.True
	}
	if d.False != nil {
		variations[falseVariation] = d.False
	}
	if d.Default != nil {
		variations[defaultVariation] = d.Default
	}

	if isScheduleStep && len(variations) == 0 {
		variations = nil
	}
	return variations
}
