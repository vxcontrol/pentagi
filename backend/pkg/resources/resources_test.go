package resources

import (
	"archive/zip"
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResourceDirsAndBlobPath(t *testing.T) {
	assert.Equal(t, "/data/resources", ResourcesDir("/data"))
	assert.Equal(t, "/data/resources/900150983cd24fb0d6963f7d28e17f72.blob",
		BlobPath("/data", "900150983cd24fb0d6963f7d28e17f72"))

	invalidPath := BlobPath("/data", "../evil")
	assert.Equal(t, "/data/resources/.invalid-blob-hash.blob", invalidPath)
	assert.NotContains(t, invalidPath, "..")
}

func TestComputeFileMD5(t *testing.T) {
	got, err := ComputeFileMD5(strings.NewReader("abc"))
	require.NoError(t, err)
	assert.Equal(t, "900150983cd24fb0d6963f7d28e17f72", got)
}

func TestIsValidBlobHash(t *testing.T) {
	tests := []struct {
		name string
		hash string
		want bool
	}{
		{"lowercase md5", "900150983cd24fb0d6963f7d28e17f72", true},
		{"uppercase md5", "900150983CD24FB0D6963F7D28E17F72", true},
		{"too short", "abc", false},
		{"non hex", "zzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzz", false},
		{"path traversal", "../900150983cd24fb0d6963f7d28e17f72", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, IsValidBlobHash(tt.hash))
		})
	}
}

func TestBlobLifecycle(t *testing.T) {
	dataDir := t.TempDir()
	hash := "900150983cd24fb0d6963f7d28e17f72"

	exists, err := BlobExists(dataDir, hash)
	require.NoError(t, err)
	assert.False(t, exists)

	tmpPath, gotHash, size, err := SaveToTemp(strings.NewReader("abc"), ResourcesDir(dataDir))
	require.NoError(t, err)
	assert.Equal(t, hash, gotHash)
	assert.Equal(t, int64(3), size)

	require.NoError(t, CommitBlob(dataDir, gotHash, tmpPath))
	exists, err = BlobExists(dataDir, hash)
	require.NoError(t, err)
	assert.True(t, exists)

	tmpPath, _, _, err = SaveToTemp(strings.NewReader("abc"), ResourcesDir(dataDir))
	require.NoError(t, err)
	require.NoError(t, CommitBlob(dataDir, hash, tmpPath))
	_, err = os.Lstat(tmpPath)
	assert.True(t, os.IsNotExist(err), "duplicate commit should remove temp file")

	require.NoError(t, DeleteBlob(dataDir, hash))
	exists, err = BlobExists(dataDir, hash)
	require.NoError(t, err)
	assert.False(t, exists)

	require.Error(t, DeleteBlob(dataDir, "../evil"))
	_, err = BlobExists(dataDir, "../evil")
	require.Error(t, err)
	require.Error(t, CommitBlob(dataDir, "../evil", filepath.Join(dataDir, "tmp")))
}

func TestSanitizeResourcePath(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		wantErr  bool
	}{
		{"plain", "report.txt", "report.txt", false},
		{"nested", "targets/ips.txt", "targets/ips.txt", false},
		{"windows separators", `targets\ips.txt`, "targets/ips.txt", false},
		{"cleans duplicate slashes", "targets//ips.txt", "targets/ips.txt", false},
		{"trims whitespace", "  targets/ips.txt  ", "targets/ips.txt", false},
		{"rejects empty", "   ", "", true},
		{"rejects absolute", "/etc/passwd", "", true},
		{"rejects traversal", "../etc/passwd", "", true},
		{"rejects nested traversal", "targets/../passwd", "", true},
		{"rejects control characters", "bad\nname.txt", "", true},
		{"rejects unsupported characters", `bad"name.txt`, "", true},
		{"rejects too long path", strings.Repeat("a", MaxPathLength+1), "", true},
		{"rejects too long component", strings.Repeat("a", MaxFileNameLength+1), "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SanitizeResourcePath(tt.input)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestSanitizeResourceDir(t *testing.T) {
	got, err := SanitizeResourceDir("   ")
	require.NoError(t, err)
	assert.Empty(t, got)

	got, err = SanitizeResourceDir(`targets\internal`)
	require.NoError(t, err)
	assert.Equal(t, "targets/internal", got)
}

func TestSanitizeResourceFileName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		wantErr  bool
	}{
		{"plain name", "report.txt", "report.txt", false},
		{"strips parent traversal", "../report.txt", "report.txt", false},
		{"normalizes windows separators", `nested\brief.md`, "brief.md", false},
		{"strips absolute path", "/etc/passwd", "passwd", false},
		{"trims whitespace", "  wordlist.txt  ", "wordlist.txt", false},
		{"rejects empty", "   ", "", true},
		{"rejects dot-only", ".", "", true},
		{"rejects root slash", "/", "", true},
		{"rejects control characters", "bad\nname.txt", "", true},
		{"rejects unsupported characters", `bad"name.txt`, "", true},
		{"rejects too long", strings.Repeat("a", MaxFileNameLength+1), "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SanitizeResourceFileName(tt.input)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestPathHelpers(t *testing.T) {
	assert.Equal(t, "report.txt", FilePath("", "report.txt"))
	assert.Equal(t, "docs/report.txt", FilePath("docs", "report.txt"))
	assert.Equal(t, "", ParentDir("report.txt"))
	assert.Equal(t, "docs", ParentDir("docs/report.txt"))
	assert.Equal(t, "report.txt", FileName("docs/report.txt"))

	assert.True(t, IsChildOf("report.txt", ""))
	assert.False(t, IsChildOf("docs/report.txt", ""))
	assert.True(t, IsChildOf("docs/report.txt", "docs"))
	assert.False(t, IsChildOf("docs/sub/report.txt", "docs"))

	assert.True(t, PathHasPrefix("docs", "docs"))
	assert.True(t, PathHasPrefix("docs/report.txt", "docs"))
	assert.False(t, PathHasPrefix("docs2/report.txt", "docs"))

	assert.Equal(t, "archive/docs/report.txt", ReplacePrefixPath("docs/report.txt", "docs", "archive/docs"))
	assert.Equal(t, "report.txt", ReplacePrefixPath("docs/report.txt", "docs", ""))
	assert.Equal(t, "archive/report.txt", ReplacePrefixPath("report.txt", "", "archive"))
}

func TestEscapeLike(t *testing.T) {
	assert.Equal(t, `a\%b\_c\\d`, EscapeLike(`a%b_c\d`))
}

func TestZipResources(t *testing.T) {
	dataDir := t.TempDir()
	hash := md5Hex("alpha")
	blobPath := BlobPath(dataDir, hash)
	require.NoError(t, os.MkdirAll(filepath.Dir(blobPath), 0755))
	require.NoError(t, os.WriteFile(blobPath, []byte("alpha"), 0644))

	var buf bytes.Buffer
	require.NoError(t, ZipResources(&buf, []ZipEntry{
		{BlobPath: blobPath, ZipPath: "docs/a.txt"},
	}))

	zr, err := zip.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	require.NoError(t, err)
	require.Len(t, zr.File, 1)
	assert.Equal(t, "docs/a.txt", zr.File[0].Name)
	rc, err := zr.File[0].Open()
	require.NoError(t, err)
	data, err := io.ReadAll(rc)
	rc.Close()
	require.NoError(t, err)
	assert.Equal(t, "alpha", string(data))

	var bad bytes.Buffer
	err = ZipResources(&bad, []ZipEntry{{BlobPath: blobPath, ZipPath: "../evil.txt"}})
	require.Error(t, err)
}

func TestZipResourcesSkipsNonRegularBlobs(t *testing.T) {
	dataDir := t.TempDir()
	target := BlobPath(dataDir, md5Hex("alpha"))
	require.NoError(t, os.MkdirAll(filepath.Dir(target), 0755))
	require.NoError(t, os.WriteFile(target, []byte("alpha"), 0644))
	link := filepath.Join(dataDir, "resources", "link.blob")
	if err := os.Symlink(target, link); err != nil {
		t.Skipf("symlink creation not available: %v", err)
	}

	var buf bytes.Buffer
	require.NoError(t, ZipResources(&buf, []ZipEntry{{BlobPath: link, ZipPath: "link.txt"}}))

	zr, err := zip.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	require.NoError(t, err)
	assert.Empty(t, zr.File)
}

func TestZipDirectory(t *testing.T) {
	src := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(src, "sub"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(src, "sub", "a.txt"), []byte("alpha"), 0644))
	if err := os.Symlink(filepath.Join(src, "sub", "a.txt"), filepath.Join(src, "link.txt")); err != nil {
		t.Skipf("symlink creation not available: %v", err)
	}

	var buf bytes.Buffer
	require.NoError(t, ZipDirectory(&buf, src))

	zr, err := zip.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	require.NoError(t, err)

	names := make([]string, len(zr.File))
	for i, f := range zr.File {
		names[i] = f.Name
	}
	assert.Contains(t, names, "sub/a.txt")
	assert.NotContains(t, names, "link.txt")
}

func md5Hex(content string) string {
	sum := md5.Sum([]byte(content))
	return hex.EncodeToString(sum[:])
}
