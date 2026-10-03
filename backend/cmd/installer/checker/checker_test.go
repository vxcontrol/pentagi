package checker

import (
	"os"
	"path/filepath"
	"sync"
	"testing"

	"pentagi/cmd/installer/loader"
)

// Drives GatherGraphitiInfo, GatherLangfuseInfo and GatherObservabilityInfo on one configuration.
func TestChecker_NeverClaimsExternalForAStackThatIsNotConnected(t *testing.T) {
	envPath := filepath.Join(t.TempDir(), ".env")
	tests := []struct {
		name                string
		vars                map[string]loader.EnvVar
		connected, external bool
	}{
		{"nothing configured", map[string]loader.EnvVar{}, false, false},
		{"configured against somebody else's endpoint", map[string]loader.EnvVar{
			"GRAPHITI_ENABLED":    {Value: "true"},
			"GRAPHITI_URL":        {Value: "https://graphiti.example.com"},
			"LANGFUSE_BASE_URL":   {Value: "https://langfuse.example.com"},
			"LANGFUSE_PROJECT_ID": {Value: "p"},
			"LANGFUSE_PUBLIC_KEY": {Value: "pk"},
			"LANGFUSE_SECRET_KEY": {Value: "sk"},
			"OTEL_HOST":           {Value: "otel.example.com:4317"},
		}, true, true},
		{"configured against the bundled endpoint", map[string]loader.EnvVar{
			"GRAPHITI_ENABLED":    {Value: "true"},
			"GRAPHITI_URL":        {Value: DefaultGraphitiEndpoint},
			"LANGFUSE_BASE_URL":   {Value: DefaultLangfuseEndpoint},
			"LANGFUSE_PROJECT_ID": {Value: "p"},
			"LANGFUSE_PUBLIC_KEY": {Value: "pk"},
			"LANGFUSE_SECRET_KEY": {Value: "sk"},
			"OTEL_HOST":           {Value: DefaultObservabilityEndpoint},
		}, true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := &defaultCheckHandler{mx: &sync.Mutex{}, appState: &mockState{vars: tt.vars, envPath: envPath}}
			got := &CheckResult{}
			for _, err := range []error{
				handler.GatherGraphitiInfo(t.Context(), got),
				handler.GatherLangfuseInfo(t.Context(), got),
				handler.GatherObservabilityInfo(t.Context(), got),
			} {
				if err != nil {
					t.Fatalf("gather: %v", err)
				}
			}

			for stack, flags := range map[string][2]bool{
				"graphiti":      {got.GraphitiConnected, got.GraphitiExternal},
				"langfuse":      {got.LangfuseConnected, got.LangfuseExternal},
				"observability": {got.ObservabilityConnected, got.ObservabilityExternal},
			} {
				if flags != [2]bool{tt.connected, tt.external} {
					t.Errorf("%s: connected=%t external=%t, want %t and %t", stack, flags[0], flags[1], tt.connected, tt.external)
				}
			}
		})
	}
}

func TestChecker_GatherPentagiInfo_LooksForTheExtractedStackBesideTheEnvFile(t *testing.T) {
	tests := []struct {
		name              string
		beside, elsewhere []string // files beside the env file, and in the working directory
		want              bool
	}{
		{
			"the compose file and both examples beside the env file",
			[]string{"docker-compose.yml", "example.custom.provider.yml", "example.ollama.provider.yml"}, nil, true,
		},
		{"the compose file alone", []string{"docker-compose.yml"}, nil, false},
		{"the ollama example missing", []string{"docker-compose.yml", "example.custom.provider.yml"}, nil, false},
		{"the custom example missing", []string{"docker-compose.yml", "example.ollama.provider.yml"}, nil, false},
		{"the compose file missing", []string{"example.custom.provider.yml", "example.ollama.provider.yml"}, nil, false},
		{
			"the examples in the working directory only",
			[]string{"docker-compose.yml"}, []string{"example.custom.provider.yml", "example.ollama.provider.yml"}, false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			envDir, workDir := t.TempDir(), t.TempDir()
			for dir, names := range map[string][]string{envDir: tt.beside, workDir: tt.elsewhere} {
				for _, name := range names {
					if err := os.WriteFile(filepath.Join(dir, name), []byte("x"), 0o644); err != nil {
						t.Fatalf("write %s: %v", name, err)
					}
				}
			}
			t.Chdir(workDir)

			handler := &defaultCheckHandler{mx: &sync.Mutex{}, appState: &mockState{envPath: filepath.Join(envDir, ".env")}}
			got := &CheckResult{}
			if err := handler.GatherPentagiInfo(t.Context(), got); err != nil {
				t.Fatalf("gather: %v", err)
			}

			if got.PentagiExtracted != tt.want {
				t.Errorf("PentagiExtracted = %t, want %t", got.PentagiExtracted, tt.want)
			}
		})
	}
}
