package flag

// ConvertV1DtoToFlag is converting a flag in the config file to the internal format.
func ConvertV1DtoToFlag(d DtoFlag, isScheduleStep bool) FlagData {
	return FlagData{
		Variations:  d.Variations,
		Rules:       d.Rules,
		DefaultRule: d.DefaultRule,
		Rollout:     d.Rollout.convertRollout(1),
		TrackEvents: d.TrackEvents,
		Disable:     d.Disable,
		Version:     d.Version,
	}
}
