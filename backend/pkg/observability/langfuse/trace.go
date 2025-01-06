package langfuse

import "time"

type TraceContext struct {
	Timestamp *time.Time  `json:"timestamp,omitempty"`
	Name      *string     `json:"name,omitempty"`
	UserId    *string     `json:"user_id,omitempty"`
	Input     interface{} `json:"input,omitempty"`
	Output    interface{} `json:"output,omitempty"`
	SessionId *string     `json:"session_id,omitempty"`
	Version   *string     `json:"version,omitempty"`
	Metadata  Metadata    `json:"metadata,omitempty"`
	Tags      []string    `json:"tags,omitempty"`
	Public    *bool       `json:"public,omitempty"`
}

type TraceContextOption func(*TraceContext)

func WithTraceTimestamp(timestamp time.Time) TraceContextOption {
	return func(t *TraceContext) {
		t.Timestamp = &timestamp
	}
}

func WithTraceName(name string) TraceContextOption {
	return func(t *TraceContext) {
		t.Name = &name
	}
}

func WithTraceUserId(userId string) TraceContextOption {
	return func(t *TraceContext) {
		t.UserId = &userId
	}
}

func WithTraceInput(input interface{}) TraceContextOption {
	return func(t *TraceContext) {
		t.Input = input
	}
}

func WithTraceOutput(output interface{}) TraceContextOption {
	return func(t *TraceContext) {
		t.Output = output
	}
}

func WithTraceSessionId(sessionId string) TraceContextOption {
	return func(t *TraceContext) {
		t.SessionId = &sessionId
	}
}

func WithTraceVersion(version string) TraceContextOption {
	return func(t *TraceContext) {
		t.Version = &version
	}
}

func WithTraceMetadata(metadata Metadata) TraceContextOption {
	return func(t *TraceContext) {
		t.Metadata = metadata
	}
}

func WithTraceTags(tags []string) TraceContextOption {
	return func(t *TraceContext) {
		t.Tags = tags
	}
}

func WithTracePublic() TraceContextOption {
	return func(t *TraceContext) {
		t.Public = getBoolRef(true)
	}
}
