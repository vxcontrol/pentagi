package langfuse

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

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

	traceCtx *TraceContext `json:"-"`
	endOnce  sync.Once     `json:"-"`
	observer enqueue       `json:"-"`
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

func withEvaluatorTraceContext(traceCtx *TraceContext) EvaluatorOption {
	return func(e *evaluator) {
		e.traceCtx = cloneTraceContext(traceCtx)
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

	return e
}

func (e *evaluator) End(opts ...EvaluatorOption) {
	e.endOnce.Do(func() {
		id := e.ObservationID
		startTime := e.StartTime
		e.EndTime = getCurrentTimeRef()
		for _, opt := range opts {
			opt(e)
		}

		// preserve the original observation ID and start time
		e.ObservationID = id
		e.StartTime = startTime

		otelSpan := buildOtelSpan(otelSpanParams{
			TraceID:             e.TraceID,
			ObservationID:       e.ObservationID,
			ParentObservationID: e.ParentObservationID,
			Name:                e.Name,
			ObservationType:     otelObservationTypeEvaluator,
			StartTime:           e.StartTime,
			EndTime:             e.EndTime,
			Level:               e.Level,
			Status:              e.Status,
			Version:             e.Version,
			Input:               convertInput(e.Input, e.Tools),
			Output:              convertOutput(e.Output),
			Metadata:            e.Metadata,
			Model:               e.Model,
			ModelParameters:     e.ModelParameters,
			TraceCtx:            e.traceCtx,
		})

		e.observer.enqueue(otelSpan)
	})
}

func (e *evaluator) String() string {
	return fmt.Sprintf("Trace(%s) Observation(%s) Evaluator(%s)", e.TraceID, e.ObservationID, e.Name)
}

func (e *evaluator) MarshalJSON() ([]byte, error) {
	return json.Marshal(e)
}

func (e *evaluator) Observation(ctx context.Context) (context.Context, Observation) {
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

func (e *evaluator) ObservationInfo() ObservationInfo {
	return ObservationInfo{
		TraceID:             e.TraceID,
		ObservationID:       e.ObservationID,
		ParentObservationID: e.ParentObservationID,
	}
}
