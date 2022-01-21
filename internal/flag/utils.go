package flag

import (
	"fmt"
	"github.com/thomaspoignant/go-feature-flag/ffuser"
)

// userToMap convert the user to a MAP to use the query on it.
func userToMap(u ffuser.User) map[string]interface{} {
	// We don't have a json copy of the user.
	userCopy := make(map[string]interface{})

	// Duplicate the map to keep User un-mutable
	for key, value := range u.GetCustom() {
		userCopy[key] = value
	}
	userCopy["anonymous"] = u.IsAnonymous()
	userCopy["key"] = u.GetKey()
	return userCopy
}

func appendIfHasValue(toString []string, key string, value string) []string {
	if value != "" {
		toString = append(toString, fmt.Sprintf("%s:[%v]", key, value))
	}
	return toString
}
