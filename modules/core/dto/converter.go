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

func ConvertInternalFlagToDto(flag flag.InternalFlag) DTO {
	experimentation := &ExperimentationDto{}
	if flag.Experimentation != nil {
		experimentation = &ExperimentationDto{
			Start: flag.Experimentation.Start,
			End:   flag.Experimentation.End,
		}
	}

	return DTO{
		TrackEvents:     flag.TrackEvents,
		Disable:         flag.Disable,
		Version:         flag.Version,
		Variations:      flag.Variations,
		Rules:           flag.Rules,
		BucketingKey:    flag.BucketingKey,
		DefaultRule:     flag.DefaultRule,
		Scheduled:       flag.Scheduled,
		Experimentation: experimentation,
		Metadata:        flag.Metadata,
	}
}
