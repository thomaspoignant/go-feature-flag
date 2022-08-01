package dto

import (
	"github.com/thomaspoignant/go-feature-flag/internal/flag"
)

var (
	LegacyRuleName  = "legacyRuleV0"
	defaultRuleName = "legacyDefaultRule"

	trueVariation     = "True"
	falseVariation    = "False"
	defaultVariation  = "Default"
	defaultPercentage = float64(0)
)

// ConvertV0DtoToInternalFlag is converting a flag in the config file to the internal format.
// this function convert only the old format of the flag (before v1.0.0), to keep
// backward support of the configurations.
func ConvertV0DtoToInternalFlag(d DTOv0, isScheduledStep bool) flag.InternalFlag {
	// Create variations based on the available definition in the flag v0
	var variations *map[string]*interface{}
	newVariations := createVariationsV0(d, isScheduledStep)
	if newVariations != nil {
		variations = &newVariations
	}

	var rules *[]flag.Rule
	if d.Rule != nil && *d.Rule != "" {
		r := make([]flag.Rule, 1)
		r[0] = createLegacyRuleV0(d)
		rules = &r
	}

	// Default rule
	hasTargetRule := rules != nil && len(*rules) > 0
	defaultRule := createDefaultLegacyRuleV0(d, hasTargetRule)

	var rollout *flag.Rollout
	if d.Rollout != nil && (d.Rollout.Experimentation != nil || d.Rollout.Scheduled != nil) {
		rollout = convertRollout(d, isScheduledStep)
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

// createDefaultLegacyRuleV0 create the default rule based on the legacy format.
func createDefaultLegacyRuleV0(d DTOv0, hasTargetRule bool) *flag.Rule {
	hasProgressiveRollout := d.Rollout != nil &&
		d.Rollout.Progressive != nil &&
		d.Rollout.Progressive.ReleaseRamp.Start != nil &&
		d.Rollout.Progressive.ReleaseRamp.End != nil

	if hasProgressiveRollout && !hasTargetRule {
		return &flag.Rule{
			Name: &defaultRuleName,
			ProgressiveRollout: &flag.ProgressiveRollout{
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
			},
		}
	}
	if d.Rule == nil {
		if d.Percentage == nil {
			d.Percentage = &defaultPercentage
		}

		p := computePercentages(*d.Percentage)
		return &flag.Rule{
			Name:        &defaultRuleName,
			Percentages: p,
		}
	}

	return &flag.Rule{
		Name:            &defaultRuleName,
		VariationResult: &defaultVariation,
	}
}

// createLegacyRuleV0 will create a rule based on the previous format
func createLegacyRuleV0(d DTOv0) flag.Rule {
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

	var percentages *map[string]float64
	if progressiveRollout == nil {
		if d.Percentage != nil {
			percentages = computePercentages(*d.Percentage)
		} else {
			percentages = &map[string]float64{
				trueVariation:  0,
				falseVariation: 100,
			}
		}
	}

	return flag.Rule{
		Name:               &LegacyRuleName,
		Query:              d.Rule,
		Percentages:        percentages,
		ProgressiveRollout: progressiveRollout,
	}
}

// createVariationsV0 will create a set of variations based on the previous format
func createVariationsV0(d DTOv0, isScheduleStep bool) map[string]*interface{} {
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

func convertRollout(dto DTOv0, isScheduledStep bool) *flag.Rollout {
	r := flag.Rollout{}
	if dto.Rollout.Experimentation != nil &&
		dto.Rollout.Experimentation.Start != nil &&
		dto.Rollout.Experimentation.End != nil {
		r.Experimentation = &flag.ExperimentationRollout{
			Start: dto.Rollout.Experimentation.Start,
			End:   dto.Rollout.Experimentation.End,
		}
	}

	// it is not allowed to have a scheduled step inside a scheduled step
	if !isScheduledStep && dto.Rollout.Scheduled != nil && dto.Rollout.Scheduled.Steps != nil {
		var convertedSteps []flag.ScheduledStep
		initialDto := dto
		scheduledSteps := dto.Rollout.Scheduled.Steps
		initialDto.Rollout.Scheduled = nil
		for _, v := range scheduledSteps {
			// Hack for simplicity
			// When we translate the DTOvO scheduled step into a flag.ScheduledStep what we are doing
			// is patching the flag the old way and putting the whole patch into the schedule step.
			initialDto = mergeDtoScheduledStep(initialDto, v.DTOv0)
			step := flag.ScheduledStep{
				InternalFlag: ConvertV0DtoToInternalFlag(initialDto, false),
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
func computePercentages(percentage float64) *map[string]float64 {
	return &map[string]float64{
		trueVariation:  percentage,
		falseVariation: 100 - percentage,
	}
}

func mergeDtoScheduledStep(origin DTOv0, toBeMerged DTOv0) DTOv0 {
	if toBeMerged.Disable != nil {
		origin.Disable = toBeMerged.Disable
	}
	if toBeMerged.False != nil {
		origin.False = toBeMerged.False
	}
	if toBeMerged.True != nil {
		origin.True = toBeMerged.True
	}
	if toBeMerged.Default != nil {
		origin.Default = toBeMerged.Default
	}
	if toBeMerged.TrackEvents != nil {
		origin.TrackEvents = toBeMerged.TrackEvents
	}
	if toBeMerged.Percentage != nil {
		origin.Percentage = toBeMerged.Percentage
	}
	if toBeMerged.Rule != nil {
		origin.Rule = toBeMerged.Rule
	}
	if toBeMerged.Rollout != nil {
		origin.Rollout = toBeMerged.Rollout
	}
	if toBeMerged.Version != nil {
		origin.Version = toBeMerged.Version
	}
	return origin
}
