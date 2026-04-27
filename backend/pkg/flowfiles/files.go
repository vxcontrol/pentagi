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
	"sort"
	"strings"
	"time"
)

const (
	UploadsDirName   = "uploads"
	ContainerDirName = "container"

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

func List(dataDir string, flowID uint64) (Files, error) {
	uploadsEntries, err := ListDirEntries(FlowUploadsDir(dataDir, flowID), UploadsDirName)
	if err != nil {
		return Files{}, fmt.Errorf("reading uploads cache: %w", err)
	}

	containerEntries, err := ListDirEntriesRecursive(FlowContainerDir(dataDir, flowID), ContainerDirName)
	if err != nil {
		return Files{}, fmt.Errorf("reading container cache: %w", err)
	}

	files := append(uploadsEntries, containerEntries...)
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

func ResolveCachedPath(dataDir string, flowID uint64, reqPath string) (string, error) {
	if strings.TrimSpace(reqPath) == "" {
		return "", errors.New("path query parameter is required")
	}

	cleaned := filepath.Clean(filepath.FromSlash(strings.ReplaceAll(reqPath, "\\", "/")))
	if filepath.IsAbs(cleaned) {
		return "", fmt.Errorf("path must be relative (no leading /)")
	}

	parts := strings.SplitN(cleaned, string(filepath.Separator), 2)
	if parts[0] != UploadsDirName && parts[0] != ContainerDirName {
		return "", fmt.Errorf("path must start with '%s' or '%s'", UploadsDirName, ContainerDirName)
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
		if files[i].ModifiedAt.Equal(files[j].ModifiedAt) {
			return files[i].Name < files[j].Name
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
	tw := tar.NewWriter(w)
	defer w.Close()
	defer tw.Close()

	if err := tw.WriteHeader(&tar.Header{
		Typeflag: tar.TypeDir,
		Name:     UploadsDirName,
		Mode:     0755,
		ModTime:  time.Now(),
	}); err != nil {
		w.CloseWithError(err)
		return fmt.Errorf("failed to write uploads directory tar header: %w", err)
	}

	var filesCount int
	var totalSize int64
	return filepath.WalkDir(uploadDir, func(entryPath string, d os.DirEntry, err error) error {
		if err != nil {
			w.CloseWithError(err)
			return err
		}
		if entryPath == uploadDir {
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

		rel, err := filepath.Rel(uploadDir, entryPath)
		if err != nil {
			w.CloseWithError(err)
			return fmt.Errorf("failed to get upload relative path: %w", err)
		}
		headerName := path.Join(UploadsDirName, filepath.ToSlash(rel))

		if info.IsDir() {
			if err := tw.WriteHeader(&tar.Header{
				Typeflag: tar.TypeDir,
				Name:     headerName,
				Mode:     int64(info.Mode().Perm()),
				ModTime:  info.ModTime(),
			}); err != nil {
				w.CloseWithError(err)
				return fmt.Errorf("failed to write upload directory tar header: %w", err)
			}
			return nil
		}

		filesCount++
		if filesCount > MaxUploadFiles {
			err := fmt.Errorf("uploads cache exceeds maximum file count of %d", MaxUploadFiles)
			w.CloseWithError(err)
			return err
		}
		totalSize += info.Size()
		if totalSize > MaxUploadTotalSize {
			err := fmt.Errorf("uploads cache exceeds maximum total size of %d bytes", MaxUploadTotalSize)
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
			return fmt.Errorf("failed to write upload file tar header: %w", err)
		}

		f, err := os.Open(entryPath)
		if err != nil {
			w.CloseWithError(err)
			return fmt.Errorf("failed to open upload file: %w", err)
		}
		defer f.Close()

		if _, err := io.Copy(tw, f); err != nil {
			w.CloseWithError(err)
			return fmt.Errorf("failed to write upload file tar content: %w", err)
		}

		return nil
	})
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

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return fmt.Errorf("failed to create zip header: %w", err)
		}
		header.Name = filepath.ToSlash(rel)
		header.Method = zip.Deflate

		zf, err := zw.CreateHeader(header)
		if err != nil {
			return fmt.Errorf("failed to create zip entry: %w", err)
		}

		f, err := os.Open(entryPath)
		if err != nil {
			return fmt.Errorf("failed to open file for zip: %w", err)
		}
		defer f.Close()

		_, err = io.Copy(zf, f)
		return err
	})
}
