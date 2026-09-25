package loader

import (
	"os"
	"path/filepath"
	"testing"
)

// loaderEnvFile writes content to .env in a fresh directory, where Save also puts its .bak.
func loaderEnvFile(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), ".env")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write env file: %v", err)
	}
	return path
}
