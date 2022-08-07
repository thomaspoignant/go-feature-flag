package flagv1

import "github.com/thomaspoignant/go-feature-flag/internal/dto"

func ConvertDtoToV1(d dto.DTOv0) FlagData {
	var r Rollout

	if d.Rollout != nil {
		r = Rollout{}

		if d.Rollout.Progressive != nil {
			r.Progressive = &Progressive{}
			r.Progressive.ReleaseRamp.Start = d.Rollout.Progressive.ReleaseRamp.Start
			r.Progressive.ReleaseRamp.End = d.Rollout.Progressive.ReleaseRamp.End
			r.Progressive.Percentage.End = d.Rollout.Progressive.Percentage.End
			r.Progressive.Percentage.Initial = d.Rollout.Progressive.Percentage.Initial
		}

		if d.Rollout.Scheduled != nil {
			r.Scheduled = &ScheduledRollout{}
			if d.Rollout.Scheduled.Steps != nil {
				r.Scheduled.Steps = []ScheduledStep{}
				for _, step := range d.Rollout.Scheduled.Steps {
					s := ScheduledStep{
						FlagData: FlagData{
							Rule:        step.Rule,
							Percentage:  step.Percentage,
							True:        step.True,
							False:       step.False,
							Default:     step.Default,
							TrackEvents: step.TrackEvents,
							Disable:     step.Disable,
						},
						Date: step.Date,
					}
					r.Scheduled.Steps = append(r.Scheduled.Steps, s)
				}
			}
		}

		if d.Rollout.Experimentation != nil {
			r.Experimentation = &Experimentation{}
			if d.Rollout.Experimentation.Start != nil {
				r.Experimentation.Start = d.Rollout.Experimentation.Start
			}
			if d.Rollout.Experimentation.End != nil {
				r.Experimentation.End = d.Rollout.Experimentation.End
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
