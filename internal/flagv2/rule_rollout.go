package flagv2

//
//import (
//	"fmt"
//	"github.com/thomaspoignant/go-feature-flag/internal/flagv1"
//)
//
//type RuleRollout struct {
//	// Progressive is your struct to configure a progressive rollout deployment of your flag.
//	// It will allow you to ramp up the percentage of your flag over time.
//	// You can decide at which percentage you starts and at what percentage you ends in your release ramp.
//	// Before the start date we will serve the initial percentage and after we will serve the end percentage.
//	Progressive *flagv1.Progressive `json:"progressive,omitempty" yaml:"progressive,omitempty" toml:"progressive,omitempty"` // nolint: lll
//}
//
//func (e RuleRollout) String() string {
//	return fmt.Sprintf("progressive:%v",e.Progressive)
//}
