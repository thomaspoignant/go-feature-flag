package flag

import (
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/thomaspoignant/go-feature-flag/modules/core/ffcontext"
)

// FlagDependency binds a flag to the state of another flag (the parent).
type FlagDependency struct {
	// Flag is the name of the parent flag.
	Flag string `json:"flag" yaml:"flag" toml:"flag"`

	// Variation is the variation the parent flag must resolve to for the dependency to be satisfied.
	// Exactly one of Variation and NotVariation must be set.
	Variation *string `json:"variation,omitempty" yaml:"variation,omitempty" toml:"variation,omitempty"`

	// NotVariation is a variation the parent flag must not resolve to for the dependency to be satisfied.
	NotVariation *string `json:"notVariation,omitempty" yaml:"notVariation,omitempty" toml:"notVariation,omitempty"`

	// AppliesTo is the list of variations of the flag that are bound by this dependency.
	// When empty, the dependency applies to all the variations.
	AppliesTo []string `json:"appliesTo,omitempty" yaml:"appliesTo,omitempty" toml:"appliesTo,omitempty"`

	// ExpiresWithParent makes the flag disabled as soon as the parent flag is disabled,
	// either by its own disable field or by one of its due scheduled rollout steps.
	ExpiresWithParent bool `json:"expiresWithParent,omitempty" yaml:"expiresWithParent,omitempty" toml:"expiresWithParent,omitempty"` // nolint: lll
}

const circularDependencyMessage = "circular dependency between flags"

// dependencyMemo is not safe for concurrent use, it belongs to one evaluation or one watcher pass.
type dependencyMemo struct {
	parents         map[string]ResolutionDetails
	parentsDisabled map[string]bool
}

func (d FlagDependency) appliesTo(variation string) bool {
	return len(d.AppliesTo) == 0 || slices.Contains(d.AppliesTo, variation)
}

func (d FlagDependency) isSatisfiedBy(parentVariation string) bool {
	switch {
	case d.Variation != nil:
		return parentVariation == *d.Variation
	case d.NotVariation != nil:
		return parentVariation != *d.NotVariation
	}
	return false
}

func (d FlagDependency) validate(variations map[string]*any) error {
	if d.Flag == "" {
		return fmt.Errorf("dependsOn: missing flag name")
	}
	if (d.Variation == nil) == (d.NotVariation == nil) {
		return fmt.Errorf("dependsOn %q: exactly one of variation and notVariation is required", d.Flag)
	}
	for _, name := range d.AppliesTo {
		if _, ok := variations[name]; !ok {
			return fmt.Errorf("dependsOn %q: appliesTo references unknown variation %q", d.Flag, name)
		}
	}
	return nil
}

// withDependencyChain returns a copy of the context that remembers the flags being resolved,
// it is used to stop on circular dependencies.
func (c Context) withDependencyChain(flagName string) Context {
	c.dependencyChain = append(slices.Clone(c.dependencyChain), flagName)
	return c
}

// unmetDependency checks the dependencies bound to the selected variation.
// It returns the error code and message of the first unmet dependency, or an empty code if all are met.
func (f *InternalFlag) unmetDependency(
	flagName, variation string,
	evaluationCtx ffcontext.Context,
	flagContext Context,
) (ErrorCode, string) {
	parentCtx := flagContext.WithMemo().withDependencyChain(flagName)
	for _, dep := range f.GetDependsOn() {
		if !dep.appliesTo(variation) {
			continue
		}
		if slices.Contains(parentCtx.dependencyChain, dep.Flag) {
			return ErrorFlagConfiguration,
				fmt.Sprintf("%s %q and %q", circularDependencyMessage, flagName, dep.Flag)
		}
		var parent Flag
		found := false
		if flagContext.GetFlag != nil {
			parent, found = flagContext.GetFlag(dep.Flag)
		}
		if !found {
			return ErrorCodeFlagDependency,
				fmt.Sprintf("flag %q depends on flag %q which does not exist", flagName, dep.Flag)
		}
		details, resolved := parentCtx.memo.parents[dep.Flag]
		if !resolved {
			_, details = parent.Value(dep.Flag, evaluationCtx, parentCtx)
			if parentCtx.memo.parents == nil {
				parentCtx.memo.parents = map[string]ResolutionDetails{}
			}
			parentCtx.memo.parents[dep.Flag] = details
		}
		if details.ErrorCode == ErrorFlagConfiguration &&
			strings.HasPrefix(details.ErrorMessage, circularDependencyMessage) {
			return details.ErrorCode, details.ErrorMessage
		}
		if details.Reason == ReasonError || details.Reason == ReasonDisabled ||
			!dep.isSatisfiedBy(details.Variant) {
			return ErrorCodeFlagDependency,
				fmt.Sprintf("variation %q of flag %q is not available: dependency on flag %q is not met",
					variation, flagName, dep.Flag)
		}
	}
	return "", ""
}

// isDisabledByParents returns true when a parent the flag expires with is disabled at the
// evaluation date, either by its own disable field or by one of its due scheduled rollout steps.
// The results of the parents reached without skipping a dependency (no cycle) are memoized,
// the ones computed on an incomplete walk (cycle guard) are not, they could be wrong for a
// later evaluation.
func (f *InternalFlag) isDisabledByParents(
	flagName string,
	evaluationDate time.Time,
	flagContext Context,
) bool {
	if flagContext.GetFlag == nil {
		return false
	}
	disabled, _ := f.isDisabledByParentsWalk(
		flagContext.WithMemo().withDependencyChain(flagName), evaluationDate)
	return disabled
}

// isDisabledByParentsWalk returns the disablement of the flag and whether the result is complete,
// a result computed after skipping a dependency of the current path (cycle guard) is not complete.
func (f *InternalFlag) isDisabledByParentsWalk(
	flagContext Context,
	evaluationDate time.Time,
) (bool, bool) {
	complete := true
	for _, dep := range f.GetDependsOn() {
		if !dep.ExpiresWithParent || slices.Contains(flagContext.dependencyChain, dep.Flag) {
			continue
		}
		parent, found := flagContext.GetFlag(dep.Flag)
		if !found {
			continue
		}
		if disabled, memoized := flagContext.memo.parentsDisabled[dep.Flag]; memoized {
			if disabled {
				return true, true
			}
			continue
		}
		parentDisabled, parentComplete := isParentDisabled(parent, dep.Flag, flagContext, evaluationDate)
		if !parentComplete {
			complete = false
		}
		if parentDisabled {
			if parentComplete {
				if flagContext.memo.parentsDisabled == nil {
					flagContext.memo.parentsDisabled = map[string]bool{}
				}
				flagContext.memo.parentsDisabled[dep.Flag] = true
			}
			return true, complete
		}
		if parentComplete {
			if flagContext.memo.parentsDisabled == nil {
				flagContext.memo.parentsDisabled = map[string]bool{}
			}
			flagContext.memo.parentsDisabled[dep.Flag] = false
		}
	}
	return false, complete
}

// isParentDisabled returns the disablement of the parent at the evaluation date and whether
// the result is complete, see isDisabledByParentsWalk.
func isParentDisabled(parent Flag, flagName string, flagContext Context, evaluationDate time.Time) (bool, bool) {
	p, ok := parent.(*InternalFlag)
	if !ok {
		return false, true
	}
	effective, err := p.applyScheduledRolloutSteps(evaluationDate)
	if err != nil {
		return false, true
	}
	if effective.IsDisable() {
		return true, true
	}
	return p.isDisabledByParentsWalk(
		flagContext.WithMemo().withDependencyChain(flagName), evaluationDate)
}
