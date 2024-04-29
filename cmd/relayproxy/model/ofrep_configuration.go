package model

type OFREPConfiguration struct {
	Name         string                  `json:"name"`
	Capabilities OFREPConfigCapabilities `json:"capabilities,omitempty"`
}

type OFREPConfigCapabilities struct {
	CacheInvalidation OFREPConfigCapabilitiesCacheInvalidation `json:"cacheInvalidation,omitempty"`
}

type OFREPConfigCapabilitiesCacheInvalidation struct {
	Polling OFREPConfigCapabilitiesCacheInvalidationPolling `json:"polling,omitempty"`
}

type OFREPConfigCapabilitiesCacheInvalidationPolling struct {
	Enabled            bool  `json:"enabled,omitempty"`
	MinPollingInterval int64 `json:"minPollingInterval,omitempty"`
}
