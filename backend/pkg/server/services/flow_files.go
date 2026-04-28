package services

import (
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

		// Best-effort push to the running primary container so the file is
		// immediately visible at /work/uploads/<name> without a container restart.
		if pushErr := s.copyFileToContainer(c.Request.Context(), flowID, p.dstPath); pushErr != nil {
			logger.FromContext(c).WithError(pushErr).WithFields(map[string]any{
				"flow_id":   flowID,
				"file_name": p.fileName,
			}).Warn("uploaded file saved locally but could not be pushed to container")
		}

		savedFiles = append(savedFiles, convertFlowFile(flowfiles.NewFile(info, flowfiles.UploadsDirName)))
	}

	sortFlowFiles(savedFiles)
	s.publishFlowFilesAdded(c.Request.Context(), flow, savedFiles)
	response.Success(c, http.StatusOK, models.FlowFiles{
		Files: savedFiles,
		Total: uint64(len(savedFiles)),
	})
}

// DeleteFlowFile is a function to delete a cached flow file or directory by path
// @Summary Delete flow file or directory by cached path
// @Tags FlowFiles
// @Produce json
// @Security BearerAuth
// @Param flowID path int true "flow id" minimum(0)
// @Param path query string true "relative path in cache: uploads/<name> or container/<name>"
// @Success 200 {object} response.successResp{data=models.FlowFiles} "flow file deleted successful"
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

	localPath, err := s.resolveCachedPath(flowID, c.Query("path"))
	if err != nil {
		logger.FromContext(c).WithError(err).WithField("flow_id", flowID).Error("invalid delete path")
		response.Error(c, response.ErrFlowFilesInvalidRequest, err)
		return
	}

	info, err := os.Lstat(localPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			response.Error(c, response.ErrFlowFilesNotFound, fmt.Errorf("'%s' not found in local cache", c.Query("path")))
		} else {
			logger.FromContext(c).WithError(err).WithField("flow_id", flowID).Error("error reading cached flow file")
			response.Error(c, response.ErrInternal, err)
		}
		return
	}
	deletedFile := convertFlowFile(flowfiles.NewFileWithPath(info, filepath.ToSlash(c.Query("path"))))

	if err := s.deleteUploadFromContainer(c.Request.Context(), flowID, c.Query("path")); err != nil {
		logger.FromContext(c).WithError(err).WithField("flow_id", flowID).Error("error deleting uploaded file from container")
		response.Error(c, response.ErrInternal, err)
		return
	}

	if err := os.RemoveAll(localPath); err != nil {
		logger.FromContext(c).WithError(err).WithField("flow_id", flowID).Error("error deleting cached flow file")
		response.Error(c, response.ErrInternal, err)
		return
	}

	s.publishFlowFileDeleted(c.Request.Context(), flow, deletedFile)
	response.Success(c, http.StatusOK, models.FlowFiles{
		Files: nil,
		Total: 0,
	})
}

// DownloadFlowFile is a function to download a flow file or directory (as ZIP) by cached path
// @Summary Download flow file or directory by cached path
// @Tags FlowFiles
// @Produce octet-stream,application/zip,json
// @Security BearerAuth
// @Param flowID path int true "flow id" minimum(0)
// @Param path query string true "relative path in cache: uploads/<name> or container/<name>"
// @Success 200 {file} binary "file content, or ZIP archive for directories"
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

	localPath, err := s.resolveCachedPath(flowID, c.Query("path"))
	if err != nil {
		logger.FromContext(c).WithError(err).WithField("flow_id", flowID).Error("invalid download path")
		response.Error(c, response.ErrFlowFilesInvalidRequest, err)
		return
	}

	info, err := os.Lstat(localPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			response.Error(c, response.ErrFlowFilesNotFound, fmt.Errorf("'%s' not found in local cache", c.Query("path")))
		} else {
			logger.FromContext(c).WithError(err).WithField("flow_id", flowID).Error("error reading cached flow file")
			response.Error(c, response.ErrInternal, err)
		}
		return
	}

	// Never serve symlinks — could point outside the flow data directory.
	if info.Mode()&os.ModeSymlink != 0 {
		response.Error(c, response.ErrFlowFilesNotFound, fmt.Errorf("'%s' not found in local cache", c.Query("path")))
		return
	}

	if info.IsDir() {
		name := filepath.Base(localPath)
		c.Header("Content-Disposition", mime.FormatMediaType("attachment", map[string]string{
			"filename": name + ".zip",
		}))
		c.Header("Content-Type", "application/zip")
		if err := flowfiles.ZipDirectory(c.Writer, localPath); err != nil {
			logger.FromContext(c).WithError(err).WithField("flow_id", flowID).Error("error creating ZIP archive")
		}
		return
	}

	if !info.Mode().IsRegular() {
		response.Error(c, response.ErrFlowFilesNotFound, fmt.Errorf("'%s' not found in local cache", c.Query("path")))
		return
	}

	c.FileAttachment(localPath, filepath.Base(localPath))
}

// PullFlowFiles is a function to sync a path from the container into the local cache
// @Summary Pull a file or directory from the running container into the local cache
// @Tags FlowFiles
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param flowID path int true "flow id" minimum(0)
// @Param request body models.PullFlowFilesRequest true "pull request"
// @Success 200 {object} response.successResp{data=models.FlowFiles} "container path synced to local cache"
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

	containerPath := strings.TrimSpace(req.Path)
	if containerPath == "" {
		err = errors.New("path is required")
		logger.FromContext(c).WithError(err).WithField("flow_id", flowID).Error("missing container path in pull request")
		response.Error(c, response.ErrFlowFilesInvalidRequest, err)
		return
	}

	// Preserve the normalized container path under the local container cache:
	// /etc/nginx/conf/ → container/etc/nginx/conf/
	// /etc/nginx/nginx.conf → container/etc/nginx/nginx.conf
	cacheRelPath, err := flowfiles.SanitizeContainerCachePath(containerPath)
	if err != nil {
		logger.FromContext(c).WithError(err).WithFields(map[string]any{
			"flow_id":        flowID,
			"container_path": containerPath,
		}).Error("invalid container path")
		response.Error(c, response.ErrFlowFilesInvalidRequest, fmt.Errorf("invalid container path: %w", err))
		return
	}

	containerDir := s.flowContainerDir(flowID)
	localTarget := filepath.Join(containerDir, filepath.FromSlash(cacheRelPath))
	targetExists := false

	// Check for existing cache entry; honour force flag.
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

	reader, _, err := s.dockerClient.CopyFromContainer(c.Request.Context(), containerName, containerPath)
	if err != nil {
		logger.FromContext(c).WithError(err).WithFields(map[string]any{
			"flow_id":        flowID,
			"container_path": containerPath,
		}).Error("error copying from container")
		response.Error(c, response.ErrInternal, err)
		return
	}
	defer reader.Close()

	stagingDir, err := os.MkdirTemp(containerDir, ".pull-*")
	if err != nil {
		logger.FromContext(c).WithError(err).WithField("flow_id", flowID).Error("error creating pull staging directory")
		response.Error(c, response.ErrInternal, err)
		return
	}
	defer os.RemoveAll(stagingDir)

	if err := flowfiles.ExtractTar(reader, stagingDir); err != nil {
		logger.FromContext(c).WithError(err).WithField("flow_id", flowID).Error("error extracting container TAR archive")
		response.Error(c, response.ErrInternal, err)
		return
	}

	stagedTarget := flowfiles.ResolvePulledStagedTarget(stagingDir, cacheRelPath)
	if stagedTarget == "" {
		err = fmt.Errorf("pulled archive did not contain expected entry '%s'", cacheRelPath)
		logger.FromContext(c).WithError(err).WithFields(map[string]any{
			"flow_id":    flowID,
			"cache_path": cacheRelPath,
		}).Error("pulled container archive did not contain expected entry")
		response.Error(c, response.ErrInternal, err)
		return
	}
	if _, err := os.Lstat(stagedTarget); err != nil {
		logger.FromContext(c).WithError(err).WithFields(map[string]any{
			"flow_id":    flowID,
			"cache_path": cacheRelPath,
		}).Error("pulled container archive did not contain expected entry")
		response.Error(c, response.ErrInternal, err)
		return
	}

	if req.Force {
		if err := os.RemoveAll(localTarget); err != nil {
			logger.FromContext(c).WithError(err).WithFields(map[string]any{
				"flow_id":    flowID,
				"cache_path": cacheRelPath,
			}).Error("error removing existing cache entry before forced pull")
			response.Error(c, response.ErrInternal, err)
			return
		}
	}
	if err := os.MkdirAll(filepath.Dir(localTarget), 0755); err != nil {
		logger.FromContext(c).WithError(err).WithFields(map[string]any{
			"flow_id":    flowID,
			"cache_path": cacheRelPath,
		}).Error("error creating container cache parent directory")
		response.Error(c, response.ErrInternal, err)
		return
	}
	if err := os.Rename(stagedTarget, localTarget); err != nil {
		logger.FromContext(c).WithError(err).WithFields(map[string]any{
			"flow_id":    flowID,
			"cache_path": cacheRelPath,
		}).Error("error committing pulled container entry")
		response.Error(c, response.ErrInternal, err)
		return
	}

	info, err := os.Lstat(localTarget)
	if err != nil {
		logger.FromContext(c).WithError(err).WithFields(map[string]any{
			"flow_id":    flowID,
			"cache_path": cacheRelPath,
		}).Error("error stating synced container entry")
		response.Error(c, response.ErrInternal, err)
		return
	}

	synced := convertFlowFile(flowfiles.NewFileWithPath(info, path.Join(flowfiles.ContainerDirName, cacheRelPath)))
	if targetExists {
		s.publishFlowFileUpdated(c.Request.Context(), flow, synced)
	} else {
		s.publishFlowFilesAdded(c.Request.Context(), flow, []models.FlowFile{synced})
	}
	response.Success(c, http.StatusOK, models.FlowFiles{
		Files: []models.FlowFile{synced},
		Total: 1,
	})
}

// GetFlowContainerFiles is a function to return non-recursive container directory files list
// @Summary Retrieve flow container directory files list
// @Tags FlowFiles
// @Produce json
// @Security BearerAuth
// @Param flowID path int true "flow id" minimum(0)
// @Param path query string false "absolute path inside the running container; defaults to /work"
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

	containerPath := strings.TrimSpace(c.Query("path"))
	if containerPath == "" {
		containerPath = docker.WorkFolderPathInContainer
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
		response.Success(c, http.StatusOK, models.ContainerFiles{
			Path:  containerPath,
			Files: []models.ContainerFile{file},
			Total: 1,
		})
		return
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

	files := convertContainerFiles(containerPath, stats)
	response.Success(c, http.StatusOK, models.ContainerFiles{
		Path:  containerPath,
		Files: files,
		Total: uint64(len(files)),
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

// copyFileToContainer packages a single local file as a TAR archive and uploads it
// to the running primary container at /work/uploads/<filename>. Best-effort: errors are
// returned but callers should only log them as warnings.
func (s *FlowFileService) copyFileToContainer(ctx context.Context, flowID uint64, localFilePath string) error {
	if s.dockerClient == nil {
		return nil
	}

	containerName := primaryContainerName(flowID)
	running, err := s.dockerClient.IsContainerRunning(ctx, containerName)
	if err != nil || !running {
		return nil // container absent or not running — skip silently
	}

	pr, pw := io.Pipe()
	errCh := make(chan error, 1)
	go func() {
		tarName := path.Join(flowfiles.UploadsDirName, filepath.Base(localFilePath))
		errCh <- flowfiles.WriteSingleFileTar(pw, localFilePath, tarName)
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

func (s *FlowFileService) deleteUploadFromContainer(ctx context.Context, flowID uint64, reqPath string) error {
	if s.dockerClient == nil {
		return nil
	}

	cleaned := filepath.Clean(filepath.FromSlash(strings.ReplaceAll(reqPath, "\\", "/")))
	parts := strings.SplitN(cleaned, string(filepath.Separator), 2)
	if len(parts) == 0 || parts[0] != flowfiles.UploadsDirName {
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

	containerPath := path.Join(docker.WorkFolderPathInContainer, filepath.ToSlash(cleaned))
	createResp, err := s.dockerClient.ContainerExecCreate(ctx, containerName, container.ExecOptions{
		Cmd:          []string{"sh", "-c", fmt.Sprintf("rm -rf -- %s", shellQuote(containerPath))},
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

func (s *FlowFileService) publishFlowFileDeleted(ctx context.Context, flow models.Flow, file models.FlowFile) {
	if s.ss == nil {
		return
	}

	publisher := s.ss.NewFlowPublisher(int64(flow.UserID), int64(flow.ID))
	publisher.FlowFileDeleted(ctx, convertModelFlowFile(file))
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

	found := make(map[string]models.UserResource, len(recs))
	for _, r := range recs {
		found[fmt.Sprintf("%d", r.ID)] = r
	}

	refs := make([]flowfiles.ResourceRef, 0, len(req.IDs))
	for _, idStr := range req.IDs {
		r, ok := found[idStr]
		if !ok {
			err := fmt.Errorf("resource %s not found", idStr)
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

// AddResourceFromFlow promotes a file from the flow cache into the user's global resource store.
// @Summary Promote a flow file to user resources
// @Description Reads a file from flow-{id}-data/{sourcePath}, computes MD5, stores the blob, inserts a user_resources row.
// @Tags FlowFiles
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param flowID path int true "flow id" minimum(0)
// @Param body body models.AddResourceFromFlowRequest true "source path, destination and force flag"
// @Success 200 {object} response.successResp{data=models.ResourceEntry} "resource created or updated"
// @Failure 400 {object} response.errorResp "invalid request"
// @Failure 403 {object} response.errorResp "not permitted"
// @Failure 404 {object} response.errorResp "source file or flow not found"
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
	isAdmin := slices.Contains(privs, "resources.admin")

	if !isAdmin {
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

	absPath, err := flowfiles.ResolveCachedPath(s.dataDir, flowID, req.SourcePath)
	if err != nil {
		logger.FromContext(c).WithError(err).Error("invalid source path for add resource from flow")
		response.Error(c, response.ErrFlowFilesInvalidData, err)
		return
	}

	if _, statErr := os.Lstat(absPath); statErr != nil {
		if os.IsNotExist(statErr) {
			response.Error(c, response.ErrFlowFilesNotFound, statErr)
			return
		}
		logger.FromContext(c).WithError(statErr).Error("error stating flow file")
		response.Error(c, response.ErrInternal, statErr)
		return
	}

	destPath, err := resources.SanitizeResourcePath(req.Destination)
	if err != nil {
		logger.FromContext(c).WithError(err).Error("invalid destination path for add resource from flow")
		response.Error(c, response.ErrFlowFilesInvalidData, err)
		return
	}

	var existingRec models.UserResource
	existErr := s.db.Where("user_id = ? AND path = ?", uid, destPath).First(&existingRec).Error
	exists := existErr == nil
	if existErr != nil && !gorm.IsRecordNotFoundError(existErr) {
		logger.FromContext(c).WithError(existErr).Error("error checking resource existence")
		response.Error(c, response.ErrInternal, existErr)
		return
	}
	if exists && !req.Force {
		response.Error(c, response.ErrFlowFilesAlreadyExists, fmt.Errorf("resource already exists at %q; use force=true to overwrite", destPath))
		return
	}

	f, err := os.Open(absPath)
	if err != nil {
		logger.FromContext(c).WithError(err).Error("error opening flow file")
		response.Error(c, response.ErrInternal, err)
		return
	}
	tmpPath, hash, size, err := resources.SaveToTemp(f, resources.ResourcesDir(s.dataDir))
	f.Close()
	if err != nil {
		logger.FromContext(c).WithError(err).Error("error saving flow file to temp")
		response.Error(c, response.ErrInternal, err)
		return
	}

	if err := resources.CommitBlob(s.dataDir, hash, tmpPath); err != nil {
		logger.FromContext(c).WithError(err).Error("error committing blob for flow file")
		response.Error(c, response.ErrInternal, err)
		return
	}

	name := filepath.Base(destPath)

	if exists {
		oldHash := existingRec.Hash
		if err := s.db.Model(&existingRec).Updates(models.UserResource{
			Hash:      hash,
			Name:      name,
			Size:      size,
			UpdatedAt: time.Now(),
		}).Error; err != nil {
			logger.FromContext(c).WithError(err).Error("error updating resource record")
			response.Error(c, response.ErrInternal, err)
			return
		}
		if s.db.Model(&models.UserResource{}).Where("hash = ?", oldHash).RecordNotFound() {
			os.Remove(resources.BlobPath(s.dataDir, oldHash))
		}
		entry := models.ResourceEntry{
			ID:        int64(existingRec.ID),
			UserID:    existingRec.UserID,
			Name:      existingRec.Name,
			Path:      existingRec.Path,
			Size:      size,
			IsDir:     false,
			CreatedAt: existingRec.CreatedAt,
			UpdatedAt: existingRec.UpdatedAt,
		}
		if s.ss != nil {
			s.ss.NewResourcePublisher(int64(uid)).ResourceUpdated(c.Request.Context(), convertResourceToModel(entry))
		}
		response.Success(c, http.StatusOK, entry)
		return
	}

	rec := models.UserResource{
		UserID: uid,
		Hash:   hash,
		Name:   name,
		Path:   destPath,
		Size:   size,
		IsDir:  false,
	}
	if err := s.db.Create(&rec).Error; err != nil {
		logger.FromContext(c).WithError(err).Error("error inserting resource from flow")
		response.Error(c, response.ErrInternal, err)
		return
	}

	entry := models.ResourceEntry{
		ID:        int64(rec.ID),
		UserID:    rec.UserID,
		Name:      rec.Name,
		Path:      rec.Path,
		Size:      rec.Size,
		IsDir:     false,
		CreatedAt: rec.CreatedAt,
		UpdatedAt: rec.UpdatedAt,
	}
	if s.ss != nil {
		s.ss.NewResourcePublisher(int64(uid)).ResourceAdded(c.Request.Context(), convertResourceToModel(entry))
	}
	response.Success(c, http.StatusOK, entry)
}
