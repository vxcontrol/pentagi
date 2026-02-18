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
	generationDefaultName = "Default Generation"
)

type Generation interface {
	End(opts ...GenerationOption)
	String() string
	MarshalJSON() ([]byte, error)
	Observation(ctx context.Context) (context.Context, Observation)
	ObservationInfo() ObservationInfo
}

type generation struct {
	Name                string           `json:"name"`
	Metadata            Metadata         `json:"metadata,omitempty"`
	Input               any              `json:"input,omitempty"`
	Output              any              `json:"output,omitempty"`
	StartTime           *time.Time       `json:"start_time,omitempty"`
	EndTime             *time.Time       `json:"end_time,omitempty"`
	CompletionStartTime *time.Time       `json:"completion_start_time,omitempty"`
	Level               ObservationLevel `json:"level"`
	Status              *string          `json:"status,omitempty"`
	Version             *string          `json:"version,omitempty"`
	Model               *string          `json:"model,omitempty"`
	ModelParameters     *ModelParameters `json:"modelParameters,omitempty" url:"modelParameters,omitempty"`
	Usage               *GenerationUsage `json:"usage,omitempty" url:"usage,omitempty"`
	PromptName          *string          `json:"promptName,omitempty" url:"promptName,omitempty"`
	PromptVersion       *int             `json:"promptVersion,omitempty" url:"promptVersion,omitempty"`
	Tools               []llms.Tool      `json:"tools,omitempty"`

	TraceID             string `json:"trace_id"`
	ObservationID       string `json:"observation_id"`
	ParentObservationID string `json:"parent_observation_id"`

	observer enqueue `json:"-"`
}

type GenerationOption func(*generation)

func withGenerationTraceID(traceID string) GenerationOption {
	return func(g *generation) {
		g.TraceID = traceID
	}
}

func withGenerationParentObservationID(parentObservationID string) GenerationOption {
	return func(g *generation) {
		g.ParentObservationID = parentObservationID
	}
}

// WithGenerationID sets on creation time
func WithGenerationID(id string) GenerationOption {
	return func(g *generation) {
		g.ObservationID = id
	}
}

func WithGenerationName(name string) GenerationOption {
	return func(g *generation) {
		g.Name = name
	}
}

func WithGenerationMetadata(metadata Metadata) GenerationOption {
	return func(g *generation) {
		g.Metadata = mergeMaps(g.Metadata, metadata)
	}
}

func WithGenerationInput(input any) GenerationOption {
	return func(g *generation) {
		g.Input = input
	}
}

func WithGenerationOutput(output any) GenerationOption {
	return func(g *generation) {
		g.Output = output
	}
}

// WithGenerationStartTime sets on creation time
func WithGenerationStartTime(time time.Time) GenerationOption {
	return func(g *generation) {
		g.StartTime = &time
	}
}

func WithGenerationEndTime(time time.Time) GenerationOption {
	return func(g *generation) {
		g.EndTime = &time
	}
}

func WithGenerationCompletionStartTime(time time.Time) GenerationOption {
	return func(g *generation) {
		g.CompletionStartTime = &time
	}
}

func WithGenerationLevel(level ObservationLevel) GenerationOption {
	return func(g *generation) {
		g.Level = level
	}
}

func WithGenerationStatus(status string) GenerationOption {
	return func(g *generation) {
		g.Status = &status
	}
}

func WithGenerationVersion(version string) GenerationOption {
	return func(g *generation) {
		g.Version = &version
	}
}

func WithGenerationModel(model string) GenerationOption {
	return func(g *generation) {
		g.Model = &model
	}
}

func WithGenerationModelParameters(parameters *ModelParameters) GenerationOption {
	return func(g *generation) {
		g.ModelParameters = parameters
	}
}

func WithGenerationUsage(usage *GenerationUsage) GenerationOption {
	return func(g *generation) {
		g.Usage = usage
	}
}

func WithGenerationPromptName(name string) GenerationOption {
	return func(g *generation) {
		g.PromptName = &name
	}
}

func WithGenerationPromptVersion(version int) GenerationOption {
	return func(g *generation) {
		g.PromptVersion = &version
	}
}

func WithGenerationTools(tools []llms.Tool) GenerationOption {
	return func(g *generation) {
		g.Tools = tools
	}
}

func newGeneration(observer enqueue, opts ...GenerationOption) Generation {
	currentTime := getCurrentTimeRef()
	g := &generation{
		Name:                generationDefaultName,
		ObservationID:       newSpanID(),
		Version:             getStringRef(firstVersion),
		StartTime:           currentTime,
		CompletionStartTime: currentTime,
		observer:            observer,
	}

	for _, opt := range opts {
		opt(g)
	}

	if g.StartTime != currentTime && g.CompletionStartTime == currentTime {
		g.CompletionStartTime = g.StartTime
	}

	genCreate := &api.IngestionEvent{IngestionEventFour: &api.IngestionEventFour{
		ID:        newSpanID(),
		Timestamp: getTimeRefString(g.StartTime),
		Type:      api.IngestionEventFourType(ingestionCreateGeneration).Ptr(),
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
			CompletionStartTime: g.CompletionStartTime,
			Level:               g.Level.ToLangfuse(),
			StatusMessage:       g.Status,
			Version:             g.Version,
			Model:               g.Model,
			ModelParameters:     g.ModelParameters.ToLangfuse(),
			PromptName:          g.PromptName,
			PromptVersion:       g.PromptVersion,
			Usage:               g.Usage.ToLangfuse(),
		},
	}}

	g.observer.enqueue(genCreate)

	return g
}

func (g *generation) End(opts ...GenerationOption) {
	id := g.ObservationID
	startTime := g.StartTime
	g.EndTime = getCurrentTimeRef()
	for _, opt := range opts {
		opt(g)
	}

	// preserve the original observation ID and start time
	g.ObservationID = id
	g.StartTime = startTime

	genUpdate := &api.IngestionEvent{IngestionEventFive: &api.IngestionEventFive{
		ID:        newSpanID(),
		Timestamp: getTimeRefString(g.EndTime),
		Type:      api.IngestionEventFiveType(ingestionUpdateGeneration).Ptr(),
		Body: &api.UpdateGenerationBody{
			ID:                  g.ObservationID,
			Name:                getStringRef(g.Name),
			Metadata:            g.Metadata,
			Input:               convertInput(g.Input, g.Tools),
			Output:              convertOutput(g.Output),
			EndTime:             g.EndTime,
			CompletionStartTime: g.CompletionStartTime,
			Level:               g.Level.ToLangfuse(),
			StatusMessage:       g.Status,
			Version:             g.Version,
			Model:               g.Model,
			ModelParameters:     g.ModelParameters.ToLangfuse(),
			PromptName:          g.PromptName,
			PromptVersion:       g.PromptVersion,
			Usage:               g.Usage.ToLangfuse(),
		},
	}}

	g.observer.enqueue(genUpdate)
}

func (g *generation) String() string {
	return fmt.Sprintf("Trace(%s) Observation(%s) Generation(%s)", g.TraceID, g.ObservationID, g.Name)
}

func (g *generation) MarshalJSON() ([]byte, error) {
	return json.Marshal(g)
}

func (g *generation) Observation(ctx context.Context) (context.Context, Observation) {
	obs := &observation{
		obsCtx: observationContext{
			TraceID:       g.TraceID,
			ObservationID: g.ObservationID,
		},
		observer: g.observer,
	}

	return putObservationContext(ctx, obs.obsCtx), obs
}

func (g *generation) ObservationInfo() ObservationInfo {
	return ObservationInfo{
		TraceID:             g.TraceID,
		ObservationID:       g.ObservationID,
		ParentObservationID: g.ParentObservationID,
	}
}
