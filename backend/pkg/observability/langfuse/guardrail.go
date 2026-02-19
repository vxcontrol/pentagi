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
	guardrailDefaultName = "Default Guardrail"
)

type Guardrail interface {
	End(opts ...GuardrailOption)
	String() string
	MarshalJSON() ([]byte, error)
	Observation(ctx context.Context) (context.Context, Observation)
	ObservationInfo() ObservationInfo
}

type guardrail struct {
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

type GuardrailOption func(*guardrail)

func withGuardrailTraceID(traceID string) GuardrailOption {
	return func(g *guardrail) {
		g.TraceID = traceID
	}
}

func withGuardrailParentObservationID(parentObservationID string) GuardrailOption {
	return func(g *guardrail) {
		g.ParentObservationID = parentObservationID
	}
}

// WithGuardrailID sets on creation time
func WithGuardrailID(id string) GuardrailOption {
	return func(g *guardrail) {
		g.ObservationID = id
	}
}

func WithGuardrailName(name string) GuardrailOption {
	return func(g *guardrail) {
		g.Name = name
	}
}

func WithGuardrailMetadata(metadata Metadata) GuardrailOption {
	return func(g *guardrail) {
		g.Metadata = mergeMaps(g.Metadata, metadata)
	}
}

func WithGuardrailInput(input any) GuardrailOption {
	return func(g *guardrail) {
		g.Input = input
	}
}

func WithGuardrailOutput(output any) GuardrailOption {
	return func(g *guardrail) {
		g.Output = output
	}
}

// WithGuardrailStartTime sets on creation time
func WithGuardrailStartTime(time time.Time) GuardrailOption {
	return func(g *guardrail) {
		g.StartTime = &time
	}
}

func WithGuardrailEndTime(time time.Time) GuardrailOption {
	return func(g *guardrail) {
		g.EndTime = &time
	}
}

func WithGuardrailLevel(level ObservationLevel) GuardrailOption {
	return func(g *guardrail) {
		g.Level = level
	}
}

func WithGuardrailStatus(status string) GuardrailOption {
	return func(g *guardrail) {
		g.Status = &status
	}
}

func WithGuardrailVersion(version string) GuardrailOption {
	return func(g *guardrail) {
		g.Version = &version
	}
}

func WithGuardrailModel(model string) GuardrailOption {
	return func(g *guardrail) {
		g.Model = &model
	}
}

func WithGuardrailModelParameters(parameters *ModelParameters) GuardrailOption {
	return func(g *guardrail) {
		g.ModelParameters = parameters
	}
}

func WithGuardrailTools(tools []llms.Tool) GuardrailOption {
	return func(g *guardrail) {
		g.Tools = tools
	}
}

func newGuardrail(observer enqueue, opts ...GuardrailOption) Guardrail {
	g := &guardrail{
		Name:          guardrailDefaultName,
		ObservationID: newSpanID(),
		Version:       getStringRef(firstVersion),
		StartTime:     getCurrentTimeRef(),
		observer:      observer,
	}

	for _, opt := range opts {
		opt(g)
	}

	obsCreate := &api.IngestionEvent{IngestionEventSixteen: &api.IngestionEventSixteen{
		ID:        newSpanID(),
		Timestamp: getTimeRefString(g.StartTime),
		Type:      api.IngestionEventSixteenType(ingestionCreateGuardrail).Ptr(),
		Body: &api.CreateGenerationBody{
			ID:                  getStringRef(g.ObservationID),
			TraceID:             getStringRef(g.TraceID),
			ParentObservationID: getStringRef(g.ParentObservationID),
			Name:                getStringRef(g.Name),
			Metadata:            g.Metadata,
			Input:               convertInput(g.Input, g.Tools),
			Output:              convertOutput(g.Output),
			StartTime:           g.StartTime,
			EndTime:             g.EndTime,
			Level:               g.Level.ToLangfuse(),
			StatusMessage:       g.Status,
			Version:             g.Version,
			Model:               g.Model,
			ModelParameters:     g.ModelParameters.ToLangfuse(),
		},
	}}

	g.observer.enqueue(obsCreate)

	return g
}

func (g *guardrail) End(opts ...GuardrailOption) {
	id := g.ObservationID
	startTime := g.StartTime
	g.EndTime = getCurrentTimeRef()
	for _, opt := range opts {
		opt(g)
	}

	// preserve the original observation ID and start time
	g.ObservationID = id
	g.StartTime = startTime

	guardrailUpdate := &api.IngestionEvent{IngestionEventSixteen: &api.IngestionEventSixteen{
		ID:        newSpanID(),
		Timestamp: getTimeRefString(g.EndTime),
		Type:      api.IngestionEventSixteenType(ingestionCreateGuardrail).Ptr(),
		Body: &api.CreateGenerationBody{
			ID:              getStringRef(g.ObservationID),
			Name:            getStringRef(g.Name),
			Metadata:        g.Metadata,
			Input:           convertInput(g.Input, g.Tools),
			Output:          convertOutput(g.Output),
			EndTime:         g.EndTime,
			Level:           g.Level.ToLangfuse(),
			StatusMessage:   g.Status,
			Version:         g.Version,
			Model:           g.Model,
			ModelParameters: g.ModelParameters.ToLangfuse(),
		},
	}}

	g.observer.enqueue(guardrailUpdate)
}

func (g *guardrail) String() string {
	return fmt.Sprintf("Trace(%s) Observation(%s) Guardrail(%s)", g.TraceID, g.ObservationID, g.Name)
}

func (g *guardrail) MarshalJSON() ([]byte, error) {
	return json.Marshal(g)
}

func (g *guardrail) Observation(ctx context.Context) (context.Context, Observation) {
	obs := &observation{
		obsCtx: observationContext{
			TraceID:       g.TraceID,
			ObservationID: g.ObservationID,
		},
		observer: g.observer,
	}

	return putObservationContext(ctx, obs.obsCtx), obs
}

func (g *guardrail) ObservationInfo() ObservationInfo {
	return ObservationInfo{
		TraceID:             g.TraceID,
		ObservationID:       g.ObservationID,
		ParentObservationID: g.ParentObservationID,
	}
}
