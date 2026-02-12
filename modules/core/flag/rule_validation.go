package flag

import (
	"fmt"
	"strings"

	"github.com/diegoholiveira/jsonlogic/v3"
	"github.com/nikunjy/rules/parser"
)

// IsValid is checking if the rule is valid
func (r *Rule) IsValid(defaultRule bool, variations map[string]*any) error {
	if !defaultRule && r.IsDisable() {
		return nil
	}

	if r.Percentages == nil && r.ProgressiveRollout == nil && r.VariationResult == nil {
		return fmt.Errorf("impossible to return value")
	}

	if err := r.isQueryValid(defaultRule); err != nil {
		return err
	}

	if err := r.validatePercentages(variations); err != nil {
		return err
	}

	if err := r.validateProgressiveRollout(variations); err != nil {
		return err
	}

	if err := r.validateVariationResult(variations); err != nil {
		return err
	}

	return nil
}

// validatePercentages validates the percentage configuration of the rule.
// It checks that percentages are not empty, the sum is not zero, and all
// referenced variations exist in the provided variations map.
func (r *Rule) validatePercentages(variations map[string]*any) error {
	if r.Percentages == nil {
		return nil
	}

	percentages := r.GetPercentages()
	if len(percentages) == 0 {
		return fmt.Errorf("invalid percentages: should not be empty")
	}

	count := float64(0)
	for k, p := range percentages {
		count += p
		if _, ok := variations[k]; !ok {
			return fmt.Errorf("invalid percentage: variation %s does not exist", k)
		}
	}

	if count == 0 {
		return fmt.Errorf("invalid percentages: should not be equal to 0")
	}

	return nil
}

// validateProgressiveRollout validates the progressive rollout configuration of the rule.
// It checks that the initial percentage is lower than the end percentage, and that
// both the initial and end variations exist in the provided variations map.
func (r *Rule) validateProgressiveRollout(variations map[string]*any) error {
	if r.ProgressiveRollout == nil {
		return nil
	}

	progRollout := r.GetProgressiveRollout()
	initialPct := progRollout.Initial.getPercentage()
	endPct := progRollout.End.getPercentage()

	if endPct < initialPct {
		return fmt.Errorf(
			"invalid progressive rollout, initial percentage should be lower "+
				"than end percentage: %v/%v",
			initialPct,
			endPct,
		)
	}

	endVar := progRollout.End.getVariation()
	if _, ok := variations[endVar]; !ok {
		return fmt.Errorf(
			"invalid progressive rollout, end variation %s does not exist",
			endVar,
		)
	}

	initialVar := progRollout.Initial.getVariation()
	if _, ok := variations[initialVar]; !ok {
		return fmt.Errorf(
			"invalid progressive rollout, initial variation %s does not exist",
			initialVar,
		)
	}

	return nil
}

// validateVariationResult validates the simple variation result configuration of the rule.
// It checks that the variation result exists in the provided variations map.
// This validation only applies when the rule uses a simple variation result
// (not percentages or progressive rollout).
func (r *Rule) validateVariationResult(variations map[string]*any) error {
	if r.Percentages != nil || r.ProgressiveRollout != nil || r.VariationResult == nil {
		return nil
	}

	if _, ok := variations[r.GetVariationResult()]; !ok {
		return fmt.Errorf("invalid variation: %s does not exist", r.GetVariationResult())
	}

	return nil
}

// isQueryValid validates the query configuration of the rule.
// It checks that the query is not empty and that the query format is valid.
// The query format can be either JSONLogic or Nikunjy.
func (r *Rule) isQueryValid(defaultRule bool) error {
	if defaultRule {
		return nil
	}

	if r.Query == nil {
		return fmt.Errorf("each targeting should have a query")
	}

	// Validate the query with the parser
	switch r.GetQueryFormat() {
	case JSONLogicQueryFormat:
		if !jsonlogic.IsValid(strings.NewReader(r.GetQuery())) {
			return fmt.Errorf("invalid jsonlogic query")
		}
		return nil
	default:
		return validateNikunjyQuery(r.GetTrimmedQuery())
	}
}

// validateNikunjyQuery validates the Nikunjy query configuration of the rule.
func validateNikunjyQuery(query string) error {
	ev, err := parser.NewEvaluator(query)
	if err != nil {
		return err
	}
	_, err = ev.Process(map[string]any{})
	if err != nil {
		return fmt.Errorf("invalid query: %w", err)
	}
	return nil
}
