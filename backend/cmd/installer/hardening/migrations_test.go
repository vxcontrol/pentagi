package hardening

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"pentagi/cmd/installer/loader"
	"pentagi/cmd/installer/wizard/controller"
)

// Test 1: Successful migrations for all variables
func TestDoMigrateSettings_SuccessfulMigrations(t *testing.T) {
	tests := []struct {
		name            string
		setupFunc       func(*testing.T) (string, func())
		varName         string
		pentagiVarName  string
		defaultPath     string
		pathType        checkPathType
		customPath      string
		expectMigration bool
	}{
		{
			name: "migrate DOCKER_CERT_PATH to PENTAGI_DOCKER_CERT_PATH",
			setupFunc: func(t *testing.T) (string, func()) {
				tmpDir, err := os.MkdirTemp("", "docker-certs-*")
				if err != nil {
					t.Fatalf("Failed to create temp dir: %v", err)
				}
				return tmpDir, func() { os.RemoveAll(tmpDir) }
			},
			varName:         "DOCKER_CERT_PATH",
			pentagiVarName:  "PENTAGI_DOCKER_CERT_PATH",
			defaultPath:     controller.DefaultDockerCertPath,
			pathType:        directory,
			expectMigration: true,
		},
		{
			name: "migrate LLM_SERVER_CONFIG_PATH to PENTAGI_LLM_SERVER_CONFIG_PATH",
			setupFunc: func(t *testing.T) (string, func()) {
				tmpFile, err := os.CreateTemp("", "custom-*.yml")
				if err != nil {
					t.Fatalf("Failed to create temp file: %v", err)
				}
				tmpFile.Close()
				return tmpFile.Name(), func() { os.Remove(tmpFile.Name()) }
			},
			varName:         "LLM_SERVER_CONFIG_PATH",
			pentagiVarName:  "PENTAGI_LLM_SERVER_CONFIG_PATH",
			defaultPath:     controller.DefaultCustomConfigsPath,
			pathType:        file,
			expectMigration: true,
		},
		{
			name: "migrate OLLAMA_SERVER_CONFIG_PATH to PENTAGI_OLLAMA_SERVER_CONFIG_PATH",
			setupFunc: func(t *testing.T) (string, func()) {
				tmpFile, err := os.CreateTemp("", "ollama-*.yml")
				if err != nil {
					t.Fatalf("Failed to create temp file: %v", err)
				}
				tmpFile.Close()
				return tmpFile.Name(), func() { os.Remove(tmpFile.Name()) }
			},
			varName:         "OLLAMA_SERVER_CONFIG_PATH",
			pentagiVarName:  "PENTAGI_OLLAMA_SERVER_CONFIG_PATH",
			defaultPath:     controller.DefaultOllamaConfigsPath,
			pathType:        file,
			expectMigration: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// setup temporary path
			customPath, cleanup := tt.setupFunc(t)
			defer cleanup()

			// create mock state with custom path set
			mockSt := &mockState{
				vars: map[string]loader.EnvVar{
					tt.varName: {
						Name:      tt.varName,
						Value:     customPath,
						Line:      1,
						IsChanged: false,
					},
				},
			}

			// execute migration
			err := DoMigrateSettings(mockSt)
			if err != nil {
				t.Fatalf("DoMigrateSettings() unexpected error = %v", err)
			}

			// verify migration occurred
			if tt.expectMigration {
				// check that PENTAGI_* variable was set to custom path
				pentagiVar, exists := mockSt.GetVar(tt.pentagiVarName)
				if !exists {
					t.Errorf("Expected %s to be set", tt.pentagiVarName)
				} else if pentagiVar.Value != customPath {
					t.Errorf("Expected %s = %q, got %q", tt.pentagiVarName, customPath, pentagiVar.Value)
				}

				// check that original variable was set to default path
				originalVar, exists := mockSt.GetVar(tt.varName)
				if !exists {
					t.Errorf("Expected %s to be set", tt.varName)
				} else if originalVar.Value != tt.defaultPath {
					t.Errorf("Expected %s = %q, got %q", tt.varName, tt.defaultPath, originalVar.Value)
				}
			}
		})
	}
}

// Test 2: No migration when variable is not set
func TestDoMigrateSettings_VariableNotSet(t *testing.T) {
	tests := []struct {
		name           string
		pentagiVarName string
	}{
		{
			name:           "DOCKER_CERT_PATH not set",
			pentagiVarName: "PENTAGI_DOCKER_CERT_PATH",
		},
		{
			name:           "LLM_SERVER_CONFIG_PATH not set",
			pentagiVarName: "PENTAGI_LLM_SERVER_CONFIG_PATH",
		},
		{
			name:           "OLLAMA_SERVER_CONFIG_PATH not set",
			pentagiVarName: "PENTAGI_OLLAMA_SERVER_CONFIG_PATH",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// create mock state with no variables set
			mockSt := &mockState{
				vars: make(map[string]loader.EnvVar),
			}

			// execute migration
			err := DoMigrateSettings(mockSt)
			if err != nil {
				t.Fatalf("DoMigrateSettings() unexpected error = %v", err)
			}

			// verify no migration occurred
			_, exists := mockSt.GetVar(tt.pentagiVarName)
			if exists {
				t.Errorf("Expected %s to not be set", tt.pentagiVarName)
			}
		})
	}
}

// Test 3: No migration when variable is empty
func TestDoMigrateSettings_EmptyVariable(t *testing.T) {
	tests := []struct {
		name           string
		varName        string
		pentagiVarName string
	}{
		{
			name:           "DOCKER_CERT_PATH is empty",
			varName:        "DOCKER_CERT_PATH",
			pentagiVarName: "PENTAGI_DOCKER_CERT_PATH",
		},
		{
			name:           "LLM_SERVER_CONFIG_PATH is empty",
			varName:        "LLM_SERVER_CONFIG_PATH",
			pentagiVarName: "PENTAGI_LLM_SERVER_CONFIG_PATH",
		},
		{
			name:           "OLLAMA_SERVER_CONFIG_PATH is empty",
			varName:        "OLLAMA_SERVER_CONFIG_PATH",
			pentagiVarName: "PENTAGI_OLLAMA_SERVER_CONFIG_PATH",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// create mock state with empty variable
			mockSt := &mockState{
				vars: map[string]loader.EnvVar{
					tt.varName: {
						Name:      tt.varName,
						Value:     "",
						Line:      1,
						IsChanged: false,
					},
				},
			}

			// execute migration
			err := DoMigrateSettings(mockSt)
			if err != nil {
				t.Fatalf("DoMigrateSettings() unexpected error = %v", err)
			}

			// verify no migration occurred
			_, exists := mockSt.GetVar(tt.pentagiVarName)
			if exists {
				t.Errorf("Expected %s to not be set", tt.pentagiVarName)
			}
		})
	}
}

// Test 4: No migration when path doesn't exist
func TestDoMigrateSettings_PathNotExist(t *testing.T) {
	tests := []struct {
		name           string
		varName        string
		pentagiVarName string
		nonExistPath   string
	}{
		{
			name:           "DOCKER_CERT_PATH points to non-existing directory",
			varName:        "DOCKER_CERT_PATH",
			pentagiVarName: "PENTAGI_DOCKER_CERT_PATH",
			nonExistPath:   "/nonexistent/docker/certs",
		},
		{
			name:           "LLM_SERVER_CONFIG_PATH points to non-existing file",
			varName:        "LLM_SERVER_CONFIG_PATH",
			pentagiVarName: "PENTAGI_LLM_SERVER_CONFIG_PATH",
			nonExistPath:   "/nonexistent/custom.provider.yml",
		},
		{
			name:           "OLLAMA_SERVER_CONFIG_PATH points to non-existing file",
			varName:        "OLLAMA_SERVER_CONFIG_PATH",
			pentagiVarName: "PENTAGI_OLLAMA_SERVER_CONFIG_PATH",
			nonExistPath:   "/nonexistent/ollama.provider.yml",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// create mock state with non-existing path
			mockSt := &mockState{
				vars: map[string]loader.EnvVar{
					tt.varName: {
						Name:      tt.varName,
						Value:     tt.nonExistPath,
						Line:      1,
						IsChanged: false,
					},
				},
			}

			// execute migration
			err := DoMigrateSettings(mockSt)
			if err != nil {
				t.Fatalf("DoMigrateSettings() unexpected error = %v", err)
			}

			// verify no migration occurred
			_, exists := mockSt.GetVar(tt.pentagiVarName)
			if exists {
				t.Errorf("Expected %s to not be set for non-existing path", tt.pentagiVarName)
			}
		})
	}
}

// Test 5: No migration when variable already has default container path value
func TestDoMigrateSettings_AlreadyDefaultValue(t *testing.T) {
	tests := []struct {
		name           string
		varName        string
		pentagiVarName string
		defaultPath    string
		description    string
	}{
		{
			name:           "DOCKER_CERT_PATH already has default container path",
			varName:        "DOCKER_CERT_PATH",
			pentagiVarName: "PENTAGI_DOCKER_CERT_PATH",
			defaultPath:    controller.DefaultDockerCertPath,
			description:    "Default container path should not be migrated",
		},
		{
			name:           "LLM_SERVER_CONFIG_PATH already has default container path",
			varName:        "LLM_SERVER_CONFIG_PATH",
			pentagiVarName: "PENTAGI_LLM_SERVER_CONFIG_PATH",
			defaultPath:    controller.DefaultCustomConfigsPath,
			description:    "Default container path should not be migrated",
		},
		{
			name:           "OLLAMA_SERVER_CONFIG_PATH already has default container path",
			varName:        "OLLAMA_SERVER_CONFIG_PATH",
			pentagiVarName: "PENTAGI_OLLAMA_SERVER_CONFIG_PATH",
			defaultPath:    controller.DefaultOllamaConfigsPath,
			description:    "Default container path should not be migrated",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// create mock state with default path
			mockSt := &mockState{
				vars: map[string]loader.EnvVar{
					tt.varName: {
						Name:      tt.varName,
						Value:     tt.defaultPath,
						Line:      1,
						IsChanged: false,
					},
				},
			}

			// execute migration
			err := DoMigrateSettings(mockSt)
			if err != nil {
				t.Fatalf("DoMigrateSettings() unexpected error = %v", err)
			}

			// verify no migration occurred
			_, exists := mockSt.GetVar(tt.pentagiVarName)
			if exists {
				t.Errorf("Expected %s to not be set when already using default", tt.pentagiVarName)
			}

			// verify original variable was not changed
			originalVar, exists := mockSt.GetVar(tt.varName)
			if !exists {
				t.Errorf("Expected %s to still exist", tt.varName)
			} else if originalVar.Value != tt.defaultPath {
				t.Errorf("Expected %s to remain %q, got %q", tt.varName, tt.defaultPath, originalVar.Value)
			}
		})
	}
}

// Test 6: No migration for embedded LLM configs
func TestDoMigrateSettings_EmbeddedConfigs(t *testing.T) {
	tests := []struct {
		name           string
		varName        string
		pentagiVarName string
		embeddedPath   string
		description    string
	}{
		{
			name:           "LLM_SERVER_CONFIG_PATH with embedded config should not migrate",
			varName:        "LLM_SERVER_CONFIG_PATH",
			pentagiVarName: "PENTAGI_LLM_SERVER_CONFIG_PATH",
			embeddedPath:   "/opt/pentagi/conf/llms/openai.yml",
			description:    "Embedded configs are inside docker image, no migration needed",
		},
		{
			name:           "OLLAMA_SERVER_CONFIG_PATH with embedded config should not migrate",
			varName:        "OLLAMA_SERVER_CONFIG_PATH",
			pentagiVarName: "PENTAGI_OLLAMA_SERVER_CONFIG_PATH",
			embeddedPath:   "/opt/pentagi/conf/llms/llama3.yml",
			description:    "Embedded configs are inside docker image, no migration needed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// create mock state with embedded config path
			mockSt := &mockState{
				vars: map[string]loader.EnvVar{
					tt.varName: {
						Name:      tt.varName,
						Value:     tt.embeddedPath,
						Line:      1,
						IsChanged: false,
					},
				},
			}

			// execute migration
			err := DoMigrateSettings(mockSt)
			if err != nil {
				t.Fatalf("DoMigrateSettings() unexpected error = %v", err)
			}

			// verify no migration occurred
			pentagiVar, exists := mockSt.GetVar(tt.pentagiVarName)
			if exists && pentagiVar.Value != "" {
				t.Errorf("Expected %s to not be set for embedded config: %s", tt.pentagiVarName, tt.description)
			}

			// verify original variable was not changed
			originalVar, exists := mockSt.GetVar(tt.varName)
			if !exists {
				t.Errorf("Expected %s to still exist", tt.varName)
			} else if originalVar.Value != tt.embeddedPath {
				t.Errorf("Expected %s to remain %q, got %q: %s", tt.varName, tt.embeddedPath, originalVar.Value, tt.description)
			}
		})
	}
}

// Test 7: Wrong path type (file instead of directory and vice versa)
func TestDoMigrateSettings_WrongPathType(t *testing.T) {
	tests := []struct {
		name           string
		setupFunc      func(*testing.T) (string, func())
		varName        string
		pentagiVarName string
		description    string
	}{
		{
			name: "DOCKER_CERT_PATH points to file instead of directory",
			setupFunc: func(t *testing.T) (string, func()) {
				tmpFile, err := os.CreateTemp("", "docker-cert-*")
				if err != nil {
					t.Fatalf("Failed to create temp file: %v", err)
				}
				tmpFile.Close()
				return tmpFile.Name(), func() { os.Remove(tmpFile.Name()) }
			},
			varName:        "DOCKER_CERT_PATH",
			pentagiVarName: "PENTAGI_DOCKER_CERT_PATH",
			description:    "File provided when directory expected",
		},
		{
			name: "LLM_SERVER_CONFIG_PATH points to directory instead of file",
			setupFunc: func(t *testing.T) (string, func()) {
				tmpDir, err := os.MkdirTemp("", "llm-config-*")
				if err != nil {
					t.Fatalf("Failed to create temp dir: %v", err)
				}
				return tmpDir, func() { os.RemoveAll(tmpDir) }
			},
			varName:        "LLM_SERVER_CONFIG_PATH",
			pentagiVarName: "PENTAGI_LLM_SERVER_CONFIG_PATH",
			description:    "Directory provided when file expected",
		},
		{
			name: "OLLAMA_SERVER_CONFIG_PATH points to directory instead of file",
			setupFunc: func(t *testing.T) (string, func()) {
				tmpDir, err := os.MkdirTemp("", "ollama-config-*")
				if err != nil {
					t.Fatalf("Failed to create temp dir: %v", err)
				}
				return tmpDir, func() { os.RemoveAll(tmpDir) }
			},
			varName:        "OLLAMA_SERVER_CONFIG_PATH",
			pentagiVarName: "PENTAGI_OLLAMA_SERVER_CONFIG_PATH",
			description:    "Directory provided when file expected",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// setup wrong path type
			wrongPath, cleanup := tt.setupFunc(t)
			defer cleanup()

			// create mock state with wrong path type
			mockSt := &mockState{
				vars: map[string]loader.EnvVar{
					tt.varName: {
						Name:      tt.varName,
						Value:     wrongPath,
						Line:      1,
						IsChanged: false,
					},
				},
			}

			// execute migration
			err := DoMigrateSettings(mockSt)
			if err != nil {
				t.Fatalf("DoMigrateSettings() unexpected error = %v", err)
			}

			// verify no migration occurred
			_, exists := mockSt.GetVar(tt.pentagiVarName)
			if exists {
				t.Errorf("Expected %s to not be set for wrong path type: %s", tt.pentagiVarName, tt.description)
			}
		})
	}
}

// Test 8: Error handling scenarios
func TestDoMigrateSettings_ErrorHandling(t *testing.T) {
	tests := []struct {
		name          string
		setupFunc     func(*testing.T) (*mockStateWithErrors, string, func())
		expectedError string
	}{
		{
			name: "SetVar error for PENTAGI_DOCKER_CERT_PATH",
			setupFunc: func(t *testing.T) (*mockStateWithErrors, string, func()) {
				tmpDir, err := os.MkdirTemp("", "docker-certs-*")
				if err != nil {
					t.Fatalf("Failed to create temp dir: %v", err)
				}
				mockSt := &mockStateWithErrors{
					vars: map[string]loader.EnvVar{
						"DOCKER_CERT_PATH": {
							Name:  "DOCKER_CERT_PATH",
							Value: tmpDir,
							Line:  1,
						},
					},
					setVarError: map[string]error{
						"PENTAGI_DOCKER_CERT_PATH": mockError,
					},
				}
				return mockSt, tmpDir, func() { os.RemoveAll(tmpDir) }
			},
			expectedError: "mocked error",
		},
		{
			name: "SetVar error for DOCKER_CERT_PATH",
			setupFunc: func(t *testing.T) (*mockStateWithErrors, string, func()) {
				tmpDir, err := os.MkdirTemp("", "docker-certs-*")
				if err != nil {
					t.Fatalf("Failed to create temp dir: %v", err)
				}
				mockSt := &mockStateWithErrors{
					vars: map[string]loader.EnvVar{
						"DOCKER_CERT_PATH": {
							Name:  "DOCKER_CERT_PATH",
							Value: tmpDir,
							Line:  1,
						},
					},
					setVarError: map[string]error{
						"DOCKER_CERT_PATH": mockError,
					},
				}
				return mockSt, tmpDir, func() { os.RemoveAll(tmpDir) }
			},
			expectedError: "mocked error",
		},
		{
			name: "SetVar error for PENTAGI_LLM_SERVER_CONFIG_PATH",
			setupFunc: func(t *testing.T) (*mockStateWithErrors, string, func()) {
				tmpFile, err := os.CreateTemp("", "custom-*.yml")
				if err != nil {
					t.Fatalf("Failed to create temp file: %v", err)
				}
				tmpFile.Close()
				mockSt := &mockStateWithErrors{
					vars: map[string]loader.EnvVar{
						"LLM_SERVER_CONFIG_PATH": {
							Name:  "LLM_SERVER_CONFIG_PATH",
							Value: tmpFile.Name(),
							Line:  1,
						},
					},
					setVarError: map[string]error{
						"PENTAGI_LLM_SERVER_CONFIG_PATH": mockError,
					},
				}
				return mockSt, tmpFile.Name(), func() { os.Remove(tmpFile.Name()) }
			},
			expectedError: "mocked error",
		},
		{
			name: "SetVar error for PENTAGI_OLLAMA_SERVER_CONFIG_PATH",
			setupFunc: func(t *testing.T) (*mockStateWithErrors, string, func()) {
				tmpFile, err := os.CreateTemp("", "ollama-*.yml")
				if err != nil {
					t.Fatalf("Failed to create temp file: %v", err)
				}
				tmpFile.Close()
				mockSt := &mockStateWithErrors{
					vars: map[string]loader.EnvVar{
						"OLLAMA_SERVER_CONFIG_PATH": {
							Name:  "OLLAMA_SERVER_CONFIG_PATH",
							Value: tmpFile.Name(),
							Line:  1,
						},
					},
					setVarError: map[string]error{
						"PENTAGI_OLLAMA_SERVER_CONFIG_PATH": mockError,
					},
				}
				return mockSt, tmpFile.Name(), func() { os.Remove(tmpFile.Name()) }
			},
			expectedError: "mocked error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// setup mock state with error condition
			mockSt, _, cleanup := tt.setupFunc(t)
			defer cleanup()

			// execute migration
			err := DoMigrateSettings(mockSt)

			// verify error was returned
			if err == nil {
				t.Error("Expected error but got none")
			} else if err.Error() != tt.expectedError {
				t.Errorf("Expected error %q, got %q", tt.expectedError, err.Error())
			}
		})
	}
}

// Test 9: Combined migrations scenario
func TestDoMigrateSettings_CombinedMigrations(t *testing.T) {
	tests := []struct {
		name         string
		setupFunc    func(*testing.T) (map[string]string, func())
		expectedVars map[string]string
		description  string
	}{
		{
			name: "migrate all three variables at once",
			setupFunc: func(t *testing.T) (map[string]string, func()) {
				// create temp directory for docker certs
				dockerCertDir, err := os.MkdirTemp("", "docker-certs-*")
				if err != nil {
					t.Fatalf("Failed to create temp dir: %v", err)
				}

				// create temp file for LLM config
				llmConfigFile, err := os.CreateTemp("", "custom-*.yml")
				if err != nil {
					os.RemoveAll(dockerCertDir)
					t.Fatalf("Failed to create temp file: %v", err)
				}
				llmConfigFile.Close()

				// create temp file for Ollama config
				ollamaConfigFile, err := os.CreateTemp("", "ollama-*.yml")
				if err != nil {
					os.RemoveAll(dockerCertDir)
					os.Remove(llmConfigFile.Name())
					t.Fatalf("Failed to create temp file: %v", err)
				}
				ollamaConfigFile.Close()

				paths := map[string]string{
					"DOCKER_CERT_PATH":          dockerCertDir,
					"LLM_SERVER_CONFIG_PATH":    llmConfigFile.Name(),
					"OLLAMA_SERVER_CONFIG_PATH": ollamaConfigFile.Name(),
				}

				cleanup := func() {
					os.RemoveAll(dockerCertDir)
					os.Remove(llmConfigFile.Name())
					os.Remove(ollamaConfigFile.Name())
				}

				return paths, cleanup
			},
			expectedVars: map[string]string{
				"DOCKER_CERT_PATH":          controller.DefaultDockerCertPath,
				"LLM_SERVER_CONFIG_PATH":    controller.DefaultCustomConfigsPath,
				"OLLAMA_SERVER_CONFIG_PATH": controller.DefaultOllamaConfigsPath,
				// PENTAGI_* vars will be checked separately as they contain dynamic temp paths
			},
			description: "All three migrations should complete successfully",
		},
		{
			name: "migrate only DOCKER_CERT_PATH, others are default",
			setupFunc: func(t *testing.T) (map[string]string, func()) {
				dockerCertDir, err := os.MkdirTemp("", "docker-certs-*")
				if err != nil {
					t.Fatalf("Failed to create temp dir: %v", err)
				}

				paths := map[string]string{
					"DOCKER_CERT_PATH":          dockerCertDir,
					"LLM_SERVER_CONFIG_PATH":    controller.DefaultCustomConfigsPath,
					"OLLAMA_SERVER_CONFIG_PATH": controller.DefaultOllamaConfigsPath,
				}

				cleanup := func() {
					os.RemoveAll(dockerCertDir)
				}

				return paths, cleanup
			},
			expectedVars: map[string]string{
				"DOCKER_CERT_PATH":          controller.DefaultDockerCertPath,
				"LLM_SERVER_CONFIG_PATH":    controller.DefaultCustomConfigsPath,
				"OLLAMA_SERVER_CONFIG_PATH": controller.DefaultOllamaConfigsPath,
			},
			description: "Only DOCKER_CERT_PATH should be migrated",
		},
		{
			name: "migrate only config paths, DOCKER_CERT_PATH is default",
			setupFunc: func(t *testing.T) (map[string]string, func()) {
				// create temp file for LLM config
				llmConfigFile, err := os.CreateTemp("", "custom-*.yml")
				if err != nil {
					t.Fatalf("Failed to create temp file: %v", err)
				}
				llmConfigFile.Close()

				// create temp file for Ollama config
				ollamaConfigFile, err := os.CreateTemp("", "ollama-*.yml")
				if err != nil {
					os.Remove(llmConfigFile.Name())
					t.Fatalf("Failed to create temp file: %v", err)
				}
				ollamaConfigFile.Close()

				paths := map[string]string{
					"DOCKER_CERT_PATH":          controller.DefaultDockerCertPath,
					"LLM_SERVER_CONFIG_PATH":    llmConfigFile.Name(),
					"OLLAMA_SERVER_CONFIG_PATH": ollamaConfigFile.Name(),
				}

				cleanup := func() {
					os.Remove(llmConfigFile.Name())
					os.Remove(ollamaConfigFile.Name())
				}

				return paths, cleanup
			},
			expectedVars: map[string]string{
				"DOCKER_CERT_PATH":          controller.DefaultDockerCertPath,
				"LLM_SERVER_CONFIG_PATH":    controller.DefaultCustomConfigsPath,
				"OLLAMA_SERVER_CONFIG_PATH": controller.DefaultOllamaConfigsPath,
			},
			description: "Only config paths should be migrated",
		},
		{
			name: "no migration for embedded configs",
			setupFunc: func(t *testing.T) (map[string]string, func()) {
				// create temp directory for docker certs
				dockerCertDir, err := os.MkdirTemp("", "docker-certs-*")
				if err != nil {
					t.Fatalf("Failed to create temp dir: %v", err)
				}

				paths := map[string]string{
					"DOCKER_CERT_PATH":          dockerCertDir,
					"LLM_SERVER_CONFIG_PATH":    "/opt/pentagi/conf/llms/openai.yml", // embedded config
					"OLLAMA_SERVER_CONFIG_PATH": "/opt/pentagi/conf/llms/llama3.yml", // embedded config
				}

				cleanup := func() {
					os.RemoveAll(dockerCertDir)
				}

				return paths, cleanup
			},
			expectedVars: map[string]string{
				"DOCKER_CERT_PATH":          controller.DefaultDockerCertPath,
				"LLM_SERVER_CONFIG_PATH":    "/opt/pentagi/conf/llms/openai.yml", // should not change
				"OLLAMA_SERVER_CONFIG_PATH": "/opt/pentagi/conf/llms/llama3.yml", // should not change
			},
			description: "Embedded configs should not be migrated, only DOCKER_CERT_PATH",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// setup paths and mock state
			paths, cleanup := tt.setupFunc(t)
			defer cleanup()

			mockSt := &mockState{
				vars: make(map[string]loader.EnvVar),
			}

			// populate mock state with initial values
			for varName, varValue := range paths {
				mockSt.vars[varName] = loader.EnvVar{
					Name:      varName,
					Value:     varValue,
					Line:      1,
					IsChanged: false,
				}
			}

			// execute migration
			err := DoMigrateSettings(mockSt)
			if err != nil {
				t.Fatalf("DoMigrateSettings() unexpected error = %v", err)
			}

			// verify all expected variables
			for varName, expectedValue := range tt.expectedVars {
				actualVar, exists := mockSt.GetVar(varName)
				if !exists {
					t.Errorf("Expected %s to be set", varName)
				} else if actualVar.Value != expectedValue {
					t.Errorf("Expected %s = %q, got %q", varName, expectedValue, actualVar.Value)
				}
			}

			// verify PENTAGI_* variables were set correctly for non-default and non-embedded values
			for varName, originalValue := range paths {
				pentagiVarName := ""
				defaultValue := ""
				isEmbedded := false

				switch varName {
				case "DOCKER_CERT_PATH":
					pentagiVarName = "PENTAGI_DOCKER_CERT_PATH"
					defaultValue = controller.DefaultDockerCertPath
				case "LLM_SERVER_CONFIG_PATH":
					pentagiVarName = "PENTAGI_LLM_SERVER_CONFIG_PATH"
					defaultValue = controller.DefaultCustomConfigsPath
					// check if it's an embedded config path
					isEmbedded = strings.HasPrefix(originalValue, "/opt/pentagi/conf/llms/")
				case "OLLAMA_SERVER_CONFIG_PATH":
					pentagiVarName = "PENTAGI_OLLAMA_SERVER_CONFIG_PATH"
					defaultValue = controller.DefaultOllamaConfigsPath
					// check if it's an embedded config path
					isEmbedded = strings.HasPrefix(originalValue, "/opt/pentagi/conf/llms/")
				}

				// migration should only occur for non-default, non-embedded, existing files
				shouldMigrate := originalValue != defaultValue && !isEmbedded

				if shouldMigrate {
					// check if file exists on host (migration only happens for existing files)
					_, err := os.Stat(originalValue)
					if err == nil {
						// migration should have occurred
						pentagiVar, exists := mockSt.GetVar(pentagiVarName)
						if !exists {
							t.Errorf("Expected %s to be set for non-default value", pentagiVarName)
						} else if pentagiVar.Value != originalValue {
							t.Errorf("Expected %s = %q, got %q", pentagiVarName, originalValue, pentagiVar.Value)
						}
					}
				} else {
					// migration should not have occurred
					pentagiVar, exists := mockSt.GetVar(pentagiVarName)
					if exists && pentagiVar.Value != "" {
						t.Errorf("Expected %s to not be set for default/embedded value, but got %q", pentagiVarName, pentagiVar.Value)
					}
				}
			}
		})
	}
}

// Test 10: checkPathInHostFS function
func TestCheckPathInHostFS(t *testing.T) {
	tests := []struct {
		name       string
		setupFunc  func(*testing.T) (string, func())
		pathType   checkPathType
		expectTrue bool
	}{
		{
			name: "valid directory returns true for directory type",
			setupFunc: func(t *testing.T) (string, func()) {
				tmpDir, err := os.MkdirTemp("", "test-dir-*")
				if err != nil {
					t.Fatalf("Failed to create temp dir: %v", err)
				}
				return tmpDir, func() { os.RemoveAll(tmpDir) }
			},
			pathType:   directory,
			expectTrue: true,
		},
		{
			name: "valid file returns false for directory type",
			setupFunc: func(t *testing.T) (string, func()) {
				tmpFile, err := os.CreateTemp("", "test-file-*")
				if err != nil {
					t.Fatalf("Failed to create temp file: %v", err)
				}
				tmpFile.Close()
				return tmpFile.Name(), func() { os.Remove(tmpFile.Name()) }
			},
			pathType:   directory,
			expectTrue: false,
		},
		{
			name: "valid file returns true for file type",
			setupFunc: func(t *testing.T) (string, func()) {
				tmpFile, err := os.CreateTemp("", "test-file-*")
				if err != nil {
					t.Fatalf("Failed to create temp file: %v", err)
				}
				tmpFile.Close()
				return tmpFile.Name(), func() { os.Remove(tmpFile.Name()) }
			},
			pathType:   file,
			expectTrue: true,
		},
		{
			name: "valid directory returns false for file type",
			setupFunc: func(t *testing.T) (string, func()) {
				tmpDir, err := os.MkdirTemp("", "test-dir-*")
				if err != nil {
					t.Fatalf("Failed to create temp dir: %v", err)
				}
				return tmpDir, func() { os.RemoveAll(tmpDir) }
			},
			pathType:   file,
			expectTrue: false,
		},
		{
			name: "non-existent path returns false",
			setupFunc: func(t *testing.T) (string, func()) {
				tmpDir, err := os.MkdirTemp("", "test-dir-*")
				if err != nil {
					t.Fatalf("Failed to create temp dir: %v", err)
				}
				nonExistPath := filepath.Join(tmpDir, "nonexistent")
				return nonExistPath, func() { os.RemoveAll(tmpDir) }
			},
			pathType:   directory,
			expectTrue: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path, cleanup := tt.setupFunc(t)
			defer cleanup()

			result := checkPathInHostFS(path, tt.pathType)
			if result != tt.expectTrue {
				t.Errorf("checkPathInHostFS(%q, %v) = %v, want %v", path, tt.pathType, result, tt.expectTrue)
			}
		})
	}
}
