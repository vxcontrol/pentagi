package langfuse

import (
	"time"

	"pentagi/pkg/observability/langfuse/api"
)

const (
	scoreDefaultName = "Default Score"
)

type score struct {
	ID        string                `json:"id"`
	Name      string                `json:"name"`
	Metadata  Metadata              `json:"metadata,omitempty"`
	StartTime *time.Time            `json:"start_time,omitempty"`
	Value     *api.CreateScoreValue `json:"value,omitempty"`
	DataType  *api.ScoreDataType    `json:"data_type,omitempty"`
	Comment   *string               `json:"comment,omitempty"`
	ConfigID  *string               `json:"config_id,omitempty"`
	QueueID   *string               `json:"queue_id,omitempty"`

	TraceID             string `json:"trace_id"`
	ObservationID       string `json:"observation_id"`
	ParentObservationID string `json:"parent_observation_id"`

	observer enqueue `json:"-"`
}

type ScoreOption func(*score)

func withScoreTraceID(traceID string) ScoreOption {
	return func(e *score) {
		e.TraceID = traceID
	}
}

func withScoreParentObservationID(parentObservationID string) ScoreOption {
	return func(e *score) {
		e.ParentObservationID = parentObservationID
	}
}

func WithScoreID(id string) ScoreOption {
	return func(e *score) {
		e.ID = id
	}
}

func WithScoreName(name string) ScoreOption {
	return func(e *score) {
		e.Name = name
	}
}

func WithScoreMetadata(metadata Metadata) ScoreOption {
	return func(e *score) {
		e.Metadata = mergeMaps(e.Metadata, metadata)
	}
}

func WithScoreTime(time time.Time) ScoreOption {
	return func(e *score) {
		e.StartTime = &time
	}
}

func WithScoreFloatValue(value float64) ScoreOption {
	return func(e *score) {
		e.Value = &api.CreateScoreValue{Double: value}
		e.DataType = api.ScoreDataTypeNumeric.Ptr()
	}
}

func WithScoreStringValue(value string) ScoreOption {
	return func(e *score) {
		e.Value = &api.CreateScoreValue{String: value}
		e.DataType = api.ScoreDataTypeCategorical.Ptr()
	}
}

func WithScoreComment(comment string) ScoreOption {
	return func(e *score) {
		e.Comment = &comment
	}
}

func WithScoreConfigID(configID string) ScoreOption {
	return func(e *score) {
		e.ConfigID = &configID
	}
}

func WithScoreQueueID(queueID string) ScoreOption {
	return func(e *score) {
		e.QueueID = &queueID
	}
}

func newScore(observer enqueue, opts ...ScoreOption) {
	s := &score{
		ID:            newSpanID(),
		Name:          scoreDefaultName,
		ObservationID: newSpanID(),
		StartTime:     getCurrentTimeRef(),
		Value:         &api.CreateScoreValue{},
		DataType:      api.ScoreDataTypeCategorical.Ptr(),
		observer:      observer,
	}

	for _, opt := range opts {
		opt(s)
	}

	obsCreate := &api.IngestionEvent{IngestionEventOne: &api.IngestionEventOne{
		ID:        newSpanID(),
		Timestamp: getTimeRefString(s.StartTime),
		Type:      api.IngestionEventOneType(ingestionCreateScore).Ptr(),
		Body: &api.ScoreBody{
			ID:            getStringRef(s.ObservationID),
			ObservationID: getStringRef(s.ParentObservationID),
			TraceID:       getStringRef(s.TraceID),
			Name:          s.Name,
			Metadata:      s.Metadata,
			Value:         s.Value,
			DataType:      s.DataType,
			Comment:       s.Comment,
			ConfigID:      s.ConfigID,
			QueueID:       s.QueueID,
		},
	}}

	s.observer.enqueue(obsCreate)
}
