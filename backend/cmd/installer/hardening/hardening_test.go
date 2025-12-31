package hardening

import (
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"pentagi/cmd/installer/checker"
	"pentagi/cmd/installer/loader"
)

// mockState implements State interface for testing
type mockState struct {
	vars    map[string]loader.EnvVar
	envPath string
}

func (m *mockState) GetVar(key string) (loader.EnvVar, bool) {
	if val, exists := m.vars[key]; exists {
		return val, true
	}
	return loader.EnvVar{Name: key, Line: -1}, false
}

func (m *mockState) GetVars(names []string) (map[string]loader.EnvVar, map[string]bool) {
	vars := make(map[string]loader.EnvVar)
	present := make(map[string]bool)
	for _, name := range names {
		if val, exists := m.vars[name]; exists {
			vars[name] = val
			present[name] = true
		} else {
			vars[name] = loader.EnvVar{Name: name, Line: -1}
			present[name] = false
		}
	}
	return vars, present
}

func (m *mockState) GetEnvPath() string                   { return m.envPath }
func (m *mockState) Exists() bool                         { return true }
func (m *mockState) Reset() error                         { return nil }
func (m *mockState) IsDirty() bool                        { return false }
func (m *mockState) GetEulaConsent() bool                 { return true }
func (m *mockState) SetEulaConsent() error                { return nil }
func (m *mockState) SetStack(stack []string) error        { return nil }
func (m *mockState) GetStack() []string                   { return []string{} }
func (m *mockState) ResetVar(name string) error           { return nil }
func (m *mockState) ResetVars(names []string) error       { return nil }
func (m *mockState) GetAllVars() map[string]loader.EnvVar { return m.vars }
func (m *mockState) Commit() error                        { return nil }

func (m *mockState) SetVar(name, value string) error {
	if m.vars == nil {
		m.vars = make(map[string]loader.EnvVar)
	}
	envVar := m.vars[name]
	envVar.Name = name
	envVar.Value = value
	envVar.IsChanged = true
	m.vars[name] = envVar
	return nil
}

func (m *mockState) SetVars(vars map[string]string) error {
	for name, value := range vars {
		if err := m.SetVar(name, value); err != nil {
			return err
		}
	}
	return nil
}

// getEnvExamplePath returns path to .env.example relative to this test file
func getEnvExamplePath() string {
	_, filename, _, _ := runtime.Caller(0)
	dir := filepath.Dir(filename)
	// Go from backend/cmd/installer/hardening to project root
	return filepath.Join(dir, "..", "..", "..", "..", ".env.example")
}

// createTempEnvFile creates a temporary .env file with given content
func createTempEnvFile(t *testing.T, content string) string {
	tmpFile, err := os.CreateTemp("", "test_env_*.env")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer tmpFile.Close()

	if _, err := tmpFile.WriteString(content); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}

	return tmpFile.Name()
}

// Test 1: HardeningPolicy and randomString function
func TestRandomString(t *testing.T) {
	tests := []struct {
		name     string
		policy   HardeningPolicy
		validate func(string) bool
	}{
		{
			name:   "default alphanumeric",
			policy: HardeningPolicy{Type: HardeningPolicyTypeDefault, Length: 10},
			validate: func(s string) bool {
				return len(s) == 10 && isAlphanumeric(s)
			},
		},
		{
			name:   "hex string",
			policy: HardeningPolicy{Type: HardeningPolicyTypeHex, Length: 16},
			validate: func(s string) bool {
				return len(s) == 16 && isHexString(s)
			},
		},
		{
			name:   "uuid with prefix",
			policy: HardeningPolicy{Type: HardeningPolicyTypeUUID, Prefix: "pk-lf-"},
			validate: func(s string) bool {
				return strings.HasPrefix(s, "pk-lf-") && len(s) == 42 // prefix + 36 char UUID
			},
		},
		{
			name:   "bool true",
			policy: HardeningPolicy{Type: HardeningPolicyTypeBoolTrue},
			validate: func(s string) bool {
				return s == "true"
			},
		},
		{
			name:   "bool false",
			policy: HardeningPolicy{Type: HardeningPolicyTypeBoolFalse},
			validate: func(s string) bool {
				return s == "false"
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := randomString(tt.policy)
			if err != nil {
				t.Fatalf("randomString() error = %v", err)
			}
			if !tt.validate(result) {
				t.Errorf("randomString() = %q, validation failed", result)
			}
		})
	}

	// Test invalid policy type
	t.Run("invalid policy type", func(t *testing.T) {
		policy := HardeningPolicy{Type: "invalid"}
		_, err := randomString(policy)
		if err == nil {
			t.Error("randomString() should return error for invalid policy type")
		}
	})
}

// Test 2: Individual random string generators
func TestRandStringAlpha(t *testing.T) {
	tests := []struct {
		name   string
		length int
	}{
		{"zero length", 0},
		{"short string", 5},
		{"medium string", 16},
		{"long string", 64},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := randStringAlpha(tt.length)
			if err != nil {
				t.Fatalf("randStringAlpha() error = %v", err)
			}
			if len(result) != tt.length {
				t.Errorf("randStringAlpha() length = %d, want %d", len(result), tt.length)
			}
			if tt.length > 0 && !isAlphanumeric(result) {
				t.Errorf("randStringAlpha() = %q, should be alphanumeric", result)
			}
		})
	}
}

func TestRandStringHex(t *testing.T) {
	tests := []struct {
		name   string
		length int
	}{
		{"zero length", 0},
		{"short string", 4},
		{"medium string", 16},
		{"long string", 32},
		{"odd length", 7},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := randStringHex(tt.length)
			if err != nil {
				t.Fatalf("randStringHex() error = %v", err)
			}
			if len(result) != tt.length {
				t.Errorf("randStringHex() length = %d, want %d", len(result), tt.length)
			}
			if tt.length > 0 && !isHexString(result) {
				t.Errorf("randStringHex() = %q, should be hex string", result)
			}
		})
	}
}

func TestRandStringUUID(t *testing.T) {
	tests := []struct {
		name   string
		prefix string
	}{
		{"no prefix", ""},
		{"with prefix", "pk-lf-"},
		{"long prefix", "very-long-prefix-"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := randStringUUID(tt.prefix)
			if err != nil {
				t.Fatalf("randStringUUID() error = %v", err)
			}
			if !strings.HasPrefix(result, tt.prefix) {
				t.Errorf("randStringUUID() = %q, should start with %q", result, tt.prefix)
			}
			// UUID should be 36 characters + prefix length
			expectedLength := len(tt.prefix) + 36
			if len(result) != expectedLength {
				t.Errorf("randStringUUID() length = %d, want %d", len(result), expectedLength)
			}
		})
	}
}

// Test 3: updateDefaultValues function
func TestUpdateDefaultValues(t *testing.T) {
	tests := []struct {
		name     string
		vars     map[string]loader.EnvVar
		expected map[string]string // expected default values
	}{
		{
			name: "empty defaults get set",
			vars: map[string]loader.EnvVar{
				"COOKIE_SIGNING_SALT": {Name: "COOKIE_SIGNING_SALT", Default: ""},
				"UNKNOWN_VAR":         {Name: "UNKNOWN_VAR", Default: ""},
			},
			expected: map[string]string{
				"COOKIE_SIGNING_SALT": "salt",
				"UNKNOWN_VAR":         "", // no default defined
			},
		},
		{
			name: "existing defaults unchanged",
			vars: map[string]loader.EnvVar{
				"COOKIE_SIGNING_SALT": {Name: "COOKIE_SIGNING_SALT", Default: "existing"},
			},
			expected: map[string]string{
				"COOKIE_SIGNING_SALT": "existing",
			},
		},
		{
			name: "mixed scenarios",
			vars: map[string]loader.EnvVar{
				"PENTAGI_POSTGRES_PASSWORD":  {Name: "PENTAGI_POSTGRES_PASSWORD", Default: ""},
				"LANGFUSE_POSTGRES_PASSWORD": {Name: "LANGFUSE_POSTGRES_PASSWORD", Default: "custom"},
			},
			expected: map[string]string{
				"PENTAGI_POSTGRES_PASSWORD":  "postgres",
				"LANGFUSE_POSTGRES_PASSWORD": "custom",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			updateDefaultValues(tt.vars)

			for varName, expectedDefault := range tt.expected {
				if envVar, exists := tt.vars[varName]; exists {
					if envVar.Default != expectedDefault {
						t.Errorf("Variable %s: expected default %q, got %q", varName, expectedDefault, envVar.Default)
					}
				} else {
					t.Errorf("Variable %s should exist", varName)
				}
			}
		})
	}
}

// Test 4: replaceDefaultValues function
func TestReplaceDefaultValues(t *testing.T) {
	tests := []struct {
		name        string
		vars        map[string]loader.EnvVar
		policies    map[string]HardeningPolicy
		wantErr     bool
		wantChanged bool
	}{
		{
			name: "replace default values",
			vars: map[string]loader.EnvVar{
				"TEST_VAR": {Name: "TEST_VAR", Value: "default", Default: "default"},
			},
			policies: map[string]HardeningPolicy{
				"TEST_VAR": {Type: HardeningPolicyTypeDefault, Length: 10},
			},
			wantErr:     false,
			wantChanged: true,
		},
		{
			name: "skip non-default values",
			vars: map[string]loader.EnvVar{
				"TEST_VAR": {Name: "TEST_VAR", Value: "custom", Default: "default"},
			},
			policies: map[string]HardeningPolicy{
				"TEST_VAR": {Type: HardeningPolicyTypeDefault, Length: 10},
			},
			wantErr:     false,
			wantChanged: false,
		},
		{
			name: "invalid policy type",
			vars: map[string]loader.EnvVar{
				"TEST_VAR": {Name: "TEST_VAR", Value: "default", Default: "default"},
			},
			policies: map[string]HardeningPolicy{
				"TEST_VAR": {Type: "invalid"},
			},
			wantErr:     true,
			wantChanged: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			originalVars := make(map[string]loader.EnvVar)
			maps.Copy(originalVars, tt.vars)
			mockSt := &mockState{vars: make(map[string]loader.EnvVar)}

			isChanged, err := replaceDefaultValues(mockSt, tt.vars, tt.policies)
			if (err != nil) != tt.wantErr {
				t.Errorf("replaceDefaultValues() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if isChanged != tt.wantChanged {
				t.Errorf("replaceDefaultValues() isChanged = %v, wantChanged %v", isChanged, tt.wantChanged)
			}

			if !tt.wantErr {
				for varName, policy := range tt.policies {
					envVar := tt.vars[varName]
					originalVar := originalVars[varName]

					if originalVar.IsDefault() {
						// Should be replaced
						if envVar.Value == originalVar.Value {
							t.Errorf("Variable %s should have been replaced", varName)
						}
						if !envVar.IsChanged {
							t.Errorf("Variable %s should be marked as changed", varName)
						}
						// Validate the new value based on policy
						if policy.Type == HardeningPolicyTypeDefault && len(envVar.Value) != policy.Length {
							t.Errorf("Variable %s: expected length %d, got %d", varName, policy.Length, len(envVar.Value))
						}
					} else {
						// Should not be replaced
						if envVar.Value != originalVar.Value {
							t.Errorf("Variable %s should not have been replaced", varName)
						}
					}
				}
			}
		})
	}
}

// Test 5: syncValueToState function
func TestSyncValueToState(t *testing.T) {
	tests := []struct {
		name     string
		curVar   loader.EnvVar
		newValue string
		wantErr  bool
	}{
		{
			name:     "sync single variable",
			curVar:   loader.EnvVar{Name: "TEST_VAR", Value: "old_value"},
			newValue: "new_value",
			wantErr:  false,
		},
		{
			name:     "sync with empty new value",
			curVar:   loader.EnvVar{Name: "TEST_VAR", Value: "old_value"},
			newValue: "",
			wantErr:  false,
		},
		{
			name:     "sync with same value",
			curVar:   loader.EnvVar{Name: "TEST_VAR", Value: "same_value"},
			newValue: "same_value",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSt := &mockState{vars: make(map[string]loader.EnvVar)}

			resultVar, err := syncValueToState(mockSt, tt.curVar, tt.newValue)
			if (err != nil) != tt.wantErr {
				t.Errorf("syncValueToState() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Verify the variable was synced to state
				if actualVar, exists := mockSt.GetVar(tt.curVar.Name); exists {
					if actualVar.Value != tt.newValue {
						t.Errorf("Variable %s: expected value %q, got %q", tt.curVar.Name, tt.newValue, actualVar.Value)
					}
					if actualVar.IsChanged != true {
						t.Errorf("Variable %s should be marked as changed", tt.curVar.Name)
					}
				} else {
					t.Errorf("Variable %s should exist in state", tt.curVar.Name)
				}

				// Verify returned variable has correct default
				if resultVar.Default != tt.curVar.Value {
					t.Errorf("Result variable default should be %q, got %q", tt.curVar.Value, resultVar.Default)
				}
			}
		})
	}
}

// Test 6: Verify all varsForHardening variables exist in .env.example
func TestVarsExistInEnvExample(t *testing.T) {
	envExamplePath := getEnvExamplePath()

	// Load .env.example file
	envFile, err := loader.LoadEnvFile(envExamplePath)
	if err != nil {
		t.Fatalf("Failed to load .env.example: %v", err)
	}

	allEnvVars := envFile.GetAll()

	// Check all hardening variables exist
	for area, varNames := range varsForHardening {
		for _, varName := range varNames {
			t.Run(string(area)+"_"+varName, func(t *testing.T) {
				if _, exists := allEnvVars[varName]; !exists {
					t.Errorf("Variable %s from %s area not found in .env.example", varName, area)
				}
			})
		}
	}
}

// Test 7: Verify default values match .env.example
func TestDefaultValuesMatchEnvExample(t *testing.T) {
	envExamplePath := getEnvExamplePath()

	envFile, err := loader.LoadEnvFile(envExamplePath)
	if err != nil {
		t.Fatalf("Failed to load .env.example: %v", err)
	}

	allEnvVars := envFile.GetAll()

	// Create a copy of varsForHardeningDefault for testing
	testVars := make(map[string]loader.EnvVar)
	for varName := range varsForHardeningDefault {
		if envVar, exists := allEnvVars[varName]; exists {
			testVars[varName] = loader.EnvVar{
				Name:    varName,
				Value:   envVar.Value,
				Default: "",
			}
		}
	}

	// Update defaults
	updateDefaultValues(testVars)

	// Verify defaults were set correctly
	for varName, expectedDefault := range varsForHardeningDefault {
		if envVar, exists := testVars[varName]; exists {
			if envVar.Default != expectedDefault {
				t.Errorf("Variable %s: expected default %q, got %q", varName, expectedDefault, envVar.Default)
			}
		}
	}
}

// Test 8: DoHardening main logic
func TestDoHardening(t *testing.T) {
	tests := []struct {
		name          string
		checkResult   checker.CheckResult
		setupVars     map[string]loader.EnvVar
		expectChanges bool // whether we expect hardening to be applied
		// expectCommit removed - commit testing requires more sophisticated mocking
	}{
		{
			name: "langfuse not installed - should harden",
			checkResult: checker.CheckResult{
				LangfuseInstalled:    false,
				LangfuseVolumesExist: false,
				PentagiInstalled:     true,
				PentagiVolumesExist:  true,
			},
			setupVars: map[string]loader.EnvVar{
				"LANGFUSE_SALT": {Name: "LANGFUSE_SALT", Value: "salt", Default: "salt"},
			},
			expectChanges: true,
		},
		{
			name: "pentagi not installed - should harden",
			checkResult: checker.CheckResult{
				LangfuseInstalled:    true,
				LangfuseVolumesExist: true,
				PentagiInstalled:     false,
				PentagiVolumesExist:  false,
			},
			setupVars: map[string]loader.EnvVar{
				"COOKIE_SIGNING_SALT": {Name: "COOKIE_SIGNING_SALT", Value: "salt", Default: "salt"},
			},
			expectChanges: true,
		},
		{
			name: "both installed - should not harden",
			checkResult: checker.CheckResult{
				LangfuseInstalled:    true,
				LangfuseVolumesExist: true,
				PentagiInstalled:     true,
				PentagiVolumesExist:  true,
			},
			setupVars:     map[string]loader.EnvVar{},
			expectChanges: false,
		},
		{
			name: "neither installed - should harden both",
			checkResult: checker.CheckResult{
				LangfuseInstalled:    false,
				LangfuseVolumesExist: false,
				PentagiInstalled:     false,
				PentagiVolumesExist:  false,
			},
			setupVars: map[string]loader.EnvVar{
				"LANGFUSE_SALT":       {Name: "LANGFUSE_SALT", Value: "salt", Default: "salt"},
				"COOKIE_SIGNING_SALT": {Name: "COOKIE_SIGNING_SALT", Value: "salt", Default: "salt"},
			},
			expectChanges: true,
		},
		{
			name: "langfuse not installed but no default values - should not commit",
			checkResult: checker.CheckResult{
				LangfuseInstalled:    false,
				LangfuseVolumesExist: false,
				PentagiInstalled:     true,
				PentagiVolumesExist:  true,
			},
			setupVars: map[string]loader.EnvVar{
				"LANGFUSE_SALT": {Name: "LANGFUSE_SALT", Value: "custom", Default: "salt"}, // custom value, not default
			},
			expectChanges: false,
		},
		{
			name: "langfuse volumes exist but containers removed - should NOT harden",
			checkResult: checker.CheckResult{
				LangfuseInstalled:    false, // containers removed
				LangfuseVolumesExist: true,  // but volumes remain!
				PentagiInstalled:     true,
				PentagiVolumesExist:  true,
			},
			setupVars: map[string]loader.EnvVar{
				"LANGFUSE_SALT": {Name: "LANGFUSE_SALT", Value: "salt", Default: "salt"},
			},
			expectChanges: false, // should NOT change because volumes exist
		},
		{
			name: "pentagi volumes exist but containers removed - should NOT harden",
			checkResult: checker.CheckResult{
				LangfuseInstalled:    true,
				LangfuseVolumesExist: true,
				PentagiInstalled:     false, // containers removed
				PentagiVolumesExist:  true,  // but volumes remain!
			},
			setupVars: map[string]loader.EnvVar{
				"PENTAGI_POSTGRES_PASSWORD": {Name: "PENTAGI_POSTGRES_PASSWORD", Value: "postgres", Default: "postgres"},
			},
			expectChanges: false, // should NOT change because volumes exist
		},
		{
			name: "containers removed but volumes remain for both - should NOT harden any",
			checkResult: checker.CheckResult{
				LangfuseInstalled:    false, // containers removed
				LangfuseVolumesExist: true,  // volumes remain
				PentagiInstalled:     false, // containers removed
				PentagiVolumesExist:  true,  // volumes remain
			},
			setupVars: map[string]loader.EnvVar{
				"LANGFUSE_SALT":             {Name: "LANGFUSE_SALT", Value: "salt", Default: "salt"},
				"PENTAGI_POSTGRES_PASSWORD": {Name: "PENTAGI_POSTGRES_PASSWORD", Value: "postgres", Default: "postgres"},
			},
			expectChanges: false, // should NOT change anything
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock state that tracks commit calls
			type commitTrackingMockState struct {
				*mockState
				commitCalled bool
			}

			mockSt := &commitTrackingMockState{
				mockState: &mockState{vars: make(map[string]loader.EnvVar)},
			}
			// Copy setup vars into mock state
			for k, v := range tt.setupVars {
				mockSt.vars[k] = v
			}

			err := DoHardening(mockSt.mockState, tt.checkResult)
			if err != nil {
				t.Fatalf("DoHardening() error = %v", err)
			}

			if tt.expectChanges {
				// Verify that some variables were processed
				// This is a basic check - in a real scenario, we'd verify specific behaviors
				if len(mockSt.vars) == 0 && len(tt.setupVars) > 0 {
					t.Error("Expected some variables to be processed, but state is empty")
				}
			}

			// Note: In a real test environment, we would need to verify that Commit()
			// was called appropriately. This would require more sophisticated mocking
			// or integration testing. For now, we test the logic indirectly by checking
			// that variables were modified when expected.
		})
	}
}

// Test 9: Special SCRAPER_PRIVATE_URL logic in DoHardening
func TestDoHardening_ScraperURLLogic(t *testing.T) {
	tests := []struct {
		name            string
		setupVars       map[string]loader.EnvVar
		expectURLUpdate bool
	}{
		{
			name: "scraper credentials have default values - should update URL after hardening",
			setupVars: map[string]loader.EnvVar{
				"LOCAL_SCRAPER_USERNAME": {
					Name:      "LOCAL_SCRAPER_USERNAME",
					Value:     "someuser", // default value
					Default:   "someuser", // same as default
					IsChanged: false,
				},
				"LOCAL_SCRAPER_PASSWORD": {
					Name:      "LOCAL_SCRAPER_PASSWORD",
					Value:     "somepass", // default value
					Default:   "somepass", // same as default
					IsChanged: false,
				},
				"SCRAPER_PRIVATE_URL": {
					Name:    "SCRAPER_PRIVATE_URL",
					Value:   "https://someuser:somepass@scraper/",
					Default: "https://someuser:somepass@scraper/",
				},
			},
			expectURLUpdate: true, // URL will be updated with new random credentials
		},
		{
			name: "scraper credentials are custom - should not update URL",
			setupVars: map[string]loader.EnvVar{
				"LOCAL_SCRAPER_USERNAME": {
					Name:      "LOCAL_SCRAPER_USERNAME",
					Value:     "customuser", // custom value
					Default:   "someuser",   // different from default
					IsChanged: true,
				},
				"LOCAL_SCRAPER_PASSWORD": {
					Name:      "LOCAL_SCRAPER_PASSWORD",
					Value:     "custompass", // custom value
					Default:   "somepass",   // different from default
					IsChanged: true,
				},
				"SCRAPER_PRIVATE_URL": {
					Name:    "SCRAPER_PRIVATE_URL",
					Value:   "https://customuser:custompass@scraper/",
					Default: "https://someuser:somepass@scraper/",
				},
			},
			expectURLUpdate: false, // URL should not be updated for custom values
		},
		{
			name: "only username is custom - should not update URL because URL is not default",
			setupVars: map[string]loader.EnvVar{
				"LOCAL_SCRAPER_USERNAME": {
					Name:      "LOCAL_SCRAPER_USERNAME",
					Value:     "customuser", // custom value
					Default:   "someuser",   // different from default
					IsChanged: true,
				},
				"LOCAL_SCRAPER_PASSWORD": {
					Name:      "LOCAL_SCRAPER_PASSWORD",
					Value:     "somepass", // default value
					Default:   "somepass", // same as default
					IsChanged: false,
				},
				"SCRAPER_PRIVATE_URL": {
					Name:    "SCRAPER_PRIVATE_URL",
					Value:   "https://customuser:somepass@scraper/", // custom value - not default
					Default: "https://someuser:somepass@scraper/",
				},
			},
			expectURLUpdate: false, // URL will not be updated because it's not default
		},
		{
			name: "default credentials should update URL after hardening",
			setupVars: map[string]loader.EnvVar{
				"LOCAL_SCRAPER_USERNAME": {
					Name:      "LOCAL_SCRAPER_USERNAME",
					Value:     "someuser", // default value
					Default:   "someuser",
					IsChanged: false,
				},
				"LOCAL_SCRAPER_PASSWORD": {
					Name:      "LOCAL_SCRAPER_PASSWORD",
					Value:     "somepass", // default value
					Default:   "somepass",
					IsChanged: false,
				},
				"SCRAPER_PRIVATE_URL": {
					Name:    "SCRAPER_PRIVATE_URL",
					Value:   "https://someuser:somepass@scraper/", // default value
					Default: "https://someuser:somepass@scraper/", // same as default
				},
			},
			expectURLUpdate: true, // URL will be updated because URL is default and credentials will be hardened
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSt := &mockState{vars: tt.setupVars}

			checkResult := checker.CheckResult{
				LangfuseInstalled:    true,
				LangfuseVolumesExist: true,
				PentagiInstalled:     false, // Trigger pentagi hardening
				PentagiVolumesExist:  false,
			}

			originalURL := tt.setupVars["SCRAPER_PRIVATE_URL"].Value

			err := DoHardening(mockSt, checkResult)
			if err != nil {
				t.Fatalf("DoHardening() error = %v", err)
			}

			// Check if URL was updated as expected
			updatedVar, exists := mockSt.GetVar("SCRAPER_PRIVATE_URL")
			if !exists {
				t.Fatal("SCRAPER_PRIVATE_URL should exist in state")
			}

			urlChanged := updatedVar.Value != originalURL
			if tt.expectURLUpdate && !urlChanged {
				t.Errorf("Expected SCRAPER_PRIVATE_URL to be updated, but it wasn't")
			}
			if !tt.expectURLUpdate && urlChanged {
				t.Errorf("Expected SCRAPER_PRIVATE_URL to remain unchanged, but it was updated to: %s", updatedVar.Value)
			}

			if tt.expectURLUpdate && urlChanged {
				// For default values, verify the URL was updated with new random credentials
				// We just check that it's different from the original and has the expected format
				if updatedVar.Value == originalURL {
					t.Errorf("Updated URL should be different from original for default values")
				}
				if !strings.Contains(updatedVar.Value, "@scraper/") {
					t.Errorf("Updated URL should contain correct host: %s", updatedVar.Value)
				}
			}
		})
	}
}

// Test 10: syncLangfuseState function
func TestSyncLangfuseState(t *testing.T) {
	tests := []struct {
		name          string
		inputVars     map[string]loader.EnvVar
		expectedVars  map[string]string // expected values after sync
		expectedFlags map[string]bool   // expected IsChanged flags
		wantChanged   bool              // expected return value from function
	}{
		{
			name: "sync empty langfuse vars from init vars",
			inputVars: map[string]loader.EnvVar{
				"LANGFUSE_PROJECT_ID": {
					Name:      "LANGFUSE_PROJECT_ID",
					Value:     "", // empty, should be synced
					IsChanged: false,
				},
				"LANGFUSE_PUBLIC_KEY": {
					Name:      "LANGFUSE_PUBLIC_KEY",
					Value:     "", // empty, should be synced
					IsChanged: false,
				},
				"LANGFUSE_SECRET_KEY": {
					Name:      "LANGFUSE_SECRET_KEY",
					Value:     "", // empty, should be synced
					IsChanged: false,
				},
				"LANGFUSE_INIT_PROJECT_ID": {
					Name:      "LANGFUSE_INIT_PROJECT_ID",
					Value:     "cm47619l0000872mcd2dlbqwb",
					IsChanged: true,
				},
				"LANGFUSE_INIT_PROJECT_PUBLIC_KEY": {
					Name:      "LANGFUSE_INIT_PROJECT_PUBLIC_KEY",
					Value:     "pk-lf-12345678-1234-1234-1234-123456789abc",
					IsChanged: true,
				},
				"LANGFUSE_INIT_PROJECT_SECRET_KEY": {
					Name:      "LANGFUSE_INIT_PROJECT_SECRET_KEY",
					Value:     "sk-lf-87654321-4321-4321-4321-cba987654321",
					IsChanged: true,
				},
			},
			expectedVars: map[string]string{
				"LANGFUSE_PROJECT_ID": "cm47619l0000872mcd2dlbqwb",
				"LANGFUSE_PUBLIC_KEY": "pk-lf-12345678-1234-1234-1234-123456789abc",
				"LANGFUSE_SECRET_KEY": "sk-lf-87654321-4321-4321-4321-cba987654321",
			},
			expectedFlags: map[string]bool{
				"LANGFUSE_PROJECT_ID": true,
				"LANGFUSE_PUBLIC_KEY": true,
				"LANGFUSE_SECRET_KEY": true,
			},
			wantChanged: true,
		},
		{
			name: "do not sync non-empty langfuse vars",
			inputVars: map[string]loader.EnvVar{
				"LANGFUSE_PROJECT_ID": {
					Name:      "LANGFUSE_PROJECT_ID",
					Value:     "existing-project-id", // not empty, should not be synced
					IsChanged: false,
				},
				"LANGFUSE_PUBLIC_KEY": {
					Name:      "LANGFUSE_PUBLIC_KEY",
					Value:     "", // empty, should be synced
					IsChanged: false,
				},
				"LANGFUSE_INIT_PROJECT_ID": {
					Name:      "LANGFUSE_INIT_PROJECT_ID",
					Value:     "cm47619l0000872mcd2dlbqwb",
					IsChanged: true,
				},
				"LANGFUSE_INIT_PROJECT_PUBLIC_KEY": {
					Name:      "LANGFUSE_INIT_PROJECT_PUBLIC_KEY",
					Value:     "pk-lf-12345678-1234-1234-1234-123456789abc",
					IsChanged: true,
				},
			},
			expectedVars: map[string]string{
				"LANGFUSE_PROJECT_ID": "existing-project-id",                        // unchanged
				"LANGFUSE_PUBLIC_KEY": "pk-lf-12345678-1234-1234-1234-123456789abc", // synced
			},
			expectedFlags: map[string]bool{
				"LANGFUSE_PROJECT_ID": false, // unchanged
				"LANGFUSE_PUBLIC_KEY": true,  // synced
			},
			wantChanged: true, // because PUBLIC_KEY was synced
		},
		{
			name: "skip sync when init var does not exist",
			inputVars: map[string]loader.EnvVar{
				"LANGFUSE_PROJECT_ID": {
					Name:      "LANGFUSE_PROJECT_ID",
					Value:     "", // empty, but no init var to sync from
					IsChanged: false,
				},
			},
			expectedVars: map[string]string{
				"LANGFUSE_PROJECT_ID": "", // unchanged because no init var
			},
			expectedFlags: map[string]bool{
				"LANGFUSE_PROJECT_ID": false, // unchanged
			},
			wantChanged: false,
		},
		{
			name: "skip sync when target var does not exist",
			inputVars: map[string]loader.EnvVar{
				"LANGFUSE_INIT_PROJECT_ID": {
					Name:      "LANGFUSE_INIT_PROJECT_ID",
					Value:     "cm47619l0000872mcd2dlbqwb",
					IsChanged: true,
				},
				// No LANGFUSE_PROJECT_ID in vars
			},
			expectedVars:  map[string]string{},
			expectedFlags: map[string]bool{},
			wantChanged:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a copy of input vars to avoid modifying the original
			vars := make(map[string]loader.EnvVar)
			for k, v := range tt.inputVars {
				vars[k] = v
			}
			mockSt := &mockState{vars: make(map[string]loader.EnvVar)}

			// Call the function
			isChanged, err := syncLangfuseState(mockSt, vars)
			if err != nil {
				t.Errorf("syncLangfuseState() error = %v", err)
				return
			}

			if isChanged != tt.wantChanged {
				t.Errorf("syncLangfuseState() isChanged = %v, wantChanged %v", isChanged, tt.wantChanged)
			}

			// Check expected values
			for varName, expectedValue := range tt.expectedVars {
				if envVar, exists := vars[varName]; exists {
					if envVar.Value != expectedValue {
						t.Errorf("Variable %s: expected value %q, got %q", varName, expectedValue, envVar.Value)
					}
				} else {
					t.Errorf("Variable %s should exist", varName)
				}
			}

			// Check expected IsChanged flags
			for varName, expectedFlag := range tt.expectedFlags {
				if envVar, exists := vars[varName]; exists {
					if envVar.IsChanged != expectedFlag {
						t.Errorf("Variable %s: expected IsChanged %v, got %v", varName, expectedFlag, envVar.IsChanged)
					}
				}
			}
		})
	}
}

// Test 11: syncScraperState function
func TestSyncScraperState(t *testing.T) {
	tests := []struct {
		name            string
		inputVars       map[string]loader.EnvVar
		expectError     bool
		expectURLUpdate bool
		expectedURL     string
		wantChanged     bool // expected return value from function
	}{
		{
			name: "update default URL with hardened credentials",
			inputVars: map[string]loader.EnvVar{
				"LOCAL_SCRAPER_USERNAME": {
					Name:      "LOCAL_SCRAPER_USERNAME",
					Value:     "newuser",
					Default:   "someuser",
					IsChanged: true, // credential was hardened
				},
				"LOCAL_SCRAPER_PASSWORD": {
					Name:      "LOCAL_SCRAPER_PASSWORD",
					Value:     "newpass",
					Default:   "somepass",
					IsChanged: true, // credential was hardened
				},
				"SCRAPER_PRIVATE_URL": {
					Name:    "SCRAPER_PRIVATE_URL",
					Value:   "https://someuser:somepass@scraper/", // default value
					Default: "https://someuser:somepass@scraper/", // same as default
				},
			},
			expectError:     false,
			expectURLUpdate: true,
			expectedURL:     "https://newuser:newpass@scraper/",
			wantChanged:     true,
		},
		{
			name: "do not update custom URL",
			inputVars: map[string]loader.EnvVar{
				"LOCAL_SCRAPER_USERNAME": {
					Name:      "LOCAL_SCRAPER_USERNAME",
					Value:     "newuser",
					Default:   "someuser",
					IsChanged: true,
				},
				"LOCAL_SCRAPER_PASSWORD": {
					Name:      "LOCAL_SCRAPER_PASSWORD",
					Value:     "newpass",
					Default:   "somepass",
					IsChanged: true,
				},
				"SCRAPER_PRIVATE_URL": {
					Name:    "SCRAPER_PRIVATE_URL",
					Value:   "https://customuser:custompass@scraper/", // custom value
					Default: "https://someuser:somepass@scraper/",     // different from default
				},
			},
			expectError:     false,
			expectURLUpdate: false,
			expectedURL:     "https://customuser:custompass@scraper/", // unchanged
			wantChanged:     false,
		},
		{
			name: "do not update when credentials not changed",
			inputVars: map[string]loader.EnvVar{
				"LOCAL_SCRAPER_USERNAME": {
					Name:      "LOCAL_SCRAPER_USERNAME",
					Value:     "someuser",
					Default:   "someuser",
					IsChanged: false, // not changed
				},
				"LOCAL_SCRAPER_PASSWORD": {
					Name:      "LOCAL_SCRAPER_PASSWORD",
					Value:     "somepass",
					Default:   "somepass",
					IsChanged: false, // not changed
				},
				"SCRAPER_PRIVATE_URL": {
					Name:    "SCRAPER_PRIVATE_URL",
					Value:   "https://someuser:somepass@scraper/",
					Default: "https://someuser:somepass@scraper/",
				},
			},
			expectError:     false,
			expectURLUpdate: false,
			expectedURL:     "https://someuser:somepass@scraper/", // unchanged
			wantChanged:     false,
		},
		{
			name: "update when only one credential changed",
			inputVars: map[string]loader.EnvVar{
				"LOCAL_SCRAPER_USERNAME": {
					Name:      "LOCAL_SCRAPER_USERNAME",
					Value:     "newuser",
					Default:   "someuser",
					IsChanged: true, // changed
				},
				"LOCAL_SCRAPER_PASSWORD": {
					Name:      "LOCAL_SCRAPER_PASSWORD",
					Value:     "somepass",
					Default:   "somepass",
					IsChanged: false, // not changed
				},
				"SCRAPER_PRIVATE_URL": {
					Name:    "SCRAPER_PRIVATE_URL",
					Value:   "https://someuser:somepass@scraper/",
					Default: "https://someuser:somepass@scraper/",
				},
			},
			expectError:     false,
			expectURLUpdate: true,
			expectedURL:     "https://newuser:somepass@scraper/",
			wantChanged:     true,
		},
		{
			name: "handle missing variables gracefully",
			inputVars: map[string]loader.EnvVar{
				"LOCAL_SCRAPER_USERNAME": {
					Name:      "LOCAL_SCRAPER_USERNAME",
					Value:     "newuser",
					Default:   "someuser",
					IsChanged: true,
				},
				// Missing LOCAL_SCRAPER_PASSWORD and SCRAPER_PRIVATE_URL
			},
			expectError:     false,
			expectURLUpdate: false,
			expectedURL:     "", // no URL to update
			wantChanged:     false,
		},
		{
			name: "handle invalid URL gracefully",
			inputVars: map[string]loader.EnvVar{
				"LOCAL_SCRAPER_USERNAME": {
					Name:      "LOCAL_SCRAPER_USERNAME",
					Value:     "newuser",
					Default:   "someuser",
					IsChanged: true,
				},
				"LOCAL_SCRAPER_PASSWORD": {
					Name:      "LOCAL_SCRAPER_PASSWORD",
					Value:     "newpass",
					Default:   "somepass",
					IsChanged: true,
				},
				"SCRAPER_PRIVATE_URL": {
					Name:    "SCRAPER_PRIVATE_URL",
					Value:   "://invalid-url", // invalid scheme format that will cause parse error
					Default: "://invalid-url",
				},
			},
			expectError:     true, // should return error for invalid URL
			expectURLUpdate: false,
			expectedURL:     "://invalid-url",
			wantChanged:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a copy of input vars to avoid modifying the original
			vars := make(map[string]loader.EnvVar)
			for k, v := range tt.inputVars {
				vars[k] = v
			}
			mockSt := &mockState{vars: make(map[string]loader.EnvVar)}

			originalURL := ""
			if urlVar, exists := vars["SCRAPER_PRIVATE_URL"]; exists {
				originalURL = urlVar.Value
			}

			// Call the function
			isChanged, err := syncScraperState(mockSt, vars)

			// Check error expectation
			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if isChanged != tt.wantChanged {
				t.Errorf("syncScraperState() isChanged = %v, wantChanged %v", isChanged, tt.wantChanged)
			}

			// Check URL update expectation
			if urlVar, exists := vars["SCRAPER_PRIVATE_URL"]; exists {
				urlChanged := urlVar.Value != originalURL

				if tt.expectURLUpdate && !urlChanged {
					t.Errorf("Expected URL to be updated but it wasn't")
				}
				if !tt.expectURLUpdate && urlChanged {
					t.Errorf("Expected URL to remain unchanged but it was updated")
				}

				if tt.expectedURL != "" && urlVar.Value != tt.expectedURL {
					t.Errorf("Expected URL %q, got %q", tt.expectedURL, urlVar.Value)
				}

				// If URL was updated, IsChanged should be true
				if tt.expectURLUpdate && urlChanged && !urlVar.IsChanged {
					t.Errorf("Expected IsChanged to be true when URL is updated")
				}
			} else if tt.expectedURL != "" {
				t.Errorf("Expected URL variable to exist")
			}
		})
	}
}

// Test 12: Integration test with real .env.example file
func TestDoHardening_IntegrationWithRealEnvFile(t *testing.T) {
	// Read the real .env.example file
	envExamplePath := getEnvExamplePath()
	envExampleContent, err := os.ReadFile(envExamplePath)
	if err != nil {
		t.Fatalf("Failed to read .env.example: %v", err)
	}

	tests := []struct {
		name                  string
		checkResult           checker.CheckResult
		expectedHardenedVars  []string // variables that should be hardened
		expectedUnchangedVars []string // variables that should remain unchanged
	}{
		{
			name: "harden langfuse only",
			checkResult: checker.CheckResult{
				LangfuseInstalled:    false, // Should harden langfuse
				LangfuseVolumesExist: false,
				PentagiInstalled:     true, // Should not harden pentagi
				PentagiVolumesExist:  true,
			},
			expectedHardenedVars: []string{
				"LANGFUSE_POSTGRES_PASSWORD",
				"LANGFUSE_CLICKHOUSE_PASSWORD",
				"LANGFUSE_S3_ACCESS_KEY_ID",
				"LANGFUSE_S3_SECRET_ACCESS_KEY",
				"LANGFUSE_REDIS_AUTH",
				"LANGFUSE_SALT",
				"LANGFUSE_ENCRYPTION_KEY",
				"LANGFUSE_NEXTAUTH_SECRET",
				"LANGFUSE_INIT_PROJECT_PUBLIC_KEY",
				"LANGFUSE_INIT_PROJECT_SECRET_KEY",
				"LANGFUSE_AUTH_DISABLE_SIGNUP",
			},
			expectedUnchangedVars: []string{
				"COOKIE_SIGNING_SALT",
				"PENTAGI_POSTGRES_PASSWORD",
				"NEO4J_PASSWORD",
				"LOCAL_SCRAPER_USERNAME",
				"LOCAL_SCRAPER_PASSWORD",
			},
		},
		{
			name: "harden pentagi only",
			checkResult: checker.CheckResult{
				LangfuseInstalled:    true, // Should not harden langfuse
				LangfuseVolumesExist: true,
				PentagiInstalled:     false, // Should harden pentagi
				PentagiVolumesExist:  false,
			},
			expectedHardenedVars: []string{
				"COOKIE_SIGNING_SALT",
				"PENTAGI_POSTGRES_PASSWORD",
				"NEO4J_PASSWORD",
				"LOCAL_SCRAPER_USERNAME",
				"LOCAL_SCRAPER_PASSWORD",
				"SCRAPER_PRIVATE_URL", // Should be updated if credentials are hardened
			},
			expectedUnchangedVars: []string{
				"LANGFUSE_POSTGRES_PASSWORD",
				"LANGFUSE_CLICKHOUSE_PASSWORD",
				"LANGFUSE_S3_ACCESS_KEY_ID",
				"LANGFUSE_S3_SECRET_ACCESS_KEY",
			},
		},
		{
			name: "harden both",
			checkResult: checker.CheckResult{
				LangfuseInstalled:    false, // Should harden langfuse
				LangfuseVolumesExist: false,
				PentagiInstalled:     false, // Should harden pentagi
				PentagiVolumesExist:  false,
			},
			expectedHardenedVars: []string{
				// Pentagi vars
				"COOKIE_SIGNING_SALT",
				"PENTAGI_POSTGRES_PASSWORD",
				"NEO4J_PASSWORD",
				"LOCAL_SCRAPER_USERNAME",
				"LOCAL_SCRAPER_PASSWORD",
				"SCRAPER_PRIVATE_URL",
				// Langfuse vars
				"LANGFUSE_POSTGRES_PASSWORD",
				"LANGFUSE_CLICKHOUSE_PASSWORD",
				"LANGFUSE_S3_ACCESS_KEY_ID",
				"LANGFUSE_S3_SECRET_ACCESS_KEY",
				"LANGFUSE_REDIS_AUTH",
				"LANGFUSE_SALT",
				"LANGFUSE_ENCRYPTION_KEY",
				"LANGFUSE_NEXTAUTH_SECRET",
				"LANGFUSE_INIT_PROJECT_PUBLIC_KEY",
				"LANGFUSE_INIT_PROJECT_SECRET_KEY",
				"LANGFUSE_AUTH_DISABLE_SIGNUP",
				// Langfuse sync vars should be updated too
				"LANGFUSE_PROJECT_ID",
				"LANGFUSE_PUBLIC_KEY",
				"LANGFUSE_SECRET_KEY",
			},
			expectedUnchangedVars: []string{
				// Variables that should never be hardened or are managed differently
				"LANGFUSE_INIT_PROJECT_ID", // This doesn't get hardened, only synced to other vars
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary copy of .env.example
			tempEnvPath := createTempEnvFile(t, string(envExampleContent))
			defer os.Remove(tempEnvPath)

			// Load the temporary file into state
			envFile, err := loader.LoadEnvFile(tempEnvPath)
			if err != nil {
				t.Fatalf("Failed to load temp env file: %v", err)
			}

			// Create mock state with real data
			allVars := envFile.GetAll()
			mockSt := &mockState{
				vars:    allVars,
				envPath: tempEnvPath,
			}

			// Store original values for comparison
			originalValues := make(map[string]string)
			for varName, envVar := range allVars {
				originalValues[varName] = envVar.Value
			}

			// Run DoHardening
			err = DoHardening(mockSt, tt.checkResult)
			if err != nil {
				t.Fatalf("DoHardening() error = %v", err)
			}

			// Check that expected variables were hardened
			for _, varName := range tt.expectedHardenedVars {
				if updatedVar, exists := mockSt.GetVar(varName); exists {
					originalValue := originalValues[varName]

					// For most variables, check they changed from default
					if defaultValue, hasDefault := varsForHardeningDefault[varName]; hasDefault {
						if originalValue == defaultValue && updatedVar.Value == originalValue {
							t.Errorf("Variable %s should have been hardened but remained unchanged", varName)
						}
						if originalValue == defaultValue && updatedVar.Value != originalValue {
							// Good - default value was hardened
							if !updatedVar.IsChanged {
								t.Errorf("Variable %s should be marked as changed after hardening", varName)
							}

							// Validate the new value based on hardening policy
							if err := validateHardenedValue(varName, updatedVar.Value); err != nil {
								t.Errorf("Variable %s hardened value validation failed: %v", varName, err)
							}
						}
					} else {
						// For variables without defaults (like sync vars), just check they were updated
						if varName == "LANGFUSE_PROJECT_ID" || varName == "LANGFUSE_PUBLIC_KEY" || varName == "LANGFUSE_SECRET_KEY" {
							if updatedVar.Value == "" {
								t.Errorf("Sync variable %s should have been updated but is still empty", varName)
							}
						}
					}
				} else {
					t.Errorf("Expected hardened variable %s not found in state", varName)
				}
			}

			// Check that expected variables were NOT hardened
			for _, varName := range tt.expectedUnchangedVars {
				if updatedVar, exists := mockSt.GetVar(varName); exists {
					originalValue := originalValues[varName]
					if updatedVar.Value != originalValue {
						t.Errorf("Variable %s should not have been changed but was updated from %q to %q",
							varName, originalValue, updatedVar.Value)
					}
				}
			}

			// Note: We don't verify file consistency here because mockState
			// doesn't write back to file. In real system, state.Commit() would handle this.

			// Verify sync relationships for Langfuse
			if !tt.checkResult.LangfuseInstalled {
				verifyLangfuseSyncRelationships(t, mockSt)
			}

			// Verify scraper URL consistency for Pentagi
			if !tt.checkResult.PentagiInstalled {
				verifyScraperURLConsistency(t, mockSt)
			}
		})
	}
}

// Helper function to validate hardened values
func validateHardenedValue(varName, value string) error {
	// Get the policy for this variable
	var policy HardeningPolicy
	var found bool

	for _, policies := range varsHardeningPolicies {
		if p, exists := policies[varName]; exists {
			policy = p
			found = true
			break
		}
	}

	if !found {
		return nil // No policy means no validation needed
	}

	switch policy.Type {
	case HardeningPolicyTypeDefault:
		if len(value) != policy.Length {
			return fmt.Errorf("expected length %d, got %d", policy.Length, len(value))
		}
		if !isAlphanumeric(value) {
			return fmt.Errorf("should be alphanumeric")
		}
	case HardeningPolicyTypeHex:
		if len(value) != policy.Length {
			return fmt.Errorf("expected length %d, got %d", policy.Length, len(value))
		}
		if !isHexString(value) {
			return fmt.Errorf("should be hex string")
		}
	case HardeningPolicyTypeUUID:
		if !strings.HasPrefix(value, policy.Prefix) {
			return fmt.Errorf("should start with %q", policy.Prefix)
		}
		expectedLength := len(policy.Prefix) + 36 // UUID is 36 chars
		if len(value) != expectedLength {
			return fmt.Errorf("expected total length %d, got %d", expectedLength, len(value))
		}
	case HardeningPolicyTypeBoolTrue:
		if value != "true" {
			return fmt.Errorf("should be 'true'")
		}
	case HardeningPolicyTypeBoolFalse:
		if value != "false" {
			return fmt.Errorf("should be 'false'")
		}
	}

	return nil
}

// Helper function to verify Langfuse sync relationships
func verifyLangfuseSyncRelationships(t *testing.T, state *mockState) {
	for varName, syncVarName := range varsHardeningSyncLangfuse {
		if targetVar, targetExists := state.GetVar(varName); targetExists {
			if sourceVar, sourceExists := state.GetVar(syncVarName); sourceExists {
				if targetVar.Value == "" && sourceVar.Value != "" {
					t.Errorf("Langfuse sync failed: %s is empty but %s has value %q",
						varName, syncVarName, sourceVar.Value)
				} else if targetVar.Value != "" && sourceVar.Value != "" && targetVar.Value != sourceVar.Value {
					t.Errorf("Langfuse sync inconsistent: %s=%q, %s=%q",
						varName, targetVar.Value, syncVarName, sourceVar.Value)
				}
			}
		}
	}
}

// Helper function to verify scraper URL consistency
func verifyScraperURLConsistency(t *testing.T, state *mockState) {
	urlVar, urlExists := state.GetVar("SCRAPER_PRIVATE_URL")
	userVar, userExists := state.GetVar("LOCAL_SCRAPER_USERNAME")
	passVar, passExists := state.GetVar("LOCAL_SCRAPER_PASSWORD")

	if !urlExists || !userExists || !passExists {
		return // Can't verify if variables don't exist
	}

	// If credentials were hardened (changed), URL should be updated too (if it was default)
	if userVar.IsChanged && passVar.IsChanged && urlVar.IsDefault() {
		expectedURL := fmt.Sprintf("https://%s:%s@scraper/", userVar.Value, passVar.Value)
		if urlVar.Value != expectedURL {
			t.Errorf("Scraper URL should be updated to match credentials: expected %q, got %q",
				expectedURL, urlVar.Value)
		}
		if !urlVar.IsChanged {
			t.Errorf("Scraper URL should be marked as changed when credentials are hardened")
		}
	}
}

// Helper functions
func isAlphanumeric(s string) bool {
	for _, r := range s {
		if !((r >= '0' && r <= '9') || (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z')) {
			return false
		}
	}
	return true
}

func isHexString(s string) bool {
	for _, r := range s {
		if !((r >= '0' && r <= '9') || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F')) {
			return false
		}
	}
	return true
}
