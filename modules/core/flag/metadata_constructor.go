package flag

import "maps"

// addMetadataEntry returns a copy of flagMetadata with key=value added.
// When value is empty, the metadata is returned unchanged (a clone of flagMetadata).
func addMetadataEntry(flagMetadata map[string]any, key, value string) map[string]any {
	metadata := maps.Clone(flagMetadata)
	if value == "" {
		return metadata
	}
	if metadata == nil {
		metadata = make(map[string]any)
	}
	metadata[key] = value
	return metadata
}

// constructMetadata is the internal generic func used to enhance model.VariationResult adding
// the targeting.rule's name (from configuration) to the Metadata.
// That way, it is possible to see when a targeting rule is match during the evaluation process.
func constructMetadata(
	flagMetadata map[string]any, ruleName *string) map[string]any {
	name := ""
	if ruleName != nil {
		name = *ruleName
	}
	return addMetadataEntry(flagMetadata, "evaluatedRuleName", name)
}

// constructNeedsMetadata enhances the flag metadata with the name of the dependency (from the
// `needs` field) that was not satisfied. That way it is possible to see, when a flag is disabled
// because of an unmet dependency, which dependency caused it.
func constructNeedsMetadata(
	flagMetadata map[string]any, unsatisfiedDependency string) map[string]any {
	return addMetadataEntry(flagMetadata, "unsatisfiedDependency", unsatisfiedDependency)
}
