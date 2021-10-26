package flag

import (
	"fmt"
	"github.com/nikunjy/rules/parser"
	"github.com/thomaspoignant/go-feature-flag/ffuser"
	"sort"
	"strings"
	"time"
)

// Rule represents a rule applied by the flag.
type Rule struct {
	// Query represents an antlr query in the nikunjy/rules format
	Query *string `json:"query,omitempty" yaml:"query,omitempty" toml:"query,omitempty"`

	// VariationResult represents the variation name to use if the rule apply for the user.
	// In case we have a percentage field in the config VariationResult is ignored
	VariationResult *string `json:"variation,omitempty" yaml:"variation,omitempty" toml:"variation,omitempty"` // nolint: lll

	// Percentages represents the percentage we should give to each variations.
	// example: variationA = 10%, variationB = 80%, variationC = 10%
	Percentages *map[string]float64 `json:"percentage,omitempty" yaml:"percentage,omitempty" toml:"percentage,omitempty"` // nolint: lll

	ProgressiveRollout *ProgressiveRollout `json:"progressiveRollout,omitempty" yaml:"progressiveRollout,omitempty" toml:"progressiveRollout,omitempty"` // nolint: lll
}

func (r *Rule) evaluate(user ffuser.User, hash uint32, defaultRule bool,
) ( /* apply */ bool /* variation name */, string, error) {
	// Check if the rule apply for this user
	ruleApply := defaultRule || r.Query == nil || *r.Query == "" || parser.Evaluate(*r.Query, userToMap(user))
	if !ruleApply {
		return false, "", nil
	}

	if r.ProgressiveRollout != nil {
		variation, err := r.getVariationFromProgressiveRollout(hash)
		if err == nil {
			return true, variation, nil
		}
		// TODO add log to explain that we cannot use the rollout flag + continue
	}

	if r.Percentages != nil {
		variationName, err := r.getVariationFromPercentage(hash)
		if err != nil {
			return false, "", err
		}
		return true, variationName, nil
	}

	if r.VariationResult != nil {
		return true, *r.VariationResult, nil
	}
	return false, "", fmt.Errorf("error in the configuration, no variation available for this rule")
}

func (r *Rule) getVariationFromProgressiveRollout(hash uint32) (string, error) {
	isRolloutValid := r.ProgressiveRollout.ReleaseRamp.Start != nil &&
		r.ProgressiveRollout.ReleaseRamp.End != nil &&
		r.ProgressiveRollout.Variation.End != nil &&
		r.ProgressiveRollout.Variation.Initial != nil

	if isRolloutValid {
		now := time.Now()
		if now.Before(*r.ProgressiveRollout.ReleaseRamp.Start) {
			return *r.ProgressiveRollout.Variation.Initial, nil
		}

		if now.After(*r.ProgressiveRollout.ReleaseRamp.End) {
			return *r.ProgressiveRollout.Variation.End, nil
		}

		// We are between initial and end
		initialPercentage := r.ProgressiveRollout.Percentage.Initial * percentageMultiplier
		if r.ProgressiveRollout.Percentage.End == 0 {
			r.ProgressiveRollout.Percentage.End = 100
		}
		endPercentage := r.ProgressiveRollout.Percentage.End * percentageMultiplier

		nbSec := r.ProgressiveRollout.ReleaseRamp.End.Unix() - r.ProgressiveRollout.ReleaseRamp.Start.Unix()
		percentage := endPercentage - initialPercentage
		percentPerSec := percentage / float64(nbSec)

		c := now.Unix() - r.ProgressiveRollout.ReleaseRamp.Start.Unix()
		currentPercentage := float64(c)*percentPerSec + initialPercentage

		if hash < uint32(currentPercentage) {
			return *r.ProgressiveRollout.Variation.End, nil
		}
		return *r.ProgressiveRollout.Variation.Initial, nil
	}
	return "", fmt.Errorf("error in the progressive rollout, missing params")
}

func (r *Rule) getVariationFromPercentage(hash uint32) (string, error) {
	buckets := r.getPercentageBuckets()
	for key, bucket := range buckets {
		if uint32(bucket.start) <= hash && uint32(bucket.end) > hash {
			return key, nil
		}
	}
	return "", fmt.Errorf("impossible to find the variation")
}

// getPercentageBuckets compute a map containing the buckets of each variation for this rule.
func (r *Rule) getPercentageBuckets() map[string]rulePercentageBucket {
	percentageBuckets := map[string]rulePercentageBucket{}

	// we sort the variation to be sure to create the buckets all the time in the same order
	percentage := r.GetPercentages()

	// we need to sort the map to affect the bucket to be sure we are constantly affecting the users to the same bucket.
	// Map are not ordered in GO, so we have to order the variationNames to be able to compute the same numbers for the buckets
	// everytime we are in this function.
	variationNames := make([]string, 0, len(percentage))
	for k := range percentage {
		variationNames = append(variationNames, k)
	}
	// HACK: we are reversing the sort to support the legacy format of the flags (before 1.0.0) and to be sure to always
	// have "True" before "False"
	sort.Sort(sort.Reverse(sort.StringSlice(variationNames)))

	bucketStart := float64(0)
	for _, varName := range variationNames {
		bucketLimit := percentage[varName] * percentageMultiplier
		bucketEnd := bucketLimit
		percentageBuckets[varName] = rulePercentageBucket{
			start: bucketStart,
			end:   bucketEnd,
		}
		bucketStart = bucketLimit
	}
	return percentageBuckets
}

func (r *Rule) GetQuery() string {
	if r.Query == nil {
		return ""
	}
	return *r.Query
}

func (r *Rule) GetVariation() string {
	if r.VariationResult == nil {
		return ""
	}
	return *r.VariationResult
}

func (r *Rule) GetPercentages() map[string]float64 {
	if r.Percentages == nil {
		return map[string]float64{}
	}
	return *r.Percentages
}

//func (r *Rule) GetRollout() *RuleRollout {
//	return r.Rollout
//}

func (r Rule) String() string {
	var toString []string
	toString = appendIfHasValue(toString, "query", fmt.Sprintf("%v", r.GetQuery()))
	toString = appendIfHasValue(toString, "variation", fmt.Sprintf("%v", r.GetVariation()))

	var percentString []string
	//TODO : fix me
	//for _, p := range r.GetPercentages() {
	//percentString = append(percentString, p.String())
	//}
	toString = appendIfHasValue(toString, "percentages", strings.Join(percentString, ","))
	//if r.GetRollout() != nil {
	//	toString = appendIfHasValue(toString, "rollout", fmt.Sprintf("%v", *r.GetRollout()))
	//}
	return fmt.Sprintf("%s", strings.Join(toString, ","))
}
