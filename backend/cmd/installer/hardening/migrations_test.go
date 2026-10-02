package hardening

import (
	"errors"
	"testing"

	"pentagi/cmd/installer/loader"
)

func TestMigrations_DoMigrateSettings_MovesAHostPathToItsPentagiVar(t *testing.T) {
	dir, file := hardeningHostPaths(t)
	host := map[string]string{
		"DOCKER_CERT_PATH": dir, "LLM_SERVER_CONFIG_PATH": file, "OLLAMA_SERVER_CONFIG_PATH": file, "BEDROCK_CONFIG_PATH": file,
	}
	container := map[string]string{
		"DOCKER_CERT_PATH":          "/opt/pentagi/docker/ssl",
		"LLM_SERVER_CONFIG_PATH":    "/opt/pentagi/conf/custom.provider.yml",
		"OLLAMA_SERVER_CONFIG_PATH": "/opt/pentagi/conf/ollama.provider.yml",
		"BEDROCK_CONFIG_PATH":       "/opt/pentagi/conf/bedrock.provider.yml",
	}
	tests := []struct {
		name string
		vars []string
	}{
		{"a docker certificates directory", []string{"DOCKER_CERT_PATH"}},
		{"an llm config file", []string{"LLM_SERVER_CONFIG_PATH"}},
		{"an ollama config file", []string{"OLLAMA_SERVER_CONFIG_PATH"}},
		{"a bedrock config file", []string{"BEDROCK_CONFIG_PATH"}},
		{"all four at once", []string{
			"DOCKER_CERT_PATH", "LLM_SERVER_CONFIG_PATH", "OLLAMA_SERVER_CONFIG_PATH", "BEDROCK_CONFIG_PATH",
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			st := &mockState{vars: map[string]loader.EnvVar{}}
			for _, name := range tt.vars {
				st.vars[name] = loader.EnvVar{Name: name, Value: host[name], Line: 1}
			}

			if err := DoMigrateSettings(st); err != nil {
				t.Fatalf("DoMigrateSettings: %v", err)
			}

			for _, name := range tt.vars {
				if got := st.vars["PENTAGI_"+name].Value; got != host[name] {
					t.Errorf("PENTAGI_%s = %q, want the host path %q", name, got, host[name])
				}
				if got := st.vars[name].Value; got != container[name] {
					t.Errorf("%s = %q, want the container path %q", name, got, container[name])
				}
			}
			if len(st.vars) != 2*len(tt.vars) {
				t.Errorf("state holds %d vars, want %d: %v", len(st.vars), 2*len(tt.vars), st.vars)
			}
		})
	}
}

func TestMigrations_DoMigrateSettings_LeavesAPathItCannotMigrate(t *testing.T) {
	dir, file := hardeningHostPaths(t)
	tests := []struct {
		name, variable, value string
	}{
		{"nothing set", "", ""},
		{"an empty docker certificates path", "DOCKER_CERT_PATH", ""},
		{"an empty llm config path", "LLM_SERVER_CONFIG_PATH", ""},
		{"an empty ollama config path", "OLLAMA_SERVER_CONFIG_PATH", ""},
		{"an empty bedrock config path", "BEDROCK_CONFIG_PATH", ""},
		{"a docker certificates directory that does not exist", "DOCKER_CERT_PATH", "/nonexistent/docker/certs"},
		{"an llm config file that does not exist", "LLM_SERVER_CONFIG_PATH", "/nonexistent/custom.provider.yml"},
		{"an ollama config file that does not exist", "OLLAMA_SERVER_CONFIG_PATH", "/nonexistent/ollama.provider.yml"},
		{"a bedrock config file that does not exist", "BEDROCK_CONFIG_PATH", "/opt/pentagi/conf/bedrock.yml"},
		{"a file where the docker certificates directory belongs", "DOCKER_CERT_PATH", file},
		{"a directory where the llm config file belongs", "LLM_SERVER_CONFIG_PATH", dir},
		{"a directory where the ollama config file belongs", "OLLAMA_SERVER_CONFIG_PATH", dir},
		{"a directory where the bedrock config file belongs", "BEDROCK_CONFIG_PATH", dir},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			st := &mockState{vars: map[string]loader.EnvVar{}}
			if tt.variable != "" {
				st.vars[tt.variable] = loader.EnvVar{Name: tt.variable, Value: tt.value, Line: 1}
			}

			if err := DoMigrateSettings(st); err != nil {
				t.Fatalf("DoMigrateSettings: %v", err)
			}

			for _, name := range []string{
				"DOCKER_CERT_PATH", "LLM_SERVER_CONFIG_PATH", "OLLAMA_SERVER_CONFIG_PATH", "BEDROCK_CONFIG_PATH",
			} {
				if got, set := st.vars["PENTAGI_"+name]; set {
					t.Errorf("PENTAGI_%s = %q, want it unset", name, got.Value)
				}
			}
			if got := st.vars[tt.variable]; tt.variable != "" && (got.Value != tt.value || got.IsChanged) {
				t.Errorf("%s = %q (changed %t), want %q untouched", tt.variable, got.Value, got.IsChanged, tt.value)
			}
		})
	}
}

func TestMigrations_DoMigrateSettings_ReturnsTheStateError(t *testing.T) {
	dir, file := hardeningHostPaths(t)
	tests := []struct {
		variable, path, failing string
	}{
		{"DOCKER_CERT_PATH", dir, "PENTAGI_DOCKER_CERT_PATH"},
		{"DOCKER_CERT_PATH", dir, "DOCKER_CERT_PATH"},
		{"LLM_SERVER_CONFIG_PATH", file, "PENTAGI_LLM_SERVER_CONFIG_PATH"},
		{"LLM_SERVER_CONFIG_PATH", file, "LLM_SERVER_CONFIG_PATH"},
		{"OLLAMA_SERVER_CONFIG_PATH", file, "PENTAGI_OLLAMA_SERVER_CONFIG_PATH"},
		{"OLLAMA_SERVER_CONFIG_PATH", file, "OLLAMA_SERVER_CONFIG_PATH"},
		{"BEDROCK_CONFIG_PATH", file, "PENTAGI_BEDROCK_CONFIG_PATH"},
		{"BEDROCK_CONFIG_PATH", file, "BEDROCK_CONFIG_PATH"},
	}

	for _, tt := range tests {
		st := &mockState{
			vars:        map[string]loader.EnvVar{tt.variable: {Name: tt.variable, Value: tt.path, Line: 1}},
			setVarError: map[string]error{tt.failing: errMocked},
		}

		if err := DoMigrateSettings(st); !errors.Is(err, errMocked) {
			t.Errorf("writing %s fails: err = %v, want %v", tt.failing, err, errMocked)
		}
	}
}
