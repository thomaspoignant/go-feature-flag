package flag

import (
	"encoding/json"
	"github.com/thomaspoignant/go-feature-flag/internal/flagv1"
	"github.com/thomaspoignant/go-feature-flag/internal/flagv2"
	"github.com/thomaspoignant/go-feature-flag/internal/rollout"
	"strconv"
)

type DtoRollout struct {
	// Experimentation is your struct to configure an experimentation, it will allow you to configure a start date and
	// an end date for your flag.
	// When the experimentation is not running, the flag will serve the default value.
	Experimentation *rollout.Experimentation `json:"experimentation,omitempty" yaml:"experimentation,omitempty" toml:"experimentation,omitempty"` // nolint: lll

	// Progressive is your struct to configure a progressive rollout deployment of your flag.
	// It will allow you to ramp up the percentage of your flag over time.
	// You can decide at which percentage you starts and at what percentage you ends in your release ramp.
	// Before the start date we will serve the initial percentage and after we will serve the end percentage.
	Progressive *flagv1.Progressive `json:"progressive,omitempty" yaml:"progressive,omitempty" toml:"progressive,omitempty"` // nolint: lll

	// Scheduled is your struct to configure an update on some fields of your flag over time.
	// You can add several steps that updates the flag, this is typically used if you want to gradually add more user
	// in your flag.
	Scheduled *map[string]interface{} `json:"scheduled,omitempty" yaml:"scheduled,omitempty" toml:"scheduled,omitempty"` // nolint: lll
}

func (dr *DtoRollout) toRolloutV1() *flagv1.Rollout {
	if dr == nil {
		return nil
	}

	r := flagv1.Rollout{
		Experimentation: dr.Experimentation,
		Progressive:     dr.Progressive,
	}
	if dr.Scheduled != nil {
		scheduled := flagv1.ScheduledRollout{}
		jsonString, _ := json.Marshal(dr.Scheduled)
		err := json.Unmarshal(jsonString, &scheduled)
		if err != nil {
			// TODO: log impossible to read scheduled
			return nil
		}
		*r.Scheduled = scheduled
	}
	return &r
}

func (dr *DtoRollout) toRolloutV2() *flagv2.Rollout {
	if dr == nil {
		return nil
	}

	r := flagv2.Rollout{
		Experimentation: dr.Experimentation,
	}
	if dr.Scheduled != nil {
		scheduled := flagv2.ScheduledRollout{}
		jsonString, _ := json.Marshal(dr.Scheduled)
		err := json.Unmarshal(jsonString, &scheduled)
		if err != nil {
			// TODO: log impossible to read scheduled
			return nil
		}
		*r.Scheduled = scheduled
	}
	return &r
}

// DtoFlag contains all the flag from flagv1 and flagv2
// we need it because it allows to have multiple flag formats in the same file.
type DtoFlag struct {
	// --- FLAGv1 FIELDS ---
	// Rule is the query use to select on which user the flag should apply.
	// Rule format is based on the nikunjy/rules module.
	// If no rule set, the flag apply to all users (percentage still apply).
	Rule *string `json:"rule,omitempty" yaml:"rule,omitempty" toml:"rule,omitempty"`

	// Percentage of the users affect by the flag.
	// Default value is 0
	Percentage *float64 `json:"percentage,omitempty" yaml:"percentage,omitempty" toml:"percentage,omitempty"`

	// True is the value return by the flag if apply to the user (rule is evaluated to true)
	// and user is in the active percentage.
	True *interface{} `json:"true,omitempty" yaml:"true,omitempty" toml:"true,omitempty"`

	// False is the value return by the flag if apply to the user (rule is evaluated to true)
	// and user is not in the active percentage.
	False *interface{} `json:"false,omitempty" yaml:"false,omitempty" toml:"false,omitempty"`

	// Default is the value return by the flag if not apply to the user (rule is evaluated to false).
	Default *interface{} `json:"default,omitempty" yaml:"default,omitempty" toml:"default,omitempty"`

	// --- COMMON FIELDS ---
	// Rollout is how we rollout the flag
	Rollout *DtoRollout `json:"rollout,omitempty" yaml:"rollout,omitempty" toml:"rollout,omitempty"`

	// TrackEvents is false if you don't want to export the data in your data exporter.
	// Default value is true
	TrackEvents *bool `json:"trackEvents,omitempty" yaml:"trackEvents,omitempty" toml:"trackEvents,omitempty"`

	// Disable is true if the flag is disabled.
	Disable *bool `json:"disable,omitempty" yaml:"disable,omitempty" toml:"disable,omitempty"`

	// Version (optional) This field contains the version of the flag.
	// The version is manually managed when you configure your flags and it is used to display the information
	// in the notifications and data collection.
	Version *string `json:"version,omitempty" yaml:"version,omitempty" toml:"version,omitempty"`

	// --- FLAGv2 FIELDS ---
	// Variations are all the variations available for this flag.
	// The minimum is 2 variations and we don't have any max limit except
	// if the variationValue is a bool, the max is 2.
	Variations *map[string]*interface{} `json:"variations,omitempty" yaml:"variations,omitempty" toml:"variations,omitempty"`

	// Rules is the list of Rule for this flag.
	// This an optional field.
	Rules *[]flagv2.Rule `json:"targeting,omitempty" yaml:"targeting,omitempty" toml:"targeting,omitempty"`

	// DefaultRule is the rule applied after checking that any other rules
	// matched the user.
	DefaultRule *flagv2.Rule `json:"defaultRule,omitempty" yaml:"defaultRule,omitempty" toml:"defaultRule,omitempty"`
}

func (dto *DtoFlag) IsFlagV2() bool {
	return dto.Variations != nil
}

func (dto *DtoFlag) ToFlagV1() flagv2.FlagData {

	variations := map[string]*interface{}{
		"True":    dto.True,
		"False":   dto.False,
		"Default": dto.Default,
	}
	defaultVariation := "Default"

	var rules []flagv2.Rule
	var defaultRule flagv2.Rule
	if dto.Rule != nil && *dto.Rule != "" {

		percentage := float64(0)
		if dto.Percentage != nil {
			percentage = *dto.Percentage
		}
		var percentages []flagv2.VariationPercentage
		percentages = append(percentages, flagv2.VariationPercentage{"True": percentage})
		percentages = append(percentages, flagv2.VariationPercentage{"False": 100 - percentage})

		newRule := flagv2.Rule{
			Percentages: &percentages,
			Query:       dto.Rule,
		}
		rules = append(rules, newRule)

		defaultRule = flagv2.Rule{
			VariationResult: &defaultVariation,
		}
	} else {
		percentage := float64(0)
		if dto.Percentage != nil {
			percentage = *dto.Percentage
		}
		var percentages []flagv2.VariationPercentage
		percentages = append(percentages, flagv2.VariationPercentage{"True": percentage})
		percentages = append(percentages, flagv2.VariationPercentage{"False": 100 - percentage})

		defaultRule = flagv2.Rule{
			Percentages: &percentages,
		}
	}

	return flagv2.FlagData{
		Variations:  &variations,
		Rules:       &rules,
		DefaultRule: &defaultRule,
		Rollout:     nil,
		TrackEvents: dto.TrackEvents,
		Disable:     dto.Disable,
		Version:     dto.Version,
	}

	//	this.variations = new Map<string, string | number | object | boolean>()
	//	this.variations.set('true', legacyFlag.true)
	//	this.variations.set('false', legacyFlag.false)
	//	this.variations.set('default', legacyFlag.default)
	//this.disable = legacyFlag.disable
	//this.trackEvents = legacyFlag.trackEvents
	//
	//// percentage
	//if (legacyFlag.rule !== undefined && legacyFlag.rule !== ''){
	//const rule = new Rule()
	//rule.query = legacyFlag.rule
	//rule.percentage = new Array<object>()
	//const percentage = legacyFlag.percentage || 0
	//rule.percentage.push({'true': percentage})
	//rule.percentage.push({'false': 100 - percentage})
	//
	//const defaultRule = new Rule()
	//defaultRule.variation = 'default'
	//
	//this.targeting = [rule]
	//this.defaultRule = defaultRule
	//} else {
	//const defaultRule = new Rule()
	//defaultRule.percentage = new Array<object>()
	//const percentage = legacyFlag.percentage || 0
	//defaultRule.percentage.push({'true': percentage})
	//defaultRule.percentage.push({'false': 100 - percentage})
	//this.defaultRule = defaultRule
	//}
	//}

}

func (dto *DtoFlag) ToFlagV2() flagv2.FlagData {
	return flagv2.FlagData{
		Variations:  dto.Variations,
		Rules:       dto.Rules,
		DefaultRule: dto.DefaultRule,
		Rollout:     dto.Rollout.toRolloutV2(),
		TrackEvents: dto.TrackEvents,
		Disable:     dto.Disable,
		Version:     dto.Version,
	}
}

func convertVersionToFloat64(version *string) *float64 {
	if version == nil {
		return nil
	}

	res, err := strconv.ParseFloat(*version, 64)
	if err != nil {
		return nil
	}

	return &res
}
