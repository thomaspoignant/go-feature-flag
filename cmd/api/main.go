package main

import (
	"encoding/json"
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
	Description     string    `json:"description"`
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
	Disable bool `json:"disable"`
}

// Define a slice to store feature flags (simulating a database).
var featureFlags []FeatureFlag

func main() {
	featureFlags = initFeatureFlag2()

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
	// time.Sleep(5 * time.Second)
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
	featureFlags[flagIndex].Disable = testconvert.Bool(statusUpdate.Disable)
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

// nolint
var request1 = "{\n  \"description\":\"This is a feature flag\",  \"variations\": {\n        \"variation1\": true,\n        \"variation2\": false\n    },\n    \"targeting\": [\n        {\n            \"name\": \"targetRule1\",\n            \"query\": \"user.id == 456\",\n            \"variation\": \"variation1\",\n            \"percentage\": {\n                \"variation1\": 70,\n                \"variation2\": 30\n            },\n            \"progressiveRollout\": {\n                \"initial\": {\n                    \"variation\": \"variation1\",\n                    \"percentage\": 10,\n                    \"date\": \"2023-01-02T00:00:00Z\"\n                },\n                \"end\": {\n                    \"variation\": \"variation2\",\n                    \"percentage\": 100,\n                    \"date\": \"2023-12-30T23:59:59Z\"\n                }\n            },\n            \"disable\": false\n        }\n    ],\n    \"defaultRule\": {\n        \"name\": \"defaultRule\",\n        \"query\": \"user.id == 789\",\n        \"variation\": \"variation2\",\n        \"percentage\": {\n            \"variation1\": 40,\n            \"variation2\": 60\n        },\n        \"progressiveRollout\": {\n            \"initial\": {\n                \"variation\": \"variation2\",\n                \"percentage\": 20,\n                \"date\": \"2023-02-01T00:00:00Z\"\n            },\n            \"end\": {\n                \"variation\": \"variation1\",\n                \"percentage\": 100,\n                \"date\": \"2023-12-29T23:59:59Z\"\n            }\n        },\n        \"disable\": true\n    },\n    \"scheduledRollout\": [\n        {\n            \"variations\": {\n                \"variation3\": \"scheduled_value1\",\n                \"variation4\": \"scheduled_value2\"\n            },\n            \"targeting\": [\n                {\n                    \"name\": \"scheduledRule1\",\n                    \"query\": \"user.id == 555\",\n                    \"variation\": \"variation3\",\n                    \"percentage\": {\n                        \"variation3\": 80,\n                        \"variation4\": 20\n                    },\n                    \"progressiveRollout\": {\n                        \"initial\": {\n                            \"variation\": \"variation3\",\n                            \"percentage\": 30,\n                            \"date\": \"2023-03-01T00:00:00Z\"\n                        },\n                        \"end\": {\n                            \"variation\": \"variation4\",\n                            \"percentage\": 100,\n                            \"date\": \"2023-12-28T23:59:59Z\"\n                        }\n                    },\n                    \"disable\": false\n                }\n            ],\n            \"defaultRule\": {\n                \"name\": \"scheduledDefaultRule\",\n                \"query\": \"user.id == 999\",\n                \"variation\": \"variation4\",\n                \"percentage\": {\n                    \"variation3\": 60,\n                    \"variation4\": 40\n                },\n                \"progressiveRollout\": {\n                    \"initial\": {\n                        \"variation\": \"variation4\",\n                        \"percentage\": 40,\n                        \"date\": \"2023-04-01T00:00:00Z\"\n                    },\n                    \"end\": {\n                        \"variation\": \"variation3\",\n                        \"percentage\": 100,\n                        \"date\": \"2023-12-27T23:59:59Z\"\n                    }\n                },\n                \"disable\": true\n            },\n            \"experimentation\": {\n                \"start\": \"2023-01-01T00:00:00Z\",\n                \"end\": \"2023-12-31T23:59:59Z\"\n            },\n            \"metadata\": {},\n            \"disable\": false,\n            \"version\": \"v2\",\n            \"trackEvents\": true,\n            \"date\": \"2023-12-28T22:00:00+01:00\"\n        }\n    ],\n    \"experimentation\": {\n        \"start\": \"2023-02-01T00:00:00Z\",\n        \"end\": \"2023-11-30T23:59:59Z\"\n    },\n    \"metadata\": {\n        \"newMetadataField\": \"newMetadataValue\"\n ,\"toto\":\"titi\"   },\n    \"disable\": false,\n    \"id\": \"f1f7a727-92ab-4a21-97ef-36390e3b8433\",\n    \"name\": \"FeatureFlag1\",\n    \"createdDate\": \"2023-12-28T21:57:24.713911+01:00\",\n    \"lastUpdatedDate\": \"2023-12-28T22:00:00+01:00\"\n}\n"

// nolint
var request2 = "{\n  \"description\":\"This is a feature flag, rekudgwakgscadkg dwksugawoig skgwskig wskugwksig wkgwskgswo olwghsksg\",  \"variations\": {\n        \"variationA\": \"valueA\",\n        \"variationB\": \"valueB\"\n    },\n    \"targeting\": [],\n    \"defaultRule\": {\n        \"name\": \"defaultRule\",\n        \"query\": \"user.id == 456\",\n        \"variation\": \"variationA\",\n        \"percentage\": {\n            \"variationA\": 30,\n            \"variationB\": 70\n        },\n        \"progressiveRollout\": {\n            \"initial\": {\n                \"variation\": \"variationA\",\n                \"percentage\": 10,\n                \"date\": \"2023-01-03T00:00:00Z\"\n            },\n            \"end\": {\n                \"variation\": \"variationB\",\n                \"percentage\": 100,\n                \"date\": \"2023-12-28T23:59:59Z\"\n            }\n        },\n        \"disable\": true\n    },\n    \"scheduledRollout\": [],\n    \"experimentation\": {\n        \"start\": \"2023-03-01T00:00:00Z\",\n        \"end\": \"2023-10-31T23:59:59Z\"\n    },\n    \"metadata\": {\n        \"customField\": \"customValue\"\n    },\n    \"disable\": false,\n    \"id\": \"a7c529f9-54a4-4e63-afaf-2581ef48d192\",\n    \"name\": \"FeatureFlag2\",\n    \"createdDate\": \"2023-12-28T21:57:24.713911+01:00\",\n    \"lastUpdatedDate\": \"2023-12-28T23:00:00+01:00\"\n}\n"

// nolint
var request3 = "{\n    \"variations\": {\n        \"option1\": \"result1\",\n        \"option2\": \"result2\"\n    },\n    \"targeting\": [\n        {\n            \"name\": \"customRule\",\n            \"query\": \"user.id == 789\",\n            \"variation\": \"option1\",\n            \"percentage\": {\n                \"option1\": 50,\n                \"option2\": 50\n            },\n            \"progressiveRollout\": {\n                \"initial\": {\n                    \"variation\": \"option1\",\n                    \"percentage\": 20,\n                    \"date\": \"2023-01-04T00:00:00Z\"\n                },\n                \"end\": {\n                    \"variation\": \"option2\",\n                    \"percentage\": 100,\n                    \"date\": \"2023-12-28T23:59:59Z\"\n                }\n            },\n            \"disable\": false\n        }\n    ],\n    \"defaultRule\": {\n        \"name\": \"defaultRule\",\n        \"query\": \"user.id == 999\",\n        \"variation\": \"option2\",\n        \"percentage\": {\n            \"option1\": 40,\n            \"option2\": 60\n        },\n        \"progressiveRollout\": {\n            \"initial\": {\n                \"variation\": \"option2\",\n                \"percentage\": 30,\n                \"date\": \"2023-01-05T00:00:00Z\"\n            },\n            \"end\": {\n                \"variation\": \"option1\",\n                \"percentage\": 100,\n                \"date\": \"2023-12-28T23:59:59Z\"\n            }\n        },\n        \"disable\": false\n    },\n    \"scheduledRollout\": [\n        {\n            \"variations\": {\n                \"scheduleA\": \"valueA\",\n                \"scheduleB\": \"valueB\"\n            },\n            \"targeting\": [\n                {\n                    \"name\": \"scheduleRule\",\n                    \"query\": \"user.id == 111\",\n                    \"variation\": \"scheduleA\",\n                    \"percentage\": {\n                        \"scheduleA\": 60,\n                        \"scheduleB\": 40\n                    },\n                    \"progressiveRollout\": {\n                        \"initial\": {\n                            \"variation\": \"scheduleA\",\n                            \"percentage\": 40,\n                            \"date\": \"2023-01-06T00:00:00Z\"\n                        },\n                        \"end\": {\n                            \"variation\": \"scheduleB\",\n                            \"percentage\": 100,\n                            \"date\": \"2023-12-28T23:59:59Z\"\n                        }\n                    },\n                    \"disable\": true\n                }\n            ],\n            \"defaultRule\": {\n                \"name\": \"scheduleDefaultRule\",\n                \"query\": \"user.id == 222\",\n                \"variation\": \"scheduleB\",\n                \"percentage\": {\n                    \"scheduleA\": 70,\n                    \"scheduleB\": 30\n                },\n                \"progressiveRollout\": {\n                    \"initial\": {\n                        \"variation\": \"scheduleB\",\n                        \"percentage\": 50,\n                        \"date\": \"2023-01-07T00:00:00Z\"\n                    },\n                    \"end\": {\n                        \"variation\": \"scheduleA\",\n                        \"percentage\": 100,\n                        \"date\": \"2023-12-28T23:59:59Z\"\n                    }\n                },\n                \"disable\": false\n            },\n            \"experimentation\": {\n                \"start\": \"2023-01-01T00:00:00Z\",\n                \"end\": \"2023-12-31T23:59:59Z\"\n            },\n            \"metadata\": {\n                \"newMetadataField\": \"newMetadataValue\"\n            },\n            \"disable\": false,\n            \"version\": \"v3\",\n            \"trackEvents\": false,\n            \"date\": \"2023-12-28T23:30:00+01:00\"\n        }\n    ],\n    \"experimentation\": {\n        \"start\": \"2023-04-01T00:00:00Z\",\n        \"end\": \"2023-09-30T23:59:59Z\"\n    },\n    \"metadata\": {},\n    \"disable\": true,\n    \"id\": \"6c7e1a2a-af5f-42d3-bb4b-9f3986a68a37\",\n    \"name\": \"FeatureFlag3\",\n    \"createdDate\": \"2023-12-28T21:57:24.713911+01:00\",\n    \"lastUpdatedDate\": \"2023-12-28T23:30:00+01:00\"\n}\n"
var request4 = "{  \"version\":\"v1.0.0\", \"variations\": {  \"option1\": 10,  \"option21341\": 20,  \"option2332\": 20,\n  \"option212\": 20,\n  \"option221\": 20,\n  \"option222\": 20,\n  \"option21\": 20,\n  \"option24\": 20,\n  \"option25\": 20,\n  \"option23\": 20,\n  \"option26\": 20,\n  \"option28\": 20,\n  \"option27\": 20,\n  \"option29\": 20},  \"targeting\": [    {      \"name\": \"customRule\",      \"query\": \"user.id == 789\",      \"variation\": \"option1\",      \"percentage\": {        \"option1\": 50,        \"option2\": 50      },      \"progressiveRollout\": {        \"initial\": {          \"variation\": \"option1\",          \"percentage\": 20,          \"date\": \"2023-01-04T00:00:00Z\"        },        \"end\": {          \"variation\": \"option2\",          \"percentage\": 100,          \"date\": \"2023-12-28T23:59:59Z\"        }      },      \"disable\": false    }  ],  \"defaultRule\": {    \"name\": \"defaultRule\",    \"query\": \"user.id == 999\",    \"variation\": \"option2\",    \"percentage\": {      \"option1\": 40,      \"option2\": 60    },    \"progressiveRollout\": {      \"initial\": {        \"variation\": \"option2\",        \"percentage\": 30,        \"date\": \"2023-01-05T00:00:00Z\"      },      \"end\": {        \"variation\": \"option1\",        \"percentage\": 100,        \"date\": \"2023-12-28T23:59:59Z\"      }    },    \"disable\": false  },  \"scheduledRollout\": [    {      \"variations\": {        \"scheduleA\": \"valueA\",        \"scheduleB\": \"valueB\"      },      \"targeting\": [        {          \"name\": \"scheduleRule\",          \"query\": \"user.id == 111\",          \"variation\": \"scheduleA\",          \"percentage\": {            \"scheduleA\": 60,            \"scheduleB\": 40          },          \"progressiveRollout\": {            \"initial\": {              \"variation\": \"scheduleA\",              \"percentage\": 40,              \"date\": \"2023-01-06T00:00:00Z\"            },            \"end\": {              \"variation\": \"scheduleB\",              \"percentage\": 100,              \"date\": \"2023-12-28T23:59:59Z\"            }          },          \"disable\": true        }      ],      \"defaultRule\": {        \"name\": \"scheduleDefaultRule\",        \"query\": \"user.id == 222\",        \"variation\": \"scheduleB\",        \"percentage\": {          \"scheduleA\": 70,          \"scheduleB\": 30        },        \"progressiveRollout\": {          \"initial\": {            \"variation\": \"scheduleB\",            \"percentage\": 50,            \"date\": \"2023-01-07T00:00:00Z\"          },          \"end\": {            \"variation\": \"scheduleA\",            \"percentage\": 100,            \"date\": \"2023-12-28T23:59:59Z\"          }        },        \"disable\": false      },      \"experimentation\": {        \"start\": \"2023-01-01T00:00:00Z\",        \"end\": \"2023-12-31T23:59:59Z\"      },      \"metadata\": {        \"newMetadataField\": \"newMetadataValue\"      },      \"disable\": false,      \"version\": \"v3\",      \"trackEvents\": false,      \"date\": \"2023-12-28T23:30:00+01:00\"    }  ],  \"experimentation\": {    \"start\": \"2023-04-01T00:00:00Z\",    \"end\": \"2023-09-30T23:59:59Z\"  },  \"metadata\": {},  \"disable\": true,  \"id\": \"6c7e1a2a-af5f-42d3-bb4b-9f3986a68b37\",  \"name\": \"FeatureFlag4\",  \"createdDate\": \"2023-12-28T21:57:24.713911+01:00\",  \"lastUpdatedDate\": \"2023-12-28T23:30:00+01:00\"}"

func initFeatureFlag2() []FeatureFlag {
	var featureFlags = make([]FeatureFlag, 4)
	if err := json.Unmarshal([]byte(request1), &featureFlags[0]); err != nil {
		fmt.Println("Error unmarshalling JSON:", err)
		return nil
	}

	if err := json.Unmarshal([]byte(request2), &featureFlags[1]); err != nil {
		fmt.Println("Error unmarshalling JSON:", err)
		return nil
	}

	if err := json.Unmarshal([]byte(request3), &featureFlags[2]); err != nil {
		fmt.Println("Error unmarshalling JSON:", err)
		return nil
	}

	if err := json.Unmarshal([]byte(request4), &featureFlags[3]); err != nil {
		fmt.Println("Error unmarshalling JSON:", err)
		return nil
	}
	return featureFlags
}
