package hardening

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"pentagi/cmd/installer/loader"
)

var errMocked = errors.New("mocked error")

// hardeningHostPaths returns a directory and a regular file that exist on the host.
func hardeningHostPaths(t *testing.T) (dir, file string) {
	t.Helper()
	dir = t.TempDir()
	file = filepath.Join(t.TempDir(), "custom.provider.yml")
	if err := os.WriteFile(file, nil, 0o600); err != nil {
		t.Fatalf("write %s: %v", file, err)
	}
	return dir, file
}

// mockState stages writes into vars the way state.State does, and counts commits.
type mockState struct {
	vars         map[string]loader.EnvVar
	setVarError  map[string]error
	setVarsError error
	commits      int
}

func (m *mockState) GetVar(key string) (loader.EnvVar, bool) {
	if val, exists := m.vars[key]; exists {
		return val, true
	}
	return loader.EnvVar{Name: key, Line: -1}, false
}

func (m *mockState) GetVars(names []string) (map[string]loader.EnvVar, map[string]bool) {
	vars := make(map[string]loader.EnvVar, len(names))
	present := make(map[string]bool, len(names))
	for _, name := range names {
		vars[name], present[name] = m.GetVar(name)
	}
	return vars, present
}

func (m *mockState) SetVar(name, value string) error {
	if err := m.setVarError[name]; err != nil {
		return err
	}
	if m.vars == nil {
		m.vars = make(map[string]loader.EnvVar)
	}
	envVar := m.vars[name]
	envVar.Name, envVar.Value, envVar.IsChanged = name, value, true
	m.vars[name] = envVar
	return nil
}

func (m *mockState) SetVars(vars map[string]string) error {
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

func (m *mockState) Commit() error                          { m.commits++; return nil }
func (m *mockState) WriteVars(vars map[string]string) error { return m.SetVars(vars) }
func (m *mockState) GetAllVars() map[string]loader.EnvVar   { return m.vars }
func (m *mockState) GetEnvPath() string                     { return "" }
func (m *mockState) Exists() bool                           { return true }
func (m *mockState) Reset() error                           { return nil }
func (m *mockState) IsDirty() bool                          { return false }
func (m *mockState) GetEulaConsent() bool                   { return true }
func (m *mockState) SetEulaConsent() error                  { return nil }
func (m *mockState) SetStack(stack []string) error          { return nil }
func (m *mockState) GetStack() []string                     { return []string{} }
func (m *mockState) ResetVar(name string) error             { return nil }
func (m *mockState) ResetVars(names []string) error         { return nil }
