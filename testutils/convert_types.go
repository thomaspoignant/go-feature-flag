package testutils

import "time"

// Bool returns a pointer to the bool value passed in.
func Bool(v bool) *bool {
	return &v
}

// Time returns a pointer to the Time value passed in.
func Time(t time.Time) *time.Time {
	return &t
}
