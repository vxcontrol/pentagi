package checker

import (
	"context"
	"path/filepath"
	"sync"
	"testing"

	"pentagi/cmd/installer/loader"
)

// A stack nobody configured is not "externally hosted".
//
// The three External flags are computed as "the configured URL differs from the
// bundled default", and an unset URL differs from everything — so an
// installation with no Graphiti at all used to report
// `graphiti_connected: false, graphiti_external: true`. That is what a released
// installer still puts on the legacy wire, and it is what the TUI told its
// operator.
//
// Nothing downstream ever changed behaviour because of it: every consumer starts
// with `Connected &&`, so the term was dead exactly when the bug fired. What this
// pins is the honesty of the field, whose one consumer that cannot apply the
// guard is a person reading it.
//
// Driven through the real gather functions rather than through a copy of the
// expression, because a test of `connected && differs` would pass with the fix
// reverted.
func TestExternalIsNeverClaimedForAStackThatIsNotConnected(t *testing.T) {
	envPath := filepath.Join(t.TempDir(), ".env")
	gather := func(t *testing.T, vars map[string]loader.EnvVar) *CheckResult {
		t.Helper()
		handler := &defaultCheckHandler{mx: &sync.Mutex{}, appState: &mockState{vars: vars, envPath: envPath}}
		result := &CheckResult{}
		ctx := context.Background()
		if err := handler.GatherGraphitiInfo(ctx, result); err != nil {
			t.Fatalf("GatherGraphitiInfo: %v", err)
		}
		if err := handler.GatherLangfuseInfo(ctx, result); err != nil {
			t.Fatalf("GatherLangfuseInfo: %v", err)
		}
		if err := handler.GatherObservabilityInfo(ctx, result); err != nil {
			t.Fatalf("GatherObservabilityInfo: %v", err)
		}
		return result
	}

	t.Run("nothing configured", func(t *testing.T) {
		got := gather(t, map[string]loader.EnvVar{})

		if got.GraphitiConnected || got.GraphitiExternal {
			t.Errorf("graphiti: connected=%v external=%v, want both false",
				got.GraphitiConnected, got.GraphitiExternal)
		}
		if got.LangfuseConnected || got.LangfuseExternal {
			t.Errorf("langfuse: connected=%v external=%v, want both false",
				got.LangfuseConnected, got.LangfuseExternal)
		}
		if got.ObservabilityConnected || got.ObservabilityExternal {
			t.Errorf("observability: connected=%v external=%v, want both false",
				got.ObservabilityConnected, got.ObservabilityExternal)
		}
	})

	t.Run("configured against somebody else's endpoint", func(t *testing.T) {
		got := gather(t, map[string]loader.EnvVar{
			"GRAPHITI_ENABLED":    {Value: "true"},
			"GRAPHITI_URL":        {Value: "https://graphiti.example.com"},
			"LANGFUSE_BASE_URL":   {Value: "https://langfuse.example.com"},
			"LANGFUSE_PROJECT_ID": {Value: "p"},
			"LANGFUSE_PUBLIC_KEY": {Value: "pk"},
			"LANGFUSE_SECRET_KEY": {Value: "sk"},
			"OTEL_HOST":           {Value: "otel.example.com:4317"},
		})

		if !got.GraphitiConnected || !got.GraphitiExternal {
			t.Errorf("graphiti: connected=%v external=%v, want both true",
				got.GraphitiConnected, got.GraphitiExternal)
		}
		if !got.LangfuseConnected || !got.LangfuseExternal {
			t.Errorf("langfuse: connected=%v external=%v, want both true",
				got.LangfuseConnected, got.LangfuseExternal)
		}
		if !got.ObservabilityConnected || !got.ObservabilityExternal {
			t.Errorf("observability: connected=%v external=%v, want both true",
				got.ObservabilityConnected, got.ObservabilityExternal)
		}
	})

	t.Run("configured against the bundled endpoint", func(t *testing.T) {
		got := gather(t, map[string]loader.EnvVar{
			"GRAPHITI_ENABLED":    {Value: "true"},
			"GRAPHITI_URL":        {Value: DefaultGraphitiEndpoint},
			"LANGFUSE_BASE_URL":   {Value: DefaultLangfuseEndpoint},
			"LANGFUSE_PROJECT_ID": {Value: "p"},
			"LANGFUSE_PUBLIC_KEY": {Value: "pk"},
			"LANGFUSE_SECRET_KEY": {Value: "sk"},
			"OTEL_HOST":           {Value: DefaultObservabilityEndpoint},
		})

		if !got.GraphitiConnected || got.GraphitiExternal {
			t.Errorf("graphiti: connected=%v external=%v, want connected and not external",
				got.GraphitiConnected, got.GraphitiExternal)
		}
		if !got.LangfuseConnected || got.LangfuseExternal {
			t.Errorf("langfuse: connected=%v external=%v, want connected and not external",
				got.LangfuseConnected, got.LangfuseExternal)
		}
		if !got.ObservabilityConnected || got.ObservabilityExternal {
			t.Errorf("observability: connected=%v external=%v, want connected and not external",
				got.ObservabilityConnected, got.ObservabilityExternal)
		}
	})
}
