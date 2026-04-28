package services

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"pentagi/pkg/flowfiles"
	"pentagi/pkg/server/models"

	"github.com/docker/docker/api/types/container"
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
