package langfuse

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"pentagi/pkg/observability/langfuse/api"
)

const (
	spanDefaultName = "Default Span"
)

type Span interface {
	End(opts ...SpanEndOption)
	String() string
	MarshalJSON() ([]byte, error)
	Observation(ctx context.Context) (context.Context, Observation)
	ObservationInfo() ObservationInfo
}

type span struct {
	Name      string           `json:"name"`
	Metadata  Metadata         `json:"metadata,omitempty"`
	Input     any              `json:"input,omitempty"`
	Output    any              `json:"output,omitempty"`
	StartTime *time.Time       `json:"start_time,omitempty"`
	EndTime   *time.Time       `json:"end_time,omitempty"`
	Level     ObservationLevel `json:"level"`
	Status    *string          `json:"status,omitempty"`
	Version   *string          `json:"version,omitempty"`

	TraceID             string `json:"trace_id"`
	ObservationID       string `json:"observation_id"`
	ParentObservationID string `json:"parent_observation_id"`

	observer enqueue `json:"-"`
}

type SpanStartOption func(*span)

func withSpanTraceID(traceID string) SpanStartOption {
	return func(s *span) {
		s.TraceID = traceID
	}
}

func withSpanParentObservationID(parentObservationID string) SpanStartOption {
	return func(s *span) {
		s.ParentObservationID = parentObservationID
	}
}

func WithStartSpanID(id string) SpanStartOption {
	return func(s *span) {
		s.ObservationID = id
	}
}

func WithStartSpanName(name string) SpanStartOption {
	return func(s *span) {
		s.Name = name
	}
}

func WithStartSpanMetadata(metadata Metadata) SpanStartOption {
	return func(s *span) {
		s.Metadata = metadata
	}
}

func WithStartSpanInput(input any) SpanStartOption {
	return func(s *span) {
		s.Input = input
	}
}

func WithStartSpanTime(time time.Time) SpanStartOption {
	return func(s *span) {
		s.StartTime = &time
	}
}

func WithStartSpanLevel(level ObservationLevel) SpanStartOption {
	return func(s *span) {
		s.Level = level
	}
}

func WithStartSpanVersion(version string) SpanStartOption {
	return func(s *span) {
		s.Version = &version
	}
}

type SpanEndOption func(*span)

func WithEndSpanTime(time time.Time) SpanEndOption {
	return func(s *span) {
		s.EndTime = &time
	}
}

func WithEndSpanOutput(output any) SpanEndOption {
	return func(s *span) {
		s.Output = output
	}
}

func WithEndSpanStatus(status string) SpanEndOption {
	return func(s *span) {
		s.Status = &status
	}
}

func WithEndSpanLevel(level ObservationLevel) SpanEndOption {
	return func(s *span) {
		s.Level = level
	}
}

func newSpan(observer enqueue, opts ...SpanStartOption) Span {
	s := &span{
		Name:          spanDefaultName,
		ObservationID: newSpanID(),
		Version:       getStringRef(firstVersion),
		StartTime:     getCurrentTimeRef(),
		observer:      observer,
	}

	for _, opt := range opts {
		opt(s)
	}

	obsCreate := api.NewIngestionEventFromIngestionEventEight(&api.IngestionEventEight{
		Id:        newSpanID(),
		Timestamp: getTimeRefString(s.StartTime),
		Type:      ingestionCreateObservation,
		Body: &api.ObservationBody{
			Id:                  getStringRef(s.ObservationID),
			TraceId:             getStringRef(s.TraceID),
			ParentObservationId: getStringRef(s.ParentObservationID),
			Type:                api.ObservationTypeSpan,
			Name:                getStringRef(s.Name),
			StartTime:           s.StartTime,
			Metadata:            s.Metadata,
			Input:               s.Input,
			Level:               s.Level.ToLangfuse(),
			StatusMessage:       s.Status,
			Version:             s.Version,
		},
	})

	s.observer.enqueue(obsCreate)

	return s
}

func (s *span) End(opts ...SpanEndOption) {
	s.EndTime = getCurrentTimeRef()
	for _, opt := range opts {
		opt(s)
	}

	obsUpdate := api.NewIngestionEventFromIngestionEventNine(&api.IngestionEventNine{
		Id:        newSpanID(),
		Timestamp: getTimeRefString(s.EndTime),
		Type:      ingestionUpdateObservation,
		Body: &api.ObservationBody{
			Id:            getStringRef(s.ObservationID),
			Type:          api.ObservationTypeSpan,
			EndTime:       s.EndTime,
			Output:        s.Output,
			StatusMessage: s.Status,
			Level:         s.Level.ToLangfuse(),
		},
	})

	s.observer.enqueue(obsUpdate)
}

func (s *span) String() string {
	return fmt.Sprintf("Trace(%s) Observation(%s) Span(%s)", s.TraceID, s.ObservationID, s.Name)
}

func (s *span) MarshalJSON() ([]byte, error) {
	return json.Marshal(s)
}

func (s *span) Observation(ctx context.Context) (context.Context, Observation) {
	obs := &observation{
		obsCtx: observationContext{
			TraceID:       s.TraceID,
			ObservationID: s.ObservationID,
		},
		observer: s.observer,
	}

	return putObservationContext(ctx, obs.obsCtx), obs
}

func (s *span) ObservationInfo() ObservationInfo {
	return ObservationInfo{
		TraceID:             s.TraceID,
		ObservationID:       s.ObservationID,
		ParentObservationID: s.ParentObservationID,
	}
}
