package langfuse

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
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

	traceCtx *TraceContext `json:"-"`
	observer enqueue       `json:"-"`
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

func withEventTraceContext(traceCtx *TraceContext) EventOption {
	return func(e *event) {
		e.traceCtx = cloneTraceContext(traceCtx)
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

	otelSpan := buildOtelSpan(otelSpanParams{
		TraceID:             e.TraceID,
		ObservationID:       e.ObservationID,
		ParentObservationID: e.ParentObservationID,
		Name:                e.Name,
		ObservationType:     otelObservationTypeEvent,
		StartTime:           e.Time,
		EndTime:             e.Time,
		Level:               e.Level,
		Status:              e.Status,
		Version:             e.Version,
		Input:               convertInput(e.Input, nil),
		Output:              convertOutput(e.Output),
		Metadata:            e.Metadata,
		TraceCtx:            e.traceCtx,
	})

	e.observer.enqueue(otelSpan)

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
	traceCtx := e.traceCtx
	if parent, ok := getObservationContext(ctx); ok && parent.TraceCtx != nil {
		traceCtx = mergeTraceContext(parent.TraceCtx, traceCtx)
	}
	obs := &observation{
		obsCtx: observationContext{
			TraceID:       e.TraceID,
			ObservationID: e.ObservationID,
			TraceCtx:      cloneTraceContext(traceCtx),
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
