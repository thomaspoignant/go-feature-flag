package model

// HealthResponse is the object returned by the health API
type HealthResponse struct {
	Initialized bool `json:"initialized"`
}
