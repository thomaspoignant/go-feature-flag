package main

import (
	"fmt"
	"github.com/labstack/echo/v4/middleware"
	custommiddleware "github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/api/middleware"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/log"
	"github.com/thomaspoignant/go-feature-flag/internal/dto"
	"github.com/thomaspoignant/go-feature-flag/testutils/testconvert"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// FeatureFlag represents a feature flag.
type FeatureFlag struct {
	dto.DTO
	ID              uuid.UUID `json:"id"`
	Name            string    `json:"name"`
	CreatedDate     time.Time `json:"createdDate"`
	LastUpdatedDate time.Time `json:"lastUpdatedDate"`
}

// ScheduledStep represents a step in scheduled rollout.
type ScheduledStep struct {
	dto.DTO
	Date time.Time `json:"date"`
}

// ExperimentationDto represents experimentation configuration.
type ExperimentationDto struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// ExperimentationRollout represents experimentation rollout configuration.
type ExperimentationRollout struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// ProgressiveRollout represents progressive rollout configuration.
type ProgressiveRollout struct {
	Initial ProgressiveRolloutStep `json:"initial"`
	End     ProgressiveRolloutStep `json:"end"`
}

// ProgressiveRolloutStep represents a step in progressive rollout.
type ProgressiveRolloutStep struct {
	Variation  string    `json:"variation"`
	Percentage float64   `json:"percentage"`
	Date       time.Time `json:"date"`
}

// FeatureFlagInput represents the input for creating a new feature flag.
type FeatureFlagInput struct {
	dto.DTO
	Name            string    `json:"name"`
	CreatedDate     time.Time `json:"createdDate"`
	LastUpdatedDate time.Time `json:"lastUpdatedDate"`
}

// FeatureFlagStatusUpdate represents the input for updating the status of a feature flag.
type FeatureFlagStatusUpdate struct {
	Enable bool `json:"enable"`
}

// Define a slice to store feature flags (simulating a database).
var featureFlags []FeatureFlag

func main() {
	featureFlags = initFeatureFlag()

	e := echo.New()
	zapLog := log.InitLogger()
	defer func() { _ = zapLog.Sync() }()
	e.Use(custommiddleware.ZapLogger(zapLog, nil))
	e.Use(middleware.CORSWithConfig(middleware.DefaultCORSConfig))
	// Routes
	e.GET("/v1/flags", getAllFeatureFlags)
	e.POST("/v1/flags", createFeatureFlag)
	e.GET("/v1/flags/:id", getFeatureFlagByID)
	e.PUT("/v1/flags/:id", updateFeatureFlagByID)
	e.DELETE("/v1/flags/:id", deleteFeatureFlagByID)
	e.PATCH("/v1/flags/:id/status", updateFeatureFlagStatus)

	// Start server
	e.Logger.Fatal(e.Start(":3001"))
}

// Implement the CRUD operations
// (getAllFeatureFlags, createFeatureFlag, getFeatureFlagByID, updateFeatureFlagByID, deleteFeatureFlagByID,
// updateFeatureFlagStatus) as described in the previous responses.
func getAllFeatureFlags(c echo.Context) error {
	if featureFlags == nil {
		featureFlags = []FeatureFlag{}
	}
	return c.JSON(http.StatusOK, featureFlags)
}

func createFeatureFlag(c echo.Context) error {
	var input FeatureFlagInput
	if err := c.Bind(&input); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	// Generate a new UUID for the feature flag
	id := uuid.New()
	now := time.Now()

	// Create the new feature flag
	newFlag := FeatureFlag{
		ID: id,
		DTO: dto.DTO{
			DTOv1: dto.DTOv1{
				Variations:      input.Variations,
				Rules:           input.Rules,
				DefaultRule:     input.DefaultRule,
				Scheduled:       input.Scheduled,
				Experimentation: input.Experimentation,
				Metadata:        input.Metadata,
			},
			Disable: testconvert.Bool(false),
		},
		Name:            input.Name,
		CreatedDate:     now,
		LastUpdatedDate: now,
	}

	// Append the new feature flag to the slice
	featureFlags = append(featureFlags, newFlag)

	return c.JSON(http.StatusCreated, newFlag)
}

func getFeatureFlagByID(c echo.Context) error {
	id := c.Param("id")
	flag, err := findFeatureFlagByID(id)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Feature flag not found"})
	}

	return c.JSON(http.StatusOK, flag)
}

func updateFeatureFlagByID(c echo.Context) error {
	id := c.Param("id")

	// Find the feature flag by ID
	flagIndex, err := findFeatureFlagIndexByID(id)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Feature flag not found"})
	}

	// Parse the update input
	var updateInput FeatureFlagInput
	if err := c.Bind(&updateInput); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	// Update the feature flag fields
	flag := &featureFlags[flagIndex]
	flag.Variations = updateInput.Variations
	flag.Rules = updateInput.Rules
	flag.DefaultRule = updateInput.DefaultRule
	flag.Scheduled = updateInput.Scheduled
	flag.Experimentation = updateInput.Experimentation
	flag.Metadata = updateInput.Metadata
	flag.Disable = updateInput.Disable
	flag.LastUpdatedDate = updateInput.LastUpdatedDate

	return c.JSON(http.StatusOK, flag)
}

func deleteFeatureFlagByID(c echo.Context) error {
	id := c.Param("id")

	// Find the feature flag by ID
	flagIndex, err := findFeatureFlagIndexByID(id)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Feature flag not found"})
	}

	// Remove the feature flag from the slice
	featureFlags = append(featureFlags[:flagIndex], featureFlags[flagIndex+1:]...)

	return c.NoContent(http.StatusNoContent)
}

func updateFeatureFlagStatus(c echo.Context) error {
	id := c.Param("id")

	// Find the feature flag by ID
	flagIndex, err := findFeatureFlagIndexByID(id)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Feature flag not found"})
	}

	// Parse the status update input
	var statusUpdate FeatureFlagStatusUpdate
	if err := c.Bind(&statusUpdate); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	// Update the feature flag status
	featureFlags[flagIndex].Disable = testconvert.Bool(!statusUpdate.Enable)
	featureFlags[flagIndex].LastUpdatedDate = time.Now()

	return c.JSON(http.StatusOK, featureFlags[flagIndex])
}

// Helper function to find a feature flag by ID
func findFeatureFlagByID(id string) (*FeatureFlag, error) {
	parsedID, err := uuid.Parse(id)
	if err != nil {
		return nil, err
	}

	for _, flag := range featureFlags {
		if flag.ID == parsedID {
			return &flag, nil
		}
	}

	return nil, fmt.Errorf("feature flag not found")
}

// Helper function to find the index of a feature flag by ID
func findFeatureFlagIndexByID(id string) (int, error) {
	parsedID, err := uuid.Parse(id)
	if err != nil {
		return -1, err
	}

	for i, flag := range featureFlags {
		if flag.ID == parsedID {
			return i, nil
		}
	}

	return -1, fmt.Errorf("feature flag not found")
}
