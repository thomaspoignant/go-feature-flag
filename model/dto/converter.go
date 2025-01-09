package dto

import (
	"github.com/thomaspoignant/go-feature-flag/internal/flag"
)

// ConvertDtoToInternalFlag is converting a DTO to a flag.InternalFlag
func ConvertDtoToInternalFlag(dto DTO) flag.InternalFlag {
	var experimentation *flag.ExperimentationRollout
	if dto.Experimentation != nil {
		experimentation = &flag.ExperimentationRollout{
			Start: dto.Experimentation.Start,
			End:   dto.Experimentation.End,
		}
	}

	// We add the information of the type of query format we have
	var rules *[]flag.Rule
	if dto.Rules != nil {
		rulesTmp := make([]flag.Rule, len(*dto.Rules))
		for i, r := range *dto.Rules {
			r.QueryFormat = flag.GetQueryFormat(r)
			rulesTmp[i] = r
		}
		rules = &rulesTmp
	}

	return flag.InternalFlag{
		BucketingKey:    dto.BucketingKey,
		Variations:      dto.Variations,
		Rules:           rules,
		DefaultRule:     dto.DefaultRule,
		TrackEvents:     dto.TrackEvents,
		Disable:         dto.Disable,
		Version:         dto.Version,
		Scheduled:       dto.Scheduled,
		Experimentation: experimentation,
		Metadata:        dto.Metadata,
	}
}
