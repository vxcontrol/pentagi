//go:generate go run generate.go

package files

import (
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// FileStatus represents file integrity status
type FileStatus string

const (
	FileStatusMissing  FileStatus = "missing"  // file does not exist
	FileStatusModified FileStatus = "modified" // file exists but differs from embedded
	FileStatusOK       FileStatus = "ok"       // file exists and matches embedded
)

// Files provides access to embedded and filesystem files
type Files interface {
	// GetContent returns file content from embedded FS or filesystem fallback
	GetContent(name string) ([]byte, error)

	// Exists checks if file/directory exists in embedded FS
	Exists(name string) bool

	// ExistsInFS checks if file/directory exists in real filesystem
	ExistsInFS(name string) bool

	// Stat returns file info from embedded FS or filesystem fallback
	Stat(name string) (fs.FileInfo, error)

	// Copy copies file/directory from embedded FS to real filesystem
	// dst is target directory, src name is preserved
	Copy(src, dst string, rewrite bool) error

	// Check returns file status comparing embedded vs filesystem
	Check(name string, workingDir string) FileStatus

	// List returns all embedded files with given prefix
	List(prefix string) ([]string, error)
}

// EmbeddedProvider interface for generated embedded filesystem
type EmbeddedProvider interface {
	GetContent(name string) ([]byte, error)
	Exists(name string) bool
	Stat(name string) (fs.FileInfo, error)
	Copy(src, dst string, rewrite bool) error
	List(prefix string) ([]string, error)
	CheckHash(name, workingDir string) (bool, error)
	ExpectedMode(name string) (fs.FileMode, bool)
}

// embeddedProvider holds reference to generated embedded provider
var embeddedProvider EmbeddedProvider = nil

// files implements Files interface with fallback logic
type files struct {
	linksDir string
}

func NewFiles() Files {
	return &files{
		linksDir: "links",
	}
}

// GetContent returns file content from embedded FS or filesystem fallback
func (f *files) GetContent(name string) ([]byte, error) {
	if embeddedProvider != nil {
		if content, err := embeddedProvider.GetContent(name); err == nil {
			return content, nil
		}
	}

	return f.getContentFromFS(name)
}

// Exists checks if file/directory exists in embedded FS
func (f *files) Exists(name string) bool {
	if embeddedProvider != nil {
		return embeddedProvider.Exists(name)
	}
	return false
}

// ExistsInFS checks if file/directory exists in real filesystem
func (f *files) ExistsInFS(name string) bool {
	path := filepath.Join(f.linksDir, name)
	_, err := os.Stat(path)
	return err == nil
}

// Stat returns file info from embedded FS or filesystem fallback
func (f *files) Stat(name string) (fs.FileInfo, error) {
	if embeddedProvider != nil {
		if info, err := embeddedProvider.Stat(name); err == nil {
			return info, nil
		}
	}

	return f.statFromFS(name)
}

// Copy copies file/directory from embedded FS to real filesystem
func (f *files) Copy(src, dst string, rewrite bool) error {
	if embeddedProvider != nil {
		if err := embeddedProvider.Copy(src, dst, rewrite); err == nil {
			return nil
		}
	}

	return f.copyFromFS(src, dst, rewrite)
}

// Check returns file status comparing embedded vs filesystem
func (f *files) Check(name string, workingDir string) FileStatus {
	targetPath := filepath.Join(workingDir, name)

	// check if file exists in filesystem
	if _, err := os.Stat(targetPath); os.IsNotExist(err) {
		return FileStatusMissing
	}

	// try hash-based comparison first (more efficient)
	if embeddedProvider != nil {
		if hashMatch, err := embeddedProvider.CheckHash(name, workingDir); err == nil {
			if hashMatch {
				// hash matches, also verify permission bits if available
				if expectedMode, ok := embeddedProvider.ExpectedMode(name); ok {
					fsInfo, err := os.Stat(targetPath)
					if err != nil {
						return FileStatusMissing
					}
					if fsInfo.Mode().Perm() != expectedMode.Perm() {
						return FileStatusModified
					}
				}
				return FileStatusOK
			}
			// hash didn't match but no error, so it's definitely modified
			return FileStatusModified
		}
		// if hash check failed (file not in metadata, etc.), fall back to content comparison
	}

	// fallback to content comparison
	embeddedContent, err := f.GetContent(name)
	if err != nil {
		// if embedded doesn't exist, filesystem file is OK by default
		return FileStatusOK
	}

	// read filesystem content
	fsContent, err := os.ReadFile(targetPath)
	if err != nil {
		// cannot read filesystem file, consider it missing
		return FileStatusMissing
	}

	// compare contents
	if string(embeddedContent) == string(fsContent) {
		// also compare permission bits when using filesystem fallback
		if infoExpected, err := f.statFromFS(name); err == nil {
			if infoFS, err := os.Stat(targetPath); err == nil {
				if infoFS.Mode().Perm() != infoExpected.Mode().Perm() {
					return FileStatusModified
				}
			}
		}
		return FileStatusOK
	}

	return FileStatusModified
}

// List returns all embedded files with given prefix
func (f *files) List(prefix string) ([]string, error) {
	if embeddedProvider != nil {
		return embeddedProvider.List(prefix)
	}

	// fallback to filesystem listing
	return f.listFromFS(prefix)
}

// getContentFromFS reads file content from real filesystem
func (f *files) getContentFromFS(name string) ([]byte, error) {
	path := filepath.Join(f.linksDir, name)
	return os.ReadFile(path)
}

// statFromFS gets file info from real filesystem
func (f *files) statFromFS(name string) (fs.FileInfo, error) {
	path := filepath.Join(f.linksDir, name)
	return os.Stat(path)
}

// copyFromFS copies file/directory from links directory to destination
func (f *files) copyFromFS(src, dst string, rewrite bool) error {
	srcPath := filepath.Join(f.linksDir, src)
	dstPath := filepath.Join(dst, src)

	srcInfo, err := os.Stat(srcPath)
	if err != nil {
		return err
	}

	if srcInfo.IsDir() {
		return f.copyDirFromFS(srcPath, dstPath, rewrite)
	}

	return f.copyFileFromFS(srcPath, dstPath, rewrite)
}

// copyFileFromFS copies single file
func (f *files) copyFileFromFS(src, dst string, rewrite bool) error {
	if !rewrite {
		if _, err := os.Stat(dst); err == nil {
			return &os.PathError{Op: "copy", Path: dst, Err: os.ErrExist}
		}
	}

	// Ensure destination directory exists
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}

	// read source mode to preserve permissions
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	if _, err = io.Copy(dstFile, srcFile); err != nil {
		return err
	}

	// apply original permissions
	if err := os.Chmod(dst, srcInfo.Mode().Perm()); err != nil {
		return err
	}

	return nil
}

// copyDirFromFS copies directory recursively
func (f *files) copyDirFromFS(src, dst string, rewrite bool) error {
	if !rewrite {
		if _, err := os.Stat(dst); err == nil {
			return &os.PathError{Op: "copy", Path: dst, Err: os.ErrExist}
		}
	}

	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		dstPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(dstPath, info.Mode())
		}

		return f.copyFileFromFS(path, dstPath, rewrite)
	})
}

// listFromFS lists files from filesystem with given prefix
func (f *files) listFromFS(prefix string) ([]string, error) {
	var files []string
	basePath := filepath.Join(f.linksDir, prefix)

	// check if prefix path exists
	if _, err := os.Stat(basePath); os.IsNotExist(err) {
		return files, nil
	}

	err := filepath.Walk(f.linksDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// skip directories
		if info.IsDir() {
			return nil
		}

		// get relative path from links directory
		relPath, err := filepath.Rel(f.linksDir, path)
		if err != nil {
			return err
		}

		// check if path starts with prefix
		if strings.HasPrefix(relPath, prefix) {
			files = append(files, relPath)
		}

		return nil
	})

	return files, err
}
