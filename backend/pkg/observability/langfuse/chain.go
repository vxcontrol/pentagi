package langfuse

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"pentagi/pkg/observability/langfuse/api"
)

const (
	chainDefaultName = "Default Chain"
)

type Chain interface {
	End(opts ...ChainOption)
	String() string
	MarshalJSON() ([]byte, error)
	Observation(ctx context.Context) (context.Context, Observation)
	ObservationInfo() ObservationInfo
}

type chain struct {
	Name      string           `json:"name"`
	Metadata  Metadata         `json:"metadata,omitempty"`
	Input     any              `json:"input,omitempty"`
	Output    any              `json:"output,omitempty"`
	StartTime *time.Time       `json:"start_time,omitempty"`
	EndTime   *time.Time       `json:"end_time,omitempty"`
	Level     ObservationLevel `json:"level"`
	Status    *string          `json:"status,omitempty"`
	Version   *string          `json:"version,omitempty"`

	TraceID             string `json:"trace_id"`
	ObservationID       string `json:"observation_id"`
	ParentObservationID string `json:"parent_observation_id"`

	observer enqueue `json:"-"`
}

type ChainOption func(*chain)

func withChainTraceID(traceID string) ChainOption {
	return func(c *chain) {
		c.TraceID = traceID
	}
}

func withChainParentObservationID(parentObservationID string) ChainOption {
	return func(c *chain) {
		c.ParentObservationID = parentObservationID
	}
}

// WithChainID sets on creation time
func WithChainID(id string) ChainOption {
	return func(c *chain) {
		c.ObservationID = id
	}
}

func WithChainName(name string) ChainOption {
	return func(c *chain) {
		c.Name = name
	}
}

func WithChainMetadata(metadata Metadata) ChainOption {
	return func(c *chain) {
		c.Metadata = mergeMaps(c.Metadata, metadata)
	}
}

func WithChainInput(input any) ChainOption {
	return func(c *chain) {
		c.Input = input
	}
}

func WithChainOutput(output any) ChainOption {
	return func(c *chain) {
		c.Output = output
	}
}

// WithChainStartTime sets on creation time
func WithChainStartTime(time time.Time) ChainOption {
	return func(c *chain) {
		c.StartTime = &time
	}
}

func WithChainEndTime(time time.Time) ChainOption {
	return func(c *chain) {
		c.EndTime = &time
	}
}

func WithChainLevel(level ObservationLevel) ChainOption {
	return func(c *chain) {
		c.Level = level
	}
}

func WithChainStatus(status string) ChainOption {
	return func(c *chain) {
		c.Status = &status
	}
}

func WithChainVersion(version string) ChainOption {
	return func(c *chain) {
		c.Version = &version
	}
}

func newChain(observer enqueue, opts ...ChainOption) Chain {
	c := &chain{
		Name:          chainDefaultName,
		ObservationID: newSpanID(),
		Version:       getStringRef(firstVersion),
		StartTime:     getCurrentTimeRef(),
		observer:      observer,
	}

	for _, opt := range opts {
		opt(c)
	}

	obsCreate := &api.IngestionEvent{IngestionEventTwelve: &api.IngestionEventTwelve{
		ID:        newSpanID(),
		Timestamp: getTimeRefString(c.StartTime),
		Type:      api.IngestionEventTwelveType(ingestionCreateChain).Ptr(),
		Body: &api.CreateGenerationBody{
			ID:                  getStringRef(c.ObservationID),
			TraceID:             getStringRef(c.TraceID),
			ParentObservationID: getStringRef(c.ParentObservationID),
			Name:                getStringRef(c.Name),
			Metadata:            c.Metadata,
			Input:               convertInput(c.Input, nil),
			Output:              convertOutput(c.Output),
			StartTime:           c.StartTime,
			EndTime:             c.EndTime,
			Level:               c.Level.ToLangfuse(),
			StatusMessage:       c.Status,
			Version:             c.Version,
		},
	}}

	c.observer.enqueue(obsCreate)

	return c
}

func (c *chain) End(opts ...ChainOption) {
	id := c.ObservationID
	startTime := c.StartTime
	c.EndTime = getCurrentTimeRef()
	for _, opt := range opts {
		opt(c)
	}

	// preserve the original observation ID and start time
	c.ObservationID = id
	c.StartTime = startTime

	chainUpdate := &api.IngestionEvent{IngestionEventTwelve: &api.IngestionEventTwelve{
		ID:        newSpanID(),
		Timestamp: getTimeRefString(c.EndTime),
		Type:      api.IngestionEventTwelveType(ingestionCreateChain).Ptr(),
		Body: &api.CreateGenerationBody{
			ID:            getStringRef(c.ObservationID),
			Name:          getStringRef(c.Name),
			Metadata:      c.Metadata,
			Input:         convertInput(c.Input, nil),
			Output:        convertOutput(c.Output),
			EndTime:       c.EndTime,
			Level:         c.Level.ToLangfuse(),
			StatusMessage: c.Status,
			Version:       c.Version,
		},
	}}

	c.observer.enqueue(chainUpdate)
}

func (c *chain) String() string {
	return fmt.Sprintf("Trace(%s) Observation(%s) Chain(%s)", c.TraceID, c.ObservationID, c.Name)
}

func (c *chain) MarshalJSON() ([]byte, error) {
	return json.Marshal(c)
}

func (c *chain) Observation(ctx context.Context) (context.Context, Observation) {
	obs := &observation{
		obsCtx: observationContext{
			TraceID:       c.TraceID,
			ObservationID: c.ObservationID,
		},
		observer: c.observer,
	}

	return putObservationContext(ctx, obs.obsCtx), obs
}

func (c *chain) ObservationInfo() ObservationInfo {
	return ObservationInfo{
		TraceID:             c.TraceID,
		ObservationID:       c.ObservationID,
		ParentObservationID: c.ParentObservationID,
	}
}
