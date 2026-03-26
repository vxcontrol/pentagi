package loader

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

func containsLine(content, line string) bool {
	lines := strings.Split(content, "\n")
	for _, l := range lines {
		if strings.TrimSpace(l) == line {
			return true
		}
	}
	return false
}

func TestLoadEnvFile(t *testing.T) {
	tests := []struct {
		name    string
		content string
		wantErr bool
	}{
		{
			name: "valid file",
			content: `# Comment
VAR1=value1
VAR2=value2
# Another comment
VAR3=value3`,
			wantErr: false,
		},
		{
			name:    "empty file",
			content: "",
			wantErr: false,
		},
		{
			name:    "comment only",
			content: "# Just a comment",
			wantErr: false,
		},
		{
			name: "malformed lines",
			content: `VAR1=value1
invalid line
VAR2=value2`,
			wantErr: false,
		},
		{
			name: "comments in value",
			content: `VAR1=value1 # comment
VAR2=value2 # comment`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpFile := createTempFile(t, tt.content)
			defer os.Remove(tmpFile)

			envFile, err := LoadEnvFile(tmpFile)
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadEnvFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && envFile == nil {
				t.Error("Expected envFile to be non-nil")
			}
		})
	}
}

func TestLoadEnvFileErrors(t *testing.T) {
	t.Run("non-existent file", func(t *testing.T) {
		_, err := LoadEnvFile("/non/existent/file")
		if err == nil {
			t.Error("Expected error for non-existent file")
		}
	})

	t.Run("directory instead of file", func(t *testing.T) {
		tmpDir := t.TempDir()
		defer os.RemoveAll(tmpDir)

		_, err := LoadEnvFile(tmpDir)
		if err == nil {
			t.Error("Expected error when path is directory")
		}
	})
}

func TestEnvVarMethods(t *testing.T) {
	envVar := &EnvVar{
		Name:    "TEST_VAR",
		Value:   "test_value",
		Default: "default_value",
		Line:    5,
	}

	if envVar.IsDefault() {
		t.Error("Expected IsDefault() to be false")
	}

	if !envVar.IsPresent() {
		t.Error("Expected IsPresent() to be true")
	}

	envVar.Value = "default_value"
	if !envVar.IsDefault() {
		t.Error("Expected IsDefault() to be true")
	}

	envVar.Line = -1
	if envVar.IsPresent() {
		t.Error("Expected IsPresent() to be false")
	}
}

func TestEnvFileSetGet(t *testing.T) {
	envFile := &envFile{
		vars: make(map[string]*EnvVar),
		mx:   &sync.Mutex{},
	}

	t.Run("set new variable", func(t *testing.T) {
		envFile.Set("NEW_VAR", "new_value")

		envVar, exists := envFile.Get("NEW_VAR")
		if !exists {
			t.Error("Expected NEW_VAR to exist")
		}
		if envVar.Name != "NEW_VAR" || envVar.Value != "new_value" {
			t.Errorf("Expected NEW_VAR=new_value, got %s=%s", envVar.Name, envVar.Value)
		}
		if !envVar.IsChanged {
			t.Error("Expected IsChanged to be true for new variable")
		}
		if envVar.Line != -1 {
			t.Error("Expected Line to be -1 for new variable")
		}
	})

	t.Run("update existing variable", func(t *testing.T) {
		envFile.Set("NEW_VAR", "updated_value")

		envVar, exists := envFile.Get("NEW_VAR")
		if !exists {
			t.Error("Expected NEW_VAR to exist")
		}
		if envVar.Value != "updated_value" {
			t.Errorf("Expected updated_value, got %s", envVar.Value)
		}
		if !envVar.IsChanged {
			t.Error("Expected IsChanged to remain true")
		}
	})

	t.Run("set same value should not mark as changed", func(t *testing.T) {
		// Reset IsChanged flag first
		envFile.vars["NEW_VAR"].IsChanged = false

		envFile.Set("NEW_VAR", "updated_value") // same value

		envVar, exists := envFile.Get("NEW_VAR")
		if !exists {
			t.Error("Expected NEW_VAR to exist")
		}
		if envVar.IsChanged {
			t.Error("Expected IsChanged to remain false when setting same value")
		}
	})

	t.Run("get non-existent variable", func(t *testing.T) {
		envVar, exists := envFile.Get("NON_EXISTENT")
		if exists {
			t.Error("Expected NON_EXISTENT to not exist")
		}
		if envVar.Name != "NON_EXISTENT" || envVar.Line != -1 {
			t.Error("Expected empty EnvVar with Line=-1 for non-existent variable")
		}
	})

	t.Run("trim whitespace", func(t *testing.T) {
		envFile.Set("  TRIM_VAR  ", "  trim_value  ")

		envVar, exists := envFile.Get("TRIM_VAR")
		if !exists {
			t.Error("Expected TRIM_VAR to exist")
		}
		if envVar.Name != "TRIM_VAR" || envVar.Value != "trim_value" {
			t.Errorf("Expected TRIM_VAR=trim_value, got %s=%s", envVar.Name, envVar.Value)
		}
	})

	t.Run("delete variable", func(t *testing.T) {
		envFile.Set("DELETE_VAR", "delete_value")

		envVar, exists := envFile.Get("DELETE_VAR")
		if !exists {
			t.Error("Expected DELETE_VAR to exist before deletion")
		}
		if envVar.Value != "delete_value" {
			t.Errorf("Expected DELETE_VAR value 'delete_value', got '%s'", envVar.Value)
		}

		envFile.Del("DELETE_VAR")

		_, exists = envFile.Get("DELETE_VAR")
		if exists {
			t.Error("Expected DELETE_VAR to not exist after deletion")
		}
	})

	t.Run("delete non-existent variable", func(t *testing.T) {
		originalCount := len(envFile.GetAll())

		envFile.Del("NON_EXISTENT_VAR")

		if len(envFile.GetAll()) != originalCount {
			t.Error("Deleting non-existent variable should not change variable count")
		}
	})

	t.Run("get all variables", func(t *testing.T) {
		allVars := envFile.GetAll()

		if len(allVars) < 2 { // should have at least NEW_VAR and TRIM_VAR
			t.Errorf("Expected at least 2 variables, got %d", len(allVars))
		}

		if newVar, exists := allVars["NEW_VAR"]; !exists {
			t.Error("Expected NEW_VAR in GetAll result")
		} else if newVar.Value != "updated_value" {
			t.Errorf("Expected NEW_VAR value 'updated_value', got '%s'", newVar.Value)
		}
	})

	t.Run("set all variables", func(t *testing.T) {
		newVars := map[string]EnvVar{
			"BATCH_VAR1":   {Name: "BATCH_VAR1", Value: "batch_value1", IsChanged: true, Line: -1},
			"BATCH_VAR2":   {Name: "BATCH_VAR2", Value: "batch_value2", IsChanged: false, Line: 5},
			"EXISTING_VAR": {Name: "EXISTING_VAR", Value: "overwritten", IsChanged: true, Line: 10},
		}

		envFile.SetAll(newVars)

		// Check that all new variables were set
		for name, expected := range newVars {
			actual, exists := envFile.Get(name)
			if !exists {
				t.Errorf("Expected variable %s to exist after SetAll", name)
				continue
			}
			if actual.Value != expected.Value {
				t.Errorf("Variable %s: expected value %s, got %s", name, expected.Value, actual.Value)
			}
			if actual.IsChanged != expected.IsChanged {
				t.Errorf("Variable %s: expected IsChanged %v, got %v", name, expected.IsChanged, actual.IsChanged)
			}
			if actual.Line != expected.Line {
				t.Errorf("Variable %s: expected Line %d, got %d", name, expected.Line, actual.Line)
			}
		}

		// Check that previous variables still exist
		if _, exists := envFile.Get("NEW_VAR"); !exists {
			t.Error("Expected NEW_VAR to still exist after SetAll")
		}
	})

	t.Run("set all empty map", func(t *testing.T) {
		originalCount := len(envFile.GetAll())

		envFile.SetAll(map[string]EnvVar{})

		if len(envFile.GetAll()) != originalCount {
			t.Error("SetAll with empty map should not change existing variables")
		}
	})
}

func TestEnvFileSave(t *testing.T) {
	content := `VAR1=value1
VAR2=value2`

	tmpFile := createTempFile(t, content)
	defer os.Remove(tmpFile)

	envFile, err := LoadEnvFile(tmpFile)
	if err != nil {
		t.Fatalf("Failed to load env file: %v", err)
	}

	envFile.Set("VAR1", "new_value1")
	envFile.Set("NEW_VAR", "new_value")

	err = envFile.Save(tmpFile)
	if err != nil {
		t.Fatalf("Failed to save env file: %v", err)
	}

	// Check backup was created
	backupDir := filepath.Join(filepath.Dir(tmpFile), ".bak")
	entries, err := os.ReadDir(backupDir)
	if err != nil {
		t.Fatalf("Failed to read backup directory: %v", err)
	}
	if len(entries) == 0 {
		t.Error("Expected backup file to be created")
	}

	// Check file content
	savedContent, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatalf("Failed to read saved file: %v", err)
	}

	expectedLines := []string{"VAR1=new_value1", "VAR2=value2", "NEW_VAR=new_value"}
	savedLines := strings.Split(strings.TrimSpace(string(savedContent)), "\n")

	for _, expected := range expectedLines {
		found := false
		for _, line := range savedLines {
			if strings.TrimSpace(line) == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected line '%s' not found in saved file", expected)
		}
	}

	// Check IsChanged flags reset
	for _, envVar := range envFile.GetAll() {
		if envVar.IsChanged {
			t.Errorf("Expected IsChanged to be false after save for %s", envVar.Name)
		}
	}

	// Cleanup backup
	os.RemoveAll(backupDir)
}

func TestEnvFileSaveNewFile(t *testing.T) {
	envFile := &envFile{
		vars: map[string]*EnvVar{
			"VAR1": {Name: "VAR1", Value: "value1", IsChanged: true, Line: -1},
		},
		perm: 0644,
		raw:  "",
		mx:   &sync.Mutex{},
	}

	tmpDir := t.TempDir()
	defer os.RemoveAll(tmpDir)

	newFile := filepath.Join(tmpDir, "new.env")

	err := envFile.Save(newFile)
	if err != nil {
		t.Fatalf("Failed to save new file: %v", err)
	}

	content, err := os.ReadFile(newFile)
	if err != nil {
		t.Fatalf("Failed to read new file: %v", err)
	}

	if !strings.Contains(string(content), "VAR1=value1") {
		t.Error("Expected VAR1=value1 in new file")
	}
}

func TestEnvFileSaveErrors(t *testing.T) {
	const defaultEmptyContent = "# Empty file\n"

	t.Run("save to directory", func(t *testing.T) {
		envFile := &envFile{
			vars: map[string]*EnvVar{
				"VAR1": {Name: "VAR1", Value: "value1", IsChanged: true, Line: 0},
			},
			mx: &sync.Mutex{},
		}

		tmpDir := t.TempDir()
		defer os.RemoveAll(tmpDir)

		err := envFile.Save(tmpDir)
		if err == nil {
			t.Error("Expected error when saving to directory")
		}
	})

	t.Run("save empty file", func(t *testing.T) {
		envFile := &envFile{
			vars: make(map[string]*EnvVar),
			mx:   &sync.Mutex{},
		}

		tmpFile := createTempFile(t, defaultEmptyContent)
		defer os.Remove(tmpFile)

		err := envFile.Save(tmpFile)
		if err != nil {
			t.Fatalf("Failed to save empty env file: %v", err)
		}

		content, err := os.ReadFile(tmpFile)
		if err != nil {
			t.Fatalf("Failed to read empty env file: %v", err)
		}
		if string(content) != defaultEmptyContent {
			t.Errorf("Expected default empty content, got '%s'", string(content))
		}
	})

	t.Run("save without changes", func(t *testing.T) {
		envFile := &envFile{
			vars: map[string]*EnvVar{
				"VAR1": {Name: "VAR1", Value: "value1", IsChanged: false, Line: 0},
			},
			mx: &sync.Mutex{},
		}

		tmpFile := createTempFile(t, defaultEmptyContent)
		defer os.Remove(tmpFile)

		err := envFile.Save(tmpFile)
		if err != nil {
			t.Fatalf("Failed to save non changed env file: %v", err)
		}

		content, err := os.ReadFile(tmpFile)
		if err != nil {
			t.Fatalf("Failed to read empty env file: %v", err)
		}
		if string(content) != defaultEmptyContent {
			t.Errorf("Expected default empty content, got '%s'", string(content))
		}
	})
}
func TestEnvFileClone(t *testing.T) {
	original := &envFile{
		vars: map[string]*EnvVar{
			"VAR1": {Name: "VAR1", Value: "value1", IsChanged: true, Line: 0},
			"VAR2": {Name: "VAR2", Value: "value2", IsChanged: false, Line: 1},
		},
		perm: 0644,
		raw:  "VAR1=value1\nVAR2=value2",
		mx:   &sync.Mutex{},
	}

	clone := original.Clone()

	// Check independence
	if clone == original {
		t.Error("Clone should return different instance")
	}

	// Check content equality
	if len(clone.GetAll()) != len(original.GetAll()) {
		t.Error("Clone should have same number of variables")
	}

	for name, origVar := range original.GetAll() {
		cloneVar, exists := clone.Get(name)
		if !exists {
			t.Errorf("Variable %s missing in clone", name)
			continue
		}

		if cloneVar.Name != origVar.Name || cloneVar.Value != origVar.Value {
			t.Errorf("Variable %s content mismatch in clone", name)
		}
	}

	// Test modification independence
	clone.Set("VAR1", "modified")
	if original.vars["VAR1"].Value == "modified" {
		t.Error("Modifying clone should not affect original")
	}
}

func TestLoadVarsEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected map[string]string
	}{
		{
			name:     "empty lines",
			content:  "\n\n\nVAR1=value1\n\n",
			expected: map[string]string{"VAR1": "value1"},
		},
		{
			name:     "commented variables",
			content:  "#VAR1=commented\nVAR2=active",
			expected: map[string]string{"VAR1": "commented", "VAR2": "active"},
		},
		{
			name:     "comments in value",
			content:  "VAR1=value1 # comment\nVAR2=value2 # comment",
			expected: map[string]string{"VAR1": "value1", "VAR2": "value2"},
		},
		{
			name:     "variables with spaces",
			content:  "VAR1 = value1\n  VAR2=value2  ",
			expected: map[string]string{"VAR1": "value1", "VAR2": "value2"},
		},
		{
			name:     "variables with equals in value",
			content:  "VAR1=value=with=equals\nVAR2=url=https://example.com",
			expected: map[string]string{"VAR1": "value=with=equals", "VAR2": "url=https://example.com"},
		},
		{
			name:     "invalid lines ignored",
			content:  "invalid line\nVAR1=value1\nanother invalid",
			expected: map[string]string{"VAR1": "value1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vars := loadVars(tt.content)

			for expectedName, expectedValue := range tt.expected {
				envVar, exists := vars[expectedName]
				if !exists {
					t.Errorf("Expected variable %s not found", expectedName)
					continue
				}
				if envVar.Value != expectedValue {
					t.Errorf("Variable %s: expected value %s, got %s", expectedName, expectedValue, envVar.Value)
				}
			}
		})
	}
}

func TestPatchRaw(t *testing.T) {
	envFile := &envFile{
		vars: map[string]*EnvVar{
			"EXISTING": {Name: "EXISTING", Value: "updated", IsChanged: true, Line: 1},
			"NEW_VAR":  {Name: "NEW_VAR", Value: "new_value", IsChanged: true, Line: -1},
		},
		raw: "# Comment line\nEXISTING=old_value\nUNCHANGED=unchanged\n",
		mx:  &sync.Mutex{},
	}

	envFile.patchRaw()

	lines := strings.Split(envFile.raw, "\n")

	// Check existing variable updated
	if lines[1] != "EXISTING=updated" {
		t.Errorf("Expected line 1 to be 'EXISTING=updated', got '%s'", lines[1])
	}

	// Check new variable added
	found := false
	for _, line := range lines {
		if line == "NEW_VAR=new_value" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected NEW_VAR=new_value to be added to file")
	}

	// Check comment not modified
	if lines[0] != "# Comment line" {
		t.Errorf("Expected comment line unchanged, got '%s'", lines[0])
	}
	// Check last line is empty
	if lines[len(lines)-1] != "" {
		t.Errorf("Expected last line to be empty, got '%s'", lines[len(lines)-1])
	}
}

func TestEnvFileDelInSave(t *testing.T) {
	content := `VAR1=value1
VAR2=value2
VAR3=value3`

	tmpFile := createTempFile(t, content)
	defer os.Remove(tmpFile)

	envFile, err := LoadEnvFile(tmpFile)
	if err != nil {
		t.Fatalf("Failed to load env file: %v", err)
	}

	// Modify, add, and delete variables
	envFile.Set("VAR1", "new_value1")
	envFile.Set("NEW_VAR", "new_value")
	envFile.Del("VAR2")

	err = envFile.Save(tmpFile)
	if err != nil {
		t.Fatalf("Failed to save env file: %v", err)
	}

	// Check file content
	savedContent, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatalf("Failed to read saved file: %v", err)
	}

	contentStr := string(savedContent)

	// Should contain updated and new variables
	if !containsLine(contentStr, "VAR1=new_value1") {
		t.Error("Expected VAR1 to be updated in saved file")
	}
	if !containsLine(contentStr, "NEW_VAR=new_value") {
		t.Error("Expected NEW_VAR to be added to saved file")
	}
	if !containsLine(contentStr, "VAR3=value3") {
		t.Error("Expected VAR3 to remain unchanged in saved file")
	}

	// Should not contain deleted variable
	if containsLine(contentStr, "VAR2=value2") {
		t.Error("Expected VAR2 to be removed from saved file")
	}

	// Cleanup backup
	backupDir := filepath.Join(filepath.Dir(tmpFile), ".bak")
	os.RemoveAll(backupDir)
}

func TestSetDefaultVarsNilURL(t *testing.T) {
	envFile := &envFile{
		vars: make(map[string]*EnvVar),
		mx:   &sync.Mutex{},
	}

	// This should not panic even with nil URL
	err := setDefaultVars(envFile)
	if err != nil {
		t.Fatalf("setDefaultVars failed: %v", err)
	}

	// Check that STATIC_URL exists (it has envDefault empty, so should be nil URL)
	if envVar, exists := envFile.vars["STATIC_URL"]; exists {
		if envVar.Default != "" {
			t.Errorf("Expected empty default for STATIC_URL, got '%s'", envVar.Default)
		}
	}
}

func TestSetDefaultVars(t *testing.T) {
	envFile := &envFile{
		vars: make(map[string]*EnvVar),
		mx:   &sync.Mutex{},
	}

	// This should not panic even with nil URL
	err := setDefaultVars(envFile)
	if err != nil {
		t.Fatalf("setDefaultVars failed: %v", err)
	}

	// Check that all variables are not present and have default value the same as current value
	for name, envVar := range envFile.vars {
		if envVar.IsPresent() {
			t.Errorf("Expected variable %s to be not present", name)
		}
		if !envVar.IsDefault() {
			t.Errorf("Expected variable %s to have default value", name)
		}
	}
}

func createTempFile(t *testing.T, content string) string {
	tmpFile, err := os.CreateTemp("", "test*.env")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	if _, err := tmpFile.WriteString(content); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}

	if err := tmpFile.Close(); err != nil {
		t.Fatalf("Failed to close temp file: %v", err)
	}

	return tmpFile.Name()
}
