package processor

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"pentagi/cmd/installer/state"
)

// TestRestartWithMissingGraphitiFile tests that restart operation handles
// missing graphiti compose file gracefully when GRAPHITI_ENABLED=false
func TestRestartWithMissingGraphitiFile(t *testing.T) {
	tmpDir := t.TempDir()
	envPath := filepath.Join(tmpDir, ".env")

	// Create .env file with GRAPHITI_ENABLED=false
	envContent := `GRAPHITI_ENABLED=false
GRAPHITI_URL=http://localhost:8000`
	err := os.WriteFile(envPath, []byte(envContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test env file: %v", err)
	}

	// Create main docker-compose.yml (pentagi stack)
	composeContent := `version: "3.8"
services:
  pentagi:
    image: pentagi:latest`
	composePath := filepath.Join(tmpDir, "docker-compose.yml")
	err = os.WriteFile(composePath, []byte(composeContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test compose file: %v", err)
	}

	// NOTE: We intentionally DO NOT create docker-compose-graphiti.yml
	// to simulate the bug scenario

	// Initialize state
	appState, err := state.NewState(envPath)
	if err != nil {
		t.Fatalf("Failed to initialize state: %v", err)
	}

	// Create processor (this would normally be done through proper initialization)
	// For this test, we'll just verify the helper function works
	proc := &processor{
		state: appState,
	}

	composeOps := newComposeOperations(proc).(*composeOperationsImpl)

	// Test 1: composeFileExists should return false for missing graphiti file
	if composeOps.composeFileExists(ProductStackGraphiti) {
		t.Error("composeFileExists should return false for missing graphiti file")
	}

	// Test 2: isComposeMissingError should detect missing file errors
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "file does not exist error",
			err:      os.ErrNotExist,
			expected: true,
		},
		{
			name:     "custom does not exist error",
			err:      &os.PathError{Op: "open", Path: "/path/to/file", Err: os.ErrNotExist},
			expected: true,
		},
		{
			name:     "other error",
			err:      os.ErrPermission,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := composeOps.isComposeMissingError(tt.err)
			if result != tt.expected {
				t.Errorf("isComposeMissingError(%v) = %v, expected %v",
					tt.err, result, tt.expected)
			}
		})
	}
}

// TestPerformStackOperationSkipsMissingFiles tests that stop operations
// gracefully skip stacks with missing compose files
func TestPerformStackOperationSkipsMissingFiles(t *testing.T) {
	tmpDir := t.TempDir()
	envPath := filepath.Join(tmpDir, ".env")

	// Create .env file
	envContent := `GRAPHITI_ENABLED=false`
	err := os.WriteFile(envPath, []byte(envContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test env file: %v", err)
	}

	// Initialize state
	appState, err := state.NewState(envPath)
	if err != nil {
		t.Fatalf("Failed to initialize state: %v", err)
	}

	proc := &processor{
		state: appState,
	}

	composeOps := newComposeOperations(proc).(*composeOperationsImpl)

	// Test that performStackOperation returns nil (skips) for missing file
	// This simulates the restart operation flow
	ctx := context.Background()
	opState := &operationState{}

	// This should not return an error because the file doesn't exist
	err = composeOps.performStackOperation(ctx, ProductStackGraphiti, opState, ProcessorOperationStop, "stop")
	if err != nil {
		t.Errorf("performStackOperation should skip missing graphiti file, but got error: %v", err)
	}
}
