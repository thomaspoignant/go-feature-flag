package flag

import (
	"fmt"
	"github.com/thomaspoignant/go-feature-flag/ffcontext"
	"sort"
	"strings"
	"time"

	"github.com/nikunjy/rules/parser"
	"github.com/thomaspoignant/go-feature-flag/internal/internalerror"
	"github.com/thomaspoignant/go-feature-flag/internal/utils"
)

// Rule represents a rule applied by the flag.
type Rule struct {
	// Name is the name of the rule, this field is mandatory if you want
	// to update the rule during scheduled rollout
	Name *string `json:"name,omitempty" yaml:"name,omitempty" toml:"name,omitempty" jsonschema:"title=name,description=Name is the name of the rule. This field is mandatory if you want to update the rule during scheduled rollout."` // nolint: lll

	// Query represents an antlr query in the nikunjy/rules format
	Query *string `json:"query,omitempty" yaml:"query,omitempty" toml:"query,omitempty" jsonschema:"title=query,description=The query that allow to check in the evaluation context match. Note: in the defaultRule field query is ignored."` // nolint: lll

	// VariationResult represents the variation name to use if the rule apply for the user.
	// In case we have a percentage field in the config VariationResult is ignored
	VariationResult *string `json:"variation,omitempty" yaml:"variation,omitempty" toml:"variation,omitempty" jsonschema:"title=variation,description=The variation name to use if the rule apply for the user. In case we have a percentage field in the config this field is ignored"` // nolint: lll

	// Percentages represents the percentage we should give to each variation.
	// example: variationA = 10%, variationB = 80%, variationC = 10%
	Percentages *map[string]float64 `json:"percentage,omitempty" yaml:"percentage,omitempty" toml:"percentage,omitempty" jsonschema:"title=percentage,description=Represents the percentage we should give to each variation."` // nolint: lll

	// ProgressiveRollout is your struct to configure a progressive rollout deployment of your flag.
	// It will allow you to ramp up the percentage of your flag over time.
	// You can decide at which percentage you starts with and at what percentage you ends with in your release ramp.
	// Before the start date we will serve the initial percentage and, after we will serve the end percentage.
	ProgressiveRollout *ProgressiveRollout `json:"progressiveRollout,omitempty" yaml:"progressiveRollout,omitempty" toml:"progressiveRollout,omitempty" jsonschema:"title=progressiveRollout,description=Configure a progressive rollout deployment of your flag."` // nolint: lll

	// Disable indicates that this rule is disabled.
	Disable *bool `json:"disable,omitempty" yaml:"disable,omitempty" toml:"disable,omitempty" jsonschema:"title=disable,description=Indicates that this rule is disabled."` // nolint: lll
}

// Evaluate is checking if the rule apply to for the user.
// If yes it returns the variation you should use for this rule.
func (r *Rule) Evaluate(ctx ffcontext.Context, flagName string, isDefault bool,
) (string, error) {
	evaluationDate := DateFromContextOrDefault(ctx, time.Now())
	// check that we have an evaluation context
	if ctx == nil {
		return "", fmt.Errorf("evaluate Rule: no evaluation context")
	}

	// Check if the rule apply for this user
	ruleApply := isDefault || r.GetQuery() == "" || parser.Evaluate(r.GetTrimmedQuery(), utils.ContextToMap(ctx))
	if !ruleApply || (!isDefault && r.IsDisable()) {
		return "", &internalerror.RuleNotApply{Context: ctx}
	}

	if r.ProgressiveRollout != nil {
		progressiveRolloutMaxPercentage := uint32(100 * PercentageMultiplier)
		hashID := utils.BuildHash(flagName, ctx.GetKey(), progressiveRolloutMaxPercentage)
		variation, err := r.getVariationFromProgressiveRollout(hashID, evaluationDate)
		if err != nil {
			return variation, err
		}
		return variation, nil
	}

	if r.Percentages != nil && len(r.GetPercentages()) > 0 {
		m := 0.0
		for _, percentage := range r.GetPercentages() {
			m += percentage
		}
		maxPercentage := uint32(m * PercentageMultiplier)
		hashID := utils.BuildHash(flagName, ctx.GetKey(), maxPercentage)
		variationName, err := r.getVariationFromPercentage(hashID)
		if err != nil {
			return "", err
		}
		return variationName, nil
	}

	if r.VariationResult != nil {
		return r.GetVariationResult(), nil
	}
	return "", fmt.Errorf("error in the configuration, no variation available for this rule")
}

// IsDynamic is a function that allows to know if the rule has a dynamic result or not.
func (r *Rule) IsDynamic() bool {
	hasPercentage100 := false
	for _, percentage := range r.GetPercentages() {
		if percentage == 100 {
			hasPercentage100 = true
			break
		}
	}
	return r.ProgressiveRollout != nil || (r.Percentages != nil && len(r.GetPercentages()) > 0 && !hasPercentage100)
}

func (r *Rule) getVariationFromProgressiveRollout(hash uint32, evaluationDate time.Time) (string, error) {
	isRolloutValid := r.ProgressiveRollout != nil &&
		r.ProgressiveRollout.Initial != nil &&
		r.ProgressiveRollout.Initial.Date != nil &&
		r.ProgressiveRollout.Initial.Variation != nil &&
		r.ProgressiveRollout.End != nil &&
		r.ProgressiveRollout.End.Date != nil &&
		r.ProgressiveRollout.End.Variation != nil &&
		r.ProgressiveRollout.End.Date.After(*r.ProgressiveRollout.Initial.Date)

	if isRolloutValid {
		if evaluationDate.Before(*r.ProgressiveRollout.Initial.Date) {
			return *r.ProgressiveRollout.Initial.Variation, nil
		}

		// We are between initial and end
		initialPercentage := r.ProgressiveRollout.Initial.getPercentage() * PercentageMultiplier
		if r.ProgressiveRollout.End.getPercentage() == 0 || r.ProgressiveRollout.End.getPercentage() > 100 {
			maxPercentage := float64(100)
			r.ProgressiveRollout.End.Percentage = &maxPercentage
		}
		endPercentage := r.ProgressiveRollout.End.getPercentage() * PercentageMultiplier
		nbSec := r.ProgressiveRollout.End.Date.Unix() - r.ProgressiveRollout.Initial.Date.Unix()
		percentage := endPercentage - initialPercentage
		percentPerSec := percentage / float64(nbSec)

		c := evaluationDate.Unix() - r.ProgressiveRollout.Initial.Date.Unix()
		currentPercentage := float64(c)*percentPerSec + initialPercentage

		if hash < uint32(currentPercentage) {
			return r.ProgressiveRollout.End.getVariation(), nil
		}
		return r.ProgressiveRollout.Initial.getVariation(), nil
	}
	return "", fmt.Errorf("error in the progressive rollout, missing params")
}

func (r *Rule) getVariationFromPercentage(hash uint32) (string, error) {
	for key, bucket := range r.getPercentageBuckets() {
		if uint32(bucket.start) <= hash && uint32(bucket.end) > hash {
			return key, nil
		}
	}
	return "", fmt.Errorf("impossible to find the variation")
}

// getPercentageBuckets compute a map containing the buckets of each variation for this rule.
func (r *Rule) getPercentageBuckets() map[string]percentageBucket {
	percentageBuckets := make(map[string]percentageBucket, len(r.GetPercentages()))
	percentage := r.GetPercentages()

	// we need to sort the map to affect the bucket to be sure we are constantly affecting the users to the same bucket.
	// Map are not ordered in GO, so we have to order the variationNames to be able to compute the same numbers for the
	// buckets every time we are in this function.
	variationNames := make([]string, 0)
	for k := range percentage {
		variationNames = append(variationNames, k)
	}
	// HACK: we are reversing the sort to support the legacy format of the flags (before 1.0.0) and to be sure to always
	// have "True" before "False"
	sort.Sort(sort.Reverse(sort.StringSlice(variationNames)))

	for index, varName := range variationNames {
		startBucket := float64(0)
		if index != 0 {
			startBucket = percentageBuckets[variationNames[index-1]].end
		}
		endBucket := startBucket + (percentage[varName] * PercentageMultiplier)
		percentageBuckets[varName] = percentageBucket{
			start: startBucket,
			end:   endBucket,
		}
	}
	return percentageBuckets
}

// MergeRules is merging 2 rules.
// It is used when we have to update a rule in a scheduled rollout.
func (r *Rule) MergeRules(updatedRule Rule) {
	if updatedRule.Query != nil {
		r.Query = updatedRule.Query
	}

	if updatedRule.VariationResult != nil {
		r.VariationResult = updatedRule.VariationResult
	}

	if updatedRule.ProgressiveRollout != nil {
		c := r.GetProgressiveRollout()
		if updatedRule.ProgressiveRollout.Initial != nil {
			c.Initial.mergeStep(updatedRule.ProgressiveRollout.Initial)
		}

		if updatedRule.ProgressiveRollout.End != nil {
			c.End.mergeStep(updatedRule.ProgressiveRollout.End)
		}
		r.ProgressiveRollout = &c
	}

	if updatedRule.Percentages != nil {
		updatedPercentages := updatedRule.GetPercentages()
		mergedPercentages := r.GetPercentages()
		for key, percentage := range updatedPercentages {
			// When you set a negative percentage we are not taking it in consideration.
			if percentage < 0 {
				delete(mergedPercentages, key)
				continue
			}
			mergedPercentages[key] = percentage
		}
		r.Percentages = &mergedPercentages
	}
}

// IsValid is checking if the rule is valid
func (r *Rule) IsValid(defaultRule bool) error {
	if !defaultRule && r.IsDisable() {
		return nil
	}

	if r.Percentages == nil && r.ProgressiveRollout == nil && r.VariationResult == nil {
		return fmt.Errorf("impossible to return value")
	}

	// targeting without query
	if !defaultRule && r.Query == nil {
		return fmt.Errorf("each targeting should have a query")
	}

	// Validate the percentage of the rule
	if r.Percentages != nil {
		count := float64(0)
		for _, p := range r.GetPercentages() {
			count += p
		}

		if len(r.GetPercentages()) == 0 {
			return fmt.Errorf("invalid percentages: should not be empty")
		}

		if count == 0 {
			return fmt.Errorf("invalid percentages: should not be equal to 0")
		}
	}

	// Progressive rollout: check that initial is lower than end
	if r.ProgressiveRollout != nil &&
		(r.GetProgressiveRollout().End.getPercentage() < r.GetProgressiveRollout().Initial.getPercentage()) {
		return fmt.Errorf("invalid progressive rollout, initial percentage should be lower "+
			"than end percentage: %v/%v",
			r.GetProgressiveRollout().Initial.getPercentage(), r.GetProgressiveRollout().End.getPercentage())
	}
	return nil
}

// GetTrimmedQuery is removing the break lines and return
func (r *Rule) GetTrimmedQuery() string {
	splitQuery := strings.Split(r.GetQuery(), "\n")
	for index, item := range splitQuery {
		splitQuery[index] = strings.TrimLeft(item, " ")
	}
	return strings.Join(splitQuery, "")
}

func (r *Rule) GetQuery() string {
	if r.Query == nil {
		return ""
	}

	return *r.Query
}

func (r *Rule) GetVariationResult() string {
	if r.VariationResult == nil {
		return ""
	}
	return *r.VariationResult
}

func (r *Rule) GetName() string {
	if r.Name == nil {
		return ""
	}
	return *r.Name
}

func (r *Rule) GetPercentages() map[string]float64 {
	if r.Percentages == nil {
		return map[string]float64{}
	}
	return *r.Percentages
}

func (r *Rule) IsDisable() bool {
	if r.Disable == nil {
		return false
	}
	return *r.Disable
}

func (r *Rule) GetProgressiveRollout() ProgressiveRollout {
	if r.ProgressiveRollout == nil {
		return ProgressiveRollout{
			Initial: &ProgressiveRolloutStep{},
			End:     &ProgressiveRolloutStep{},
		}
	}
	return *r.ProgressiveRollout
}

// MergeSetOfRules is taking a collection of rules and merge it with the updates
// from a schedule steps.
// If you want to edit a rule this rule should have a name already to be able to
// target the updates to the right place.
func MergeSetOfRules(initialRules []Rule, updates []Rule) *[]Rule {
	collection := initialRules
	modified := make(map[string]Rule, 0)
	for _, update := range updates {
		ruleName := update.Name
		if ruleName != nil {
			modified[update.GetName()] = update
		}
	}

	mergedUpdates := make([]string, 0)
	for index, rule := range collection {
		if _, ok := modified[rule.GetName()]; ok {
			rule.MergeRules(modified[rule.GetName()])
			collection[index] = rule
			mergedUpdates = append(mergedUpdates, rule.GetName())
		}
	}

	for _, update := range updates {
		if !utils.Contains(mergedUpdates, update.GetName()) {
			collection = append(collection, update)
		}
	}

	return &collection
}
