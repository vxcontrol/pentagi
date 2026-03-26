package main

import (
	"os"
	"path/filepath"
	"testing"

	"pentagi/pkg/version"
)

func TestParseFlags(t *testing.T) {
	tests := []struct {
		name            string
		args            []string
		expectedEnv     string
		expectedVersion bool
	}{
		{
			name:            "default values",
			args:            []string{},
			expectedEnv:     ".env",
			expectedVersion: false,
		},
		{
			name:            "custom env path",
			args:            []string{"-e", "config/.env"},
			expectedEnv:     "config/.env",
			expectedVersion: false,
		},
		{
			name:            "version flag",
			args:            []string{"-v"},
			expectedEnv:     ".env",
			expectedVersion: true,
		},
		{
			name:            "both flags",
			args:            []string{"-e", "test.env", "-v"},
			expectedEnv:     "test.env",
			expectedVersion: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := parseFlags(append([]string{"test"}, tt.args...))

			if config.envPath != tt.expectedEnv {
				t.Errorf("Expected envPath %s, got %s", tt.expectedEnv, config.envPath)
			}
			if config.showVersion != tt.expectedVersion {
				t.Errorf("Expected showVersion %v, got %v", tt.expectedVersion, config.showVersion)
			}
		})
	}
}

func TestValidateEnvPath(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name        string
		path        string
		setup       func() string
		expectError bool
	}{
		{
			name: "existing file",
			setup: func() string {
				path := filepath.Join(tmpDir, "existing.env")
				os.WriteFile(path, []byte("VAR=value"), 0644)
				return path
			},
			expectError: false,
		},
		{
			name: "non-existent file in existing directory",
			setup: func() string {
				return filepath.Join(tmpDir, "new.env")
			},
			expectError: false,
		},
		{
			name: "non-existent directory",
			setup: func() string {
				return filepath.Join(tmpDir, "nonexistent", "file.env")
			},
			expectError: false,
		},
		{
			name: "directory instead of file",
			setup: func() string {
				os.Mkdir(filepath.Join(tmpDir, "dir"), 0755)
				return filepath.Join(tmpDir, "dir")
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := tt.setup()

			result, err := validateEnvPath(path)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if result == "" {
					t.Error("Expected non-empty result path")
				}
				// Check that file exists after validation
				if _, err := os.Stat(result); os.IsNotExist(err) {
					t.Error("Expected file to exist after validation")
				}
			}
		})
	}
}

func TestCreateEmptyEnvFile(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "test.env")

	err := createInitialEnvFile(path)
	if err != nil {
		t.Fatalf("Failed to create empty env file: %v", err)
	}

	// Check file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Error("Expected file to be created")
	}

	// Check file content
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("Failed to read created file: %v", err)
	}

	contentStr := string(content)
	if !containsString(contentStr, "PentAGI Environment Configuration") {
		t.Error("Expected file to contain header comment")
	}
	if !containsString(contentStr, version.GetBinaryVersion()) {
		t.Error("Expected file to contain version")
	}
}

func TestInitializeState(t *testing.T) {
	tmpDir := t.TempDir()
	envPath := filepath.Join(tmpDir, "test.env")

	// Create test env file
	err := os.WriteFile(envPath, []byte("VAR1=value1"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test env file: %v", err)
	}

	state, err := initializeState(envPath)
	if err != nil {
		t.Fatalf("Failed to initialize state: %v", err)
	}

	if state == nil {
		t.Error("Expected non-nil state")
	}

	// Test that state can access variables
	envVar, exists := state.GetVar("VAR1")
	if !exists {
		t.Error("Expected VAR1 to exist in state")
	}
	if envVar.Value != "value1" {
		t.Errorf("Expected VAR1 value 'value1', got '%s'", envVar.Value)
	}
}

func containsString(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr ||
			containsString(s[1:], substr) ||
			(len(s) > 0 && s[:len(substr)] == substr))
}
