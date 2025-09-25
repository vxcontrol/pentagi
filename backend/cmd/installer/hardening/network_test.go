package hardening

import (
	"errors"
	"os"
	"testing"

	"pentagi/cmd/installer/loader"
)

var mockError = errors.New("mocked error")

// mockStateWithErrors is an extension of mockState that can simulate errors
type mockStateWithErrors struct {
	vars         map[string]loader.EnvVar
	envPath      string
	setVarError  map[string]error
	setVarsError error
}

func (m *mockStateWithErrors) GetVar(key string) (loader.EnvVar, bool) {
	if val, exists := m.vars[key]; exists {
		return val, true
	}
	return loader.EnvVar{Name: key, Line: -1}, false
}

func (m *mockStateWithErrors) GetVars(names []string) (map[string]loader.EnvVar, map[string]bool) {
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

func (m *mockStateWithErrors) SetVar(name, value string) error {
	if m.setVarError != nil {
		if err, hasError := m.setVarError[name]; hasError {
			return err
		}
	}

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

func (m *mockStateWithErrors) SetVars(vars map[string]string) error {
	if m.setVarsError != nil {
		return m.setVarsError
	}

	for name, value := range vars {
		if err := m.SetVar(name, value); err != nil {
			return err
		}
	}
	return nil
}

func (m *mockStateWithErrors) GetEnvPath() string                   { return m.envPath }
func (m *mockStateWithErrors) Exists() bool                         { return true }
func (m *mockStateWithErrors) Reset() error                         { return nil }
func (m *mockStateWithErrors) Commit() error                        { return nil }
func (m *mockStateWithErrors) IsDirty() bool                        { return false }
func (m *mockStateWithErrors) GetEulaConsent() bool                 { return true }
func (m *mockStateWithErrors) SetEulaConsent() error                { return nil }
func (m *mockStateWithErrors) SetStack(stack []string) error        { return nil }
func (m *mockStateWithErrors) GetStack() []string                   { return []string{} }
func (m *mockStateWithErrors) ResetVar(name string) error           { return nil }
func (m *mockStateWithErrors) ResetVars(names []string) error       { return nil }
func (m *mockStateWithErrors) GetAllVars() map[string]loader.EnvVar { return m.vars }

// Test 1: HTTP_PROXY and HTTPS_PROXY synchronization
func TestDoSyncNetworkSettings_ProxySettings(t *testing.T) {
	tests := []struct {
		name        string
		httpProxy   string
		httpsProxy  string
		setHTTP     bool
		setHTTPS    bool
		expectedVar string
		expectedVal string
		wantErr     bool
	}{
		{
			name:        "set HTTP_PROXY only",
			httpProxy:   "http://proxy.example.com:8080",
			setHTTP:     true,
			setHTTPS:    false,
			expectedVar: "PROXY_URL",
			expectedVal: "http://proxy.example.com:8080",
			wantErr:     false,
		},
		{
			name:        "set HTTPS_PROXY only",
			httpsProxy:  "https://proxy.example.com:8443",
			setHTTP:     false,
			setHTTPS:    true,
			expectedVar: "PROXY_URL",
			expectedVal: "https://proxy.example.com:8443",
			wantErr:     false,
		},
		{
			name:        "HTTPS_PROXY overrides HTTP_PROXY",
			httpProxy:   "http://proxy.example.com:8080",
			httpsProxy:  "https://proxy.example.com:8443",
			setHTTP:     true,
			setHTTPS:    true,
			expectedVar: "PROXY_URL",
			expectedVal: "https://proxy.example.com:8443", // HTTPS takes precedence
			wantErr:     false,
		},
		{
			name:        "empty HTTP_PROXY should not set PROXY_URL",
			httpProxy:   "",
			setHTTP:     true,
			setHTTPS:    false,
			expectedVar: "",
			expectedVal: "",
			wantErr:     false,
		},
		{
			name:        "empty HTTPS_PROXY should not set PROXY_URL",
			httpsProxy:  "",
			setHTTP:     false,
			setHTTPS:    true,
			expectedVar: "",
			expectedVal: "",
			wantErr:     false,
		},
		{
			name:        "no proxy variables set",
			setHTTP:     false,
			setHTTPS:    false,
			expectedVar: "",
			expectedVal: "",
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original environment
			originalHTTP := os.Getenv("HTTP_PROXY")
			originalHTTPS := os.Getenv("HTTPS_PROXY")

			// Clean up after test
			defer func() {
				if originalHTTP != "" {
					os.Setenv("HTTP_PROXY", originalHTTP)
				} else {
					os.Unsetenv("HTTP_PROXY")
				}
				if originalHTTPS != "" {
					os.Setenv("HTTPS_PROXY", originalHTTPS)
				} else {
					os.Unsetenv("HTTPS_PROXY")
				}
			}()

			// Set up test environment
			os.Unsetenv("HTTP_PROXY")
			os.Unsetenv("HTTPS_PROXY")

			if tt.setHTTP {
				os.Setenv("HTTP_PROXY", tt.httpProxy)
			}
			if tt.setHTTPS {
				os.Setenv("HTTPS_PROXY", tt.httpsProxy)
			}

			// Create mock state
			mockSt := &mockState{vars: make(map[string]loader.EnvVar)}

			// Execute function
			err := DoSyncNetworkSettings(mockSt)

			// Check error expectation
			if (err != nil) != tt.wantErr {
				t.Errorf("DoSyncNetworkSettings() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Check expected variable was set
			if tt.expectedVar != "" {
				if actualVar, exists := mockSt.GetVar(tt.expectedVar); exists {
					if actualVar.Value != tt.expectedVal {
						t.Errorf("Expected %s = %q, got %q", tt.expectedVar, tt.expectedVal, actualVar.Value)
					}
					if !actualVar.IsChanged {
						t.Errorf("Variable %s should be marked as changed", tt.expectedVar)
					}
				} else {
					t.Errorf("Expected variable %s to be set in state", tt.expectedVar)
				}
			} else {
				// No variable should be set
				if actualVar, exists := mockSt.GetVar("PROXY_URL"); exists && actualVar.Value != "" {
					t.Errorf("No proxy variable should be set, but PROXY_URL = %q", actualVar.Value)
				}
			}
		})
	}
}

// Test 2: Docker environment variables synchronization
func TestDoSyncNetworkSettings_DockerEnvVars(t *testing.T) {
	tests := []struct {
		name         string
		dockerVars   map[string]string // variable name -> value
		setVars      map[string]bool   // variable name -> should be set
		expectSync   bool              // should Docker vars be synced
		expectedVars map[string]string // expected state variables
		wantErr      bool
	}{
		{
			name: "set all Docker variables",
			dockerVars: map[string]string{
				"DOCKER_HOST":       "tcp://docker.example.com:2376",
				"DOCKER_TLS_VERIFY": "1",
				"DOCKER_CERT_PATH":  "/path/to/certs",
			},
			setVars: map[string]bool{
				"DOCKER_HOST":       true,
				"DOCKER_TLS_VERIFY": true,
				"DOCKER_CERT_PATH":  true,
			},
			expectSync: true,
			expectedVars: map[string]string{
				"DOCKER_HOST":       "tcp://docker.example.com:2376",
				"DOCKER_TLS_VERIFY": "1",
				"DOCKER_CERT_PATH":  "/path/to/certs",
			},
			wantErr: false,
		},
		{
			name: "set only DOCKER_HOST",
			dockerVars: map[string]string{
				"DOCKER_HOST": "tcp://docker.example.com:2376",
			},
			setVars: map[string]bool{
				"DOCKER_HOST":       true,
				"DOCKER_TLS_VERIFY": false,
				"DOCKER_CERT_PATH":  false,
			},
			expectSync: true,
			expectedVars: map[string]string{
				"DOCKER_HOST":       "tcp://docker.example.com:2376",
				"DOCKER_TLS_VERIFY": "", // empty value should be synced too
				"DOCKER_CERT_PATH":  "", // empty value should be synced too
			},
			wantErr: false,
		},
		{
			name: "set only DOCKER_TLS_VERIFY",
			dockerVars: map[string]string{
				"DOCKER_TLS_VERIFY": "1",
			},
			setVars: map[string]bool{
				"DOCKER_HOST":       false,
				"DOCKER_TLS_VERIFY": true,
				"DOCKER_CERT_PATH":  false,
			},
			expectSync: true,
			expectedVars: map[string]string{
				"DOCKER_HOST":       "",
				"DOCKER_TLS_VERIFY": "1",
				"DOCKER_CERT_PATH":  "",
			},
			wantErr: false,
		},
		{
			name: "set only DOCKER_CERT_PATH",
			dockerVars: map[string]string{
				"DOCKER_CERT_PATH": "/path/to/certs",
			},
			setVars: map[string]bool{
				"DOCKER_HOST":       false,
				"DOCKER_TLS_VERIFY": false,
				"DOCKER_CERT_PATH":  true,
			},
			expectSync: true,
			expectedVars: map[string]string{
				"DOCKER_HOST":       "",
				"DOCKER_TLS_VERIFY": "",
				"DOCKER_CERT_PATH":  "/path/to/certs",
			},
			wantErr: false,
		},
		{
			name: "empty Docker variables should not trigger sync",
			dockerVars: map[string]string{
				"DOCKER_HOST":       "",
				"DOCKER_TLS_VERIFY": "",
				"DOCKER_CERT_PATH":  "",
			},
			setVars: map[string]bool{
				"DOCKER_HOST":       true,
				"DOCKER_TLS_VERIFY": true,
				"DOCKER_CERT_PATH":  true,
			},
			expectSync:   false,
			expectedVars: map[string]string{},
			wantErr:      false,
		},
		{
			name:         "no Docker variables set should not trigger sync",
			dockerVars:   map[string]string{},
			setVars:      map[string]bool{},
			expectSync:   false,
			expectedVars: map[string]string{},
			wantErr:      false,
		},
		{
			name: "mixed empty and non-empty Docker variables",
			dockerVars: map[string]string{
				"DOCKER_HOST":       "tcp://docker.example.com:2376",
				"DOCKER_TLS_VERIFY": "", // empty
				"DOCKER_CERT_PATH":  "/path/to/certs",
			},
			setVars: map[string]bool{
				"DOCKER_HOST":       true,
				"DOCKER_TLS_VERIFY": true, // set but empty
				"DOCKER_CERT_PATH":  true,
			},
			expectSync: true, // should sync because DOCKER_HOST and DOCKER_CERT_PATH are non-empty
			expectedVars: map[string]string{
				"DOCKER_HOST":       "tcp://docker.example.com:2376",
				"DOCKER_TLS_VERIFY": "", // empty value gets synced too
				"DOCKER_CERT_PATH":  "/path/to/certs",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original environment
			dockerEnvVarNames := []string{"DOCKER_HOST", "DOCKER_TLS_VERIFY", "DOCKER_CERT_PATH"}
			originalEnv := make(map[string]string)
			for _, varName := range dockerEnvVarNames {
				originalEnv[varName] = os.Getenv(varName)
			}

			// Clean up after test
			defer func() {
				for _, varName := range dockerEnvVarNames {
					if originalVal := originalEnv[varName]; originalVal != "" {
						os.Setenv(varName, originalVal)
					} else {
						os.Unsetenv(varName)
					}
				}
			}()

			// Clear environment first
			for _, varName := range dockerEnvVarNames {
				os.Unsetenv(varName)
			}

			// Set up test environment
			for varName, shouldSet := range tt.setVars {
				if shouldSet {
					value := tt.dockerVars[varName]
					os.Setenv(varName, value)
				}
			}

			// Create mock state
			mockSt := &mockState{vars: make(map[string]loader.EnvVar)}

			// Execute function
			err := DoSyncNetworkSettings(mockSt)

			// Check error expectation
			if (err != nil) != tt.wantErr {
				t.Errorf("DoSyncNetworkSettings() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Check if variables were synced as expected
			if tt.expectSync {
				for varName, expectedVal := range tt.expectedVars {
					if actualVar, exists := mockSt.GetVar(varName); exists {
						if actualVar.Value != expectedVal {
							t.Errorf("Expected %s = %q, got %q", varName, expectedVal, actualVar.Value)
						}
						if !actualVar.IsChanged {
							t.Errorf("Variable %s should be marked as changed", varName)
						}
					} else {
						t.Errorf("Expected variable %s to be set in state", varName)
					}
				}
			} else {
				// No Docker variables should be synced
				for _, varName := range dockerEnvVarNames {
					if actualVar, exists := mockSt.GetVar(varName); exists && actualVar.Value != "" {
						t.Errorf("Docker variable %s should not be synced when expectSync=false, but got %q", varName, actualVar.Value)
					}
				}
			}
		})
	}
}

// Test 3: Combined proxy and Docker variables test
func TestDoSyncNetworkSettings_CombinedScenarios(t *testing.T) {
	tests := []struct {
		name            string
		httpProxy       string
		httpsProxy      string
		dockerVars      map[string]string
		setProxyVars    map[string]bool
		setDockerVars   map[string]bool
		expectedResults map[string]string
		wantErr         bool
	}{
		{
			name:       "both proxy and Docker variables set",
			httpProxy:  "http://proxy.example.com:8080",
			httpsProxy: "https://proxy.example.com:8443",
			dockerVars: map[string]string{
				"DOCKER_HOST":       "tcp://docker.example.com:2376",
				"DOCKER_TLS_VERIFY": "1",
				"DOCKER_CERT_PATH":  "/path/to/certs",
			},
			setProxyVars: map[string]bool{
				"HTTP_PROXY":  true,
				"HTTPS_PROXY": true,
			},
			setDockerVars: map[string]bool{
				"DOCKER_HOST":       true,
				"DOCKER_TLS_VERIFY": true,
				"DOCKER_CERT_PATH":  true,
			},
			expectedResults: map[string]string{
				"PROXY_URL":         "https://proxy.example.com:8443", // HTTPS takes precedence
				"DOCKER_HOST":       "tcp://docker.example.com:2376",
				"DOCKER_TLS_VERIFY": "1",
				"DOCKER_CERT_PATH":  "/path/to/certs",
			},
			wantErr: false,
		},
		{
			name:      "only proxy variables, no Docker",
			httpProxy: "http://proxy.example.com:8080",
			setProxyVars: map[string]bool{
				"HTTP_PROXY": true,
			},
			setDockerVars: map[string]bool{},
			expectedResults: map[string]string{
				"PROXY_URL": "http://proxy.example.com:8080",
				// No Docker variables should be set
			},
			wantErr: false,
		},
		{
			name: "only Docker variables, no proxy",
			dockerVars: map[string]string{
				"DOCKER_HOST": "tcp://docker.example.com:2376",
			},
			setProxyVars: map[string]bool{},
			setDockerVars: map[string]bool{
				"DOCKER_HOST": true,
			},
			expectedResults: map[string]string{
				"DOCKER_HOST":       "tcp://docker.example.com:2376",
				"DOCKER_TLS_VERIFY": "", // empty values get synced too
				"DOCKER_CERT_PATH":  "", // empty values get synced too
				// No PROXY_URL should be set
			},
			wantErr: false,
		},
		{
			name:            "no environment variables set",
			setProxyVars:    map[string]bool{},
			setDockerVars:   map[string]bool{},
			expectedResults: map[string]string{},
			wantErr:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original environment
			allEnvVars := []string{"HTTP_PROXY", "HTTPS_PROXY", "DOCKER_HOST", "DOCKER_TLS_VERIFY", "DOCKER_CERT_PATH"}
			originalEnv := make(map[string]string)
			for _, varName := range allEnvVars {
				originalEnv[varName] = os.Getenv(varName)
			}

			// Clean up after test
			defer func() {
				for _, varName := range allEnvVars {
					if originalVal := originalEnv[varName]; originalVal != "" {
						os.Setenv(varName, originalVal)
					} else {
						os.Unsetenv(varName)
					}
				}
			}()

			// Clear all environment variables first
			for _, varName := range allEnvVars {
				os.Unsetenv(varName)
			}

			// Set up proxy variables
			if tt.setProxyVars["HTTP_PROXY"] {
				os.Setenv("HTTP_PROXY", tt.httpProxy)
			}
			if tt.setProxyVars["HTTPS_PROXY"] {
				os.Setenv("HTTPS_PROXY", tt.httpsProxy)
			}

			// Set up Docker variables
			for varName, shouldSet := range tt.setDockerVars {
				if shouldSet {
					value := tt.dockerVars[varName]
					os.Setenv(varName, value)
				}
			}

			// Create mock state
			mockSt := &mockState{vars: make(map[string]loader.EnvVar)}

			// Execute function
			err := DoSyncNetworkSettings(mockSt)

			// Check error expectation
			if (err != nil) != tt.wantErr {
				t.Errorf("DoSyncNetworkSettings() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Check all expected results
			for varName, expectedVal := range tt.expectedResults {
				if actualVar, exists := mockSt.GetVar(varName); exists {
					if actualVar.Value != expectedVal {
						t.Errorf("Expected %s = %q, got %q", varName, expectedVal, actualVar.Value)
					}
					if !actualVar.IsChanged {
						t.Errorf("Variable %s should be marked as changed", varName)
					}
				} else {
					t.Errorf("Expected variable %s to be set in state", varName)
				}
			}

			// Verify no unexpected variables were set
			allStateVars := mockSt.GetAllVars()
			for varName, actualVar := range allStateVars {
				if actualVar.Value != "" {
					if _, expected := tt.expectedResults[varName]; !expected {
						t.Errorf("Unexpected variable %s = %q was set in state", varName, actualVar.Value)
					}
				}
			}
		})
	}
}

// Test 4: Error handling scenarios
func TestDoSyncNetworkSettings_ErrorHandling(t *testing.T) {
	tests := []struct {
		name          string
		setupFunc     func(*testing.T) *mockStateWithErrors
		expectedError string
	}{
		{
			name: "SetVar error for PROXY_URL",
			setupFunc: func(t *testing.T) *mockStateWithErrors {
				return &mockStateWithErrors{
					vars:         make(map[string]loader.EnvVar),
					setVarError:  map[string]error{"PROXY_URL": mockError},
					setVarsError: nil,
				}
			},
			expectedError: "mocked error",
		},
		{
			name: "SetVars error for Docker variables",
			setupFunc: func(t *testing.T) *mockStateWithErrors {
				return &mockStateWithErrors{
					vars:         make(map[string]loader.EnvVar),
					setVarError:  nil,
					setVarsError: mockError,
				}
			},
			expectedError: "mocked error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original environment
			allEnvVars := []string{"HTTP_PROXY", "HTTPS_PROXY", "DOCKER_HOST", "DOCKER_TLS_VERIFY", "DOCKER_CERT_PATH"}
			originalEnv := make(map[string]string)
			for _, varName := range allEnvVars {
				originalEnv[varName] = os.Getenv(varName)
			}

			// Clean up after test
			defer func() {
				for _, varName := range allEnvVars {
					if originalVal := originalEnv[varName]; originalVal != "" {
						os.Setenv(varName, originalVal)
					} else {
						os.Unsetenv(varName)
					}
				}
			}()

			// Set up environment to trigger the error path
			switch tt.name {
			case "SetVar error for PROXY_URL":
				os.Setenv("HTTP_PROXY", "http://proxy.example.com:8080")
			case "SetVars error for Docker variables":
				os.Setenv("DOCKER_HOST", "tcp://docker.example.com:2376")
			}

			// Create mock state with error conditions
			mockSt := tt.setupFunc(t)

			// Execute function
			err := DoSyncNetworkSettings(mockSt)

			// Check that error was returned
			if err == nil {
				t.Error("Expected error but got none")
			} else if err.Error() != tt.expectedError {
				t.Errorf("Expected error %q, got %q", tt.expectedError, err.Error())
			}
		})
	}
}

// Test 5: Edge cases and boundary conditions
func TestDoSyncNetworkSettings_EdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		setupEnv    func()
		expectSync  bool
		description string
	}{
		{
			name: "whitespace-only proxy values will set PROXY_URL",
			setupEnv: func() {
				os.Setenv("HTTP_PROXY", "   ")
				os.Setenv("HTTPS_PROXY", "\t\n")
			},
			expectSync:  true, // Function doesn't trim whitespace, so it will sync
			description: "Whitespace-only values are treated as non-empty by the function",
		},
		{
			name: "Docker variable with whitespace-only value will trigger sync",
			setupEnv: func() {
				os.Setenv("DOCKER_HOST", "   ")
				os.Setenv("DOCKER_TLS_VERIFY", "\t")
				os.Setenv("DOCKER_CERT_PATH", "\n")
			},
			expectSync:  true, // Function doesn't trim whitespace, so it will sync
			description: "Docker variables with whitespace are treated as non-empty by the function",
		},
		{
			name: "special characters in proxy URL",
			setupEnv: func() {
				os.Setenv("HTTP_PROXY", "http://user%40domain:p%40ssw0rd@proxy.example.com:8080")
			},
			expectSync:  true,
			description: "Proxy URLs with special characters should be handled correctly",
		},
		{
			name: "truly empty proxy variables should not sync",
			setupEnv: func() {
				os.Setenv("HTTP_PROXY", "")
				os.Setenv("HTTPS_PROXY", "")
			},
			expectSync:  false,
			description: "Empty string values should not trigger sync",
		},
		{
			name: "truly empty Docker variables should not sync",
			setupEnv: func() {
				os.Setenv("DOCKER_HOST", "")
				os.Setenv("DOCKER_TLS_VERIFY", "")
				os.Setenv("DOCKER_CERT_PATH", "")
			},
			expectSync:  false,
			description: "Empty string Docker values should not trigger sync",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original environment
			allEnvVars := []string{"HTTP_PROXY", "HTTPS_PROXY", "DOCKER_HOST", "DOCKER_TLS_VERIFY", "DOCKER_CERT_PATH"}
			originalEnv := make(map[string]string)
			for _, varName := range allEnvVars {
				originalEnv[varName] = os.Getenv(varName)
			}

			// Clean up after test
			defer func() {
				for _, varName := range allEnvVars {
					if originalVal := originalEnv[varName]; originalVal != "" {
						os.Setenv(varName, originalVal)
					} else {
						os.Unsetenv(varName)
					}
				}
			}()

			// Clear environment first
			for _, varName := range allEnvVars {
				os.Unsetenv(varName)
			}

			// Set up test environment
			tt.setupEnv()

			// Create mock state
			mockSt := &mockState{vars: make(map[string]loader.EnvVar)}

			// Execute function
			err := DoSyncNetworkSettings(mockSt)
			if err != nil {
				t.Fatalf("DoSyncNetworkSettings() unexpected error = %v", err)
			}

			// Check if any variables were synced
			allStateVars := mockSt.GetAllVars()
			anyVarSet := false
			for _, envVar := range allStateVars {
				if envVar.Value != "" {
					anyVarSet = true
					break
				}
			}

			if tt.expectSync && !anyVarSet {
				t.Errorf("Expected some variables to be synced for case: %s", tt.description)
			}
			if !tt.expectSync && anyVarSet {
				t.Errorf("Expected no variables to be synced for case: %s", tt.description)
			}
		})
	}
}
