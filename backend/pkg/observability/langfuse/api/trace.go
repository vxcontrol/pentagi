// This file was auto-generated by Fern from our API Definition.

package api

import (
	json "encoding/json"
	fmt "fmt"
	internal "pentagi/pkg/observability/langfuse/api/internal"
	time "time"
)

type TraceListRequest struct {
	// Page number, starts at 1
	Page *int `json:"-" url:"page,omitempty"`
	// Limit of items per page. If you encounter api issues due to too large page sizes, try to reduce the limit.
	Limit     *int    `json:"-" url:"limit,omitempty"`
	UserId    *string `json:"-" url:"userId,omitempty"`
	Name      *string `json:"-" url:"name,omitempty"`
	SessionId *string `json:"-" url:"sessionId,omitempty"`
	// Optional filter to only include traces with a trace.timestamp on or after a certain datetime (ISO 8601)
	FromTimestamp *time.Time `json:"-" url:"fromTimestamp,omitempty"`
	// Optional filter to only include traces with a trace.timestamp before a certain datetime (ISO 8601)
	ToTimestamp *time.Time `json:"-" url:"toTimestamp,omitempty"`
	// Format of the string [field].[asc/desc]. Fields: id, timestamp, name, userId, release, version, public, bookmarked, sessionId. Example: timestamp.asc
	OrderBy *string `json:"-" url:"orderBy,omitempty"`
	// Only traces that include all of these tags will be returned.
	Tags []*string `json:"-" url:"tags,omitempty"`
	// Optional filter to only include traces with a certain version.
	Version *string `json:"-" url:"version,omitempty"`
	// Optional filter to only include traces with a certain release.
	Release *string `json:"-" url:"release,omitempty"`
}

type TraceWithDetails struct {
	// The unique identifier of a trace
	Id string `json:"id" url:"id"`
	// The timestamp when the trace was created
	Timestamp time.Time `json:"timestamp" url:"timestamp"`
	// The name of the trace
	Name   *string     `json:"name,omitempty" url:"name,omitempty"`
	Input  interface{} `json:"input,omitempty" url:"input,omitempty"`
	Output interface{} `json:"output,omitempty" url:"output,omitempty"`
	// The session identifier associated with the trace
	SessionId *string `json:"sessionId,omitempty" url:"sessionId,omitempty"`
	// The release version of the application when the trace was created
	Release *string `json:"release,omitempty" url:"release,omitempty"`
	// The version of the trace
	Version *string `json:"version,omitempty" url:"version,omitempty"`
	// The user identifier associated with the trace
	UserId   *string     `json:"userId,omitempty" url:"userId,omitempty"`
	Metadata interface{} `json:"metadata,omitempty" url:"metadata,omitempty"`
	// The tags associated with the trace. Can be an array of strings or null.
	Tags []string `json:"tags,omitempty" url:"tags,omitempty"`
	// Public traces are accessible via url without login
	Public *bool `json:"public,omitempty" url:"public,omitempty"`
	// Path of trace in Langfuse UI
	HtmlPath string `json:"htmlPath" url:"htmlPath"`
	// Latency of trace in seconds
	Latency float64 `json:"latency" url:"latency"`
	// Cost of trace in USD
	TotalCost float64 `json:"totalCost" url:"totalCost"`
	// List of observation ids
	Observations []string `json:"observations,omitempty" url:"observations,omitempty"`
	// List of score ids
	Scores []string `json:"scores,omitempty" url:"scores,omitempty"`

	extraProperties map[string]interface{}
	rawJSON         json.RawMessage
}

func (t *TraceWithDetails) GetId() string {
	if t == nil {
		return ""
	}
	return t.Id
}

func (t *TraceWithDetails) GetTimestamp() time.Time {
	if t == nil {
		return time.Time{}
	}
	return t.Timestamp
}

func (t *TraceWithDetails) GetName() *string {
	if t == nil {
		return nil
	}
	return t.Name
}

func (t *TraceWithDetails) GetInput() interface{} {
	if t == nil {
		return nil
	}
	return t.Input
}

func (t *TraceWithDetails) GetOutput() interface{} {
	if t == nil {
		return nil
	}
	return t.Output
}

func (t *TraceWithDetails) GetSessionId() *string {
	if t == nil {
		return nil
	}
	return t.SessionId
}

func (t *TraceWithDetails) GetRelease() *string {
	if t == nil {
		return nil
	}
	return t.Release
}

func (t *TraceWithDetails) GetVersion() *string {
	if t == nil {
		return nil
	}
	return t.Version
}

func (t *TraceWithDetails) GetUserId() *string {
	if t == nil {
		return nil
	}
	return t.UserId
}

func (t *TraceWithDetails) GetMetadata() interface{} {
	if t == nil {
		return nil
	}
	return t.Metadata
}

func (t *TraceWithDetails) GetTags() []string {
	if t == nil {
		return nil
	}
	return t.Tags
}

func (t *TraceWithDetails) GetPublic() *bool {
	if t == nil {
		return nil
	}
	return t.Public
}

func (t *TraceWithDetails) GetHtmlPath() string {
	if t == nil {
		return ""
	}
	return t.HtmlPath
}

func (t *TraceWithDetails) GetLatency() float64 {
	if t == nil {
		return 0
	}
	return t.Latency
}

func (t *TraceWithDetails) GetTotalCost() float64 {
	if t == nil {
		return 0
	}
	return t.TotalCost
}

func (t *TraceWithDetails) GetObservations() []string {
	if t == nil {
		return nil
	}
	return t.Observations
}

func (t *TraceWithDetails) GetScores() []string {
	if t == nil {
		return nil
	}
	return t.Scores
}

func (t *TraceWithDetails) GetExtraProperties() map[string]interface{} {
	return t.extraProperties
}

func (t *TraceWithDetails) UnmarshalJSON(data []byte) error {
	type embed TraceWithDetails
	var unmarshaler = struct {
		embed
		Timestamp *internal.DateTime `json:"timestamp"`
	}{
		embed: embed(*t),
	}
	if err := json.Unmarshal(data, &unmarshaler); err != nil {
		return err
	}
	*t = TraceWithDetails(unmarshaler.embed)
	t.Timestamp = unmarshaler.Timestamp.Time()
	extraProperties, err := internal.ExtractExtraProperties(data, *t)
	if err != nil {
		return err
	}
	t.extraProperties = extraProperties
	t.rawJSON = json.RawMessage(data)
	return nil
}

func (t *TraceWithDetails) MarshalJSON() ([]byte, error) {
	type embed TraceWithDetails
	var marshaler = struct {
		embed
		Timestamp *internal.DateTime `json:"timestamp"`
	}{
		embed:     embed(*t),
		Timestamp: internal.NewDateTime(t.Timestamp),
	}
	return json.Marshal(marshaler)
}

func (t *TraceWithDetails) String() string {
	if len(t.rawJSON) > 0 {
		if value, err := internal.StringifyJSON(t.rawJSON); err == nil {
			return value
		}
	}
	if value, err := internal.StringifyJSON(t); err == nil {
		return value
	}
	return fmt.Sprintf("%#v", t)
}

type TraceWithFullDetails struct {
	// The unique identifier of a trace
	Id string `json:"id" url:"id"`
	// The timestamp when the trace was created
	Timestamp time.Time `json:"timestamp" url:"timestamp"`
	// The name of the trace
	Name   *string     `json:"name,omitempty" url:"name,omitempty"`
	Input  interface{} `json:"input,omitempty" url:"input,omitempty"`
	Output interface{} `json:"output,omitempty" url:"output,omitempty"`
	// The session identifier associated with the trace
	SessionId *string `json:"sessionId,omitempty" url:"sessionId,omitempty"`
	// The release version of the application when the trace was created
	Release *string `json:"release,omitempty" url:"release,omitempty"`
	// The version of the trace
	Version *string `json:"version,omitempty" url:"version,omitempty"`
	// The user identifier associated with the trace
	UserId   *string     `json:"userId,omitempty" url:"userId,omitempty"`
	Metadata interface{} `json:"metadata,omitempty" url:"metadata,omitempty"`
	// The tags associated with the trace. Can be an array of strings or null.
	Tags []string `json:"tags,omitempty" url:"tags,omitempty"`
	// Public traces are accessible via url without login
	Public *bool `json:"public,omitempty" url:"public,omitempty"`
	// Path of trace in Langfuse UI
	HtmlPath string `json:"htmlPath" url:"htmlPath"`
	// Latency of trace in seconds
	Latency float64 `json:"latency" url:"latency"`
	// Cost of trace in USD
	TotalCost float64 `json:"totalCost" url:"totalCost"`
	// List of observations
	Observations []*ObservationsView `json:"observations,omitempty" url:"observations,omitempty"`
	// List of scores
	Scores []*Score `json:"scores,omitempty" url:"scores,omitempty"`

	extraProperties map[string]interface{}
	rawJSON         json.RawMessage
}

func (t *TraceWithFullDetails) GetId() string {
	if t == nil {
		return ""
	}
	return t.Id
}

func (t *TraceWithFullDetails) GetTimestamp() time.Time {
	if t == nil {
		return time.Time{}
	}
	return t.Timestamp
}

func (t *TraceWithFullDetails) GetName() *string {
	if t == nil {
		return nil
	}
	return t.Name
}

func (t *TraceWithFullDetails) GetInput() interface{} {
	if t == nil {
		return nil
	}
	return t.Input
}

func (t *TraceWithFullDetails) GetOutput() interface{} {
	if t == nil {
		return nil
	}
	return t.Output
}

func (t *TraceWithFullDetails) GetSessionId() *string {
	if t == nil {
		return nil
	}
	return t.SessionId
}

func (t *TraceWithFullDetails) GetRelease() *string {
	if t == nil {
		return nil
	}
	return t.Release
}

func (t *TraceWithFullDetails) GetVersion() *string {
	if t == nil {
		return nil
	}
	return t.Version
}

func (t *TraceWithFullDetails) GetUserId() *string {
	if t == nil {
		return nil
	}
	return t.UserId
}

func (t *TraceWithFullDetails) GetMetadata() interface{} {
	if t == nil {
		return nil
	}
	return t.Metadata
}

func (t *TraceWithFullDetails) GetTags() []string {
	if t == nil {
		return nil
	}
	return t.Tags
}

func (t *TraceWithFullDetails) GetPublic() *bool {
	if t == nil {
		return nil
	}
	return t.Public
}

func (t *TraceWithFullDetails) GetHtmlPath() string {
	if t == nil {
		return ""
	}
	return t.HtmlPath
}

func (t *TraceWithFullDetails) GetLatency() float64 {
	if t == nil {
		return 0
	}
	return t.Latency
}

func (t *TraceWithFullDetails) GetTotalCost() float64 {
	if t == nil {
		return 0
	}
	return t.TotalCost
}

func (t *TraceWithFullDetails) GetObservations() []*ObservationsView {
	if t == nil {
		return nil
	}
	return t.Observations
}

func (t *TraceWithFullDetails) GetScores() []*Score {
	if t == nil {
		return nil
	}
	return t.Scores
}

func (t *TraceWithFullDetails) GetExtraProperties() map[string]interface{} {
	return t.extraProperties
}

func (t *TraceWithFullDetails) UnmarshalJSON(data []byte) error {
	type embed TraceWithFullDetails
	var unmarshaler = struct {
		embed
		Timestamp *internal.DateTime `json:"timestamp"`
	}{
		embed: embed(*t),
	}
	if err := json.Unmarshal(data, &unmarshaler); err != nil {
		return err
	}
	*t = TraceWithFullDetails(unmarshaler.embed)
	t.Timestamp = unmarshaler.Timestamp.Time()
	extraProperties, err := internal.ExtractExtraProperties(data, *t)
	if err != nil {
		return err
	}
	t.extraProperties = extraProperties
	t.rawJSON = json.RawMessage(data)
	return nil
}

func (t *TraceWithFullDetails) MarshalJSON() ([]byte, error) {
	type embed TraceWithFullDetails
	var marshaler = struct {
		embed
		Timestamp *internal.DateTime `json:"timestamp"`
	}{
		embed:     embed(*t),
		Timestamp: internal.NewDateTime(t.Timestamp),
	}
	return json.Marshal(marshaler)
}

func (t *TraceWithFullDetails) String() string {
	if len(t.rawJSON) > 0 {
		if value, err := internal.StringifyJSON(t.rawJSON); err == nil {
			return value
		}
	}
	if value, err := internal.StringifyJSON(t); err == nil {
		return value
	}
	return fmt.Sprintf("%#v", t)
}

type Traces struct {
	Data []*TraceWithDetails `json:"data,omitempty" url:"data,omitempty"`
	Meta *UtilsMetaResponse  `json:"meta,omitempty" url:"meta,omitempty"`

	extraProperties map[string]interface{}
	rawJSON         json.RawMessage
}

func (t *Traces) GetData() []*TraceWithDetails {
	if t == nil {
		return nil
	}
	return t.Data
}

func (t *Traces) GetMeta() *UtilsMetaResponse {
	if t == nil {
		return nil
	}
	return t.Meta
}

func (t *Traces) GetExtraProperties() map[string]interface{} {
	return t.extraProperties
}

func (t *Traces) UnmarshalJSON(data []byte) error {
	type unmarshaler Traces
	var value unmarshaler
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	*t = Traces(value)
	extraProperties, err := internal.ExtractExtraProperties(data, *t)
	if err != nil {
		return err
	}
	t.extraProperties = extraProperties
	t.rawJSON = json.RawMessage(data)
	return nil
}

func (t *Traces) String() string {
	if len(t.rawJSON) > 0 {
		if value, err := internal.StringifyJSON(t.rawJSON); err == nil {
			return value
		}
	}
	if value, err := internal.StringifyJSON(t); err == nil {
		return value
	}
	return fmt.Sprintf("%#v", t)
}