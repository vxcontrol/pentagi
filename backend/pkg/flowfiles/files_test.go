package flowfiles

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"io"
	"mime/multipart"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type tarTestEntry struct {
	name     string
	typeflag byte
	content  string
	linkname string
}

func buildTar(t *testing.T, entries []tarTestEntry) *bytes.Buffer {
	t.Helper()

	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)
	for _, entry := range entries {
		hdr := &tar.Header{
			Name:     entry.name,
			Typeflag: entry.typeflag,
			Mode:     0644,
			Size:     int64(len(entry.content)),
			Linkname: entry.linkname,
		}
		if entry.typeflag == tar.TypeDir {
			hdr.Mode = 0755
		}
		require.NoError(t, tw.WriteHeader(hdr))
		if len(entry.content) > 0 {
			_, _ = tw.Write([]byte(entry.content))
		}
	}
	require.NoError(t, tw.Close())
	return &buf
}

func TestFlowDirs(t *testing.T) {
	assert.Equal(t, "/data/flow-42-data", FlowDataDir("/data", 42))
	assert.Equal(t, "/data/flow-42-data/uploads", FlowUploadsDir("/data", 42))
	assert.Equal(t, "/data/flow-42-data/container", FlowContainerDir("/data", 42))
}

func TestSanitizeFileName(t *testing.T) {
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
		{"rejects too long", string(bytes.Repeat([]byte("a"), MaxFileNameLength+1)), "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SanitizeFileName(tt.input)
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
		{"normalizes windows separators", `etc\nginx\nginx.conf`, "etc/nginx/nginx.conf", false},
		{"rejects empty", "   ", "", true},
		{"rejects root", "/", "", true},
		{"rejects bad component", "/etc/bad\nname", "", true},
		{"rejects unsupported component", `/etc/bad"name`, "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SanitizeContainerCachePath(tt.input)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestResolveCachedPath(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{"uploads file", "uploads/report.txt", "/data/flow-1-data/uploads/report.txt", false},
		{"container file", "container/etc/nginx/nginx.conf", "/data/flow-1-data/container/etc/nginx/nginx.conf", false},
		{"container windows separators", `container\etc\nginx.conf`, "/data/flow-1-data/container/etc/nginx.conf", false},
		{"empty path", "", "", true},
		{"wrong prefix", "tmp/evil.sh", "", true},
		{"absolute path", "/etc/passwd", "", true},
		{"path traversal", "uploads/../../etc/passwd", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ResolveCachedPath("/data", 1, tt.input)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestListDirEntries(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "a.txt"), []byte("a"), 0644))
	require.NoError(t, os.Mkdir(filepath.Join(dir, "sub"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".upload-temp"), []byte("tmp"), 0644))
	require.NoError(t, os.Mkdir(filepath.Join(dir, ".pull-temp"), 0755))
	if err := os.Symlink(filepath.Join(dir, "a.txt"), filepath.Join(dir, "link.txt")); err != nil {
		t.Skipf("symlink creation not available: %v", err)
	}

	entries, err := ListDirEntries(dir, UploadsDirName)
	require.NoError(t, err)
	require.Len(t, entries, 2)

	paths := make([]string, len(entries))
	for i, entry := range entries {
		paths[i] = entry.Path
	}
	assert.Contains(t, paths, "uploads/a.txt")
	assert.Contains(t, paths, "uploads/sub")
	assert.NotContains(t, paths, "uploads/.upload-temp")
	assert.NotContains(t, paths, "uploads/.pull-temp")
	assert.NotContains(t, paths, "uploads/link.txt")
}

func TestListDirEntriesMissingDir(t *testing.T) {
	entries, err := ListDirEntries(filepath.Join(t.TempDir(), "missing"), UploadsDirName)
	require.NoError(t, err)
	assert.Empty(t, entries)

	entries, err = ListDirEntriesRecursive(filepath.Join(t.TempDir(), "missing"), ContainerDirName)
	require.NoError(t, err)
	assert.Empty(t, entries)
}

func TestListDirEntriesRecursivePreservesNestedPaths(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "etc", "nginx", "conf"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "etc", "nginx", "nginx.conf"), []byte("nginx"), 0644))
	require.NoError(t, os.Mkdir(filepath.Join(dir, ".pull-temp"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".pull-temp", "tmp.txt"), []byte("tmp"), 0644))

	entries, err := ListDirEntriesRecursive(dir, ContainerDirName)
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

func TestListBothSources(t *testing.T) {
	dataDir := t.TempDir()
	uploadsDir := FlowUploadsDir(dataDir, 7)
	containerDir := FlowContainerDir(dataDir, 7)
	require.NoError(t, os.MkdirAll(uploadsDir, 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(containerDir, "etc", "nginx", "conf"), 0755))

	require.NoError(t, os.WriteFile(filepath.Join(uploadsDir, "wordlist.txt"), []byte("words"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(containerDir, "etc", "nginx", "nginx.conf"), []byte("nginx"), 0644))

	files, err := List(dataDir, 7)
	require.NoError(t, err)
	assert.Equal(t, uint64(len(files.Files)), files.Total)

	paths := make([]string, len(files.Files))
	for i, file := range files.Files {
		paths[i] = file.Path
	}
	assert.Contains(t, paths, "uploads/wordlist.txt")
	assert.Contains(t, paths, "container/etc/nginx/conf")
	assert.Contains(t, paths, "container/etc/nginx/nginx.conf")
}

func TestLocalEntryExistsAndRegularFileInfo(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "f.txt")

	exists, err := LocalEntryExists(filePath)
	require.NoError(t, err)
	assert.False(t, exists)

	require.NoError(t, os.WriteFile(filePath, []byte("x"), 0644))

	exists, err = LocalEntryExists(filePath)
	require.NoError(t, err)
	assert.True(t, exists)

	info, err := RegularFileInfo(filePath)
	require.NoError(t, err)
	assert.Equal(t, "f.txt", info.Name())

	_, err = RegularFileInfo(dir)
	require.Error(t, err)
}

func TestSaveUploadedFileToTemp(t *testing.T) {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("file", "report.txt")
	require.NoError(t, err)
	_, err = part.Write([]byte("payload"))
	require.NoError(t, err)
	require.NoError(t, writer.Close())

	req := httptest.NewRequest("POST", "/", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	require.NoError(t, req.ParseMultipartForm(1024))
	fh := req.MultipartForm.File["file"][0]

	tmpPath, err := SaveUploadedFileToTemp(fh, t.TempDir())
	require.NoError(t, err)

	data, err := os.ReadFile(tmpPath)
	require.NoError(t, err)
	assert.Equal(t, "payload", string(data))
}

func TestIsWithinDir(t *testing.T) {
	assert.True(t, IsWithinDir("/data/flow-1/uploads/file.txt", "/data/flow-1/uploads"))
	assert.True(t, IsWithinDir("/data/flow-1/uploads/sub/file.txt", "/data/flow-1/uploads"))
	assert.False(t, IsWithinDir("/data/flow-1/../evil.txt", "/data/flow-1/uploads"))
	assert.False(t, IsWithinDir("/data/flow-2/uploads/file.txt", "/data/flow-1/uploads"))
}

func TestResolvePulledStagedTarget(t *testing.T) {
	t.Run("full cache path archive", func(t *testing.T) {
		stagingDir := t.TempDir()
		target := filepath.Join(stagingDir, "etc", "nginx", "nginx.conf")
		require.NoError(t, os.MkdirAll(filepath.Dir(target), 0755))
		require.NoError(t, os.WriteFile(target, []byte("nginx"), 0644))

		assert.Equal(t, target, ResolvePulledStagedTarget(stagingDir, "etc/nginx/nginx.conf"))
	})

	t.Run("basename archive", func(t *testing.T) {
		stagingDir := t.TempDir()
		target := filepath.Join(stagingDir, "nginx.conf")
		require.NoError(t, os.WriteFile(target, []byte("nginx"), 0644))

		assert.Equal(t, target, ResolvePulledStagedTarget(stagingDir, "etc/nginx/nginx.conf"))
	})
}

func TestWriteUploadsTar(t *testing.T) {
	uploadDir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(uploadDir, "a.txt"), []byte("alpha"), 0644))
	require.NoError(t, os.Mkdir(filepath.Join(uploadDir, "sub"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(uploadDir, "sub", "b.txt"), []byte("bravo"), 0644))
	if err := os.Symlink(filepath.Join(uploadDir, "a.txt"), filepath.Join(uploadDir, "link.txt")); err != nil {
		t.Skipf("symlink creation not available: %v", err)
	}

	pr, pw := io.Pipe()
	errCh := make(chan error, 1)
	go func() {
		errCh <- WriteUploadsTar(pw, uploadDir)
	}()

	var buf bytes.Buffer
	_, err := io.Copy(&buf, pr)
	require.NoError(t, err)
	require.NoError(t, <-errCh)

	tr := tar.NewReader(bytes.NewReader(buf.Bytes()))
	contents := map[string]string{}
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		require.NoError(t, err)
		if hdr.Typeflag != tar.TypeReg {
			continue
		}
		data, err := io.ReadAll(tr)
		require.NoError(t, err)
		contents[hdr.Name] = string(data)
	}

	assert.Equal(t, "alpha", contents["uploads/a.txt"])
	assert.Equal(t, "bravo", contents["uploads/sub/b.txt"])
	assert.NotContains(t, contents, "uploads/link.txt")
}

func TestExtractTarRegularFiles(t *testing.T) {
	destDir := t.TempDir()
	buf := buildTar(t, []tarTestEntry{
		{name: "dir/", typeflag: tar.TypeDir},
		{name: "dir/file.txt", typeflag: tar.TypeReg, content: "hello"},
	})

	require.NoError(t, ExtractTar(buf, destDir))

	data, err := os.ReadFile(filepath.Join(destDir, "dir", "file.txt"))
	require.NoError(t, err)
	assert.Equal(t, "hello", string(data))
}

func TestExtractTarSkipsSymlinks(t *testing.T) {
	destDir := t.TempDir()
	buf := buildTar(t, []tarTestEntry{
		{name: "link.txt", typeflag: tar.TypeSymlink, linkname: "/etc/passwd"},
	})

	require.NoError(t, ExtractTar(buf, destDir))

	_, err := os.Lstat(filepath.Join(destDir, "link.txt"))
	assert.True(t, os.IsNotExist(err), "symlink must not be created in cache")
}

func TestExtractTarSkipsPathTraversal(t *testing.T) {
	destDir := t.TempDir()
	buf := buildTar(t, []tarTestEntry{
		{name: "../../evil.txt", typeflag: tar.TypeReg, content: "evil"},
	})

	require.NoError(t, ExtractTar(buf, destDir))

	evilPath := filepath.Join(filepath.Dir(destDir), "evil.txt")
	_, err := os.Lstat(evilPath)
	assert.True(t, os.IsNotExist(err), "path traversal file must not be created")
}

func TestExtractTarRejectsTooManyFiles(t *testing.T) {
	destDir := t.TempDir()
	entries := make([]tarTestEntry, 0, MaxPullFiles+1)
	for i := 0; i < MaxPullFiles+1; i++ {
		entries = append(entries, tarTestEntry{
			name:     filepath.Join("many", "file-"+strconv.Itoa(i)+".txt"),
			typeflag: tar.TypeReg,
			content:  "x",
		})
	}

	err := ExtractTar(buildTar(t, entries), destDir)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "maximum file count")
}

func TestZipDirectory(t *testing.T) {
	src := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(src, "a.txt"), []byte("hello"), 0644))
	require.NoError(t, os.MkdirAll(filepath.Join(src, "sub"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(src, "sub", "b.txt"), []byte("world"), 0644))

	var buf bytes.Buffer
	require.NoError(t, ZipDirectory(&buf, src))

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

func TestZipDirectoryExcludesSymlinks(t *testing.T) {
	src := t.TempDir()
	target := filepath.Join(src, "real.txt")
	link := filepath.Join(src, "link.txt")
	require.NoError(t, os.WriteFile(target, []byte("real"), 0644))
	if err := os.Symlink(target, link); err != nil {
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
	assert.Contains(t, names, "real.txt")
	assert.NotContains(t, names, "link.txt")
}

func TestSort(t *testing.T) {
	now := time.Now()
	files := []File{
		{Name: "b.txt", ModifiedAt: now.Add(-2 * time.Hour)},
		{Name: "a.txt", ModifiedAt: now.Add(-1 * time.Hour)},
		{Name: "c.txt", ModifiedAt: now.Add(-1 * time.Hour)},
	}
	Sort(files)

	assert.Equal(t, "a.txt", files[0].Name)
	assert.Equal(t, "c.txt", files[1].Name)
	assert.Equal(t, "b.txt", files[2].Name)
}
