package flag

import (
	"fmt"
	"strings"
	"time"
)

type ProgressiveRollout struct {
	// TODO: comments
	Initial *ProgressiveRolloutStep `json:"initial,omitempty" yaml:"initial,omitempty" toml:"initial,omitempty"`
	End     *ProgressiveRolloutStep `json:"end,omitempty" yaml:"end,omitempty" toml:"end,omitempty"`
}

type ProgressiveRolloutStep struct {
	Variation  *string
	Percentage float64
	Date       *time.Time
}

func (p *ProgressiveRolloutStep) getVariation() string {
	if p.Variation == nil {
		return ""
	}
	return *p.Variation
}

func (p *ProgressiveRolloutStep) getPercentage() float64 {
	return p.Percentage
}

func (p ProgressiveRollout) String() string {
	var initial []string
	initial = appendIfHasValue(initial, "Variation", fmt.Sprintf("%v", p.Initial.getVariation()))
	initial = appendIfHasValue(initial, "Percentage", fmt.Sprintf("%v", p.Initial.getPercentage()))
	if p.Initial.Date != nil {
		initialDate := *p.Initial.Date
		initial = appendIfHasValue(initial, "Date", fmt.Sprintf("%v", initialDate.Format(time.RFC3339)))
	}

	var end []string
	end = appendIfHasValue(end, "Variation", fmt.Sprintf("%v", p.End.getVariation()))
	end = appendIfHasValue(end, "Percentage", fmt.Sprintf("%v", p.End.getPercentage()))
	if p.End.Date != nil {
		endDate := *p.End.Date
		end = appendIfHasValue(end, "Date", fmt.Sprintf("%v", endDate.Format(time.RFC3339)))
	}

	return fmt.Sprintf("Initial:[%v], End:[%v]", strings.Join(initial, ", "), strings.Join(end, ", "))

}
