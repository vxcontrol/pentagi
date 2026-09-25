package loader

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestLoader_LoadEnvFile_ParsesAssignmentsAndCommentedOutVariables(t *testing.T) {
	type parsed struct {
		value   string
		comment bool
		changed bool // the line is not in canonical NAME=value form, so Save rewrites it
	}

	tests := []struct {
		name    string
		content string
		want    map[string]parsed // every variable the file itself defines
	}{
		{name: "empty file", content: "", want: map[string]parsed{}},
		{name: "comment only", content: "# Just a comment", want: map[string]parsed{}},
		{
			name:    "assignments between comments",
			content: "# Comment\nVAR1=value1\nVAR2=value2\n# Another comment\nVAR3=value3",
			want:    map[string]parsed{"VAR1": {value: "value1"}, "VAR2": {value: "value2"}, "VAR3": {value: "value3"}},
		},
		{name: "empty lines", content: "\n\n\nVAR1=value1\n\n", want: map[string]parsed{"VAR1": {value: "value1"}}},
		{
			name:    "commented-out variable",
			content: "#VAR1=commented\nVAR2=active",
			want:    map[string]parsed{"VAR1": {value: "commented", comment: true}, "VAR2": {value: "active"}},
		},
		{
			name:    "inline comment is stripped",
			content: "VAR1=value1 # comment\nVAR2=value2 # comment",
			want:    map[string]parsed{"VAR1": {value: "value1", changed: true}, "VAR2": {value: "value2", changed: true}},
		},
		{
			name:    "spaces around names and values",
			content: "VAR1 = value1\n  VAR2=value2  ",
			want:    map[string]parsed{"VAR1": {value: "value1", changed: true}, "VAR2": {value: "value2"}},
		},
		{
			name:    "equals signs inside values",
			content: "VAR1=value=with=equals\nVAR2=url=https://example.com",
			want:    map[string]parsed{"VAR1": {value: "value=with=equals"}, "VAR2": {value: "url=https://example.com"}},
		},
		{
			name:    "lines without an assignment are ignored",
			content: "invalid line\nVAR1=value1\nanother invalid",
			want:    map[string]parsed{"VAR1": {value: "value1"}},
		},
		{
			name: "comment prose containing an equals sign is not a variable",
			content: strings.Join([]string{
				"## Default: 1200 (20 minutes). Range: 1-10800 (up to 3 hours). Values <= 0 or above 10800 are clamped to 10800.",
				"TERMINAL_TOOL_TIMEOUT=1200",
				"",
				"## Go pprof HTTP listener (empty = disabled). Use host:port, e.g. :7777.",
				"PPROF_ADDR=",
				"",
				"## one worker node, one Neo4j/Graphiti, one Langfuse (empty = standalone instance).",
				"TENANT_ID=",
				"",
				"#REAL_COMMENTED_VAR=commented_value",
			}, "\n"),
			want: map[string]parsed{
				"TERMINAL_TOOL_TIMEOUT": {value: "1200"},
				"PPROF_ADDR":            {},
				"TENANT_ID":             {},
				"REAL_COMMENTED_VAR":    {value: "commented_value", comment: true},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			envFile, err := LoadEnvFile(loaderEnvFile(t, tt.content))
			if err != nil {
				t.Fatalf("LoadEnvFile() error = %v", err)
			}

			got := map[string]parsed{}
			for name, envVar := range envFile.GetAll() {
				if envVar.IsPresent() {
					got[name] = parsed{value: envVar.Value, comment: envVar.IsComment, changed: envVar.IsChanged}
				}
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("variables from the file = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestLoader_LoadEnvFile_RefusesAMissingFileOrADirectory(t *testing.T) {
	dir := t.TempDir()
	tests := []struct {
		name    string
		path    string
		wantErr string
	}{
		{name: "a missing file", path: filepath.Join(dir, "absent.env"), wantErr: "failed to stat"},
		{name: "a directory", path: dir, wantErr: "is a directory"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := LoadEnvFile(tt.path); err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("LoadEnvFile() error = %v, want it to mention %q", err, tt.wantErr)
			}
		})
	}
}

func TestLoader_LoadEnvFile_FillsInConfigDefaults(t *testing.T) {
	// Defaults are read through the process environment, so a developer's shell must not answer instead.
	for _, name := range []string{"SERVER_PORT", "DEBUG", "DATABASE_URL", "STATIC_URL"} {
		t.Setenv(name, "")
		os.Unsetenv(name)
	}

	envFile, err := LoadEnvFile(loaderEnvFile(t, "SERVER_PORT=9090\n"))
	if err != nil {
		t.Fatalf("LoadEnvFile() error = %v", err)
	}

	if port, _ := envFile.Get("SERVER_PORT"); port.Value != "9090" || port.Default != "8080" || !port.IsPresent() {
		t.Errorf("SERVER_PORT = %+v, want the file's 9090 with default 8080", port)
	}
	if debug, ok := envFile.Get("DEBUG"); !ok || debug.Default != "false" {
		t.Errorf("DEBUG = %+v, %v; want default false", debug, ok)
	}
	if db, ok := envFile.Get("DATABASE_URL"); !ok || db.Default != "postgres://pentagiuser:pentagipass@pgvector:5432/pentagidb?sslmode=disable" {
		t.Errorf("DATABASE_URL = %+v, %v; want the bundled database default", db, ok)
	}
	if _, ok := envFile.Get("STATIC_URL"); ok {
		t.Error("STATIC_URL declares no default and is not in the file, yet it was added")
	}

	for name, envVar := range envFile.GetAll() {
		if name != "SERVER_PORT" && (envVar.IsPresent() || envVar.Value != "" || envVar.IsChanged) {
			t.Errorf("%s = %+v; a default must stay absent, empty and unchanged", name, envVar)
		}
	}
}
