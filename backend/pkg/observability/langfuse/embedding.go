package langfuse

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"pentagi/pkg/observability/langfuse/api"
)

const (
	embeddingDefaultName = "Default Embedding"
)

type Embedding interface {
	End(opts ...EmbeddingOption)
	String() string
	MarshalJSON() ([]byte, error)
	Observation(ctx context.Context) (context.Context, Observation)
	ObservationInfo() ObservationInfo
}

type embedding struct {
	Name      string           `json:"name"`
	Metadata  Metadata         `json:"metadata,omitempty"`
	Input     any              `json:"input,omitempty"`
	Output    any              `json:"output,omitempty"`
	StartTime *time.Time       `json:"start_time,omitempty"`
	EndTime   *time.Time       `json:"end_time,omitempty"`
	Level     ObservationLevel `json:"level"`
	Status    *string          `json:"status,omitempty"`
	Version   *string          `json:"version,omitempty"`
	Model     *string          `json:"model,omitempty"`

	TraceID             string `json:"trace_id"`
	ObservationID       string `json:"observation_id"`
	ParentObservationID string `json:"parent_observation_id"`

	observer enqueue `json:"-"`
}

type EmbeddingOption func(*embedding)

func withEmbeddingTraceID(traceID string) EmbeddingOption {
	return func(e *embedding) {
		e.TraceID = traceID
	}
}

func withEmbeddingParentObservationID(parentObservationID string) EmbeddingOption {
	return func(e *embedding) {
		e.ParentObservationID = parentObservationID
	}
}

// WithEmbeddingID sets on creation time
func WithEmbeddingID(id string) EmbeddingOption {
	return func(e *embedding) {
		e.ObservationID = id
	}
}

func WithEmbeddingName(name string) EmbeddingOption {
	return func(e *embedding) {
		e.Name = name
	}
}

func WithEmbeddingMetadata(metadata Metadata) EmbeddingOption {
	return func(e *embedding) {
		e.Metadata = mergeMaps(e.Metadata, metadata)
	}
}

func WithEmbeddingInput(input any) EmbeddingOption {
	return func(e *embedding) {
		e.Input = input
	}
}

func WithEmbeddingOutput(output any) EmbeddingOption {
	return func(e *embedding) {
		e.Output = output
	}
}

// WithEmbeddingStartTime sets on creation time
func WithEmbeddingStartTime(time time.Time) EmbeddingOption {
	return func(e *embedding) {
		e.StartTime = &time
	}
}

func WithEmbeddingEndTime(time time.Time) EmbeddingOption {
	return func(e *embedding) {
		e.EndTime = &time
	}
}

func WithEmbeddingLevel(level ObservationLevel) EmbeddingOption {
	return func(e *embedding) {
		e.Level = level
	}
}

func WithEmbeddingStatus(status string) EmbeddingOption {
	return func(e *embedding) {
		e.Status = &status
	}
}

func WithEmbeddingVersion(version string) EmbeddingOption {
	return func(e *embedding) {
		e.Version = &version
	}
}

func WithEmbeddingModel(model string) EmbeddingOption {
	return func(e *embedding) {
		e.Model = &model
	}
}

func newEmbedding(observer enqueue, opts ...EmbeddingOption) Embedding {
	e := &embedding{
		Name:          embeddingDefaultName,
		ObservationID: newSpanID(),
		Version:       getStringRef(firstVersion),
		StartTime:     getCurrentTimeRef(),
		observer:      observer,
	}

	for _, opt := range opts {
		opt(e)
	}

	obsCreate := &api.IngestionEvent{IngestionEventFifteen: &api.IngestionEventFifteen{
		ID:        newSpanID(),
		Timestamp: getTimeRefString(e.StartTime),
		Type:      api.IngestionEventFifteenType(ingestionCreateEmbedding).Ptr(),
		Body: &api.CreateGenerationBody{
			ID:                  getStringRef(e.ObservationID),
			TraceID:             getStringRef(e.TraceID),
			ParentObservationID: getStringRef(e.ParentObservationID),
			Name:                getStringRef(e.Name),
			Metadata:            e.Metadata,
			Input:               convertInput(e.Input, nil),
			Output:              convertOutput(e.Output),
			StartTime:           e.StartTime,
			EndTime:             e.EndTime,
			Level:               e.Level.ToLangfuse(),
			StatusMessage:       e.Status,
			Version:             e.Version,
			Model:               e.Model,
		},
	}}

	e.observer.enqueue(obsCreate)

	return e
}

func (e *embedding) End(opts ...EmbeddingOption) {
	id := e.ObservationID
	startTime := e.StartTime
	e.EndTime = getCurrentTimeRef()
	for _, opt := range opts {
		opt(e)
	}

	// preserve the original observation ID and start time
	e.ObservationID = id
	e.StartTime = startTime

	embeddingUpdate := &api.IngestionEvent{IngestionEventFifteen: &api.IngestionEventFifteen{
		ID:        newSpanID(),
		Timestamp: getTimeRefString(e.EndTime),
		Type:      api.IngestionEventFifteenType(ingestionCreateEmbedding).Ptr(),
		Body: &api.CreateGenerationBody{
			ID:            getStringRef(e.ObservationID),
			Name:          getStringRef(e.Name),
			Metadata:      e.Metadata,
			Input:         convertInput(e.Input, nil),
			Output:        convertOutput(e.Output),
			EndTime:       e.EndTime,
			Level:         e.Level.ToLangfuse(),
			StatusMessage: e.Status,
			Version:       e.Version,
			Model:         e.Model,
		},
	}}

	e.observer.enqueue(embeddingUpdate)
}

func (e *embedding) String() string {
	return fmt.Sprintf("Trace(%s) Observation(%s) Embedding(%s)", e.TraceID, e.ObservationID, e.Name)
}

func (e *embedding) MarshalJSON() ([]byte, error) {
	return json.Marshal(e)
}

func (e *embedding) Observation(ctx context.Context) (context.Context, Observation) {
	obs := &observation{
		obsCtx: observationContext{
			TraceID:       e.TraceID,
			ObservationID: e.ObservationID,
		},
		observer: e.observer,
	}

	return putObservationContext(ctx, obs.obsCtx), obs
}

func (e *embedding) ObservationInfo() ObservationInfo {
	return ObservationInfo{
		TraceID:             e.TraceID,
		ObservationID:       e.ObservationID,
		ParentObservationID: e.ParentObservationID,
	}
}
