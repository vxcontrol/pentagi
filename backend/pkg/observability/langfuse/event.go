package langfuse

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"pentagi/pkg/observability/langfuse/api"
)

const (
	eventDefaultName = "Default Event"
)

type Event interface {
	String() string
	MarshalJSON() ([]byte, error)
	Observation(ctx context.Context) (context.Context, Observation)
	ObservationInfo() ObservationInfo
}

type event struct {
	Name     string           `json:"name"`
	Metadata Metadata         `json:"metadata,omitempty"`
	Input    any              `json:"input,omitempty"`
	Output   any              `json:"output,omitempty"`
	Time     *time.Time       `json:"time"`
	Level    ObservationLevel `json:"level"`
	Status   *string          `json:"status,omitempty"`
	Version  *string          `json:"version,omitempty"`

	TraceID             string `json:"trace_id"`
	ObservationID       string `json:"observation_id"`
	ParentObservationID string `json:"parent_observation_id"`

	observer enqueue `json:"-"`
}

type EventOption func(*event)

func withEventTraceID(traceID string) EventOption {
	return func(e *event) {
		e.TraceID = traceID
	}
}

func withEventParentObservationID(parentObservationID string) EventOption {
	return func(e *event) {
		e.ParentObservationID = parentObservationID
	}
}

func WithEventName(name string) EventOption {
	return func(e *event) {
		e.Name = name
	}
}

func WithEventMetadata(metadata Metadata) EventOption {
	return func(e *event) {
		e.Metadata = metadata
	}
}

func WithEventInput(input any) EventOption {
	return func(e *event) {
		e.Input = input
	}
}

func WithEventOutput(output any) EventOption {
	return func(e *event) {
		e.Output = output
	}
}

func WithEventTime(time time.Time) EventOption {
	return func(e *event) {
		e.Time = &time
	}
}

func WithEventLevel(level ObservationLevel) EventOption {
	return func(e *event) {
		e.Level = level
	}
}

func WithEventStatus(status string) EventOption {
	return func(e *event) {
		e.Status = &status
	}
}

func WithEventVersion(version string) EventOption {
	return func(e *event) {
		e.Version = &version
	}
}

func newEvent(observer enqueue, opts ...EventOption) Event {
	currentTime := getCurrentTimeRef()
	e := &event{
		Name:          eventDefaultName,
		ObservationID: newSpanID(),
		Version:       getStringRef(firstVersion),
		Time:          currentTime,
		observer:      observer,
	}

	for _, opt := range opts {
		opt(e)
	}

	obsCreate := &api.IngestionEvent{IngestionEventSix: &api.IngestionEventSix{
		ID:        newSpanID(),
		Timestamp: getTimeRefString(e.Time),
		Type:      api.IngestionEventSixType(ingestionCreateEvent).Ptr(),
		Body: &api.CreateEventBody{
			ID:                  getStringRef(e.ObservationID),
			TraceID:             getStringRef(e.TraceID),
			ParentObservationID: getStringRef(e.ParentObservationID),
			Name:                getStringRef(e.Name),
			StartTime:           e.Time,
			Metadata:            e.Metadata,
			Input:               e.Input,
			Output:              e.Output,
			Level:               e.Level.ToLangfuse(),
			StatusMessage:       e.Status,
			Version:             e.Version,
		},
	}}

	e.observer.enqueue(obsCreate)

	return e
}

func (e *event) String() string {
	return fmt.Sprintf("Trace(%s) Observation(%s) Event(%s)", e.TraceID, e.ObservationID, e.Name)
}

func (e *event) MarshalJSON() ([]byte, error) {
	type alias event
	return json.Marshal(alias(*e))
}

func (e *event) Observation(ctx context.Context) (context.Context, Observation) {
	obs := &observation{
		obsCtx: observationContext{
			TraceID:       e.TraceID,
			ObservationID: e.ObservationID,
		},
		observer: e.observer,
	}

	return putObservationContext(ctx, obs.obsCtx), obs
}

func (e *event) ObservationInfo() ObservationInfo {
	return ObservationInfo{
		TraceID:             e.TraceID,
		ObservationID:       e.ObservationID,
		ParentObservationID: e.ParentObservationID,
	}
}
