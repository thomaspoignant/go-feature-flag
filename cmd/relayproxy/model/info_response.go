package model

import "time"

// InfoResponse is the object returned by the info API
type InfoResponse struct {
	LatestCacheRefresh time.Time `json:"cacheRefresh"`
}
