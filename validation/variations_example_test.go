package validation_test

import (
	"encoding/json"
	"fmt"

	"github.com/thomaspoignant/go-feature-flag/modules/core/dto"
	"github.com/thomaspoignant/go-feature-flag/validation"
)

func ExampleValidateVariations() {
	const flagKey = "storefront-banner"
	const flagJSON = `{
		"variations": {
			"baseline": {"reward": 3000}
		},
		"defaultRule": {"variation": "baseline"},
		"scheduledRollout": [
			{
				"date": "2026-12-01T10:00:00Z",
				"variations": {"baseline": {"reward": "5000"}}
			}
		]
	}`
	const schema = `{
		"type": "object",
		"required": ["reward"],
		"properties": {"reward": {"type": "integer"}}
	}`

	var definition dto.DTO
	if err := json.Unmarshal([]byte(flagJSON), &definition); err != nil {
		fmt.Println(err)
		return
	}

	violations, err := validation.ValidateVariations(flagKey, definition, []byte(schema))
	if err != nil {
		fmt.Println(err)
		return
	}

	for _, violation := range violations {
		source := "base variation"
		if violation.ScheduledStep != nil {
			source = fmt.Sprintf("scheduledRollout[%d]", *violation.ScheduledStep)
		}
		fmt.Printf("%s %s %s %s: %s\n",
			violation.FlagKey, violation.Variation, source, violation.InstanceLocation, violation.Message)
	}

	// Output:
	// storefront-banner baseline scheduledRollout[0] /reward: expected integer, but got string
}
