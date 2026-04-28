package resources

import (
	"archive/zip"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

const (
	ResourcesDirName = "resources"

	invalidBlobHashFileName = ".invalid-blob-hash.blob"

	MaxUploadFileSize    = 300 * 1024 * 1024      // 300 MB
	MaxUploadFiles       = 1000                   // files per request
	MaxUploadTotalSize   = 2 * 1024 * 1024 * 1024 // 2 GB
	MaxUploadRequestSize = MaxUploadTotalSize + 64*1024*1024
	MaxPathLength        = 4096
	MaxFileNameLength    = 255
)

// ResourceEntry is a file or directory record as returned by the package layer.
type ResourceEntry struct {
	Hash      string
	Name      string
	Path      string
	Size      int64
	IsDir     bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

// ZipEntry describes a single file to include in a ZIP archive.
type ZipEntry struct {
	// BlobPath is the absolute path to the .blob file on disk.
	BlobPath string
	// ZipPath is the relative path inside the archive (uses original name / subdirectory structure).
	ZipPath string
}

// ResourcesDir returns the absolute path to the flat blob storage directory.
func ResourcesDir(dataDir string) string {
	return filepath.Join(dataDir, ResourcesDirName)
}

// BlobPath returns the absolute path to the .blob file for a given MD5 hash.
func BlobPath(dataDir, hash string) string {
	cleanHash := strings.ToLower(strings.TrimSpace(hash))
	if !IsValidBlobHash(cleanHash) {
		return filepath.Join(ResourcesDir(dataDir), invalidBlobHashFileName)
	}
	return filepath.Join(ResourcesDir(dataDir), cleanHash+".blob")
}

// EnsureResourcesDir creates the resources storage directory if it does not exist.
func EnsureResourcesDir(dataDir string) error {
	if err := os.MkdirAll(ResourcesDir(dataDir), 0755); err != nil {
		return fmt.Errorf("failed to create resources directory: %w", err)
	}
	return nil
}

// ComputeFileMD5 reads r to EOF and returns the lowercase hex MD5 digest.
func ComputeFileMD5(r io.Reader) (string, error) {
	h := md5.New()
	if _, err := io.Copy(h, r); err != nil {
		return "", fmt.Errorf("failed to compute MD5: %w", err)
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// IsValidBlobHash reports whether hash is a hex-encoded MD5 digest.
func IsValidBlobHash(hash string) bool {
	if len(hash) != md5.Size*2 {
		return false
	}
	for _, r := range hash {
		if (r >= '0' && r <= '9') || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F') {
			continue
		}
		return false
	}
	return true
}

func validateBlobHash(hash string) error {
	if !IsValidBlobHash(hash) {
		return fmt.Errorf("invalid blob hash %q", hash)
	}
	return nil
}

// BlobExists returns true if the .blob file for hash already exists on disk.
func BlobExists(dataDir, hash string) (bool, error) {
	if err := validateBlobHash(hash); err != nil {
		return false, err
	}
	_, err := os.Lstat(BlobPath(dataDir, hash))
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// DeleteBlob removes the .blob file for hash. It is safe to call if the file
// does not exist (returns nil in that case).
func DeleteBlob(dataDir, hash string) error {
	if err := validateBlobHash(hash); err != nil {
		return err
	}
	err := os.Remove(BlobPath(dataDir, hash))
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete blob %s: %w", hash, err)
	}
	return nil
}

// SanitizeResourcePath normalises a client-supplied virtual path and ensures it
// is safe to use:
//   - trims whitespace
//   - converts backslashes to forward slashes
//   - cleans the path (removes .., double slashes, etc.)
//   - rejects absolute paths, dot-only components, and paths that exceed MaxPathLength
//   - returns an error for the empty path
func SanitizeResourcePath(p string) (string, error) {
	trimmed := strings.TrimSpace(p)
	if trimmed == "" {
		return "", fmt.Errorf("path must not be empty")
	}
	if len(trimmed) > MaxPathLength {
		return "", fmt.Errorf("path exceeds maximum allowed length of %d characters", MaxPathLength)
	}

	normalized := strings.ReplaceAll(trimmed, "\\", "/")
	if strings.HasPrefix(normalized, "/") {
		return "", fmt.Errorf("path must be relative")
	}
	for _, part := range strings.Split(normalized, "/") {
		if part == ".." {
			return "", fmt.Errorf("path must not contain parent directory traversal")
		}
	}
	cleaned := path.Clean("/" + normalized)
	// Remove the leading "/" we added for Clean, making the path relative.
	rel := strings.TrimPrefix(cleaned, "/")
	if rel == "" || rel == "." {
		return "", fmt.Errorf("invalid path")
	}

	// Validate every path component.
	parts := strings.Split(rel, "/")
	for _, part := range parts {
		if err := validatePathComponent(part); err != nil {
			return "", err
		}
	}

	return rel, nil
}

// SanitizeResourceDir is like SanitizeResourcePath but also accepts an empty
// string to mean "root". It returns "" for root, or a clean relative path.
func SanitizeResourceDir(p string) (string, error) {
	if strings.TrimSpace(p) == "" {
		return "", nil
	}
	return SanitizeResourcePath(p)
}

// SanitizeResourceFileName strips path separators and validates the basename.
func SanitizeResourceFileName(fileName string) (string, error) {
	trimmed := strings.TrimSpace(fileName)
	if trimmed == "" {
		return "", fmt.Errorf("file name is required")
	}

	normalized := strings.ReplaceAll(trimmed, "\\", "/")
	cleanName := path.Base(path.Clean("/" + normalized))

	if err := validatePathComponent(cleanName); err != nil {
		return "", err
	}
	return cleanName, nil
}

// FilePath builds the virtual file path for a file named name inside dir.
// dir may be "" (root).
func FilePath(dir, name string) string {
	if dir == "" {
		return name
	}
	return dir + "/" + name
}

// ParentDir returns the parent directory of a virtual path, or "" for root.
func ParentDir(p string) string {
	parent := path.Dir(p)
	if parent == "." || parent == "/" || parent == "" {
		return ""
	}
	return parent
}

// FileName returns the final component of a virtual path.
func FileName(p string) string {
	return path.Base(p)
}

// IsChildOf returns true when candidate is a direct child of dir (not a deeper
// descendant).  dir == "" means root.
func IsChildOf(candidate, dir string) bool {
	if dir == "" {
		return !strings.Contains(candidate, "/")
	}
	prefix := dir + "/"
	if !strings.HasPrefix(candidate, prefix) {
		return false
	}
	rest := candidate[len(prefix):]
	return !strings.Contains(rest, "/")
}

// PathHasPrefix returns true when p equals dir or starts with dir+"/".
// dir == "" matches everything.
func PathHasPrefix(p, dir string) bool {
	if dir == "" {
		return true
	}
	return p == dir || strings.HasPrefix(p, dir+"/")
}

// ReplacePrefixPath replaces the leading oldPrefix in p with newPrefix.
// The prefix must match exactly; the function panics if it does not.
func ReplacePrefixPath(p, oldPrefix, newPrefix string) string {
	if oldPrefix == "" {
		if newPrefix == "" {
			return p
		}
		return newPrefix + "/" + p
	}
	if p == oldPrefix {
		return newPrefix
	}
	rest := strings.TrimPrefix(p, oldPrefix+"/")
	if newPrefix == "" {
		return rest
	}
	return newPrefix + "/" + rest
}

// EscapeLike escapes special LIKE pattern characters (%, _) in s so the string
// can be safely embedded in a SQL LIKE clause.
func EscapeLike(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `%`, `\%`)
	s = strings.ReplaceAll(s, `_`, `\_`)
	return s
}

// SaveToTemp writes r into a new temporary file in dir and returns the path to
// the temp file. The caller is responsible for removing the temp file on error.
func SaveToTemp(r io.Reader, dir string) (tmpPath string, hash string, size int64, err error) {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", "", 0, fmt.Errorf("failed to create temp directory: %w", err)
	}

	tmp, err := os.CreateTemp(dir, ".resource-upload-*")
	if err != nil {
		return "", "", 0, fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpPath = tmp.Name()
	defer tmp.Close()

	h := md5.New()
	mw := io.MultiWriter(tmp, h)
	written, copyErr := io.Copy(mw, r)
	if copyErr != nil {
		os.Remove(tmpPath)
		return "", "", 0, fmt.Errorf("failed to write temp file: %w", copyErr)
	}
	if err := tmp.Chmod(0644); err != nil {
		os.Remove(tmpPath)
		return "", "", 0, fmt.Errorf("failed to set temp file permissions: %w", err)
	}

	return tmpPath, hex.EncodeToString(h.Sum(nil)), written, nil
}

// CommitBlob atomically moves tmpPath to the .blob destination for hash.  If
// the blob already exists (race with concurrent upload of identical file) the
// tmp file is removed and no error is returned.
func CommitBlob(dataDir, hash, tmpPath string) error {
	if err := validateBlobHash(hash); err != nil {
		return err
	}
	if err := EnsureResourcesDir(dataDir); err != nil {
		return err
	}

	dest := BlobPath(dataDir, hash)
	if _, err := os.Lstat(dest); err == nil {
		// Already exists — remove tmp and consider success.
		os.Remove(tmpPath)
		return nil
	}
	if err := os.Rename(tmpPath, dest); err != nil {
		return fmt.Errorf("failed to commit blob %s: %w", hash, err)
	}
	return nil
}

// ZipResources writes a ZIP archive to w containing all entries in files.
// Each ZipEntry maps a .blob file on disk to a path inside the archive.
func ZipResources(w io.Writer, entries []ZipEntry) error {
	zw := zip.NewWriter(w)
	defer zw.Close()

	for _, e := range entries {
		info, err := os.Lstat(e.BlobPath)
		if err != nil {
			return fmt.Errorf("failed to stat blob %s: %w", e.BlobPath, err)
		}
		if !info.Mode().IsRegular() {
			continue
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return fmt.Errorf("failed to create zip header for %s: %w", e.ZipPath, err)
		}
		zipPath, err := SanitizeResourcePath(e.ZipPath)
		if err != nil {
			return fmt.Errorf("invalid zip entry path %q: %w", e.ZipPath, err)
		}
		header.Name = filepath.ToSlash(zipPath)
		header.Method = zip.Deflate

		zf, err := zw.CreateHeader(header)
		if err != nil {
			return fmt.Errorf("failed to create zip entry %s: %w", e.ZipPath, err)
		}

		f, err := os.Open(e.BlobPath)
		if err != nil {
			return fmt.Errorf("failed to open blob %s: %w", e.BlobPath, err)
		}
		_, copyErr := io.Copy(zf, f)
		f.Close()
		if copyErr != nil {
			return fmt.Errorf("failed to write zip entry %s: %w", e.ZipPath, copyErr)
		}
	}
	return nil
}

// ZipDirectory is kept for symmetry with flowfiles and wraps ZipResources by
// walking a real FS directory.  In the resources service, use ZipResources
// instead since files are stored as blobs, not in a directory structure.
func ZipDirectory(w io.Writer, dirPath string) error {
	zw := zip.NewWriter(w)
	defer zw.Close()

	return filepath.WalkDir(dirPath, func(entryPath string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || d.Type()&os.ModeSymlink != 0 || !d.Type().IsRegular() {
			return nil
		}

		rel, err := filepath.Rel(dirPath, entryPath)
		if err != nil {
			return fmt.Errorf("failed to get relative path: %w", err)
		}
		info, err := d.Info()
		if err != nil {
			return fmt.Errorf("failed to get file info: %w", err)
		}
		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return fmt.Errorf("failed to create zip header: %w", err)
		}
		zipPath, err := SanitizeResourcePath(filepath.ToSlash(rel))
		if err != nil {
			return fmt.Errorf("invalid zip entry path %q: %w", rel, err)
		}
		header.Name = filepath.ToSlash(zipPath)
		header.Method = zip.Deflate

		zf, err := zw.CreateHeader(header)
		if err != nil {
			return fmt.Errorf("failed to create zip entry: %w", err)
		}
		f, err := os.Open(entryPath)
		if err != nil {
			return fmt.Errorf("failed to open file: %w", err)
		}
		_, copyErr := io.Copy(zf, f)
		f.Close()
		return copyErr
	})
}

// validatePathComponent rejects empty, dot/dotdot, overly long, or
// character-unsafe path segments.  Mirrors flowfiles.validatePathComponent.
func validatePathComponent(component string) error {
	clean := strings.TrimSpace(component)
	if clean == "" || clean == "." || clean == ".." {
		return fmt.Errorf("invalid path component %q", component)
	}
	if len(clean) > MaxFileNameLength {
		return fmt.Errorf("path component %q exceeds maximum length of %d", component, MaxFileNameLength)
	}
	for _, r := range clean {
		if r < 0x20 || r == 0x7f {
			return fmt.Errorf("path component contains control characters")
		}
		switch r {
		case '/', '\\', ':', '*', '?', '"', '<', '>', '|':
			return fmt.Errorf("path component contains unsupported character %q", r)
		}
	}
	return nil
}
