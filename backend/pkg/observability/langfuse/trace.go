package langfuse

import (
	"maps"
	"slices"
	"time"
)

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

func cloneTraceContext(src *TraceContext) *TraceContext {
	if src == nil {
		return nil
	}
	cp := *src
	if src.Metadata != nil {
		cp.Metadata = maps.Clone(src.Metadata)
	}
	if src.Tags != nil {
		cp.Tags = slices.Clone(src.Tags)
	}
	return &cp
}

func mergeTraceContext(base, update *TraceContext) *TraceContext {
	if update == nil {
		return cloneTraceContext(base)
	}
	if base == nil {
		return cloneTraceContext(update)
	}
	merged := cloneTraceContext(base)
	if update.Timestamp != nil {
		merged.Timestamp = update.Timestamp
	}
	if update.Name != nil {
		merged.Name = update.Name
	}
	if update.UserID != nil {
		merged.UserID = update.UserID
	}
	if update.Input != nil {
		merged.Input = update.Input
	}
	if update.Output != nil {
		merged.Output = update.Output
	}
	if update.SessionID != nil {
		merged.SessionID = update.SessionID
	}
	if update.Version != nil {
		merged.Version = update.Version
	}
	if update.Metadata != nil {
		merged.Metadata = mergeMaps(merged.Metadata, update.Metadata)
	}
	if len(update.Tags) > 0 {
		merged.Tags = slices.Clone(update.Tags)
	}
	if update.Public != nil {
		merged.Public = update.Public
	}
	return merged
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
