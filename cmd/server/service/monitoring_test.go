package service_test

import (
	"context"
	"github.com/stretchr/testify/assert"
	ffclient "github.com/thomaspoignant/go-feature-flag"
	"github.com/thomaspoignant/go-feature-flag/retriever/fileretriever"
	"testing"
	"time"

	"github.com/thomaspoignant/go-feature-flag/cmd/server/service"
)

// Mock implementation of GoFeatureFlag for testing
var testGOFeatureFlag, _ = ffclient.New(ffclient.Config{
	PollingInterval: 1 * time.Minute,
	Context:         context.TODO(),
	Retriever:       &fileretriever.Retriever{Path: "../testdata/controller/config_flags.yaml"},
	FileFormat:      "yaml",
})

// Test the Health function of monitoringImpl
func TestMonitoringImpl_Health(t *testing.T) {
	m := service.NewMonitoring(testGOFeatureFlag)
	health := m.Health()
	assert.True(t, health.Initialized, "Expected initialized to be true, got false")
}

// Test the Info function of monitoringImpl
func TestMonitoringImpl_Info(t *testing.T) {
	m := service.NewMonitoring(testGOFeatureFlag)
	info := m.Info()
	assert.False(t, info.LatestCacheRefresh.IsZero(), "Expected LatestCacheRefresh to not be zero, got zero")
}
