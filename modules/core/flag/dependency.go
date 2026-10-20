package flag

import (
	"fmt"
	"slices"
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

	// ExpiresWithParent makes the flag expire when the parent flag expires.
	ExpiresWithParent bool `json:"expiresWithParent,omitempty" yaml:"expiresWithParent,omitempty" toml:"expiresWithParent,omitempty"` // nolint: lll
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
	parentCtx := flagContext.withDependencyChain(flagName)
	for _, dep := range f.GetDependsOn() {
		if !dep.appliesTo(variation) {
			continue
		}
		if slices.Contains(parentCtx.dependencyChain, dep.Flag) {
			return ErrorFlagConfiguration,
				fmt.Sprintf("circular dependency between flags %q and %q", flagName, dep.Flag)
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
		_, details := parent.Value(dep.Flag, evaluationCtx, parentCtx)
		if details.Reason == ReasonError || details.Reason == ReasonDisabled ||
			!dep.isSatisfiedBy(details.Variant) {
			return ErrorCodeFlagDependency,
				fmt.Sprintf("variation %q of flag %q is not available: dependency on flag %q is not met",
					variation, flagName, dep.Flag)
		}
	}
	return "", ""
}

// effectiveExpiration is the expiration date of the flag, taking into account the parents
// the flag expires with. The earliest date wins.
func (f *InternalFlag) effectiveExpiration(flagContext Context) (time.Time, bool) {
	expiration, ok := f.ExpirationDate(flagContext.DefaultTTL)
	if flagContext.GetFlag == nil {
		return expiration, ok
	}
	for _, dep := range f.GetDependsOn() {
		if !dep.ExpiresWithParent || slices.Contains(flagContext.dependencyChain, dep.Flag) {
			continue
		}
		parent, found := flagContext.GetFlag(dep.Flag)
		if !found {
			continue
		}
		p, isExpirable := parent.(interface {
			effectiveExpiration(Context) (time.Time, bool)
		})
		if !isExpirable {
			continue
		}
		parentExpiration, parentOk := p.effectiveExpiration(flagContext.withDependencyChain(dep.Flag))
		if parentOk && (!ok || parentExpiration.Before(expiration)) {
			expiration, ok = parentExpiration, true
		}
	}
	return expiration, ok
}
