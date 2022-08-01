package dto

import (
	"github.com/thomaspoignant/go-feature-flag/internal/flag"
)

var (
	LegacyRuleName  = "legacyRuleV0"
	defaultRuleName = "legacyDefaultRule"
)

var (
	trueVariation    = "True"
	falseVariation   = "False"
	defaultVariation = "Default"
)

// ConvertV0DtoToInternalFlag is converting a flag in the config file to the internal format.
// this function convert only the old format of the flag (before v1.0.0), to keep
// backward support of the configurations.
func ConvertV0DtoToInternalFlag(d DTO, isScheduleStep bool) flag.InternalFlag {
	// Create variations based on the available definition in the flag v0
	var variations *map[string]*interface{}
	newVariations := createVariationsV0(d, isScheduleStep)
	if newVariations != nil {
		variations = &newVariations
	}

	var rules *[]flag.Rule
	if d.Rule != nil && *d.Rule != "" {
		r := make([]flag.Rule, 1)
		r[0] = createLegacyRuleV0(d)
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

	var rollout *flag.Rollout
	if d.Rollout != nil {
		rollout = convertRollout(*d.Rollout, isScheduleStep)
	}

	return flag.InternalFlag{
		Variations:  variations,
		Rules:       rules,
		DefaultRule: defaultRule,
		Rollout:     rollout,
		TrackEvents: d.TrackEvents,
		Disable:     d.Disable,
		Version:     d.Version,
	}
}

// createLegacyRuleV0 will create a rule based on the previous format
func createLegacyRuleV0(d DTO) flag.Rule {
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
	variations := make(map[string]*interface{}, 3)
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

func convertRollout(rollout RolloutV0, isScheduledStep bool) *flag.Rollout {
	r := flag.Rollout{}
	if rollout.Experimentation != nil && rollout.Experimentation.Start != nil && rollout.Experimentation.End != nil {
		r.Experimentation = &flag.ExperimentationRollout{
			Start: rollout.Experimentation.Start,
			End:   rollout.Experimentation.End,
		}
	}

	// it is not allowed to have a scheduled step inside a scheduled step
	if !isScheduledStep && rollout.Scheduled != nil && rollout.Scheduled.Steps != nil {
		var convertedSteps []flag.ScheduledStep
		for _, v := range rollout.Scheduled.Steps {
			converter := "v0"
			toConvert := DTO{
				DTOv0:     v.DTOv0,
				Converter: &converter,
			}
			step := flag.ScheduledStep{
				InternalFlag: ConvertV0DtoToInternalFlag(toConvert, true),
				Date:         v.Date,
			}
			convertedSteps = append(convertedSteps, step)
		}
		r.Scheduled = &convertedSteps
	}

	return &r
}

// computePercentages is creating the percentage structure based on the
// field percentage in the DTO.
func computePercentages(percentage float64) map[string]float64 {
	return map[string]float64{
		trueVariation:  percentage,
		falseVariation: 100 - percentage,
	}
}
