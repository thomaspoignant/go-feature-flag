package dto

import (
	"fmt"
	"github.com/thomaspoignant/go-feature-flag/internal/flag"
	"github.com/thomaspoignant/go-feature-flag/testutils/testconvert"
)

var (
	LegacyRuleName  = "legacyRuleV0"
	defaultRuleName = "legacyDefaultRule"

	trueVariation     = "True"
	falseVariation    = "False"
	defaultVariation  = "Default"
	defaultPercentage = float64(0)

	emptyVarRes      = ""
	disableRuleValue = true
	enableRuleValue  = false
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
	var defaultRule *flag.Rule
	if d.Rule != nil && *d.Rule != "" {
		rules = &[]flag.Rule{createLegacyRuleV0(d)}
		defaultRule = createDefaultLegacyRuleV0(d, true)
	} else {
		rules = nil
		defaultRule = createDefaultLegacyRuleV0(d, false)
	}

	internalFlag := flag.InternalFlag{
		Variations:  variations,
		Rules:       rules,
		DefaultRule: defaultRule,
		TrackEvents: d.TrackEvents,
		Disable:     d.Disable,
		Version:     d.Version,
	}

	var rollout *flag.Rollout
	if d.Rollout != nil && (d.Rollout.Experimentation != nil || d.Rollout.Scheduled != nil) {
		rollout = convertRollout(d, internalFlag)
	}

	internalFlag.Rollout = rollout
	return internalFlag
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
			Name:            &defaultRuleName,
			Percentages:     p,
			VariationResult: nil,
		}
	}

	return &flag.Rule{
		Name:            &defaultRuleName,
		VariationResult: &defaultVariation,
		Percentages:     nil,
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

func createScheduledStep(f flag.InternalFlag, dto ScheduledStepV0) flag.ScheduledStep {
	step := flag.ScheduledStep{
		Date: dto.Date,
	}

	variations := createVariationsV0(dto.DTOv0, true)
	step.Variations = &variations

	ruleIndex := f.GetRuleIndexByName(LegacyRuleName)
	hasRuleBefore := ruleIndex != nil && !f.GetRules()[*ruleIndex].IsDisable()
	updateRule := dto.Rule != nil

	// rules management
	switch {

	case hasRuleBefore && !updateRule:
		// deactivate rule + update the default rule
		// activate the target rule
		if dto.Percentage != nil {
			step.Rules = &[]flag.Rule{{
				Name:        testconvert.String(LegacyRuleName),
				Percentages: computePercentages(*dto.Percentage),
			}}
		}
		fmt.Println("1")
		break

	case !hasRuleBefore && updateRule:
		if *dto.Rule == "" {
			// disable target + update default
			step.Rules = &[]flag.Rule{{
				Name:    testconvert.String(LegacyRuleName),
				Disable: &disableRuleValue,
			}}

			step.DefaultRule = &flag.Rule{
				Name: testconvert.String(defaultRuleName),
			}

			if dto.Percentage != nil {
				step.DefaultRule.Percentages = computePercentages(*dto.Percentage)
			}
		} else {
			// TODO: Update target
		}
		fmt.Println("2")
		break

	case !hasRuleBefore && !updateRule:
		//update the defaultRule
		if dto.Percentage != nil {
			step.DefaultRule = &flag.Rule{
				VariationResult: &emptyVarRes,
				Name:            testconvert.String(defaultRuleName),
				Percentages:     computePercentages(*dto.Percentage),
			}
		}
		fmt.Println("3")
		break

	case hasRuleBefore && updateRule:
		// update targeting
		if dto.Rule != nil {
			r := flag.Rule{
				Name:    testconvert.String(LegacyRuleName),
				Query:   dto.Rule,
				Disable: &enableRuleValue,
			}
			if dto.Percentage != nil {
				r.VariationResult = &emptyVarRes
				r.Percentages = computePercentages(*dto.Percentage)
			}
			step.Rules = &[]flag.Rule{r}
		} else {
			// TODO: disable targeting and update default
		}
		fmt.Println("4")
		break
	}

	// we have only a default rule
	//if dto.Rule == nil && ruleIndex == nil {
	//	if dto.Percentage != nil {
	//		step.DefaultRule = &flag.Rule{
	//			Name:        &defaultRuleName,
	//			Percentages: computePercentages(*dto.Percentage),
	//		}
	//	}
	//}
	//
	//if dto.Rule != nil && *dto.Rule == "" && ruleIndex != nil {
	//	disable := true
	//	r := flag.Rule{
	//		Name:    testconvert.String(LegacyRuleName),
	//		Disable: &disable,
	//	}
	//	step.Rules = &[]flag.Rule{r}
	//}
	//
	//// We add a rule before or we update the query
	//if dto.Rule != nil || (ruleIndex != nil && !f.GetRules()[*ruleIndex].IsDisable()) {
	//	r := flag.Rule{
	//		Name:  testconvert.String(LegacyRuleName),
	//		Query: dto.Rule,
	//	}
	//
	//	if dto.Percentage != nil {
	//		if ruleIndex != nil && f.GetRules()[*ruleIndex].VariationResult != nil {
	//			r.VariationResult = testconvert.String("")
	//		}
	//		r.Percentages = computePercentages(*dto.Percentage)
	//	} else if ruleIndex != nil && f.GetRules()[*ruleIndex].Percentages != nil {
	//		r.Percentages = deepCopyPercentages(f.GetRules()[*ruleIndex].GetPercentages())
	//	} else if ruleIndex == nil && f.GetDefaultRule() != nil {
	//		r.Percentages = deepCopyPercentages(f.GetDefaultRule().GetPercentages())
	//	}
	//
	//	// if we did not have rules before we migrate the percentage into the rule
	//	if ruleIndex == nil && dto.Percentage == nil {
	//		r.Percentages = deepCopyPercentages(f.GetDefaultRule().GetPercentages())
	//	}
	//
	//	// remove all percentages from default rule
	//	if f.GetDefaultRule().Percentages != nil {
	//		removedPercentages := map[string]float64{}
	//		for k, _ := range f.GetDefaultRule().GetPercentages() {
	//			removedPercentages[k] = -1
	//		}
	//		step.DefaultRule = &flag.Rule{
	//			Percentages:     &removedPercentages,
	//			Name:            &defaultRuleName,
	//			VariationResult: &defaultVariation,
	//		}
	//	}
	//
	//	step.Rules = &[]flag.Rule{r}
	//}

	if dto.Disable != nil {
		step.Disable = dto.Disable
	}

	if dto.TrackEvents != nil {
		step.TrackEvents = dto.TrackEvents
	}

	if dto.Version != nil {
		step.Version = dto.Version
	}

	return step
}

func convertRollout(dto DTOv0, f flag.InternalFlag) *flag.Rollout {
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
	if dto.Rollout.Scheduled != nil && dto.Rollout.Scheduled.Steps != nil {
		var convertedSteps []flag.ScheduledStep
		for _, v := range dto.Rollout.Scheduled.Steps {
			convertedSteps = append(convertedSteps, createScheduledStep(f, v))
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

func deepCopyPercentages(in map[string]float64) *map[string]float64 {
	p := make(map[string]float64, len(in))
	// deep copy of the percentages to avoid being override
	for k, v := range in {
		p[k] = v
	}
	return &p
}
