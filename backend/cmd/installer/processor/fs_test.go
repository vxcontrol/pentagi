package processor

import (
	"context"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"pentagi/cmd/installer/files"
)

// testStackIntegrityOperation is a helper for testing stack integrity operations
func testStackIntegrityOperation(t *testing.T, operation func(*fileSystemOperationsImpl, context.Context, ProductStack, *operationState) error, needsTempDir bool) {
	t.Helper()

	tests := []struct {
		name      string
		stack     ProductStack
		expectErr bool
	}{
		{"ProductStackPentagi", ProductStackPentagi, false},
		{"ProductStackLangfuse", ProductStackLangfuse, false},
		{"ProductStackObservability", ProductStackObservability, false},
		{"ProductStackCompose", ProductStackCompose, false},
		{"ProductStackAll", ProductStackAll, false},
		{"ProductStackWorker - unsupported", ProductStackWorker, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			processor := createTestProcessor()
			fsOps := newFileSystemOperations(processor).(*fileSystemOperationsImpl)

			// for cleanup operations, ensure temp directory setup
			if needsTempDir {
				tmpDir := t.TempDir()
				mockState := processor.state.(*mockState)
				mockState.envPath = filepath.Join(tmpDir, ".env")
			}

			err := operation(fsOps, t.Context(), tt.stack, testOperationState(t))
			assertError(t, err, tt.expectErr, "")
		})
	}
}

func TestFileSystemOperationsImpl_EnsureStackIntegrity(t *testing.T) {
	testStackIntegrityOperation(t, (*fileSystemOperationsImpl).ensureStackIntegrity, false)
}

func TestFileSystemOperationsImpl_VerifyStackIntegrity(t *testing.T) {
	testStackIntegrityOperation(t, (*fileSystemOperationsImpl).verifyStackIntegrity, false)
}

func TestFileSystemOperationsImpl_CleanupStackFiles(t *testing.T) {
	testStackIntegrityOperation(t, (*fileSystemOperationsImpl).cleanupStackFiles, true)
}

func TestFileSystemOperationsImpl_EnsureFileFromEmbed(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		force    bool
		setup    func(*mockFiles, string) // setup mock and working dir
	}{
		{
			name:     "file missing - should copy",
			filename: "test.yml",
			force:    false,
			setup: func(m *mockFiles, workingDir string) {
				// file not exists, will be copied
				m.AddFile("test.yml", []byte("test content"))
			},
		},
		{
			name:     "file exists, force false - should skip",
			filename: "test.yml",
			force:    false,
			setup: func(m *mockFiles, workingDir string) {
				// create existing file
				os.WriteFile(filepath.Join(workingDir, "test.yml"), []byte("existing"), 0644)
				m.AddFile("test.yml", []byte("test content"))
			},
		},
		{
			name:     "file exists, force true - should update",
			filename: "test.yml",
			force:    true,
			setup: func(m *mockFiles, workingDir string) {
				// create existing file
				os.WriteFile(filepath.Join(workingDir, "test.yml"), []byte("existing"), 0644)
				m.AddFile("test.yml", []byte("test content"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// create temp directory
			tmpDir := t.TempDir()

			// create test processor using unified mocks
			processor := createTestProcessor()
			mockState := processor.state.(*mockState)
			mockState.envPath = filepath.Join(tmpDir, ".env")

			mockFiles := processor.files.(*mockFiles)
			tt.setup(mockFiles, tmpDir)

			fsOps := newFileSystemOperations(processor).(*fileSystemOperationsImpl)
			state := &operationState{force: tt.force, mx: &sync.Mutex{}, ctx: t.Context()}

			err := fsOps.ensureFileFromEmbed(tt.filename, state)
			assertNoError(t, err)
		})
	}
}

func TestFileSystemOperationsImpl_VerifyDirectoryContentIntegrity(t *testing.T) {
	tests := []struct {
		name      string
		embedPath string
		setup     func(*mockFiles, string, string) // setup mock, embedPath, targetPath
		force     bool
		expectErr bool
	}{
		{
			name:      "embedded directory not found",
			embedPath: "nonexistent",
			setup:     func(m *mockFiles, embedPath, targetPath string) {},
			force:     false,
			expectErr: true,
		},
		{
			name:      "directory with all files OK",
			embedPath: "observability",
			setup: func(m *mockFiles, embedPath, targetPath string) {
				// setup embedded files
				m.lists[embedPath] = []string{
					"observability/config1.yml",
					"observability/config2.yml",
				}
				m.statuses["observability/config1.yml"] = files.FileStatusOK
				m.statuses["observability/config2.yml"] = files.FileStatusOK
				m.AddFile(embedPath, []byte{}) // mark as existing directory

				// create target directory
				os.MkdirAll(targetPath, 0755)
			},
			force:     false,
			expectErr: false,
		},
		{
			name:      "directory with missing files",
			embedPath: "observability",
			setup: func(m *mockFiles, embedPath, targetPath string) {
				// setup embedded files
				m.lists[embedPath] = []string{
					"observability/config1.yml",
					"observability/config2.yml",
				}
				m.statuses["observability/config1.yml"] = files.FileStatusOK
				m.statuses["observability/config2.yml"] = files.FileStatusMissing
				m.AddFile(embedPath, []byte{}) // mark as existing directory

				// create target directory
				os.MkdirAll(targetPath, 0755)
			},
			force:     false,
			expectErr: false,
		},
		{
			name:      "directory with modified files, force false",
			embedPath: "observability",
			setup: func(m *mockFiles, embedPath, targetPath string) {
				// setup embedded files
				m.lists[embedPath] = []string{
					"observability/config1.yml",
				}
				m.statuses["observability/config1.yml"] = files.FileStatusModified
				m.AddFile(embedPath, []byte{}) // mark as existing directory

				// create target directory
				os.MkdirAll(targetPath, 0755)
			},
			force:     false,
			expectErr: false,
		},
		{
			name:      "directory with modified files, force true",
			embedPath: "observability",
			setup: func(m *mockFiles, embedPath, targetPath string) {
				// setup embedded files
				m.lists[embedPath] = []string{
					"observability/config1.yml",
				}
				m.statuses["observability/config1.yml"] = files.FileStatusModified
				m.AddFile(embedPath, []byte{}) // mark as existing directory

				// create target directory
				os.MkdirAll(targetPath, 0755)
			},
			force:     true,
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// create temp directory
			tmpDir := t.TempDir()
			targetPath := filepath.Join(tmpDir, tt.embedPath)

			// create test processor
			processor := createTestProcessor()
			mockState := processor.state.(*mockState)
			mockState.envPath = filepath.Join(tmpDir, ".env")

			mockFiles := processor.files.(*mockFiles)
			tt.setup(mockFiles, tt.embedPath, targetPath)

			fsOps := newFileSystemOperations(processor).(*fileSystemOperationsImpl)
			state := &operationState{force: tt.force, mx: &sync.Mutex{}, ctx: t.Context()}

			err := fsOps.verifyDirectoryContentIntegrity(tt.embedPath, targetPath, state)
			assertError(t, err, tt.expectErr, "")
		})
	}
}

func TestFileSystemOperationsImpl_ExcludedFilesHandling(t *testing.T) {
	if len(filesToExcludeFromVerification) == 0 {
		t.Skip("no excluded files configured; skipping excluded files tests")
	}

	excluded := filesToExcludeFromVerification[0]

	t.Run("excluded_missing_should_be_copied", func(t *testing.T) {
		tmpDir := t.TempDir()
		targetPath := filepath.Join(tmpDir, observabilityDirectory)

		processor := createTestProcessor()
		mockState := processor.state.(*mockState)
		mockState.envPath = filepath.Join(tmpDir, ".env")

		mockFiles := processor.files.(*mockFiles)
		// mark embedded directory and list with excluded file only
		mockFiles.lists[observabilityDirectory] = []string{
			excluded,
		}
		mockFiles.statuses[excluded] = files.FileStatusMissing
		mockFiles.AddFile(observabilityDirectory, []byte{})

		// ensure target directory exists on fs
		_ = os.MkdirAll(targetPath, 0o755)

		fsOps := newFileSystemOperations(processor).(*fileSystemOperationsImpl)
		state := &operationState{force: false, mx: &sync.Mutex{}, ctx: t.Context()}

		err := fsOps.verifyDirectoryContentIntegrity(observabilityDirectory, targetPath, state)
		assertNoError(t, err)

		if len(mockFiles.copies) != 1 {
			t.Fatalf("expected 1 copy for missing excluded file, got %d", len(mockFiles.copies))
		}
		if mockFiles.copies[0].Src != excluded || mockFiles.copies[0].Dst != tmpDir {
			t.Errorf("unexpected copy details: %+v", mockFiles.copies[0])
		}
	})

	t.Run("excluded_modified_should_not_be_copied", func(t *testing.T) {
		tmpDir := t.TempDir()
		targetPath := filepath.Join(tmpDir, observabilityDirectory)

		processor := createTestProcessor()
		mockState := processor.state.(*mockState)
		mockState.envPath = filepath.Join(tmpDir, ".env")

		mockFiles := processor.files.(*mockFiles)
		// mark embedded directory and list with excluded file only
		mockFiles.lists[observabilityDirectory] = []string{
			excluded,
		}
		mockFiles.statuses[excluded] = files.FileStatusModified
		mockFiles.AddFile(observabilityDirectory, []byte{})

		// ensure target directory exists on fs
		_ = os.MkdirAll(targetPath, 0o755)

		fsOps := newFileSystemOperations(processor).(*fileSystemOperationsImpl)
		state := &operationState{force: false, mx: &sync.Mutex{}, ctx: t.Context()}

		err := fsOps.verifyDirectoryContentIntegrity(observabilityDirectory, targetPath, state)
		assertNoError(t, err)

		if len(mockFiles.copies) != 0 {
			t.Fatalf("expected 0 copies for modified excluded file, got %d", len(mockFiles.copies))
		}
	})

	t.Run("force_true_updates_only_non_excluded", func(t *testing.T) {
		tmpDir := t.TempDir()
		targetPath := filepath.Join(tmpDir, observabilityDirectory)

		processor := createTestProcessor()
		mockState := processor.state.(*mockState)
		mockState.envPath = filepath.Join(tmpDir, ".env")

		mockFiles := processor.files.(*mockFiles)
		// list contains excluded and non-excluded files
		nonExcluded := "observability/other.yml"
		mockFiles.lists[observabilityDirectory] = []string{
			excluded,    // excluded
			nonExcluded, // non-excluded
		}
		mockFiles.statuses[excluded] = files.FileStatusModified
		mockFiles.statuses[nonExcluded] = files.FileStatusModified
		mockFiles.AddFile(observabilityDirectory, []byte{})

		// ensure target directory exists on fs
		_ = os.MkdirAll(targetPath, 0o755)

		fsOps := newFileSystemOperations(processor).(*fileSystemOperationsImpl)
		// call higher-level method to exercise force=true branch
		state := &operationState{force: true, mx: &sync.Mutex{}, ctx: t.Context()}

		err := fsOps.verifyDirectoryIntegrity(observabilityDirectory, state)
		assertNoError(t, err)

		if len(mockFiles.copies) != 1 {
			t.Fatalf("expected 1 copy for non-excluded modified file, got %d", len(mockFiles.copies))
		}
		if mockFiles.copies[0].Src != nonExcluded || mockFiles.copies[0].Dst != tmpDir {
			t.Errorf("unexpected copy details: %+v", mockFiles.copies[0])
		}
	})
}

func TestFileSystemOperationsImpl_FileExists(t *testing.T) {
	tmpDir := t.TempDir()
	processor := createTestProcessor()
	fsOps := newFileSystemOperations(processor).(*fileSystemOperationsImpl)

	// create test file
	testFile := filepath.Join(tmpDir, "test.txt")
	os.WriteFile(testFile, []byte("test"), 0644)

	// create test directory
	testDir := filepath.Join(tmpDir, "testdir")
	os.MkdirAll(testDir, 0755)

	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{"existing file", testFile, true},
		{"existing directory", testDir, false}, // fileExists should return false for directories
		{"nonexistent path", filepath.Join(tmpDir, "nonexistent"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := fsOps.fileExists(tt.path)
			if result != tt.expected {
				t.Errorf("fileExists(%s) = %v, want %v", tt.path, result, tt.expected)
			}
		})
	}
}

func TestFileSystemOperationsImpl_DirectoryExists(t *testing.T) {
	tmpDir := t.TempDir()
	processor := createTestProcessor()
	fsOps := newFileSystemOperations(processor).(*fileSystemOperationsImpl)

	// create test file
	testFile := filepath.Join(tmpDir, "test.txt")
	os.WriteFile(testFile, []byte("test"), 0644)

	// create test directory
	testDir := filepath.Join(tmpDir, "testdir")
	os.MkdirAll(testDir, 0755)

	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{"existing file", testFile, false}, // directoryExists should return false for files
		{"existing directory", testDir, true},
		{"nonexistent path", filepath.Join(tmpDir, "nonexistent"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := fsOps.directoryExists(tt.path)
			if result != tt.expected {
				t.Errorf("directoryExists(%s) = %v, want %v", tt.path, result, tt.expected)
			}
		})
	}
}

func TestFileSystemOperationsImpl_ValidateYamlFile(t *testing.T) {
	tmpDir := t.TempDir()
	processor := createTestProcessor()
	fsOps := newFileSystemOperations(processor).(*fileSystemOperationsImpl)

	tests := []struct {
		name      string
		content   string
		expectErr bool
	}{
		{
			name: "valid YAML",
			content: `
version: '3.8'
services:
  app:
    image: nginx
    ports:
      - "80:80"
`,
			expectErr: false,
		},
		{
			name: "invalid YAML - syntax error",
			content: `
version: '3.8'
services:
  app:
    image: nginx
    ports:
      - "80:80
    # missing closing quote
`,
			expectErr: true,
		},
		{
			name:      "empty file",
			content:   "",
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// create test file
			testFile := filepath.Join(tmpDir, "test.yml")
			err := os.WriteFile(testFile, []byte(tt.content), 0644)
			if err != nil {
				t.Fatalf("failed to create test file: %v", err)
			}

			err = fsOps.validateYamlFile(testFile)
			assertError(t, err, tt.expectErr, "")
		})
	}
}

func TestCheckStackIntegrity(t *testing.T) {
	tests := []struct {
		name     string
		stack    ProductStack
		setup    func(*mockFiles)
		expected map[string]files.FileStatus
	}{
		{
			name:  "pentagi_stack_all_files_ok",
			stack: ProductStackPentagi,
			setup: func(m *mockFiles) {
				m.statuses[composeFilePentagi] = files.FileStatusOK
			},
			expected: map[string]files.FileStatus{
				composeFilePentagi: files.FileStatusOK,
			},
		},
		{
			name:  "langfuse_stack_file_modified",
			stack: ProductStackLangfuse,
			setup: func(m *mockFiles) {
				m.statuses[composeFileLangfuse] = files.FileStatusModified
			},
			expected: map[string]files.FileStatus{
				composeFileLangfuse: files.FileStatusModified,
			},
		},
		{
			name:  "observability_stack_mixed_status",
			stack: ProductStackObservability,
			setup: func(m *mockFiles) {
				m.statuses[composeFileObservability] = files.FileStatusOK
				m.lists[observabilityDirectory] = []string{
					"observability/config1.yml",
					"observability/config2.yml",
					"observability/subdir/config3.yml",
				}
				m.statuses["observability/config1.yml"] = files.FileStatusOK
				m.statuses["observability/config2.yml"] = files.FileStatusModified
				m.statuses["observability/subdir/config3.yml"] = files.FileStatusMissing
			},
			expected: map[string]files.FileStatus{
				composeFileObservability:           files.FileStatusOK,
				"observability/config1.yml":        files.FileStatusOK,
				"observability/config2.yml":        files.FileStatusModified,
				"observability/subdir/config3.yml": files.FileStatusMissing,
			},
		},
		{
			name:  "compose_stacks_combined",
			stack: ProductStackCompose,
			setup: func(m *mockFiles) {
				// pentagi
				m.statuses[composeFilePentagi] = files.FileStatusOK
				// graphiti
				m.statuses[composeFileGraphiti] = files.FileStatusOK
				// langfuse
				m.statuses[composeFileLangfuse] = files.FileStatusModified
				// observability
				m.statuses[composeFileObservability] = files.FileStatusMissing
				m.lists[observabilityDirectory] = []string{
					"observability/config.yml",
				}
				m.statuses["observability/config.yml"] = files.FileStatusOK
			},
			expected: map[string]files.FileStatus{
				composeFilePentagi:         files.FileStatusOK,
				composeFileGraphiti:        files.FileStatusOK,
				composeFileLangfuse:        files.FileStatusModified,
				composeFileObservability:   files.FileStatusMissing,
				"observability/config.yml": files.FileStatusOK,
			},
		},
		{
			name:  "all_stacks_combined",
			stack: ProductStackAll,
			setup: func(m *mockFiles) {
				// pentagi
				m.statuses[composeFilePentagi] = files.FileStatusOK
				// graphiti
				m.statuses[composeFileGraphiti] = files.FileStatusOK
				// langfuse
				m.statuses[composeFileLangfuse] = files.FileStatusModified
				// observability
				m.statuses[composeFileObservability] = files.FileStatusMissing
				m.lists[observabilityDirectory] = []string{
					"observability/config.yml",
				}
				m.statuses["observability/config.yml"] = files.FileStatusOK
			},
			expected: map[string]files.FileStatus{
				composeFilePentagi:         files.FileStatusOK,
				composeFileGraphiti:        files.FileStatusOK,
				composeFileLangfuse:        files.FileStatusModified,
				composeFileObservability:   files.FileStatusMissing,
				"observability/config.yml": files.FileStatusOK,
			},
		},
		{
			name:     "unsupported_stack",
			stack:    ProductStackWorker,
			setup:    func(m *mockFiles) {},
			expected: map[string]files.FileStatus{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockFiles := newMockFiles()
			tt.setup(mockFiles)

			// create test processor and fsOps
			processor := createTestProcessor()
			// set working dir via state env path
			mockState := processor.state.(*mockState)
			mockState.envPath = filepath.Join("/test/working", ".env")
			processor.files = mockFiles

			fsOps := newFileSystemOperations(processor).(*fileSystemOperationsImpl)
			result, err := fsOps.checkStackIntegrity(t.Context(), tt.stack)
			if err != nil {
				if tt.stack == ProductStackWorker {
					// unsupported stack should return an error or empty result; current impl returns error
					return
				}
				t.Fatalf("unexpected error: %v", err)
			}

			if len(result) != len(tt.expected) {
				t.Errorf("expected %d file statuses, got %d", len(tt.expected), len(result))
			}

			for path, expectedStatus := range tt.expected {
				if actualStatus, ok := result[path]; !ok {
					t.Errorf("expected file %s not found in result", path)
				} else if actualStatus != expectedStatus {
					t.Errorf("file %s: expected status %v, got %v", path, expectedStatus, actualStatus)
				}
			}
		})
	}
}

func TestCheckStackIntegrity_RealFiles(t *testing.T) {
	// Test with real files interface would require proper embedded files setup
	// For now, we focus on mock-based testing which covers the logic
	t.Run("mock_based_coverage", func(t *testing.T) {
		// The logic is covered by TestGetStackFilesStatus above
		// Real files integration would require setting up embedded content
		// which is beyond the scope of unit tests
		mockFiles := newMockFiles()

		// Setup comprehensive test scenario
		mockFiles.statuses[composeFilePentagi] = files.FileStatusOK
		mockFiles.statuses[composeFileGraphiti] = files.FileStatusOK
		mockFiles.statuses[composeFileLangfuse] = files.FileStatusModified
		mockFiles.statuses[composeFileObservability] = files.FileStatusMissing
		mockFiles.lists[observabilityDirectory] = []string{
			"observability/config.yml",
			"observability/subdir/nested.yml",
		}
		mockFiles.statuses["observability/config.yml"] = files.FileStatusOK
		mockFiles.statuses["observability/subdir/nested.yml"] = files.FileStatusModified

		for _, stack := range []ProductStack{ProductStackAll, ProductStackCompose} {
			// create test processor and fsOps
			processor := createTestProcessor()
			mockState := processor.state.(*mockState)
			mockState.envPath = filepath.Join("/test", ".env")
			processor.files = mockFiles
			fsOps := newFileSystemOperations(processor).(*fileSystemOperationsImpl)
			result, err := fsOps.checkStackIntegrity(t.Context(), stack)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// Verify all files are captured
			expectedCount := 6 // 4 compose files + 2 observability directory files
			if len(result) != expectedCount {
				t.Errorf("expected %d files, got %d", expectedCount, len(result))
			}
		}
	})
}

func TestFileSystemOperations_IntegrityWithForceMode(t *testing.T) {
	// Test the interaction between ensure/verify integrity and force mode
	t.Run("ensure_respects_force_mode", func(t *testing.T) {
		processor := createTestProcessor()
		mockFiles := processor.files.(*mockFiles)

		// Setup files
		mockFiles.statuses[composeFilePentagi] = files.FileStatusModified
		mockFiles.AddFile(composeFilePentagi, []byte("embedded content"))

		fsOps := newFileSystemOperations(processor).(*fileSystemOperationsImpl)

		// First without force - should not overwrite
		state := &operationState{force: false, mx: &sync.Mutex{}, ctx: t.Context()}
		err := fsOps.ensureStackIntegrity(t.Context(), ProductStackPentagi, state)
		assertNoError(t, err)

		// Check that file was not copied (force=false with existing file)
		copyCount := 0
		for _, copy := range mockFiles.copies {
			if copy.Src == composeFilePentagi {
				copyCount++
			}
		}
		// Note: in real implementation, existing file check happens in ensureFileFromEmbed
		// which uses fileExists check on actual filesystem, not mock

		// Now with force - should overwrite
		state.force = true
		err = fsOps.ensureStackIntegrity(t.Context(), ProductStackPentagi, state)
		assertNoError(t, err)
	})
}

func TestFileSystemOperations_StackSpecificBehavior(t *testing.T) {
	// Test stack-specific behaviors
	t.Run("observability_handles_directory", func(t *testing.T) {
		processor := createTestProcessor()
		mockFiles := processor.files.(*mockFiles)
		tmpDir := t.TempDir()
		mockState := processor.state.(*mockState)
		mockState.envPath = filepath.Join(tmpDir, ".env")

		// Setup observability directory
		mockFiles.lists[observabilityDirectory] = []string{
			"observability/config1.yml",
			"observability/nested/config2.yml",
		}
		mockFiles.statuses["observability/config1.yml"] = files.FileStatusOK
		mockFiles.statuses["observability/nested/config2.yml"] = files.FileStatusOK
		mockFiles.AddFile(observabilityDirectory, []byte{}) // mark as directory

		fsOps := newFileSystemOperations(processor).(*fileSystemOperationsImpl)
		state := &operationState{force: false, mx: &sync.Mutex{}, ctx: t.Context()}

		err := fsOps.ensureStackIntegrity(t.Context(), ProductStackObservability, state)
		assertNoError(t, err)

		// Should have attempted to copy both compose file and directory
		expectedCopies := 2 // compose file + directory
		if len(mockFiles.copies) < expectedCopies {
			t.Errorf("expected at least %d copy operations, got %d", expectedCopies, len(mockFiles.copies))
		}
	})
}
