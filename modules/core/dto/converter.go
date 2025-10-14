package dto

import (
	"github.com/thomaspoignant/go-feature-flag/modules/core/flag"
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

	return flag.InternalFlag{
		BucketingKey:    dto.BucketingKey,
		Variations:      dto.Variations,
		Rules:           dto.Rules,
		DefaultRule:     dto.DefaultRule,
		TrackEvents:     dto.TrackEvents,
		Disable:         dto.Disable,
		Version:         dto.Version,
		Scheduled:       dto.Scheduled,
		Experimentation: experimentation,
		Metadata:        dto.Metadata,
	}
}
