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
	assert.Equal(t, "/data/flow-42-data/resources", FlowResourcesDir("/data", 42))
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
		{"resources file", "resources/creds/passwords.txt", "/data/flow-1-data/resources/creds/passwords.txt", false},
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
	resourcesDir := FlowResourcesDir(dataDir, 7)
	require.NoError(t, os.MkdirAll(uploadsDir, 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(containerDir, "etc", "nginx", "conf"), 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(resourcesDir, "creds"), 0755))

	require.NoError(t, os.WriteFile(filepath.Join(uploadsDir, "wordlist.txt"), []byte("words"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(containerDir, "etc", "nginx", "nginx.conf"), []byte("nginx"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(resourcesDir, "creds", "passwords.txt"), []byte("secret"), 0644))

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
	assert.Contains(t, paths, "resources/creds")
	assert.Contains(t, paths, "resources/creds/passwords.txt")
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

func TestCopyResourcesToFlow(t *testing.T) {
	dataDir := t.TempDir()
	storeDir := filepath.Join(dataDir, "resources")
	require.NoError(t, os.MkdirAll(storeDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(storeDir, "hash-a.blob"), []byte("alpha"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(storeDir, "hash-b.blob"), []byte("bravo"), 0644))

	added, err := CopyResourcesToFlow(dataDir, storeDir, 3, []ResourceRef{
		{VirtualPath: "creds", IsDir: true},
		{Hash: "hash-a", VirtualPath: "creds/passwords.txt", Name: "passwords.txt"},
		{Hash: "hash-b", VirtualPath: "notes.txt", Name: "notes.txt"},
	}, false)
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{
		"resources/creds/passwords.txt",
		"resources/notes.txt",
	}, added)

	data, err := os.ReadFile(filepath.Join(FlowResourcesDir(dataDir, 3), "creds", "passwords.txt"))
	require.NoError(t, err)
	assert.Equal(t, "alpha", string(data))

	added, err = CopyResourcesToFlow(dataDir, storeDir, 3, []ResourceRef{
		{Hash: "hash-a", VirtualPath: "creds/passwords.txt", Name: "passwords.txt"},
	}, false)
	require.NoError(t, err)
	assert.Empty(t, added)

	require.NoError(t, os.WriteFile(filepath.Join(storeDir, "hash-a.blob"), []byte("updated"), 0644))
	added, err = CopyResourcesToFlow(dataDir, storeDir, 3, []ResourceRef{
		{Hash: "hash-a", VirtualPath: "creds/passwords.txt", Name: "passwords.txt"},
	}, true)
	require.NoError(t, err)
	assert.Equal(t, []string{"resources/creds/passwords.txt"}, added)

	data, err = os.ReadFile(filepath.Join(FlowResourcesDir(dataDir, 3), "creds", "passwords.txt"))
	require.NoError(t, err)
	assert.Equal(t, "updated", string(data))
}

func TestCopyResourcesToFlowRejectsEscapingPath(t *testing.T) {
	dataDir := t.TempDir()
	storeDir := filepath.Join(dataDir, "resources")
	require.NoError(t, os.MkdirAll(storeDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(storeDir, "hash.blob"), []byte("x"), 0644))

	_, err := CopyResourcesToFlow(dataDir, storeDir, 3, []ResourceRef{
		{Hash: "hash", VirtualPath: "../evil.txt", Name: "evil.txt"},
	}, false)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "escapes resources directory")
}

func TestWriteResourcesTar(t *testing.T) {
	resourcesDir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(resourcesDir, "creds"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(resourcesDir, "creds", "passwords.txt"), []byte("secret"), 0644))
	if err := os.Symlink(filepath.Join(resourcesDir, "creds", "passwords.txt"), filepath.Join(resourcesDir, "link.txt")); err != nil {
		t.Skipf("symlink creation not available: %v", err)
	}

	pr, pw := io.Pipe()
	errCh := make(chan error, 1)
	go func() {
		errCh <- WriteResourcesTar(pw, resourcesDir)
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

	assert.Equal(t, "secret", contents["resources/creds/passwords.txt"])
	assert.NotContains(t, contents, "resources/link.txt")
}

func TestFileListingForPrompt(t *testing.T) {
	dataDir := t.TempDir()
	uploadsDir := FlowUploadsDir(dataDir, 11)
	resourcesDir := FlowResourcesDir(dataDir, 11)
	require.NoError(t, os.MkdirAll(filepath.Join(uploadsDir, "targets"), 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(resourcesDir, "creds"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(uploadsDir, "targets", "ips.txt"), []byte("127.0.0.1"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(resourcesDir, "creds", "passwords.txt"), []byte("secret"), 0644))
	if err := os.Symlink(filepath.Join(uploadsDir, "targets", "ips.txt"), filepath.Join(uploadsDir, "link.txt")); err != nil {
		t.Skipf("symlink creation not available: %v", err)
	}

	listing := FileListingForPrompt(dataDir, 11)

	assert.Contains(t, listing, "<task_files>")
	assert.Contains(t, listing, `<uploads base="/work/uploads">`)
	assert.Contains(t, listing, "targets/ips.txt\n")
	assert.Contains(t, listing, `<resources base="/work/resources">`)
	assert.Contains(t, listing, "creds/passwords.txt\n")
	assert.NotContains(t, listing, "link.txt")
}

func TestFileListingForPromptEmpty(t *testing.T) {
	assert.Empty(t, FileListingForPrompt(t.TempDir(), 11))
}

func TestBaseName(t *testing.T) {
	assert.Equal(t, "passwords.txt", BaseName("resources/creds/passwords.txt"))
	assert.Equal(t, "passwords.txt", BaseName(`resources\creds\passwords.txt`))
	assert.Equal(t, "plain.txt", BaseName("plain.txt"))
}

func TestWriteSingleFileTar(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "passwords.txt")
	require.NoError(t, os.WriteFile(filePath, []byte("secret"), 0644))

	pr, pw := io.Pipe()
	errCh := make(chan error, 1)
	go func() {
		errCh <- WriteSingleFileTar(pw, filePath, "resources/creds/passwords.txt")
	}()

	var buf bytes.Buffer
	_, err := io.Copy(&buf, pr)
	require.NoError(t, err)
	require.NoError(t, <-errCh)

	tr := tar.NewReader(bytes.NewReader(buf.Bytes()))
	seenDirs := map[string]bool{}
	contents := map[string]string{}
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		require.NoError(t, err)
		switch hdr.Typeflag {
		case tar.TypeDir:
			seenDirs[hdr.Name] = true
		case tar.TypeReg:
			data, err := io.ReadAll(tr)
			require.NoError(t, err)
			contents[hdr.Name] = string(data)
		}
	}

	assert.True(t, seenDirs["resources"])
	assert.True(t, seenDirs["resources/creds"])
	assert.Equal(t, "secret", contents["resources/creds/passwords.txt"])
}

func TestWriteFilesTar(t *testing.T) {
	dir := t.TempDir()
	firstPath := filepath.Join(dir, "passwords.txt")
	secondPath := filepath.Join(dir, "ips.txt")
	missingPath := filepath.Join(dir, "missing.txt")
	require.NoError(t, os.WriteFile(firstPath, []byte("secret"), 0644))
	require.NoError(t, os.WriteFile(secondPath, []byte("127.0.0.1"), 0644))
	if err := os.Symlink(firstPath, filepath.Join(dir, "link.txt")); err != nil {
		t.Skipf("symlink creation not available: %v", err)
	}

	pr, pw := io.Pipe()
	errCh := make(chan error, 1)
	go func() {
		errCh <- WriteFilesTar(pw, []TarEntry{
			{LocalPath: firstPath, TarPath: "resources/creds/passwords.txt"},
			{LocalPath: secondPath, TarPath: "uploads/targets/ips.txt"},
			{LocalPath: filepath.Join(dir, "link.txt"), TarPath: "uploads/link.txt"},
			{LocalPath: missingPath, TarPath: "uploads/missing.txt"},
		})
	}()

	var buf bytes.Buffer
	_, err := io.Copy(&buf, pr)
	require.NoError(t, err)
	require.NoError(t, <-errCh)

	tr := tar.NewReader(bytes.NewReader(buf.Bytes()))
	seenDirs := map[string]bool{}
	contents := map[string]string{}
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		require.NoError(t, err)
		switch hdr.Typeflag {
		case tar.TypeDir:
			seenDirs[hdr.Name] = true
		case tar.TypeReg:
			data, err := io.ReadAll(tr)
			require.NoError(t, err)
			contents[hdr.Name] = string(data)
		}
	}

	assert.True(t, seenDirs["resources"])
	assert.True(t, seenDirs["resources/creds"])
	assert.True(t, seenDirs["uploads"])
	assert.True(t, seenDirs["uploads/targets"])
	assert.Equal(t, "secret", contents["resources/creds/passwords.txt"])
	assert.Equal(t, "127.0.0.1", contents["uploads/targets/ips.txt"])
	assert.NotContains(t, contents, "uploads/link.txt")
	assert.NotContains(t, contents, "uploads/missing.txt")
}

func TestWriteFilesTarRejectsInvalidTarPath(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "passwords.txt")
	require.NoError(t, os.WriteFile(filePath, []byte("secret"), 0644))

	pr, pw := io.Pipe()
	errCh := make(chan error, 1)
	go func() {
		errCh <- WriteFilesTar(pw, []TarEntry{{LocalPath: filePath, TarPath: "../evil.txt"}})
	}()

	_, err := io.Copy(io.Discard, pr)
	require.Error(t, err)
	require.Error(t, <-errCh)
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

func TestZipRelativePaths(t *testing.T) {
	t.Run("files and directory contents with cache-relative names", func(t *testing.T) {
		base := t.TempDir()
		require.NoError(t, os.MkdirAll(filepath.Join(base, "uploads"), 0755))
		require.NoError(t, os.WriteFile(filepath.Join(base, "uploads", "a.txt"), []byte("alpha"), 0644))
		require.NoError(t, os.MkdirAll(filepath.Join(base, "container", "etc", "nginx"), 0755))
		require.NoError(t, os.WriteFile(filepath.Join(base, "container", "etc", "nginx", "nginx.conf"), []byte("nginx"), 0644))
		if err := os.Symlink(filepath.Join(base, "uploads", "a.txt"), filepath.Join(base, "uploads", "link.txt")); err != nil {
			t.Skipf("symlink creation not available: %v", err)
		}

		var buf bytes.Buffer
		err := ZipRelativePaths(&buf, base, []string{
			"uploads/a.txt",
			"uploads/link.txt",    // symlink: skipped
			"container/etc",       // directory: entries under container/etc/...
			"uploads/missing.txt", // missing: silently skipped
		})
		require.NoError(t, err)

		zr, err := zip.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
		require.NoError(t, err)
		contents := map[string]string{}
		for _, f := range zr.File {
			rc, err := f.Open()
			require.NoError(t, err)
			data, _ := io.ReadAll(rc)
			rc.Close()
			contents[f.Name] = string(data)
		}

		assert.Equal(t, "alpha", contents["uploads/a.txt"])
		assert.Equal(t, "nginx", contents["container/etc/nginx/nginx.conf"])
		assert.NotContains(t, contents, "uploads/link.txt", "symlinks must be excluded")
		assert.NotContains(t, contents, "uploads/missing.txt", "missing files must be silently skipped")
	})

	t.Run("empty relPaths produces empty zip", func(t *testing.T) {
		var buf bytes.Buffer
		require.NoError(t, ZipRelativePaths(&buf, t.TempDir(), nil))

		zr, err := zip.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
		require.NoError(t, err)
		assert.Empty(t, zr.File)
	})
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

func TestDeduplicatePaths(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		// ── core user scenario ────────────────────────────────────────────────
		{
			name: "user example: parent covers nested paths, underscore sibling survives",
			input: []string{
				"uploads/my_dir1/my_dir2/",
				"uploads/my_dir1/my_file1.txt",
				"uploads/my_dir1/my_dir2/my_file2.txt",
				"uploads/my_dir1_temp",
				"uploads/my_dir1",
			},
			// uploads/my_dir1 covers the first three; uploads/my_dir1_temp is independent
			expected: []string{"uploads/my_dir1_temp", "uploads/my_dir1"},
		},

		// ── exact deduplication ───────────────────────────────────────────────
		{
			name:     "exact string duplicates keep first occurrence",
			input:    []string{"uploads/file.txt", "uploads/file.txt", "uploads/file.txt"},
			expected: []string{"uploads/file.txt"},
		},
		{
			name:     "trailing slash and no trailing slash are the same path",
			input:    []string{"uploads/dir/", "uploads/dir"},
			expected: []string{"uploads/dir/"},
		},
		{
			name:  "normalisation via .. collapses to same path as plain entry",
			input: []string{"uploads/my_dir1_temp/../my_dir1/", "uploads/my_dir1"},
			// both clean to "uploads/my_dir1"; original of first occurrence is returned
			expected: []string{"uploads/my_dir1_temp/../my_dir1/"},
		},

		// ── parent-covers-child ───────────────────────────────────────────────
		{
			name:     "parent covers direct child file",
			input:    []string{"uploads/dir/file.txt", "uploads/dir"},
			expected: []string{"uploads/dir"},
		},
		{
			name:     "parent covers nested subdirectory",
			input:    []string{"uploads/dir/sub/", "uploads/dir"},
			expected: []string{"uploads/dir"},
		},
		{
			name:     "top-level covers deep nesting",
			input:    []string{"uploads/a/b/c/d.txt", "uploads/a"},
			expected: []string{"uploads/a"},
		},
		{
			name:     "all descendants collapsed to single ancestor",
			input:    []string{"uploads/a/x.txt", "uploads/a/y.txt", "uploads/a/sub/z.txt", "uploads/a"},
			expected: []string{"uploads/a"},
		},
		{
			name:     "ancestor given last still covers earlier entries",
			input:    []string{"uploads/dir/file.txt", "uploads/dir/sub/", "uploads/dir"},
			expected: []string{"uploads/dir"},
		},
		{
			name:     "intermediate node covers its subtree but is covered by root",
			input:    []string{"uploads/a", "uploads/a/b", "uploads/a/b/c.txt"},
			expected: []string{"uploads/a"},
		},

		// ── no false positives ────────────────────────────────────────────────
		{
			name:     "underscore suffix prevents false parent match",
			input:    []string{"uploads/dir", "uploads/dir_extra"},
			expected: []string{"uploads/dir", "uploads/dir_extra"},
		},
		{
			name:     "numeric suffix does not create false match",
			input:    []string{"uploads/dir1", "uploads/dir"},
			expected: []string{"uploads/dir1", "uploads/dir"},
		},
		{
			name:     "sibling directories both survive",
			input:    []string{"uploads/a", "uploads/b"},
			expected: []string{"uploads/a", "uploads/b"},
		},
		{
			name:     "independent flat files both survive",
			input:    []string{"uploads/a.txt", "uploads/b.txt"},
			expected: []string{"uploads/a.txt", "uploads/b.txt"},
		},

		// ── order preservation ────────────────────────────────────────────────
		{
			name:     "input order preserved when no path is covered",
			input:    []string{"uploads/z.txt", "uploads/a.txt", "uploads/m.txt"},
			expected: []string{"uploads/z.txt", "uploads/a.txt", "uploads/m.txt"},
		},

		// ── multiple namespaces ───────────────────────────────────────────────
		{
			name:     "uploads resources container do not interfere with each other",
			input:    []string{"uploads/dir", "resources/dir", "container/dir"},
			expected: []string{"uploads/dir", "resources/dir", "container/dir"},
		},
		{
			name:     "same relative name in different namespaces are independent",
			input:    []string{"uploads/dir/file.txt", "container/dir"},
			expected: []string{"uploads/dir/file.txt", "container/dir"},
		},
		{
			name:     "coverage is scoped within each namespace",
			input:    []string{"uploads/dir/file.txt", "uploads/dir", "resources/dir/file.txt"},
			expected: []string{"uploads/dir", "resources/dir/file.txt"},
		},

		// ── original value preservation ───────────────────────────────────────
		{
			name:     "original path with trailing slash returned when it survives",
			input:    []string{"uploads/dir/", "resources/other.txt"},
			expected: []string{"uploads/dir/", "resources/other.txt"},
		},
		{
			name:  "original dotdot path returned when it survives and is safe",
			input: []string{"uploads/tmp/../keep/"},
			// path.Clean → "uploads/keep"; not absolute, not leading ".." → safe; original returned
			expected: []string{"uploads/tmp/../keep/"},
		},

		// ── security: reject path traversal ──────────────────────────────────
		{
			name:     "leading dotdot path rejected",
			input:    []string{"../etc/passwd"},
			expected: nil,
		},
		{
			name:     "absolute path rejected",
			input:    []string{"/etc/passwd"},
			expected: nil,
		},
		{
			name:  "path that cleans to dotdot escape rejected",
			input: []string{"uploads/../../etc/passwd"},
			// path.Clean → "../../etc/passwd" → starts with "../" → rejected
			expected: nil,
		},
		{
			name:  "dotdot that fully escapes root rejected",
			input: []string{"../"},
			// path.Clean → ".." → equals ".." → rejected
			expected: nil,
		},
		{
			name:     "mixed safe and unsafe: only safe paths returned",
			input:    []string{"uploads/safe.txt", "../etc/passwd", "/etc/shadow", "uploads/../../evil"},
			expected: []string{"uploads/safe.txt"},
		},
		{
			name:     "all unsafe inputs return nil",
			input:    []string{"../a", "/b", "uploads/../../c"},
			expected: nil,
		},

		// ── edge cases ────────────────────────────────────────────────────────
		{
			name:     "nil input returns nil",
			input:    nil,
			expected: nil,
		},
		{
			name:     "empty slice returns nil",
			input:    []string{},
			expected: nil,
		},
		{
			name:     "whitespace-only entries dropped",
			input:    []string{"", "   ", "\t", "uploads/file.txt"},
			expected: []string{"uploads/file.txt"},
		},
		{
			name:     "single safe path returned unchanged",
			input:    []string{"uploads/report.txt"},
			expected: []string{"uploads/report.txt"},
		},
		{
			name:     "backslash normalised to slash for comparison, original returned",
			input:    []string{`uploads\file.txt`},
			expected: []string{`uploads\file.txt`},
		},
		{
			name:  "backslash and slash variants of same path are deduplicated",
			input: []string{`uploads\file.txt`, "uploads/file.txt"},
			// both clean to "uploads/file.txt"; first occurrence wins
			expected: []string{`uploads\file.txt`},
		},
		{
			name:     "dot in middle is cleaned and deduped correctly",
			input:    []string{"uploads/./file.txt", "uploads/file.txt"},
			expected: []string{"uploads/./file.txt"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DeduplicatePaths(tt.input)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestSort(t *testing.T) {
	now := time.Now()
	files := []File{
		{Name: "b.txt", Path: "b.txt", ModifiedAt: now.Add(-2 * time.Hour)},
		{Name: "a.txt", Path: "a.txt", ModifiedAt: now.Add(-1 * time.Hour)},
		{Name: "c.txt", Path: "c.txt", ModifiedAt: now.Add(-1 * time.Hour)},
	}
	Sort(files)

	assert.Equal(t, "a.txt", files[0].Name)
	assert.Equal(t, "c.txt", files[2].Name)
	assert.Equal(t, "b.txt", files[1].Name)
}
