package ffclient

import (
	"time"

	"github.com/thomaspoignant/go-feature-flag/exporter"
	"github.com/thomaspoignant/go-feature-flag/ffcontext"
)

// Track is used to track an event.
// Note: Use this function only if you are using multiple go-feature-flag instances.
func (g *GoFeatureFlag) Track(
	trackingEventName string,
	ctx ffcontext.EvaluationContext,
	trackingEventDetails exporter.TrackingEventDetails,
) {
	if g != nil && g.trackingEventDataExporter != nil {
		contextKind := "user"
		if ctx.IsAnonymous() {
			contextKind = "anonymousUser"
		}
		event := exporter.TrackingEvent{
			Kind:              "tracking",
			ContextKind:       contextKind,
			UserKey:           ctx.GetKey(),
			CreationDate:      time.Now().Unix(),
			Key:               trackingEventName,
			EvaluationContext: ctx.ToMap(),
			TrackingDetails:   trackingEventDetails,
		}
		g.trackingEventDataExporter.AddEvent(event)
	}
}

// Track is used to track an event.
func Track(
	trackingEventName string,
	ctx ffcontext.EvaluationContext,
	trackingEventDetails exporter.TrackingEventDetails,
) {
	ff.Track(trackingEventName, ctx, trackingEventDetails)
}
