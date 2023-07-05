package service

import (
	ffclient "github.com/thomaspoignant/go-feature-flag"
	"github.com/thomaspoignant/go-feature-flag/cmd/server/model"
)

// Monitoring is the interface of the monitoring service
type Monitoring interface {
	Health() model.HealthResponse
	Info() model.InfoResponse
}

// NewMonitoring creates a new implementation of Monitoring
func NewMonitoring(goFF *ffclient.GoFeatureFlag) Monitoring {
	return &monitoringImpl{
		goFF: goFF,
	}
}

// monitoringImpl is one implementation of the Monitoring interface
type monitoringImpl struct {
	goFF *ffclient.GoFeatureFlag
}

// Health returns a static object to show that the server is initialized
func (m *monitoringImpl) Health() model.HealthResponse {
	return model.HealthResponse{
		Initialized: true,
	}
}

// Info returns information about the relay-proxy
func (m *monitoringImpl) Info() model.InfoResponse {
	return model.InfoResponse{
		LatestCacheRefresh: m.goFF.GetCacheRefreshDate(),
	}
}
