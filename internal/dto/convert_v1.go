package dto

import "github.com/thomaspoignant/go-feature-flag/internal/flag"

func ConvertV1DtoToInternalFlag(dto DTO) flag.InternalFlag {
	var rollout *flag.Rollout
	if dto.Rollout != nil {
		rollout = &flag.Rollout{}
		if dto.Rollout.V1Rollout.Scheduled != nil {
			rollout.Scheduled = dto.Rollout.V1Rollout.Scheduled
		}

		if dto.Rollout.Experimentation != nil {
			rollout.Experimentation = &flag.ExperimentationRollout{
				Start: dto.Rollout.Experimentation.Start,
				End:   dto.Rollout.Experimentation.End,
			}
		}
	}

	return flag.InternalFlag{
		Variations:  dto.Variations,
		Rules:       dto.Rules,
		DefaultRule: dto.DefaultRule,
		Rollout:     rollout,
		TrackEvents: dto.TrackEvents,
		Disable:     dto.Disable,
		Version:     dto.Version,
	}
}
