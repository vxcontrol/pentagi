package langfuse

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"pentagi/pkg/observability/langfuse/api"
)

const (
	generationDefaultName = "Default Generation"
)

type Generation interface {
	End(opts ...GenerationEndOption)
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

	TraceID             string `json:"trace_id"`
	ObservationID       string `json:"observation_id"`
	ParentObservationID string `json:"parent_observation_id"`

	observer enqueue `json:"-"`
}

type GenerationStartOption func(*generation)

func withGenerationTraceID(traceID string) GenerationStartOption {
	return func(g *generation) {
		g.TraceID = traceID
	}
}

func withGenerationParentObservationID(parentObservationID string) GenerationStartOption {
	return func(g *generation) {
		g.ParentObservationID = parentObservationID
	}
}

func WithStartGenerationID(id string) GenerationStartOption {
	return func(g *generation) {
		g.ObservationID = id
	}
}

func WithStartGenerationName(name string) GenerationStartOption {
	return func(g *generation) {
		g.Name = name
	}
}

func WithStartGenerationMetadata(metadata Metadata) GenerationStartOption {
	return func(g *generation) {
		g.Metadata = metadata
	}
}

func WithStartGenerationInput(input any) GenerationStartOption {
	return func(g *generation) {
		g.Input = input
	}
}

func WithStartGenerationTime(time time.Time) GenerationStartOption {
	return func(g *generation) {
		g.StartTime = &time
	}
}

func WithStartGenerationCompletionStartTime(time time.Time) GenerationStartOption {
	return func(g *generation) {
		g.CompletionStartTime = &time
	}
}

func WithStartGenerationModel(model string) GenerationStartOption {
	return func(g *generation) {
		g.Model = &model
	}
}

func WithStartGenerationModelParameters(parameters *ModelParameters) GenerationStartOption {
	return func(g *generation) {
		g.ModelParameters = parameters
	}
}

func WithStartGenerationLevel(level ObservationLevel) GenerationStartOption {
	return func(g *generation) {
		g.Level = level
	}
}

func WithStartGenerationVersion(version string) GenerationStartOption {
	return func(g *generation) {
		g.Version = &version
	}
}

func WithStartGenerationPromptName(name string) GenerationStartOption {
	return func(g *generation) {
		g.PromptName = &name
	}
}

func WithStartGenerationPromptVersion(version int) GenerationStartOption {
	return func(g *generation) {
		g.PromptVersion = &version
	}
}

type GenerationEndOption func(*generation)

func WithEndGenerationTime(time time.Time) GenerationEndOption {
	return func(g *generation) {
		g.EndTime = &time
	}
}

func WithEndGenerationOutput(output any) GenerationEndOption {
	return func(g *generation) {
		g.Output = output
	}
}

func WithEndGenerationStatus(status string) GenerationEndOption {
	return func(g *generation) {
		g.Status = &status
	}
}

func WithEndGenerationLevel(level ObservationLevel) GenerationEndOption {
	return func(g *generation) {
		g.Level = level
	}
}

func WithEndGenerationUsage(usage *GenerationUsage) GenerationEndOption {
	return func(g *generation) {
		g.Usage = usage
	}
}

func newGeneration(observer enqueue, opts ...GenerationStartOption) Generation {
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

	genCreate := api.NewIngestionEventFromIngestionEventFour(&api.IngestionEventFour{
		Id:        newSpanID(),
		Timestamp: getTimeRefString(g.StartTime),
		Type:      ingestionCreateGeneration,
		Body: &api.CreateGenerationBody{
			Id:                  getStringRef(g.ObservationID),
			TraceId:             getStringRef(g.TraceID),
			ParentObservationId: getStringRef(g.ParentObservationID),
			Name:                getStringRef(g.Name),
			StartTime:           g.StartTime,
			CompletionStartTime: g.CompletionStartTime,
			Metadata:            g.Metadata,
			Input:               g.Input,
			Level:               g.Level.ToLangfuse(),
			Model:               g.Model,
			ModelParameters:     g.ModelParameters.ToLangfuse(),
			PromptName:          g.PromptName,
			PromptVersion:       g.PromptVersion,
			Version:             g.Version,
		},
	})

	g.observer.enqueue(genCreate)

	return g
}

func (g *generation) End(opts ...GenerationEndOption) {
	g.EndTime = getCurrentTimeRef()
	for _, opt := range opts {
		opt(g)
	}

	genUpdate := api.NewIngestionEventFromIngestionEventFive(&api.IngestionEventFive{
		Id:        newSpanID(),
		Timestamp: getTimeRefString(g.EndTime),
		Type:      ingestionUpdateGeneration,
		Body: &api.UpdateGenerationBody{
			Id:            g.ObservationID,
			EndTime:       g.EndTime,
			Output:        g.Output,
			StatusMessage: g.Status,
			Level:         g.Level.ToLangfuse(),
			Usage:         g.Usage.ToLangfuse(),
		},
	})

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
