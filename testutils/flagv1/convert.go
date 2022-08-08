package flagv1

import "github.com/thomaspoignant/go-feature-flag/internal/dto"

func ConvertDtoToV1(d dto.DTO) FlagData {
	var r Rollout

	if d.RolloutV0 != nil {
		r = Rollout{}

		if d.RolloutV0.Progressive != nil {
			r.Progressive = &Progressive{}
			r.Progressive.ReleaseRamp.Start = d.RolloutV0.Progressive.ReleaseRamp.Start
			r.Progressive.ReleaseRamp.End = d.RolloutV0.Progressive.ReleaseRamp.End
			r.Progressive.Percentage.End = d.RolloutV0.Progressive.Percentage.End
			r.Progressive.Percentage.Initial = d.RolloutV0.Progressive.Percentage.Initial
		}

		if d.RolloutV0.Scheduled != nil {
			r.Scheduled = &ScheduledRollout{}
			if d.RolloutV0.Scheduled.Steps != nil {
				r.Scheduled.Steps = []ScheduledStep{}
				for _, step := range d.RolloutV0.Scheduled.Steps {
					f := FlagData{
						Rule:        step.Rule,
						Percentage:  step.Percentage,
						True:        step.True,
						False:       step.False,
						Default:     step.Default,
						TrackEvents: step.TrackEvents,
						Disable:     step.Disable,
					}
					if step.RolloutV0 != nil && step.RolloutV0.Progressive != nil {
						f.Rollout = &Rollout{
							Progressive: &Progressive{
								ReleaseRamp: ProgressiveReleaseRamp{
									Start: step.RolloutV0.Progressive.ReleaseRamp.Start,
									End:   step.RolloutV0.Progressive.ReleaseRamp.End,
								},
								Percentage: ProgressivePercentage{
									Initial: step.RolloutV0.Progressive.Percentage.Initial,
									End:     step.RolloutV0.Progressive.Percentage.End,
								},
							},
						}
					}

					s := ScheduledStep{
						FlagData: f,
						Date:     step.Date,
					}
					r.Scheduled.Steps = append(r.Scheduled.Steps, s)
				}
			}
		}

		if d.RolloutV0.Experimentation != nil {
			r.Experimentation = &Experimentation{}
			if d.RolloutV0.Experimentation.Start != nil {
				r.Experimentation.Start = d.RolloutV0.Experimentation.Start
			}
			if d.RolloutV0.Experimentation.End != nil {
				r.Experimentation.End = d.RolloutV0.Experimentation.End
			}
		}
	}

	return FlagData{
		Rule:        d.Rule,
		Percentage:  d.Percentage,
		True:        d.True,
		False:       d.False,
		Default:     d.Default,
		TrackEvents: d.TrackEvents,
		Disable:     d.Disable,
		Rollout:     &r,
	}
}
