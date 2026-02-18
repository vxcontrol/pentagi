package langfuse

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"pentagi/pkg/observability/langfuse/api"
)

const (
	toolDefaultName = "Default Tool"
)

type Tool interface {
	End(opts ...ToolOption)
	String() string
	MarshalJSON() ([]byte, error)
	Observation(ctx context.Context) (context.Context, Observation)
	ObservationInfo() ObservationInfo
}

type tool struct {
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

type ToolOption func(*tool)

func withToolTraceID(traceID string) ToolOption {
	return func(t *tool) {
		t.TraceID = traceID
	}
}

func withToolParentObservationID(parentObservationID string) ToolOption {
	return func(t *tool) {
		t.ParentObservationID = parentObservationID
	}
}

// WithToolID sets on creation time
func WithToolID(id string) ToolOption {
	return func(t *tool) {
		t.ObservationID = id
	}
}

func WithToolName(name string) ToolOption {
	return func(t *tool) {
		t.Name = name
	}
}

func WithToolMetadata(metadata Metadata) ToolOption {
	return func(t *tool) {
		t.Metadata = mergeMaps(t.Metadata, metadata)
	}
}

func WithToolInput(input any) ToolOption {
	return func(t *tool) {
		t.Input = input
	}
}

func WithToolOutput(output any) ToolOption {
	return func(t *tool) {
		t.Output = output
	}
}

// WithToolStartTime sets on creation time
func WithToolStartTime(time time.Time) ToolOption {
	return func(t *tool) {
		t.StartTime = &time
	}
}

func WithToolEndTime(time time.Time) ToolOption {
	return func(t *tool) {
		t.EndTime = &time
	}
}

func WithToolLevel(level ObservationLevel) ToolOption {
	return func(t *tool) {
		t.Level = level
	}
}

func WithToolStatus(status string) ToolOption {
	return func(t *tool) {
		t.Status = &status
	}
}

func WithToolVersion(version string) ToolOption {
	return func(t *tool) {
		t.Version = &version
	}
}

func newTool(observer enqueue, opts ...ToolOption) Tool {
	t := &tool{
		Name:          toolDefaultName,
		ObservationID: newSpanID(),
		Version:       getStringRef(firstVersion),
		StartTime:     getCurrentTimeRef(),
		observer:      observer,
	}

	for _, opt := range opts {
		opt(t)
	}

	obsCreate := &api.IngestionEvent{IngestionEventEleven: &api.IngestionEventEleven{
		ID:        newSpanID(),
		Timestamp: getTimeRefString(t.StartTime),
		Type:      api.IngestionEventElevenType(ingestionCreateTool).Ptr(),
		Body: &api.CreateGenerationBody{
			ID:                  getStringRef(t.ObservationID),
			TraceID:             getStringRef(t.TraceID),
			ParentObservationID: getStringRef(t.ParentObservationID),
			Name:                getStringRef(t.Name),
			Metadata:            t.Metadata,
			Input:               convertInput(t.Input, nil),
			Output:              convertOutput(t.Output),
			StartTime:           t.StartTime,
			EndTime:             t.EndTime,
			Level:               t.Level.ToLangfuse(),
			StatusMessage:       t.Status,
			Version:             t.Version,
		},
	}}

	t.observer.enqueue(obsCreate)

	return t
}

func (t *tool) End(opts ...ToolOption) {
	id := t.ObservationID
	startTime := t.StartTime
	t.EndTime = getCurrentTimeRef()
	for _, opt := range opts {
		opt(t)
	}

	// preserve the original observation ID and start time
	t.ObservationID = id
	t.StartTime = startTime

	toolUpdate := &api.IngestionEvent{IngestionEventEleven: &api.IngestionEventEleven{
		ID:        newSpanID(),
		Timestamp: getTimeRefString(t.EndTime),
		Type:      api.IngestionEventElevenType(ingestionCreateTool).Ptr(),
		Body: &api.CreateGenerationBody{
			ID:            getStringRef(t.ObservationID),
			Name:          getStringRef(t.Name),
			Metadata:      t.Metadata,
			Input:         convertInput(t.Input, nil),
			Output:        convertOutput(t.Output),
			EndTime:       t.EndTime,
			Level:         t.Level.ToLangfuse(),
			StatusMessage: t.Status,
			Version:       t.Version,
		},
	}}

	t.observer.enqueue(toolUpdate)
}

func (t *tool) String() string {
	return fmt.Sprintf("Trace(%s) Observation(%s) Tool(%s)", t.TraceID, t.ObservationID, t.Name)
}

func (t *tool) MarshalJSON() ([]byte, error) {
	return json.Marshal(t)
}

func (t *tool) Observation(ctx context.Context) (context.Context, Observation) {
	obs := &observation{
		obsCtx: observationContext{
			TraceID:       t.TraceID,
			ObservationID: t.ObservationID,
		},
		observer: t.observer,
	}

	return putObservationContext(ctx, obs.obsCtx), obs
}

func (t *tool) ObservationInfo() ObservationInfo {
	return ObservationInfo{
		TraceID:             t.TraceID,
		ObservationID:       t.ObservationID,
		ParentObservationID: t.ParentObservationID,
	}
}
