package utils

import (
	"github.com/thomaspoignant/go-feature-flag/modules/core/ffcontext"
)

// ContextToMap convert the context to a MAP to use the query on it.
func ContextToMap(ctx ffcontext.Context) map[string]any {
	// We don't have a json copy of the user.
	custom := ctx.GetCustom()
	userCopy := make(map[string]any, len(custom)+3)

	// Duplicate the map to keep User un-mutable
	for key, value := range custom {
		userCopy[key] = value
	}
	userCopy["anonymous"] = ctx.IsAnonymous()
	// box the key once for both entries
	k := any(ctx.GetKey())
	userCopy["key"] = k
	userCopy["targetingKey"] = k
	return userCopy
}
