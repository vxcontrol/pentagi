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
	agentDefaultName = "Default Agent"
)

type Agent interface {
	End(opts ...AgentOption)
	String() string
	MarshalJSON() ([]byte, error)
	Observation(ctx context.Context) (context.Context, Observation)
	ObservationInfo() ObservationInfo
}

type agent struct {
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

type AgentOption func(*agent)

func withAgentTraceID(traceID string) AgentOption {
	return func(a *agent) {
		a.TraceID = traceID
	}
}

func withAgentParentObservationID(parentObservationID string) AgentOption {
	return func(a *agent) {
		a.ParentObservationID = parentObservationID
	}
}

func withAgentTraceContext(traceCtx *TraceContext) AgentOption {
	return func(a *agent) {
		a.traceCtx = cloneTraceContext(traceCtx)
	}
}

// WithAgentID sets on creation time
func WithAgentID(id string) AgentOption {
	return func(a *agent) {
		a.ObservationID = id
	}
}

func WithAgentName(name string) AgentOption {
	return func(a *agent) {
		a.Name = name
	}
}

func WithAgentMetadata(metadata Metadata) AgentOption {
	return func(a *agent) {
		a.Metadata = mergeMaps(a.Metadata, metadata)
	}
}

func WithAgentInput(input any) AgentOption {
	return func(a *agent) {
		a.Input = input
	}
}

func WithAgentOutput(output any) AgentOption {
	return func(a *agent) {
		a.Output = output
	}
}

// WithAgentStartTime sets on creation time
func WithAgentStartTime(time time.Time) AgentOption {
	return func(a *agent) {
		a.StartTime = &time
	}
}

func WithAgentEndTime(time time.Time) AgentOption {
	return func(a *agent) {
		a.EndTime = &time
	}
}

func WithAgentLevel(level ObservationLevel) AgentOption {
	return func(a *agent) {
		a.Level = level
	}
}

func WithAgentStatus(status string) AgentOption {
	return func(a *agent) {
		a.Status = &status
	}
}

func WithAgentVersion(version string) AgentOption {
	return func(a *agent) {
		a.Version = &version
	}
}

func WithAgentModel(model string) AgentOption {
	return func(a *agent) {
		a.Model = &model
	}
}

func WithAgentModelParameters(parameters *ModelParameters) AgentOption {
	return func(a *agent) {
		a.ModelParameters = parameters
	}
}

func WithAgentTools(tools []llms.Tool) AgentOption {
	return func(a *agent) {
		a.Tools = tools
	}
}

func newAgent(observer enqueue, opts ...AgentOption) Agent {
	a := &agent{
		Name:          agentDefaultName,
		ObservationID: newSpanID(),
		Version:       getStringRef(firstVersion),
		StartTime:     getCurrentTimeRef(),
		observer:      observer,
	}

	for _, opt := range opts {
		opt(a)
	}

	return a
}

func (a *agent) End(opts ...AgentOption) {
	a.endOnce.Do(func() {
		id := a.ObservationID
		startTime := a.StartTime
		a.EndTime = getCurrentTimeRef()
		for _, opt := range opts {
			opt(a)
		}

		// preserve the original observation ID and start time
		a.ObservationID = id
		a.StartTime = startTime

		otelSpan := buildOtelSpan(otelSpanParams{
			TraceID:             a.TraceID,
			ObservationID:       a.ObservationID,
			ParentObservationID: a.ParentObservationID,
			Name:                a.Name,
			ObservationType:     otelObservationTypeAgent,
			StartTime:           a.StartTime,
			EndTime:             a.EndTime,
			Level:               a.Level,
			Status:              a.Status,
			Version:             a.Version,
			Input:               convertInput(a.Input, a.Tools),
			Output:              convertOutput(a.Output),
			Metadata:            a.Metadata,
			Model:               a.Model,
			ModelParameters:     a.ModelParameters,
			TraceCtx:            a.traceCtx,
		})

		a.observer.enqueue(otelSpan)
	})
}

func (a *agent) String() string {
	return fmt.Sprintf("Trace(%s) Observation(%s) Agent(%s)", a.TraceID, a.ObservationID, a.Name)
}

func (a *agent) MarshalJSON() ([]byte, error) {
	return json.Marshal(a)
}

func (a *agent) Observation(ctx context.Context) (context.Context, Observation) {
	traceCtx := a.traceCtx
	if parent, ok := getObservationContext(ctx); ok && parent.TraceCtx != nil {
		traceCtx = mergeTraceContext(parent.TraceCtx, traceCtx)
	}
	obs := &observation{
		obsCtx: observationContext{
			TraceID:       a.TraceID,
			ObservationID: a.ObservationID,
			TraceCtx:      cloneTraceContext(traceCtx),
		},
		observer: a.observer,
	}

	return putObservationContext(ctx, obs.obsCtx), obs
}

func (a *agent) ObservationInfo() ObservationInfo {
	return ObservationInfo{
		TraceID:             a.TraceID,
		ObservationID:       a.ObservationID,
		ParentObservationID: a.ParentObservationID,
	}
}
