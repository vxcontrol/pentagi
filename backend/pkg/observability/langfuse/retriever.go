package langfuse

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"pentagi/pkg/observability/langfuse/api"

	"github.com/vxcontrol/langchaingo/llms"
)

const (
	retrieverDefaultName = "Default Retriever"
)

type Retriever interface {
	End(opts ...RetrieverOption)
	String() string
	MarshalJSON() ([]byte, error)
	Observation(ctx context.Context) (context.Context, Observation)
	ObservationInfo() ObservationInfo
}

type retriever struct {
	Name            string           `json:"name"`
	Metadata        Metadata         `json:"metadata,omitempty"`
	Input           any              `json:"input,omitempty"`
	Output          any              `json:"output,omitempty"`
	StartTime       *time.Time       `json:"start_time,omitempty"`
	EndTime         *time.Time       `json:"end_time,omitempty"`
	Level           ObservationLevel `json:"level"`
	Status          *string          `json:"status,omitempty"`
	Version         *string          `json:"version,omitempty"`
	Model           *string          `json:"model,omitempty"`
	ModelParameters *ModelParameters `json:"modelParameters,omitempty" url:"modelParameters,omitempty"`
	Tools           []llms.Tool      `json:"tools,omitempty"`

	TraceID             string `json:"trace_id"`
	ObservationID       string `json:"observation_id"`
	ParentObservationID string `json:"parent_observation_id"`

	observer enqueue `json:"-"`
}

type RetrieverOption func(*retriever)

func withRetrieverTraceID(traceID string) RetrieverOption {
	return func(r *retriever) {
		r.TraceID = traceID
	}
}

func withRetrieverParentObservationID(parentObservationID string) RetrieverOption {
	return func(r *retriever) {
		r.ParentObservationID = parentObservationID
	}
}

// WithRetrieverID sets on creation time
func WithRetrieverID(id string) RetrieverOption {
	return func(r *retriever) {
		r.ObservationID = id
	}
}

func WithRetrieverName(name string) RetrieverOption {
	return func(r *retriever) {
		r.Name = name
	}
}

func WithRetrieverMetadata(metadata Metadata) RetrieverOption {
	return func(r *retriever) {
		r.Metadata = mergeMaps(r.Metadata, metadata)
	}
}

// WithRetrieverInput sets on creation time
func WithRetrieverInput(input any) RetrieverOption {
	return func(r *retriever) {
		r.Input = input
	}
}

func WithRetrieverOutput(output any) RetrieverOption {
	return func(r *retriever) {
		r.Output = output
	}
}

func WithRetrieverStartTime(time time.Time) RetrieverOption {
	return func(r *retriever) {
		r.StartTime = &time
	}
}

func WithRetrieverEndTime(time time.Time) RetrieverOption {
	return func(r *retriever) {
		r.EndTime = &time
	}
}

func WithRetrieverLevel(level ObservationLevel) RetrieverOption {
	return func(r *retriever) {
		r.Level = level
	}
}

func WithRetrieverStatus(status string) RetrieverOption {
	return func(r *retriever) {
		r.Status = &status
	}
}

func WithRetrieverVersion(version string) RetrieverOption {
	return func(r *retriever) {
		r.Version = &version
	}
}

func WithRetrieverModel(model string) RetrieverOption {
	return func(r *retriever) {
		r.Model = &model
	}
}

func WithRetrieverModelParameters(parameters *ModelParameters) RetrieverOption {
	return func(r *retriever) {
		r.ModelParameters = parameters
	}
}

func WithRetrieverTools(tools []llms.Tool) RetrieverOption {
	return func(r *retriever) {
		r.Tools = tools
	}
}

func newRetriever(observer enqueue, opts ...RetrieverOption) Retriever {
	r := &retriever{
		Name:          retrieverDefaultName,
		ObservationID: newSpanID(),
		Version:       getStringRef(firstVersion),
		StartTime:     getCurrentTimeRef(),
		observer:      observer,
	}

	for _, opt := range opts {
		opt(r)
	}

	obsCreate := &api.IngestionEvent{IngestionEventThirteen: &api.IngestionEventThirteen{
		ID:        newSpanID(),
		Timestamp: getTimeRefString(r.StartTime),
		Type:      api.IngestionEventThirteenType(ingestionCreateRetriever).Ptr(),
		Body: &api.CreateGenerationBody{
			ID:                  getStringRef(r.ObservationID),
			TraceID:             getStringRef(r.TraceID),
			ParentObservationID: getStringRef(r.ParentObservationID),
			Name:                getStringRef(r.Name),
			Metadata:            r.Metadata,
			Input:               convertInput(r.Input, r.Tools),
			Output:              convertOutput(r.Output),
			StartTime:           r.StartTime,
			EndTime:             r.EndTime,
			Level:               r.Level.ToLangfuse(),
			StatusMessage:       r.Status,
			Version:             r.Version,
			Model:               r.Model,
			ModelParameters:     r.ModelParameters.ToLangfuse(),
		},
	}}

	r.observer.enqueue(obsCreate)

	return r
}

func (r *retriever) End(opts ...RetrieverOption) {
	id := r.ObservationID
	startTime := r.StartTime
	r.EndTime = getCurrentTimeRef()
	for _, opt := range opts {
		opt(r)
	}

	// preserve the original observation ID and start time
	r.ObservationID = id
	r.StartTime = startTime

	retrieverUpdate := &api.IngestionEvent{IngestionEventThirteen: &api.IngestionEventThirteen{
		ID:        newSpanID(),
		Timestamp: getTimeRefString(r.EndTime),
		Type:      api.IngestionEventThirteenType(ingestionCreateRetriever).Ptr(),
		Body: &api.CreateGenerationBody{
			ID:              getStringRef(r.ObservationID),
			Name:            getStringRef(r.Name),
			Metadata:        r.Metadata,
			Input:           convertInput(r.Input, r.Tools),
			Output:          convertOutput(r.Output),
			EndTime:         r.EndTime,
			Level:           r.Level.ToLangfuse(),
			StatusMessage:   r.Status,
			Version:         r.Version,
			Model:           r.Model,
			ModelParameters: r.ModelParameters.ToLangfuse(),
		},
	}}

	r.observer.enqueue(retrieverUpdate)
}

func (r *retriever) String() string {
	return fmt.Sprintf("Trace(%s) Observation(%s) Retriever(%s)", r.TraceID, r.ObservationID, r.Name)
}

func (r *retriever) MarshalJSON() ([]byte, error) {
	return json.Marshal(r)
}

func (r *retriever) Observation(ctx context.Context) (context.Context, Observation) {
	obs := &observation{
		obsCtx: observationContext{
			TraceID:       r.TraceID,
			ObservationID: r.ObservationID,
		},
		observer: r.observer,
	}

	return putObservationContext(ctx, obs.obsCtx), obs
}

func (r *retriever) ObservationInfo() ObservationInfo {
	return ObservationInfo{
		TraceID:             r.TraceID,
		ObservationID:       r.ObservationID,
		ParentObservationID: r.ParentObservationID,
	}
}
