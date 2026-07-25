package postgresqlretriever

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/thomaspoignant/go-feature-flag/modules/core/dto"
)

// computeETag returns a strong ETag for a flag definition. It is computed over the
// jsonEncodedDefinition JSON encoding of the definition (parsed into a dto.DTO and re-marshaled, which
// deterministically sorts map keys) rather than the raw bytes: the raw bytes may come from
// Postgres JSONB, which does not guarantee stable key order or whitespace across rows/versions.
func computeETag(definition []byte) (string, error) {
	var flagDefinitionDto dto.DTO
	if unmarshallingError := json.Unmarshal(definition, &flagDefinitionDto); unmarshallingError != nil {
		return "", fmt.Errorf("impossible to compute etag, invalid flag definition: %w", unmarshallingError)
	}
	jsonEncodedDefinition, marshallingError := json.Marshal(flagDefinitionDto)
	if marshallingError != nil {
		return "", fmt.Errorf("impossible to compute etag: %w", marshallingError)
	}
	sum := sha256.Sum256(jsonEncodedDefinition)
	return `"` + hex.EncodeToString(sum[:]) + `"`, nil
}
