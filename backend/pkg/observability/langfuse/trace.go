package langfuse

import "time"

type TraceContext struct {
	Timestamp *time.Time `json:"timestamp,omitempty"`
	Name      *string    `json:"name,omitempty"`
	UserID    *string    `json:"user_id,omitempty"`
	Input     any        `json:"input,omitempty"`
	Output    any        `json:"output,omitempty"`
	SessionID *string    `json:"session_id,omitempty"`
	Version   *string    `json:"version,omitempty"`
	Metadata  Metadata   `json:"metadata,omitempty"`
	Tags      []string   `json:"tags,omitempty"`
	Public    *bool      `json:"public,omitempty"`
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

func WithTraceUserID(userID string) TraceContextOption {
	return func(t *TraceContext) {
		t.UserID = &userID
	}
}

func WithTraceInput(input any) TraceContextOption {
	return func(t *TraceContext) {
		t.Input = convertInput(input, nil)
	}
}

func WithTraceOutput(output any) TraceContextOption {
	return func(t *TraceContext) {
		t.Output = convertOutput(output)
	}
}

func WithTraceSessionID(sessionID string) TraceContextOption {
	return func(t *TraceContext) {
		t.SessionID = &sessionID
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
