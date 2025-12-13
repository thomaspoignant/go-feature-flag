package service

import (
	"fmt"
	"time"

	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/model"
)

// Monitoring is the interface of the monitoring service
type Monitoring interface {
	Health() model.HealthResponse
	Info() (model.InfoResponse, error)
}

// NewMonitoring creates a new implementation of Monitoring
func NewMonitoring(flagsetManager FlagsetManager) Monitoring {
	return &monitoringImpl{
		flagsetManager: flagsetManager,
	}
}

// monitoringImpl is one implementation of the Monitoring interface
type monitoringImpl struct {
	flagsetManager FlagsetManager
}

// Health returns a static object to show that the server is initialized
func (m *monitoringImpl) Health() model.HealthResponse {
	return model.HealthResponse{
		Initialized: true,
	}
}

// Info returns information about the relay-proxy
func (m *monitoringImpl) Info() (model.InfoResponse, error) {
	if m.flagsetManager == nil {
		return model.InfoResponse{}, fmt.Errorf("flagset manager is not initialized")
	}
	flagSets, err := m.flagsetManager.AllFlagSets()
	if err != nil {
		return model.InfoResponse{}, err
	}

	if m.flagsetManager.IsDefaultFlagSet() {
		if m.flagsetManager.Default() == nil {
			return model.InfoResponse{}, fmt.Errorf("no default flagset configured")
		}
		cacheRefreshDate := m.flagsetManager.Default().GetCacheRefreshDate()
		return model.InfoResponse{
			LatestCacheRefresh: &cacheRefreshDate,
		}, nil
	}

	refreshDates := make(map[string]time.Time, len(flagSets))
	latestRefreshDate := time.Time{}
	for flagsetName, flagset := range flagSets {
		refreshDates[flagsetName] = flagset.GetCacheRefreshDate()
		if refreshDates[flagsetName].After(latestRefreshDate) {
			latestRefreshDate = refreshDates[flagsetName]
		}
	}
	return model.InfoResponse{
		Flagsets:           refreshDates,
		LatestCacheRefresh: &latestRefreshDate,
	}, nil
}
