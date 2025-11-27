package flag

import "maps"

// constructMetadata is the internal generic func used to enhance model.VariationResult adding
// the targeting.rule's name (from configuration) to the Metadata.
// That way, it is possible to see when a targeting rule is match during the evaluation process.
func constructMetadata(
	flagMetadata map[string]any, ruleName *string) map[string]any {
	metadata := maps.Clone(flagMetadata)
	if ruleName == nil || *ruleName == "" {
		return metadata
	}
	if metadata == nil {
		metadata = make(map[string]any)
	}
	metadata["evaluatedRuleName"] = *ruleName
	return metadata
}
