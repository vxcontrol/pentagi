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
	End(opts ...EventEndOption)
	String() string
	MarshalJSON() ([]byte, error)
	Observation(ctx context.Context) (context.Context, Observation)
	ObservationInfo() ObservationInfo
}

type event struct {
	Name      string           `json:"name"`
	Metadata  Metadata         `json:"metadata,omitempty"`
	Input     any              `json:"input,omitempty"`
	Output    any              `json:"output,omitempty"`
	StartTime *time.Time       `json:"start_time"`
	EndTime   *time.Time       `json:"end_time"`
	Level     ObservationLevel `json:"level"`
	Status    *string          `json:"status,omitempty"`
	Version   *string          `json:"version,omitempty"`

	TraceID             string `json:"trace_id"`
	ObservationID       string `json:"observation_id"`
	ParentObservationID string `json:"parent_observation_id"`

	observer enqueue `json:"-"`
}

type EventStartOption func(*event)

func withEventTraceID(traceID string) EventStartOption {
	return func(e *event) {
		e.TraceID = traceID
	}
}

func withEventParentObservationID(parentObservationID string) EventStartOption {
	return func(e *event) {
		e.ParentObservationID = parentObservationID
	}
}

func WithStartEventName(name string) EventStartOption {
	return func(e *event) {
		e.Name = name
	}
}

func WithStartEventMetadata(metadata Metadata) EventStartOption {
	return func(e *event) {
		e.Metadata = metadata
	}
}

func WithStartEventInput(input any) EventStartOption {
	return func(e *event) {
		e.Input = input
	}
}

func WithStartEventOutput(output any) EventStartOption {
	return func(e *event) {
		e.Output = output
	}
}

func WithStartEventStartTime(time time.Time) EventStartOption {
	return func(e *event) {
		e.StartTime = &time
	}
}

func WithStartEventEndTime(time time.Time) EventStartOption {
	return func(e *event) {
		e.EndTime = &time
	}
}

func WithStartEventLevel(level ObservationLevel) EventStartOption {
	return func(e *event) {
		e.Level = level
	}
}

func WithStartEventStatus(status string) EventStartOption {
	return func(e *event) {
		e.Status = &status
	}
}

func WithStartEventVersion(version string) EventStartOption {
	return func(e *event) {
		e.Version = &version
	}
}

type EventEndOption func(*event)

func WithEndEventTime(time time.Time) EventEndOption {
	return func(e *event) {
		e.EndTime = &time
	}
}

func newEvent(observer enqueue, opts ...EventStartOption) Event {
	currentTime := getCurrentTimeRef()
	e := &event{
		Name:          spanDefaultName,
		ObservationID: newSpanID(),
		Version:       getStringRef(firstVersion),
		StartTime:     currentTime,
		EndTime:       currentTime,
		observer:      observer,
	}

	for _, opt := range opts {
		opt(e)
	}

	if e.StartTime != currentTime && e.EndTime == currentTime {
		e.EndTime = e.StartTime
	}

	obsCreate := api.NewIngestionEventFromIngestionEventEight(&api.IngestionEventEight{
		Id:        newSpanID(),
		Timestamp: getTimeRefString(e.StartTime),
		Type:      ingestionCreateObservation,
		Body: &api.ObservationBody{
			Id:                  getStringRef(e.ObservationID),
			TraceId:             getStringRef(e.TraceID),
			ParentObservationId: getStringRef(e.ParentObservationID),
			Type:                api.ObservationTypeEvent,
			Name:                getStringRef(e.Name),
			StartTime:           e.StartTime,
			EndTime:             e.EndTime,
			Metadata:            e.Metadata,
			Input:               e.Input,
			Output:              e.Output,
			Level:               e.Level.ToLangfuse(),
			StatusMessage:       e.Status,
			Version:             e.Version,
		},
	})

	e.observer.enqueue(obsCreate)

	return e
}

func (e *event) End(opts ...EventEndOption) {
	e.EndTime = getCurrentTimeRef()
	for _, opt := range opts {
		opt(e)
	}

	obsUpdate := api.NewIngestionEventFromIngestionEventNine(&api.IngestionEventNine{
		Id:        newSpanID(),
		Timestamp: getTimeRefString(e.EndTime),
		Type:      ingestionUpdateObservation,
		Body: &api.ObservationBody{
			Id:      getStringRef(e.ObservationID),
			Type:    api.ObservationTypeEvent,
			EndTime: e.EndTime,
		},
	})

	e.observer.enqueue(obsUpdate)
}

func (e *event) String() string {
	return fmt.Sprintf("Trace(%s) Observation(%s) Event(%s)", e.TraceID, e.ObservationID, e.Name)
}

func (e *event) MarshalJSON() ([]byte, error) {
	return json.Marshal(e)
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
