package langfuse

import (
	"context"
	"fmt"

	"pentagi/pkg/observability/langfuse/api"

	"github.com/sirupsen/logrus"
)

type Observation interface {
	ID() string
	TraceID() string
	String() string
	Log(ctx context.Context, message string)
	Score(opts ...ScoreOption)
	Event(opts ...EventOption) Event
	Span(opts ...SpanOption) Span
	Generation(opts ...GenerationOption) Generation
	Agent(opts ...AgentOption) Agent
	Tool(opts ...ToolOption) Tool
	Chain(opts ...ChainOption) Chain
	Retriever(opts ...RetrieverOption) Retriever
	Evaluator(opts ...EvaluatorOption) Evaluator
	Embedding(opts ...EmbeddingOption) Embedding
	Guardrail(opts ...GuardrailOption) Guardrail
}

type observation struct {
	obsCtx   observationContext
	observer enqueue
}

func (o *observation) ID() string {
	return o.obsCtx.ObservationID
}

func (o *observation) TraceID() string {
	return o.obsCtx.TraceID
}

func (o *observation) String() string {
	return fmt.Sprintf("Trace(%s) Observation(%s)", o.obsCtx.TraceID, o.obsCtx.ObservationID)
}

func (o *observation) Log(ctx context.Context, message string) {
	logID := newSpanID()
	logrus.WithContext(ctx).WithFields(logrus.Fields{
		"langfuse_trace_id":       o.obsCtx.TraceID,
		"langfuse_observation_id": o.obsCtx.ObservationID,
		"langfuse_log_id":         logID,
	}).Info(message)

	obsLog := &api.IngestionEvent{IngestionEventSeven: &api.IngestionEventSeven{
		ID:        logID,
		Timestamp: getCurrentTimeString(),
		Type:      api.IngestionEventSevenType(ingestionPutLog).Ptr(),
		Body: &api.SdkLogBody{
			Log: message,
		},
	}}

	o.observer.enqueue(obsLog)
}

func (o *observation) Score(opts ...ScoreOption) {
	opts = append(opts,
		withScoreTraceID(o.obsCtx.TraceID),
		withScoreParentObservationID(o.obsCtx.ObservationID),
	)
	newScore(o.observer, opts...)
}

func (o *observation) Event(opts ...EventOption) Event {
	opts = append(opts,
		withEventTraceID(o.obsCtx.TraceID),
		withEventParentObservationID(o.obsCtx.ObservationID),
	)
	return newEvent(o.observer, opts...)
}

func (o *observation) Span(opts ...SpanOption) Span {
	opts = append(opts,
		withSpanTraceID(o.obsCtx.TraceID),
		withSpanParentObservationID(o.obsCtx.ObservationID),
	)
	return newSpan(o.observer, opts...)
}

func (o *observation) Generation(opts ...GenerationOption) Generation {
	opts = append(opts,
		withGenerationTraceID(o.obsCtx.TraceID),
		withGenerationParentObservationID(o.obsCtx.ObservationID),
	)
	return newGeneration(o.observer, opts...)
}

func (o *observation) Agent(opts ...AgentOption) Agent {
	opts = append(opts,
		withAgentTraceID(o.obsCtx.TraceID),
		withAgentParentObservationID(o.obsCtx.ObservationID),
	)
	return newAgent(o.observer, opts...)
}

func (o *observation) Tool(opts ...ToolOption) Tool {
	opts = append(opts,
		withToolTraceID(o.obsCtx.TraceID),
		withToolParentObservationID(o.obsCtx.ObservationID),
	)
	return newTool(o.observer, opts...)
}

func (o *observation) Chain(opts ...ChainOption) Chain {
	opts = append(opts,
		withChainTraceID(o.obsCtx.TraceID),
		withChainParentObservationID(o.obsCtx.ObservationID),
	)
	return newChain(o.observer, opts...)
}

func (o *observation) Retriever(opts ...RetrieverOption) Retriever {
	opts = append(opts,
		withRetrieverTraceID(o.obsCtx.TraceID),
		withRetrieverParentObservationID(o.obsCtx.ObservationID),
	)
	return newRetriever(o.observer, opts...)
}

func (o *observation) Evaluator(opts ...EvaluatorOption) Evaluator {
	opts = append(opts,
		withEvaluatorTraceID(o.obsCtx.TraceID),
		withEvaluatorParentObservationID(o.obsCtx.ObservationID),
	)
	return newEvaluator(o.observer, opts...)
}

func (o *observation) Embedding(opts ...EmbeddingOption) Embedding {
	opts = append(opts,
		withEmbeddingTraceID(o.obsCtx.TraceID),
		withEmbeddingParentObservationID(o.obsCtx.ObservationID),
	)
	return newEmbedding(o.observer, opts...)
}

func (o *observation) Guardrail(opts ...GuardrailOption) Guardrail {
	opts = append(opts,
		withGuardrailTraceID(o.obsCtx.TraceID),
		withGuardrailParentObservationID(o.obsCtx.ObservationID),
	)
	return newGuardrail(o.observer, opts...)
}

type ObservationInfo struct {
	TraceID             string `json:"trace_id"`
	ObservationID       string `json:"observation_id"`
	ParentObservationID string `json:"parent_observation_id"`
}

type ObservationContext struct {
	TraceID       string
	TraceCtx      *TraceContext
	ObservationID string
}

type ObservationContextOption func(*ObservationContext)

func WithObservationTraceID(traceID string) ObservationContextOption {
	return func(o *ObservationContext) {
		o.TraceID = traceID
	}
}

func WithObservationID(observationID string) ObservationContextOption {
	return func(o *ObservationContext) {
		o.ObservationID = observationID
	}
}

func WithObservationTraceContext(opts ...TraceContextOption) ObservationContextOption {
	traceCtx := &TraceContext{
		Timestamp: getCurrentTimeRef(),
		Version:   getStringRef(firstVersion),
	}
	for _, opt := range opts {
		opt(traceCtx)
	}

	return func(o *ObservationContext) {
		o.TraceCtx = traceCtx
	}
}
