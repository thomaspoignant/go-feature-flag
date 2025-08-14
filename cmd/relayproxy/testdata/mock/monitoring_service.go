package mock

import "github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/model"

// MockMonitoringService is a mock implementation for testing error scenarios
type MockMonitoringService struct {
	HealthResponse model.HealthResponse
	InfoResponse   model.InfoResponse
	InfoError      error
}

func (m *MockMonitoringService) Health() model.HealthResponse {
	return m.HealthResponse
}

func (m *MockMonitoringService) Info() (model.InfoResponse, error) {
	return m.InfoResponse, m.InfoError
}
