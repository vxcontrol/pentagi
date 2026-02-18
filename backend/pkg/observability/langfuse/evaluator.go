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
	evaluatorDefaultName = "Default Evaluator"
)

type Evaluator interface {
	End(opts ...EvaluatorOption)
	String() string
	MarshalJSON() ([]byte, error)
	Observation(ctx context.Context) (context.Context, Observation)
	ObservationInfo() ObservationInfo
}

type evaluator struct {
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

type EvaluatorOption func(*evaluator)

func withEvaluatorTraceID(traceID string) EvaluatorOption {
	return func(e *evaluator) {
		e.TraceID = traceID
	}
}

func withEvaluatorParentObservationID(parentObservationID string) EvaluatorOption {
	return func(e *evaluator) {
		e.ParentObservationID = parentObservationID
	}
}

// WithEvaluatorID sets on creation time
func WithEvaluatorID(id string) EvaluatorOption {
	return func(e *evaluator) {
		e.ObservationID = id
	}
}

func WithEvaluatorName(name string) EvaluatorOption {
	return func(e *evaluator) {
		e.Name = name
	}
}

func WithEvaluatorMetadata(metadata Metadata) EvaluatorOption {
	return func(e *evaluator) {
		e.Metadata = mergeMaps(e.Metadata, metadata)
	}
}

// WithEvaluatorInput sets on creation time
func WithEvaluatorInput(input any) EvaluatorOption {
	return func(e *evaluator) {
		e.Input = input
	}
}

func WithEvaluatorOutput(output any) EvaluatorOption {
	return func(e *evaluator) {
		e.Output = output
	}
}

func WithEvaluatorStartTime(time time.Time) EvaluatorOption {
	return func(e *evaluator) {
		e.StartTime = &time
	}
}

func WithEvaluatorEndTime(time time.Time) EvaluatorOption {
	return func(e *evaluator) {
		e.EndTime = &time
	}
}

func WithEvaluatorLevel(level ObservationLevel) EvaluatorOption {
	return func(e *evaluator) {
		e.Level = level
	}
}

func WithEvaluatorStatus(status string) EvaluatorOption {
	return func(e *evaluator) {
		e.Status = &status
	}
}

func WithEvaluatorVersion(version string) EvaluatorOption {
	return func(e *evaluator) {
		e.Version = &version
	}
}

func WithEvaluatorModel(model string) EvaluatorOption {
	return func(e *evaluator) {
		e.Model = &model
	}
}

func WithEvaluatorModelParameters(parameters *ModelParameters) EvaluatorOption {
	return func(e *evaluator) {
		e.ModelParameters = parameters
	}
}

func WithEvaluatorTools(tools []llms.Tool) EvaluatorOption {
	return func(e *evaluator) {
		e.Tools = tools
	}
}

func newEvaluator(observer enqueue, opts ...EvaluatorOption) Evaluator {
	e := &evaluator{
		Name:          evaluatorDefaultName,
		ObservationID: newSpanID(),
		Version:       getStringRef(firstVersion),
		StartTime:     getCurrentTimeRef(),
		observer:      observer,
	}

	for _, opt := range opts {
		opt(e)
	}

	obsCreate := &api.IngestionEvent{IngestionEventFourteen: &api.IngestionEventFourteen{
		ID:        newSpanID(),
		Timestamp: getTimeRefString(e.StartTime),
		Type:      api.IngestionEventFourteenType(ingestionCreateEvaluator).Ptr(),
		Body: &api.CreateGenerationBody{
			ID:                  getStringRef(e.ObservationID),
			TraceID:             getStringRef(e.TraceID),
			ParentObservationID: getStringRef(e.ParentObservationID),
			Name:                getStringRef(e.Name),
			Metadata:            e.Metadata,
			Input:               convertInput(e.Input, e.Tools),
			Output:              convertOutput(e.Output),
			StartTime:           e.StartTime,
			EndTime:             e.EndTime,
			Level:               e.Level.ToLangfuse(),
			StatusMessage:       e.Status,
			Version:             e.Version,
			Model:               e.Model,
			ModelParameters:     e.ModelParameters.ToLangfuse(),
		},
	}}

	e.observer.enqueue(obsCreate)

	return e
}

func (e *evaluator) End(opts ...EvaluatorOption) {
	id := e.ObservationID
	startTime := e.StartTime
	e.EndTime = getCurrentTimeRef()
	for _, opt := range opts {
		opt(e)
	}

	// preserve the original observation ID and start time
	e.ObservationID = id
	e.StartTime = startTime

	evaluatorUpdate := &api.IngestionEvent{IngestionEventFourteen: &api.IngestionEventFourteen{
		ID:        newSpanID(),
		Timestamp: getTimeRefString(e.EndTime),
		Type:      api.IngestionEventFourteenType(ingestionCreateEvaluator).Ptr(),
		Body: &api.CreateGenerationBody{
			ID:              getStringRef(e.ObservationID),
			Name:            getStringRef(e.Name),
			Metadata:        e.Metadata,
			Input:           convertInput(e.Input, e.Tools),
			Output:          convertOutput(e.Output),
			EndTime:         e.EndTime,
			Level:           e.Level.ToLangfuse(),
			StatusMessage:   e.Status,
			Version:         e.Version,
			Model:           e.Model,
			ModelParameters: e.ModelParameters.ToLangfuse(),
		},
	}}

	e.observer.enqueue(evaluatorUpdate)
}

func (e *evaluator) String() string {
	return fmt.Sprintf("Trace(%s) Observation(%s) Evaluator(%s)", e.TraceID, e.ObservationID, e.Name)
}

func (e *evaluator) MarshalJSON() ([]byte, error) {
	return json.Marshal(e)
}

func (e *evaluator) Observation(ctx context.Context) (context.Context, Observation) {
	obs := &observation{
		obsCtx: observationContext{
			TraceID:       e.TraceID,
			ObservationID: e.ObservationID,
		},
		observer: e.observer,
	}

	return putObservationContext(ctx, obs.obsCtx), obs
}

func (e *evaluator) ObservationInfo() ObservationInfo {
	return ObservationInfo{
		TraceID:             e.TraceID,
		ObservationID:       e.ObservationID,
		ParentObservationID: e.ParentObservationID,
	}
}
