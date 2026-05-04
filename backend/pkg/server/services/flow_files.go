package services

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"

	"pentagi/pkg/docker"
	"pentagi/pkg/flowfiles"
	"pentagi/pkg/graph/model"
	"pentagi/pkg/graph/subscriptions"
	"pentagi/pkg/resources"
	"pentagi/pkg/server/logger"
	"pentagi/pkg/server/models"
	"pentagi/pkg/server/response"
	"pentagi/pkg/tools"

	"github.com/docker/docker/api/types/container"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

// Local storage layout under {dataDir}/flow-{id}-data/:
//
//	uploads/    ← user-uploaded files; also pushed to container /work/uploads/
//	resources/  ← user resources copied from user_resources table; also pushed to container /work/resources/
//	container/  ← files synced from the container via pull; never sent back to container

type pendingUpload struct {
	fileName string
	dstPath  string
	tmpPath  string
}

// FlowFileService manages flow-scoped files with two distinct sources:
//   - user uploads   → {dataDir}/flow-{id}-data/uploads/  (also pushed to container /work/uploads/)
//   - user resources → {dataDir}/flow-{id}-data/resources/ (copied from user_resources table)
//   - container sync → {dataDir}/flow-{id}-data/container/ (pulled from container, never sent back)
type FlowFileService struct {
	dataDir      string
	db           *gorm.DB
	dockerClient docker.DockerClient
	ss           subscriptions.SubscriptionsController
}

func NewFlowFileService(
	db *gorm.DB,
	dataDir string,
	dockerClient docker.DockerClient,
	ss subscriptions.SubscriptionsController,
) *FlowFileService {
	return &FlowFileService{
		dataDir:      dataDir,
		db:           db,
		dockerClient: dockerClient,
		ss:           ss,
	}
}

// GetFlowFiles is a function to return flow files list
// @Summary Retrieve flow files list
// @Tags FlowFiles
// @Produce json
// @Security BearerAuth
// @Param flowID path int true "flow id" minimum(0)
// @Success 200 {object} response.successResp{data=models.FlowFiles} "flow files list received successful"
// @Failure 400 {object} response.errorResp "invalid flow files request data"
// @Failure 403 {object} response.errorResp "getting flow files not permitted"
// @Failure 404 {object} response.errorResp "flow not found"
// @Failure 500 {object} response.errorResp "internal error on getting flow files"
// @Router /flows/{flowID}/files/ [get]
func (s *FlowFileService) GetFlowFiles(c *gin.Context) {
	flowID, err := parseFlowIDParam(c)
	if err != nil {
		logger.FromContext(c).WithError(err).Error("error parsing flow id")
		response.Error(c, response.ErrFlowFilesInvalidRequest, err)
		return
	}

	if _, err := s.getFlow(c, flowID, false); err != nil {
		s.handleFlowLookupError(c, flowID, err)
		return
	}

	files, err := s.listFlowFiles(flowID)
	if err != nil {
		logger.FromContext(c).WithError(err).WithField("flow_id", flowID).Error("error listing flow files")
		response.Error(c, response.ErrInternal, err)
		return
	}

	response.Success(c, http.StatusOK, models.FlowFiles{
		Files: files,
		Total: uint64(len(files)),
	})
}

// UploadFlowFiles is a function to upload files to the flow workspace
// @Summary Upload files to flow workspace
// @Tags FlowFiles
// @Accept mpfd
// @Produce json
// @Security BearerAuth
// @Param flowID path int true "flow id" minimum(0)
// @Param files formData file true "files to upload (multipart, field name: files or file)"
// @Success 200 {object} response.successResp{data=models.FlowFiles} "flow files uploaded successful"
// @Failure 400 {object} response.errorResp "invalid flow files request data"
// @Failure 403 {object} response.errorResp "uploading flow files not permitted"
// @Failure 404 {object} response.errorResp "flow not found"
// @Failure 409 {object} response.errorResp "flow file already exists"
// @Failure 500 {object} response.errorResp "internal error on uploading flow files"
// @Router /flows/{flowID}/files/ [post]
func (s *FlowFileService) UploadFlowFiles(c *gin.Context) {
	flowID, err := parseFlowIDParam(c)
	if err != nil {
		logger.FromContext(c).WithError(err).Error("error parsing flow id")
		response.Error(c, response.ErrFlowFilesInvalidRequest, err)
		return
	}

	flow, err := s.getFlow(c, flowID, true)
	if err != nil {
		s.handleFlowLookupError(c, flowID, err)
		return
	}

	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, flowfiles.MaxUploadRequestSize)
	multipartForm, err := c.MultipartForm()
	if err != nil {
		logger.FromContext(c).WithError(err).WithField("flow_id", flowID).Error("error reading multipart form")
		response.Error(c, response.ErrFlowFilesInvalidRequest, err)
		return
	}

	fileHeaders := multipartForm.File["files"]
	if len(fileHeaders) == 0 {
		fileHeader, formErr := c.FormFile("file")
		if formErr == nil && fileHeader != nil {
			fileHeaders = append(fileHeaders, fileHeader)
		}
	}
	if len(fileHeaders) == 0 {
		err = errors.New("at least one uploaded file is required")
		logger.FromContext(c).WithError(err).WithField("flow_id", flowID).Error("missing uploaded files")
		response.Error(c, response.ErrFlowFilesInvalidRequest, err)
		return
	}
	if len(fileHeaders) > flowfiles.MaxUploadFiles {
		err = fmt.Errorf("too many uploaded files: %d exceeds the maximum allowed count of %d",
			len(fileHeaders), flowfiles.MaxUploadFiles)
		logger.FromContext(c).WithError(err).WithField("flow_id", flowID).Error("too many uploaded files")
		response.Error(c, response.ErrFlowFilesInvalidRequest, err)
		return
	}

	uploadDir := s.flowUploadsDir(flowID)
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		logger.FromContext(c).WithError(err).WithField("flow_id", flowID).Error("error creating upload directory")
		response.Error(c, response.ErrInternal, err)
		return
	}

	// Pre-validate all files before saving any to avoid partial writes on error.
	pending := make([]pendingUpload, 0, len(fileHeaders))
	var totalSize int64

	for _, fh := range fileHeaders {
		// SanitizeFileName strips path separators — only the basename is kept.
		// This prevents path traversal via crafted multipart filenames.
		fileName, err := flowfiles.SanitizeFileName(fh.Filename)
		if err != nil {
			logger.FromContext(c).WithError(err).WithField("flow_id", flowID).Error("invalid uploaded file name")
			response.Error(c, response.ErrFlowFilesInvalidData, err)
			return
		}

		if fh.Size > flowfiles.MaxUploadFileSize {
			err = fmt.Errorf("file '%s' size %d bytes exceeds the maximum allowed upload size of %d bytes",
				fileName, fh.Size, flowfiles.MaxUploadFileSize)
			logger.FromContext(c).WithError(err).WithFields(map[string]any{
				"flow_id":   flowID,
				"file_name": fileName,
			}).Error("uploaded file too large")
			response.Error(c, response.ErrFlowFilesInvalidRequest, err)
			return
		}
		if fh.Size < 0 {
			err = fmt.Errorf("file '%s' has invalid size %d", fileName, fh.Size)
			logger.FromContext(c).WithError(err).WithFields(map[string]any{
				"flow_id":   flowID,
				"file_name": fileName,
			}).Error("uploaded file has invalid size")
			response.Error(c, response.ErrFlowFilesInvalidRequest, err)
			return
		}
		totalSize += fh.Size
		if totalSize > flowfiles.MaxUploadTotalSize {
			err = fmt.Errorf("uploaded files total size %d bytes exceeds the maximum allowed size of %d bytes",
				totalSize, flowfiles.MaxUploadTotalSize)
			logger.FromContext(c).WithError(err).WithField("flow_id", flowID).Error("uploaded files total size too large")
			response.Error(c, response.ErrFlowFilesInvalidRequest, err)
			return
		}

		dstPath := filepath.Join(uploadDir, fileName)
		exists, err := flowfiles.LocalEntryExists(dstPath)
		if err != nil {
			logger.FromContext(c).WithError(err).WithFields(map[string]any{
				"flow_id":   flowID,
				"file_name": fileName,
			}).Error("error checking existing uploaded file")
			response.Error(c, response.ErrInternal, err)
			return
		}
		if exists {
			err = fmt.Errorf("flow file '%s' already exists", fileName)
			logger.FromContext(c).WithError(err).WithFields(map[string]any{
				"flow_id":   flowID,
				"file_name": fileName,
			}).Error("uploaded file already exists")
			response.Error(c, response.ErrFlowFilesAlreadyExists, err)
			return
		}

		pending = append(pending, pendingUpload{fileName: fileName, dstPath: dstPath})
	}

	// All files passed validation — write temporary files first to avoid partial commits on copy errors.
	for i := range pending {
		tmpPath, err := flowfiles.SaveUploadedFileToTemp(fileHeaders[i], uploadDir)
		if err != nil {
			logger.FromContext(c).WithError(err).WithFields(map[string]any{
				"flow_id":   flowID,
				"file_name": pending[i].fileName,
			}).Error("error saving uploaded file to temporary cache")
			cleanupPendingUploads(pending)
			response.Error(c, response.ErrInternal, err)
			return
		}
		pending[i].tmpPath = tmpPath
	}

	// Temporary files are complete — move them into place, then push to container.
	savedFiles := make([]models.FlowFile, 0, len(fileHeaders))
	pushEntries := make([]flowfiles.TarEntry, 0, len(pending))
	for i := range pending {
		p := &pending[i]
		if err := os.Rename(p.tmpPath, p.dstPath); err != nil {
			logger.FromContext(c).WithError(err).WithFields(map[string]any{
				"flow_id":   flowID,
				"file_name": p.fileName,
			}).Error("error committing uploaded file")
			cleanupPendingUploads(pending)
			response.Error(c, response.ErrInternal, err)
			return
		}
		p.tmpPath = ""

		info, err := flowfiles.RegularFileInfo(p.dstPath)
		if err != nil {
			logger.FromContext(c).WithError(err).WithFields(map[string]any{
				"flow_id":   flowID,
				"file_name": p.fileName,
			}).Error("error stating uploaded file")
			response.Error(c, response.ErrInternal, err)
			return
		}

		pushEntries = append(pushEntries, flowfiles.TarEntry{
			LocalPath: p.dstPath,
			TarPath:   path.Join(flowfiles.UploadsDirName, p.fileName),
		})
		savedFiles = append(savedFiles, convertFlowFile(flowfiles.NewFile(info, flowfiles.UploadsDirName)))
	}

	// Best-effort: one tar stream + one CopyToContainer for all new uploads.
	if pushErr := s.copyLocalFilesToPrimaryWork(c.Request.Context(), flowID, pushEntries); pushErr != nil {
		logger.FromContext(c).WithError(pushErr).WithFields(map[string]any{
			"flow_id": flowID,
			"count":   len(pushEntries),
		}).Warn("uploaded files saved locally but could not be pushed to container")
	}

	sortFlowFiles(savedFiles)
	s.publishFlowFilesAdded(c.Request.Context(), flow, savedFiles)
	response.Success(c, http.StatusOK, models.FlowFiles{
		Files: savedFiles,
		Total: uint64(len(savedFiles)),
	})
}

// DeleteFlowFile is a function to delete one or more cached flow files or directories by path.
// Accepts a single "path" query parameter and/or repeated "paths[]" parameters; both may be
// combined and duplicates (after path normalization) are silently ignored.
// @Summary Delete flow files or directories by cached paths
// @Tags FlowFiles
// @Produce json
// @Security BearerAuth
// @Param flowID path int true "flow id" minimum(0)
// @Param path query string false "relative path in cache: uploads/, resources/, or container/ (may be combined with paths[])"
// @Param paths[] query []string false "relative paths in cache (repeatable): uploads/, resources/, or container/"
// @Success 200 {object} response.successResp{data=models.FlowFiles} "flow files deleted successfully"
// @Failure 400 {object} response.errorResp "invalid flow file request data"
// @Failure 403 {object} response.errorResp "deleting flow file not permitted"
// @Failure 404 {object} response.errorResp "flow file not found"
// @Failure 500 {object} response.errorResp "internal error on deleting flow file"
// @Router /flows/{flowID}/files/ [delete]
func (s *FlowFileService) DeleteFlowFile(c *gin.Context) {
	flowID, err := parseFlowIDParam(c)
	if err != nil {
		logger.FromContext(c).WithError(err).Error("error parsing flow id")
		response.Error(c, response.ErrFlowFilesInvalidRequest, err)
		return
	}

	flow, err := s.getFlow(c, flowID, true)
	if err != nil {
		s.handleFlowLookupError(c, flowID, err)
		return
	}

	// Collect paths from both "path" and "paths[]" query params, then deduplicate strings.
	rawPaths := c.QueryArray("paths[]")
	if singlePath := c.Query("path"); singlePath != "" {
		rawPaths = append(rawPaths, singlePath)
	}
	reqPaths := flowfiles.DeduplicatePaths(rawPaths)
	if len(reqPaths) == 0 {
		err = errors.New("at least one path is required (use 'path' or 'paths[]' query parameters)")
		logger.FromContext(c).WithError(err).WithField("flow_id", flowID).Error("missing delete paths")
		response.Error(c, response.ErrFlowFilesInvalidRequest, err)
		return
	}

	// Resolve and validate every path before performing any deletion; deduplicate
	// by resolved local path to handle path-equivalent inputs (e.g. "uploads/./x" == "uploads/x").
	type resolvedEntry struct {
		reqPath   string
		localPath string
		info      os.FileInfo
	}
	seenLocal := make(map[string]struct{}, len(reqPaths))
	entries := make([]resolvedEntry, 0, len(reqPaths))
	for _, reqPath := range reqPaths {
		localPath, err := s.resolveCachedPath(flowID, reqPath)
		if err != nil {
			logger.FromContext(c).WithError(err).WithField("flow_id", flowID).Error("invalid delete path")
			response.Error(c, response.ErrFlowFilesInvalidRequest, err)
			return
		}
		if _, dup := seenLocal[localPath]; dup {
			continue
		}
		seenLocal[localPath] = struct{}{}

		info, err := os.Lstat(localPath)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				response.Error(c, response.ErrFlowFilesNotFound, fmt.Errorf("'%s' not found in local cache", reqPath))
			} else {
				logger.FromContext(c).WithError(err).WithField("flow_id", flowID).Error("error reading cached flow file")
				response.Error(c, response.ErrInternal, err)
			}
			return
		}
		entries = append(entries, resolvedEntry{
			reqPath:   reqPath,
			localPath: localPath,
			info:      info,
		})
	}

	// Expand each entry to collect the full set of files and directories that will
	// actually disappear from the cache. For directories we walk recursively so
	// that every removed sub-entry is reported to the caller and published via
	// subscription — giving the client a precise picture of what was deleted.
	// The actual deletion below still operates on the minimal deduplicated set.
	expandedFiles := make([]flowfiles.File, 0, len(entries))
	for _, e := range entries {
		relPath := filepath.ToSlash(e.reqPath)
		expandedFiles = append(expandedFiles, flowfiles.NewFileWithPath(e.info, relPath))
		if e.info.IsDir() {
			nested, nestErr := flowfiles.ListDirEntriesRecursive(e.localPath, relPath)
			if nestErr != nil {
				logger.FromContext(c).WithError(nestErr).WithFields(map[string]any{
					"flow_id": flowID,
					"path":    relPath,
				}).Warn("could not expand directory contents before deletion; reporting top-level entry only")
			}
			expandedFiles = append(expandedFiles, nested...)
		}
	}
	flowfiles.Sort(expandedFiles)
	deletedFiles := convertFlowFiles(expandedFiles)

	// One Docker exec removes all relevant container paths in a single API call.
	allReqPaths := make([]string, 0, len(entries))
	for _, e := range entries {
		allReqPaths = append(allReqPaths, e.reqPath)
	}
	if err := s.deleteUploadsFromContainer(c.Request.Context(), flowID, allReqPaths); err != nil {
		logger.FromContext(c).WithError(err).WithField("flow_id", flowID).Error("error deleting files from primary container")
		response.Error(c, response.ErrInternal, err)
		return
	}

	// Remove local cache entries after the container sync succeeded.
	for _, e := range entries {
		if err := os.RemoveAll(e.localPath); err != nil {
			logger.FromContext(c).WithError(err).WithField("flow_id", flowID).Error("error deleting cached flow file")
			response.Error(c, response.ErrInternal, err)
			return
		}
	}

	s.publishFlowFilesDeleted(c.Request.Context(), flow, deletedFiles)
	response.Success(c, http.StatusOK, models.FlowFiles{
		Files: deletedFiles,
		Total: uint64(len(deletedFiles)),
	})
}

// DownloadFlowFile is a function to download flow file(s) or directory by cached path(s)
// @Summary Download flow file(s) or directory by cached path(s)
// @Description Single regular file: served as a direct attachment.
// @Description Single directory or multiple paths (any mix): packaged as a ZIP archive.
// @Tags FlowFiles
// @Produce application/octet-stream,application/zip,json
// @Security BearerAuth
// @Param flowID path int true "flow id" minimum(0)
// @Param path query string false "relative path in cache: uploads/, resources/, or container/ (may be combined with paths[])"
// @Param paths[] query []string false "relative paths in cache (repeatable)"
// @Success 200 {file} binary "file content, or ZIP archive for directories / multiple paths"
// @Failure 400 {object} response.errorResp "invalid flow file request data"
// @Failure 403 {object} response.errorResp "downloading flow file not permitted"
// @Failure 404 {object} response.errorResp "flow file not found"
// @Failure 500 {object} response.errorResp "internal error on downloading flow file"
// @Router /flows/{flowID}/files/download [get]
func (s *FlowFileService) DownloadFlowFile(c *gin.Context) {
	flowID, err := parseFlowIDParam(c)
	if err != nil {
		logger.FromContext(c).WithError(err).Error("error parsing flow id")
		response.Error(c, response.ErrFlowFilesInvalidRequest, err)
		return
	}

	if _, err := s.getFlow(c, flowID, false); err != nil {
		s.handleFlowLookupError(c, flowID, err)
		return
	}

	// Collect paths from both "path" and "paths[]" query params, then deduplicate.
	rawPaths := c.QueryArray("paths[]")
	if singlePath := c.Query("path"); singlePath != "" {
		rawPaths = append(rawPaths, singlePath)
	}
	reqPaths := flowfiles.DeduplicatePaths(rawPaths)
	if len(reqPaths) == 0 {
		err = errors.New("at least one path is required (use 'path' or 'paths[]' query parameters)")
		logger.FromContext(c).WithError(err).WithField("flow_id", flowID).Error("missing download paths")
		response.Error(c, response.ErrFlowFilesInvalidRequest, err)
		return
	}

	// Resolve and validate every path; deduplicate by resolved local path.
	type resolvedEntry struct {
		reqPath   string
		localPath string
		info      os.FileInfo
	}
	seenLocal := make(map[string]struct{}, len(reqPaths))
	entries := make([]resolvedEntry, 0, len(reqPaths))
	for _, reqPath := range reqPaths {
		localPath, err := s.resolveCachedPath(flowID, reqPath)
		if err != nil {
			logger.FromContext(c).WithError(err).WithField("flow_id", flowID).Error("invalid download path")
			response.Error(c, response.ErrFlowFilesInvalidRequest, err)
			return
		}
		if _, dup := seenLocal[localPath]; dup {
			continue
		}
		seenLocal[localPath] = struct{}{}

		info, err := os.Lstat(localPath)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				response.Error(c, response.ErrFlowFilesNotFound, fmt.Errorf("'%s' not found in local cache", reqPath))
			} else {
				logger.FromContext(c).WithError(err).WithField("flow_id", flowID).Error("error reading cached flow file")
				response.Error(c, response.ErrInternal, err)
			}
			return
		}
		// Never serve symlinks — could point outside the flow data directory.
		if info.Mode()&os.ModeSymlink != 0 {
			response.Error(c, response.ErrFlowFilesNotFound, fmt.Errorf("'%s' not found in local cache", reqPath))
			return
		}
		if !info.IsDir() && !info.Mode().IsRegular() {
			response.Error(c, response.ErrFlowFilesNotFound, fmt.Errorf("'%s' not found in local cache", reqPath))
			return
		}
		entries = append(entries, resolvedEntry{reqPath: reqPath, localPath: localPath, info: info})
	}

	if len(entries) == 0 {
		response.Error(c, response.ErrFlowFilesNotFound, errors.New("no accessible paths found"))
		return
	}

	// Single regular file → serve as a direct attachment.
	// We open the file explicitly and use DataFromReader with a known Content-Length
	// so that Gin's response writer emits a proper Content-Length header instead of
	// relying on http.ServeFile, which conflicts with Gin's middleware-set headers
	// and can produce content-length: 0 in certain request contexts.
	if len(entries) == 1 && entries[0].info.Mode().IsRegular() {
		e := entries[0]
		f, err := os.Open(e.localPath)
		if err != nil {
			logger.FromContext(c).WithError(err).WithField("flow_id", flowID).Error("error opening flow file for download")
			response.Error(c, response.ErrInternal, err)
			return
		}
		defer f.Close()

		contentType := "application/octet-stream"
		c.DataFromReader(http.StatusOK, e.info.Size(), contentType, f,
			map[string]string{
				"Content-Disposition": mime.FormatMediaType("attachment", map[string]string{
					"filename": filepath.Base(e.localPath),
				}),
			})
		return
	}

	// Single directory → ZIP with paths relative to that directory (backward-compat).
	// The archive is buffered so an explicit Content-Length can be set; this allows
	// clients (including Swagger UI) to recognise and download the file correctly.
	if len(entries) == 1 && entries[0].info.IsDir() {
		name := filepath.Base(entries[0].localPath)
		var buf bytes.Buffer
		if err := flowfiles.ZipDirectory(&buf, entries[0].localPath); err != nil {
			logger.FromContext(c).WithError(err).WithField("flow_id", flowID).Error("error creating ZIP archive")
			response.Error(c, response.ErrInternal, err)
			return
		}
		c.DataFromReader(http.StatusOK, int64(buf.Len()), "application/zip", &buf,
			map[string]string{
				"Content-Disposition": mime.FormatMediaType("attachment", map[string]string{
					"filename": name + ".zip",
				}),
			})
		return
	}

	// Multiple paths (any mix of files and directories) → ZIP with cache-relative paths.
	// Buffered for the same reason as above.
	relPaths := make([]string, 0, len(entries))
	for _, e := range entries {
		relPaths = append(relPaths, filepath.ToSlash(e.reqPath))
	}
	var buf bytes.Buffer
	if err := flowfiles.ZipRelativePaths(&buf, s.flowDataDir(flowID), relPaths); err != nil {
		logger.FromContext(c).WithError(err).WithField("flow_id", flowID).Error("error creating ZIP archive")
		response.Error(c, response.ErrInternal, err)
		return
	}
	c.DataFromReader(http.StatusOK, int64(buf.Len()), "application/zip", &buf,
		map[string]string{
			"Content-Disposition": mime.FormatMediaType("attachment", map[string]string{
				"filename": "download.zip",
			}),
		})
}

// PullFlowFiles is a function to sync one or more paths from the container into the local cache
// @Summary Pull files or directories from the running container into the local cache
// @Tags FlowFiles
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param flowID path int true "flow id" minimum(0)
// @Param request body models.PullFlowFilesRequest true "pull request: path or paths are required"
// @Success 200 {object} response.successResp{data=models.FlowFiles} "container paths synced to local cache, sorted by cache path"
// @Failure 400 {object} response.errorResp "invalid pull request or container not running"
// @Failure 403 {object} response.errorResp "pulling flow files not permitted"
// @Failure 404 {object} response.errorResp "flow not found"
// @Failure 409 {object} response.errorResp "cache entry already exists; set force=true to overwrite"
// @Failure 500 {object} response.errorResp "internal error on pulling flow files"
// @Router /flows/{flowID}/files/pull [post]
func (s *FlowFileService) PullFlowFiles(c *gin.Context) {
	flowID, err := parseFlowIDParam(c)
	if err != nil {
		logger.FromContext(c).WithError(err).Error("error parsing flow id")
		response.Error(c, response.ErrFlowFilesInvalidRequest, err)
		return
	}

	flow, err := s.getFlow(c, flowID, true)
	if err != nil {
		s.handleFlowLookupError(c, flowID, err)
		return
	}

	// Container interaction additionally requires containers.view (or containers.admin).
	privs := c.GetStringSlice("prm")
	if !slices.Contains(privs, "containers.admin") && !slices.Contains(privs, "containers.view") {
		response.Error(c, response.ErrNotPermitted, fmt.Errorf("containers.view privilege is required to pull from container"))
		return
	}

	var req models.PullFlowFilesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.FromContext(c).WithError(err).WithField("flow_id", flowID).Error("error parsing pull request")
		response.Error(c, response.ErrFlowFilesInvalidRequest, err)
		return
	}

	// Collect container paths from both req.Path and req.Paths, then deduplicate
	// with coverage semantics using flowfiles.DeduplicatePaths.
	// Container paths (/etc/nginx.conf) are normalised to their relative form
	// (etc/nginx.conf) so that DeduplicatePaths can apply ancestor coverage:
	// /etc/ covers /etc/nginx.conf → only /etc/ is pulled.
	rawPaths := req.Paths
	if strings.TrimSpace(req.Path) != "" {
		rawPaths = append(rawPaths, req.Path)
	}
	relPaths := make([]string, 0, len(rawPaths))
	for _, p := range rawPaths {
		trimmed := strings.TrimSpace(p)
		if trimmed == "" {
			continue
		}
		// Mirror SanitizeContainerCachePath normalisation: prepend "/", Clean, strip "/".
		rel := strings.TrimPrefix(path.Clean("/"+strings.ReplaceAll(trimmed, "\\", "/")), "/")
		if rel != "" && rel != "." {
			relPaths = append(relPaths, rel)
		}
	}
	dedupedRel := flowfiles.DeduplicatePaths(relPaths)
	if len(dedupedRel) == 0 {
		err = errors.New("at least one valid path is required (use 'path' or 'paths' fields)")
		logger.FromContext(c).WithError(err).WithField("flow_id", flowID).Error("missing container paths in pull request")
		response.Error(c, response.ErrFlowFilesInvalidRequest, err)
		return
	}
	// Reconstruct absolute paths for the Docker API.
	containerPaths := make([]string, len(dedupedRel))
	for i, rel := range dedupedRel {
		containerPaths[i] = "/" + rel
	}

	containerDir := s.flowContainerDir(flowID)

	// ── Phase 1: validate every path and check for cache conflicts ────────────
	// All validation is done upfront so that no docker operations are started
	// if any path is invalid or already cached without force=true.
	type pullEntry struct {
		containerPath string
		cacheRelPath  string
		localTarget   string
		targetExists  bool
	}
	seenCachePaths := make(map[string]struct{}, len(containerPaths))
	entries := make([]pullEntry, 0, len(containerPaths))
	for _, containerPath := range containerPaths {
		cacheRelPath, err := flowfiles.SanitizeContainerCachePath(containerPath)
		if err != nil {
			logger.FromContext(c).WithError(err).WithFields(map[string]any{
				"flow_id":        flowID,
				"container_path": containerPath,
			}).Error("invalid container path")
			response.Error(c, response.ErrFlowFilesInvalidRequest, fmt.Errorf("invalid container path %q: %w", containerPath, err))
			return
		}

		// Deduplicate by resolved cache path (handles /etc/foo and etc/foo being equivalent).
		if _, dup := seenCachePaths[cacheRelPath]; dup {
			continue
		}
		seenCachePaths[cacheRelPath] = struct{}{}

		localTarget := filepath.Join(containerDir, filepath.FromSlash(cacheRelPath))
		targetExists := false
		if _, statErr := os.Lstat(localTarget); statErr == nil {
			targetExists = true
			if !req.Force {
				err = fmt.Errorf("'%s' already exists in the container cache; set force=true to overwrite", cacheRelPath)
				logger.FromContext(c).WithError(err).WithFields(map[string]any{
					"flow_id":    flowID,
					"cache_path": cacheRelPath,
				}).Error("pull target already exists")
				response.Error(c, response.ErrFlowFilesAlreadyExists, err)
				return
			}
		} else if !errors.Is(statErr, os.ErrNotExist) {
			logger.FromContext(c).WithError(statErr).WithField("flow_id", flowID).Error("error checking container cache entry")
			response.Error(c, response.ErrInternal, statErr)
			return
		}
		entries = append(entries, pullEntry{
			containerPath: containerPath,
			cacheRelPath:  cacheRelPath,
			localTarget:   localTarget,
			targetExists:  targetExists,
		})
	}

	if err := os.MkdirAll(containerDir, 0755); err != nil {
		logger.FromContext(c).WithError(err).WithField("flow_id", flowID).Error("error creating container cache directory")
		response.Error(c, response.ErrInternal, err)
		return
	}

	if s.dockerClient == nil {
		err = errors.New("docker client not configured on this server")
		logger.FromContext(c).WithError(err).WithField("flow_id", flowID).Error("docker client unavailable for pull")
		response.Error(c, response.ErrInternal, err)
		return
	}

	containerName := primaryContainerName(flowID)
	running, err := s.dockerClient.IsContainerRunning(c.Request.Context(), containerName)
	if err != nil {
		logger.FromContext(c).WithError(err).WithFields(map[string]any{
			"flow_id":        flowID,
			"container_name": containerName,
		}).Error("error checking container status for pull")
		response.Error(c, response.ErrInternal, err)
		return
	}
	if !running {
		err = fmt.Errorf("container '%s' is not running; start the flow before pulling files", containerName)
		logger.FromContext(c).WithError(err).WithField("flow_id", flowID).Error("container not running for pull")
		response.Error(c, response.ErrFlowFilesContainerNotRunning, err)
		return
	}

	// ── Phase 2: pull each validated entry from the container ─────────────────
	syncedFiles := make([]models.FlowFile, 0, len(entries))
	addedFiles := make([]models.FlowFile, 0, len(entries))
	updatedFiles := make([]models.FlowFile, 0)

	for _, entry := range entries {
		reader, _, err := s.dockerClient.CopyFromContainer(c.Request.Context(), containerName, entry.containerPath)
		if err != nil {
			logger.FromContext(c).WithError(err).WithFields(map[string]any{
				"flow_id":        flowID,
				"container_path": entry.containerPath,
			}).Error("error copying from container")
			response.Error(c, response.ErrInternal, err)
			return
		}

		stagingDir, err := os.MkdirTemp(containerDir, ".pull-*")
		if err != nil {
			reader.Close()
			logger.FromContext(c).WithError(err).WithField("flow_id", flowID).Error("error creating pull staging directory")
			response.Error(c, response.ErrInternal, err)
			return
		}

		extractErr := flowfiles.ExtractTar(reader, stagingDir)
		reader.Close()
		if extractErr != nil {
			os.RemoveAll(stagingDir)
			logger.FromContext(c).WithError(extractErr).WithField("flow_id", flowID).Error("error extracting container TAR archive")
			response.Error(c, response.ErrInternal, extractErr)
			return
		}

		stagedTarget := flowfiles.ResolvePulledStagedTarget(stagingDir, entry.cacheRelPath)
		if stagedTarget == "" {
			os.RemoveAll(stagingDir)
			err = fmt.Errorf("pulled archive did not contain expected entry '%s'", entry.cacheRelPath)
			logger.FromContext(c).WithError(err).WithFields(map[string]any{
				"flow_id":    flowID,
				"cache_path": entry.cacheRelPath,
			}).Error("pulled container archive did not contain expected entry")
			response.Error(c, response.ErrInternal, err)
			return
		}
		if _, statErr := os.Lstat(stagedTarget); statErr != nil {
			os.RemoveAll(stagingDir)
			logger.FromContext(c).WithError(statErr).WithFields(map[string]any{
				"flow_id":    flowID,
				"cache_path": entry.cacheRelPath,
			}).Error("pulled container archive did not contain expected entry")
			response.Error(c, response.ErrInternal, statErr)
			return
		}

		if entry.targetExists {
			if err := os.RemoveAll(entry.localTarget); err != nil {
				os.RemoveAll(stagingDir)
				logger.FromContext(c).WithError(err).WithFields(map[string]any{
					"flow_id":    flowID,
					"cache_path": entry.cacheRelPath,
				}).Error("error removing existing cache entry before forced pull")
				response.Error(c, response.ErrInternal, err)
				return
			}
		}
		if err := os.MkdirAll(filepath.Dir(entry.localTarget), 0755); err != nil {
			os.RemoveAll(stagingDir)
			logger.FromContext(c).WithError(err).WithFields(map[string]any{
				"flow_id":    flowID,
				"cache_path": entry.cacheRelPath,
			}).Error("error creating container cache parent directory")
			response.Error(c, response.ErrInternal, err)
			return
		}
		if err := os.Rename(stagedTarget, entry.localTarget); err != nil {
			os.RemoveAll(stagingDir)
			logger.FromContext(c).WithError(err).WithFields(map[string]any{
				"flow_id":    flowID,
				"cache_path": entry.cacheRelPath,
			}).Error("error committing pulled container entry")
			response.Error(c, response.ErrInternal, err)
			return
		}
		os.RemoveAll(stagingDir)

		info, err := os.Lstat(entry.localTarget)
		if err != nil {
			logger.FromContext(c).WithError(err).WithFields(map[string]any{
				"flow_id":    flowID,
				"cache_path": entry.cacheRelPath,
			}).Error("error stating synced container entry")
			response.Error(c, response.ErrInternal, err)
			return
		}

		// Build the full list of synced entries: root entry + recursive contents for
		// directories. Mirrors the expansion done in DeleteFlowFile so the caller gets
		// a complete picture of what was actually written to the cache.
		relPath := path.Join(flowfiles.ContainerDirName, entry.cacheRelPath)
		rootFile := convertFlowFile(flowfiles.NewFileWithPath(info, relPath))
		expanded := []models.FlowFile{rootFile}
		if info.IsDir() {
			nested, nestErr := flowfiles.ListDirEntriesRecursive(entry.localTarget, relPath)
			if nestErr != nil {
				logger.FromContext(c).WithError(nestErr).WithFields(map[string]any{
					"flow_id": flowID,
					"path":    relPath,
				}).Warn("could not list pulled directory contents for response")
			} else {
				for _, f := range nested {
					expanded = append(expanded, convertFlowFile(f))
				}
			}
		}
		syncedFiles = append(syncedFiles, expanded...)
		if entry.targetExists {
			updatedFiles = append(updatedFiles, expanded...)
		} else {
			addedFiles = append(addedFiles, expanded...)
		}
	}

	// Sort results by cache path for deterministic output.
	sort.Slice(syncedFiles, func(i, j int) bool {
		if syncedFiles[i].Path < syncedFiles[j].Path {
			return true
		} else if syncedFiles[i].Path > syncedFiles[j].Path {
			return false
		} else if syncedFiles[i].Name < syncedFiles[j].Name {
			return true
		} else if syncedFiles[i].Name > syncedFiles[j].Name {
			return false
		}
		return syncedFiles[i].ModifiedAt.After(syncedFiles[j].ModifiedAt)
	})

	ctx := c.Request.Context()
	if len(addedFiles) > 0 {
		s.publishFlowFilesAdded(ctx, flow, addedFiles)
	}
	for _, f := range updatedFiles {
		s.publishFlowFileUpdated(ctx, flow, f)
	}

	response.Success(c, http.StatusOK, models.FlowFiles{
		Files: syncedFiles,
		Total: uint64(len(syncedFiles)),
	})
}

// GetFlowContainerFiles is a function to return non-recursive container directory files list
// @Summary Retrieve flow container directory files list
// @Tags FlowFiles
// @Produce json
// @Security BearerAuth
// @Param flowID path int true "flow id" minimum(0)
// @Param path query string false "absolute path inside the running container; defaults to /work (may be combined with paths[])"
// @Param paths[] query []string false "absolute paths inside the running container (repeatable)"
// @Success 200 {object} response.successResp{data=models.ContainerFiles} "container files list received successful"
// @Failure 400 {object} response.errorResp "invalid flow files request data or container not running"
// @Failure 403 {object} response.errorResp "getting flow container files not permitted"
// @Failure 404 {object} response.errorResp "flow not found"
// @Failure 500 {object} response.errorResp "internal error on getting flow container files"
// @Router /flows/{flowID}/files/container [get]
func (s *FlowFileService) GetFlowContainerFiles(c *gin.Context) {
	flowID, err := parseFlowIDParam(c)
	if err != nil {
		logger.FromContext(c).WithError(err).Error("error parsing flow id")
		response.Error(c, response.ErrFlowFilesInvalidRequest, err)
		return
	}

	if _, err := s.getFlow(c, flowID, false); err != nil {
		s.handleFlowLookupError(c, flowID, err)
		return
	}

	// Container interaction additionally requires containers.view (or containers.admin).
	privs := c.GetStringSlice("prm")
	if !slices.Contains(privs, "containers.admin") && !slices.Contains(privs, "containers.view") {
		response.Error(c, response.ErrNotPermitted, fmt.Errorf("containers.view privilege is required to list container files"))
		return
	}

	if s.dockerClient == nil {
		err = errors.New("docker client not configured on this server")
		logger.FromContext(c).WithError(err).WithField("flow_id", flowID).Error("docker client unavailable for container files list")
		response.Error(c, response.ErrInternal, err)
		return
	}

	// Collect paths from both "path" and "paths[]" query params, deduplicate by
	// exact string. If explicit paths are provided but all are whitespace-only,
	// return an error. When nothing is provided at all, default to /work.
	rawPaths := c.QueryArray("paths[]")
	if singlePath := strings.TrimSpace(c.Query("path")); singlePath != "" {
		rawPaths = append(rawPaths, singlePath)
	}
	containerPaths := deduplicateContainerPaths(rawPaths)
	if len(rawPaths) > 0 && len(containerPaths) == 0 {
		err = errors.New("at least one valid path is required (use 'path' or 'paths[]' query parameters)")
		logger.FromContext(c).WithError(err).WithField("flow_id", flowID).Error("missing container paths")
		response.Error(c, response.ErrFlowFilesInvalidRequest, err)
		return
	}
	if len(containerPaths) == 0 {
		containerPaths = []string{docker.WorkFolderPathInContainer}
	}

	containerName := primaryContainerName(flowID)
	running, err := s.dockerClient.IsContainerRunning(c.Request.Context(), containerName)
	if err != nil {
		logger.FromContext(c).WithError(err).WithFields(map[string]any{
			"flow_id":        flowID,
			"container_name": containerName,
		}).Error("error checking container status for files list")
		response.Error(c, response.ErrInternal, err)
		return
	}
	if !running {
		err = fmt.Errorf("container '%s' is not running; start the flow before listing container files", containerName)
		logger.FromContext(c).WithError(err).WithField("flow_id", flowID).Error("container not running for files list")
		response.Error(c, response.ErrFlowFilesContainerNotRunning, err)
		return
	}

	// Query each path and collect results. Output is deduplicated by the
	// resolved file.Path so that overlapping queries never return the same
	// entry twice.
	seenPaths := make(map[string]struct{})
	allFiles := make([]models.ContainerFile, 0)
	for _, containerPath := range containerPaths {
		pathStat, err := s.dockerClient.ContainerStatPath(c.Request.Context(), containerName, containerPath)
		if err != nil {
			logger.FromContext(c).WithError(err).WithFields(map[string]any{
				"flow_id":        flowID,
				"container_path": containerPath,
			}).Error("error stating container path")
			response.Error(c, response.ErrInternal, err)
			return
		}

		if !pathStat.Mode.IsDir() {
			file := convertContainerFile(path.Dir(containerPath), pathStat)
			if _, seen := seenPaths[file.Path]; !seen {
				seenPaths[file.Path] = struct{}{}
				allFiles = append(allFiles, file)
			}
			continue
		}

		stats, err := s.dockerClient.ListContainerDir(c.Request.Context(), containerName, containerPath)
		if err != nil {
			logger.FromContext(c).WithError(err).WithFields(map[string]any{
				"flow_id":        flowID,
				"container_path": containerPath,
			}).Error("error listing container directory")
			response.Error(c, response.ErrInternal, err)
			return
		}
		for _, stat := range stats {
			file := convertContainerFile(containerPath, stat)
			if _, seen := seenPaths[file.Path]; !seen {
				seenPaths[file.Path] = struct{}{}
				allFiles = append(allFiles, file)
			}
		}
	}

	// Sort by name for deterministic output across all paths.
	sort.Slice(allFiles, func(i, j int) bool {
		if allFiles[i].Path < allFiles[j].Path {
			return true
		} else if allFiles[i].Path > allFiles[j].Path {
			return false
		} else if allFiles[i].Name < allFiles[j].Name {
			return true
		} else if allFiles[i].Name > allFiles[j].Name {
			return false
		}
		return allFiles[i].ModifiedAt.Before(allFiles[j].ModifiedAt)
	})

	// Backward-compat: when exactly one container path was queried, echo it back
	// as Path so existing callers receive the same field value as before.
	responsePath := ""
	if len(containerPaths) == 1 {
		responsePath = containerPaths[0]
	}

	response.Success(c, http.StatusOK, models.ContainerFiles{
		Path:  responsePath,
		Files: allFiles,
		Total: uint64(len(allFiles)),
	})
}

func (s *FlowFileService) getFlow(c *gin.Context, flowID uint64, writeAccess bool) (models.Flow, error) {
	var flow models.Flow

	uid := c.GetUint64("uid")
	privs := c.GetStringSlice("prm")
	scope := flowScopeForFiles(privs, uid, flowID, writeAccess)
	if scope == nil {
		return flow, response.ErrNotPermitted
	}

	if err := s.db.Model(&flow).Scopes(scope).Take(&flow).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return flow, response.ErrFlowsNotFound
		}
		return flow, err
	}

	return flow, nil
}

// listFlowFiles lists top-level uploads and recursive container cache entries.
// Container entries preserve pulled paths, e.g. container/etc/nginx/nginx.conf.
func (s *FlowFileService) listFlowFiles(flowID uint64) ([]models.FlowFile, error) {
	files, err := flowfiles.List(s.dataDir, flowID)
	if err != nil {
		return nil, err
	}

	return convertFlowFiles(files.Files), nil
}

func (s *FlowFileService) handleFlowLookupError(c *gin.Context, flowID uint64, err error) {
	fields := map[string]any{"flow_id": flowID}

	switch err {
	case response.ErrNotPermitted:
		logger.FromContext(c).WithFields(fields).Error("error filtering user role permissions: permission not found")
		response.Error(c, response.ErrNotPermitted, nil)
	case response.ErrFlowsNotFound:
		logger.FromContext(c).WithFields(fields).Error("error finding flow for flow files")
		response.Error(c, response.ErrFlowsNotFound, err)
	default:
		logger.FromContext(c).WithError(err).WithFields(fields).Error("error loading flow for flow files")
		response.Error(c, response.ErrInternal, err)
	}
}

// flowDataDir returns the root of the flow-scoped local cache directory.
func (s *FlowFileService) flowDataDir(flowID uint64) string {
	return flowfiles.FlowDataDir(s.dataDir, flowID)
}

// flowUploadsDir returns the subdirectory for user-uploaded files.
func (s *FlowFileService) flowUploadsDir(flowID uint64) string {
	return flowfiles.FlowUploadsDir(s.dataDir, flowID)
}

// flowContainerDir returns the subdirectory for files synced from the container.
func (s *FlowFileService) flowContainerDir(flowID uint64) string {
	return flowfiles.FlowContainerDir(s.dataDir, flowID)
}

// resolveCachedPath validates reqPath, ensures it starts with "uploads" or "container",
// and returns the absolute host path. Returns an error if the path would escape flowDataDir.
func (s *FlowFileService) resolveCachedPath(flowID uint64, reqPath string) (string, error) {
	return flowfiles.ResolveCachedPath(s.dataDir, flowID, reqPath)
}

// copyLocalFilesToPrimaryWork streams local regular files into the running primary
// container under /work using TarPath for each entry (e.g. "uploads/a.txt",
// "resources/dir/b.yml"). One tar archive and one CopyToContainer call per invocation.
// Returns nil when Docker is unavailable, the container is not running, or entries is empty.
func (s *FlowFileService) copyLocalFilesToPrimaryWork(ctx context.Context, flowID uint64, entries []flowfiles.TarEntry) error {
	if s.dockerClient == nil || len(entries) == 0 {
		return nil
	}

	containerName := primaryContainerName(flowID)
	running, err := s.dockerClient.IsContainerRunning(ctx, containerName)
	if err != nil || !running {
		return nil
	}

	pr, pw := io.Pipe()
	errCh := make(chan error, 1)
	go func() {
		errCh <- flowfiles.WriteFilesTar(pw, entries)
	}()

	copyErr := s.dockerClient.CopyToContainer(ctx, containerName, docker.WorkFolderPathInContainer, pr,
		container.CopyToContainerOptions{AllowOverwriteDirWithFile: true})
	pr.Close()
	writeErr := <-errCh
	if copyErr != nil {
		return copyErr
	}
	return writeErr
}

// pushResourcePathsToContainer copies flow resource files into the running primary
// container at /work/resources/... in a single tar stream. Best-effort: skips when
// Docker is unavailable or the container is not running.
func (s *FlowFileService) pushResourcePathsToContainer(c *gin.Context, flowID uint64, relPaths []string) {
	if len(relPaths) == 0 {
		return
	}

	resourcesDir := flowfiles.FlowResourcesDir(s.dataDir, flowID)
	prefixLen := len(flowfiles.ResourcesDirName) + 1
	entries := make([]flowfiles.TarEntry, 0, len(relPaths))
	for _, relPath := range relPaths {
		if len(relPath) <= prefixLen {
			continue
		}
		fsRelPath := relPath[prefixLen:]
		entries = append(entries, flowfiles.TarEntry{
			LocalPath: filepath.Join(resourcesDir, filepath.FromSlash(fsRelPath)),
			TarPath:   path.Join(flowfiles.ResourcesDirName, filepath.ToSlash(fsRelPath)),
		})
	}

	if len(entries) == 0 {
		return
	}
	if pushErr := s.copyLocalFilesToPrimaryWork(c.Request.Context(), flowID, entries); pushErr != nil {
		logger.FromContext(c).WithError(pushErr).WithFields(map[string]any{
			"flow_id": flowID,
			"count":   len(entries),
		}).Warn("resource files saved locally but could not be pushed to container")
	}
}

// deleteUploadsFromContainer removes cache paths that mirror the primary container's
// /work tree (uploads/... and resources/...) in a single Docker exec call.
// Container mirror paths (container/...) stay host-only and are not touched inside
// the running container. When Docker is unavailable or the container is not running,
// the function returns nil — the cache divergence will be resolved on next container start.
func (s *FlowFileService) deleteUploadsFromContainer(ctx context.Context, flowID uint64, reqPaths []string) error {
	if s.dockerClient == nil || len(reqPaths) == 0 {
		return nil
	}

	// Collect container-side paths for entries that mirror into /work (uploads/ and resources/ only).
	containerPaths := make([]string, 0, len(reqPaths))
	for _, reqPath := range reqPaths {
		cleaned := filepath.Clean(filepath.FromSlash(strings.ReplaceAll(reqPath, "\\", "/")))
		parts := strings.SplitN(cleaned, string(filepath.Separator), 2)
		if len(parts) == 0 || (parts[0] != flowfiles.UploadsDirName && parts[0] != flowfiles.ResourcesDirName) {
			continue
		}
		containerPaths = append(containerPaths, path.Join(docker.WorkFolderPathInContainer, filepath.ToSlash(cleaned)))
	}
	if len(containerPaths) == 0 {
		return nil
	}

	containerName := primaryContainerName(flowID)
	running, err := s.dockerClient.IsContainerRunning(ctx, containerName)
	if err != nil {
		return nil // container absent or unavailable — cache deletion will be synced on next start
	}
	if !running {
		return nil
	}

	// Build a single rm -rf command for all paths to minimise Docker API round-trips.
	quotedPaths := make([]string, len(containerPaths))
	for i, p := range containerPaths {
		quotedPaths[i] = shellQuote(p)
	}
	cmd := "rm -rf -- " + strings.Join(quotedPaths, " ")

	createResp, err := s.dockerClient.ContainerExecCreate(ctx, containerName, container.ExecOptions{
		Cmd:          []string{"sh", "-c", cmd},
		AttachStdout: true,
		AttachStderr: true,
	})
	if err != nil {
		return fmt.Errorf("failed to create container delete exec: %w", err)
	}

	resp, err := s.dockerClient.ContainerExecAttach(ctx, createResp.ID, container.ExecAttachOptions{})
	if err != nil {
		return fmt.Errorf("failed to attach container delete exec: %w", err)
	}
	output, copyErr := io.ReadAll(resp.Reader)
	resp.Close()
	if copyErr != nil {
		return fmt.Errorf("failed to read container delete output: %w", copyErr)
	}

	inspect, err := s.dockerClient.ContainerExecInspect(ctx, createResp.ID)
	if err != nil {
		return fmt.Errorf("failed to inspect container delete exec: %w", err)
	}
	if inspect.ExitCode != 0 {
		return fmt.Errorf("container delete command failed with exit code %d: %s", inspect.ExitCode, string(output))
	}

	return nil
}

func shellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "'\"'\"'") + "'"
}

// deduplicateContainerPaths returns a slice of trimmed, non-empty, unique
// container path strings, preserving first-occurrence order. Unlike
// flowfiles.DeduplicatePaths it applies no path.Clean or safety checks —
// those semantics are specific to host-side cache paths, not container-internal
// absolute paths.
func deduplicateContainerPaths(paths []string) []string {
	seen := make(map[string]struct{}, len(paths))
	result := make([]string, 0, len(paths))
	for _, p := range paths {
		trimmed := strings.TrimSpace(p)
		if trimmed == "" {
			continue
		}
		if _, dup := seen[trimmed]; dup {
			continue
		}
		seen[trimmed] = struct{}{}
		result = append(result, trimmed)
	}
	return result
}

func flowScopeForFiles(
	privs []string,
	uid uint64,
	flowID uint64,
	writeAccess bool,
) func(db *gorm.DB) *gorm.DB {
	if slices.Contains(privs, "flow_files.admin") {
		return func(db *gorm.DB) *gorm.DB {
			return db.Where("id = ?", flowID)
		}
	}

	if writeAccess && slices.Contains(privs, "flow_files.upload") {
		return func(db *gorm.DB) *gorm.DB {
			return db.Where("id = ? AND user_id = ?", flowID, uid)
		}
	}

	if !writeAccess && slices.Contains(privs, "flow_files.view") {
		return func(db *gorm.DB) *gorm.DB {
			return db.Where("id = ? AND user_id = ?", flowID, uid)
		}
	}

	return nil
}

func parseFlowIDParam(c *gin.Context) (uint64, error) {
	return strconv.ParseUint(c.Param("flowID"), 10, 64)
}

func cleanupPendingUploads(pending []pendingUpload) {
	for _, p := range pending {
		if p.tmpPath != "" {
			os.Remove(p.tmpPath)
		}
	}
}

func convertFlowFiles(files []flowfiles.File) []models.FlowFile {
	converted := make([]models.FlowFile, 0, len(files))
	for _, file := range files {
		converted = append(converted, convertFlowFile(file))
	}
	return converted
}

func convertFlowFile(file flowfiles.File) models.FlowFile {
	return models.FlowFile{
		ID:         file.ID,
		Name:       file.Name,
		Path:       file.Path,
		Size:       file.Size,
		IsDir:      file.IsDir,
		ModifiedAt: file.ModifiedAt,
	}
}

func convertModelFlowFile(file models.FlowFile) *model.FlowFile {
	return &model.FlowFile{
		ID:         file.ID,
		Name:       file.Name,
		Path:       file.Path,
		Size:       int(file.Size),
		IsDir:      file.IsDir,
		ModifiedAt: file.ModifiedAt,
	}
}

func convertContainerFiles(basePath string, stats []container.PathStat) []models.ContainerFile {
	files := make([]models.ContainerFile, 0, len(stats))
	for _, stat := range stats {
		files = append(files, convertContainerFile(basePath, stat))
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i].Name < files[j].Name
	})

	return files
}

func convertContainerFile(basePath string, stat container.PathStat) models.ContainerFile {
	filePath := path.Join(basePath, stat.Name)

	return models.ContainerFile{
		ID:         flowfiles.ID(filePath),
		Name:       stat.Name,
		Path:       filePath,
		Size:       stat.Size,
		IsDir:      stat.Mode.IsDir(),
		ModifiedAt: stat.Mtime,
	}
}

func (s *FlowFileService) publishFlowFilesAdded(ctx context.Context, flow models.Flow, files []models.FlowFile) {
	if s.ss == nil {
		return
	}

	publisher := s.ss.NewFlowPublisher(int64(flow.UserID), int64(flow.ID))
	for _, file := range files {
		publisher.FlowFileAdded(ctx, convertModelFlowFile(file))
	}
}

func (s *FlowFileService) publishFlowFileUpdated(ctx context.Context, flow models.Flow, file models.FlowFile) {
	if s.ss == nil {
		return
	}

	publisher := s.ss.NewFlowPublisher(int64(flow.UserID), int64(flow.ID))
	publisher.FlowFileUpdated(ctx, convertModelFlowFile(file))
}

func (s *FlowFileService) publishFlowFilesDeleted(ctx context.Context, flow models.Flow, files []models.FlowFile) {
	if s.ss == nil {
		return
	}

	publisher := s.ss.NewFlowPublisher(int64(flow.UserID), int64(flow.ID))
	for _, file := range files {
		publisher.FlowFileDeleted(ctx, convertModelFlowFile(file))
	}
}

func sortFlowFiles(files []models.FlowFile) {
	sort.Slice(files, func(i, j int) bool {
		if files[i].ModifiedAt.Equal(files[j].ModifiedAt) {
			return files[i].Name < files[j].Name
		}
		return files[i].ModifiedAt.After(files[j].ModifiedAt)
	})
}

// primaryContainerName returns the Docker container name for a flow's primary terminal.
func primaryContainerName(flowID uint64) string {
	return fmt.Sprintf("%s%d", tools.PrimaryTerminalNamePrefix, flowID)
}

// AddResourcesToFlow copies user-owned resources into the flow resources directory.
// @Summary Copy user resources into a flow
// @Description Copies one or more user resources (identified by ID) into flow-{id}-data/resources/.
// @Description Files already present in the flow are skipped unless force=true, in which case they are replaced.
// @Description When the primary container for this flow is running, new or replaced files are pushed to /work/resources/ (best-effort), same as uploads.
// @Tags FlowFiles
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param flowID path int true "flow id" minimum(0)
// @Param body body models.AddResourcesRequest true "resource IDs and force flag"
// @Success 200 {object} response.successResp{data=models.FlowFiles} "resources copied successfully"
// @Failure 400 {object} response.errorResp "invalid request"
// @Failure 403 {object} response.errorResp "not permitted"
// @Failure 404 {object} response.errorResp "flow not found"
// @Failure 409 {object} response.errorResp "file already exists and force=false"
// @Failure 500 {object} response.errorResp "internal error"
// @Router /flows/{flowID}/files/resources [post]
func (s *FlowFileService) AddResourcesToFlow(c *gin.Context) {
	flowID, err := parseFlowIDParam(c)
	if err != nil {
		logger.FromContext(c).WithError(err).Error("error parsing flow id")
		response.Error(c, response.ErrFlowFilesInvalidRequest, err)
		return
	}

	flow, err := s.getFlow(c, flowID, true)
	if err != nil {
		s.handleFlowLookupError(c, flowID, err)
		return
	}

	var req models.AddResourcesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.FromContext(c).WithError(err).WithField("flow_id", flowID).Error("invalid add resources request")
		response.Error(c, response.ErrFlowFilesInvalidRequest, err)
		return
	}

	uid := c.GetUint64("uid")
	privs := c.GetStringSlice("prm")
	isAdmin := slices.Contains(privs, "resources.admin") || slices.Contains(privs, "flow_files.admin")

	// Check privilege: need both flow_files.upload and resources.view.
	if !isAdmin {
		if !slices.Contains(privs, "flow_files.upload") {
			response.Error(c, response.ErrNotPermitted, nil)
			return
		}
		if !slices.Contains(privs, "resources.view") && !slices.Contains(privs, "resources.admin") {
			response.Error(c, response.ErrNotPermitted, nil)
			return
		}
	}

	// Fetch resources and validate ownership.
	var recs []models.UserResource
	if err := s.db.Model(&models.UserResource{}).Where("id IN (?)", req.IDs).Find(&recs).Error; err != nil {
		logger.FromContext(c).WithError(err).WithField("flow_id", flowID).Error("error fetching resources by IDs")
		response.Error(c, response.ErrInternal, err)
		return
	}

	found := make(map[uint64]models.UserResource, len(recs))
	for _, r := range recs {
		found[r.ID] = r
	}

	refs := make([]flowfiles.ResourceRef, 0, len(req.IDs))
	for _, id := range req.IDs {
		r, ok := found[id]
		if !ok {
			err := fmt.Errorf("resource %d not found", id)
			response.Error(c, response.ErrFlowFilesInvalidRequest, err)
			return
		}
		if !isAdmin && r.UserID != uid {
			response.Error(c, response.ErrNotPermitted, nil)
			return
		}
		refs = append(refs, flowfiles.ResourceRef{
			Hash:        r.Hash,
			VirtualPath: r.Path,
			Name:        r.Name,
			IsDir:       r.IsDir,
		})
	}

	storeDir := filepath.Join(s.dataDir, "resources")
	addedPaths, err := flowfiles.CopyResourcesToFlow(s.dataDir, storeDir, flowID, refs, req.Force)
	if err != nil {
		logger.FromContext(c).WithError(err).WithField("flow_id", flowID).Error("error copying resources to flow")
		response.Error(c, response.ErrInternal, err)
		return
	}

	if len(addedPaths) > 0 {
		s.pushResourcePathsToContainer(c, flowID, addedPaths)
	}

	// Publish subscription events for newly added/updated files.
	if len(addedPaths) > 0 {
		addedFiles := make([]models.FlowFile, 0, len(addedPaths))
		resourcesDir := flowfiles.FlowResourcesDir(s.dataDir, flowID)
		for _, relPath := range addedPaths {
			fsRelPath := relPath[len(flowfiles.ResourcesDirName)+1:]
			absPath := filepath.Join(resourcesDir, filepath.FromSlash(fsRelPath))
			var size int64
			var modTime time.Time
			if info, err := os.Lstat(absPath); err == nil {
				size = info.Size()
				modTime = info.ModTime()
			} else {
				modTime = time.Now()
			}
			addedFiles = append(addedFiles, models.FlowFile{
				ID:         flowfiles.ID(relPath),
				Name:       path.Base(relPath),
				Path:       relPath,
				Size:       size,
				IsDir:      false,
				ModifiedAt: modTime,
			})
		}

		if req.Force {
			for _, f := range addedFiles {
				s.publishFlowFileUpdated(c.Request.Context(), flow, f)
			}
		} else {
			s.publishFlowFilesAdded(c.Request.Context(), flow, addedFiles)
		}
	}

	files, err := s.listFlowFiles(flowID)
	if err != nil {
		logger.FromContext(c).WithError(err).WithField("flow_id", flowID).Error("error listing flow files after resource add")
		response.Error(c, response.ErrInternal, err)
		return
	}

	response.Success(c, http.StatusOK, models.FlowFiles{
		Files: files,
		Total: uint64(len(files)),
	})
}

// AddResourceFromFlow promotes a flow cache file or directory tree into the
// user's global resource store. The response always lists every created or
// updated entry, including any virtual parent directories that were ensured
// during the promotion.
// @Summary Promote a flow file or directory to user resources
// @Tags FlowFiles
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param flowID path int true "flow id" minimum(0)
// @Param body body models.AddResourceFromFlowRequest true "source path, destination and force flag"
// @Success 200 {object} response.successResp{data=models.ResourceList} "resources created or updated"
// @Failure 400 {object} response.errorResp "invalid request"
// @Failure 403 {object} response.errorResp "not permitted"
// @Failure 404 {object} response.errorResp "source or flow not found"
// @Failure 409 {object} response.errorResp "destination already exists and force=false"
// @Failure 500 {object} response.errorResp "internal error"
// @Router /flows/{flowID}/files/to-resources [post]
func (s *FlowFileService) AddResourceFromFlow(c *gin.Context) {
	flowID, err := parseFlowIDParam(c)
	if err != nil {
		logger.FromContext(c).WithError(err).Error("error parsing flow id")
		response.Error(c, response.ErrFlowFilesInvalidRequest, err)
		return
	}

	uid := c.GetUint64("uid")
	privs := c.GetStringSlice("prm")
	if !slices.Contains(privs, "resources.admin") {
		if !slices.Contains(privs, "resources.upload") {
			response.Error(c, response.ErrNotPermitted, nil)
			return
		}
		if !slices.Contains(privs, "flow_files.view") && !slices.Contains(privs, "flow_files.admin") {
			response.Error(c, response.ErrNotPermitted, nil)
			return
		}
	}

	if _, err := s.getFlow(c, flowID, false); err != nil {
		s.handleFlowLookupError(c, flowID, err)
		return
	}

	var req models.AddResourceFromFlowRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.FromContext(c).WithError(err).Error("invalid add resource from flow request")
		response.Error(c, response.ErrFlowFilesInvalidRequest, err)
		return
	}

	absPath, err := flowfiles.ResolveCachedPath(s.dataDir, flowID, req.Source)
	if err != nil {
		logger.FromContext(c).WithError(err).Error("invalid source path for add resource from flow")
		response.Error(c, response.ErrFlowFilesInvalidData, err)
		return
	}

	info, statErr := os.Lstat(absPath)
	if statErr != nil {
		if os.IsNotExist(statErr) {
			response.Error(c, response.ErrFlowFilesNotFound, statErr)
			return
		}
		logger.FromContext(c).WithError(statErr).Error("error stating flow source")
		response.Error(c, response.ErrInternal, statErr)
		return
	}
	if !info.IsDir() && !info.Mode().IsRegular() {
		// Reject symlinks and other special files: os.Open would follow
		// symlinks at the OS level, bypassing the cache-root constraint.
		response.Error(c, response.ErrFlowFilesInvalidData,
			fmt.Errorf("source must be a regular file or a directory"))
		return
	}

	destPath, err := resources.SanitizeResourcePath(req.Destination)
	if err != nil {
		logger.FromContext(c).WithError(err).Error("invalid destination path for add resource from flow")
		response.Error(c, response.ErrFlowFilesInvalidData, err)
		return
	}

	saved, ok := s.promoteToResources(c, uid, absPath, destPath, info.IsDir(), req.Force)
	if !ok {
		return
	}
	response.Success(c, http.StatusOK,
		models.ResourceList{Items: saved, Total: uint64(len(saved))})
}

// promoteToResources promotes a single regular file or a directory tree from
// the flow cache into the user's resource store. Mirrors the three-phase
// pattern used by UploadResources:
//  1. walk and stream every regular file to a temporary blob (MD5-keyed)
//  2. atomically commit blobs into the content-addressed store
//  3. persist all directory and file rows inside one DB transaction
//
// Returns the slice of resulting entries (newly-created dirs, added files,
// updated files in that order) on success, or false after writing an error
// response on failure.
func (s *FlowFileService) promoteToResources(
	c *gin.Context, uid uint64,
	absSourceRoot, destPath string,
	sourceIsDir, force bool,
) ([]models.ResourceEntry, bool) {
	log := logger.FromContext(c)
	ctx := c.Request.Context()

	if err := resources.EnsureResourcesDir(s.dataDir); err != nil {
		log.WithError(err).Error("failed to ensure resources directory")
		response.Error(c, response.ErrInternal, err)
		return nil, false
	}
	blobsDir := resources.ResourcesDir(s.dataDir)

	// ── Phase 1: stream every regular file to a temporary blob ────────────────
	var pending []pendingResourceUpload
	walkErr := filepath.Walk(absSourceRoot, func(filePath string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// filepath.Walk uses Lstat → symlinks and special files are NOT regular
		// and are skipped here (defense-in-depth against symlinked-target reads).
		if !fi.Mode().IsRegular() {
			return nil
		}

		var virtPath string
		if sourceIsDir {
			rel, relErr := filepath.Rel(absSourceRoot, filePath)
			if relErr != nil {
				return fmt.Errorf("relative path: %w", relErr)
			}
			sanitized, sanitizeErr := resources.SanitizeResourcePath(destPath + "/" + filepath.ToSlash(rel))
			if sanitizeErr != nil {
				log.WithError(sanitizeErr).Warnf("skipping unsafe file name %q", rel)
				return nil
			}
			virtPath = sanitized
		} else {
			virtPath = destPath
		}

		// Phase 1 conflict pre-check: avoid IO when the destination obviously
		// conflicts. Phase 3 re-checks inside the TX as the source of truth.
		if !force {
			_, exists, dbErr := findResourceByPath(s.db, uid, virtPath)
			if dbErr != nil {
				return fmt.Errorf("checking %s: %w", virtPath, dbErr)
			}
			if exists {
				return fmt.Errorf("%w: resource already exists at %q", errResourceConflict, virtPath)
			}
		}

		f, openErr := os.Open(filePath)
		if openErr != nil {
			return fmt.Errorf("open %s: %w", filePath, openErr)
		}
		tmpPath, hash, size, saveErr := resources.SaveToTemp(f, blobsDir)
		f.Close()
		if saveErr != nil {
			return fmt.Errorf("save temp for %s: %w", filePath, saveErr)
		}
		pending = append(pending, pendingResourceUpload{
			name:    path.Base(virtPath),
			vPath:   virtPath,
			tmpPath: tmpPath,
			hash:    hash,
			size:    size,
		})
		return nil
	})
	if walkErr != nil {
		cleanupResourceUploads(pending)
		if errors.Is(walkErr, errResourceConflict) {
			response.Error(c, response.ErrFlowFilesAlreadyExists, walkErr)
		} else {
			log.WithError(walkErr).Error("error reading source files for promotion")
			response.Error(c, response.ErrInternal, walkErr)
		}
		return nil, false
	}

	// ── Phase 2: commit blobs (atomic rename; idempotent on duplicate hashes) ─
	for i := range pending {
		p := &pending[i]
		existed, blobErr := resources.BlobExists(s.dataDir, p.hash)
		if blobErr != nil {
			cleanupResourceUploads(pending[i:])
			s.deleteOrphanBlobsIfUnreferenced(pending[:i])
			log.WithError(blobErr).Error("error checking blob existence")
			response.Error(c, response.ErrInternal, blobErr)
			return nil, false
		}
		if commitErr := resources.CommitBlob(s.dataDir, p.hash, p.tmpPath); commitErr != nil {
			cleanupResourceUploads(pending[i:])
			s.deleteOrphanBlobsIfUnreferenced(pending[:i])
			log.WithError(commitErr).Error("error committing blob")
			response.Error(c, response.ErrInternal, commitErr)
			return nil, false
		}
		p.tmpPath = ""
		p.newBlob = !existed
	}

	// Determine the directory chain to ensure: the destination itself for a
	// directory promotion, or the parent of the file otherwise. May be empty
	// when promoting a single file at the resource root.
	parentDir := destPath
	if !sourceIsDir {
		if d := path.Dir(destPath); d != "." && d != "/" {
			parentDir = d
		} else {
			parentDir = ""
		}
	}

	// ── Phase 3: persist all rows in a single DB transaction ──────────────────
	tx := s.db.Begin()
	if tx.Error != nil {
		s.deleteOrphanBlobsIfUnreferenced(pending)
		log.WithError(tx.Error).Error("failed to begin transaction")
		response.Error(c, response.ErrInternal, tx.Error)
		return nil, false
	}

	var allDirs, added, updated []models.UserResource
	var orphanHashes []string
	var txErr error

	if parentDir != "" {
		rootDirs, _, rootOrphans, dirErr := ensureResourceDirs(tx, uid, parentDir, force)
		if dirErr != nil {
			txErr = dirErr
		} else {
			allDirs = append(allDirs, rootDirs...)
			orphanHashes = append(orphanHashes, rootOrphans...)
		}
	}

	// Ensure each unique parent directory chain only once across all files.
	ensuredDirs := make(map[string]bool, len(pending))
	if parentDir != "" {
		ensuredDirs[parentDir] = true
	}
	for _, p := range pending {
		if txErr != nil {
			break
		}

		if subParent := path.Dir(p.vPath); subParent != "." && subParent != "/" && !ensuredDirs[subParent] {
			ensuredDirs[subParent] = true
			subDirs, _, subOrphans, dirErr := ensureResourceDirs(tx, uid, subParent, force)
			if dirErr != nil {
				txErr = dirErr
				break
			}
			allDirs = append(allDirs, subDirs...)
			orphanHashes = append(orphanHashes, subOrphans...)
		}

		existing, exists, lookErr := findResourceByPath(tx, uid, p.vPath)
		if lookErr != nil {
			txErr = lookErr
			break
		}
		if exists {
			if !force {
				txErr = fmt.Errorf("%w: resource already exists at %q", errResourceConflict, p.vPath)
				break
			}
			if existing.IsDir {
				txErr = fmt.Errorf("%w: a directory already exists at %q", errResourceConflict, p.vPath)
				break
			}
			if existing.Hash != "" && existing.Hash != p.hash {
				orphanHashes = append(orphanHashes, existing.Hash)
			}
			now := time.Now()
			if updErr := tx.Model(&existing).Updates(models.UserResource{
				Hash: p.hash, Name: p.name, Size: p.size, UpdatedAt: now,
			}).Error; updErr != nil {
				txErr = updErr
				break
			}
			existing.Hash, existing.Name, existing.Size, existing.UpdatedAt = p.hash, p.name, p.size, now
			updated = append(updated, existing)
			continue
		}

		rec := models.UserResource{
			UserID: uid, Hash: p.hash, Name: p.name, Path: p.vPath, Size: p.size,
		}
		if createErr := tx.Create(&rec).Error; createErr != nil {
			txErr = createErr
			break
		}
		added = append(added, rec)
	}

	if txErr != nil {
		tx.Rollback()
		s.deleteOrphanBlobsIfUnreferenced(pending)
		if isUniqueViolation(txErr) || errors.Is(txErr, errResourceConflict) {
			response.Error(c, response.ErrFlowFilesAlreadyExists, txErr)
		} else {
			log.WithError(txErr).Error("failed to persist promoted resources")
			response.Error(c, response.ErrInternal, txErr)
		}
		return nil, false
	}
	if commitErr := tx.Commit().Error; commitErr != nil {
		tx.Rollback()
		s.deleteOrphanBlobsIfUnreferenced(pending)
		log.WithError(commitErr).Error("failed to commit promotion transaction")
		response.Error(c, response.ErrInternal, commitErr)
		return nil, false
	}

	s.cleanupOrphanHashes(orphanHashes)

	if s.ss != nil {
		pub := s.ss.NewResourcePublisher(int64(uid))
		for _, r := range allDirs {
			pub.ResourceAdded(ctx, convertResourceToModel(convertResource(r)))
		}
		for _, r := range added {
			pub.ResourceAdded(ctx, convertResourceToModel(convertResource(r)))
		}
		for _, r := range updated {
			pub.ResourceUpdated(ctx, convertResourceToModel(convertResource(r)))
		}
	}

	out := convertResources(allDirs)
	out = append(out, convertResources(added)...)
	out = append(out, convertResources(updated)...)
	return out, true
}

// deleteOrphanBlobsIfUnreferenced removes freshly-committed blobs from this
// promotion ONLY if no DB row references them. Mirrors
// ResourceService.deleteOrphanBlob — used after a rolled-back transaction or
// a partial Phase 2 to avoid deleting blobs that other concurrent uploads of
// identical content may now reference.
func (s *FlowFileService) deleteOrphanBlobsIfUnreferenced(pending []pendingResourceUpload) {
	for _, p := range pending {
		if !p.newBlob || p.hash == "" {
			continue
		}
		var count int64
		if err := s.db.Model(&models.UserResource{}).
			Where("hash = ?", p.hash).
			Count(&count).Error; err != nil || count > 0 {
			continue
		}
		_ = resources.DeleteBlob(s.dataDir, p.hash)
	}
}

// cleanupOrphanHashes mirrors ResourceService.cleanupOrphanBlobs: it removes
// .blob files whose hashes are no longer referenced by any user_resources row,
// after a successful TX that may have orphaned blobs (overwrites).
func (s *FlowFileService) cleanupOrphanHashes(hashes []string) {
	if len(hashes) == 0 {
		return
	}
	var stillReferenced []string
	if err := s.db.Model(&models.UserResource{}).
		Where("hash IN (?)", hashes).
		Pluck("DISTINCT hash", &stillReferenced).Error; err != nil {
		return
	}
	refSet := make(map[string]bool, len(stillReferenced))
	for _, h := range stillReferenced {
		refSet[h] = true
	}
	seen := make(map[string]bool, len(hashes))
	for _, h := range hashes {
		if h == "" || refSet[h] || seen[h] {
			continue
		}
		seen[h] = true
		_ = resources.DeleteBlob(s.dataDir, h)
	}
}
