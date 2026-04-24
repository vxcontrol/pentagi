package services

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"

	"pentagi/pkg/server/logger"
	"pentagi/pkg/server/models"
	"pentagi/pkg/server/response"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

const flowUploadsDirName = "uploads"

var errFlowFileNotRegular = errors.New("flow file is not a regular file")

type flowFile struct {
	Name       string    `json:"name"`
	Path       string    `json:"path"`
	Size       int64     `json:"size"`
	ModifiedAt time.Time `json:"modifiedAt"`
}

type flowFiles struct {
	Files []flowFile `json:"files"`
	Total uint64     `json:"total"`
}

type FlowFileService struct {
	dataDir string
	db      *gorm.DB
}

func NewFlowFileService(db *gorm.DB, dataDir string) *FlowFileService {
	return &FlowFileService{
		dataDir: dataDir,
		db:      db,
	}
}

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

	response.Success(c, http.StatusOK, flowFiles{
		Files: files,
		Total: uint64(len(files)),
	})
}

func (s *FlowFileService) UploadFlowFiles(c *gin.Context) {
	flowID, err := parseFlowIDParam(c)
	if err != nil {
		logger.FromContext(c).WithError(err).Error("error parsing flow id")
		response.Error(c, response.ErrFlowFilesInvalidRequest, err)
		return
	}

	if _, err := s.getFlow(c, flowID, true); err != nil {
		s.handleFlowLookupError(c, flowID, err)
		return
	}

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

	uploadDir := s.flowUploadsDir(flowID)
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		logger.FromContext(c).WithError(err).WithField("flow_id", flowID).Error("error creating upload directory")
		response.Error(c, response.ErrInternal, err)
		return
	}

	savedFiles := make([]flowFile, 0, len(fileHeaders))
	for _, fileHeader := range fileHeaders {
		fileName, err := sanitizeFlowFileName(fileHeader.Filename)
		if err != nil {
			logger.FromContext(c).WithError(err).WithField("flow_id", flowID).Error("invalid uploaded file name")
			response.Error(c, response.ErrFlowFilesInvalidData, err)
			return
		}

		dstPath := filepath.Join(uploadDir, fileName)
		exists, err := flowFileExists(dstPath)
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

		if err := c.SaveUploadedFile(fileHeader, dstPath); err != nil {
			logger.FromContext(c).WithError(err).WithFields(map[string]any{
				"flow_id":   flowID,
				"file_name": fileName,
			}).Error("error saving uploaded file")
			response.Error(c, response.ErrInternal, err)
			return
		}

		info, err := regularFlowFileInfo(dstPath)
		if err != nil {
			logger.FromContext(c).WithError(err).WithFields(map[string]any{
				"flow_id":   flowID,
				"file_name": fileName,
			}).Error("error stating uploaded file")
			response.Error(c, response.ErrInternal, err)
			return
		}

		savedFiles = append(savedFiles, newFlowFile(info))
	}

	sortFlowFiles(savedFiles)
	response.Success(c, http.StatusOK, flowFiles{
		Files: savedFiles,
		Total: uint64(len(savedFiles)),
	})
}

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

	fileName, err := sanitizeFlowFileName(c.Param("fileName"))
	if err != nil {
		logger.FromContext(c).WithError(err).WithField("flow_id", flowID).Error("invalid download file name")
		response.Error(c, response.ErrFlowFilesInvalidRequest, err)
		return
	}

	filePath := filepath.Join(s.flowUploadsDir(flowID), fileName)
	_, err = regularFlowFileInfo(filePath)
	if err != nil {
		logger.FromContext(c).WithError(err).WithFields(map[string]any{
			"flow_id":   flowID,
			"file_name": fileName,
		}).Error("error reading flow file")
		if errors.Is(err, os.ErrNotExist) || errors.Is(err, errFlowFileNotRegular) {
			response.Error(c, response.ErrFlowFilesNotFound, err)
		} else {
			response.Error(c, response.ErrInternal, err)
		}
		return
	}

	c.FileAttachment(filePath, fileName)
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

func (s *FlowFileService) listFlowFiles(flowID uint64) ([]flowFile, error) {
	entries, err := os.ReadDir(s.flowUploadsDir(flowID))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return []flowFile{}, nil
		}
		return nil, err
	}

	files := make([]flowFile, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || entry.Type()&os.ModeSymlink != 0 {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			return nil, err
		}
		if !info.Mode().IsRegular() {
			continue
		}

		files = append(files, newFlowFile(info))
	}

	sortFlowFiles(files)
	return files, nil
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

func (s *FlowFileService) flowUploadsDir(flowID uint64) string {
	return filepath.Join(s.dataDir, fmt.Sprintf("flow-%d", flowID), flowUploadsDirName)
}

func flowScopeForFiles(
	privs []string,
	uid uint64,
	flowID uint64,
	writeAccess bool,
) func(db *gorm.DB) *gorm.DB {
	if slices.Contains(privs, "flows.admin") {
		return func(db *gorm.DB) *gorm.DB {
			return db.Where("id = ?", flowID)
		}
	}

	if writeAccess && slices.Contains(privs, "flows.edit") {
		return func(db *gorm.DB) *gorm.DB {
			return db.Where("id = ? AND user_id = ?", flowID, uid)
		}
	}

	if !writeAccess && slices.Contains(privs, "flows.view") {
		return func(db *gorm.DB) *gorm.DB {
			return db.Where("id = ? AND user_id = ?", flowID, uid)
		}
	}

	return nil
}

func parseFlowIDParam(c *gin.Context) (uint64, error) {
	return strconv.ParseUint(c.Param("flowID"), 10, 64)
}

func sanitizeFlowFileName(fileName string) (string, error) {
	trimmedName := strings.TrimSpace(fileName)
	if trimmedName == "" {
		return "", fmt.Errorf("file name is required")
	}

	normalizedName := strings.ReplaceAll(trimmedName, "\\", "/")
	cleanName := path.Base(path.Clean("/" + normalizedName))
	if cleanName == "." || cleanName == "/" || cleanName == "" {
		return "", fmt.Errorf("invalid file name")
	}

	return cleanName, nil
}

func flowFileExists(filePath string) (bool, error) {
	if _, err := os.Lstat(filePath); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

func regularFlowFileInfo(filePath string) (os.FileInfo, error) {
	info, err := os.Lstat(filePath)
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() {
		return nil, fmt.Errorf("%w: %s", errFlowFileNotRegular, filePath)
	}

	return info, nil
}

func newFlowFile(info os.FileInfo) flowFile {
	return flowFile{
		Name:       info.Name(),
		Path:       path.Join("/work", flowUploadsDirName, info.Name()),
		Size:       info.Size(),
		ModifiedAt: info.ModTime(),
	}
}

func sortFlowFiles(files []flowFile) {
	sort.Slice(files, func(i, j int) bool {
		if files[i].ModifiedAt.Equal(files[j].ModifiedAt) {
			return files[i].Name < files[j].Name
		}

		return files[i].ModifiedAt.After(files[j].ModifiedAt)
	})
}
