package testconvert

import "time"

// Bool returns a pointer to the bool value passed in.
func Bool(v bool) *bool {
	return &v
}

// Time returns a pointer to the Time value passed in.
func Time(t time.Time) *time.Time {
	return &t
}

// Float64 returns a pointer to the float64 value passed in.
func Float64(t float64) *float64 {
	return &t
}

// Int returns a pointer to the float64 value passed in.
func Int(t int) *int {
	return &t
}

// Interface returns a pointer to the interface value passed in.
func Interface(v interface{}) *interface{} {
	return &v
}

// String returns a pointer to the string value passed in.
func String(v string) *string {
	return &v
}
