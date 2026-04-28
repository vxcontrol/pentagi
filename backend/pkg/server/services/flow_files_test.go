package services

import (
	"archive/tar"
	"archive/zip"
	"bufio"
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"pentagi/pkg/database"
	"pentagi/pkg/docker"
	"pentagi/pkg/flowfiles"
	graphmodel "pentagi/pkg/graph/model"
	"pentagi/pkg/graph/subscriptions"
	"pentagi/pkg/providers/pconfig"
	"pentagi/pkg/resources"
	"pentagi/pkg/server/models"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ── sanitizeFlowFileName ─────────────────────────────────────────────────────

func TestSanitizeFlowFileName(t *testing.T) {
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
		{"strips deep traversal", "../../etc/shadow", "shadow", false},
		{"trims whitespace", "  wordlist.txt  ", "wordlist.txt", false},
		{"rejects empty", "   ", "", true},
		{"rejects dot-only", ".", "", true},
		{"rejects root slash", "/", "", true},
		{"rejects control characters", "bad\nname.txt", "", true},
		{"rejects unsupported header characters", `bad"name.txt`, "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := flowfiles.SanitizeFileName(tt.input)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestSanitizeContainerCachePath(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		wantErr  bool
	}{
		{"absolute directory", "/etc/nginx/conf/", "etc/nginx/conf", false},
		{"absolute file", "/etc/nginx/nginx.conf", "etc/nginx/nginx.conf", false},
		{"relative file", "var/log/app.log", "var/log/app.log", false},
		{"normalizes traversal", "../../etc/shadow", "etc/shadow", false},
		{"rejects empty", "   ", "", true},
		{"rejects root", "/", "", true},
		{"rejects bad component", "/etc/bad\nname", "", true},
		{"rejects unsupported component", `/etc/bad"name`, "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := flowfiles.SanitizeContainerCachePath(tt.input)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.expected, got)
		})
	}
}

// ── directory helpers ─────────────────────────────────────────────────────────

func TestFlowFileService_Dirs(t *testing.T) {
	svc := NewFlowFileService(nil, "/data", nil, nil)
	assert.Equal(t, "/data/flow-42-data", svc.flowDataDir(42))
	assert.Equal(t, "/data/flow-42-data/uploads", svc.flowUploadsDir(42))
	assert.Equal(t, "/data/flow-42-data/container", svc.flowContainerDir(42))
	assert.Equal(t, "/data/flow-42-data/resources", flowfiles.FlowResourcesDir("/data", 42))
}

func TestMaxUploadFileSize(t *testing.T) {
	const expectedMaxMB = 300
	assert.Equal(t, int64(expectedMaxMB*1024*1024), int64(flowfiles.MaxUploadFileSize))
}

// ── resolveCachedPath ─────────────────────────────────────────────────────────

func TestResolveCachedPath(t *testing.T) {
	svc := NewFlowFileService(nil, "/data", nil, nil)

	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{"uploads file", "uploads/report.txt", "/data/flow-1-data/uploads/report.txt", false},
		{"container file", "container/conf/nginx.conf", "/data/flow-1-data/container/conf/nginx.conf", false},
		{"container top-level", "container/conf", "/data/flow-1-data/container/conf", false},
		{"resources file", "resources/creds/passwords.txt", "/data/flow-1-data/resources/creds/passwords.txt", false},
		{"empty path", "", "", true},
		{"wrong prefix", "tmp/evil.sh", "", true},
		{"absolute path", "/etc/passwd", "", true},
		{"path traversal", "uploads/../../etc/passwd", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := svc.resolveCachedPath(1, tt.input)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

// ── listDirEntries ────────────────────────────────────────────────────────────

func TestListDirEntries_Basic(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "a.txt"), []byte("a"), 0644))
	require.NoError(t, os.Mkdir(filepath.Join(dir, "sub"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".upload-temp"), []byte("tmp"), 0644))
	require.NoError(t, os.Mkdir(filepath.Join(dir, ".pull-temp"), 0755))
	_ = os.Symlink(filepath.Join(dir, "a.txt"), filepath.Join(dir, "link.txt"))

	entries, err := flowfiles.ListDirEntries(dir, "uploads")
	require.NoError(t, err)
	require.Len(t, entries, 2) // a.txt + sub; link.txt excluded

	names := make([]string, len(entries))
	for i, e := range entries {
		names[i] = e.Name
	}
	assert.Contains(t, names, "a.txt")
	assert.Contains(t, names, "sub")
	assert.NotContains(t, names, ".upload-temp")
	assert.NotContains(t, names, ".pull-temp")
	assert.NotContains(t, names, "link.txt")

	for _, e := range entries {
		if e.Name == "sub" {
			assert.True(t, e.IsDir)
			assert.Equal(t, "uploads/sub", e.Path)
		} else {
			assert.False(t, e.IsDir)
			assert.Equal(t, "uploads/a.txt", e.Path)
		}
	}
}

func TestListDirEntries_MissingDir(t *testing.T) {
	entries, err := flowfiles.ListDirEntries("/nonexistent/path", "uploads")
	require.NoError(t, err)
	assert.Empty(t, entries)
}

func TestListDirEntriesRecursive_PreservesNestedPaths(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "etc", "nginx", "conf"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "etc", "nginx", "nginx.conf"), []byte("nginx"), 0644))
	require.NoError(t, os.Mkdir(filepath.Join(dir, ".pull-temp"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".pull-temp", "tmp.txt"), []byte("tmp"), 0644))

	entries, err := flowfiles.ListDirEntriesRecursive(dir, "container")
	require.NoError(t, err)

	paths := make([]string, len(entries))
	for i, entry := range entries {
		paths[i] = entry.Path
	}
	assert.Contains(t, paths, "container/etc")
	assert.Contains(t, paths, "container/etc/nginx")
	assert.Contains(t, paths, "container/etc/nginx/conf")
	assert.Contains(t, paths, "container/etc/nginx/nginx.conf")
	assert.NotContains(t, paths, "container/.pull-temp")
	assert.NotContains(t, paths, "container/.pull-temp/tmp.txt")
}

// ── listFlowFiles ─────────────────────────────────────────────────────────────

func TestFlowFileService_ListFlowFiles_BothSources(t *testing.T) {
	dataDir := t.TempDir()
	svc := NewFlowFileService(nil, dataDir, nil, nil)

	uploadsDir := filepath.Join(dataDir, "flow-7-data", "uploads")
	containerDir := filepath.Join(dataDir, "flow-7-data", "container")
	resourcesDir := filepath.Join(dataDir, "flow-7-data", "resources")
	require.NoError(t, os.MkdirAll(uploadsDir, 0755))
	require.NoError(t, os.MkdirAll(containerDir, 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(resourcesDir, "creds"), 0755))

	require.NoError(t, os.WriteFile(filepath.Join(uploadsDir, "wordlist.txt"), []byte("words"), 0644))
	require.NoError(t, os.MkdirAll(filepath.Join(containerDir, "etc", "nginx", "conf"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(containerDir, "etc", "nginx", "nginx.conf"), []byte("nginx"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(resourcesDir, "creds", "passwords.txt"), []byte("secret"), 0644))

	files, err := svc.listFlowFiles(7)
	require.NoError(t, err)

	paths := make([]string, len(files))
	for i, f := range files {
		paths[i] = f.Path
	}
	assert.Contains(t, paths, "uploads/wordlist.txt")
	assert.Contains(t, paths, "container/etc/nginx/conf")
	assert.Contains(t, paths, "container/etc/nginx/nginx.conf")
	assert.Contains(t, paths, "resources/creds")
	assert.Contains(t, paths, "resources/creds/passwords.txt")
}

func TestFlowFileService_ListFlowFiles_EmptyDirs(t *testing.T) {
	svc := NewFlowFileService(nil, t.TempDir(), nil, nil)
	files, err := svc.listFlowFiles(999)
	require.NoError(t, err)
	assert.Empty(t, files)
}

func TestConvertModelFlowFile_PreservesID(t *testing.T) {
	file := models.FlowFile{
		ID:         "flow-file-id",
		Name:       "report.md",
		Path:       "uploads/report.md",
		Size:       42,
		IsDir:      false,
		ModifiedAt: time.Now(),
	}

	modelFile := convertModelFlowFile(file)

	require.NotNil(t, modelFile)
	assert.Equal(t, file.ID, modelFile.ID)
	assert.Equal(t, file.Name, modelFile.Name)
	assert.Equal(t, file.Path, modelFile.Path)
	assert.Equal(t, int(file.Size), modelFile.Size)
	assert.Equal(t, file.IsDir, modelFile.IsDir)
	assert.Equal(t, file.ModifiedAt, modelFile.ModifiedAt)
}

func TestConvertContainerFiles(t *testing.T) {
	mtime := time.Now()
	files := convertContainerFiles("/work", []container.PathStat{
		{
			Name:  "zeta.txt",
			Size:  10,
			Mode:  0644,
			Mtime: mtime,
		},
		{
			Name:  "alpha",
			Mode:  os.ModeDir | 0755,
			Mtime: mtime.Add(time.Second),
		},
	})

	require.Len(t, files, 2)
	assert.Equal(t, "alpha", files[0].Name)
	assert.Equal(t, "/work/alpha", files[0].Path)
	assert.Equal(t, flowfiles.ID("/work/alpha"), files[0].ID)
	assert.True(t, files[0].IsDir)
	assert.Equal(t, int64(0), files[0].Size)

	assert.Equal(t, "zeta.txt", files[1].Name)
	assert.Equal(t, "/work/zeta.txt", files[1].Path)
	assert.Equal(t, flowfiles.ID("/work/zeta.txt"), files[1].ID)
	assert.False(t, files[1].IsDir)
	assert.Equal(t, int64(10), files[1].Size)
}

// ── localEntryExists ─────────────────────────────────────────────────────────

func TestLocalEntryExists(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "f.txt")

	exists, err := flowfiles.LocalEntryExists(p)
	require.NoError(t, err)
	assert.False(t, exists)

	require.NoError(t, os.WriteFile(p, []byte("x"), 0644))

	exists, err = flowfiles.LocalEntryExists(p)
	require.NoError(t, err)
	assert.True(t, exists)
}

// ── isWithinDir ───────────────────────────────────────────────────────────────

func TestIsWithinDir(t *testing.T) {
	assert.True(t, flowfiles.IsWithinDir("/data/flow-1/uploads/file.txt", "/data/flow-1/uploads"))
	assert.True(t, flowfiles.IsWithinDir("/data/flow-1/uploads/sub/file.txt", "/data/flow-1/uploads"))
	assert.False(t, flowfiles.IsWithinDir("/data/flow-1/../evil.txt", "/data/flow-1/uploads"))
	assert.False(t, flowfiles.IsWithinDir("/data/flow-2/uploads/file.txt", "/data/flow-1/uploads"))
}

func TestResolvePulledStagedTarget(t *testing.T) {
	t.Run("full cache path archive", func(t *testing.T) {
		stagingDir := t.TempDir()
		target := filepath.Join(stagingDir, "etc", "nginx", "nginx.conf")
		require.NoError(t, os.MkdirAll(filepath.Dir(target), 0755))
		require.NoError(t, os.WriteFile(target, []byte("nginx"), 0644))

		assert.Equal(t, target, flowfiles.ResolvePulledStagedTarget(stagingDir, "etc/nginx/nginx.conf"))
	})

	t.Run("basename archive", func(t *testing.T) {
		stagingDir := t.TempDir()
		target := filepath.Join(stagingDir, "nginx.conf")
		require.NoError(t, os.WriteFile(target, []byte("nginx"), 0644))

		assert.Equal(t, target, flowfiles.ResolvePulledStagedTarget(stagingDir, "etc/nginx/nginx.conf"))
	})
}

// ── extractTar ────────────────────────────────────────────────────────────────

func buildTar(entries []tarTestEntry) *bytes.Buffer {
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)
	for _, e := range entries {
		hdr := &tar.Header{
			Name:     e.name,
			Typeflag: e.typeflag,
			Mode:     0644,
			Size:     int64(len(e.content)),
			Linkname: e.linkname,
		}
		if e.typeflag == tar.TypeDir {
			hdr.Mode = 0755
		}
		_ = tw.WriteHeader(hdr)
		if len(e.content) > 0 {
			_, _ = tw.Write([]byte(e.content))
		}
	}
	tw.Close()
	return &buf
}

type tarTestEntry struct {
	name     string
	typeflag byte
	content  string
	linkname string
}

func TestExtractTar_RegularFiles(t *testing.T) {
	destDir := t.TempDir()

	buf := buildTar([]tarTestEntry{
		{name: "dir/", typeflag: tar.TypeDir},
		{name: "dir/file.txt", typeflag: tar.TypeReg, content: "hello"},
	})

	require.NoError(t, flowfiles.ExtractTar(buf, destDir))

	data, err := os.ReadFile(filepath.Join(destDir, "dir", "file.txt"))
	require.NoError(t, err)
	assert.Equal(t, "hello", string(data))
}

func TestExtractTar_SkipsSymlinks(t *testing.T) {
	destDir := t.TempDir()

	buf := buildTar([]tarTestEntry{
		{name: "link.txt", typeflag: tar.TypeSymlink, linkname: "/etc/passwd"},
	})

	require.NoError(t, flowfiles.ExtractTar(buf, destDir))

	_, err := os.Lstat(filepath.Join(destDir, "link.txt"))
	assert.True(t, os.IsNotExist(err), "symlink must not be created in cache")
}

func TestExtractTar_PathTraversal(t *testing.T) {
	destDir := t.TempDir()

	buf := buildTar([]tarTestEntry{
		{name: "../../evil.txt", typeflag: tar.TypeReg, content: "evil"},
	})

	require.NoError(t, flowfiles.ExtractTar(buf, destDir))

	// evil.txt must NOT appear outside destDir
	evilPath := filepath.Join(filepath.Dir(destDir), "evil.txt")
	_, err := os.Lstat(evilPath)
	assert.True(t, os.IsNotExist(err), "path traversal file must not be created")
}

func TestExtractTar_RejectsTooManyFiles(t *testing.T) {
	destDir := t.TempDir()
	entries := make([]tarTestEntry, 0, flowfiles.MaxPullFiles+1)
	for i := 0; i < flowfiles.MaxPullFiles+1; i++ {
		entries = append(entries, tarTestEntry{
			name:     filepath.Join("many", "file-"+strconv.Itoa(i)+".txt"),
			typeflag: tar.TypeReg,
			content:  "x",
		})
	}

	err := flowfiles.ExtractTar(buildTar(entries), destDir)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "maximum file count")
}

// ── zipDirectory ─────────────────────────────────────────────────────────────

func TestZipDirectory(t *testing.T) {
	src := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(src, "a.txt"), []byte("hello"), 0644))
	require.NoError(t, os.MkdirAll(filepath.Join(src, "sub"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(src, "sub", "b.txt"), []byte("world"), 0644))

	var buf bytes.Buffer
	require.NoError(t, flowfiles.ZipDirectory(&buf, src))

	zr, err := zip.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	require.NoError(t, err)

	contents := make(map[string]string)
	for _, f := range zr.File {
		rc, err := f.Open()
		require.NoError(t, err)
		data, _ := io.ReadAll(rc)
		rc.Close()
		contents[f.Name] = string(data)
	}
	assert.Equal(t, "hello", contents["a.txt"])
	assert.Equal(t, "world", contents["sub/b.txt"])
}

func TestZipDirectory_ExcludesSymlinks(t *testing.T) {
	src := t.TempDir()
	target := filepath.Join(src, "real.txt")
	link := filepath.Join(src, "link.txt")
	require.NoError(t, os.WriteFile(target, []byte("real"), 0644))
	if err := os.Symlink(target, link); err != nil {
		t.Skipf("symlink creation not available: %v", err)
	}

	var buf bytes.Buffer
	require.NoError(t, flowfiles.ZipDirectory(&buf, src))

	zr, err := zip.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	require.NoError(t, err)

	names := make([]string, len(zr.File))
	for i, f := range zr.File {
		names[i] = f.Name
	}
	assert.Contains(t, names, "real.txt")
	assert.NotContains(t, names, "link.txt")
}

// ── primaryContainerName ─────────────────────────────────────────────────────

func TestPrimaryContainerName(t *testing.T) {
	assert.Equal(t, "pentagi-terminal-42", primaryContainerName(42))
}

// ── sorting ───────────────────────────────────────────────────────────────────

func TestSortFlowFiles(t *testing.T) {
	now := time.Now()
	files := []models.FlowFile{
		{Name: "b.txt", ModifiedAt: now.Add(-2 * time.Hour)},
		{Name: "a.txt", ModifiedAt: now.Add(-1 * time.Hour)},
		{Name: "c.txt", ModifiedAt: now.Add(-1 * time.Hour)},
	}
	sortFlowFiles(files)

	// Newer first; ties broken alphabetically.
	assert.Equal(t, "a.txt", files[0].Name)
	assert.Equal(t, "c.txt", files[1].Name)
	assert.Equal(t, "b.txt", files[2].Name)
}

// ── pure helpers ─────────────────────────────────────────────────────────────

func TestShellQuote(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"plain path", "/work/uploads/report.txt", "'/work/uploads/report.txt'"},
		{"path with space", "/work/uploads/my report.txt", "'/work/uploads/my report.txt'"},
		{"path with single quote", "/tmp/it's-mine.txt", `'/tmp/it'"'"'s-mine.txt'`},
		{"empty", "", "''"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, shellQuote(tt.input))
		})
	}
}

func TestParseFlowIDParam(t *testing.T) {
	tests := []struct {
		name    string
		param   string
		want    uint64
		wantErr bool
	}{
		{"numeric id", "42", 42, false},
		{"zero", "0", 0, false},
		{"negative", "-1", 0, true},
		{"non-numeric", "abc", 0, true},
		{"empty", "", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			c.Params = gin.Params{{Key: "flowID", Value: tt.param}}

			got, err := parseFlowIDParam(c)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestCleanupPendingUploads(t *testing.T) {
	dir := t.TempDir()
	keep := filepath.Join(dir, "keep.txt")
	tmp1 := filepath.Join(dir, "tmp1.txt")
	tmp2 := filepath.Join(dir, "tmp2.txt")
	require.NoError(t, os.WriteFile(keep, []byte("keep"), 0644))
	require.NoError(t, os.WriteFile(tmp1, []byte("a"), 0644))
	require.NoError(t, os.WriteFile(tmp2, []byte("b"), 0644))

	cleanupPendingUploads([]pendingUpload{
		{tmpPath: tmp1},
		{tmpPath: ""}, // committed already, must be skipped
		{tmpPath: tmp2},
	})

	_, err := os.Lstat(tmp1)
	assert.True(t, os.IsNotExist(err))
	_, err = os.Lstat(tmp2)
	assert.True(t, os.IsNotExist(err))
	_, err = os.Lstat(keep)
	assert.NoError(t, err, "non-pending files must not be touched")
}

func TestFlowScopeForFiles(t *testing.T) {
	tests := []struct {
		name        string
		privs       []string
		writeAccess bool
		wantNil     bool
		wantSQL     string
	}{
		{
			name:    "admin bypasses uid filter on read",
			privs:   []string{"flow_files.admin"},
			wantSQL: "id = ?",
		},
		{
			name:        "admin bypasses uid filter on write",
			privs:       []string{"flow_files.admin"},
			writeAccess: true,
			wantSQL:     "id = ?",
		},
		{
			name:        "upload privilege grants write to own flow",
			privs:       []string{"flow_files.upload"},
			writeAccess: true,
			wantSQL:     "id = ? AND user_id = ?",
		},
		{
			name:    "view privilege grants read to own flow",
			privs:   []string{"flow_files.view"},
			wantSQL: "id = ? AND user_id = ?",
		},
		{
			name:        "view privilege does not grant write",
			privs:       []string{"flow_files.view"},
			writeAccess: true,
			wantNil:     true,
		},
		{
			name:    "upload privilege does not grant read",
			privs:   []string{"flow_files.upload"},
			wantNil: true,
		},
		{
			name:        "no privileges denies access",
			privs:       []string{},
			wantNil:     true,
			writeAccess: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scope := flowScopeForFiles(tt.privs, 1, 42, tt.writeAccess)
			if tt.wantNil {
				assert.Nil(t, scope)
				return
			}
			require.NotNil(t, scope)
		})
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Test infrastructure for FlowFileService HTTP handlers.
// ─────────────────────────────────────────────────────────────────────────────

func setupFlowFileServiceTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	db.LogMode(false)

	require.NoError(t, db.Exec(`
		CREATE TABLE flows (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			status TEXT NOT NULL DEFAULT 'created',
			title TEXT NOT NULL DEFAULT 'untitled',
			model TEXT NOT NULL DEFAULT '',
			model_provider_name TEXT NOT NULL DEFAULT '',
			model_provider_type TEXT NOT NULL DEFAULT 'openai',
			language TEXT NOT NULL DEFAULT 'english',
			functions TEXT NOT NULL DEFAULT '{}',
			tool_call_id_template TEXT NOT NULL DEFAULT '',
			trace_id TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			deleted_at DATETIME
		)
	`).Error)

	require.NoError(t, db.Exec(`
		CREATE TABLE user_resources (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			hash TEXT NOT NULL DEFAULT '',
			name TEXT NOT NULL,
			path TEXT NOT NULL,
			size INTEGER NOT NULL DEFAULT 0,
			is_dir BOOLEAN NOT NULL DEFAULT FALSE,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(user_id, path)
		)
	`).Error)

	t.Cleanup(func() {
		require.NoError(t, db.Close())
	})

	return db
}

func seedFlow(t *testing.T, db *gorm.DB, id, userID uint64) {
	t.Helper()

	require.NoError(t, db.Exec(
		`INSERT INTO flows (id, user_id, model, model_provider_name, tool_call_id_template, trace_id) VALUES (?, ?, 'gpt', 'openai', 'tcid', '')`,
		id, userID,
	).Error)
}

func seedUserResource(t *testing.T, db *gorm.DB, rec models.UserResource) models.UserResource {
	t.Helper()

	require.NoError(t, db.Create(&rec).Error)
	return rec
}

func md5HexForFlowFiles(content string) string {
	sum := md5.Sum([]byte(content))
	return hex.EncodeToString(sum[:])
}

// newFlowFileTestContext creates a gin test context with the path param,
// uid, prm and request body pre-populated.
func newFlowFileTestContext(
	method, target string,
	body io.Reader,
	privs []string,
	uid, flowID uint64,
) (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("uid", uid)
	c.Set("prm", privs)
	c.Params = gin.Params{
		{Key: "flowID", Value: strconv.FormatUint(flowID, 10)},
	}
	c.Request = httptest.NewRequest(method, target, body)
	return c, w
}

func decodeFlowFilesResponse(t *testing.T, w *httptest.ResponseRecorder) models.FlowFiles {
	t.Helper()

	var resp struct {
		Status string           `json:"status"`
		Data   models.FlowFiles `json:"data"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	require.Equal(t, "success", resp.Status)
	return resp.Data
}

func decodeContainerFilesResponse(t *testing.T, w *httptest.ResponseRecorder) models.ContainerFiles {
	t.Helper()

	var resp struct {
		Status string                `json:"status"`
		Data   models.ContainerFiles `json:"data"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	require.Equal(t, "success", resp.Status)
	return resp.Data
}

func decodeResourceEntryResponse(t *testing.T, w *httptest.ResponseRecorder) models.ResourceEntry {
	t.Helper()

	var resp struct {
		Status string               `json:"status"`
		Data   models.ResourceEntry `json:"data"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	require.Equal(t, "success", resp.Status)
	return resp.Data
}

// flowFileMultipartBody builds a multipart body for upload tests.
type flowFileUploadFile struct {
	name    string
	content string
}

func flowFileMultipartBody(t *testing.T, files []flowFileUploadFile, fieldName string) (*bytes.Buffer, string) {
	t.Helper()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	for _, file := range files {
		part, err := writer.CreateFormFile(fieldName, file.name)
		require.NoError(t, err)
		_, err = part.Write([]byte(file.content))
		require.NoError(t, err)
	}
	require.NoError(t, writer.Close())
	return &body, writer.FormDataContentType()
}

// ─────────────────────────────────────────────────────────────────────────────
// Subscription capture for handler tests.
// ─────────────────────────────────────────────────────────────────────────────

type flowFileEvent struct {
	channel string // "flow" or "resource"
	action  string // "added", "updated", "deleted"
	path    string
	id      string
}

type flowFileCaptureSubscriptions struct {
	mu     sync.Mutex
	events []flowFileEvent
}

func (s *flowFileCaptureSubscriptions) NewFlowSubscriber(int64, int64) subscriptions.FlowSubscriber {
	return nil
}
func (s *flowFileCaptureSubscriptions) NewFlowPublisher(int64, int64) subscriptions.FlowPublisher {
	return &captureFlowPublisher{events: s}
}
func (s *flowFileCaptureSubscriptions) NewResourceSubscriber(int64) subscriptions.ResourceSubscriber {
	return nil
}
func (s *flowFileCaptureSubscriptions) NewResourcePublisher(int64) subscriptions.ResourcePublisher {
	return &captureResourcePublisherForFlow{events: s}
}
func (s *flowFileCaptureSubscriptions) NewProviderSubscriber(int64) subscriptions.ProviderSubscriber {
	return nil
}
func (s *flowFileCaptureSubscriptions) NewProviderPublisher(int64) subscriptions.ProviderPublisher {
	return nil
}
func (s *flowFileCaptureSubscriptions) NewAPITokenSubscriber(int64) subscriptions.APITokenSubscriber {
	return nil
}
func (s *flowFileCaptureSubscriptions) NewAPITokenPublisher(int64) subscriptions.APITokenPublisher {
	return nil
}
func (s *flowFileCaptureSubscriptions) NewSettingsSubscriber(int64) subscriptions.SettingsSubscriber {
	return nil
}
func (s *flowFileCaptureSubscriptions) NewSettingsPublisher(int64) subscriptions.SettingsPublisher {
	return nil
}
func (s *flowFileCaptureSubscriptions) NewFlowTemplateSubscriber(int64) subscriptions.FlowTemplateSubscriber {
	return nil
}
func (s *flowFileCaptureSubscriptions) NewFlowTemplatePublisher(int64) subscriptions.FlowTemplatePublisher {
	return nil
}

func (s *flowFileCaptureSubscriptions) record(e flowFileEvent) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.events = append(s.events, e)
}

func (s *flowFileCaptureSubscriptions) snapshot() []flowFileEvent {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]flowFileEvent, len(s.events))
	copy(out, s.events)
	return out
}

// captureFlowPublisher records FlowFile events; all other methods are no-ops.
type captureFlowPublisher struct {
	flowID int64
	userID int64
	events *flowFileCaptureSubscriptions
}

func (p *captureFlowPublisher) GetFlowID() int64   { return p.flowID }
func (p *captureFlowPublisher) SetFlowID(id int64) { p.flowID = id }
func (p *captureFlowPublisher) GetUserID() int64   { return p.userID }
func (p *captureFlowPublisher) SetUserID(id int64) { p.userID = id }
func (p *captureFlowPublisher) FlowCreated(_ context.Context, _ database.Flow, _ []database.Container) {
}
func (p *captureFlowPublisher) FlowDeleted(_ context.Context, _ database.Flow, _ []database.Container) {
}
func (p *captureFlowPublisher) FlowUpdated(_ context.Context, _ database.Flow, _ []database.Container) {
}
func (p *captureFlowPublisher) TaskCreated(_ context.Context, _ database.Task, _ []database.Subtask) {
}
func (p *captureFlowPublisher) TaskUpdated(_ context.Context, _ database.Task, _ []database.Subtask) {
}
func (p *captureFlowPublisher) AssistantCreated(_ context.Context, _ database.Assistant) {}
func (p *captureFlowPublisher) AssistantUpdated(_ context.Context, _ database.Assistant) {}
func (p *captureFlowPublisher) AssistantDeleted(_ context.Context, _ database.Assistant) {}
func (p *captureFlowPublisher) FlowFileAdded(_ context.Context, file *graphmodel.FlowFile) {
	p.events.record(flowFileEvent{channel: "flow", action: "added", path: file.Path, id: file.ID})
}
func (p *captureFlowPublisher) FlowFileUpdated(_ context.Context, file *graphmodel.FlowFile) {
	p.events.record(flowFileEvent{channel: "flow", action: "updated", path: file.Path, id: file.ID})
}
func (p *captureFlowPublisher) FlowFileDeleted(_ context.Context, file *graphmodel.FlowFile) {
	p.events.record(flowFileEvent{channel: "flow", action: "deleted", path: file.Path, id: file.ID})
}
func (p *captureFlowPublisher) ScreenshotAdded(_ context.Context, _ database.Screenshot) {}
func (p *captureFlowPublisher) TerminalLogAdded(_ context.Context, _ database.Termlog)   {}
func (p *captureFlowPublisher) MessageLogAdded(_ context.Context, _ database.Msglog)     {}
func (p *captureFlowPublisher) MessageLogUpdated(_ context.Context, _ database.Msglog)   {}
func (p *captureFlowPublisher) AgentLogAdded(_ context.Context, _ database.Agentlog)     {}
func (p *captureFlowPublisher) SearchLogAdded(_ context.Context, _ database.Searchlog)   {}
func (p *captureFlowPublisher) VectorStoreLogAdded(_ context.Context, _ database.Vecstorelog) {
}
func (p *captureFlowPublisher) AssistantLogAdded(_ context.Context, _ database.Assistantlog) {}
func (p *captureFlowPublisher) AssistantLogUpdated(_ context.Context, _ database.Assistantlog, _ bool) {
}

var _ subscriptions.FlowPublisher = (*captureFlowPublisher)(nil)
var _ pconfig.ProviderConfig = pconfig.ProviderConfig{} // ensure pconfig import is referenced

// captureResourcePublisherForFlow records Resource events emitted by
// AddResourceFromFlow.  Distinct from captureResourcePublisher used by
// resources_test.go to avoid coupling between test files.
type captureResourcePublisherForFlow struct {
	userID int64
	events *flowFileCaptureSubscriptions
}

func (p *captureResourcePublisherForFlow) GetUserID() int64   { return p.userID }
func (p *captureResourcePublisherForFlow) SetUserID(id int64) { p.userID = id }
func (p *captureResourcePublisherForFlow) ResourceAdded(_ context.Context, r *graphmodel.UserResource) {
	p.events.record(flowFileEvent{channel: "resource", action: "added", path: r.Path})
}
func (p *captureResourcePublisherForFlow) ResourceUpdated(_ context.Context, r *graphmodel.UserResource) {
	p.events.record(flowFileEvent{channel: "resource", action: "updated", path: r.Path})
}
func (p *captureResourcePublisherForFlow) ResourceDeleted(_ context.Context, r *graphmodel.UserResource) {
	p.events.record(flowFileEvent{channel: "resource", action: "deleted", path: r.Path})
}

var _ subscriptions.ResourcePublisher = (*captureResourcePublisherForFlow)(nil)
var _ subscriptions.SubscriptionsController = (*flowFileCaptureSubscriptions)(nil)

// ─────────────────────────────────────────────────────────────────────────────
// fakeDockerClient: a minimal in-memory DockerClient stub for handler tests.
// ─────────────────────────────────────────────────────────────────────────────

type copyToCall struct {
	containerID string
	dstPath     string
	body        []byte
	options     container.CopyToContainerOptions
}

type fakeDockerClient struct {
	mu sync.Mutex

	// IsContainerRunning behaviour.
	running    bool
	runningErr error

	// ContainerStatPath behaviour.
	statPath    container.PathStat
	statPathErr error

	// ListContainerDir behaviour.
	listDir    []container.PathStat
	listDirErr error

	// CopyFromContainer behaviour.
	copyFromBody []byte
	copyFromStat container.PathStat
	copyFromErr  error

	// CopyToContainer behaviour.
	copyToErr   error
	copyToCalls []copyToCall

	// Exec behaviour.
	execCreateID    string
	execCreateErr   error
	execAttachOut   string
	execAttachErr   error
	execInspectCode int
	execInspectErr  error
	execCommands    []string
}

func (f *fakeDockerClient) RunContainer(_ context.Context, _ string, _ database.ContainerType,
	_ int64, _ *container.Config, _ *container.HostConfig) (database.Container, error) {
	return database.Container{}, nil
}
func (f *fakeDockerClient) StopContainer(_ context.Context, _ string, _ int64) error   { return nil }
func (f *fakeDockerClient) RemoveContainer(_ context.Context, _ string, _ int64) error { return nil }
func (f *fakeDockerClient) IsContainerRunning(_ context.Context, _ string) (bool, error) {
	return f.running, f.runningErr
}
func (f *fakeDockerClient) ContainerExecCreate(_ context.Context, _ string, opts container.ExecOptions) (container.ExecCreateResponse, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if len(opts.Cmd) > 0 {
		f.execCommands = append(f.execCommands, strings.Join(opts.Cmd, " "))
	}
	if f.execCreateErr != nil {
		return container.ExecCreateResponse{}, f.execCreateErr
	}
	id := f.execCreateID
	if id == "" {
		id = "exec-id"
	}
	return container.ExecCreateResponse{ID: id}, nil
}
func (f *fakeDockerClient) ContainerExecAttach(_ context.Context, _ string, _ container.ExecAttachOptions) (types.HijackedResponse, error) {
	if f.execAttachErr != nil {
		return types.HijackedResponse{}, f.execAttachErr
	}
	pr, pw := net.Pipe()
	go func() {
		_, _ = pw.Write([]byte(f.execAttachOut))
		_ = pw.Close()
	}()
	return types.HijackedResponse{Conn: pr, Reader: bufio.NewReader(pr)}, nil
}
func (f *fakeDockerClient) ContainerExecInspect(_ context.Context, _ string) (container.ExecInspect, error) {
	if f.execInspectErr != nil {
		return container.ExecInspect{}, f.execInspectErr
	}
	return container.ExecInspect{ExitCode: f.execInspectCode}, nil
}
func (f *fakeDockerClient) ContainerStatPath(_ context.Context, _ string, _ string) (container.PathStat, error) {
	return f.statPath, f.statPathErr
}
func (f *fakeDockerClient) ListContainerDir(_ context.Context, _ string, _ string) ([]container.PathStat, error) {
	return f.listDir, f.listDirErr
}
func (f *fakeDockerClient) CopyToContainer(_ context.Context, containerID string, dstPath string, content io.Reader, options container.CopyToContainerOptions) error {
	body, _ := io.ReadAll(content)
	f.mu.Lock()
	f.copyToCalls = append(f.copyToCalls, copyToCall{
		containerID: containerID,
		dstPath:     dstPath,
		body:        body,
		options:     options,
	})
	f.mu.Unlock()
	return f.copyToErr
}
func (f *fakeDockerClient) CopyFromContainer(_ context.Context, _ string, _ string) (io.ReadCloser, container.PathStat, error) {
	if f.copyFromErr != nil {
		return nil, container.PathStat{}, f.copyFromErr
	}
	return io.NopCloser(bytes.NewReader(f.copyFromBody)), f.copyFromStat, nil
}
func (f *fakeDockerClient) Cleanup(_ context.Context) error { return nil }
func (f *fakeDockerClient) GetDefaultImage() string         { return "test-image" }

var _ docker.DockerClient = (*fakeDockerClient)(nil)

// buildContainerTar packages files into a TAR stream that mimics
// docker's CopyFromContainer output for the given root entry.
func buildContainerTar(entries []tarTestEntry) []byte {
	return buildTar(entries).Bytes()
}

// ─────────────────────────────────────────────────────────────────────────────
// GetFlowFiles handler tests.
// ─────────────────────────────────────────────────────────────────────────────

func TestFlowFileService_GetFlowFilesScenarios(t *testing.T) {
	type seedFile struct {
		dir     string // "uploads", "container/<sub>", "resources/<sub>"
		name    string
		content string
	}

	tests := []struct {
		name              string
		seedFiles         []seedFile
		seedFlow          *struct{ id, userID uint64 }
		flowID            uint64
		uid               uint64
		privs             []string
		flowIDOverride    string
		wantStatus        int
		wantResponsePaths []string
	}{
		{
			name:              "view privilege lists own flow files",
			seedFlow:          &struct{ id, userID uint64 }{1, 1},
			flowID:            1,
			uid:               1,
			privs:             []string{"flow_files.view"},
			seedFiles:         []seedFile{{dir: "uploads", name: "report.txt", content: "r"}},
			wantStatus:        http.StatusOK,
			wantResponsePaths: []string{"uploads/report.txt"},
		},
		{
			name:              "admin lists other user's flow files",
			seedFlow:          &struct{ id, userID uint64 }{1, 2},
			flowID:            1,
			uid:               1,
			privs:             []string{"flow_files.admin"},
			seedFiles:         []seedFile{{dir: "uploads", name: "report.txt", content: "r"}},
			wantStatus:        http.StatusOK,
			wantResponsePaths: []string{"uploads/report.txt"},
		},
		{
			name:       "non-admin cannot list other user's flow",
			seedFlow:   &struct{ id, userID uint64 }{1, 2},
			flowID:     1,
			uid:        1,
			privs:      []string{"flow_files.view"},
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "missing privilege returns forbidden",
			seedFlow:   &struct{ id, userID uint64 }{1, 1},
			flowID:     1,
			uid:        1,
			privs:      []string{"resources.view"},
			wantStatus: http.StatusForbidden,
		},
		{
			name:           "non-numeric flow id returns bad request",
			flowIDOverride: "abc",
			privs:          []string{"flow_files.view"},
			uid:            1,
			wantStatus:     http.StatusBadRequest,
		},
		{
			name:       "missing flow returns not found",
			flowID:     999,
			uid:        1,
			privs:      []string{"flow_files.view"},
			wantStatus: http.StatusNotFound,
		},
		{
			name:     "list aggregates uploads, container, resources",
			seedFlow: &struct{ id, userID uint64 }{1, 1},
			flowID:   1,
			uid:      1,
			privs:    []string{"flow_files.view"},
			seedFiles: []seedFile{
				{dir: "uploads", name: "report.txt", content: "r"},
				{dir: "container/etc", name: "nginx.conf", content: "n"},
				{dir: "resources/creds", name: "p.txt", content: "p"},
			},
			wantStatus: http.StatusOK,
			wantResponsePaths: []string{
				"uploads/report.txt",
				"container/etc",
				"container/etc/nginx.conf",
				"resources/creds",
				"resources/creds/p.txt",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := setupFlowFileServiceTestDB(t)
			dataDir := t.TempDir()
			ss := &flowFileCaptureSubscriptions{}
			svc := NewFlowFileService(db, dataDir, nil, ss)

			if tt.seedFlow != nil {
				seedFlow(t, db, tt.seedFlow.id, tt.seedFlow.userID)
			}
			for _, f := range tt.seedFiles {
				dir := filepath.Join(dataDir, fmt.Sprintf("flow-%d-data", tt.flowID), filepath.FromSlash(f.dir))
				require.NoError(t, os.MkdirAll(dir, 0755))
				require.NoError(t, os.WriteFile(filepath.Join(dir, f.name), []byte(f.content), 0644))
			}

			c, w := newFlowFileTestContext(http.MethodGet, "/flows/1/files/", nil, tt.privs, tt.uid, tt.flowID)
			if tt.flowIDOverride != "" {
				c.Params = gin.Params{{Key: "flowID", Value: tt.flowIDOverride}}
			}
			svc.GetFlowFiles(c)

			require.Equal(t, tt.wantStatus, w.Code)
			if tt.wantStatus == http.StatusOK {
				resp := decodeFlowFilesResponse(t, w)
				paths := make([]string, len(resp.Files))
				for i, f := range resp.Files {
					paths[i] = f.Path
				}
				assert.ElementsMatch(t, tt.wantResponsePaths, paths)
				assert.Equal(t, uint64(len(tt.wantResponsePaths)), resp.Total)
			}
		})
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// UploadFlowFiles handler tests.
// ─────────────────────────────────────────────────────────────────────────────

func TestFlowFileService_UploadFlowFilesScenarios(t *testing.T) {
	tests := []struct {
		name              string
		flowOwner         uint64
		uid               uint64
		flowID            uint64
		privs             []string
		files             []flowFileUploadFile
		fieldName         string
		seedExisting      []string // pre-existing file names in uploads dir
		nonMultipart      bool
		wantStatus        int
		wantResponsePaths []string
		wantPushed        []string // basenames pushed to container
		dockerRunning     bool
	}{
		{
			name:              "upload single file with files field",
			flowOwner:         1,
			uid:               1,
			flowID:            1,
			privs:             []string{"flow_files.upload"},
			files:             []flowFileUploadFile{{name: "report.txt", content: "r"}},
			wantStatus:        http.StatusOK,
			wantResponsePaths: []string{"uploads/report.txt"},
		},
		{
			name:              "upload via singular file field works as fallback",
			flowOwner:         1,
			uid:               1,
			flowID:            1,
			privs:             []string{"flow_files.upload"},
			files:             []flowFileUploadFile{{name: "report.txt", content: "r"}},
			fieldName:         "file",
			wantStatus:        http.StatusOK,
			wantResponsePaths: []string{"uploads/report.txt"},
		},
		{
			name:              "upload multiple files",
			flowOwner:         1,
			uid:               1,
			flowID:            1,
			privs:             []string{"flow_files.upload"},
			files:             []flowFileUploadFile{{name: "a.txt", content: "a"}, {name: "b.txt", content: "b"}},
			wantStatus:        http.StatusOK,
			wantResponsePaths: []string{"uploads/a.txt", "uploads/b.txt"},
		},
		{
			name:              "admin can upload to other user's flow",
			flowOwner:         2,
			uid:               1,
			flowID:            1,
			privs:             []string{"flow_files.admin"},
			files:             []flowFileUploadFile{{name: "report.txt", content: "r"}},
			wantStatus:        http.StatusOK,
			wantResponsePaths: []string{"uploads/report.txt"},
		},
		{
			name:       "view privilege cannot upload",
			flowOwner:  1,
			uid:        1,
			flowID:     1,
			privs:      []string{"flow_files.view"},
			files:      []flowFileUploadFile{{name: "report.txt", content: "r"}},
			wantStatus: http.StatusForbidden,
		},
		{
			name:         "non-multipart body returns bad request",
			flowOwner:    1,
			uid:          1,
			flowID:       1,
			privs:        []string{"flow_files.upload"},
			nonMultipart: true,
			wantStatus:   http.StatusBadRequest,
		},
		{
			name:       "empty multipart returns bad request",
			flowOwner:  1,
			uid:        1,
			flowID:     1,
			privs:      []string{"flow_files.upload"},
			files:      nil,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "invalid filename returns bad request",
			flowOwner:  1,
			uid:        1,
			flowID:     1,
			privs:      []string{"flow_files.upload"},
			files:      []flowFileUploadFile{{name: "bad\nname.txt", content: "x"}},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:         "upload duplicate of existing file returns conflict",
			flowOwner:    1,
			uid:          1,
			flowID:       1,
			privs:        []string{"flow_files.upload"},
			seedExisting: []string{"report.txt"},
			files:        []flowFileUploadFile{{name: "report.txt", content: "new"}},
			wantStatus:   http.StatusConflict,
		},
		{
			name:       "missing flow returns not found",
			uid:        1,
			flowID:     99,
			privs:      []string{"flow_files.upload"},
			files:      []flowFileUploadFile{{name: "report.txt", content: "r"}},
			wantStatus: http.StatusNotFound,
		},
		{
			name:              "successful upload pushes file to running container",
			flowOwner:         1,
			uid:               1,
			flowID:            1,
			privs:             []string{"flow_files.upload"},
			files:             []flowFileUploadFile{{name: "report.txt", content: "r"}},
			dockerRunning:     true,
			wantStatus:        http.StatusOK,
			wantResponsePaths: []string{"uploads/report.txt"},
			wantPushed:        []string{"report.txt"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := setupFlowFileServiceTestDB(t)
			dataDir := t.TempDir()
			ss := &flowFileCaptureSubscriptions{}
			fakeDocker := &fakeDockerClient{running: tt.dockerRunning}
			svc := NewFlowFileService(db, dataDir, fakeDocker, ss)

			if tt.flowOwner != 0 {
				seedFlow(t, db, tt.flowID, tt.flowOwner)
			}
			for _, name := range tt.seedExisting {
				dir := filepath.Join(dataDir, fmt.Sprintf("flow-%d-data", tt.flowID), "uploads")
				require.NoError(t, os.MkdirAll(dir, 0755))
				require.NoError(t, os.WriteFile(filepath.Join(dir, name), []byte("old"), 0644))
			}

			fieldName := tt.fieldName
			if fieldName == "" {
				fieldName = "files"
			}

			var body io.Reader
			var contentType string
			if tt.nonMultipart {
				body = bytes.NewBufferString(`{"hello":"world"}`)
				contentType = "application/json"
			} else {
				buf, ct := flowFileMultipartBody(t, tt.files, fieldName)
				body = buf
				contentType = ct
			}

			c, w := newFlowFileTestContext(http.MethodPost, "/flows/1/files/", body, tt.privs, tt.uid, tt.flowID)
			c.Request.Header.Set("Content-Type", contentType)

			svc.UploadFlowFiles(c)

			require.Equal(t, tt.wantStatus, w.Code)
			if tt.wantStatus == http.StatusOK {
				resp := decodeFlowFilesResponse(t, w)
				paths := make([]string, len(resp.Files))
				for i, f := range resp.Files {
					paths[i] = f.Path
				}
				assert.ElementsMatch(t, tt.wantResponsePaths, paths)

				addedEvents := []string{}
				for _, ev := range ss.snapshot() {
					if ev.channel == "flow" && ev.action == "added" {
						addedEvents = append(addedEvents, ev.path)
					}
				}
				assert.ElementsMatch(t, tt.wantResponsePaths, addedEvents)

				if tt.wantPushed != nil {
					pushed := []string{}
					for _, call := range fakeDocker.copyToCalls {
						pushed = append(pushed, call.dstPath)
					}
					assert.NotEmpty(t, pushed, "expected at least one CopyToContainer call")
				}
			}
		})
	}
}

// TestFlowFileService_UploadFlowFilesPathTraversalSecurity exercises the upload
// handler end-to-end with malicious multipart filenames. The contract is:
//
//   - For each "sanitisable" filename, the handler MUST accept the upload
//     and store the file at flow-{id}-data/uploads/<basename>, with no
//     entries appearing outside that directory anywhere on disk.
//   - For each "unsafe-by-construction" filename, the handler MUST reject
//     the upload (400) and not write anything to the local cache.
//   - At no point may a request escape flow-{id}-data/uploads/, even if the
//     filename embeds traversal segments, absolute paths, NUL bytes, or
//     Windows-style separators.
//
// The directory tree of the entire dataDir is enumerated after each request
// and any unexpected entry fails the test, ensuring the protection chain
// (SanitizeFileName + filepath.Join + LocalEntryExists) is intact.
func TestFlowFileService_UploadFlowFilesPathTraversalSecurity(t *testing.T) {
	tests := []struct {
		name              string
		uploadName        string
		wantStatus        int
		wantStoredAs      string // basename expected on disk under uploads/, empty if rejected
		wantContentInFile string
	}{
		{
			name:              "parent traversal is reduced to basename",
			uploadName:        "../report.txt",
			wantStatus:        http.StatusOK,
			wantStoredAs:      "report.txt",
			wantContentInFile: "payload",
		},
		{
			name:              "deep parent traversal targeting /etc/passwd is reduced to basename",
			uploadName:        "../../etc/passwd",
			wantStatus:        http.StatusOK,
			wantStoredAs:      "passwd",
			wantContentInFile: "payload",
		},
		{
			name:              "absolute unix path is reduced to basename",
			uploadName:        "/etc/shadow",
			wantStatus:        http.StatusOK,
			wantStoredAs:      "shadow",
			wantContentInFile: "payload",
		},
		{
			name:              "windows-style backslash separators are reduced to basename",
			uploadName:        `nested\..\evil.txt`,
			wantStatus:        http.StatusOK,
			wantStoredAs:      "evil.txt",
			wantContentInFile: "payload",
		},
		{
			name:              "double-slash and dotdot mix collapses to basename",
			uploadName:        "..//..//.//attack.bin",
			wantStatus:        http.StatusOK,
			wantStoredAs:      "attack.bin",
			wantContentInFile: "payload",
		},
		{
			name:              "leading dot file is preserved verbatim",
			uploadName:        ".env",
			wantStatus:        http.StatusOK,
			wantStoredAs:      ".env",
			wantContentInFile: "payload",
		},
		{
			name:       "literal parent directory is rejected",
			uploadName: "..",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "literal current directory is rejected",
			uploadName: ".",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "lone slash is rejected",
			uploadName: "/",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "embedded NUL byte is rejected",
			uploadName: "evil\x00.txt",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "newline in filename is rejected",
			uploadName: "evil\nname.txt",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "carriage return in filename is rejected",
			uploadName: "evil\rname.txt",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "DEL control character is rejected",
			uploadName: "evil\x7fname.txt",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "windows reserved colon is rejected",
			uploadName: "con:1.txt",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "wildcard star is rejected",
			uploadName: "evil*.txt",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "redirection chevrons are rejected",
			uploadName: "evil<>name.txt",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "filename longer than MaxFileNameLength is rejected",
			uploadName: strings.Repeat("a", flowfiles.MaxFileNameLength+1) + ".txt",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "blank filename is rejected",
			uploadName: "   ",
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := setupFlowFileServiceTestDB(t)
			dataDir := t.TempDir()
			ss := &flowFileCaptureSubscriptions{}
			svc := NewFlowFileService(db, dataDir, nil, ss)
			seedFlow(t, db, 1, 1)

			body, contentType := flowFileMultipartBody(t,
				[]flowFileUploadFile{{name: tt.uploadName, content: "payload"}},
				"files",
			)
			c, w := newFlowFileTestContext(http.MethodPost, "/flows/1/files/", body,
				[]string{"flow_files.upload"}, 1, 1)
			c.Request.Header.Set("Content-Type", contentType)

			svc.UploadFlowFiles(c)

			require.Equal(t, tt.wantStatus, w.Code, "unexpected status for upload of %q", tt.uploadName)

			// Walk dataDir and capture every entry's relative path so we can
			// prove no file ever escaped flow-1-data/uploads/.
			seenPaths := []string{}
			require.NoError(t, filepath.Walk(dataDir, func(p string, info os.FileInfo, err error) error {
				require.NoError(t, err)
				rel, err := filepath.Rel(dataDir, p)
				require.NoError(t, err)
				if rel == "." {
					return nil
				}
				seenPaths = append(seenPaths, filepath.ToSlash(rel))
				return nil
			}))

			// No entry must live outside flow-1-data/.
			for _, p := range seenPaths {
				assert.True(t,
					p == "flow-1-data" || strings.HasPrefix(p, "flow-1-data/"),
					"file %q escaped the flow data directory", p)
			}

			if tt.wantStatus == http.StatusOK {
				require.NotEmpty(t, tt.wantStoredAs, "test bug: success scenario must declare wantStoredAs")
				expectedRel := "flow-1-data/uploads/" + tt.wantStoredAs
				absExpected := filepath.Join(dataDir, "flow-1-data", "uploads", tt.wantStoredAs)

				// File must exist at the basename location with the original payload.
				data, err := os.ReadFile(absExpected)
				require.NoError(t, err, "expected sanitised file %q on disk", expectedRel)
				assert.Equal(t, tt.wantContentInFile, string(data))

				// Only one regular file must have been created (besides directories).
				regularFiles := []string{}
				for _, p := range seenPaths {
					info, err := os.Lstat(filepath.Join(dataDir, filepath.FromSlash(p)))
					require.NoError(t, err)
					if info.Mode().IsRegular() {
						regularFiles = append(regularFiles, p)
					}
				}
				assert.Equal(t, []string{expectedRel}, regularFiles,
					"the upload must have produced exactly one regular file at the sanitised location")

				// Subscription event must reference the sanitised path, never
				// the raw user-supplied name.
				events := ss.snapshot()
				require.Len(t, events, 1)
				assert.Equal(t, "flow", events[0].channel)
				assert.Equal(t, "added", events[0].action)
				assert.Equal(t, "uploads/"+tt.wantStoredAs, events[0].path)
				assert.NotContains(t, events[0].path, "..", "event path must not contain traversal segments")
				assert.NotContains(t, events[0].path, "\\", "event path must not contain windows separators")
				return
			}

			// Rejected request: nothing should have been published, and
			// uploads/ either doesn't exist or is empty.
			assert.Empty(t, ss.snapshot(), "rejected upload must not emit any subscription events")
			uploadsDir := filepath.Join(dataDir, "flow-1-data", "uploads")
			if entries, err := os.ReadDir(uploadsDir); err == nil {
				for _, e := range entries {
					if !strings.HasPrefix(e.Name(), ".upload-") {
						t.Errorf("unexpected file %q persisted after rejected upload", e.Name())
					}
				}
			} else {
				assert.True(t, os.IsNotExist(err))
			}
		})
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// DeleteFlowFile handler tests.
// ─────────────────────────────────────────────────────────────────────────────

func TestFlowFileService_DeleteFlowFileScenarios(t *testing.T) {
	type seedFile struct {
		relPath string
		content string
		isDir   bool
	}

	tests := []struct {
		name              string
		flowOwner         uint64
		uid               uint64
		flowID            uint64
		privs             []string
		seedFlow          bool
		seedFiles         []seedFile
		queryPath         string
		dockerRunning     bool
		wantStatus        int
		wantDeletedPath   string
		wantRemainingFile string
		wantContainerExec bool
	}{
		{
			name:            "delete uploaded file",
			flowOwner:       1,
			uid:             1,
			flowID:          1,
			privs:           []string{"flow_files.upload"},
			seedFlow:        true,
			seedFiles:       []seedFile{{relPath: "uploads/report.txt", content: "r"}},
			queryPath:       "uploads/report.txt",
			wantStatus:      http.StatusOK,
			wantDeletedPath: "uploads/report.txt",
		},
		{
			name:            "delete container directory recursively",
			flowOwner:       1,
			uid:             1,
			flowID:          1,
			privs:           []string{"flow_files.upload"},
			seedFlow:        true,
			seedFiles:       []seedFile{{relPath: "container/etc", isDir: true}, {relPath: "container/etc/nginx.conf", content: "n"}},
			queryPath:       "container/etc",
			wantStatus:      http.StatusOK,
			wantDeletedPath: "container/etc",
		},
		{
			name:            "delete resources file",
			flowOwner:       1,
			uid:             1,
			flowID:          1,
			privs:           []string{"flow_files.upload"},
			seedFlow:        true,
			seedFiles:       []seedFile{{relPath: "resources/creds/p.txt", content: "p"}},
			queryPath:       "resources/creds/p.txt",
			wantStatus:      http.StatusOK,
			wantDeletedPath: "resources/creds/p.txt",
		},
		{
			name:              "delete uploads file invokes container exec when running",
			flowOwner:         1,
			uid:               1,
			flowID:            1,
			privs:             []string{"flow_files.upload"},
			seedFlow:          true,
			seedFiles:         []seedFile{{relPath: "uploads/report.txt", content: "r"}},
			queryPath:         "uploads/report.txt",
			dockerRunning:     true,
			wantStatus:        http.StatusOK,
			wantDeletedPath:   "uploads/report.txt",
			wantContainerExec: true,
		},
		{
			name:       "view privilege cannot delete",
			flowOwner:  1,
			uid:        1,
			flowID:     1,
			privs:      []string{"flow_files.view"},
			seedFlow:   true,
			queryPath:  "uploads/report.txt",
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "missing flow returns not found",
			uid:        1,
			flowID:     99,
			privs:      []string{"flow_files.upload"},
			queryPath:  "uploads/report.txt",
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "missing path returns bad request",
			flowOwner:  1,
			uid:        1,
			flowID:     1,
			privs:      []string{"flow_files.upload"},
			seedFlow:   true,
			queryPath:  "",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "wrong-prefix path returns bad request",
			flowOwner:  1,
			uid:        1,
			flowID:     1,
			privs:      []string{"flow_files.upload"},
			seedFlow:   true,
			queryPath:  "tmp/x.txt",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "path traversal returns bad request",
			flowOwner:  1,
			uid:        1,
			flowID:     1,
			privs:      []string{"flow_files.upload"},
			seedFlow:   true,
			queryPath:  "uploads/../../etc/passwd",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "non-existent file returns not found",
			flowOwner:  1,
			uid:        1,
			flowID:     1,
			privs:      []string{"flow_files.upload"},
			seedFlow:   true,
			queryPath:  "uploads/missing.txt",
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := setupFlowFileServiceTestDB(t)
			dataDir := t.TempDir()
			ss := &flowFileCaptureSubscriptions{}
			fakeDocker := &fakeDockerClient{running: tt.dockerRunning}
			svc := NewFlowFileService(db, dataDir, fakeDocker, ss)

			if tt.seedFlow {
				owner := tt.flowOwner
				if owner == 0 {
					owner = tt.uid
				}
				seedFlow(t, db, tt.flowID, owner)
			}
			for _, f := range tt.seedFiles {
				abs := filepath.Join(dataDir, fmt.Sprintf("flow-%d-data", tt.flowID), filepath.FromSlash(f.relPath))
				if f.isDir {
					require.NoError(t, os.MkdirAll(abs, 0755))
					continue
				}
				require.NoError(t, os.MkdirAll(filepath.Dir(abs), 0755))
				require.NoError(t, os.WriteFile(abs, []byte(f.content), 0644))
			}

			target := "/flows/1/files/"
			if tt.queryPath != "" {
				target += "?path=" + tt.queryPath
			}
			c, w := newFlowFileTestContext(http.MethodDelete, target, nil, tt.privs, tt.uid, tt.flowID)

			svc.DeleteFlowFile(c)

			require.Equal(t, tt.wantStatus, w.Code)
			if tt.wantStatus == http.StatusOK {
				abs := filepath.Join(dataDir, fmt.Sprintf("flow-%d-data", tt.flowID), filepath.FromSlash(tt.queryPath))
				_, err := os.Lstat(abs)
				assert.True(t, os.IsNotExist(err), "deleted file/dir must be gone from cache")

				deleted := []string{}
				for _, ev := range ss.snapshot() {
					if ev.channel == "flow" && ev.action == "deleted" {
						deleted = append(deleted, ev.path)
					}
				}
				assert.Contains(t, deleted, tt.wantDeletedPath)

				if tt.wantContainerExec {
					assert.NotEmpty(t, fakeDocker.execCommands, "expected docker exec to be invoked")
					assert.Contains(t, fakeDocker.execCommands[0], "rm -rf")
				}
			}
		})
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// DownloadFlowFile handler tests.
// ─────────────────────────────────────────────────────────────────────────────

func TestFlowFileService_DownloadFlowFileScenarios(t *testing.T) {
	tests := []struct {
		name             string
		flowOwner        uint64
		uid              uint64
		flowID           uint64
		privs            []string
		queryPath        string
		setupFile        func(t *testing.T, dataDir string, flowID uint64) string
		wantStatus       int
		wantBody         string
		wantContentType  string
		wantDispContains string
		wantZipEntries   map[string]string
	}{
		{
			name:      "download regular uploaded file",
			flowOwner: 1, uid: 1, flowID: 1,
			privs:     []string{"flow_files.view"},
			queryPath: "uploads/report.txt",
			setupFile: func(t *testing.T, dataDir string, flowID uint64) string {
				dir := filepath.Join(dataDir, fmt.Sprintf("flow-%d-data", flowID), "uploads")
				require.NoError(t, os.MkdirAll(dir, 0755))
				p := filepath.Join(dir, "report.txt")
				require.NoError(t, os.WriteFile(p, []byte("payload"), 0644))
				return p
			},
			wantStatus:       http.StatusOK,
			wantBody:         "payload",
			wantDispContains: "report.txt",
		},
		{
			name:      "download container directory as zip",
			flowOwner: 1, uid: 1, flowID: 1,
			privs:     []string{"flow_files.view"},
			queryPath: "container/etc",
			setupFile: func(t *testing.T, dataDir string, flowID uint64) string {
				dir := filepath.Join(dataDir, fmt.Sprintf("flow-%d-data", flowID), "container", "etc")
				require.NoError(t, os.MkdirAll(filepath.Join(dir, "nginx"), 0755))
				require.NoError(t, os.WriteFile(filepath.Join(dir, "nginx", "nginx.conf"), []byte("nginx"), 0644))
				return dir
			},
			wantStatus:       http.StatusOK,
			wantContentType:  "application/zip",
			wantDispContains: "etc.zip",
			wantZipEntries:   map[string]string{"nginx/nginx.conf": "nginx"},
		},
		{
			name:      "admin downloads other user's file",
			flowOwner: 2, uid: 1, flowID: 1,
			privs:     []string{"flow_files.admin"},
			queryPath: "uploads/report.txt",
			setupFile: func(t *testing.T, dataDir string, flowID uint64) string {
				dir := filepath.Join(dataDir, fmt.Sprintf("flow-%d-data", flowID), "uploads")
				require.NoError(t, os.MkdirAll(dir, 0755))
				p := filepath.Join(dir, "report.txt")
				require.NoError(t, os.WriteFile(p, []byte("admin"), 0644))
				return p
			},
			wantStatus: http.StatusOK,
			wantBody:   "admin",
		},
		{
			name:      "non-admin cannot download other user's file",
			flowOwner: 2, uid: 1, flowID: 1,
			privs:      []string{"flow_files.view"},
			queryPath:  "uploads/report.txt",
			wantStatus: http.StatusNotFound,
		},
		{
			name:      "missing privilege returns forbidden",
			flowOwner: 1, uid: 1, flowID: 1,
			privs:      []string{"resources.view"},
			queryPath:  "uploads/report.txt",
			wantStatus: http.StatusForbidden,
		},
		{
			name:      "empty path returns bad request",
			flowOwner: 1, uid: 1, flowID: 1,
			privs:      []string{"flow_files.view"},
			queryPath:  "",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:      "non-existent path returns not found",
			flowOwner: 1, uid: 1, flowID: 1,
			privs:      []string{"flow_files.view"},
			queryPath:  "uploads/missing.txt",
			wantStatus: http.StatusNotFound,
		},
		{
			name:      "symlink is rejected",
			flowOwner: 1, uid: 1, flowID: 1,
			privs:     []string{"flow_files.view"},
			queryPath: "uploads/link.txt",
			setupFile: func(t *testing.T, dataDir string, flowID uint64) string {
				dir := filepath.Join(dataDir, fmt.Sprintf("flow-%d-data", flowID), "uploads")
				require.NoError(t, os.MkdirAll(dir, 0755))
				target := filepath.Join(dir, "real.txt")
				require.NoError(t, os.WriteFile(target, []byte("x"), 0644))
				link := filepath.Join(dir, "link.txt")
				if err := os.Symlink(target, link); err != nil {
					t.Skipf("symlink unavailable: %v", err)
				}
				return link
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := setupFlowFileServiceTestDB(t)
			dataDir := t.TempDir()
			svc := NewFlowFileService(db, dataDir, nil, nil)

			if tt.flowOwner != 0 {
				seedFlow(t, db, tt.flowID, tt.flowOwner)
			}
			if tt.setupFile != nil {
				tt.setupFile(t, dataDir, tt.flowID)
			}

			target := "/flows/1/files/download"
			if tt.queryPath != "" {
				target += "?path=" + tt.queryPath
			}
			c, w := newFlowFileTestContext(http.MethodGet, target, nil, tt.privs, tt.uid, tt.flowID)

			svc.DownloadFlowFile(c)

			require.Equal(t, tt.wantStatus, w.Code)
			if tt.wantStatus != http.StatusOK {
				return
			}
			if tt.wantContentType != "" {
				assert.Equal(t, tt.wantContentType, w.Header().Get("Content-Type"))
			}
			if tt.wantDispContains != "" {
				assert.Contains(t, w.Header().Get("Content-Disposition"), tt.wantDispContains)
			}
			if tt.wantZipEntries != nil {
				zr, err := zip.NewReader(bytes.NewReader(w.Body.Bytes()), int64(w.Body.Len()))
				require.NoError(t, err)
				got := map[string]string{}
				for _, f := range zr.File {
					rc, err := f.Open()
					require.NoError(t, err)
					data, err := io.ReadAll(rc)
					rc.Close()
					require.NoError(t, err)
					got[f.Name] = string(data)
				}
				assert.Equal(t, tt.wantZipEntries, got)
			} else if tt.wantBody != "" {
				assert.Equal(t, tt.wantBody, w.Body.String())
			}
		})
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// PullFlowFiles handler tests.
// ─────────────────────────────────────────────────────────────────────────────

func TestFlowFileService_PullFlowFilesScenarios(t *testing.T) {
	tests := []struct {
		name             string
		flowOwner        uint64
		uid              uint64
		flowID           uint64
		privs            []string
		body             any // marshaled via json
		rawBody          string
		seedExisting     bool
		dockerSetup      func(*fakeDockerClient)
		dockerNil        bool
		wantStatus       int
		wantResponsePath string
		wantEventChannel string // "added" | "updated" | ""
	}{
		{
			name:      "pull new file",
			flowOwner: 1, uid: 1, flowID: 1,
			privs: []string{"flow_files.upload", "containers.view"},
			body:  models.PullFlowFilesRequest{Path: "/etc/nginx.conf"},
			dockerSetup: func(d *fakeDockerClient) {
				d.running = true
				d.copyFromBody = buildContainerTar([]tarTestEntry{
					{name: "nginx.conf", typeflag: tar.TypeReg, content: "nginx-config"},
				})
			},
			wantStatus:       http.StatusOK,
			wantResponsePath: "container/etc/nginx.conf",
			wantEventChannel: "added",
		},
		{
			name:      "pull existing without force conflicts",
			flowOwner: 1, uid: 1, flowID: 1,
			privs:        []string{"flow_files.upload", "containers.view"},
			body:         models.PullFlowFilesRequest{Path: "/etc/nginx.conf"},
			seedExisting: true,
			wantStatus:   http.StatusConflict,
		},
		{
			name:      "pull existing with force overwrites",
			flowOwner: 1, uid: 1, flowID: 1,
			privs:        []string{"flow_files.upload", "containers.view"},
			body:         models.PullFlowFilesRequest{Path: "/etc/nginx.conf", Force: true},
			seedExisting: true,
			dockerSetup: func(d *fakeDockerClient) {
				d.running = true
				d.copyFromBody = buildContainerTar([]tarTestEntry{
					{name: "nginx.conf", typeflag: tar.TypeReg, content: "new"},
				})
			},
			wantStatus:       http.StatusOK,
			wantResponsePath: "container/etc/nginx.conf",
			wantEventChannel: "updated",
		},
		{
			name:      "missing flow_files privilege",
			flowOwner: 1, uid: 1, flowID: 1,
			privs:      []string{"flow_files.view", "containers.view"},
			body:       models.PullFlowFilesRequest{Path: "/etc/nginx.conf"},
			wantStatus: http.StatusForbidden,
		},
		{
			name:      "missing containers.view privilege",
			flowOwner: 1, uid: 1, flowID: 1,
			privs:      []string{"flow_files.upload"},
			body:       models.PullFlowFilesRequest{Path: "/etc/nginx.conf"},
			wantStatus: http.StatusForbidden,
		},
		{
			name:      "containers.admin privilege is sufficient",
			flowOwner: 1, uid: 1, flowID: 1,
			privs: []string{"flow_files.upload", "containers.admin"},
			body:  models.PullFlowFilesRequest{Path: "/etc/nginx.conf"},
			dockerSetup: func(d *fakeDockerClient) {
				d.running = true
				d.copyFromBody = buildContainerTar([]tarTestEntry{
					{name: "nginx.conf", typeflag: tar.TypeReg, content: "data"},
				})
			},
			wantStatus:       http.StatusOK,
			wantResponsePath: "container/etc/nginx.conf",
			wantEventChannel: "added",
		},
		{
			name:      "malformed json returns bad request",
			flowOwner: 1, uid: 1, flowID: 1,
			privs:      []string{"flow_files.upload", "containers.view"},
			rawBody:    `{not json`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:      "empty container path returns bad request",
			flowOwner: 1, uid: 1, flowID: 1,
			privs:      []string{"flow_files.upload", "containers.view"},
			body:       models.PullFlowFilesRequest{Path: "  "},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:      "invalid container path returns bad request",
			flowOwner: 1, uid: 1, flowID: 1,
			privs:      []string{"flow_files.upload", "containers.view"},
			body:       models.PullFlowFilesRequest{Path: "/"},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:      "docker client not configured returns internal",
			flowOwner: 1, uid: 1, flowID: 1,
			privs:      []string{"flow_files.upload", "containers.view"},
			body:       models.PullFlowFilesRequest{Path: "/etc/nginx.conf"},
			dockerNil:  true,
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:      "container not running returns 400",
			flowOwner: 1, uid: 1, flowID: 1,
			privs: []string{"flow_files.upload", "containers.view"},
			body:  models.PullFlowFilesRequest{Path: "/etc/nginx.conf"},
			dockerSetup: func(d *fakeDockerClient) {
				d.running = false
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:      "running check error returns internal",
			flowOwner: 1, uid: 1, flowID: 1,
			privs: []string{"flow_files.upload", "containers.view"},
			body:  models.PullFlowFilesRequest{Path: "/etc/nginx.conf"},
			dockerSetup: func(d *fakeDockerClient) {
				d.runningErr = fmt.Errorf("docker daemon down")
			},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:      "copy from container error returns internal",
			flowOwner: 1, uid: 1, flowID: 1,
			privs: []string{"flow_files.upload", "containers.view"},
			body:  models.PullFlowFilesRequest{Path: "/etc/nginx.conf"},
			dockerSetup: func(d *fakeDockerClient) {
				d.running = true
				d.copyFromErr = fmt.Errorf("not found")
			},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:      "archive missing expected entry returns internal",
			flowOwner: 1, uid: 1, flowID: 1,
			privs: []string{"flow_files.upload", "containers.view"},
			body:  models.PullFlowFilesRequest{Path: "/etc/nginx.conf"},
			dockerSetup: func(d *fakeDockerClient) {
				d.running = true
				d.copyFromBody = buildContainerTar([]tarTestEntry{
					{name: "other.conf", typeflag: tar.TypeReg, content: "x"},
				})
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := setupFlowFileServiceTestDB(t)
			dataDir := t.TempDir()
			ss := &flowFileCaptureSubscriptions{}

			var dockerClient docker.DockerClient
			fakeDocker := &fakeDockerClient{}
			if tt.dockerSetup != nil {
				tt.dockerSetup(fakeDocker)
			}
			if !tt.dockerNil {
				dockerClient = fakeDocker
			}
			svc := NewFlowFileService(db, dataDir, dockerClient, ss)

			seedFlow(t, db, tt.flowID, tt.flowOwner)

			if tt.seedExisting {
				dir := filepath.Join(dataDir, fmt.Sprintf("flow-%d-data", tt.flowID), "container", "etc")
				require.NoError(t, os.MkdirAll(dir, 0755))
				require.NoError(t, os.WriteFile(filepath.Join(dir, "nginx.conf"), []byte("old"), 0644))
			}

			var bodyReader io.Reader
			if tt.rawBody != "" {
				bodyReader = bytes.NewBufferString(tt.rawBody)
			} else {
				payload, err := json.Marshal(tt.body)
				require.NoError(t, err)
				bodyReader = bytes.NewBuffer(payload)
			}
			c, w := newFlowFileTestContext(http.MethodPost, "/flows/1/files/pull", bodyReader, tt.privs, tt.uid, tt.flowID)
			c.Request.Header.Set("Content-Type", "application/json")

			svc.PullFlowFiles(c)

			require.Equal(t, tt.wantStatus, w.Code)
			if tt.wantStatus != http.StatusOK {
				return
			}
			resp := decodeFlowFilesResponse(t, w)
			require.Len(t, resp.Files, 1)
			assert.Equal(t, tt.wantResponsePath, resp.Files[0].Path)
			if tt.wantEventChannel != "" {
				events := ss.snapshot()
				require.Len(t, events, 1)
				assert.Equal(t, "flow", events[0].channel)
				assert.Equal(t, tt.wantEventChannel, events[0].action)
				assert.Equal(t, tt.wantResponsePath, events[0].path)
			}
		})
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// GetFlowContainerFiles handler tests.
// ─────────────────────────────────────────────────────────────────────────────

func TestFlowFileService_GetFlowContainerFilesScenarios(t *testing.T) {
	tests := []struct {
		name           string
		flowOwner      uint64
		uid            uint64
		flowID         uint64
		privs          []string
		queryPath      string
		dockerSetup    func(*fakeDockerClient)
		dockerNil      bool
		wantStatus     int
		wantPathInResp string
		wantFileNames  []string
	}{
		{
			name:      "default path lists work directory",
			flowOwner: 1, uid: 1, flowID: 1,
			privs: []string{"flow_files.view", "containers.view"},
			dockerSetup: func(d *fakeDockerClient) {
				d.running = true
				d.statPath = container.PathStat{Mode: os.ModeDir | 0755}
				d.listDir = []container.PathStat{
					{Name: "uploads", Mode: os.ModeDir | 0755},
					{Name: "report.txt", Size: 7, Mode: 0644},
				}
			},
			wantStatus:     http.StatusOK,
			wantPathInResp: "/work",
			wantFileNames:  []string{"report.txt", "uploads"},
		},
		{
			name:      "custom path is honoured",
			flowOwner: 1, uid: 1, flowID: 1,
			privs:     []string{"flow_files.view", "containers.view"},
			queryPath: "/etc",
			dockerSetup: func(d *fakeDockerClient) {
				d.running = true
				d.statPath = container.PathStat{Mode: os.ModeDir | 0755}
				d.listDir = []container.PathStat{
					{Name: "passwd", Size: 1, Mode: 0644},
				}
			},
			wantStatus:     http.StatusOK,
			wantPathInResp: "/etc",
			wantFileNames:  []string{"passwd"},
		},
		{
			name:      "regular file path returns single file entry",
			flowOwner: 1, uid: 1, flowID: 1,
			privs:     []string{"flow_files.view", "containers.view"},
			queryPath: "/etc/passwd",
			dockerSetup: func(d *fakeDockerClient) {
				d.running = true
				d.statPath = container.PathStat{Name: "passwd", Size: 5, Mode: 0644}
			},
			wantStatus:     http.StatusOK,
			wantPathInResp: "/etc/passwd",
			wantFileNames:  []string{"passwd"},
		},
		{
			name:      "missing flow_files privilege",
			flowOwner: 1, uid: 1, flowID: 1,
			privs:      []string{"containers.view"},
			wantStatus: http.StatusForbidden,
		},
		{
			name:      "missing containers privilege",
			flowOwner: 1, uid: 1, flowID: 1,
			privs:      []string{"flow_files.view"},
			wantStatus: http.StatusForbidden,
		},
		{
			name:      "docker client not configured returns internal",
			flowOwner: 1, uid: 1, flowID: 1,
			privs:      []string{"flow_files.view", "containers.view"},
			dockerNil:  true,
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:      "container not running returns 400",
			flowOwner: 1, uid: 1, flowID: 1,
			privs: []string{"flow_files.view", "containers.view"},
			dockerSetup: func(d *fakeDockerClient) {
				d.running = false
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:      "stat path error returns internal",
			flowOwner: 1, uid: 1, flowID: 1,
			privs: []string{"flow_files.view", "containers.view"},
			dockerSetup: func(d *fakeDockerClient) {
				d.running = true
				d.statPathErr = fmt.Errorf("not found")
			},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:      "list dir error returns internal",
			flowOwner: 1, uid: 1, flowID: 1,
			privs: []string{"flow_files.view", "containers.view"},
			dockerSetup: func(d *fakeDockerClient) {
				d.running = true
				d.statPath = container.PathStat{Mode: os.ModeDir | 0755}
				d.listDirErr = fmt.Errorf("permission denied")
			},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:       "missing flow returns not found",
			uid:        1,
			flowID:     99,
			privs:      []string{"flow_files.view", "containers.view"},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := setupFlowFileServiceTestDB(t)
			dataDir := t.TempDir()

			var dockerClient docker.DockerClient
			fakeDocker := &fakeDockerClient{}
			if tt.dockerSetup != nil {
				tt.dockerSetup(fakeDocker)
			}
			if !tt.dockerNil {
				dockerClient = fakeDocker
			}
			svc := NewFlowFileService(db, dataDir, dockerClient, nil)

			if tt.flowOwner != 0 {
				seedFlow(t, db, tt.flowID, tt.flowOwner)
			}

			target := "/flows/1/files/container"
			if tt.queryPath != "" {
				target += "?path=" + tt.queryPath
			}
			c, w := newFlowFileTestContext(http.MethodGet, target, nil, tt.privs, tt.uid, tt.flowID)

			svc.GetFlowContainerFiles(c)

			require.Equal(t, tt.wantStatus, w.Code)
			if tt.wantStatus != http.StatusOK {
				return
			}
			resp := decodeContainerFilesResponse(t, w)
			assert.Equal(t, tt.wantPathInResp, resp.Path)
			names := make([]string, len(resp.Files))
			for i, f := range resp.Files {
				names[i] = f.Name
			}
			assert.Equal(t, tt.wantFileNames, names)
		})
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// AddResourcesToFlow handler tests.
// ─────────────────────────────────────────────────────────────────────────────

func TestFlowFileService_AddResourcesToFlowScenarios(t *testing.T) {
	type seedRes struct {
		userID  uint64
		path    string
		content string
	}

	tests := []struct {
		name             string
		flowOwner        uint64
		uid              uint64
		flowID           uint64
		privs            []string
		seedResources    []seedRes
		body             any
		rawBody          string
		seedExisting     map[string]string // relative path under flow resources/ -> content
		wantStatus       int
		wantFileExists   string // relative to flow resources/
		wantContent      string
		wantEventChannel string // "added" or "updated"
	}{
		{
			name:      "copy single resource",
			flowOwner: 1, uid: 1, flowID: 1,
			privs:         []string{"flow_files.upload", "resources.view"},
			seedResources: []seedRes{{path: "creds/p.txt", content: "secret"}},
			body: map[string]any{
				"ids": []string{"1"},
			},
			wantStatus:       http.StatusOK,
			wantFileExists:   "creds/p.txt",
			wantContent:      "secret",
			wantEventChannel: "added",
		},
		{
			name:      "skip existing resource without force",
			flowOwner: 1, uid: 1, flowID: 1,
			privs:         []string{"flow_files.upload", "resources.view"},
			seedResources: []seedRes{{path: "creds/p.txt", content: "secret"}},
			seedExisting:  map[string]string{"creds/p.txt": "old"},
			body: map[string]any{
				"ids": []string{"1"},
			},
			wantStatus:     http.StatusOK,
			wantFileExists: "creds/p.txt",
			wantContent:    "old",
		},
		{
			name:      "force overwrite emits updated event",
			flowOwner: 1, uid: 1, flowID: 1,
			privs:         []string{"flow_files.upload", "resources.view"},
			seedResources: []seedRes{{path: "creds/p.txt", content: "new"}},
			seedExisting:  map[string]string{"creds/p.txt": "old"},
			body: map[string]any{
				"ids":   []string{"1"},
				"force": true,
			},
			wantStatus:       http.StatusOK,
			wantFileExists:   "creds/p.txt",
			wantContent:      "new",
			wantEventChannel: "updated",
		},
		{
			name:      "admin can copy other user's resource",
			flowOwner: 1, uid: 1, flowID: 1,
			privs:         []string{"flow_files.admin", "resources.admin"},
			seedResources: []seedRes{{userID: 2, path: "creds/p.txt", content: "secret"}},
			body: map[string]any{
				"ids": []string{"1"},
			},
			wantStatus:       http.StatusOK,
			wantFileExists:   "creds/p.txt",
			wantContent:      "secret",
			wantEventChannel: "added",
		},
		{
			name:      "non-admin cannot copy another user's resource",
			flowOwner: 1, uid: 1, flowID: 1,
			privs:         []string{"flow_files.upload", "resources.view"},
			seedResources: []seedRes{{userID: 2, path: "creds/p.txt", content: "secret"}},
			body: map[string]any{
				"ids": []string{"1"},
			},
			wantStatus: http.StatusForbidden,
		},
		{
			name:      "missing flow_files.upload privilege",
			flowOwner: 1, uid: 1, flowID: 1,
			privs: []string{"resources.view"},
			body: map[string]any{
				"ids": []string{"1"},
			},
			wantStatus: http.StatusForbidden,
		},
		{
			name:      "missing resources.view privilege",
			flowOwner: 1, uid: 1, flowID: 1,
			privs: []string{"flow_files.upload"},
			body: map[string]any{
				"ids": []string{"1"},
			},
			wantStatus: http.StatusForbidden,
		},
		{
			name:      "malformed json returns bad request",
			flowOwner: 1, uid: 1, flowID: 1,
			privs:      []string{"flow_files.upload", "resources.view"},
			rawBody:    `{not json`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:      "empty ids returns bad request",
			flowOwner: 1, uid: 1, flowID: 1,
			privs: []string{"flow_files.upload", "resources.view"},
			body: map[string]any{
				"ids": []string{},
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:      "non-existent resource id returns bad request",
			flowOwner: 1, uid: 1, flowID: 1,
			privs: []string{"flow_files.upload", "resources.view"},
			body: map[string]any{
				"ids": []string{"999"},
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "missing flow returns not found",
			uid:        1,
			flowID:     99,
			privs:      []string{"flow_files.upload", "resources.view"},
			body:       map[string]any{"ids": []string{"1"}},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := setupFlowFileServiceTestDB(t)
			dataDir := t.TempDir()
			ss := &flowFileCaptureSubscriptions{}
			svc := NewFlowFileService(db, dataDir, nil, ss)

			require.NoError(t, resources.EnsureResourcesDir(dataDir))
			if tt.flowOwner != 0 {
				seedFlow(t, db, tt.flowID, tt.flowOwner)
			}
			for _, r := range tt.seedResources {
				userID := r.userID
				if userID == 0 {
					userID = tt.uid
				}
				hash := md5HexForFlowFiles(r.content)
				blobPath := resources.BlobPath(dataDir, hash)
				require.NoError(t, os.MkdirAll(filepath.Dir(blobPath), 0755))
				require.NoError(t, os.WriteFile(blobPath, []byte(r.content), 0644))
				seedUserResource(t, db, models.UserResource{
					UserID: userID,
					Hash:   hash,
					Name:   filepath.Base(r.path),
					Path:   r.path,
					Size:   int64(len(r.content)),
				})
			}
			for relPath, content := range tt.seedExisting {
				abs := filepath.Join(dataDir, fmt.Sprintf("flow-%d-data", tt.flowID), "resources", filepath.FromSlash(relPath))
				require.NoError(t, os.MkdirAll(filepath.Dir(abs), 0755))
				require.NoError(t, os.WriteFile(abs, []byte(content), 0644))
			}

			var bodyReader io.Reader
			if tt.rawBody != "" {
				bodyReader = bytes.NewBufferString(tt.rawBody)
			} else {
				payload, err := json.Marshal(tt.body)
				require.NoError(t, err)
				bodyReader = bytes.NewBuffer(payload)
			}
			c, w := newFlowFileTestContext(http.MethodPost, "/flows/1/files/resources", bodyReader, tt.privs, tt.uid, tt.flowID)
			c.Request.Header.Set("Content-Type", "application/json")

			svc.AddResourcesToFlow(c)

			require.Equal(t, tt.wantStatus, w.Code)
			if tt.wantStatus != http.StatusOK {
				return
			}
			abs := filepath.Join(dataDir, fmt.Sprintf("flow-%d-data", tt.flowID), "resources", filepath.FromSlash(tt.wantFileExists))
			data, err := os.ReadFile(abs)
			require.NoError(t, err)
			assert.Equal(t, tt.wantContent, string(data))

			if tt.wantEventChannel != "" {
				matched := false
				for _, ev := range ss.snapshot() {
					if ev.channel == "flow" && ev.action == tt.wantEventChannel {
						matched = true
						break
					}
				}
				assert.True(t, matched, "expected event %q not emitted", tt.wantEventChannel)
			} else {
				for _, ev := range ss.snapshot() {
					assert.NotEqual(t, "flow", ev.channel, "no flow events expected for skipped copy")
				}
			}
		})
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// AddResourceFromFlow handler tests.
// ─────────────────────────────────────────────────────────────────────────────

func TestFlowFileService_AddResourceFromFlowScenarios(t *testing.T) {
	type sourceFile struct {
		relPath string
		content string
	}

	tests := []struct {
		name             string
		flowOwner        uint64
		uid              uint64
		flowID           uint64
		privs            []string
		sourceFiles      []sourceFile
		existingResource *models.UserResource
		body             any
		rawBody          string
		wantStatus       int
		wantResourcePath string
		wantEventChannel string // "added" | "updated"
	}{
		{
			name:      "promote uploaded file to new resource",
			flowOwner: 1, uid: 1, flowID: 1,
			privs:       []string{"resources.upload", "flow_files.view"},
			sourceFiles: []sourceFile{{relPath: "uploads/report.txt", content: "payload"}},
			body: map[string]any{
				"sourcePath":  "uploads/report.txt",
				"destination": "promoted/report.txt",
			},
			wantStatus:       http.StatusOK,
			wantResourcePath: "promoted/report.txt",
			wantEventChannel: "added",
		},
		{
			name:      "promote container file to user resources",
			flowOwner: 1, uid: 1, flowID: 1,
			privs:       []string{"resources.upload", "flow_files.view"},
			sourceFiles: []sourceFile{{relPath: "container/etc/nginx.conf", content: "nginx"}},
			body: map[string]any{
				"sourcePath":  "container/etc/nginx.conf",
				"destination": "configs/nginx.conf",
			},
			wantStatus:       http.StatusOK,
			wantResourcePath: "configs/nginx.conf",
			wantEventChannel: "added",
		},
		{
			name:      "force overwrite existing resource",
			flowOwner: 1, uid: 1, flowID: 1,
			privs:       []string{"resources.upload", "flow_files.view"},
			sourceFiles: []sourceFile{{relPath: "uploads/report.txt", content: "new"}},
			existingResource: &models.UserResource{
				Hash: md5HexForFlowFiles("old"), Name: "report.txt", Path: "promoted/report.txt", Size: 3,
			},
			body: map[string]any{
				"sourcePath":  "uploads/report.txt",
				"destination": "promoted/report.txt",
				"force":       true,
			},
			wantStatus:       http.StatusOK,
			wantResourcePath: "promoted/report.txt",
			wantEventChannel: "updated",
		},
		{
			name:      "existing resource without force returns conflict",
			flowOwner: 1, uid: 1, flowID: 1,
			privs:       []string{"resources.upload", "flow_files.view"},
			sourceFiles: []sourceFile{{relPath: "uploads/report.txt", content: "new"}},
			existingResource: &models.UserResource{
				Hash: md5HexForFlowFiles("old"), Name: "report.txt", Path: "promoted/report.txt", Size: 3,
			},
			body: map[string]any{
				"sourcePath":  "uploads/report.txt",
				"destination": "promoted/report.txt",
			},
			wantStatus: http.StatusConflict,
		},
		{
			name:      "resources admin still requires flow_files privilege to read flow",
			flowOwner: 1, uid: 1, flowID: 1,
			privs:       []string{"resources.admin"},
			sourceFiles: []sourceFile{{relPath: "uploads/report.txt", content: "payload"}},
			body: map[string]any{
				"sourcePath":  "uploads/report.txt",
				"destination": "promoted/report.txt",
			},
			wantStatus: http.StatusForbidden,
		},
		{
			name:      "resources admin with flow_files.admin can promote any flow file",
			flowOwner: 2, uid: 1, flowID: 1,
			privs:       []string{"resources.admin", "flow_files.admin"},
			sourceFiles: []sourceFile{{relPath: "uploads/report.txt", content: "payload"}},
			body: map[string]any{
				"sourcePath":  "uploads/report.txt",
				"destination": "promoted/report.txt",
			},
			wantStatus:       http.StatusOK,
			wantResourcePath: "promoted/report.txt",
			wantEventChannel: "added",
		},
		{
			name:      "missing resources.upload returns forbidden",
			flowOwner: 1, uid: 1, flowID: 1,
			privs: []string{"flow_files.view"},
			body: map[string]any{
				"sourcePath": "uploads/report.txt", "destination": "promoted/report.txt",
			},
			wantStatus: http.StatusForbidden,
		},
		{
			name:      "missing flow_files.view returns forbidden",
			flowOwner: 1, uid: 1, flowID: 1,
			privs: []string{"resources.upload"},
			body: map[string]any{
				"sourcePath": "uploads/report.txt", "destination": "promoted/report.txt",
			},
			wantStatus: http.StatusForbidden,
		},
		{
			name: "missing flow returns not found",
			uid:  1, flowID: 99,
			privs: []string{"resources.upload", "flow_files.view"},
			body: map[string]any{
				"sourcePath": "uploads/report.txt", "destination": "promoted/report.txt",
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:      "malformed json returns bad request",
			flowOwner: 1, uid: 1, flowID: 1,
			privs:      []string{"resources.upload", "flow_files.view"},
			rawBody:    `{not json`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:      "missing required fields returns bad request",
			flowOwner: 1, uid: 1, flowID: 1,
			privs:      []string{"resources.upload", "flow_files.view"},
			body:       map[string]any{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:      "invalid source path returns bad data",
			flowOwner: 1, uid: 1, flowID: 1,
			privs: []string{"resources.upload", "flow_files.view"},
			body: map[string]any{
				"sourcePath": "tmp/evil.txt", "destination": "promoted/report.txt",
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:      "missing source file returns not found",
			flowOwner: 1, uid: 1, flowID: 1,
			privs: []string{"resources.upload", "flow_files.view"},
			body: map[string]any{
				"sourcePath": "uploads/missing.txt", "destination": "promoted/report.txt",
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:      "invalid destination path returns bad data",
			flowOwner: 1, uid: 1, flowID: 1,
			privs:       []string{"resources.upload", "flow_files.view"},
			sourceFiles: []sourceFile{{relPath: "uploads/report.txt", content: "x"}},
			body: map[string]any{
				"sourcePath": "uploads/report.txt", "destination": "../escape.txt",
			},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := setupFlowFileServiceTestDB(t)
			dataDir := t.TempDir()
			ss := &flowFileCaptureSubscriptions{}
			svc := NewFlowFileService(db, dataDir, nil, ss)

			require.NoError(t, resources.EnsureResourcesDir(dataDir))
			if tt.flowOwner != 0 {
				seedFlow(t, db, tt.flowID, tt.flowOwner)
			}
			for _, sf := range tt.sourceFiles {
				abs := filepath.Join(dataDir, fmt.Sprintf("flow-%d-data", tt.flowID), filepath.FromSlash(sf.relPath))
				require.NoError(t, os.MkdirAll(filepath.Dir(abs), 0755))
				require.NoError(t, os.WriteFile(abs, []byte(sf.content), 0644))
			}
			if tt.existingResource != nil {
				existing := *tt.existingResource
				existing.UserID = tt.uid
				if existing.Hash != "" {
					blobPath := resources.BlobPath(dataDir, existing.Hash)
					require.NoError(t, os.MkdirAll(filepath.Dir(blobPath), 0755))
					require.NoError(t, os.WriteFile(blobPath, []byte("old"), 0644))
				}
				seedUserResource(t, db, existing)
			}

			var bodyReader io.Reader
			if tt.rawBody != "" {
				bodyReader = bytes.NewBufferString(tt.rawBody)
			} else {
				payload, err := json.Marshal(tt.body)
				require.NoError(t, err)
				bodyReader = bytes.NewBuffer(payload)
			}
			c, w := newFlowFileTestContext(http.MethodPost, "/flows/1/files/to-resources", bodyReader, tt.privs, tt.uid, tt.flowID)
			c.Request.Header.Set("Content-Type", "application/json")

			svc.AddResourceFromFlow(c)

			require.Equal(t, tt.wantStatus, w.Code)
			if tt.wantStatus != http.StatusOK {
				return
			}
			entry := decodeResourceEntryResponse(t, w)
			assert.Equal(t, tt.wantResourcePath, entry.Path)
			assert.False(t, entry.IsDir)

			var rec models.UserResource
			require.NoError(t, db.Where("user_id = ? AND path = ?", tt.uid, tt.wantResourcePath).First(&rec).Error)
			assert.NotEmpty(t, rec.Hash)

			if tt.wantEventChannel != "" {
				events := ss.snapshot()
				matched := false
				for _, ev := range events {
					if ev.channel == "resource" && ev.action == tt.wantEventChannel {
						matched = true
						break
					}
				}
				assert.True(t, matched, "expected resource event %q not emitted", tt.wantEventChannel)
			}
		})
	}
}
