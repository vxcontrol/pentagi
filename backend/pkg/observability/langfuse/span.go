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
	End(opts ...SpanOption)
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

type SpanOption func(*span)

func withSpanTraceID(traceID string) SpanOption {
	return func(s *span) {
		s.TraceID = traceID
	}
}

func withSpanParentObservationID(parentObservationID string) SpanOption {
	return func(s *span) {
		s.ParentObservationID = parentObservationID
	}
}

// WithSpanID sets on creation time
func WithSpanID(id string) SpanOption {
	return func(s *span) {
		s.ObservationID = id
	}
}

func WithSpanName(name string) SpanOption {
	return func(s *span) {
		s.Name = name
	}
}

func WithSpanMetadata(metadata Metadata) SpanOption {
	return func(s *span) {
		s.Metadata = mergeMaps(s.Metadata, metadata)
	}
}

func WithSpanInput(input any) SpanOption {
	return func(s *span) {
		s.Input = input
	}
}

func WithSpanOutput(output any) SpanOption {
	return func(s *span) {
		s.Output = output
	}
}

// WithSpanStartTime sets on creation time
func WithSpanStartTime(time time.Time) SpanOption {
	return func(s *span) {
		s.StartTime = &time
	}
}

func WithSpanEndTime(time time.Time) SpanOption {
	return func(s *span) {
		s.EndTime = &time
	}
}

func WithSpanLevel(level ObservationLevel) SpanOption {
	return func(s *span) {
		s.Level = level
	}
}

func WithSpanStatus(status string) SpanOption {
	return func(s *span) {
		s.Status = &status
	}
}

func WithSpanVersion(version string) SpanOption {
	return func(s *span) {
		s.Version = &version
	}
}

func newSpan(observer enqueue, opts ...SpanOption) Span {
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

	obsCreate := &api.IngestionEvent{IngestionEventTwo: &api.IngestionEventTwo{
		ID:        newSpanID(),
		Timestamp: getTimeRefString(s.StartTime),
		Type:      api.IngestionEventTwoType(ingestionCreateSpan).Ptr(),
		Body: &api.CreateSpanBody{
			ID:                  getStringRef(s.ObservationID),
			TraceID:             getStringRef(s.TraceID),
			ParentObservationID: getStringRef(s.ParentObservationID),
			Name:                getStringRef(s.Name),
			Input:               convertInput(s.Input, nil),
			Output:              convertOutput(s.Output),
			StartTime:           s.StartTime,
			EndTime:             s.EndTime,
			Metadata:            s.Metadata,
			Level:               s.Level.ToLangfuse(),
			StatusMessage:       s.Status,
			Version:             s.Version,
		},
	}}

	s.observer.enqueue(obsCreate)

	return s
}

func (s *span) End(opts ...SpanOption) {
	id := s.ObservationID
	startTime := s.StartTime
	s.EndTime = getCurrentTimeRef()
	for _, opt := range opts {
		opt(s)
	}

	// preserve the original observation ID and start time
	s.ObservationID = id
	s.StartTime = startTime

	obsUpdate := &api.IngestionEvent{IngestionEventThree: &api.IngestionEventThree{
		ID:        newSpanID(),
		Timestamp: getTimeRefString(s.EndTime),
		Type:      api.IngestionEventThreeType(ingestionUpdateSpan).Ptr(),
		Body: &api.UpdateSpanBody{
			ID:            s.ObservationID,
			Name:          getStringRef(s.Name),
			Metadata:      s.Metadata,
			Input:         convertInput(s.Input, nil),
			Output:        convertOutput(s.Output),
			EndTime:       s.EndTime,
			Level:         s.Level.ToLangfuse(),
			StatusMessage: s.Status,
			Version:       s.Version,
		},
	}}

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
