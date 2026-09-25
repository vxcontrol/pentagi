package loader

import (
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"sync"
	"testing"
)

func TestFile_EnvVar_ReportsDefaultAndPresence(t *testing.T) {
	tests := []struct {
		name        string
		envVar      EnvVar
		wantDefault bool
		wantPresent bool
	}{
		{name: "a file value other than the default", envVar: EnvVar{Value: "test_value", Default: "default_value", Line: 5}, wantPresent: true},
		{name: "a file value equal to the default", envVar: EnvVar{Value: "default_value", Default: "default_value", Line: 5}, wantDefault: true, wantPresent: true},
		{name: "an empty value not in the file falls back to the default", envVar: EnvVar{Default: "default_value", Line: -1}, wantDefault: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.envVar.IsDefault(); got != tt.wantDefault {
				t.Errorf("IsDefault() = %v, want %v", got, tt.wantDefault)
			}
			if got := tt.envVar.IsPresent(); got != tt.wantPresent {
				t.Errorf("IsPresent() = %v, want %v", got, tt.wantPresent)
			}
		})
	}
}

// TestFile_EnvFile_KeepsVariablesInMemory: subtests run in order over one file and are keyed by method.
func TestFile_EnvFile_KeepsVariablesInMemory(t *testing.T) {
	envFile := &envFile{vars: make(map[string]*EnvVar), mx: &sync.Mutex{}}

	t.Run("set adds a new variable as changed and absent from the file", func(t *testing.T) {
		envFile.Set("NEW_VAR", "new_value")

		envVar, exists := envFile.Get("NEW_VAR")
		if !exists || envVar.Name != "NEW_VAR" || envVar.Value != "new_value" || !envVar.IsChanged || envVar.Line != -1 {
			t.Errorf("Get(NEW_VAR) = %+v, %v; want a changed new_value on line -1", envVar, exists)
		}
	})

	t.Run("set updates an existing variable", func(t *testing.T) {
		envFile.Set("NEW_VAR", "updated_value")

		if envVar, exists := envFile.Get("NEW_VAR"); !exists || envVar.Value != "updated_value" || !envVar.IsChanged {
			t.Errorf("Get(NEW_VAR) = %+v, %v; want a changed updated_value", envVar, exists)
		}
	})

	t.Run("set with the same value does not mark it changed", func(t *testing.T) {
		envFile.vars["NEW_VAR"].IsChanged = false
		envFile.Set("NEW_VAR", "updated_value")

		if envVar, exists := envFile.Get("NEW_VAR"); !exists || envVar.IsChanged {
			t.Errorf("Get(NEW_VAR) = %+v, %v; want it unchanged", envVar, exists)
		}
	})

	t.Run("get reports an unknown variable as absent", func(t *testing.T) {
		if envVar, exists := envFile.Get("NON_EXISTENT"); exists || envVar.Name != "NON_EXISTENT" || envVar.Line != -1 {
			t.Errorf("Get(NON_EXISTENT) = %+v, %v; want an absent variable on line -1", envVar, exists)
		}
	})

	t.Run("set trims names and values", func(t *testing.T) {
		envFile.Set("  TRIM_VAR  ", "  trim_value  ")

		if envVar, exists := envFile.Get("TRIM_VAR"); !exists || envVar.Name != "TRIM_VAR" || envVar.Value != "trim_value" {
			t.Errorf("Get(TRIM_VAR) = %+v, %v; want TRIM_VAR=trim_value", envVar, exists)
		}
	})

	t.Run("del removes a variable", func(t *testing.T) {
		envFile.Set("DELETE_VAR", "delete_value")
		if envVar, exists := envFile.Get("DELETE_VAR"); !exists || envVar.Value != "delete_value" {
			t.Fatalf("Get(DELETE_VAR) before Del = %+v, %v", envVar, exists)
		}

		envFile.Del("DELETE_VAR")

		if _, exists := envFile.Get("DELETE_VAR"); exists {
			t.Error("DELETE_VAR still exists after Del")
		}
	})

	t.Run("del of an unknown variable changes nothing", func(t *testing.T) {
		envFile.Del("NON_EXISTENT_VAR")

		if got := len(envFile.GetAll()); got != 2 {
			t.Errorf("GetAll() holds %d variables after Del of an unknown one, want 2", got)
		}
	})

	t.Run("get all returns every variable", func(t *testing.T) {
		allVars := envFile.GetAll()

		if len(allVars) != 2 || allVars["NEW_VAR"].Value != "updated_value" || allVars["TRIM_VAR"].Value != "trim_value" {
			t.Errorf("GetAll() = %+v, want NEW_VAR=updated_value and TRIM_VAR=trim_value", allVars)
		}
	})

	t.Run("set all adds and overwrites variables as given", func(t *testing.T) {
		newVars := map[string]EnvVar{
			"BATCH_VAR1": {Name: "BATCH_VAR1", Value: "batch_value1", IsChanged: true, Line: -1},
			"BATCH_VAR2": {Name: "BATCH_VAR2", Value: "batch_value2", IsChanged: false, Line: 5},
			"NEW_VAR":    {Name: "NEW_VAR", Value: "overwritten", IsChanged: true, Line: 10},
		}

		envFile.SetAll(newVars)

		for name, expected := range newVars {
			if actual, exists := envFile.Get(name); !exists || actual != expected {
				t.Errorf("Get(%s) = %+v, %v; want %+v", name, actual, exists, expected)
			}
		}
		if _, exists := envFile.Get("TRIM_VAR"); !exists {
			t.Error("SetAll dropped TRIM_VAR, which it was not given")
		}
	})

	t.Run("set all with an empty map changes nothing", func(t *testing.T) {
		envFile.SetAll(map[string]EnvVar{})

		if got := len(envFile.GetAll()); got != 4 {
			t.Errorf("GetAll() holds %d variables after an empty SetAll, want 4", got)
		}
	})
}

func TestFile_Save_RewritesChangedLinesAndKeepsABackup(t *testing.T) {
	const original = "# Comment line\nVAR1=value1\nVAR2=value2\nVAR3=value3\n"
	path := loaderEnvFile(t, original)

	envFile, err := LoadEnvFile(path)
	if err != nil {
		t.Fatalf("LoadEnvFile: %v", err)
	}
	envFile.Set("VAR1", "new_value1")
	envFile.Set("NEW_VAR", "new_value")
	envFile.Del("VAR2")

	if err := envFile.Save(path); err != nil {
		t.Fatalf("Save: %v", err)
	}

	saved, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read saved file: %v", err)
	}
	if want := "# Comment line\nVAR1=new_value1\nVAR3=value3\nNEW_VAR=new_value\n"; string(saved) != want {
		t.Errorf("saved file = %q, want %q", saved, want)
	}

	backupDir := filepath.Join(filepath.Dir(path), ".bak")
	backups, err := os.ReadDir(backupDir)
	if err != nil || len(backups) != 1 || !strings.HasPrefix(backups[0].Name(), ".env.") {
		t.Fatalf("backups = %v, %v; want one .env.<timestamp>", backups, err)
	}
	if backup, err := os.ReadFile(filepath.Join(backupDir, backups[0].Name())); err != nil || string(backup) != original {
		t.Errorf("backup = %q, %v; want the original file", backup, err)
	}

	if info, err := os.Stat(path); err != nil {
		t.Errorf("stat saved file: %v", err)
	} else if runtime.GOOS != "windows" && info.Mode().Perm() != 0o600 {
		t.Errorf("saved file mode = %v, want the original -rw-------", info.Mode().Perm())
	}

	for _, envVar := range envFile.GetAll() {
		if envVar.IsChanged {
			t.Errorf("%s is still marked changed after Save", envVar.Name)
		}
	}
}

func TestFile_Save_WritesOnlyPendingChangesToAFile(t *testing.T) {
	tests := []struct {
		name    string
		vars    map[string]*EnvVar
		target  func(t *testing.T) string
		wantErr string
		want    string
	}{
		{
			name:   "a missing file is created",
			vars:   map[string]*EnvVar{"VAR1": {Name: "VAR1", Value: "value1", IsChanged: true, Line: -1}},
			target: func(t *testing.T) string { return filepath.Join(t.TempDir(), "new.env") },
			want:   "VAR1=value1\n",
		},
		{
			name:    "a directory is refused",
			vars:    map[string]*EnvVar{"VAR1": {Name: "VAR1", Value: "value1", IsChanged: true, Line: 0}},
			target:  func(t *testing.T) string { return t.TempDir() },
			wantErr: "is a directory",
		},
		{
			name:   "nothing changed leaves the file untouched",
			vars:   map[string]*EnvVar{"VAR1": {Name: "VAR1", Value: "value1", IsChanged: false, Line: 0}},
			target: func(t *testing.T) string { return loaderEnvFile(t, "# Empty file\n") },
			want:   "# Empty file\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			envFile := &envFile{vars: tt.vars, perm: 0o644, mx: &sync.Mutex{}}
			path := tt.target(t)

			err := envFile.Save(path)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("Save() error = %v, want it to mention %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("Save() error = %v", err)
			}
			if content, err := os.ReadFile(path); err != nil || string(content) != tt.want {
				t.Errorf("file = %q, %v; want %q", content, err, tt.want)
			}
		})
	}
}

func TestFile_Clone_IsIndependentOfTheOriginal(t *testing.T) {
	original := &envFile{
		vars: map[string]*EnvVar{
			"VAR1": {Name: "VAR1", Value: "value1", IsChanged: true, Line: 0},
			"VAR2": {Name: "VAR2", Value: "value2", IsChanged: false, Line: 1},
		},
		perm: 0o644,
		raw:  "VAR1=value1\nVAR2=value2",
		mx:   &sync.Mutex{},
	}

	clone := original.Clone()

	if !reflect.DeepEqual(clone.GetAll(), original.GetAll()) {
		t.Errorf("clone variables = %+v, want %+v", clone.GetAll(), original.GetAll())
	}
	if c := clone.(*envFile); c.raw != original.raw || c.perm != original.perm {
		t.Errorf("clone raw, perm = %q, %v; want the original's, or its Save loses the file", c.raw, c.perm)
	}

	clone.Set("VAR1", "modified")
	if original.vars["VAR1"].Value != "value1" {
		t.Error("modifying the clone changed the original")
	}
}
