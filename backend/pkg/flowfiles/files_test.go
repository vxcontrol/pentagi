package flowfiles

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"errors"
	"io"
	"io/fs"
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

var errFilesWriteFailed = errors.New("write failed")

type errWriter struct{}

func (errWriter) Write([]byte) (int, error) { return 0, errFilesWriteFailed }

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

// filesStreamTar runs write against a pipe and returns the directory headers and file contents a reader saw,
// the error that ended the stream, and write's own error.
func filesStreamTar(
	t *testing.T,
	write func(*io.PipeWriter) error,
) (dirs map[string]bool, contents map[string]string, readErr, writeErr error) {
	t.Helper()

	pr, pw := io.Pipe()
	defer pr.Close()
	writeDone := make(chan error, 1)
	go func() { writeDone <- write(pw) }()

	dirs, contents = map[string]bool{}, map[string]string{}
	readDone := make(chan error, 1)
	go func() {
		tr := tar.NewReader(pr)
		for {
			hdr, err := tr.Next()
			if err != nil {
				if errors.Is(err, io.EOF) {
					err = nil
				}
				readDone <- err
				return
			}
			switch hdr.Typeflag {
			case tar.TypeDir:
				dirs[hdr.Name] = true
			case tar.TypeReg:
				data, err := io.ReadAll(tr)
				if err != nil {
					readDone <- err
					return
				}
				contents[hdr.Name] = string(data)
			}
		}
	}()

	timeout := time.After(10 * time.Second)
	select {
	case readErr = <-readDone:
		// A writer still in Write when the reader failed gets its error instead of blocking out the budget.
		_ = pr.CloseWithError(readErr)
	case <-timeout:
		t.Fatal("the tar stream never ended: the writer left its pipe open")
	}
	select {
	case writeErr = <-writeDone:
	case <-timeout:
		t.Fatal("the tar writer never returned after its stream ended")
	}
	return dirs, contents, readErr, writeErr
}

func filesReadTar(t *testing.T, write func(*io.PipeWriter) error) (map[string]bool, map[string]string) {
	t.Helper()

	dirs, contents, readErr, writeErr := filesStreamTar(t, write)
	require.NoError(t, readErr)
	require.NoError(t, writeErr)
	return dirs, contents
}

// filesReadZip returns the name and content of every entry of a zip archive.
func filesReadZip(t *testing.T, data []byte) map[string]string {
	t.Helper()

	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	require.NoError(t, err)

	contents := map[string]string{}
	for _, f := range zr.File {
		rc, err := f.Open()
		require.NoError(t, err)
		body, err := io.ReadAll(rc)
		rc.Close()
		require.NoError(t, err)
		contents[f.Name] = string(body)
	}
	return contents
}

func filesSymlink(t *testing.T, target, link string) {
	t.Helper()
	if err := os.Symlink(target, link); err != nil {
		t.Skipf("symlink creation not available: %v", err)
	}
}

// Covers FlowDataDir and the three cache directories built on it.
func TestFiles_FlowDirectoriesNestUnderTheFlowDataDir(t *testing.T) {
	assert.Equal(t, "/data/flow-42-data", FlowDataDir("/data", 42))
	assert.Equal(t, "/data/flow-42-data/uploads", FlowUploadsDir("/data", 42))
	assert.Equal(t, "/data/flow-42-data/container", FlowContainerDir("/data", 42))
	assert.Equal(t, "/data/flow-42-data/resources", FlowResourcesDir("/data", 42))
}

func TestFiles_SanitizeFileName_KeepsOnlyASafeBaseName(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr string
	}{
		{name: "plain name", input: "report.txt", want: "report.txt"},
		{name: "strips parent traversal", input: "../report.txt", want: "report.txt"},
		{name: "normalizes windows separators", input: `nested\brief.md`, want: "brief.md"},
		{name: "strips absolute path", input: "/etc/passwd", want: "passwd"},
		{name: "strips deep traversal", input: "../../etc/shadow", want: "shadow"},
		{name: "trims whitespace", input: "  wordlist.txt  ", want: "wordlist.txt"},
		{name: "rejects empty", input: "   ", wantErr: "file name is required"},
		{name: "rejects dot-only", input: ".", wantErr: "invalid file name"},
		{name: "rejects root slash", input: "/", wantErr: "invalid file name"},
		{name: "rejects control characters", input: "bad\nname.txt", wantErr: "control characters"},
		{name: "rejects unsupported header characters", input: `bad"name.txt`, wantErr: "unsupported characters"},
		{name: "rejects too long", input: string(bytes.Repeat([]byte("a"), MaxFileNameLength+1)), wantErr: "too long"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SanitizeFileName(tt.input)
			if tt.wantErr != "" {
				assert.ErrorContains(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestFiles_SanitizeContainerCachePath_KeepsARelativeSafePath(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr string
	}{
		{name: "absolute directory", input: "/etc/nginx/conf/", want: "etc/nginx/conf"},
		{name: "absolute file", input: "/etc/nginx/nginx.conf", want: "etc/nginx/nginx.conf"},
		{name: "relative file", input: "var/log/app.log", want: "var/log/app.log"},
		{name: "normalizes traversal", input: "../../etc/shadow", want: "etc/shadow"},
		{name: "normalizes windows separators", input: `etc\nginx\nginx.conf`, want: "etc/nginx/nginx.conf"},
		{name: "rejects empty", input: "   ", wantErr: "path is required"},
		{name: "rejects root", input: "/", wantErr: "invalid path"},
		{name: "rejects bad component", input: "/etc/bad\nname", wantErr: "control characters"},
		{name: "rejects unsupported component", input: `/etc/bad"name`, wantErr: "unsupported characters"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SanitizeContainerCachePath(tt.input)
			if tt.wantErr != "" {
				assert.ErrorContains(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestFiles_ResolveCachedPath_StaysInsideTheFlowCache(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr string
	}{
		{name: "uploads file", input: "uploads/report.txt", want: "/data/flow-1-data/uploads/report.txt"},
		{name: "container file", input: "container/etc/nginx/nginx.conf", want: "/data/flow-1-data/container/etc/nginx/nginx.conf"},
		{name: "resources file", input: "resources/creds/passwords.txt", want: "/data/flow-1-data/resources/creds/passwords.txt"},
		{name: "container windows separators", input: `container\etc\nginx.conf`, want: "/data/flow-1-data/container/etc/nginx.conf"},
		{name: "empty path", input: "", wantErr: "path query parameter is required"},
		{name: "wrong prefix", input: "tmp/evil.sh", wantErr: "path must start with"},
		{name: "absolute path", input: "/etc/passwd", wantErr: "path must be relative"},
		{name: "path traversal", input: "uploads/../../etc/passwd", wantErr: "path must start with"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ResolveCachedPath("/data", 1, tt.input)
			if tt.wantErr != "" {
				assert.ErrorContains(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestFiles_ListDirEntries_SkipsTempEntriesAndSymlinks(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "a.txt"), []byte("a"), 0644))
	require.NoError(t, os.Mkdir(filepath.Join(dir, "sub"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".upload-temp"), []byte("tmp"), 0644))
	require.NoError(t, os.Mkdir(filepath.Join(dir, ".pull-temp"), 0755))
	filesSymlink(t, filepath.Join(dir, "a.txt"), filepath.Join(dir, "link.txt"))

	entries, err := ListDirEntries(dir, UploadsDirName)
	require.NoError(t, err)

	type listed struct {
		name, path string
		isDir      bool
	}
	got := make([]listed, len(entries))
	for i, entry := range entries {
		got[i] = listed{entry.Name, entry.Path, entry.IsDir}
	}
	assert.ElementsMatch(t, []listed{{"a.txt", "uploads/a.txt", false}, {"sub", "uploads/sub", true}}, got)

	entries, err = ListDirEntries(filepath.Join(t.TempDir(), "missing"), UploadsDirName)
	require.NoError(t, err, "a missing directory lists nothing")
	assert.Empty(t, entries)
}

func TestFiles_ListDirEntriesRecursive_KeepsNestedPathsAndSkipsTempTrees(t *testing.T) {
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
	assert.ElementsMatch(t, []string{
		"container/etc",
		"container/etc/nginx",
		"container/etc/nginx/conf",
		"container/etc/nginx/nginx.conf",
	}, paths)

	entries, err = ListDirEntriesRecursive(filepath.Join(t.TempDir(), "missing"), ContainerDirName)
	require.NoError(t, err, "a missing directory lists nothing")
	assert.Empty(t, entries)
}

func TestFiles_List_MergesEveryCacheSource(t *testing.T) {
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

func TestFiles_LocalEntryExists_ReportsWhetherThePathExists(t *testing.T) {
	filePath := filepath.Join(t.TempDir(), "f.txt")

	exists, err := LocalEntryExists(filePath)
	require.NoError(t, err)
	assert.False(t, exists)

	require.NoError(t, os.WriteFile(filePath, []byte("x"), 0644))

	exists, err = LocalEntryExists(filePath)
	require.NoError(t, err)
	assert.True(t, exists)
}

func TestFiles_RegularFileInfo_RefusesADirectory(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "f.txt")
	require.NoError(t, os.WriteFile(filePath, []byte("x"), 0644))

	info, err := RegularFileInfo(filePath)
	require.NoError(t, err)
	assert.Equal(t, "f.txt", info.Name())

	_, err = RegularFileInfo(dir)
	assert.ErrorContains(t, err, "is not a regular file")
}

func TestFiles_SaveUploadedFileToTemp_WritesTheUploadedBody(t *testing.T) {
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

func TestFiles_MaxUploadFileSize_Is300MiB(t *testing.T) {
	assert.Equal(t, int64(300*1024*1024), int64(MaxUploadFileSize))
}

func TestFiles_IsWithinDir_RejectsSiblingsAndEscapes(t *testing.T) {
	assert.True(t, IsWithinDir("/data/flow-1/uploads/file.txt", "/data/flow-1/uploads"))
	assert.True(t, IsWithinDir("/data/flow-1/uploads/sub/file.txt", "/data/flow-1/uploads"))
	assert.False(t, IsWithinDir("/data/flow-1/../evil.txt", "/data/flow-1/uploads"))
	assert.False(t, IsWithinDir("/data/flow-2/uploads/file.txt", "/data/flow-1/uploads"))
}

func TestFiles_ResolvePulledStagedTarget_FindsTheFileInsideStagingOnly(t *testing.T) {
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

	t.Run("escaping cache path is rejected", func(t *testing.T) {
		root := t.TempDir()
		stagingDir := filepath.Join(root, "staging")
		require.NoError(t, os.MkdirAll(stagingDir, 0755))
		// The escaping candidate exists outside staging; the basename candidate does not exist inside it.
		require.NoError(t, os.WriteFile(filepath.Join(root, "evil.conf"), []byte("evil"), 0644))

		assert.Equal(t, "", ResolvePulledStagedTarget(stagingDir, "../evil.conf"))
	})
}

// WriteUploadsTar and WriteResourcesTar share writeDirectoryTar; each row is one of them.
func TestFiles_WriteDirectoryTar_PrefixesEntriesAndSkipsSymlinks(t *testing.T) {
	tests := []struct {
		name  string
		write func(*io.PipeWriter, string) error
		root  string
	}{
		{name: "the uploads tar is rooted at uploads", write: WriteUploadsTar, root: "uploads"},
		{name: "the resources tar is rooted at resources", write: WriteResourcesTar, root: "resources"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			require.NoError(t, os.WriteFile(filepath.Join(dir, "a.txt"), []byte("alpha"), 0644))
			require.NoError(t, os.Mkdir(filepath.Join(dir, "sub"), 0755))
			require.NoError(t, os.WriteFile(filepath.Join(dir, "sub", "b.txt"), []byte("bravo"), 0644))
			filesSymlink(t, filepath.Join(dir, "a.txt"), filepath.Join(dir, "link.txt"))

			dirs, contents := filesReadTar(t, func(pw *io.PipeWriter) error { return tt.write(pw, dir) })

			assert.Equal(t, map[string]bool{tt.root: true, tt.root + "/sub": true}, dirs)
			assert.Equal(t, map[string]string{tt.root + "/a.txt": "alpha", tt.root + "/sub/b.txt": "bravo"}, contents)
		})
	}
}

func TestFiles_CopyResourcesToFlow_CopiesOnceAndOverwritesOnlyWhenForced(t *testing.T) {
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

func TestFiles_CopyResourcesToFlow_RejectsAnEscapingPath(t *testing.T) {
	dataDir := t.TempDir()
	storeDir := filepath.Join(dataDir, "resources")
	require.NoError(t, os.MkdirAll(storeDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(storeDir, "hash.blob"), []byte("x"), 0644))

	_, err := CopyResourcesToFlow(dataDir, storeDir, 3, []ResourceRef{
		{Hash: "hash", VirtualPath: "../evil.txt", Name: "evil.txt"},
	}, false)
	assert.ErrorContains(t, err, "escapes resources directory")
}

func TestFiles_FileListingForPrompt_ListsUploadsAndResourcesOrNothing(t *testing.T) {
	dataDir := t.TempDir()
	assert.Empty(t, FileListingForPrompt(dataDir, 11), "a flow without files adds nothing to the prompt")

	uploadsDir := FlowUploadsDir(dataDir, 11)
	resourcesDir := FlowResourcesDir(dataDir, 11)
	require.NoError(t, os.MkdirAll(filepath.Join(uploadsDir, "targets"), 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(resourcesDir, "creds"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(uploadsDir, "targets", "ips.txt"), []byte("127.0.0.1"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(resourcesDir, "creds", "passwords.txt"), []byte("secret"), 0644))
	filesSymlink(t, filepath.Join(uploadsDir, "targets", "ips.txt"), filepath.Join(uploadsDir, "link.txt"))

	listing := FileListingForPrompt(dataDir, 11)

	assert.Contains(t, listing, "<task_files>")
	assert.Contains(t, listing, `<uploads base="/work/uploads">`)
	assert.Contains(t, listing, "targets/ips.txt\n")
	assert.Contains(t, listing, `<resources base="/work/resources">`)
	assert.Contains(t, listing, "creds/passwords.txt\n")
	assert.NotContains(t, listing, "link.txt")
}

func TestFiles_BaseName_TakesTheLastComponentOfEitherSeparator(t *testing.T) {
	assert.Equal(t, "passwords.txt", BaseName("resources/creds/passwords.txt"))
	assert.Equal(t, "passwords.txt", BaseName(`resources\creds\passwords.txt`))
	assert.Equal(t, "plain.txt", BaseName("plain.txt"))
}

func TestFiles_WriteSingleFileTar_WritesTheFileWithItsParentDirectories(t *testing.T) {
	filePath := filepath.Join(t.TempDir(), "passwords.txt")
	require.NoError(t, os.WriteFile(filePath, []byte("secret"), 0644))

	dirs, contents := filesReadTar(t, func(pw *io.PipeWriter) error {
		return WriteSingleFileTar(pw, filePath, "resources/creds/passwords.txt")
	})

	assert.Equal(t, map[string]bool{"resources": true, "resources/creds": true}, dirs)
	assert.Equal(t, map[string]string{"resources/creds/passwords.txt": "secret"}, contents)
}

func TestFiles_WriteFilesTar_SkipsSymlinksAndMissingFiles(t *testing.T) {
	dir := t.TempDir()
	firstPath := filepath.Join(dir, "passwords.txt")
	secondPath := filepath.Join(dir, "ips.txt")
	require.NoError(t, os.WriteFile(firstPath, []byte("secret"), 0644))
	require.NoError(t, os.WriteFile(secondPath, []byte("127.0.0.1"), 0644))
	filesSymlink(t, firstPath, filepath.Join(dir, "link.txt"))

	dirs, contents := filesReadTar(t, func(pw *io.PipeWriter) error {
		return WriteFilesTar(pw, []TarEntry{
			{LocalPath: firstPath, TarPath: "resources/creds/passwords.txt"},
			{LocalPath: secondPath, TarPath: "uploads/targets/ips.txt"},
			{LocalPath: filepath.Join(dir, "link.txt"), TarPath: "uploads/link.txt"},
			{LocalPath: filepath.Join(dir, "missing.txt"), TarPath: "uploads/missing.txt"},
		})
	})

	assert.Equal(t, map[string]bool{"resources": true, "resources/creds": true, "uploads": true, "uploads/targets": true}, dirs)
	assert.Equal(t, map[string]string{
		"resources/creds/passwords.txt": "secret",
		"uploads/targets/ips.txt":       "127.0.0.1",
	}, contents)
}

func TestFiles_WriteFilesTar_RejectsAnEscapingTarPath(t *testing.T) {
	filePath := filepath.Join(t.TempDir(), "passwords.txt")
	require.NoError(t, os.WriteFile(filePath, []byte("secret"), 0644))

	_, _, readErr, writeErr := filesStreamTar(t, func(pw *io.PipeWriter) error {
		return WriteFilesTar(pw, []TarEntry{{LocalPath: filePath, TarPath: "../evil.txt"}})
	})

	assert.ErrorContains(t, readErr, "invalid tar entry path", "the reader must see the failure")
	assert.ErrorContains(t, writeErr, "invalid tar entry path")
}

func TestFiles_ExtractTar_WritesOnlyRegularFilesInsideTheDestination(t *testing.T) {
	tooMany := make([]tarTestEntry, 0, MaxPullFiles+1)
	for i := range MaxPullFiles + 1 {
		tooMany = append(tooMany, tarTestEntry{name: "many/file-" + strconv.Itoa(i) + ".txt", typeflag: tar.TypeReg, content: "x"})
	}

	tests := []struct {
		name    string
		entries []tarTestEntry
		want    map[string]string
		absent  []string
		wantErr string
	}{
		{
			name: "regular files",
			entries: []tarTestEntry{
				{name: "dir/", typeflag: tar.TypeDir},
				{name: "dir/file.txt", typeflag: tar.TypeReg, content: "hello"},
			},
			want: map[string]string{"dir/file.txt": "hello"},
		},
		{
			name:    "a symlink is skipped",
			entries: []tarTestEntry{{name: "link.txt", typeflag: tar.TypeSymlink, linkname: "/etc/passwd"}},
			absent:  []string{"cache/dest/link.txt"},
		},
		{
			name:    "a path traversal is skipped",
			entries: []tarTestEntry{{name: "../../evil.txt", typeflag: tar.TypeReg, content: "evil"}},
			absent:  []string{"evil.txt"},
		},
		{name: "too many files", entries: tooMany, wantErr: "maximum file count"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Two levels deep, so an escape of two levels still lands inside this test's directory.
			root := t.TempDir()
			destDir := filepath.Join(root, "cache", "dest")
			require.NoError(t, os.MkdirAll(destDir, 0755))

			err := ExtractTar(buildTar(t, tt.entries), destDir)
			if tt.wantErr != "" {
				assert.ErrorContains(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)

			for rel, content := range tt.want {
				data, err := os.ReadFile(filepath.Join(destDir, rel))
				require.NoError(t, err)
				assert.Equal(t, content, string(data))
			}
			for _, rel := range tt.absent {
				_, err := os.Lstat(filepath.Join(root, rel))
				assert.ErrorIs(t, err, fs.ErrNotExist, "%s must not be created", rel)
			}
		})
	}
}

func TestFiles_ZipRelativePaths_StoresEachPathUnderItsCacheName(t *testing.T) {
	t.Run("files and directory contents with cache-relative names", func(t *testing.T) {
		base := t.TempDir()
		require.NoError(t, os.MkdirAll(filepath.Join(base, "uploads"), 0755))
		require.NoError(t, os.WriteFile(filepath.Join(base, "uploads", "a.txt"), []byte("alpha"), 0644))
		require.NoError(t, os.MkdirAll(filepath.Join(base, "container", "etc", "nginx"), 0755))
		require.NoError(t, os.WriteFile(filepath.Join(base, "container", "etc", "nginx", "nginx.conf"), []byte("nginx"), 0644))
		filesSymlink(t, filepath.Join(base, "uploads", "a.txt"), filepath.Join(base, "uploads", "link.txt"))

		var buf bytes.Buffer
		require.NoError(t, ZipRelativePaths(&buf, base, []string{
			"uploads/a.txt",
			"uploads/link.txt",    // a symlink is skipped
			"container/etc",       // a directory contributes the files under it
			"uploads/missing.txt", // a missing file is skipped silently
		}))

		assert.Equal(t, map[string]string{
			"uploads/a.txt":                  "alpha",
			"container/etc/nginx/nginx.conf": "nginx",
		}, filesReadZip(t, buf.Bytes()))
	})

	t.Run("empty relPaths produces empty zip", func(t *testing.T) {
		var buf bytes.Buffer
		require.NoError(t, ZipRelativePaths(&buf, t.TempDir(), nil))
		assert.Empty(t, filesReadZip(t, buf.Bytes()))
	})

	t.Run("escaping relPath is skipped", func(t *testing.T) {
		root := t.TempDir()
		base := filepath.Join(root, "flow-1-data")
		require.NoError(t, os.MkdirAll(filepath.Join(base, "uploads"), 0755))
		require.NoError(t, os.WriteFile(filepath.Join(base, "uploads", "a.txt"), []byte("alpha"), 0644))
		require.NoError(t, os.WriteFile(filepath.Join(root, "secret.txt"), []byte("secret"), 0644))

		var buf bytes.Buffer
		require.NoError(t, ZipRelativePaths(&buf, base, []string{"uploads/a.txt", "../secret.txt"}))

		assert.Equal(t, map[string]string{"uploads/a.txt": "alpha"}, filesReadZip(t, buf.Bytes()))
	})

	// With nothing to write, the only write is Close's central-directory flush.
	t.Run("a failed close is an error", func(t *testing.T) {
		assert.ErrorIs(t, ZipRelativePaths(errWriter{}, t.TempDir(), nil), errFilesWriteFailed)
	})
}

func TestFiles_ZipDirectory_StoresRegularFilesUnderRelativeNames(t *testing.T) {
	assert.ErrorIs(t, ZipDirectory(errWriter{}, t.TempDir()), errFilesWriteFailed, "a failed close is an error")

	src := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(src, "a.txt"), []byte("hello"), 0644))
	require.NoError(t, os.MkdirAll(filepath.Join(src, "sub"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(src, "sub", "b.txt"), []byte("world"), 0644))
	filesSymlink(t, filepath.Join(src, "a.txt"), filepath.Join(src, "link.txt"))

	var buf bytes.Buffer
	require.NoError(t, ZipDirectory(&buf, src))
	assert.Equal(t, map[string]string{"a.txt": "hello", "sub/b.txt": "world"}, filesReadZip(t, buf.Bytes()))
}

func TestFiles_DeduplicatePaths_KeepsTheFirstSafeCoveringPath(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name: "user example: parent covers nested paths, underscore sibling survives",
			input: []string{
				"uploads/my_dir1/my_dir2/",
				"uploads/my_dir1/my_file1.txt",
				"uploads/my_dir1/my_dir2/my_file2.txt",
				"uploads/my_dir1_temp",
				"uploads/my_dir1",
			},
			expected: []string{"uploads/my_dir1_temp", "uploads/my_dir1"},
		},
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
			name:     "normalisation via .. collapses to same path as plain entry",
			input:    []string{"uploads/my_dir1_temp/../my_dir1/", "uploads/my_dir1"},
			expected: []string{"uploads/my_dir1_temp/../my_dir1/"},
		},
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
			name:     "all descendants collapsed to single ancestor",
			input:    []string{"uploads/a/x.txt", "uploads/a/y.txt", "uploads/a/sub/z.txt", "uploads/a"},
			expected: []string{"uploads/a"},
		},
		{
			name:     "intermediate node covers its subtree but is covered by root",
			input:    []string{"uploads/a", "uploads/a/b", "uploads/a/b/c.txt"},
			expected: []string{"uploads/a"},
		},
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
			name:     "input order preserved when no path is covered",
			input:    []string{"uploads/z.txt", "uploads/a.txt", "uploads/m.txt"},
			expected: []string{"uploads/z.txt", "uploads/a.txt", "uploads/m.txt"},
		},
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
		{
			name:     "original dotdot path returned when it survives and is safe",
			input:    []string{"uploads/tmp/../keep/"},
			expected: []string{"uploads/tmp/../keep/"},
		},
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
			name:     "path that cleans to dotdot escape rejected",
			input:    []string{"uploads/../../etc/passwd"},
			expected: nil,
		},
		{
			name:     "dotdot that fully escapes root rejected",
			input:    []string{"../"},
			expected: nil,
		},
		{
			name:     "mixed safe and unsafe: only safe paths returned",
			input:    []string{"uploads/safe.txt", "../etc/passwd", "/etc/shadow", "uploads/../../evil"},
			expected: []string{"uploads/safe.txt"},
		},
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
			name:     "backslash normalised to slash for comparison, original returned",
			input:    []string{`uploads\file.txt`},
			expected: []string{`uploads\file.txt`},
		},
		{
			name:     "backslash and slash variants of same path are deduplicated",
			input:    []string{`uploads\file.txt`, "uploads/file.txt"},
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
			assert.Equal(t, tt.expected, DeduplicatePaths(tt.input))
		})
	}
}

func TestFiles_Sort_OrdersByPath(t *testing.T) {
	files := []File{{Name: "b.txt", Path: "b.txt"}, {Name: "a.txt", Path: "a.txt"}, {Name: "c.txt", Path: "c.txt"}}
	Sort(files)

	assert.Equal(t, "a.txt", files[0].Name)
	assert.Equal(t, "b.txt", files[1].Name)
	assert.Equal(t, "c.txt", files[2].Name)
}

func TestFiles_FirstUnsafePath_ReportsTheFirstEscape(t *testing.T) {
	tests := []struct {
		name      string
		paths     []string
		wantPath  string
		wantFound bool
	}{
		{name: "an empty list has none"},
		{name: "safe paths have none", paths: []string{"uploads/a.txt", "container/etc/hosts"}},
		{name: "blank entries are not unsafe", paths: []string{"  ", "uploads/a.txt"}},
		{name: "an absolute path is unsafe", paths: []string{"uploads/a.txt", "/etc/passwd"}, wantPath: "/etc/passwd", wantFound: true},
		{name: "a parent escape is unsafe", paths: []string{"uploads/a.txt", "../../etc/passwd"}, wantPath: "../../etc/passwd", wantFound: true},
		{name: "a bare parent is unsafe", paths: []string{".."}, wantPath: "..", wantFound: true},
		{name: "a backslash-rooted path is unsafe", paths: []string{"\\etc\\passwd"}, wantPath: "\\etc\\passwd", wantFound: true},
		{name: "escape hidden mid-path", paths: []string{"uploads/../../etc"}, wantPath: "uploads/../../etc", wantFound: true},
		{name: "reports the first one", paths: []string{"/a", "/b"}, wantPath: "/a", wantFound: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, found := FirstUnsafePath(tt.paths)
			assert.Equal(t, tt.wantFound, found)
			assert.Equal(t, tt.wantPath, got)
		})
	}
}

// A path the upload check lets through must never be one DeduplicatePaths drops as unsafe, and back.
func TestFiles_FirstUnsafePath_AgreesWithDeduplicatePaths(t *testing.T) {
	tests := []struct {
		name string
		path string
	}{
		{"an absolute path", "/etc/passwd"},
		{"a parent escape", "../escape"},
		{"a bare parent", ".."},
		{"a backslash absolute path", "\\etc\\passwd"},
		{"an escape hidden mid-path", "uploads/../../etc"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, found := FirstUnsafePath([]string{tt.path})
			assert.True(t, found, "FirstUnsafePath did not flag %q", tt.path)
			assert.Empty(t, DeduplicatePaths([]string{tt.path}), "DeduplicatePaths kept %q", tt.path)
		})
	}
}
