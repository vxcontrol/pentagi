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
	StartTime *time.Time            `json:"start_time,omitempty"`
	Value     *api.CreateScoreValue `json:"value,omitempty"`
	DataType  *api.ScoreDataType    `json:"data_type,omitempty"`
	Comment   *string               `json:"comment,omitempty"`
	ConfigId  *string               `json:"config_id,omitempty"`

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

func WithScoreTime(time time.Time) ScoreOption {
	return func(e *score) {
		e.StartTime = &time
	}
}

func WithScoreFloatValue(value float64) ScoreOption {
	return func(e *score) {
		e.Value = api.NewCreateScoreValueFromDouble(value)
		e.DataType = api.ScoreDataTypeNumeric.Ptr()
	}
}

func WithScoreStringValue(value string) ScoreOption {
	return func(e *score) {
		e.Value = api.NewCreateScoreValueFromString(value)
		e.DataType = api.ScoreDataTypeCategorical.Ptr()
	}
}

func WithScoreComment(comment string) ScoreOption {
	return func(e *score) {
		e.Comment = &comment
	}
}

func WithScoreConfigId(configId string) ScoreOption {
	return func(e *score) {
		e.ConfigId = &configId
	}
}

func newScore(observer enqueue, opts ...ScoreOption) {
	s := &score{
		ID:            newSpanID(),
		Name:          scoreDefaultName,
		ObservationID: newSpanID(),
		StartTime:     getCurrentTimeRef(),
		Value:         api.NewCreateScoreValueFromDouble(0),
		DataType:      api.ScoreDataTypeCategorical.Ptr(),
		observer:      observer,
	}

	for _, opt := range opts {
		opt(s)
	}

	obsCreate := api.NewIngestionEventFromIngestionEventOne(&api.IngestionEventOne{
		Id:        newSpanID(),
		Timestamp: getTimeRefString(s.StartTime),
		Type:      ingestionCreateScore,
		Body: &api.ScoreBody{
			Id:            getStringRef(s.ObservationID),
			ObservationId: getStringRef(s.ParentObservationID),
			TraceId:       s.TraceID,
			Name:          s.Name,
			Value:         s.Value,
			DataType:      s.DataType,
			Comment:       s.Comment,
			ConfigId:      s.ConfigId,
		},
	})

	s.observer.enqueue(obsCreate)
}
