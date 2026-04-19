package services

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSanitizeFlowFileName(t *testing.T) {
	testCases := []struct {
		expected string
		fileName string
		name     string
		wantErr  bool
	}{
		{
			name:     "keeps plain file name",
			fileName: "report.txt",
			expected: "report.txt",
		},
		{
			name:     "collapses parent traversal",
			fileName: "../report.txt",
			expected: "report.txt",
		},
		{
			name:     "normalizes windows separators",
			fileName: `nested\brief.md`,
			expected: "brief.md",
		},
		{
			name:     "rejects empty names",
			fileName: "   ",
			wantErr:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := sanitizeFlowFileName(tc.fileName)

			if tc.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.expected, actual)
		})
	}
}

func TestFlowFileService_ListFlowFiles(t *testing.T) {
	dataDir := t.TempDir()
	service := NewFlowFileService(nil, dataDir)
	uploadDir := filepath.Join(dataDir, "flow-7", "uploads")

	require.NoError(t, os.MkdirAll(uploadDir, 0755))
	require.NoError(t, os.Mkdir(filepath.Join(uploadDir, "nested"), 0755))

	oldPath := filepath.Join(uploadDir, "old.txt")
	newPath := filepath.Join(uploadDir, "new.txt")

	require.NoError(t, os.WriteFile(oldPath, []byte("old"), 0644))
	require.NoError(t, os.WriteFile(newPath, []byte("newer content"), 0644))

	oldTime := time.Now().Add(-2 * time.Hour)
	newTime := time.Now().Add(-1 * time.Hour)
	require.NoError(t, os.Chtimes(oldPath, oldTime, oldTime))
	require.NoError(t, os.Chtimes(newPath, newTime, newTime))

	files, err := service.listFlowFiles(7)
	require.NoError(t, err)
	require.Len(t, files, 2)

	assert.Equal(t, "new.txt", files[0].Name)
	assert.Equal(t, "/work/uploads/new.txt", files[0].Path)
	assert.Equal(t, int64(len("newer content")), files[0].Size)

	assert.Equal(t, "old.txt", files[1].Name)
	assert.Equal(t, "/work/uploads/old.txt", files[1].Path)
	assert.Equal(t, int64(len("old")), files[1].Size)
}

func TestFlowFileService_ListFlowFiles_MissingDirectory(t *testing.T) {
	service := NewFlowFileService(nil, t.TempDir())

	files, err := service.listFlowFiles(999)
	require.NoError(t, err)
	assert.Empty(t, files)
}
