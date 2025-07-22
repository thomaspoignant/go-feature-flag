package model

import "time"

// InfoResponse is the object returned by the info API
type InfoResponse struct {
	// LatestCacheRefresh is the last time when your flag file was read and stored in the internal cache.
	// This field is used for backward compatibility when using the default flagset.
	LatestCacheRefresh *time.Time `json:"cacheRefresh,omitempty" example:"2022-06-13T11:22:55.941628+02:00"`

	// Flagsets contains the cache refresh dates for each flagset when using multiple flagsets.
	// The format is {"flagset name": "2022-06-13T11:22:55.941628+02:00"}
	Flagsets map[string]time.Time `json:"flagsets,omitempty" example:"default:2022-06-13T11:22:55.941628+02:00,feature-flags:2022-06-13T11:22:55.941628+02:00"` //nolint: lll
}
