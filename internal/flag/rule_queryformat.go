package flag

import "github.com/thomaspoignant/go-feature-flag/internal/utils"

type QueryFormat = string

const (
	NikunjyQueryFormat   QueryFormat = "nikunjy"
	JSONLogicQueryFormat QueryFormat = "jsonlogic"
)

func GetQueryFormat(rule Rule) QueryFormat {
	if utils.IsJSONObject(rule.GetTrimmedQuery()) {
		return JSONLogicQueryFormat
	}
	return NikunjyQueryFormat
}
