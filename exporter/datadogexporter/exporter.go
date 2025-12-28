package datadogexporter

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/thomaspoignant/go-feature-flag/exporter"
	"github.com/thomaspoignant/go-feature-flag/internal"
	"github.com/thomaspoignant/go-feature-flag/utils/fflog"
)

const (
	// DatadogDefaultSite is the default Datadog site (US1).
	DatadogDefaultSite = "datadoghq.com"
	// DatadogDefaultSource is the default source tag for logs.
	DatadogDefaultSource = "go-feature-flag"
	// DatadogDefaultService is the default service tag for logs.
	DatadogDefaultService = "go-feature-flag"
)

// Exporter sends feature flag evaluation events to Datadog using the HTTP Log intake API.
// Events are sent as structured logs that can be queried in Datadog Log Explorer
// and correlated with RUM sessions using user identifiers.
type Exporter struct {
	// APIKey is the Datadog API key used for authentication.
	// This is required and can be obtained from your Datadog account.
	APIKey string

	// Site is the Datadog site to send data to (e.g., "datadoghq.com", "datadoghq.eu", "us5.datadoghq.com").
	// Defaults to "datadoghq.com" (US1) if not specified.
	Site string

	// Source is the source tag for logs. Defaults to "go-feature-flag".
	Source string

	// Service is the service name for logs. Defaults to "go-feature-flag".
	Service string

	// Tags is a list of tags to attach to every log entry (e.g., ["env:production", "team:backend"]).
	Tags []string

	httpClient internal.HTTPClient
	init       sync.Once
}

// datadogLogEntry represents a single log entry to send to Datadog.
type datadogLogEntry struct {
	// DDSource is the source of the log (integration name).
	DDSource string `json:"ddsource"`
	// DDTags is a comma-separated list of tags.
	DDTags string `json:"ddtags,omitempty"`
	// Hostname is the name of the originating host.
	Hostname string `json:"hostname,omitempty"`
	// Service is the name of the application or service generating the logs.
	Service string `json:"service,omitempty"`
	// Message is the log message.
	Message string `json:"message"`
	// Timestamp is when the event occurred (ISO8601 format).
	Timestamp string `json:"timestamp,omitempty"`
	// FeatureFlag contains the feature flag evaluation details.
	FeatureFlag datadogFeatureFlag `json:"feature_flag"`
	// User contains the user context.
	User datadogUser `json:"usr,omitempty"`
}

// datadogFeatureFlag represents feature flag data in Datadog's expected format.
type datadogFeatureFlag struct {
	// Key is the feature flag key.
	Key string `json:"key"`
	// Value is the evaluated value of the feature flag.
	Value any `json:"value"`
	// Variation is the variation name that was evaluated.
	Variation string `json:"variation"`
	// Version is the version of the flag configuration.
	Version string `json:"version,omitempty"`
	// Default indicates if the default value was returned.
	Default bool `json:"default"`
}

// datadogUser represents user context for correlation with RUM.
type datadogUser struct {
	// ID is the user identifier.
	ID string `json:"id,omitempty"`
}

// Export sends feature flag evaluation events to Datadog.
func (e *Exporter) Export(
	ctx context.Context,
	_ *fflog.FFLogger,
	events []exporter.ExportableEvent,
) error {
	e.init.Do(func() {
		if e.httpClient == nil {
			e.httpClient = internal.DefaultHTTPClient()
		}
		if e.Site == "" {
			e.Site = DatadogDefaultSite
		}
		if e.Source == "" {
			e.Source = DatadogDefaultSource
		}
		if e.Service == "" {
			e.Service = DatadogDefaultService
		}
	})

	if e.APIKey == "" {
		return fmt.Errorf("datadog exporter: API key is required")
	}

	hostname, _ := os.Hostname()

	logEntries := make([]datadogLogEntry, 0, len(events))
	for _, event := range events {
		featureEvent, ok := event.(exporter.FeatureEvent)
		if !ok {
			continue
		}
		msg := fmt.Sprintf(
			"Feature flag '%s' evaluated to '%v' for user '%s'",
			featureEvent.Key, featureEvent.Value, featureEvent.UserKey)
		entry := datadogLogEntry{
			DDSource:  e.Source,
			DDTags:    e.buildTags(),
			Hostname:  hostname,
			Service:   e.Service,
			Message:   msg,
			Timestamp: time.Unix(featureEvent.CreationDate, 0).Format(time.RFC3339),
			FeatureFlag: datadogFeatureFlag{
				Key:       featureEvent.Key,
				Value:     featureEvent.Value,
				Variation: featureEvent.Variation,
				Version:   featureEvent.Version,
				Default:   featureEvent.Default,
			},
			User: datadogUser{
				ID: featureEvent.UserKey,
			},
		}
		logEntries = append(logEntries, entry)
	}

	if len(logEntries) == 0 {
		return nil
	}

	payload, err := json.Marshal(logEntries)
	if err != nil {
		return fmt.Errorf("datadog exporter: failed to marshal events: %w", err)
	}

	endpoint := fmt.Sprintf("https://http-intake.logs.%s/api/v2/logs", e.Site)
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, io.NopCloser(bytes.NewReader(payload)))
	if err != nil {
		return fmt.Errorf("datadog exporter: failed to create request: %w", err)
	}

	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("DD-API-KEY", e.APIKey)

	response, err := e.httpClient.Do(request)
	if err != nil {
		return fmt.Errorf("datadog exporter: failed to send request: %w", err)
	}
	defer func() { _ = response.Body.Close() }()

	if response.StatusCode >= 400 {
		body, _ := io.ReadAll(response.Body)
		return fmt.Errorf("datadog exporter: received HTTP %d: %s", response.StatusCode, string(body))
	}

	return nil
}

// buildTags creates a comma-separated string of tags.
func (e *Exporter) buildTags() string {
	if len(e.Tags) == 0 {
		return ""
	}
	result := ""
	for i, tag := range e.Tags {
		if i > 0 {
			result += ","
		}
		result += tag
	}
	return result
}

// IsBulk returns true because this exporter sends events in bulk.
func (e *Exporter) IsBulk() bool {
	return true
}
