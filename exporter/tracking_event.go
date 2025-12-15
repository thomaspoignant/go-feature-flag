package exporter

import (
	"bytes"
	"encoding/json"
	"text/template"
	"time"
)

type TrackingEventDetails = map[string]any

// TrackingEvent represent an Event that we store in the data storage
// nolint:lll
type TrackingEvent struct {
	// Kind for a feature event is feature.
	// A feature event will only be generated if the trackEvents attribute of the flag is set to true.
	Kind string `json:"kind" example:"feature" parquet:"name=kind, type=BYTE_ARRAY, convertedtype=UTF8"`

	// ContextKind is the kind of context which generated an event. This will only be "anonymousUser" for events generated
	// on behalf of an anonymous user or the reserved word "user" for events generated on behalf of a non-anonymous user
	ContextKind string `json:"contextKind,omitempty" example:"user" parquet:"name=contextKind, type=BYTE_ARRAY, convertedtype=UTF8"`

	// UserKey The key of the user object used in a feature flag evaluation. Details for the user object used in a feature
	// flag evaluation as reported by the "feature" event are transmitted periodically with a separate index event.
	UserKey string `json:"userKey" example:"94a25909-20d8-40cc-8500-fee99b569345" parquet:"name=userKey, type=BYTE_ARRAY, convertedtype=UTF8"`

	// CreationDate When the feature flag was requested at Unix epoch time in milliseconds.
	CreationDate int64 `json:"creationDate" example:"1680246000011" parquet:"name=creationDate, type=INT64"`

	// Key of the event.
	Key string `json:"key" example:"my-feature-flag" parquet:"name=key, type=BYTE_ARRAY, convertedtype=UTF8"`

	// EvaluationContext contains the evaluation context used for the tracking
	EvaluationContext map[string]any `json:"evaluationContext" parquet:"name=evaluationContext, type=MAP, keytype=BYTE_ARRAY, keyconvertedtype=UTF8, valuetype=BYTE_ARRAY, valueconvertedtype=UTF8"`

	// TrackingDetails contains the details of the tracking event
	TrackingDetails TrackingEventDetails `json:"trackingEventDetails" parquet:"name=trackingEventDetails, type=MAP, keytype=BYTE_ARRAY, keyconvertedtype=UTF8, valuetype=BYTE_ARRAY, valueconvertedtype=UTF8"`
}

func (f TrackingEvent) GetKey() string {
	return f.Key
}

// GetUserKey returns the user key of the event
func (f TrackingEvent) GetUserKey() string {
	return f.UserKey
}

// GetCreationDate returns the creationDate of the event.
func (f TrackingEvent) GetCreationDate() int64 {
	return f.CreationDate
}

func (f TrackingEvent) FormatInCSV(csvTemplate *template.Template) ([]byte, error) {
	var buf bytes.Buffer
	err := csvTemplate.Execute(&buf, struct {
		TrackingEvent
		FormattedDate string
	}{
		TrackingEvent: f,
		FormattedDate: time.Unix(f.GetCreationDate(), 0).Format(time.RFC3339),
	})
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (f TrackingEvent) FormatInJSON() ([]byte, error) {
	b, err := json.Marshal(f)
	b = append(b, []byte("\n")...)
	return b, err
}
