package state

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// stateEnvFile writes content to .env in a fresh directory, beside which the state keeps .state and .bak.
func stateEnvFile(t *testing.T, content string) string {
	t.Helper()
	envPath := filepath.Join(t.TempDir(), ".env")
	if err := os.WriteFile(envPath, []byte(content), 0o600); err != nil {
		t.Fatalf("write env file: %v", err)
	}
	return envPath
}

func stateFixture(t *testing.T, content string) (State, string) {
	t.Helper()
	envPath := stateEnvFile(t, content)
	s, err := NewState(envPath)
	if err != nil {
		t.Fatalf("NewState: %v", err)
	}
	return s, envPath
}

func TestState_NewState_RestoresTheStagedSessionOrFallsBackToTheEnvFile(t *testing.T) {
	stagingFile := func(t *testing.T, envPath string) string {
		t.Helper()
		path := filepath.Join(filepath.Dir(envPath), ".state", ".env.state")
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("MkdirAll: %v", err)
		}
		return path
	}

	tests := []struct {
		name      string
		setup     func(t *testing.T, envPath string) string // returns the path handed to NewState
		wantErr   string
		wantStack []string
		wantVars  map[string]string
	}{
		{
			name:     "an env file without a staged session is loaded",
			wantVars: map[string]string{"VAR1": "value1", "VAR2": "value2", "VAR3": "value3"},
		},
		{
			name: "a staged session is restored",
			setup: func(t *testing.T, envPath string) string {
				previous, err := NewState(envPath)
				if err != nil {
					t.Fatalf("NewState: %v", err)
				}
				if err := previous.SetStack([]string{"persistence_test"}); err != nil {
					t.Fatalf("SetStack: %v", err)
				}
				if err := previous.SetVar("PERSISTENT_VAR", "persistent_value"); err != nil {
					t.Fatalf("SetVar: %v", err)
				}
				return envPath
			},
			wantStack: []string{"persistence_test"},
			wantVars:  map[string]string{"VAR1": "value1", "PERSISTENT_VAR": "persistent_value"},
		},
		{
			name: "a corrupted staging file falls back to the env file",
			setup: func(t *testing.T, envPath string) string {
				if err := os.WriteFile(stagingFile(t, envPath), []byte("invalid json content"), 0o600); err != nil {
					t.Fatalf("write staging file: %v", err)
				}
				return envPath
			},
			wantVars: map[string]string{"VAR1": "value1"},
		},
		{
			name:    "a missing env file is refused",
			setup:   func(t *testing.T, envPath string) string { return filepath.Join(filepath.Dir(envPath), "absent.env") },
			wantErr: "failed to stat",
		},
		{
			name:    "a directory in place of the env file is refused",
			setup:   func(t *testing.T, envPath string) string { return filepath.Dir(envPath) },
			wantErr: "is a directory",
		},
		{
			name: "a directory in place of the staging file is refused",
			setup: func(t *testing.T, envPath string) string {
				if err := os.MkdirAll(stagingFile(t, envPath), 0o755); err != nil {
					t.Fatalf("MkdirAll: %v", err)
				}
				return envPath
			},
			wantErr: "is a directory",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := stateEnvFile(t, "VAR1=value1\nVAR2=value2\n# Comment\nVAR3=value3")
			if tt.setup != nil {
				path = tt.setup(t, path)
			}

			s, err := NewState(path)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("NewState() error = %v, want it to mention %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("NewState() error = %v", err)
			}

			if s.GetEnvPath() != path {
				t.Errorf("GetEnvPath() = %q, want %q", s.GetEnvPath(), path)
			}
			if got := s.GetStack(); strings.Join(got, ",") != strings.Join(tt.wantStack, ",") {
				t.Errorf("GetStack() = %q, want %q", got, tt.wantStack)
			}
			for name, value := range tt.wantVars {
				if envVar, exists := s.GetVar(name); !exists || envVar.Value != value {
					t.Errorf("GetVar(%s) = %q, %v; want %q", name, envVar.Value, exists, value)
				}
			}
		})
	}
}

func TestState_Exists_ReportsTheStagingFile(t *testing.T) {
	s, _ := stateFixture(t, "VAR1=value1")
	if s.Exists() {
		t.Error("Exists() = true before anything was staged")
	}

	if err := s.SetVar("NEW_VAR", "new_value"); err != nil {
		t.Fatalf("SetVar: %v", err)
	}
	if !s.Exists() {
		t.Error("Exists() = false after a variable was staged")
	}
}

func TestState_SetStack_ReplacesTheWizardStack(t *testing.T) {
	s, _ := stateFixture(t, "VAR1=value1")
	if stack := s.GetStack(); len(stack) != 0 {
		t.Errorf("GetStack() = %q before anything was set, want empty", stack)
	}

	if err := s.SetStack([]string{"configure_database"}); err != nil {
		t.Fatalf("SetStack: %v", err)
	}
	if got := s.GetStack(); !reflect.DeepEqual(got, []string{"configure_database"}) {
		t.Errorf("GetStack() = %q, want [configure_database]", got)
	}

	if err := s.SetStack(append(s.GetStack(), "configure_api")); err != nil {
		t.Fatalf("SetStack: %v", err)
	}
	if got := s.GetStack(); !reflect.DeepEqual(got, []string{"configure_database", "configure_api"}) {
		t.Errorf("GetStack() = %q, want [configure_database configure_api]", got)
	}
}

// TestState_StagesAndResetsVariables: subtests run in order over one state and are keyed by method.
func TestState_StagesAndResetsVariables(t *testing.T) {
	s, _ := stateFixture(t, "EXISTING_VAR=existing_value")
	staged := func(t *testing.T, name, value string) {
		t.Helper()
		if envVar, exists := s.GetVar(name); !exists || envVar.Value != value || !envVar.IsChanged {
			t.Errorf("GetVar(%s) = %+v, %v; want a staged %q", name, envVar, exists, value)
		}
	}

	t.Run("get var reads the env file", func(t *testing.T) {
		if envVar, exists := s.GetVar("EXISTING_VAR"); !exists || envVar.Value != "existing_value" {
			t.Errorf("GetVar(EXISTING_VAR) = %q, %v; want existing_value", envVar.Value, exists)
		}
	})

	t.Run("set var stages a new variable", func(t *testing.T) {
		if err := s.SetVar("NEW_VAR", "new_value"); err != nil {
			t.Fatalf("SetVar: %v", err)
		}
		staged(t, "NEW_VAR", "new_value")
	})

	t.Run("set var stages an update", func(t *testing.T) {
		if err := s.SetVar("EXISTING_VAR", "updated_value"); err != nil {
			t.Fatalf("SetVar: %v", err)
		}
		staged(t, "EXISTING_VAR", "updated_value")
	})

	t.Run("get vars reports values and presence", func(t *testing.T) {
		vars, present := s.GetVars([]string{"EXISTING_VAR", "NEW_VAR", "NON_EXISTENT"})

		if len(vars) != 3 || len(present) != 3 {
			t.Errorf("GetVars() = %d values and %d flags, want 3 and 3", len(vars), len(present))
		}
		if !present["EXISTING_VAR"] || !present["NEW_VAR"] || present["NON_EXISTENT"] {
			t.Errorf("presence = %v, want EXISTING_VAR and NEW_VAR only", present)
		}
		if vars["EXISTING_VAR"].Value != "updated_value" || vars["NEW_VAR"].Value != "new_value" {
			t.Errorf("values = %q, %q; want updated_value, new_value", vars["EXISTING_VAR"].Value, vars["NEW_VAR"].Value)
		}
	})

	t.Run("set vars stages every variable", func(t *testing.T) {
		vars := map[string]string{"BATCH_VAR1": "batch_value1", "BATCH_VAR2": "batch_value2", "EXISTING_VAR": "batch_updated"}
		if err := s.SetVars(vars); err != nil {
			t.Fatalf("SetVars: %v", err)
		}
		for name, value := range vars {
			staged(t, name, value)
		}
	})

	t.Run("reset var restores the env file value", func(t *testing.T) {
		if err := s.SetVar("EXISTING_VAR", "modified_again"); err != nil {
			t.Fatalf("SetVar: %v", err)
		}
		if err := s.ResetVar("EXISTING_VAR"); err != nil {
			t.Fatalf("ResetVar: %v", err)
		}
		if envVar, exists := s.GetVar("EXISTING_VAR"); !exists || envVar.Value != "existing_value" {
			t.Errorf("GetVar(EXISTING_VAR) = %q, %v; want existing_value", envVar.Value, exists)
		}
	})

	t.Run("reset vars drops variables the env file lacks", func(t *testing.T) {
		if err := s.SetVars(map[string]string{"RESET_VAR1": "reset_value1", "RESET_VAR2": "reset_value2", "EXISTING_VAR": "modified_value"}); err != nil {
			t.Fatalf("SetVars: %v", err)
		}
		if err := s.ResetVars([]string{"RESET_VAR1", "RESET_VAR2", "EXISTING_VAR"}); err != nil {
			t.Fatalf("ResetVars: %v", err)
		}

		for _, name := range []string{"RESET_VAR1", "RESET_VAR2"} {
			if _, exists := s.GetVar(name); exists {
				t.Errorf("%s still exists after ResetVars", name)
			}
		}
		if envVar, exists := s.GetVar("EXISTING_VAR"); !exists || envVar.Value != "existing_value" {
			t.Errorf("GetVar(EXISTING_VAR) = %q, %v; want existing_value", envVar.Value, exists)
		}
	})

	t.Run("reset var of an unknown variable leaves it absent", func(t *testing.T) {
		if err := s.ResetVar("NON_EXISTENT_VAR"); err != nil {
			t.Fatalf("ResetVar: %v", err)
		}
		if _, exists := s.GetVar("NON_EXISTENT_VAR"); exists {
			t.Error("NON_EXISTENT_VAR exists after ResetVar")
		}
	})

	t.Run("get all vars includes the staged variables", func(t *testing.T) {
		allVars := s.GetAllVars()
		want := map[string]string{"EXISTING_VAR": "existing_value", "NEW_VAR": "new_value", "BATCH_VAR1": "batch_value1"}
		for name, value := range want {
			if allVars[name].Value != value {
				t.Errorf("GetAllVars()[%s] = %q, want %q", name, allVars[name].Value, value)
			}
		}
	})
}

func TestState_Commit_WritesStagedChangesAndLeavesNothingPending(t *testing.T) {
	s, envPath := stateFixture(t, "ORIGINAL_VAR=original_value")
	if s.IsDirty() {
		t.Error("IsDirty() = true without a staging file")
	}

	if err := s.SetStack([]string{"testing_commit"}); err != nil {
		t.Fatalf("SetStack: %v", err)
	}
	if err := s.SetVar("ORIGINAL_VAR", "modified_value"); err != nil {
		t.Fatalf("SetVar: %v", err)
	}
	if err := s.SetVar("NEW_VAR", "new_value"); err != nil {
		t.Fatalf("SetVar: %v", err)
	}
	if !s.Exists() || !s.IsDirty() {
		t.Errorf("Exists(), IsDirty() = %v, %v with staged changes; want true, true", s.Exists(), s.IsDirty())
	}

	if err := s.Commit(); err != nil {
		t.Fatalf("Commit: %v", err)
	}

	if !s.Exists() || s.IsDirty() {
		t.Errorf("Exists(), IsDirty() = %v, %v after Commit; want true, false", s.Exists(), s.IsDirty())
	}
	if content, err := os.ReadFile(envPath); err != nil || string(content) != "ORIGINAL_VAR=modified_value\nNEW_VAR=new_value" {
		t.Errorf("env file = %q, %v; want the staged changes", content, err)
	}
	backups, err := os.ReadDir(filepath.Join(filepath.Dir(envPath), ".bak"))
	if err != nil || len(backups) != 1 {
		t.Fatalf("backups = %v, %v; want one", backups, err)
	}
	if backup, err := os.ReadFile(filepath.Join(filepath.Dir(envPath), ".bak", backups[0].Name())); err != nil || string(backup) != "ORIGINAL_VAR=original_value" {
		t.Errorf("backup = %q, %v; want the original env file", backup, err)
	}
}

func TestState_Reset_DiscardsStagedChangesAndKeepsTheEnvFile(t *testing.T) {
	tests := []struct {
		name  string
		stage bool
	}{
		{name: "staged changes are discarded", stage: true},
		{name: "nothing staged is not an error", stage: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, envPath := stateFixture(t, "ORIGINAL_VAR=original_value")
			if tt.stage {
				if err := s.SetStack([]string{"testing_reset"}); err != nil {
					t.Fatalf("SetStack: %v", err)
				}
				if err := s.SetVar("ORIGINAL_VAR", "modified_value"); err != nil {
					t.Fatalf("SetVar: %v", err)
				}
			}

			if err := s.Reset(); err != nil {
				t.Fatalf("Reset: %v", err)
			}

			if envVar, _ := s.GetVar("ORIGINAL_VAR"); envVar.Value != "original_value" || envVar.IsChanged {
				t.Errorf("GetVar(ORIGINAL_VAR) = %+v, want the unchanged original_value", envVar)
			}
			if !s.Exists() || s.IsDirty() {
				t.Errorf("Exists(), IsDirty() = %v, %v after Reset; want true, false", s.Exists(), s.IsDirty())
			}
			if content, err := os.ReadFile(envPath); err != nil || string(content) != "ORIGINAL_VAR=original_value" {
				t.Errorf("env file = %q, %v; want it untouched", content, err)
			}
		})
	}
}

func TestState_SetEulaConsent_IsRememberedAcrossSessions(t *testing.T) {
	s, envPath := stateFixture(t, "VAR1=value1")
	if s.GetEulaConsent() {
		t.Fatal("GetEulaConsent() = true before consent was given")
	}

	if err := s.SetEulaConsent(); err != nil {
		t.Fatalf("SetEulaConsent: %v", err)
	}

	next, err := NewState(envPath)
	if err != nil || !next.GetEulaConsent() {
		t.Errorf("a new session does not see the consent (%v)", err)
	}
}

func TestState_WriteVars_SavesToTheEnvFileWithoutStagedEdits(t *testing.T) {
	s, envPath := stateFixture(t, "VAR1=value1\n")
	if err := s.SetVar("VAR1", "typed_in_a_form"); err != nil {
		t.Fatalf("SetVar: %v", err)
	}

	if err := s.WriteVars(map[string]string{"PENTAGI_IMAGE": "vxcontrol/pentagi:2.3.4"}); err != nil {
		t.Fatalf("WriteVars: %v", err)
	}

	if content, err := os.ReadFile(envPath); err != nil || string(content) != "VAR1=value1\nPENTAGI_IMAGE=vxcontrol/pentagi:2.3.4\n" {
		t.Errorf("env file = %q, %v; want the written variable and not the staged edit", content, err)
	}
	if envVar, _ := s.GetVar("PENTAGI_IMAGE"); envVar.Value != "vxcontrol/pentagi:2.3.4" {
		t.Errorf("GetVar(PENTAGI_IMAGE) = %q; a later Commit would write the old value back", envVar.Value)
	}
	if envVar, _ := s.GetVar("VAR1"); envVar.Value != "typed_in_a_form" || !envVar.IsChanged {
		t.Errorf("GetVar(VAR1) = %+v; want the staged edit kept for the user to apply", envVar)
	}
}
