package flag

import (
	"encoding/json"
	"fmt"
	"maps"
	"sort"
	"time"

	"github.com/thomaspoignant/go-feature-flag/modules/core/ffcontext"
	"github.com/thomaspoignant/go-feature-flag/modules/core/internalerror"
	"github.com/thomaspoignant/go-feature-flag/modules/core/utils"
)

const (
	PercentageMultiplier = float64(1000)
)

// InternalFlag is the internal representation of a flag when using go-feature-flag.
// All the flags in your configuration files can have different format but will be
// converted into an InternalFlag to be used in the library.
type InternalFlag struct {
	// Variations are all the variations available for this flag. You can have as many variation as needed.
	Variations *map[string]*any `json:"variations,omitempty" yaml:"variations,omitempty" toml:"variations,omitempty"` // nolint:lll

	// Rules is the list of Rule for this flag.
	// This an optional field.
	Rules *[]Rule `json:"targeting,omitempty" yaml:"targeting,omitempty" toml:"targeting,omitempty"`

	// BucketingKey defines a source for a dynamic targeting key
	BucketingKey *string `json:"bucketingKey,omitempty" yaml:"bucketingKey,omitempty" toml:"bucketingKey,omitempty"`

	// DefaultRule is the originalRule applied after checking that any other rules
	// matched the user.
	DefaultRule *Rule `json:"defaultRule,omitempty" yaml:"defaultRule,omitempty" toml:"defaultRule,omitempty"`

	// Experimentation is your struct to configure an experimentation, it will allow you to configure a start date and
	// an end date for your flag.
	// When the experimentation is not running, the flag will serve the default value.
	Experimentation *ExperimentationRollout `json:"experimentation,omitempty" yaml:"experimentation,omitempty" toml:"experimentation,omitempty"` // nolint: lll

	// Scheduled is your struct to configure an update on some fields of your flag over time.
	// You can add several steps that updates the flag, this is typically used if you want to gradually add more user
	// in your flag.
	Scheduled *[]ScheduledStep `json:"scheduledRollout,omitempty" yaml:"scheduledRollout,omitempty" toml:"scheduledRollout,omitempty"` // nolint: lll

	// TrackEvents is false if you don't want to export the data in your data exporter.
	// Default value is true
	TrackEvents *bool `json:"trackEvents,omitempty" yaml:"trackEvents,omitempty" toml:"trackEvents,omitempty"`

	// Disable is true if the flag is disabled.
	Disable *bool `json:"disable,omitempty" yaml:"disable,omitempty" toml:"disable,omitempty"`

	// Version (optional) This field contains the version of the flag.
	// The version is manually managed when you configure your flags, and it is used to display the information
	// in the notifications and data collection.
	Version *string `json:"version,omitempty" yaml:"version,omitempty" toml:"version,omitempty"`

	// Metadata is a field containing information about your flag such as an issue tracker link, a description, etc ...
	Metadata *map[string]any `json:"metadata,omitempty" yaml:"metadata,omitempty" toml:"metadata,omitempty"`

	// Needs is the list of dependencies this flag has on other flags.
	// The flag is only evaluated normally when all the dependencies are satisfied; if any of them
	// is unmet the flag is treated as disabled. Dependency resolution is limited to one level (the
	// dependency flag's own `needs` field is ignored).
	Needs *[]NeedsDependency `json:"needs,omitempty" yaml:"needs,omitempty" toml:"needs,omitempty"`
}

// Value is returning the Value associate to the flag
func (f *InternalFlag) Value(
	flagName string,
	evaluationCtx ffcontext.Context,
	flagContext Context,
) (any, ResolutionDetails) {
	return f.value(flagName, evaluationCtx, flagContext, true)
}

// value evaluates the flag. When resolveNeeds is true, the flag's `needs` dependencies are
// checked first and, if any of them is unmet, the flag is treated as disabled. Dependency flags
// are evaluated with resolveNeeds set to false so that their own `needs` field is ignored
// (one-level resolution, which also makes dependency cycles safe).
func (f *InternalFlag) value(
	flagName string,
	evaluationCtx ffcontext.Context,
	flagContext Context,
	resolveNeeds bool,
) (any, ResolutionDetails) {
	// if the evaluation context is nil, we create a new one with an empty key
	// this is to avoid any nil pointer exception.
	if evaluationCtx == nil {
		evaluationCtx = ffcontext.NewEvaluationContext("")
	}

	evaluationDate := DateFromContextOrDefault(evaluationCtx, time.Now())
	flag, err := f.applyScheduledRolloutSteps(evaluationDate)
	if err != nil {
		return flagContext.DefaultSdkValue, ResolutionDetails{
			Variant:      VariationSDKDefault,
			Reason:       ReasonError,
			ErrorCode:    ErrorCodeGeneral,
			ErrorMessage: err.Error(),
		}
	}

	if flagContext.EvaluationContextEnrichment != nil {
		maps.Copy(evaluationCtx.GetCustom(), flagContext.EvaluationContextEnrichment)
	}

	// A flag that declares dependencies through `needs` is only evaluated when all of them are
	// satisfied. If any dependency is unmet, the flag is disabled (same behavior as disable: true).
	// needsCacheable carries the cacheability of the dependencies so a result gated by a time-based
	// dependency is never cached.
	needsCacheable := true
	if resolveNeeds {
		unsatisfiedDependency, satisfied, dependenciesCacheable := flag.checkNeeds(evaluationCtx, flagContext)
		needsCacheable = dependenciesCacheable
		if !satisfied {
			return flagContext.DefaultSdkValue, ResolutionDetails{
				Variant:   VariationSDKDefault,
				Reason:    ReasonDisabled,
				Cacheable: f.isCacheable() && dependenciesCacheable,
				Metadata:  constructNeedsMetadata(f.GetMetadata(), unsatisfiedDependency),
			}
		}
	}

	key, keyError := flag.GetBucketingKeyValue(evaluationCtx)
	if keyError != nil {
		return flagContext.DefaultSdkValue, ResolutionDetails{
			Variant:      VariationSDKDefault,
			Reason:       ReasonError,
			ErrorCode:    ErrorCodeTargetingKeyMissing,
			ErrorMessage: keyError.Error(),
			Metadata:     f.GetMetadata(),
		}
	}

	if flag.IsDisable() || flag.isExperimentationOver(evaluationDate) {
		return flagContext.DefaultSdkValue, ResolutionDetails{
			Variant:   VariationSDKDefault,
			Reason:    ReasonDisabled,
			Cacheable: f.isCacheable(),
			Metadata:  f.GetMetadata(),
		}
	}

	variationSelection, err := flag.selectVariation(flagName, key, evaluationCtx)
	if err != nil {
		return flagContext.DefaultSdkValue,
			ResolutionDetails{
				Variant:      VariationSDKDefault,
				Reason:       ReasonError,
				ErrorCode:    ErrorFlagConfiguration,
				ErrorMessage: err.Error(),
				Metadata:     flag.GetMetadata(),
			}
	}

	return flag.GetVariationValue(variationSelection.name), ResolutionDetails{
		Variant:   variationSelection.name,
		Reason:    variationSelection.reason,
		RuleIndex: variationSelection.ruleIndex,
		RuleName:  variationSelection.ruleName,
		Cacheable: variationSelection.cacheable && needsCacheable,
		Metadata:  constructMetadata(flag.GetMetadata(), variationSelection.ruleName),
	}
}

// selectEvaluationReason is choosing which reason has been chosen for the evaluation.
func selectEvaluationReason(
	hasRule, targetingMatch, isDynamic, isDefaultRule bool,
) ResolutionReason {
	if hasRule && targetingMatch {
		if isDynamic {
			return ReasonTargetingMatchSplit
		}
		return ReasonTargetingMatch
	}

	if isDefaultRule {
		if isDynamic {
			return ReasonSplit
		}
		if hasRule {
			return ReasonDefault
		}
		return ReasonStatic
	}
	return ReasonUnknown
}

func (f *InternalFlag) isCacheable() bool {
	isDynamic := (f.Scheduled != nil && len(*f.Scheduled) > 0) || f.Experimentation != nil
	return !isDynamic
}

// checkNeeds evaluates all the dependencies declared in the flag's `needs` field.
// It returns the name of the first unsatisfied dependency and satisfied=false as soon as a
// dependency is unmet, or an empty string and satisfied=true when all dependencies are satisfied
// (or the flag declares none).
//
// The third return value, cacheable, is the AND of the cacheability of every dependency evaluated
// so far. When a dependency is time-based (scheduled/experimentation/progressive) its result is
// not cacheable, so the dependent flag's result must not be cached either — otherwise a client
// would keep a stale enabled/disabled value after the dependency flips over time.
//
// A dependency is unmet when the dependency flag cannot be resolved (missing flag or no resolver
// available, i.e. fail-closed) or when its resolved value does not match the expected value.
func (f *InternalFlag) checkNeeds(
	evaluationCtx ffcontext.Context,
	flagContext Context,
) (unsatisfiedDependency string, satisfied bool, cacheable bool) {
	if f.Needs == nil {
		return "", true, true
	}

	// Dependencies are resolved using the same evaluation context, but must not leak the dependent
	// flag's SDK default value nor trigger a transitive `needs` resolution.
	dependencyContext := flagContext
	dependencyContext.DefaultSdkValue = nil
	dependencyContext.DependencyFlagResolver = nil

	cacheable = true
	for _, need := range *f.Needs {
		dependencyName := need.GetFlag()
		if flagContext.DependencyFlagResolver == nil {
			return dependencyName, false, cacheable
		}
		dependencyFlag, found := flagContext.DependencyFlagResolver(dependencyName)
		if !found || dependencyFlag == nil {
			return dependencyName, false, cacheable
		}

		var resolvedValue any
		var resolvedDetails ResolutionDetails
		if internalDependency, ok := dependencyFlag.(*InternalFlag); ok {
			// Ignore the dependency's own `needs` field (one level only).
			resolvedValue, resolvedDetails = internalDependency.value(dependencyName, evaluationCtx, dependencyContext, false)
		} else {
			// Fallback for any other flag.Flag implementation. InternalFlag is the only one today,
			// so this branch is currently unreachable. A future implementation is responsible for
			// enforcing the one-level rule itself, as we cannot skip its `needs` resolution here.
			resolvedValue, resolvedDetails = dependencyFlag.Value(dependencyName, evaluationCtx, dependencyContext)
		}
		cacheable = cacheable && resolvedDetails.Cacheable

		// A dependency that is itself disabled or errored has no meaningful value; the need is
		// unmet regardless of the expected value (fail closed). This also prevents a `value: null`
		// expectation from spuriously matching a disabled dependency that resolves to nil.
		if resolvedDetails.Reason == ReasonDisabled || resolvedDetails.Reason == ReasonError {
			return dependencyName, false, cacheable
		}

		if !needsValueEqual(resolvedValue, need.GetExpectedValue()) {
			return dependencyName, false, cacheable
		}
	}
	return "", true, cacheable
}

// selectVariation is doing the magic to select the variation that should be used for this specific user
// to always affect the user to the same segment we are using a hash of the flag name + key
func (f *InternalFlag) selectVariation(
	flagName string, key string, ctx ffcontext.Context) (*variationSelection, error) {
	hasRule := len(f.GetRules()) != 0
	// Check all targeting in order, the first to match will be the one used.
	if hasRule {
		for ruleIndex, target := range f.GetRules() {
			variationName, err := target.Evaluate(key, ctx, flagName, false)
			if err != nil {
				// the targeting does not apply
				if _, ok := err.(*internalerror.RuleNotApplyError); ok {
					continue
				}
				return nil, err
			}
			reason := selectEvaluationReason(hasRule, true, target.IsDynamic(), false)
			return &variationSelection{
				name:      variationName,
				reason:    reason,
				ruleIndex: &ruleIndex,
				ruleName:  f.GetRules()[ruleIndex].Name,
				cacheable: f.isCacheable() && target.ProgressiveRollout == nil,
			}, err
		}
	}

	if f.DefaultRule == nil {
		return nil, fmt.Errorf("no default targeting for the flag")
	}

	variationName, err := f.GetDefaultRule().Evaluate(key, ctx, flagName, true)
	if err != nil {
		return nil, err
	}

	reason := selectEvaluationReason(hasRule, false, f.GetDefaultRule().IsDynamic(), true)
	return &variationSelection{
		name:      variationName,
		reason:    reason,
		cacheable: f.isCacheable() && f.GetDefaultRule().ProgressiveRollout == nil,
	}, nil
}

// nolint: gocognit
// applyScheduledRolloutSteps is checking if the flag has a scheduled rollout configured.
// If yes we merge the changes to the current flag.
func (f *InternalFlag) applyScheduledRolloutSteps(evaluationDate time.Time) (*InternalFlag, error) {
	if f.Scheduled == nil {
		return f, nil
	}

	// We are doing a deep copy the flag to avoid modifying the original flag.
	// The deep copy is done to fix this issue https://github.com/thomaspoignant/go-feature-flag/issues/2256
	data, err := json.Marshal(f)
	if err != nil {
		return &InternalFlag{}, err
	}
	var flagCopy *InternalFlag
	if err := json.Unmarshal(data, &flagCopy); err != nil {
		return &InternalFlag{}, err
	}

	// We only keep the steps that are already active (date in the past or now).
	dueSteps := make([]ScheduledStep, 0, len(*f.Scheduled))
	for _, step := range *f.Scheduled {
		if step.Date != nil &&
			(step.Date.Before(evaluationDate) || step.Date.Equal(evaluationDate)) {
			dueSteps = append(dueSteps, step)
		}
	}

	// We apply the steps in chronological order (oldest first) so that, when
	// several steps edit the same rule/field, the most recent step always wins
	// regardless of the order they are declared in the configuration.
	// The sort is stable so that steps sharing the exact same date keep their
	// declaration order (the last one declared is applied last, and wins).
	sort.SliceStable(dueSteps, func(i, j int) bool {
		return dueSteps[i].Date.Before(*dueSteps[j].Date)
	})

	// We apply the scheduled rollout
	for _, steps := range dueSteps {
		flagCopy.Rules = MergeSetOfRules(flagCopy.GetRules(), steps.GetRules())
		if steps.Disable != nil {
			flagCopy.Disable = steps.Disable
		}

		if steps.TrackEvents != nil {
			flagCopy.TrackEvents = steps.TrackEvents
		}

		if steps.DefaultRule != nil {
			flagCopy.DefaultRule.MergeRules(*steps.DefaultRule)
		}

		if steps.Variations != nil {
			maps.Copy(flagCopy.GetVariations(), steps.GetVariations())
		}

		if steps.Version != nil {
			flagCopy.Version = steps.Version
		}

		if steps.Experimentation != nil {
			if flagCopy.Experimentation == nil {
				flagCopy.Experimentation = &ExperimentationRollout{}
			}
			if steps.Experimentation.Start != nil {
				flagCopy.Experimentation.Start = steps.Experimentation.Start
			}
			if steps.Experimentation.End != nil {
				flagCopy.Experimentation.End = steps.Experimentation.End
			}
		}
	}
	return flagCopy, nil
}

// isExperimentationOver checks if we are in an experimentation or not
func (f *InternalFlag) isExperimentationOver(evaluationDate time.Time) bool {
	return f.Experimentation != nil &&
		((f.Experimentation.Start != nil && evaluationDate.Before(*f.Experimentation.Start)) ||
			(f.Experimentation.End != nil && evaluationDate.After(*f.Experimentation.End)))
}

// IsValid is checking if the current flag is valid.
// It does not validate the `needs` self-dependency rule because that requires knowing the flag's
// own name; use IsValidWithFlagName when the flag name is available (e.g. from the config map key).
func (f *InternalFlag) IsValid() error {
	return f.IsValidWithFlagName("")
}

// IsValidWithFlagName is checking if the current flag is valid, including that it does not depend
// on itself through its `needs` field. flagName is the name of the flag being validated; passing
// an empty string skips only the self-dependency check.
func (f *InternalFlag) IsValidWithFlagName(flagName string) error {
	if len(f.GetVariations()) == 0 {
		return fmt.Errorf("no variation available")
	}

	// Check that all variation has the same types
	expectedVarType := ""
	for name, value := range f.GetVariations() {
		if value == nil {
			return fmt.Errorf("nil value for variation: %s", name)
		}
		if expectedVarType != "" {
			currentType, err := utils.JSONTypeExtractor(*value)
			if err != nil {
				return err
			}
			if currentType != expectedVarType {
				return fmt.Errorf("all variations should have the same type")
			}
		} else {
			var err error
			expectedVarType, err = utils.JSONTypeExtractor(*value)
			if err != nil {
				return err
			}
		}
	}

	// Validate that we have a default Rule
	if f.GetDefaultRule() == nil {
		return fmt.Errorf("missing default rule")
	}

	const isDefaultRule = true
	// Validate rules
	if err := f.GetDefaultRule().IsValid(isDefaultRule, f.GetVariations()); err != nil {
		return err
	}

	ruleNames := map[string]any{}
	for _, rule := range f.GetRules() {
		if err := rule.IsValid(!isDefaultRule, f.GetVariations()); err != nil {
			return err
		}

		// Check if we have duplicated rule name
		if _, ok := ruleNames[rule.GetName()]; ok && rule.GetName() != "" {
			return fmt.Errorf("duplicated rule name: %s", rule.GetName())
		} else if rule.GetName() != "" {
			ruleNames[rule.GetName()] = nil
		}
	}

	// Validate the `needs` dependencies.
	if f.Needs != nil {
		for _, need := range *f.Needs {
			dependencyName := need.GetFlag()
			if dependencyName == "" {
				return fmt.Errorf("needs: a dependency is missing its flag name")
			}
			if dependencyName == flagName {
				return fmt.Errorf("needs: a flag cannot depend on itself (%s)", flagName)
			}
		}
	}

	return nil
}

// GetVariations is the getter of the field Variations
func (f *InternalFlag) GetVariations() map[string]*any {
	if f.Variations == nil {
		return map[string]*any{}
	}
	return *f.Variations
}

func (f *InternalFlag) GetRules() []Rule {
	if f.Rules == nil {
		return []Rule{}
	}
	return *f.Rules
}

func (f *InternalFlag) GetRuleIndexByName(name string) *int {
	for index, rule := range f.GetRules() {
		if rule.GetName() == name {
			return &index
		}
	}
	return nil
}

// GetDefaultRule is the getter of the field DefaultRule
func (f *InternalFlag) GetDefaultRule() *Rule {
	return f.DefaultRule
}

// IsTrackEvents is the getter of the field TrackEvents
func (f *InternalFlag) IsTrackEvents() bool {
	if f.TrackEvents == nil {
		return true
	}
	return *f.TrackEvents
}

// IsDisable is the getter for the field Disable
func (f *InternalFlag) IsDisable() bool {
	if f.Disable == nil {
		return false
	}
	return *f.Disable
}

// GetVersion is the getter for the field Version
func (f *InternalFlag) GetVersion() string {
	if f.Version == nil {
		return ""
	}
	return *f.Version
}

// GetVariationValue return the value of variation from his name
func (f *InternalFlag) GetVariationValue(name string) any {
	if v, exists := f.GetVariations()[name]; exists && v != nil {
		return *v
	}
	return nil
}

// GetBucketingKey return the name of the custom bucketing key if we are using one.
func (f *InternalFlag) GetBucketingKey() string {
	if f.BucketingKey == nil {
		return ""
	}
	return *f.BucketingKey
}

// RequiresBucketing checks if the flag requires a bucketing key for evaluation
// A flag requires bucketing if it has percentage-based rules or progressive rollouts,
// including those introduced by scheduled rollout steps
func (f *InternalFlag) RequiresBucketing() bool {
	// Check if default rule requires bucketing and if any targeting rule requires bucketing
	if f.defaultRuleRequiresBucketing() || f.rulesRequireBucketing(f.GetRules()) {
		return true
	}

	// Check if any scheduled rollout steps introduce bucketing requirements
	if f.Scheduled != nil {
		for _, step := range *f.Scheduled {
			if step.defaultRuleRequiresBucketing() || step.rulesRequireBucketing(step.GetRules()) {
				return true
			}
		}
	}
	return false
}

// rulesRequireBucketing checks if any of the rules require bucketing
func (f *InternalFlag) rulesRequireBucketing(rules []Rule) bool {
	for _, rule := range rules {
		if rule.RequiresBucketing() {
			return true
		}
	}
	return false
}

// defaultRuleRequiresBucketing checks if the default rule requires bucketing
func (f *InternalFlag) defaultRuleRequiresBucketing() bool {
	return f.DefaultRule != nil && f.DefaultRule.RequiresBucketing()
}

// GetBucketingKeyValue return the value of the bucketing key from the context
// If requiresBucketing is false, it allows empty keys for flags that don't need them
func (f *InternalFlag) GetBucketingKeyValue(ctx ffcontext.Context) (string, error) {
	// Cache the bucketing requirement check to avoid multiple calls
	requiresBucketing := f.RequiresBucketing()

	// Check if custom bucketing key is provided
	if f.BucketingKey != nil {
		key := f.GetBucketingKey()
		if key == "" {
			return ctx.GetKey(), nil
		}

		value, err := utils.GetNestedFieldValue(ctx.GetCustom(), key)
		if err != nil {
			return "", fmt.Errorf("impossible to find bucketingKey in context: %w", err)
		}

		switch v := value.(type) {
		case string:
			if v == "" {
				return f.requiresBucketingCheck(requiresBucketing, "bucketing")
			}
			return v, nil
		default:
			return "", fmt.Errorf("invalid bucketing key")
		}
	}

	// Check if targeting key is required for this flag
	if ctx.GetKey() == "" {
		return f.requiresBucketingCheck(requiresBucketing, "targeting")
	}

	return ctx.GetKey(), nil
}

// requiresBucketingCheck checks if the bucketing is required and returns the appropriate error
func (f *InternalFlag) requiresBucketingCheck(requiresBucketing bool, key string) (string, error) {
	if requiresBucketing {
		return "", &internalerror.EmptyBucketingKeyError{Message: fmt.Sprintf("Empty %s key", key)}
	}
	return "", nil
}

// GetMetadata return the metadata associated to the flag
func (f *InternalFlag) GetMetadata() map[string]any {
	if f.Metadata == nil {
		return nil
	}
	return *f.Metadata
}

// GetNeeds is the getter for the field Needs
func (f *InternalFlag) GetNeeds() []NeedsDependency {
	if f.Needs == nil {
		return nil
	}
	return *f.Needs
}

func DateFromContextOrDefault(ctx ffcontext.Context, defaultDate time.Time) time.Time {
	if ctx == nil || ctx.ExtractGOFFProtectedFields().CurrentDateTime == nil {
		return defaultDate
	}
	return *ctx.ExtractGOFFProtectedFields().CurrentDateTime
}
