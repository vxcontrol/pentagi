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

	traceCtx *TraceContext `json:"-"`
	endOnce  sync.Once     `json:"-"`
	observer enqueue       `json:"-"`
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

func withGenerationTraceContext(traceCtx *TraceContext) GenerationOption {
	return func(g *generation) {
		g.traceCtx = cloneTraceContext(traceCtx)
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

	return g
}

func (g *generation) End(opts ...GenerationOption) {
	g.endOnce.Do(func() {
		id := g.ObservationID
		startTime := g.StartTime
		g.EndTime = getCurrentTimeRef()
		for _, opt := range opts {
			opt(g)
		}

		// preserve the original observation ID and start time
		g.ObservationID = id
		g.StartTime = startTime

		otelSpan := buildOtelSpan(otelSpanParams{
			TraceID:             g.TraceID,
			ObservationID:       g.ObservationID,
			ParentObservationID: g.ParentObservationID,
			Name:                g.Name,
			ObservationType:     otelObservationTypeGeneration,
			StartTime:           g.StartTime,
			EndTime:             g.EndTime,
			CompletionStartTime: g.CompletionStartTime,
			Level:               g.Level,
			Status:              g.Status,
			Version:             g.Version,
			Input:               convertInput(g.Input, g.Tools),
			Output:              convertOutput(g.Output),
			Metadata:            g.Metadata,
			Model:               g.Model,
			ModelParameters:     g.ModelParameters,
			Usage:               g.Usage,
			PromptName:          g.PromptName,
			PromptVersion:       g.PromptVersion,
			TraceCtx:            g.traceCtx,
		})

		g.observer.enqueue(otelSpan)
	})
}

func (g *generation) String() string {
	return fmt.Sprintf("Trace(%s) Observation(%s) Generation(%s)", g.TraceID, g.ObservationID, g.Name)
}

func (g *generation) MarshalJSON() ([]byte, error) {
	return json.Marshal(g)
}

func (g *generation) Observation(ctx context.Context) (context.Context, Observation) {
	traceCtx := g.traceCtx
	if parent, ok := getObservationContext(ctx); ok && parent.TraceCtx != nil {
		traceCtx = mergeTraceContext(parent.TraceCtx, traceCtx)
	}
	obs := &observation{
		obsCtx: observationContext{
			TraceID:       g.TraceID,
			ObservationID: g.ObservationID,
			TraceCtx:      cloneTraceContext(traceCtx),
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
