package dto

import "github.com/thomaspoignant/go-feature-flag/internal/flag"

var LegacyRuleName = "legacyRuleV0"
var defaultRuleName = "legacyDefaultRule"

var trueVariation = "True"
var falseVariation = "False"
var defaultVariation = "Default"

// ConvertV0DtoToFlag is converting a flag in the config file to the internal format.
// this function convert only the old format of the flag (before v1.0.0), to keep
// backward support of the configurations.
func ConvertV0DtoToFlag(d DTO, isScheduleStep bool) flag.InternalFlag {
	// Create variations based on the available definition in the flag v0
	var variations *map[string]*interface{}
	newVariations := createVariationsV0(d, isScheduleStep)
	if newVariations != nil {
		variations = &newVariations
	}

	var rules *[]flag.Rule
	if d.Rule != nil && *d.Rule != "" {
		r := make([]flag.Rule, 1)
		r[0] = createLegacyRuleV0(d, false)
		rules = &r
	}

	// Percentage for the default rule
	var defaultRule *flag.Rule
	if (d.Rule == nil || *d.Rule == "") && d.Percentage != nil {
		p := computePercentages(*d.Percentage)
		defaultRule = &flag.Rule{
			Name:        &defaultRuleName,
			Percentages: &p,
		}
	} else {
		defaultRule = &flag.Rule{
			Name:            &defaultRuleName,
			VariationResult: &defaultVariation,
		}
	}

	return flag.InternalFlag{
		Variations:  variations,
		Rules:       rules,
		DefaultRule: defaultRule,
		Rollout:     nil,
		TrackEvents: d.TrackEvents,
		Disable:     d.Disable,
		Version:     d.Version,
	}
}

// createLegacyRuleV0 will create a rule based on the previous format
func createLegacyRuleV0(d DTO, isScheduleStep bool) flag.Rule {
	// Handle the specific use case of progressive rollout.
	var progressiveRollout *flag.ProgressiveRollout
	if d.Rollout != nil &&
		d.Rollout.Progressive != nil &&
		d.Rollout.Progressive.ReleaseRamp.Start != nil &&
		d.Rollout.Progressive.ReleaseRamp.End != nil {
		progressiveRollout = &flag.ProgressiveRollout{
			Initial: &flag.ProgressiveRolloutStep{
				Variation:  &falseVariation,
				Percentage: &d.Rollout.Progressive.Percentage.Initial,
				Date:       d.Rollout.Progressive.ReleaseRamp.Start,
			},
			End: &flag.ProgressiveRolloutStep{
				Variation:  &trueVariation,
				Percentage: &d.Rollout.Progressive.Percentage.End,
				Date:       d.Rollout.Progressive.ReleaseRamp.End,
			},
		}
	}

	var percentages map[string]float64
	if d.Percentage != nil {
		percentages = computePercentages(*d.Percentage)
	} else {
		percentages = map[string]float64{
			trueVariation:  0,
			falseVariation: 100,
		}
	}
	legacyRule := flag.Rule{
		Name:               &LegacyRuleName,
		Query:              d.Rule,
		Percentages:        &percentages,
		ProgressiveRollout: progressiveRollout,
	}

	return legacyRule
}

// createVariationsV0 will create a set of variations based on the previous format
func createVariationsV0(d DTO, isScheduleStep bool) map[string]*interface{} {
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

func computePercentages(percentage float64) map[string]float64 {
	return map[string]float64{
		trueVariation:  percentage,
		falseVariation: 100 - percentage,
	}
}
