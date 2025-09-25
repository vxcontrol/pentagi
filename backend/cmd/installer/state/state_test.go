package state

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewState(t *testing.T) {
	t.Run("create new state from existing env file", func(t *testing.T) {
		tmpDir := t.TempDir()
		defer os.RemoveAll(tmpDir)

		envPath := filepath.Join(tmpDir, ".env")

		// Create test .env file
		envContent := `VAR1=value1
VAR2=value2
# Comment
VAR3=value3`
		err := os.WriteFile(envPath, []byte(envContent), 0644)
		if err != nil {
			t.Fatalf("Failed to create test env file: %v", err)
		}

		state, err := NewState(envPath)
		if err != nil {
			t.Fatalf("Failed to create new state: %v", err)
		}

		if state == nil {
			t.Fatal("Expected state to be non-nil")
		}

		// Check that variables were loaded
		allVars := state.GetAllVars()
		if len(allVars) < 3 {
			t.Errorf("Expected at least 3 variables, got %d", len(allVars))
		}

		if envVar, exists := state.GetVar("VAR1"); !exists {
			t.Error("Expected VAR1 to exist")
		} else if envVar.Value != "value1" {
			t.Errorf("Expected VAR1 value 'value1', got '%s'", envVar.Value)
		}
	})

	t.Run("create state with non-existent env file", func(t *testing.T) {
		_, err := NewState("/non/existent/file.env")
		if err == nil {
			t.Error("Expected error when creating state with non-existent file")
		}
	})

	t.Run("create state with directory instead of file", func(t *testing.T) {
		tmpDir := t.TempDir()
		defer os.RemoveAll(tmpDir)

		_, err := NewState(tmpDir)
		if err == nil {
			t.Error("Expected error when creating state with directory")
		}
	})
}

func TestStateExists(t *testing.T) {
	tmpDir := t.TempDir()
	defer os.RemoveAll(tmpDir)

	envPath := filepath.Join(tmpDir, ".env")

	// Create test .env file
	err := os.WriteFile(envPath, []byte("VAR1=value1"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test env file: %v", err)
	}

	state, err := NewState(envPath)
	if err != nil {
		t.Fatalf("Failed to create state: %v", err)
	}

	// Initially no state file exists
	if state.Exists() {
		t.Error("Expected state to not exist initially")
	}

	// After setting a variable, state should exist
	err = state.SetVar("NEW_VAR", "new_value")
	if err != nil {
		t.Fatalf("Failed to set variable: %v", err)
	}

	if !state.Exists() {
		t.Error("Expected state to exist after setting variable")
	}
}

func TestStateStepManagement(t *testing.T) {
	tmpDir := t.TempDir()
	defer os.RemoveAll(tmpDir)

	envPath := filepath.Join(tmpDir, ".env")

	err := os.WriteFile(envPath, []byte("VAR1=value1"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test env file: %v", err)
	}

	state, err := NewState(envPath)
	if err != nil {
		t.Fatalf("Failed to create state: %v", err)
	}

	// Initially no step
	if stack := state.GetStack(); len(stack) > 0 {
		t.Errorf("Expected empty stack initially, got '%s'", stack)
	}

	// Set step
	err = state.SetStack([]string{"configure_database"})
	if err != nil {
		t.Fatalf("Failed to set step: %v", err)
	}

	if stack := state.GetStack(); len(stack) < 1 {
		t.Errorf("Expected step 'configure_database', got '%s'", stack)
	}

	// Append step
	err = state.SetStack(append(state.GetStack(), "configure_api"))
	if err != nil {
		t.Fatalf("Failed to update step: %v", err)
	}

	if stack := state.GetStack(); len(stack) < 2 {
		t.Errorf("Expected step 'configure_api', got '%s'", stack)
	}
}

func TestStateVariableManagement(t *testing.T) {
	tmpDir := t.TempDir()
	defer os.RemoveAll(tmpDir)

	envPath := filepath.Join(tmpDir, ".env")

	err := os.WriteFile(envPath, []byte("EXISTING_VAR=existing_value"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test env file: %v", err)
	}

	state, err := NewState(envPath)
	if err != nil {
		t.Fatalf("Failed to create state: %v", err)
	}

	t.Run("get existing variable", func(t *testing.T) {
		envVar, exists := state.GetVar("EXISTING_VAR")
		if !exists {
			t.Error("Expected EXISTING_VAR to exist")
		}
		if envVar.Value != "existing_value" {
			t.Errorf("Expected value 'existing_value', got '%s'", envVar.Value)
		}
	})

	t.Run("set new variable", func(t *testing.T) {
		err := state.SetVar("NEW_VAR", "new_value")
		if err != nil {
			t.Fatalf("Failed to set new variable: %v", err)
		}

		envVar, exists := state.GetVar("NEW_VAR")
		if !exists {
			t.Error("Expected NEW_VAR to exist")
		}
		if envVar.Value != "new_value" {
			t.Errorf("Expected value 'new_value', got '%s'", envVar.Value)
		}
		if !envVar.IsChanged {
			t.Error("Expected IsChanged to be true for new variable")
		}
	})

	t.Run("update existing variable", func(t *testing.T) {
		err := state.SetVar("EXISTING_VAR", "updated_value")
		if err != nil {
			t.Fatalf("Failed to update existing variable: %v", err)
		}

		envVar, exists := state.GetVar("EXISTING_VAR")
		if !exists {
			t.Error("Expected EXISTING_VAR to exist")
		}
		if envVar.Value != "updated_value" {
			t.Errorf("Expected value 'updated_value', got '%s'", envVar.Value)
		}
		if !envVar.IsChanged {
			t.Error("Expected IsChanged to be true for updated variable")
		}
	})

	t.Run("get multiple variables", func(t *testing.T) {
		names := []string{"EXISTING_VAR", "NEW_VAR", "NON_EXISTENT"}
		vars, present := state.GetVars(names)

		if len(vars) != 3 {
			t.Errorf("Expected 3 variables in result, got %d", len(vars))
		}
		if len(present) != 3 {
			t.Errorf("Expected 3 presence flags, got %d", len(present))
		}

		if !present["EXISTING_VAR"] {
			t.Error("Expected EXISTING_VAR to be present")
		}
		if !present["NEW_VAR"] {
			t.Error("Expected NEW_VAR to be present")
		}
		if present["NON_EXISTENT"] {
			t.Error("Expected NON_EXISTENT to not be present")
		}

		if vars["EXISTING_VAR"].Value != "updated_value" {
			t.Errorf("Expected EXISTING_VAR value 'updated_value', got '%s'", vars["EXISTING_VAR"].Value)
		}
		if vars["NEW_VAR"].Value != "new_value" {
			t.Errorf("Expected NEW_VAR value 'new_value', got '%s'", vars["NEW_VAR"].Value)
		}
	})

	t.Run("set multiple variables", func(t *testing.T) {
		vars := map[string]string{
			"BATCH_VAR1":   "batch_value1",
			"BATCH_VAR2":   "batch_value2",
			"EXISTING_VAR": "batch_updated",
		}

		err := state.SetVars(vars)
		if err != nil {
			t.Fatalf("Failed to set multiple variables: %v", err)
		}

		for name, expectedValue := range vars {
			envVar, exists := state.GetVar(name)
			if !exists {
				t.Errorf("Expected variable %s to exist after SetVars", name)
				continue
			}
			if envVar.Value != expectedValue {
				t.Errorf("Variable %s: expected value %s, got %s", name, expectedValue, envVar.Value)
			}
			if !envVar.IsChanged {
				t.Errorf("Variable %s: expected IsChanged to be true", name)
			}
		}
	})

	t.Run("reset single variable", func(t *testing.T) {
		// First modify a variable
		err := state.SetVar("EXISTING_VAR", "modified_again")
		if err != nil {
			t.Fatalf("Failed to modify variable: %v", err)
		}

		// Verify it was changed
		envVar, exists := state.GetVar("EXISTING_VAR")
		if !exists || envVar.Value != "modified_again" {
			t.Fatalf("Variable was not modified as expected")
		}

		// Reset it
		err = state.ResetVar("EXISTING_VAR")
		if err != nil {
			t.Fatalf("Failed to reset variable: %v", err)
		}

		// Verify it was reset to original value
		envVar, exists = state.GetVar("EXISTING_VAR")
		if !exists {
			t.Error("Expected EXISTING_VAR to exist after reset")
		}
		if envVar.Value != "existing_value" {
			t.Errorf("Expected EXISTING_VAR to be reset to 'existing_value', got '%s'", envVar.Value)
		}
	})

	t.Run("reset multiple variables", func(t *testing.T) {
		// Set some variables first
		vars := map[string]string{
			"RESET_VAR1":   "reset_value1",
			"RESET_VAR2":   "reset_value2",
			"EXISTING_VAR": "modified_value",
		}
		err := state.SetVars(vars)
		if err != nil {
			t.Fatalf("Failed to set variables: %v", err)
		}

		// Reset multiple variables
		names := []string{"RESET_VAR1", "RESET_VAR2", "EXISTING_VAR"}
		err = state.ResetVars(names)
		if err != nil {
			t.Fatalf("Failed to reset variables: %v", err)
		}

		// RESET_VAR1 and RESET_VAR2 should be deleted (not in original file)
		if _, exists := state.GetVar("RESET_VAR1"); exists {
			t.Error("Expected RESET_VAR1 to be deleted after reset")
		}
		if _, exists := state.GetVar("RESET_VAR2"); exists {
			t.Error("Expected RESET_VAR2 to be deleted after reset")
		}

		// EXISTING_VAR should be reset to original value
		envVar, exists := state.GetVar("EXISTING_VAR")
		if !exists {
			t.Error("Expected EXISTING_VAR to exist after reset")
		}
		if envVar.Value != "existing_value" {
			t.Errorf("Expected EXISTING_VAR to be reset to 'existing_value', got '%s'", envVar.Value)
		}
	})

	t.Run("reset non-existent variable", func(t *testing.T) {
		err := state.ResetVar("NON_EXISTENT_VAR")
		if err != nil {
			t.Errorf("Expected reset of non-existent variable to succeed, got error: %v", err)
		}
	})

	t.Run("get all variables", func(t *testing.T) {
		allVars := state.GetAllVars()
		if len(allVars) < 2 {
			t.Errorf("Expected at least 2 variables, got %d", len(allVars))
		}

		if _, exists := allVars["EXISTING_VAR"]; !exists {
			t.Error("Expected EXISTING_VAR in GetAllVars result")
		}
		if _, exists := allVars["NEW_VAR"]; !exists {
			t.Error("Expected NEW_VAR in GetAllVars result")
		}
	})
}

func TestStateCommit(t *testing.T) {
	tmpDir := t.TempDir()
	defer os.RemoveAll(tmpDir)

	envPath := filepath.Join(tmpDir, ".env")

	originalContent := "ORIGINAL_VAR=original_value"
	err := os.WriteFile(envPath, []byte(originalContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test env file: %v", err)
	}

	state, err := NewState(envPath)
	if err != nil {
		t.Fatalf("Failed to create state: %v", err)
	}

	// Make changes
	err = state.SetStack([]string{"testing_commit"})
	if err != nil {
		t.Fatalf("Failed to set step: %v", err)
	}

	err = state.SetVar("ORIGINAL_VAR", "modified_value")
	if err != nil {
		t.Fatalf("Failed to set variable: %v", err)
	}

	err = state.SetVar("NEW_VAR", "new_value")
	if err != nil {
		t.Fatalf("Failed to set new variable: %v", err)
	}

	// Verify state exists
	if !state.Exists() {
		t.Error("Expected state to exist before commit")
	}

	// Commit changes
	err = state.Commit()
	if err != nil {
		t.Fatalf("Failed to commit state: %v", err)
	}

	// Verify state file was reloaded and exists
	if !state.Exists() {
		t.Error("Expected state to exist after commit")
	}

	// Verify .env file was updated
	content, err := os.ReadFile(envPath)
	if err != nil {
		t.Fatalf("Failed to read env file after commit: %v", err)
	}

	contentStr := string(content)
	if !containsLine(contentStr, "ORIGINAL_VAR=modified_value") {
		t.Error("Expected ORIGINAL_VAR to be updated in env file")
	}
	if !containsLine(contentStr, "NEW_VAR=new_value") {
		t.Error("Expected NEW_VAR to be added to env file")
	}

	// Verify backup was created
	backupDir := filepath.Join(tmpDir, ".bak")
	entries, err := os.ReadDir(backupDir)
	if err != nil {
		t.Fatalf("Failed to read backup directory: %v", err)
	}
	if len(entries) == 0 {
		t.Error("Expected backup file to be created")
	}
}

func TestStateReset(t *testing.T) {
	tmpDir := t.TempDir()
	defer os.RemoveAll(tmpDir)

	envPath := filepath.Join(tmpDir, ".env")

	originalContent := "ORIGINAL_VAR=original_value"
	err := os.WriteFile(envPath, []byte(originalContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test env file: %v", err)
	}

	state, err := NewState(envPath)
	if err != nil {
		t.Fatalf("Failed to create state: %v", err)
	}

	// Make changes
	err = state.SetStack([]string{"testing_reset"})
	if err != nil {
		t.Fatalf("Failed to set step: %v", err)
	}

	err = state.SetVar("ORIGINAL_VAR", "modified_value")
	if err != nil {
		t.Fatalf("Failed to set variable: %v", err)
	}

	// Verify state exists
	if !state.Exists() {
		t.Error("Expected state to exist before reset")
	}

	// Reset state
	err = state.Reset()
	if err != nil {
		t.Fatalf("Failed to reset state: %v", err)
	}

	// Verify state file was reloaded and exists
	if !state.Exists() {
		t.Error("Expected state to exist after reset")
	}

	// Verify .env file was NOT changed
	content, err := os.ReadFile(envPath)
	if err != nil {
		t.Fatalf("Failed to read env file after reset: %v", err)
	}

	if string(content) != originalContent {
		t.Errorf("Expected env file to remain unchanged after reset, got: %s", string(content))
	}
}

func TestStatePersistence(t *testing.T) {
	tmpDir := t.TempDir()
	defer os.RemoveAll(tmpDir)

	envPath := filepath.Join(tmpDir, ".env")

	err := os.WriteFile(envPath, []byte("VAR1=value1"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test env file: %v", err)
	}

	// Create first state instance and make changes
	state1, err := NewState(envPath)
	if err != nil {
		t.Fatalf("Failed to create first state: %v", err)
	}

	err = state1.SetStack([]string{"persistence_test"})
	if err != nil {
		t.Fatalf("Failed to set step: %v", err)
	}

	err = state1.SetVar("PERSISTENT_VAR", "persistent_value")
	if err != nil {
		t.Fatalf("Failed to set variable: %v", err)
	}

	// Create second state instance (should load saved state)
	state2, err := NewState(envPath)
	if err != nil {
		t.Fatalf("Failed to create second state: %v", err)
	}

	// Verify step was restored
	if step := state2.GetStack()[0]; step != "persistence_test" {
		t.Errorf("Expected step 'persistence_test', got '%s'", step)
	}

	// Verify variable was restored
	envVar, exists := state2.GetVar("PERSISTENT_VAR")
	if !exists {
		t.Error("Expected PERSISTENT_VAR to exist in restored state")
	}
	if envVar.Value != "persistent_value" {
		t.Errorf("Expected value 'persistent_value', got '%s'", envVar.Value)
	}
}

func TestStateErrors(t *testing.T) {
	t.Run("state file is directory", func(t *testing.T) {
		tmpDir := t.TempDir()
		defer os.RemoveAll(tmpDir)

		envPath := filepath.Join(tmpDir, ".env")

		err := os.WriteFile(envPath, []byte("VAR1=value1"), 0644)
		if err != nil {
			t.Fatalf("Failed to create test env file: %v", err)
		}

		// Create directory where state file should be
		stateDir := filepath.Join(tmpDir, ".state")
		statePath := filepath.Join(stateDir, ".env.state")
		err = os.MkdirAll(statePath, 0755) // Create directory instead of file
		if err != nil {
			t.Fatalf("Failed to create state directory: %v", err)
		}

		_, err = NewState(envPath)
		if err == nil {
			t.Error("Expected error when state file is directory")
		}
	})

	t.Run("corrupted state file", func(t *testing.T) {
		tmpDir := t.TempDir()
		defer os.RemoveAll(tmpDir)

		envPath := filepath.Join(tmpDir, ".env")

		err := os.WriteFile(envPath, []byte("VAR1=value1"), 0644)
		if err != nil {
			t.Fatalf("Failed to create test env file: %v", err)
		}

		// Create corrupted state file
		stateDir := filepath.Join(tmpDir, ".state")
		err = os.MkdirAll(stateDir, 0755)
		if err != nil {
			t.Fatalf("Failed to create state directory: %v", err)
		}

		statePath := filepath.Join(stateDir, ".env.state")
		err = os.WriteFile(statePath, []byte("invalid json content"), 0644)
		if err != nil {
			t.Fatalf("Failed to create corrupted state file: %v", err)
		}

		// try to reload original env file if state file is corrupted
		state, err := NewState(envPath)
		if err != nil {
			t.Errorf("Expected reload of original env file when state file is corrupted: %v", err)
		} else {
			envVar, exist := state.GetVar("VAR1")
			if !exist {
				t.Error("Expected VAR1 to exist in restored state")
			}
			if envVar.Value != "value1" {
				t.Errorf("Expected value 'value1', got '%s'", envVar.Value)
			}
		}
	})

	t.Run("empty state file", func(t *testing.T) {
		tmpDir := t.TempDir()
		defer os.RemoveAll(tmpDir)

		envPath := filepath.Join(tmpDir, ".env")

		err := os.WriteFile(envPath, []byte("VAR1=value1"), 0644)
		if err != nil {
			t.Fatalf("Failed to create test env file: %v", err)
		}

		// Create empty state file
		stateDir := filepath.Join(tmpDir, ".state")
		err = os.MkdirAll(stateDir, 0755)
		if err != nil {
			t.Fatalf("Failed to create state directory: %v", err)
		}

		statePath := filepath.Join(stateDir, ".env.state")
		err = os.WriteFile(statePath, []byte(""), 0644)
		if err != nil {
			t.Fatalf("Failed to create empty state file: %v", err)
		}

		state, err := NewState(envPath)
		if err != nil {
			t.Errorf("Expected reload of original env file when state file is empty: %v", err)
		} else {
			envVar, exist := state.GetVar("VAR1")
			if !exist {
				t.Error("Expected VAR1 to exist in restored state")
			}
			if envVar.Value != "value1" {
				t.Errorf("Expected value 'value1', got '%s'", envVar.Value)
			}
		}
	})

	t.Run("reset non-existent state should succeed", func(t *testing.T) {
		tmpDir := t.TempDir()
		defer os.RemoveAll(tmpDir)

		envPath := filepath.Join(tmpDir, ".env")

		err := os.WriteFile(envPath, []byte("VAR1=value1"), 0644)
		if err != nil {
			t.Fatalf("Failed to create test env file: %v", err)
		}

		state, err := NewState(envPath)
		if err != nil {
			t.Fatalf("Failed to create state: %v", err)
		}

		// Reset when no state file exists should succeed (idempotent operation)
		err = state.Reset()
		if err != nil {
			t.Errorf("Expected reset of non-existent state to succeed, got error: %v", err)
		}

		// Multiple resets should also succeed
		err = state.Reset()
		if err != nil {
			t.Errorf("Expected multiple resets to succeed, got error: %v", err)
		}
	})

	t.Run("reset after commit should succeed", func(t *testing.T) {
		tmpDir := t.TempDir()
		defer os.RemoveAll(tmpDir)

		envPath := filepath.Join(tmpDir, ".env")

		err := os.WriteFile(envPath, []byte("VAR1=value1"), 0644)
		if err != nil {
			t.Fatalf("Failed to create test env file: %v", err)
		}

		state, err := NewState(envPath)
		if err != nil {
			t.Fatalf("Failed to create state: %v", err)
		}

		// Make changes
		err = state.SetVar("NEW_VAR", "new_value")
		if err != nil {
			t.Fatalf("Failed to set variable: %v", err)
		}

		// Commit (which should reset state internally)
		err = state.Commit()
		if err != nil {
			t.Fatalf("Failed to commit: %v", err)
		}

		// Additional reset should still succeed
		err = state.Reset()
		if err != nil {
			t.Errorf("Expected reset after commit to succeed, got error: %v", err)
		}
	})
}

func containsLine(content, line string) bool {
	lines := strings.Split(content, "\n")
	for _, l := range lines {
		if strings.TrimSpace(l) == line {
			return true
		}
	}
	return false
}
