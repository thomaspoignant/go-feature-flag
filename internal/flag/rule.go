package flag

import (
	"errors"
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

	// ProgressiveRollout is your struct to configure a progressive rollout deployment of your flag.
	// It will allow you to ramp up the percentage of your flag over time.
	// You can decide at which percentage you starts and at what percentage you ends in your release ramp.
	// Before the start date we will serve the initial percentage and after we will serve the end percentage.
	ProgressiveRollout *ProgressiveRollout `json:"progressiveRollout,omitempty" yaml:"progressiveRollout,omitempty" toml:"progressiveRollout,omitempty"` // nolint: lll
}

func (r *Rule) mergeChanges(updatedRule Rule) {
	if updatedRule.Query != nil {
		r.Query = updatedRule.Query
	}

	if updatedRule.VariationResult != nil {
		r.VariationResult = updatedRule.VariationResult
	}

	if updatedRule.Percentages != nil {
		updatedPercentages := *updatedRule.Percentages
		mergedPercentages := r.GetPercentages()
		for key, percentage := range updatedPercentages {
			// TODO add documentation about the -1
			if percentage == -1 {
				delete(mergedPercentages, key)
				continue
			}
			mergedPercentages[key] = percentage
		}
		r.Percentages = &mergedPercentages
	}
}

func (r *Rule) Evaluate(user ffuser.User, hashID uint32, defaultRule bool,
) ( /* apply */ bool /* variation name */, string, error) {
	// Check if the rule apply for this user
	ruleApply := defaultRule || r.Query == nil || *r.Query == "" || parser.Evaluate(*r.Query, userToMap(user))
	if !ruleApply {
		return false, "", nil
	}

	if r.ProgressiveRollout != nil {
		variation, err := r.getVariationFromProgressiveRollout(hashID)
		if err == nil {
			return true, variation, nil
		}
		// TODO add log to explain that we cannot use the rollout flag + continue
	}

	if r.Percentages != nil {
		variationName, err := r.getVariationFromPercentage(hashID)
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
	isRolloutValid :=
		r.ProgressiveRollout.Initial.Date != nil &&
			r.ProgressiveRollout.End.Date != nil &&
			r.ProgressiveRollout.Initial.Variation != nil &&
			r.ProgressiveRollout.End.Variation != nil

	if isRolloutValid {
		now := time.Now()
		if now.Before(*r.ProgressiveRollout.Initial.Date) {
			return *r.ProgressiveRollout.Initial.Variation, nil
		}

		if now.After(*r.ProgressiveRollout.End.Date) {
			return *r.ProgressiveRollout.End.Variation, nil
		}

		// We are between initial and end
		initialPercentage := r.ProgressiveRollout.Initial.Percentage * PercentageMultiplier
		if r.ProgressiveRollout.End.Percentage == 0 || r.ProgressiveRollout.End.Percentage > 100 {
			r.ProgressiveRollout.End.Percentage = 100
		}
		endPercentage := r.ProgressiveRollout.End.Percentage * PercentageMultiplier

		nbSec := r.ProgressiveRollout.End.Date.Unix() - r.ProgressiveRollout.Initial.Date.Unix()
		percentage := endPercentage - initialPercentage
		percentPerSec := percentage / float64(nbSec)

		c := now.Unix() - r.ProgressiveRollout.Initial.Date.Unix()
		currentPercentage := float64(c)*percentPerSec + initialPercentage

		if hash < uint32(currentPercentage) {
			return *r.ProgressiveRollout.End.Variation, nil
		}
		return *r.ProgressiveRollout.Initial.Variation, nil
	}
	return "", fmt.Errorf("error in the progressive rollout, missing params")
}

func (r *Rule) getVariationFromPercentage(hash uint32) (string, error) {
	buckets, err := r.getPercentageBuckets()
	if err != nil {
		return "", err
	}

	for key, bucket := range buckets {
		if uint32(bucket.start) <= hash && uint32(bucket.end) > hash {
			return key, nil
		}
	}
	return "", fmt.Errorf("impossible to find the variation")
}

// getPercentageBuckets compute a map containing the buckets of each variation for this rule.
func (r *Rule) getPercentageBuckets() (map[string]rulePercentageBucket, error) {
	percentageBuckets := map[string]rulePercentageBucket{}
	totalPercentage := float64(0)
	percentage := r.GetPercentages()

	// we need to sort the map to affect the bucket to be sure we are constantly affecting the users to the same bucket.
	// Map are not ordered in GO, so we have to order the variationNames to be able to compute the same numbers for the
	// buckets everytime we are in this function.
	variationNames := make([]string, 0)
	for k := range percentage {
		variationNames = append(variationNames, k)
	}
	// HACK: we are reversing the sort to support the legacy format of the flags (before 1.0.0) and to be sure to always
	// have "True" before "False"
	sort.Sort(sort.Reverse(sort.StringSlice(variationNames)))

	bucketStart := float64(0)
	for _, varName := range variationNames {
		totalPercentage += percentage[varName]
		bucketLimit := percentage[varName] * PercentageMultiplier
		bucketEnd := bucketStart + bucketLimit
		percentageBuckets[varName] = rulePercentageBucket{
			start: bucketStart,
			end:   bucketEnd,
		}
		bucketStart = bucketLimit
	}

	if totalPercentage != float64(100) {
		return nil, errors.New("invalid rule because percentage are not representing 100%")
	}

	return percentageBuckets, nil
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

func (r Rule) String() string {
	var toString []string
	toString = appendIfHasValue(toString, "query", fmt.Sprintf("%v", r.GetQuery()))
	toString = appendIfHasValue(toString, "variation", fmt.Sprintf("%v", r.GetVariation()))

	var percentString = make([]string, 0)
	for key, val := range r.GetPercentages() {
		percentString = append(percentString, fmt.Sprintf("%s=%.2f", key, val))
	}
	sort.Strings(percentString)

	if len(percentString) == 0 {
		percentString = nil
	}
	toString = appendIfHasValue(toString, "percentages", strings.Join(percentString, ","))
	if r.ProgressiveRollout != nil {
		toString = appendIfHasValue(toString, "progressiveRollout", fmt.Sprintf("%v", r.ProgressiveRollout))
	}
	return strings.Join(toString, ", ")
}
