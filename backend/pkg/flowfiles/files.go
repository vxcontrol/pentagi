package flowfiles

import (
	"archive/tar"
	"archive/zip"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"mime/multipart"
	"os"
	"path"
	"path/filepath"
	"pentagi/pkg/docker"
	"sort"
	"strings"
	"time"
)

const (
	UploadsDirName   = "uploads"
	ContainerDirName = "container"
	ResourcesDirName = "resources"

	MaxUploadFileSize    = 300 * 1024 * 1024      // 300 MB
	MaxUploadFiles       = 1000                   // files
	MaxUploadTotalSize   = 2 * 1024 * 1024 * 1024 // 2 GB
	MaxUploadRequestSize = MaxUploadTotalSize + 64*1024*1024
	MaxPullFiles         = 1000                   // files
	MaxPullTotalSize     = 2 * 1024 * 1024 * 1024 // 2 GB
	MaxFileNameLength    = 255
)

type File struct {
	ID         string
	Name       string
	Path       string
	Size       int64
	IsDir      bool
	ModifiedAt time.Time
}

type Files struct {
	Files []File
	Total uint64
}

// TarEntry describes a local regular file to include in a TAR archive.
type TarEntry struct {
	LocalPath string
	TarPath   string
}

func List(dataDir string, flowID uint64) (Files, error) {
	uploadsEntries, err := ListDirEntries(FlowUploadsDir(dataDir, flowID), UploadsDirName)
	if err != nil {
		return Files{}, fmt.Errorf("reading uploads cache: %w", err)
	}

	containerEntries, err := ListDirEntriesRecursive(FlowContainerDir(dataDir, flowID), ContainerDirName)
	if err != nil {
		return Files{}, fmt.Errorf("reading container cache: %w", err)
	}

	resourcesEntries, err := ListDirEntriesRecursive(FlowResourcesDir(dataDir, flowID), ResourcesDirName)
	if err != nil {
		return Files{}, fmt.Errorf("reading resources cache: %w", err)
	}

	files := append(uploadsEntries, containerEntries...)
	files = append(files, resourcesEntries...)
	Sort(files)
	return Files{
		Files: files,
		Total: uint64(len(files)),
	}, nil
}

func FlowDataDir(dataDir string, flowID uint64) string {
	return filepath.Join(dataDir, fmt.Sprintf("flow-%d-data", flowID))
}

func FlowUploadsDir(dataDir string, flowID uint64) string {
	return filepath.Join(FlowDataDir(dataDir, flowID), UploadsDirName)
}

func FlowContainerDir(dataDir string, flowID uint64) string {
	return filepath.Join(FlowDataDir(dataDir, flowID), ContainerDirName)
}

func FlowResourcesDir(dataDir string, flowID uint64) string {
	return filepath.Join(FlowDataDir(dataDir, flowID), ResourcesDirName)
}

func ResolveCachedPath(dataDir string, flowID uint64, reqPath string) (string, error) {
	if strings.TrimSpace(reqPath) == "" {
		return "", errors.New("path query parameter is required")
	}

	cleaned := filepath.Clean(filepath.FromSlash(strings.ReplaceAll(reqPath, "\\", "/")))
	if filepath.IsAbs(cleaned) {
		return "", fmt.Errorf("path must be relative (no leading /)")
	}

	parts := strings.SplitN(cleaned, string(filepath.Separator), 2)
	if parts[0] != UploadsDirName && parts[0] != ContainerDirName && parts[0] != ResourcesDirName {
		return "", fmt.Errorf("path must start with '%s', '%s', or '%s'", UploadsDirName, ContainerDirName, ResourcesDirName)
	}

	flowDataDir := FlowDataDir(dataDir, flowID)
	absPath := filepath.Join(flowDataDir, cleaned)
	if !IsWithinDir(absPath, flowDataDir) {
		return "", fmt.Errorf("path escapes the flow data directory")
	}

	return absPath, nil
}

func SanitizeFileName(fileName string) (string, error) {
	trimmedName := strings.TrimSpace(fileName)
	if trimmedName == "" {
		return "", fmt.Errorf("file name is required")
	}

	normalizedName := strings.ReplaceAll(trimmedName, "\\", "/")
	cleanName := path.Base(path.Clean("/" + normalizedName))

	return validatePathComponent(cleanName)
}

func SanitizeContainerCachePath(containerPath string) (string, error) {
	trimmedPath := strings.TrimSpace(containerPath)
	if trimmedPath == "" {
		return "", fmt.Errorf("path is required")
	}

	normalizedPath := strings.ReplaceAll(trimmedPath, "\\", "/")
	cleanPath := strings.TrimPrefix(path.Clean("/"+normalizedPath), "/")
	if cleanPath == "." || cleanPath == "" {
		return "", fmt.Errorf("invalid path")
	}

	parts := strings.Split(cleanPath, "/")
	for i, part := range parts {
		cleanPart, err := validatePathComponent(part)
		if err != nil {
			return "", fmt.Errorf("invalid path component '%s': %w", part, err)
		}
		parts[i] = cleanPart
	}

	return path.Join(parts...), nil
}

func validatePathComponent(component string) (string, error) {
	cleanName := strings.TrimSpace(component)
	if cleanName == "." || cleanName == ".." || cleanName == "/" || cleanName == "" {
		return "", fmt.Errorf("invalid file name")
	}
	if len(cleanName) > MaxFileNameLength {
		return "", fmt.Errorf("file name is too long")
	}
	for _, r := range cleanName {
		if r < 0x20 || r == 0x7f {
			return "", fmt.Errorf("file name contains control characters")
		}
		switch r {
		case '/', '\\', ':', '*', '?', '"', '<', '>', '|':
			return "", fmt.Errorf("file name contains unsupported characters")
		}
	}

	return cleanName, nil
}

func NewFile(info os.FileInfo, sourceDir string) File {
	return NewFileWithPath(info, path.Join(sourceDir, info.Name()))
}

func NewFileWithPath(info os.FileInfo, filePath string) File {
	return File{
		ID:         ID(filePath),
		Name:       info.Name(),
		Path:       filePath,
		Size:       info.Size(),
		IsDir:      info.IsDir(),
		ModifiedAt: info.ModTime(),
	}
}

func ID(filePath string) string {
	sum := md5.Sum([]byte(filePath))
	return hex.EncodeToString(sum[:])
}

func ListDirEntries(dir, sourceDir string) ([]File, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}

	files := make([]File, 0, len(entries))
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), ".upload-") || strings.HasPrefix(entry.Name(), ".pull-") {
			continue
		}
		if entry.Type()&os.ModeSymlink != 0 {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			return nil, err
		}

		if !info.Mode().IsRegular() && !info.IsDir() {
			continue
		}

		files = append(files, NewFile(info, sourceDir))
	}

	return files, nil
}

func ListDirEntriesRecursive(dir, sourceDir string) ([]File, error) {
	if _, err := os.Lstat(dir); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}

	var files []File
	err := filepath.WalkDir(dir, func(entryPath string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entryPath == dir {
			return nil
		}
		if strings.HasPrefix(entry.Name(), ".pull-") {
			if entry.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if entry.Type()&os.ModeSymlink != 0 {
			return nil
		}

		info, err := entry.Info()
		if err != nil {
			return err
		}
		if !info.Mode().IsRegular() && !info.IsDir() {
			return nil
		}

		rel, err := filepath.Rel(dir, entryPath)
		if err != nil {
			return fmt.Errorf("failed to resolve relative cache path: %w", err)
		}
		files = append(files, NewFileWithPath(info, path.Join(sourceDir, filepath.ToSlash(rel))))
		return nil
	})
	if err != nil {
		return nil, err
	}

	return files, nil
}

func Sort(files []File) {
	sort.Slice(files, func(i, j int) bool {
		if files[i].Path < files[j].Path {
			return true
		} else if files[i].Path > files[j].Path {
			return false
		} else if files[i].Name < files[j].Name {
			return true
		} else if files[i].Name > files[j].Name {
			return false
		}
		return files[i].ModifiedAt.After(files[j].ModifiedAt)
	})
}

func LocalEntryExists(filePath string) (bool, error) {
	if _, err := os.Lstat(filePath); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

func RegularFileInfo(filePath string) (os.FileInfo, error) {
	info, err := os.Lstat(filePath)
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() {
		return nil, fmt.Errorf("'%s' is not a regular file", filePath)
	}

	return info, nil
}

func SaveUploadedFileToTemp(fh *multipart.FileHeader, dir string) (string, error) {
	src, err := fh.Open()
	if err != nil {
		return "", fmt.Errorf("failed to open uploaded file: %w", err)
	}
	defer src.Close()

	dst, err := os.CreateTemp(dir, ".upload-*")
	if err != nil {
		return "", fmt.Errorf("failed to create temporary upload file: %w", err)
	}
	tmpPath := dst.Name()
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		os.Remove(tmpPath)
		return "", fmt.Errorf("failed to write temporary upload file: %w", err)
	}
	if err := dst.Chmod(0644); err != nil {
		os.Remove(tmpPath)
		return "", fmt.Errorf("failed to set temporary upload file permissions: %w", err)
	}

	return tmpPath, nil
}

func IsWithinDir(absPath, dir string) bool {
	return strings.HasPrefix(
		filepath.Clean(absPath)+string(filepath.Separator),
		filepath.Clean(dir)+string(filepath.Separator),
	)
}

func ResolvePulledStagedTarget(stagingDir, cacheRelPath string) string {
	candidates := []string{
		filepath.Join(stagingDir, filepath.FromSlash(cacheRelPath)),
		filepath.Join(stagingDir, path.Base(cacheRelPath)),
	}

	for _, candidate := range candidates {
		if _, err := os.Lstat(candidate); err == nil {
			return candidate
		}
	}

	return ""
}

func WriteUploadsTar(w *io.PipeWriter, uploadDir string) error {
	return writeDirectoryTar(w, uploadDir, UploadsDirName, "upload", "uploads")
}

func ExtractTar(r io.Reader, destDir string) error {
	tr := tar.NewReader(r)
	var filesCount int
	var totalSize int64
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read tar entry: %w", err)
		}
		if hdr.Typeflag == tar.TypeSymlink || hdr.Typeflag == tar.TypeLink {
			continue
		}

		entryPath := filepath.Join(destDir, filepath.Clean(filepath.FromSlash(hdr.Name)))
		if !IsWithinDir(entryPath, destDir) {
			continue
		}

		switch hdr.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(entryPath, 0755); err != nil {
				return fmt.Errorf("failed to create directory '%s': %w", entryPath, err)
			}
		case tar.TypeReg, tar.TypeRegA:
			if hdr.Size < 0 {
				return fmt.Errorf("tar entry '%s' has invalid size %d", hdr.Name, hdr.Size)
			}
			filesCount++
			if filesCount > MaxPullFiles {
				return fmt.Errorf("tar archive exceeds maximum file count of %d", MaxPullFiles)
			}
			totalSize += hdr.Size
			if totalSize > MaxPullTotalSize {
				return fmt.Errorf("tar archive exceeds maximum total size of %d bytes", MaxPullTotalSize)
			}

			if err := os.MkdirAll(filepath.Dir(entryPath), 0755); err != nil {
				return fmt.Errorf("failed to create parent directory for '%s': %w", entryPath, err)
			}

			f, err := os.OpenFile(entryPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
			if err != nil {
				return fmt.Errorf("failed to create file '%s': %w", entryPath, err)
			}
			_, copyErr := io.CopyN(f, tr, hdr.Size)
			f.Close()
			if copyErr != nil {
				return fmt.Errorf("failed to write file '%s': %w", entryPath, copyErr)
			}
		}
	}

	return nil
}

func ZipDirectory(w io.Writer, dirPath string) error {
	zw := zip.NewWriter(w)
	defer zw.Close()

	return filepath.WalkDir(dirPath, func(entryPath string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if d.Type()&os.ModeSymlink != 0 || !d.Type().IsRegular() {
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

		return zipWriteFile(zw, entryPath, filepath.ToSlash(rel), info)
	})
}

// ZipRelativePaths writes a set of cache-relative paths into a single ZIP
// stream. Each relPath is resolved against baseDir:
//   - Regular files are stored under their relPath (e.g. "uploads/file.txt").
//   - Directories are walked recursively; each nested regular file is stored
//     under relPath + "/" + path-relative-to-dir (e.g. "container/etc/nginx.conf").
//   - Symlinks and other special entries are silently skipped.
//   - Missing entries are silently skipped to tolerate concurrent cache changes.
//
// The caller is responsible for deduplicating relPaths before calling this
// function; no internal deduplication of ZIP entry names is performed.
func ZipRelativePaths(w io.Writer, baseDir string, relPaths []string) error {
	zw := zip.NewWriter(w)
	defer zw.Close()

	for _, relPath := range relPaths {
		localPath := filepath.Join(baseDir, filepath.FromSlash(relPath))

		info, err := os.Lstat(localPath)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				continue
			}
			return fmt.Errorf("stat %q: %w", relPath, err)
		}
		if info.Mode()&os.ModeSymlink != 0 {
			continue
		}

		if info.Mode().IsRegular() {
			if err := zipWriteFile(zw, localPath, relPath, info); err != nil {
				return err
			}
			continue
		}

		if !info.IsDir() {
			continue
		}

		if err := filepath.WalkDir(localPath, func(entryPath string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.Type()&os.ModeSymlink != 0 || d.IsDir() || !d.Type().IsRegular() {
				return nil
			}
			rel, err := filepath.Rel(localPath, entryPath)
			if err != nil {
				return fmt.Errorf("relative path: %w", err)
			}
			fi, err := d.Info()
			if err != nil {
				return err
			}
			return zipWriteFile(zw, entryPath, path.Join(relPath, filepath.ToSlash(rel)), fi)
		}); err != nil {
			return fmt.Errorf("walking %q: %w", relPath, err)
		}
	}

	return nil
}

// zipWriteFile adds a single regular file to an open zip.Writer.
func zipWriteFile(zw *zip.Writer, localPath, zipName string, info os.FileInfo) error {
	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return fmt.Errorf("failed to create zip header for %q: %w", zipName, err)
	}
	header.Name = zipName
	header.Method = zip.Deflate

	zf, err := zw.CreateHeader(header)
	if err != nil {
		return fmt.Errorf("failed to create zip entry %q: %w", zipName, err)
	}

	f, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("failed to open %q for zip: %w", localPath, err)
	}
	_, copyErr := io.Copy(zf, f)
	f.Close()
	return copyErr
}

// ResourceRef is a lightweight reference to a user resource for copying into a flow.
type ResourceRef struct {
	Hash        string // MD5 hex — name of the .blob file in the resources store
	VirtualPath string // virtual path (relative) used to restore directory hierarchy
	Name        string // display name
	IsDir       bool
}

// CopyResourcesToFlow copies user resource blobs into flow-{id}-data/resources/,
// restoring the virtual directory hierarchy from VirtualPath.
// Already-present files are skipped (idempotent); when force is true, existing files are replaced.
// Returns relative paths ("resources/<virtualPath>") of newly added or replaced files.
func CopyResourcesToFlow(dataDir, resourcesStoreDir string, flowID uint64, refs []ResourceRef, force bool) ([]string, error) {
	resourcesDir := FlowResourcesDir(dataDir, flowID)
	var added []string

	for _, ref := range refs {
		if ref.IsDir {
			dst := filepath.Join(resourcesDir, filepath.FromSlash(ref.VirtualPath))
			if !IsWithinDir(dst, resourcesDir) {
				return added, fmt.Errorf("resource path '%s' escapes resources directory", ref.VirtualPath)
			}
			if err := os.MkdirAll(dst, 0755); err != nil {
				return added, fmt.Errorf("failed to create resource directory '%s': %w", ref.VirtualPath, err)
			}
			continue
		}

		dst := filepath.Join(resourcesDir, filepath.FromSlash(ref.VirtualPath))
		if !IsWithinDir(dst, resourcesDir) {
			return added, fmt.Errorf("resource path '%s' escapes resources directory", ref.VirtualPath)
		}

		if _, err := os.Lstat(dst); err == nil {
			if !force {
				continue // already present, skip
			}
			if err := os.Remove(dst); err != nil {
				return added, fmt.Errorf("failed to remove existing resource '%s': %w", ref.VirtualPath, err)
			}
		}

		if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
			return added, fmt.Errorf("failed to create parent directory for resource '%s': %w", ref.VirtualPath, err)
		}

		src := filepath.Join(resourcesStoreDir, ref.Hash+".blob")
		if err := copyFile(src, dst); err != nil {
			return added, fmt.Errorf("failed to copy resource '%s': %w", ref.VirtualPath, err)
		}

		added = append(added, path.Join(ResourcesDirName, filepath.ToSlash(ref.VirtualPath)))
	}

	return added, nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer in.Close()

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return fmt.Errorf("failed to copy file content: %w", err)
	}
	return nil
}

// WriteResourcesTar writes flow-{id}-data/resources/ as a tar stream
// with entries prefixed as "resources/<path>", mirroring WriteUploadsTar.
func WriteResourcesTar(w *io.PipeWriter, resourcesDir string) error {
	return writeDirectoryTar(w, resourcesDir, ResourcesDirName, "resource", "resources")
}

func writeDirectoryTar(w *io.PipeWriter, rootDir, tarRootName, entryLabel, cacheLabel string) error {
	tw := tar.NewWriter(w)
	defer w.Close()
	defer tw.Close()

	if err := tw.WriteHeader(&tar.Header{
		Typeflag: tar.TypeDir,
		Name:     tarRootName,
		Mode:     0755,
		ModTime:  time.Now(),
	}); err != nil {
		w.CloseWithError(err)
		return fmt.Errorf("failed to write %s directory tar header: %w", cacheLabel, err)
	}

	var filesCount int
	var totalSize int64
	return filepath.WalkDir(rootDir, func(entryPath string, d os.DirEntry, err error) error {
		if err != nil {
			w.CloseWithError(err)
			return err
		}
		if entryPath == rootDir {
			return nil
		}
		if d.Type()&os.ModeSymlink != 0 {
			return nil
		}

		info, err := d.Info()
		if err != nil {
			w.CloseWithError(err)
			return err
		}
		if !info.Mode().IsRegular() && !info.IsDir() {
			return nil
		}

		rel, err := filepath.Rel(rootDir, entryPath)
		if err != nil {
			w.CloseWithError(err)
			return fmt.Errorf("failed to get %s relative path: %w", entryLabel, err)
		}
		headerName := path.Join(tarRootName, filepath.ToSlash(rel))

		if info.IsDir() {
			if err := tw.WriteHeader(&tar.Header{
				Typeflag: tar.TypeDir,
				Name:     headerName,
				Mode:     int64(info.Mode().Perm()),
				ModTime:  info.ModTime(),
			}); err != nil {
				w.CloseWithError(err)
				return fmt.Errorf("failed to write %s directory tar header: %w", entryLabel, err)
			}
			return nil
		}

		filesCount++
		if filesCount > MaxUploadFiles {
			err := fmt.Errorf("%s cache exceeds maximum file count of %d", cacheLabel, MaxUploadFiles)
			w.CloseWithError(err)
			return err
		}
		totalSize += info.Size()
		if totalSize > MaxUploadTotalSize {
			err := fmt.Errorf("%s cache exceeds maximum total size of %d bytes", cacheLabel, MaxUploadTotalSize)
			w.CloseWithError(err)
			return err
		}

		if err := tw.WriteHeader(&tar.Header{
			Typeflag: tar.TypeReg,
			Name:     headerName,
			Mode:     int64(info.Mode().Perm()),
			Size:     info.Size(),
			ModTime:  info.ModTime(),
		}); err != nil {
			w.CloseWithError(err)
			return fmt.Errorf("failed to write %s file tar header: %w", entryLabel, err)
		}

		f, err := os.Open(entryPath)
		if err != nil {
			w.CloseWithError(err)
			return fmt.Errorf("failed to open %s file: %w", entryLabel, err)
		}
		defer f.Close()

		if _, err := io.Copy(tw, f); err != nil {
			w.CloseWithError(err)
			return fmt.Errorf("failed to write %s file tar content: %w", entryLabel, err)
		}

		return nil
	})
}

// FileListingForPrompt returns a compact XML block listing user-accessible files in
// uploads/ and resources/ for injection into agent system prompts.
// Each tag carries a base attribute with the container-side root path; individual
// entries are slash-separated relative paths (no repeated prefix per file).
// Returns empty string when both directories are empty or absent.
func FileListingForPrompt(dataDir string, flowID uint64) string {
	uploadsBase := filepath.Join(docker.WorkFolderPathInContainer, UploadsDirName)
	uploadFiles := collectRelativeFilePaths(FlowUploadsDir(dataDir, flowID))
	resourcesBase := filepath.Join(docker.WorkFolderPathInContainer, ResourcesDirName)
	resourceFiles := collectRelativeFilePaths(FlowResourcesDir(dataDir, flowID))

	if len(uploadFiles) == 0 && len(resourceFiles) == 0 {
		return ""
	}

	var b strings.Builder
	b.WriteString("<task_files>\n")

	if len(uploadFiles) > 0 {
		b.WriteString("<uploads base=\"")
		b.WriteString(uploadsBase)
		b.WriteString("\">\n")
		for _, p := range uploadFiles {
			b.WriteString(p)
			b.WriteByte('\n')
		}
		b.WriteString("</uploads>\n")
	}

	if len(resourceFiles) > 0 {
		b.WriteString("<resources base=\"")
		b.WriteString(resourcesBase)
		b.WriteString("\">\n")
		for _, p := range resourceFiles {
			b.WriteString(p)
			b.WriteByte('\n')
		}
		b.WriteString("</resources>\n")
	}

	b.WriteString("</task_files>")
	return b.String()
}

// collectRelativeFilePaths walks dir and returns slash-separated paths relative to dir.
// Symlinks and directories are skipped; only regular files are included.
func collectRelativeFilePaths(dir string) []string {
	if _, err := os.Lstat(dir); err != nil {
		return nil
	}

	var paths []string
	_ = filepath.WalkDir(dir, func(entryPath string, d fs.DirEntry, err error) error {
		if err != nil || entryPath == dir {
			return nil
		}
		if d.Type()&os.ModeSymlink != 0 || d.IsDir() {
			return nil
		}
		if !d.Type().IsRegular() {
			return nil
		}
		rel, err := filepath.Rel(dir, entryPath)
		if err != nil {
			return nil
		}
		paths = append(paths, filepath.ToSlash(rel))
		return nil
	})
	return paths
}

// DeduplicatePaths returns a deduplicated, coverage-minimised slice of the
// original input paths, preserving first-occurrence order. The rules applied:
//
//  1. Whitespace-only entries are dropped.
//  2. Entries whose path.Clean form is an absolute path or starts with "../"
//     are silently dropped (path-traversal safety).
//  3. Paths whose path.Clean forms are equal are deduplicated; the first
//     occurrence (original value) wins.
//  4. A path that is a descendant of another surviving path is dropped because
//     the ancestor already covers it. For example, when both "uploads/dir" and
//     "uploads/dir/file.txt" are present, only "uploads/dir" is kept.
//
// The "/" suffix trick is used for the ancestor check to ensure that
// "uploads/my_dir" does not erroneously cover "uploads/my_dir_extra".
//
// Original (un-normalised) values of surviving paths are returned.
func DeduplicatePaths(paths []string) []string {
	type entry struct {
		original string
		clean    string // path.Clean result used only for comparison
	}

	entries := make([]entry, 0, len(paths))
	seen := make(map[string]struct{}, len(paths))

	for _, p := range paths {
		trimmed := strings.TrimSpace(p)
		if trimmed == "" {
			continue
		}

		// Normalise for comparison (backslashes → forward slashes, then Clean).
		normalized := path.Clean(strings.ReplaceAll(trimmed, "\\", "/"))

		// Safety: reject absolute paths and anything that escapes via "..".
		if path.IsAbs(normalized) || normalized == ".." || strings.HasPrefix(normalized, "../") {
			continue
		}

		// Exact-path deduplication (after normalisation).
		if _, dup := seen[normalized]; dup {
			continue
		}
		seen[normalized] = struct{}{}

		entries = append(entries, entry{original: trimmed, clean: normalized})
	}

	if len(entries) == 0 {
		return nil
	}

	// Coverage deduplication: drop entry B if any other entry A is an ancestor
	// of B. A is an ancestor when B's clean path with "/" appended starts with
	// A's clean path with "/" appended:
	//   strings.HasPrefix(cleanB+"/", cleanA+"/")
	// The "/" suffix on both sides prevents "uploads/dir_x" from matching
	// "uploads/dir" as a parent.
	result := make([]string, 0, len(entries))
	for i, b := range entries {
		bPrefix := b.clean + "/"
		covered := false
		for j, a := range entries {
			if i == j {
				continue
			}
			if strings.HasPrefix(bPrefix, a.clean+"/") {
				covered = true
				break
			}
		}
		if !covered {
			result = append(result, b.original)
		}
	}

	return result
}

// BaseName returns the last path component of a slash-separated path.
func BaseName(p string) string {
	for i := len(p) - 1; i >= 0; i-- {
		if p[i] == '/' || p[i] == '\\' {
			return p[i+1:]
		}
	}
	return p
}

// WriteSingleFileTar writes a single regular file as a tar stream.
// tarName is the path inside the tar archive (e.g. "resources/creds/pass.txt").
func WriteSingleFileTar(w *io.PipeWriter, absPath, tarName string) error {
	return WriteFilesTar(w, []TarEntry{{LocalPath: absPath, TarPath: tarName}})
}

// WriteFilesTar writes regular files into a tar stream and emits parent
// directory headers once per archive. Missing files are skipped to tolerate
// concurrent cache changes between listing and copying.
func WriteFilesTar(w *io.PipeWriter, entries []TarEntry) error {
	tw := tar.NewWriter(w)
	defer w.Close()
	defer tw.Close()

	createdDirs := make(map[string]bool)
	for _, entry := range entries {
		tarName := path.Clean(strings.ReplaceAll(entry.TarPath, "\\", "/"))
		if tarName == "." || strings.HasPrefix(tarName, "../") || strings.HasPrefix(tarName, "/") {
			err := fmt.Errorf("invalid tar entry path '%s'", entry.TarPath)
			w.CloseWithError(err)
			return err
		}

		info, err := os.Lstat(entry.LocalPath)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				continue
			}
			w.CloseWithError(err)
			return fmt.Errorf("failed to stat file '%s': %w", entry.LocalPath, err)
		}
		if !info.Mode().IsRegular() {
			continue
		}

		if err := writeTarParentDirs(tw, path.Dir(tarName), createdDirs); err != nil {
			w.CloseWithError(err)
			return fmt.Errorf("failed to write directory tar headers for '%s': %w", tarName, err)
		}

		if err := tw.WriteHeader(&tar.Header{
			Typeflag: tar.TypeReg,
			Name:     tarName,
			Mode:     int64(info.Mode().Perm()),
			Size:     info.Size(),
			ModTime:  info.ModTime(),
		}); err != nil {
			w.CloseWithError(err)
			return fmt.Errorf("failed to write file tar header for '%s': %w", tarName, err)
		}

		f, err := os.Open(entry.LocalPath)
		if err != nil {
			w.CloseWithError(err)
			return fmt.Errorf("failed to open file '%s': %w", entry.LocalPath, err)
		}
		if _, err := io.Copy(tw, f); err != nil {
			f.Close()
			w.CloseWithError(err)
			return fmt.Errorf("failed to write file '%s' into tar: %w", entry.LocalPath, err)
		}
		f.Close()
	}

	return nil
}

func writeTarParentDirs(tw *tar.Writer, dirPath string, created map[string]bool) error {
	parts := strings.Split(dirPath, "/")
	built := ""
	for _, part := range parts {
		if part == "" || part == "." {
			continue
		}
		if built == "" {
			built = part
		} else {
			built = built + "/" + part
		}
		if created[built] {
			continue
		}
		if err := tw.WriteHeader(&tar.Header{
			Typeflag: tar.TypeDir,
			Name:     built,
			Mode:     0755,
			ModTime:  time.Now(),
		}); err != nil {
			return err
		}
		created[built] = true
	}
	return nil
}
