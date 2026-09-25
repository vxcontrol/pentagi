package files

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// filesLinksFixture returns a Files reading from a fresh links directory that holds
// test.txt, an executable run.sh and testdir/nested.txt.
func filesLinksFixture(t *testing.T) Files {
	t.Helper()
	links := filepath.Join(t.TempDir(), "links")
	filesWrite(t, filepath.Join(links, "test.txt"), "test content", 0o644)
	filesWrite(t, filepath.Join(links, "run.sh"), "#!/bin/sh\necho hi\n", 0o755)
	filesWrite(t, filepath.Join(links, "testdir", "nested.txt"), "nested content", 0o644)
	return &files{linksDir: links}
}

// filesWrite creates path with an exact mode: WriteFile alone is subject to the umask.
func filesWrite(t *testing.T, path, content string, mode os.FileMode) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), mode); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	if err := os.Chmod(path, mode); err != nil {
		t.Fatalf("Chmod: %v", err)
	}
}

// filesEmbedded reads the generated source of an embedded file, independently of the provider.
func filesEmbedded(t *testing.T, name string) []byte {
	t.Helper()
	content, err := os.ReadFile(filepath.Join("fs", name))
	if err != nil {
		t.Fatalf("read generated fs/%s: %v", name, err)
	}
	return content
}

func filesRequireEmbedded(t *testing.T) {
	t.Helper()
	if embeddedProvider == nil {
		t.Skip("embedded provider not generated: run go generate ./cmd/installer/files/")
	}
}

// filesWithoutEmbeddedProvider makes the rest of the test run as a build without go generate.
func filesWithoutEmbeddedProvider(t *testing.T) {
	saved := embeddedProvider
	embeddedProvider = nil
	t.Cleanup(func() { embeddedProvider = saved })
}

func TestFiles_GetContent_ReadsTheEmbeddedFileThenTheLinksDirectory(t *testing.T) {
	t.Run("an embedded file comes from the provider", func(t *testing.T) {
		filesRequireEmbedded(t)
		content, err := NewFiles().GetContent("docker-compose.yml")
		if err != nil || string(content) != string(filesEmbedded(t, "docker-compose.yml")) {
			t.Errorf("GetContent(docker-compose.yml) = %d bytes, %v; want the generated file", len(content), err)
		}
	})
	t.Run("a file only in the links directory falls back to it", func(t *testing.T) {
		content, err := filesLinksFixture(t).GetContent("test.txt")
		if err != nil || string(content) != "test content" {
			t.Errorf("GetContent(test.txt) = %q, %v; want %q", content, err, "test content")
		}
	})
}

func TestFiles_ExistsInFS_ReportsWhatTheLinksDirectoryHolds(t *testing.T) {
	f := filesLinksFixture(t)
	if !f.ExistsInFS("test.txt") {
		t.Error("ExistsInFS(test.txt) = false, want true")
	}
	if f.ExistsInFS("nonexistent.txt") {
		t.Error("ExistsInFS(nonexistent.txt) = true, want false")
	}
}

func TestFiles_Stat_FallsBackToTheLinksDirectory(t *testing.T) {
	info, err := filesLinksFixture(t).Stat("test.txt")
	if err != nil {
		t.Fatalf("Stat() error = %v", err)
	}
	if info.IsDir() || info.Size() != 12 {
		t.Errorf("Stat() = dir %v, size %d; want a file of 12 bytes", info.IsDir(), info.Size())
	}
}

func TestFiles_Copy_WritesTheSourceUnderTheDestination(t *testing.T) {
	tests := []struct {
		name     string
		embedded bool
		src      string
		existing string // content already at the destination
		rewrite  bool
		path     string // file checked under the destination
		content  string
		mode     os.FileMode
		wantErr  error
	}{
		{name: "a linked file keeps its content and mode", src: "run.sh", path: "run.sh", content: "#!/bin/sh\necho hi\n", mode: 0o755},
		{name: "a linked directory is copied recursively", src: "testdir", path: "testdir/nested.txt", content: "nested content"},
		{name: "an existing file is kept without rewrite", src: "test.txt", existing: "existing", path: "test.txt", content: "existing", wantErr: os.ErrExist},
		{name: "an existing file is replaced with rewrite", src: "test.txt", existing: "existing", rewrite: true, path: "test.txt", content: "test content"},
		{name: "an embedded file is copied byte for byte", embedded: true, src: "docker-compose.yml", path: "docker-compose.yml"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, want := filesLinksFixture(t), tt.content
			if tt.embedded {
				filesRequireEmbedded(t)
				f, want = NewFiles(), string(filesEmbedded(t, tt.src))
			}
			dst := t.TempDir()
			if tt.existing != "" {
				filesWrite(t, filepath.Join(dst, tt.src), tt.existing, 0o644)
			}

			if err := f.Copy(tt.src, dst, tt.rewrite); !errors.Is(err, tt.wantErr) {
				t.Fatalf("Copy() error = %v, want %v", err, tt.wantErr)
			}

			target := filepath.Join(dst, filepath.FromSlash(tt.path))
			got, err := os.ReadFile(target)
			if err != nil {
				t.Fatalf("read copied file: %v", err)
			}
			if string(got) != want {
				t.Errorf("copied content = %q, want %q", got, want)
			}
			if tt.mode != 0 && runtime.GOOS != "windows" {
				info, err := os.Stat(target)
				if err != nil {
					t.Fatalf("stat copied file: %v", err)
				}
				if info.Mode().Perm() != tt.mode {
					t.Errorf("copied mode = %v, want %v", info.Mode().Perm(), tt.mode)
				}
			}
		})
	}
}

func TestFiles_Check_ComparesTheWorkingCopyWithTheSource(t *testing.T) {
	permMismatch := FileStatusModified
	if runtime.GOOS == "windows" {
		permMismatch = FileStatusOK // Windows has no permission bits to compare
	}
	copyOf := func(name string) func(*testing.T, Files, string) {
		return func(t *testing.T, f Files, work string) {
			if err := f.Copy(name, work, false); err != nil {
				t.Fatalf("Copy(%q) error = %v", name, err)
			}
		}
	}
	write := func(name, content string) func(*testing.T, Files, string) {
		return func(t *testing.T, _ Files, work string) {
			filesWrite(t, filepath.Join(work, name), content, 0o644)
		}
	}
	copyWithMode := func(name string, mode os.FileMode) func(*testing.T, Files, string) {
		return func(t *testing.T, f Files, work string) {
			copyOf(name)(t, f, work)
			if err := os.Chmod(filepath.Join(work, name), mode); err != nil {
				t.Fatalf("Chmod: %v", err)
			}
		}
	}

	tests := []struct {
		name     string
		embedded bool
		file     string
		setup    func(t *testing.T, f Files, work string)
		want     FileStatus
	}{
		{name: "a file absent from the working directory is missing", file: "test.txt", want: FileStatusMissing},
		{name: "an exact copy of a linked file is ok", file: "test.txt", setup: copyOf("test.txt"), want: FileStatusOK},
		{name: "a linked file with other content is modified", file: "test.txt", setup: write("test.txt", "modified content"), want: FileStatusModified},
		{name: "a linked file with other permissions is modified", file: "run.sh", setup: copyWithMode("run.sh", 0o644), want: permMismatch},
		{name: "an exact copy of an embedded file is ok", embedded: true, file: "docker-compose.yml", setup: copyOf("docker-compose.yml"), want: FileStatusOK},
		{name: "an embedded file with other permissions is modified", embedded: true, file: "docker-compose.yml", setup: copyWithMode("docker-compose.yml", 0o755), want: permMismatch},
		{
			name:     "an embedded file with other bytes of the same size is modified",
			embedded: true,
			file:     "docker-compose.yml",
			setup: func(t *testing.T, f Files, work string) {
				write("docker-compose.yml", strings.Repeat("X", len(filesEmbedded(t, "docker-compose.yml"))))(t, f, work)
			},
			want: FileStatusModified,
		},
		{name: "an embedded file of another size is modified", embedded: true, file: "docker-compose.yml", setup: write("docker-compose.yml", "different size content"), want: FileStatusModified},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := filesLinksFixture(t)
			if tt.embedded {
				filesRequireEmbedded(t)
				// As a released installer runs: no links directory to take an expected mode from.
				f = &files{linksDir: filepath.Join(t.TempDir(), "links")}
			}
			work := t.TempDir()
			if tt.setup != nil {
				tt.setup(t, f, work)
			}
			if got := f.Check(tt.file, work); got != tt.want {
				t.Errorf("Check(%q) = %v, want %v", tt.file, got, tt.want)
			}
		})
	}
}

func TestFiles_Exists_AnswersFromTheEmbeddedProviderOnly(t *testing.T) {
	tests := []struct {
		name       string
		file       string
		noProvider bool
		want       bool
	}{
		{name: "an embedded file exists", file: "docker-compose.yml", want: true},
		{name: "an unknown name does not", file: "any.txt", want: false},
		{name: "nothing exists without an embedded provider", file: "docker-compose.yml", noProvider: true, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.noProvider {
				filesWithoutEmbeddedProvider(t)
			} else {
				filesRequireEmbedded(t)
			}
			if got := NewFiles().Exists(tt.file); got != tt.want {
				t.Errorf("Exists(%q) = %v, want %v", tt.file, got, tt.want)
			}
		})
	}
}

func TestFiles_List_FiltersEmbeddedFilesByPrefix(t *testing.T) {
	filesRequireEmbedded(t)
	f := NewFiles()

	listed, err := f.List("observability")
	if err != nil || len(listed) == 0 {
		t.Fatalf("List(observability) = %v, %v; want files", listed, err)
	}
	for _, name := range listed {
		if !strings.HasPrefix(name, "observability/") {
			t.Errorf("List(observability) returned %q", name)
		}
	}

	if unknown, err := f.List("nonexistent-prefix"); err != nil || len(unknown) != 0 {
		t.Errorf("List(nonexistent-prefix) = %v, %v; want no files", unknown, err)
	}
}
