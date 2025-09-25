package files

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// newTestFiles creates a Files instance for testing with a custom links directory
func newTestFiles(linksDir string) Files {
	return &files{
		linksDir: linksDir,
	}
}

func TestNewFiles(t *testing.T) {
	f := NewFiles()
	if f == nil {
		t.Fatal("NewFiles() returned nil")
	}
}

func TestGetContent_FromFS(t *testing.T) {
	// Create temporary test directory structure
	tmpDir := t.TempDir()
	defer os.RemoveAll(tmpDir)

	testLinksDir := filepath.Join(tmpDir, "links")

	setupTestLinksInDir(t, testLinksDir)

	f := newTestFiles(testLinksDir)
	content, err := f.GetContent("test.txt")
	if err != nil {
		t.Fatalf("GetContent() error = %v", err)
	}

	expected := "test content"
	if string(content) != expected {
		t.Errorf("GetContent() = %q, want %q", string(content), expected)
	}
}

func TestExistsInFS(t *testing.T) {
	tmpDir := t.TempDir()
	defer os.RemoveAll(tmpDir)

	testLinksDir := filepath.Join(tmpDir, "links")

	setupTestLinksInDir(t, testLinksDir)

	f := newTestFiles(testLinksDir)

	if !f.ExistsInFS("test.txt") {
		t.Error("ExistsInFS() = false, want true for existing file")
	}

	if f.ExistsInFS("nonexistent.txt") {
		t.Error("ExistsInFS() = true, want false for non-existent file")
	}
}

func TestStat_FromFS(t *testing.T) {
	tmpDir := t.TempDir()
	defer os.RemoveAll(tmpDir)

	testLinksDir := filepath.Join(tmpDir, "links")

	setupTestLinksInDir(t, testLinksDir)

	f := newTestFiles(testLinksDir)
	info, err := f.Stat("test.txt")
	if err != nil {
		t.Fatalf("Stat() error = %v", err)
	}

	if info.IsDir() {
		t.Error("Stat() IsDir() = true, want false for file")
	}

	if info.Size() == 0 {
		t.Error("Stat() Size() = 0, want > 0")
	}
}

func TestCopy_File(t *testing.T) {
	tmpDir := t.TempDir()
	defer os.RemoveAll(tmpDir)

	testLinksDir := filepath.Join(tmpDir, "links")

	setupTestLinksInDir(t, testLinksDir)

	copyDstDir := t.TempDir()
	defer os.RemoveAll(copyDstDir)

	f := newTestFiles(testLinksDir)

	err := f.Copy("test.txt", copyDstDir, false)
	if err != nil {
		t.Fatalf("Copy() error = %v", err)
	}

	// Verify file was copied
	copiedPath := filepath.Join(copyDstDir, "test.txt")
	content, err := os.ReadFile(copiedPath)
	if err != nil {
		t.Fatalf("Failed to read copied file: %v", err)
	}

	expected := "test content"
	if string(content) != expected {
		t.Errorf("Copied file content = %q, want %q", string(content), expected)
	}
}

func TestCopy_PreservesExecutable_FromFS(t *testing.T) {
	tmpDir := t.TempDir()
	defer os.RemoveAll(tmpDir)

	testLinksDir := filepath.Join(tmpDir, "links")
	if err := os.MkdirAll(testLinksDir, 0755); err != nil {
		t.Fatalf("failed to create links dir: %v", err)
	}

	// create executable source file
	src := filepath.Join(testLinksDir, "run.sh")
	if err := os.WriteFile(src, []byte("#!/bin/sh\necho hi\n"), 0755); err != nil {
		t.Fatalf("failed to create exec file: %v", err)
	}

	copyDstDir := t.TempDir()
	defer os.RemoveAll(copyDstDir)

	f := newTestFiles(testLinksDir)

	if err := f.Copy("run.sh", copyDstDir, false); err != nil {
		t.Fatalf("Copy() error = %v", err)
	}

	// verify executable bit preserved
	copied := filepath.Join(copyDstDir, "run.sh")
	info, err := os.Stat(copied)
	if err != nil {
		t.Fatalf("failed to stat copied: %v", err)
	}
	if info.Mode().Perm() != 0755 {
		t.Errorf("copied mode = %o, want %o", info.Mode().Perm(), os.FileMode(0755))
	}
}

func TestCheck_DetectsPermissionMismatch_FromFS(t *testing.T) {
	tmpDir := t.TempDir()
	defer os.RemoveAll(tmpDir)

	testLinksDir := filepath.Join(tmpDir, "links")
	if err := os.MkdirAll(testLinksDir, 0755); err != nil {
		t.Fatalf("failed to create links dir: %v", err)
	}

	// create executable source file
	src := filepath.Join(testLinksDir, "tool.sh")
	if err := os.WriteFile(src, []byte("#!/bin/sh\necho tool\n"), 0755); err != nil {
		t.Fatalf("failed to create exec file: %v", err)
	}

	f := newTestFiles(testLinksDir)

	workingDir := t.TempDir()
	defer os.RemoveAll(workingDir)

	if err := f.Copy("tool.sh", workingDir, false); err != nil {
		t.Fatalf("Copy() error = %v", err)
	}

	// change permissions to remove executable bit
	target := filepath.Join(workingDir, "tool.sh")
	if err := os.Chmod(target, 0644); err != nil {
		t.Fatalf("failed to chmod: %v", err)
	}

	status := f.Check("tool.sh", workingDir)
	if status != FileStatusModified {
		t.Errorf("Check() perms mismatch = %v, want %v", status, FileStatusModified)
	}
}

func TestCopy_Directory(t *testing.T) {
	tmpDir := t.TempDir()
	defer os.RemoveAll(tmpDir)

	testLinksDir := filepath.Join(tmpDir, "links")

	setupTestLinksWithDirInDir(t, testLinksDir)

	copyDstDir := t.TempDir()
	defer os.RemoveAll(copyDstDir)

	f := newTestFiles(testLinksDir)

	err := f.Copy("testdir", copyDstDir, false)
	if err != nil {
		t.Fatalf("Copy() error = %v", err)
	}

	// Verify directory structure was copied
	copiedFile := filepath.Join(copyDstDir, "testdir", "nested.txt")
	content, err := os.ReadFile(copiedFile)
	if err != nil {
		t.Fatalf("Failed to read copied nested file: %v", err)
	}

	expected := "nested content"
	if string(content) != expected {
		t.Errorf("Copied nested file content = %q, want %q", string(content), expected)
	}
}

func TestCopy_WithoutRewrite(t *testing.T) {
	tmpDir := t.TempDir()
	defer os.RemoveAll(tmpDir)

	testLinksDir := filepath.Join(tmpDir, "links")

	setupTestLinksInDir(t, testLinksDir)

	copyDstDir := t.TempDir()
	defer os.RemoveAll(copyDstDir)

	f := newTestFiles(testLinksDir)

	// Create existing file
	existingPath := filepath.Join(copyDstDir, "test.txt")
	err := os.WriteFile(existingPath, []byte("existing"), 0644)
	if err != nil {
		t.Fatalf("Failed to create existing file: %v", err)
	}

	// Try to copy without rewrite
	err = f.Copy("test.txt", copyDstDir, false)
	if err == nil {
		t.Error("Copy() without rewrite should fail for existing file")
	}
}

func TestCopy_WithRewrite(t *testing.T) {
	tmpDir := t.TempDir()
	defer os.RemoveAll(tmpDir)

	testLinksDir := filepath.Join(tmpDir, "links")

	setupTestLinksInDir(t, testLinksDir)

	copyDstDir := t.TempDir()
	defer os.RemoveAll(copyDstDir)

	f := newTestFiles(testLinksDir)

	// Create existing file
	existingPath := filepath.Join(copyDstDir, "test.txt")
	err := os.WriteFile(existingPath, []byte("existing"), 0644)
	if err != nil {
		t.Fatalf("Failed to create existing file: %v", err)
	}

	// Copy with rewrite
	err = f.Copy("test.txt", copyDstDir, true)
	if err != nil {
		t.Fatalf("Copy() with rewrite error = %v", err)
	}

	// Verify file was overwritten
	content, err := os.ReadFile(existingPath)
	if err != nil {
		t.Fatalf("Failed to read overwritten file: %v", err)
	}

	expected := "test content"
	if string(content) != expected {
		t.Errorf("Overwritten file content = %q, want %q", string(content), expected)
	}
}

func TestExists_WithoutEmbedded(t *testing.T) {
	f := NewFiles()

	// Without embedded provider, Exists should return false
	if f.Exists("any.txt") {
		t.Error("Exists() = true, want false when no embedded provider")
	}
}

func TestCopy_FromEmbedded(t *testing.T) {
	f := NewFiles()

	// This test only runs if embedded provider is available
	if !f.Exists("docker-compose.yml") {
		t.Skip("Skipping embedded test - no embedded provider")
	}

	tmpDir := t.TempDir()
	defer os.RemoveAll(tmpDir)

	err := f.Copy("docker-compose.yml", tmpDir, false)
	if err != nil {
		t.Fatalf("Copy() from embedded error = %v", err)
	}

	// Verify file was copied from embedded FS
	copiedPath := filepath.Join(tmpDir, "docker-compose.yml")
	content, err := os.ReadFile(copiedPath)
	if err != nil {
		t.Fatalf("Failed to read copied file: %v", err)
	}

	if len(content) == 0 {
		t.Error("Copied file is empty")
	}

	// Verify content matches what we get from embedded FS
	embeddedContent, err := f.GetContent("docker-compose.yml")
	if err != nil {
		t.Fatalf("Failed to get embedded content: %v", err)
	}

	if string(content) != string(embeddedContent) {
		t.Error("Copied file content doesn't match embedded content")
	}
}

func TestCheck_Missing(t *testing.T) {
	tmpDir := t.TempDir()
	defer os.RemoveAll(tmpDir)

	testLinksDir := filepath.Join(tmpDir, "links")
	setupTestLinksInDir(t, testLinksDir)

	f := newTestFiles(testLinksDir)

	workingDir := t.TempDir()
	defer os.RemoveAll(workingDir)

	status := f.Check("test.txt", workingDir)
	if status != FileStatusMissing {
		t.Errorf("Check() = %v, want %v for missing file", status, FileStatusMissing)
	}
}

func TestCheck_OK(t *testing.T) {
	tmpDir := t.TempDir()
	defer os.RemoveAll(tmpDir)

	testLinksDir := filepath.Join(tmpDir, "links")
	setupTestLinksInDir(t, testLinksDir)

	f := newTestFiles(testLinksDir)

	workingDir := t.TempDir()
	defer os.RemoveAll(workingDir)

	// copy file to working directory
	err := f.Copy("test.txt", workingDir, false)
	if err != nil {
		t.Fatalf("Copy() error = %v", err)
	}

	status := f.Check("test.txt", workingDir)
	if status != FileStatusOK {
		t.Errorf("Check() = %v, want %v for matching file", status, FileStatusOK)
	}
}

func TestCheck_Modified(t *testing.T) {
	tmpDir := t.TempDir()
	defer os.RemoveAll(tmpDir)

	testLinksDir := filepath.Join(tmpDir, "links")
	setupTestLinksInDir(t, testLinksDir)

	f := newTestFiles(testLinksDir)

	workingDir := t.TempDir()
	defer os.RemoveAll(workingDir)

	// create modified file in working directory
	modifiedPath := filepath.Join(workingDir, "test.txt")
	err := os.WriteFile(modifiedPath, []byte("modified content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create modified file: %v", err)
	}

	status := f.Check("test.txt", workingDir)
	if status != FileStatusModified {
		t.Errorf("Check() = %v, want %v for modified file", status, FileStatusModified)
	}
}

func TestList(t *testing.T) {
	// test with real embedded provider (if available)
	f := NewFiles()

	// test listing with observability prefix (should exist in embedded)
	files, err := f.List("observability")
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	// we should have at least some observability files
	if len(files) == 0 {
		t.Error("List() with 'observability' prefix returned no files from embedded")
	}

	// verify we get some expected files
	foundObservabilityFile := false
	for _, file := range files {
		if strings.HasPrefix(file, "observability/") {
			foundObservabilityFile = true
			break
		}
	}

	if !foundObservabilityFile {
		t.Error("List() with 'observability' prefix did not include any observability files")
	}

	// test listing with non-existent prefix
	emptyFiles, err := f.List("nonexistent-prefix")
	if err != nil {
		t.Fatalf("List() with non-existent prefix error = %v", err)
	}

	if len(emptyFiles) != 0 {
		t.Errorf("List() with non-existent prefix returned %d files, want 0", len(emptyFiles))
	}
}

func TestList_NonExistentPrefix(t *testing.T) {
	tmpDir := t.TempDir()
	defer os.RemoveAll(tmpDir)

	testLinksDir := filepath.Join(tmpDir, "links")
	setupTestLinksInDir(t, testLinksDir)

	f := newTestFiles(testLinksDir)

	files, err := f.List("nonexistent")
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	if len(files) != 0 {
		t.Errorf("List() for nonexistent prefix = %v, want empty slice", files)
	}
}

func TestCheck_HashComparison_Embedded(t *testing.T) {
	// test with real embedded files that have metadata
	f := NewFiles()

	workingDir := t.TempDir()
	defer os.RemoveAll(workingDir)

	// copy embedded file to working directory
	embeddedFile := "docker-compose.yml"
	err := f.Copy(embeddedFile, workingDir, false)
	if err != nil {
		t.Fatalf("Copy() error = %v", err)
	}

	// check should return OK (hash matches)
	status := f.Check(embeddedFile, workingDir)
	if status != FileStatusOK {
		t.Errorf("Check() hash comparison = %v, want %v for embedded file", status, FileStatusOK)
	}
}

func TestCheck_HashComparison_SameSize_DifferentContent(t *testing.T) {
	// test case where file has same size but different content (different hash)
	f := NewFiles()

	workingDir := t.TempDir()
	defer os.RemoveAll(workingDir)

	// get metadata for docker-compose.yml to know its size
	embeddedContent, err := f.GetContent("docker-compose.yml")
	if err != nil {
		t.Skip("Skipping test - docker-compose.yml not available")
	}

	originalSize := len(embeddedContent)

	// create file with same size but different content
	modifiedContent := make([]byte, originalSize)
	for i := range modifiedContent {
		modifiedContent[i] = 'X' // fill with different content
	}

	modifiedPath := filepath.Join(workingDir, "docker-compose.yml")
	err = os.WriteFile(modifiedPath, modifiedContent, 0644)
	if err != nil {
		t.Fatalf("Failed to create modified file: %v", err)
	}

	// check should return Modified (same size, different hash)
	status := f.Check("docker-compose.yml", workingDir)
	if status != FileStatusModified {
		t.Errorf("Check() same size different hash = %v, want %v", status, FileStatusModified)
	}
}

func TestCheck_HashComparison_DifferentSize(t *testing.T) {
	// test case where file has different size (quick size check should catch this)
	f := NewFiles()

	workingDir := t.TempDir()
	defer os.RemoveAll(workingDir)

	// create file with different size
	modifiedContent := []byte("different size content")
	modifiedPath := filepath.Join(workingDir, "docker-compose.yml")
	err := os.WriteFile(modifiedPath, modifiedContent, 0644)
	if err != nil {
		t.Fatalf("Failed to create modified file: %v", err)
	}

	// check should return Modified (different size detected quickly)
	status := f.Check("docker-compose.yml", workingDir)
	if status != FileStatusModified {
		t.Errorf("Check() different size = %v, want %v", status, FileStatusModified)
	}
}

// Helper functions

// setupTestLinksInDir creates test files structure in specified directory
func setupTestLinksInDir(t *testing.T, linksDir string) {
	err := os.MkdirAll(linksDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create test links directory: %v", err)
	}

	testFile := filepath.Join(linksDir, "test.txt")
	err = os.WriteFile(testFile, []byte("test content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
}

// setupTestLinksWithDirInDir creates test files and directories structure in specified directory
func setupTestLinksWithDirInDir(t *testing.T, linksDir string) {
	setupTestLinksInDir(t, linksDir)

	testDir := filepath.Join(linksDir, "testdir")
	err := os.MkdirAll(testDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	nestedFile := filepath.Join(testDir, "nested.txt")
	err = os.WriteFile(nestedFile, []byte("nested content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create nested test file: %v", err)
	}
}
