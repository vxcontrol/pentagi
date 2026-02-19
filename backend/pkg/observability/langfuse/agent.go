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

	observer enqueue `json:"-"`
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

	obsCreate := &api.IngestionEvent{IngestionEventTen: &api.IngestionEventTen{
		ID:        newSpanID(),
		Timestamp: getTimeRefString(a.StartTime),
		Type:      api.IngestionEventTenType(ingestionCreateAgent).Ptr(),
		Body: &api.CreateGenerationBody{
			ID:                  getStringRef(a.ObservationID),
			TraceID:             getStringRef(a.TraceID),
			ParentObservationID: getStringRef(a.ParentObservationID),
			Name:                getStringRef(a.Name),
			Metadata:            a.Metadata,
			Input:               convertInput(a.Input, a.Tools),
			Output:              convertOutput(a.Output),
			StartTime:           a.StartTime,
			EndTime:             a.EndTime,
			Level:               a.Level.ToLangfuse(),
			StatusMessage:       a.Status,
			Version:             a.Version,
			Model:               a.Model,
			ModelParameters:     a.ModelParameters.ToLangfuse(),
		},
	}}

	a.observer.enqueue(obsCreate)

	return a
}

func (a *agent) End(opts ...AgentOption) {
	id := a.ObservationID
	startTime := a.StartTime
	a.EndTime = getCurrentTimeRef()
	for _, opt := range opts {
		opt(a)
	}

	// preserve the original observation ID and start time
	a.ObservationID = id
	a.StartTime = startTime

	agentUpdate := &api.IngestionEvent{IngestionEventTen: &api.IngestionEventTen{
		ID:        newSpanID(),
		Timestamp: getTimeRefString(a.EndTime),
		Type:      api.IngestionEventTenType(ingestionCreateAgent).Ptr(),
		Body: &api.CreateGenerationBody{
			ID:              getStringRef(a.ObservationID),
			Name:            getStringRef(a.Name),
			Metadata:        a.Metadata,
			Input:           convertInput(a.Input, a.Tools),
			Output:          convertOutput(a.Output),
			EndTime:         a.EndTime,
			Level:           a.Level.ToLangfuse(),
			StatusMessage:   a.Status,
			Version:         a.Version,
			Model:           a.Model,
			ModelParameters: a.ModelParameters.ToLangfuse(),
		},
	}}

	a.observer.enqueue(agentUpdate)
}

func (a *agent) String() string {
	return fmt.Sprintf("Trace(%s) Observation(%s) Agent(%s)", a.TraceID, a.ObservationID, a.Name)
}

func (a *agent) MarshalJSON() ([]byte, error) {
	return json.Marshal(a)
}

func (a *agent) Observation(ctx context.Context) (context.Context, Observation) {
	obs := &observation{
		obsCtx: observationContext{
			TraceID:       a.TraceID,
			ObservationID: a.ObservationID,
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
