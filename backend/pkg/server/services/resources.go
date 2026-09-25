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
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"

	"pentagi/pkg/graph/model"
	"pentagi/pkg/graph/subscriptions"
	"pentagi/pkg/resources"
	"pentagi/pkg/server/logger"
	"pentagi/pkg/server/models"
	"pentagi/pkg/server/response"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

// ---- request/response types ------------------------------------------------

// pendingResourceUpload holds state for a single file in the upload pipeline.
type pendingResourceUpload struct {
	name    string
	vPath   string
	tmpPath string
	hash    string
	size    int64
	newBlob bool
}

// ---- service ---------------------------------------------------------------

// ResourceService manages user-owned resource files.
// Physical file content is stored as {DATA_DIR}/resources/{md5hash}.blob;
// all metadata (virtual path, name, size, …) lives in user_resources (PostgreSQL).
type ResourceService struct {
	dataDir string
	db      *gorm.DB
	ss      subscriptions.SubscriptionsController
}

func NewResourceService(
	db *gorm.DB,
	dataDir string,
	ss subscriptions.SubscriptionsController,
) *ResourceService {
	return &ResourceService{
		dataDir: dataDir,
		db:      db,
		ss:      ss,
	}
}

// ---- GET /resources/ -------------------------------------------------------

// ListResources returns a list of user-owned resources.
// @Summary List user resources
// @Tags Resources
// @Produce json
// @Security BearerAuth
// @Param path query string false "virtual directory path; empty for root (may be combined with paths[])"
// @Param paths[] query []string false "additional virtual directory paths (repeatable)"
// @Param recursive query bool false "list recursively (default false)"
// @Success 200 {object} response.successResp{data=models.ResourceList}
// @Failure 400 {object} response.errorResp
// @Failure 403 {object} response.errorResp
// @Failure 500 {object} response.errorResp
// @Router /resources/ [get]
func (s *ResourceService) ListResources(c *gin.Context) {
	uid := c.GetUint64("uid")
	privs := c.GetStringSlice("prm")
	isAdmin := slices.Contains(privs, "resources.admin")

	if !isAdmin && !slices.Contains(privs, "resources.view") {
		response.Error(c, response.ErrNotPermitted, nil)
		return
	}

	recursive := c.Query("recursive") == "true" || c.Query("recursive") == "1"

	// Collect paths from both "path" and "paths[]" query params.
	rawPaths := c.QueryArray("paths[]")
	if singlePath := strings.TrimSpace(c.Query("path")); singlePath != "" {
		rawPaths = append(rawPaths, singlePath)
	}

	// When no paths are provided, list the root (backward-compat behaviour).
	if len(rawPaths) == 0 {
		items, err := s.queryResources(uid, isAdmin, "", recursive)
		if err != nil {
			logger.FromContext(c).WithError(err).Error("error listing resources")
			response.Error(c, response.ErrInternal, err)
			return
		}
		items = resources.DedupeByPathPreferringOwner(items, uid, resourceEntryOwner)
		response.Success(c, http.StatusOK, models.ResourceList{Items: items, Total: uint64(len(items))})
		return
	}

	dirPaths, err := collectAndSanitizeResourcePaths(rawPaths)
	if err != nil {
		logger.FromContext(c).WithError(err).Error("invalid path for list resources")
		response.Error(c, response.ErrResourcesInvalidRequest, err)
		return
	}

	// If all provided paths were whitespace-only, fall back to root listing.
	if len(dirPaths) == 0 {
		items, err := s.queryResources(uid, isAdmin, "", recursive)
		if err != nil {
			logger.FromContext(c).WithError(err).Error("error listing resources")
			response.Error(c, response.ErrInternal, err)
			return
		}
		items = resources.DedupeByPathPreferringOwner(items, uid, resourceEntryOwner)
		response.Success(c, http.StatusOK, models.ResourceList{Items: items, Total: uint64(len(items))})
		return
	}

	// Query each path and merge results, deduplicating by virtual path.
	merged := make([]models.ResourceEntry, 0)
	for _, dirPath := range dirPaths {
		items, err := s.queryResources(uid, isAdmin, dirPath, recursive)
		if err != nil {
			logger.FromContext(c).WithError(err).Error("error listing resources")
			response.Error(c, response.ErrInternal, err)
			return
		}
		merged = append(merged, items...)
	}
	allItems := resources.DedupeByPathPreferringOwner(merged, uid, resourceEntryOwner)
	seenPaths := make(map[string]struct{}, len(allItems))
	for _, item := range allItems {
		seenPaths[item.Path] = struct{}{}
	}

	// Ensure all ancestor directories of queried paths are present so that
	// clients can construct a complete tree without dangling nodes.
	// For example, querying "temp2/temp" must also return "temp2".
	ancestorCandidates := make(map[string]struct{})
	for _, dirPath := range dirPaths {
		parts := strings.Split(dirPath, "/")
		for i := 1; i < len(parts); i++ {
			ancestor := strings.Join(parts[:i], "/")
			if _, already := seenPaths[ancestor]; !already {
				ancestorCandidates[ancestor] = struct{}{}
			}
		}
	}
	if len(ancestorCandidates) > 0 {
		neededAncestors := make([]string, 0, len(ancestorCandidates))
		for p := range ancestorCandidates {
			neededAncestors = append(neededAncestors, p)
		}
		q := scopeToCaller(s.db.Model(&models.UserResource{}).
			Order("updated_at DESC, name ASC, id ASC").
			Where("path IN (?) AND is_dir = true", neededAncestors), uid, isAdmin)
		var ancestorRecs []models.UserResource
		if err := q.Find(&ancestorRecs).Error; err != nil {
			logger.FromContext(c).WithError(err).Error("error fetching ancestor directories for resources listing")
			response.Error(c, response.ErrInternal, err)
			return
		}
		for _, rec := range resources.DedupeByPathPreferringOwner(ancestorRecs, uid, resourceRecordOwner) {
			if _, seen := seenPaths[rec.Path]; !seen {
				seenPaths[rec.Path] = struct{}{}
				allItems = append(allItems, convertResource(rec))
			}
		}
	}

	sort.Slice(allItems, func(i, j int) bool {
		if allItems[i].Path < allItems[j].Path {
			return true
		} else if allItems[i].Path > allItems[j].Path {
			return false
		} else if allItems[i].Name < allItems[j].Name {
			return true
		} else if allItems[i].Name > allItems[j].Name {
			return false
		}
		return allItems[i].CreatedAt.Before(allItems[j].CreatedAt)
	})
	response.Success(c, http.StatusOK, models.ResourceList{Items: allItems, Total: uint64(len(allItems))})
}

// collectAndSanitizeResourcePaths sanitizes and deduplicates a slice of raw
// virtual resource paths. Whitespace-only entries are skipped; invalid paths
// return an error. Returns nil when all entries reduce to empty (caller decides
// the fallback behaviour).
func collectAndSanitizeResourcePaths(rawPaths []string) ([]string, error) {
	seen := make(map[string]struct{}, len(rawPaths))
	result := make([]string, 0, len(rawPaths))
	for _, p := range rawPaths {
		trimmed := strings.TrimSpace(p)
		if trimmed == "" {
			continue
		}
		sanitized, err := resources.SanitizeResourcePath(trimmed)
		if err != nil {
			return nil, err
		}
		if _, dup := seen[sanitized]; !dup {
			seen[sanitized] = struct{}{}
			result = append(result, sanitized)
		}
	}
	return result, nil
}

// ---- POST /resources/ ------------------------------------------------------

// UploadResources uploads one or more files into the user's resource storage.
// @Summary Upload files to resource storage
// @Tags Resources
// @Accept mpfd
// @Produce json
// @Security BearerAuth
// @Param dir query string false "target virtual directory path; empty or omitted means root"
// @Param files formData file true "files to upload (field name: files or file)"
// @Success 200 {object} response.successResp{data=models.ResourceList}
// @Failure 400 {object} response.errorResp
// @Failure 403 {object} response.errorResp
// @Failure 409 {object} response.errorResp
// @Failure 500 {object} response.errorResp
// @Router /resources/ [post]
func (s *ResourceService) UploadResources(c *gin.Context) {
	uid := c.GetUint64("uid")
	privs := c.GetStringSlice("prm")

	if !slices.Contains(privs, "resources.upload") && !slices.Contains(privs, "resources.admin") {
		response.Error(c, response.ErrNotPermitted, nil)
		return
	}

	dirRaw := strings.TrimSpace(c.Query("dir"))
	dirPath, err := resources.SanitizeResourceDir(dirRaw)
	if err != nil {
		logger.FromContext(c).WithError(err).Error("invalid upload target directory")
		response.Error(c, response.ErrResourcesInvalidRequest, err)
		return
	}

	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, resources.MaxUploadRequestSize)
	multipartForm, err := c.MultipartForm()
	if err != nil {
		logger.FromContext(c).WithError(err).Error("error reading multipart form for resources upload")
		response.Error(c, response.ErrResourcesInvalidRequest, err)
		return
	}

	fileHeaders := multipartForm.File["files"]
	if len(fileHeaders) == 0 {
		if fh, ferr := c.FormFile("file"); ferr == nil && fh != nil {
			fileHeaders = append(fileHeaders, fh)
		}
	}
	if len(fileHeaders) == 0 {
		response.Error(c, response.ErrResourcesInvalidRequest, errors.New("at least one file is required"))
		return
	}
	if len(fileHeaders) > resources.MaxUploadFiles {
		response.Error(c, response.ErrResourcesInvalidRequest,
			fmt.Errorf("too many files: %d exceeds the limit of %d", len(fileHeaders), resources.MaxUploadFiles))
		return
	}

	if err := resources.EnsureResourcesDir(s.dataDir); err != nil {
		logger.FromContext(c).WithError(err).Error("failed to ensure resources directory")
		response.Error(c, response.ErrInternal, err)
		return
	}
	blobsDir := resources.ResourcesDir(s.dataDir)

	var totalSize int64
	pendingList := make([]pendingResourceUpload, 0, len(fileHeaders))
	seenPaths := make(map[string]struct{}, len(fileHeaders))

	// Phase 1: validate, stream to tmp, compute MD5.
	for _, fh := range fileHeaders {
		name, err := resources.SanitizeResourceFileName(fh.Filename)
		if err != nil {
			cleanupResourceUploads(pendingList)
			response.Error(c, response.ErrResourcesInvalidData, err)
			return
		}

		if fh.Size > resources.MaxUploadFileSize {
			cleanupResourceUploads(pendingList)
			response.Error(c, response.ErrResourcesInvalidRequest,
				fmt.Errorf("file %q size %d bytes exceeds the limit of %d bytes", name, fh.Size, resources.MaxUploadFileSize))
			return
		}
		if fh.Size < 0 {
			cleanupResourceUploads(pendingList)
			response.Error(c, response.ErrResourcesInvalidRequest,
				fmt.Errorf("file %q has invalid size %d", name, fh.Size))
			return
		}
		totalSize += fh.Size
		if totalSize > resources.MaxUploadTotalSize {
			cleanupResourceUploads(pendingList)
			response.Error(c, response.ErrResourcesInvalidRequest,
				fmt.Errorf("total upload size exceeds the limit of %d bytes", resources.MaxUploadTotalSize))
			return
		}

		vPath := resources.FilePath(dirPath, name)
		if _, ok := seenPaths[vPath]; ok {
			cleanupResourceUploads(pendingList)
			response.Error(c, response.ErrResourcesAlreadyExists,
				fmt.Errorf("resource %q is duplicated in upload request", vPath))
			return
		}
		seenPaths[vPath] = struct{}{}

		exists, err := s.resourceExists(uid, vPath)
		if err != nil {
			cleanupResourceUploads(pendingList)
			logger.FromContext(c).WithError(err).Error("error checking resource existence")
			response.Error(c, response.ErrInternal, err)
			return
		}
		if exists {
			cleanupResourceUploads(pendingList)
			response.Error(c, response.ErrResourcesAlreadyExists,
				fmt.Errorf("resource %q already exists", vPath))
			return
		}

		src, openErr := fh.Open()
		if openErr != nil {
			cleanupResourceUploads(pendingList)
			response.Error(c, response.ErrInternal, fmt.Errorf("failed to open uploaded file: %w", openErr))
			return
		}
		tmpPath, hash, size, saveErr := resources.SaveToTemp(src, blobsDir)
		src.Close()
		if saveErr != nil {
			cleanupResourceUploads(pendingList)
			response.Error(c, response.ErrInternal, saveErr)
			return
		}

		pendingList = append(pendingList, pendingResourceUpload{
			name:    name,
			vPath:   vPath,
			tmpPath: tmpPath,
			hash:    hash,
			size:    size,
		})
	}

	// Phase 2: commit blobs (atomic rename; safe against duplicate hashes).
	for i := range pendingList {
		p := &pendingList[i]
		blobAlreadyExisted, err := resources.BlobExists(s.dataDir, p.hash)
		if err != nil {
			cleanupResourceUploads(pendingList)
			response.Error(c, response.ErrInternal, err)
			return
		}
		if err := resources.CommitBlob(s.dataDir, p.hash, p.tmpPath); err != nil {
			cleanupResourceUploads(pendingList)
			response.Error(c, response.ErrInternal, err)
			return
		}
		p.tmpPath = ""
		p.newBlob = !blobAlreadyExisted
	}

	// Phase 3: insert DB records in a transaction.
	tx := s.db.Begin()
	if tx.Error != nil {
		for _, p := range pendingList {
			if p.newBlob {
				s.deleteOrphanBlob(context.Background(), p.hash)
			}
		}
		logger.FromContext(c).WithError(tx.Error).Error("failed to begin transaction for upload")
		response.Error(c, response.ErrInternal, tx.Error)
		return
	}

	var saved []models.UserResource
	var createdDirs []models.UserResource
	var txErr error
	if dirPath != "" {
		createdDirs, _, _, txErr = ensureDestinationDirs(tx, uid, dirPath, false)
	}
	for _, p := range pendingList {
		if txErr != nil {
			break
		}
		rec := models.UserResource{
			UserID: uid,
			Hash:   p.hash,
			Name:   p.name,
			Path:   p.vPath,
			Size:   p.size,
			IsDir:  false,
		}
		if err := tx.Create(&rec).Error; err != nil {
			txErr = fmt.Errorf("failed to insert resource %q: %w", p.vPath, err)
			break
		}
		saved = append(saved, rec)
	}

	if txErr != nil {
		tx.Rollback()
		for _, p := range pendingList {
			if p.newBlob {
				s.deleteOrphanBlob(context.Background(), p.hash)
			}
		}
		if isUniqueViolation(txErr) {
			response.Error(c, response.ErrResourcesAlreadyExists, txErr)
		} else if errors.Is(txErr, errResourceConflict) {
			response.Error(c, response.ErrResourcesAlreadyExists, txErr)
		} else {
			logger.FromContext(c).WithError(txErr).Error("failed to persist uploaded resources")
			response.Error(c, response.ErrInternal, txErr)
		}
		return
	}
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		for _, p := range pendingList {
			if p.newBlob {
				s.deleteOrphanBlob(context.Background(), p.hash)
			}
		}
		logger.FromContext(c).WithError(err).Error("failed to commit upload transaction")
		response.Error(c, response.ErrInternal, err)
		return
	}

	entries := convertResources(saved)
	s.publishResourcesAdded(c.Request.Context(), convertResources(createdDirs))
	s.publishResourcesAdded(c.Request.Context(), entries)
	response.Success(c, http.StatusOK, models.ResourceList{Items: entries, Total: uint64(len(entries))})
}

// ---- POST /resources/mkdir -------------------------------------------------

// MkdirResource creates a virtual directory entry and every missing parent of
// it (idempotent). The reply names the leaf; the parents reach clients through
// the resource-added subscription.
// @Summary Create a virtual directory and any missing parents
// @Tags Resources
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body models.MkdirResourceRequest true "mkdir request"
// @Success 200 {object} response.successResp{data=models.ResourceEntry}
// @Failure 400 {object} response.errorResp
// @Failure 403 {object} response.errorResp
// @Failure 500 {object} response.errorResp
// @Router /resources/mkdir [post]
func (s *ResourceService) MkdirResource(c *gin.Context) {
	uid := c.GetUint64("uid")
	privs := c.GetStringSlice("prm")

	if !slices.Contains(privs, "resources.edit") && !slices.Contains(privs, "resources.admin") {
		response.Error(c, response.ErrNotPermitted, nil)
		return
	}

	var req models.MkdirResourceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, response.ErrResourcesInvalidRequest, err)
		return
	}

	dirPath, err := resources.SanitizeResourcePath(req.Path)
	if err != nil {
		response.Error(c, response.ErrResourcesInvalidRequest, err)
		return
	}

	created, err := s.createDirChain(uid, dirPath)
	if err != nil && !errors.Is(err, errResourceConflict) {
		// A racing insert of the same segment leaves this transaction aborted,
		// and Postgres then refuses every later statement in it — the recovery
		// read inside the chain walk included. A fresh attempt sees the rows.
		created, err = s.createDirChain(uid, dirPath)
	}
	if err != nil {
		if errors.Is(err, errResourceConflict) {
			response.Error(c, response.ErrResourcesAlreadyExists, err)
			return
		}
		logger.FromContext(c).WithError(err).Error("error creating resource directory")
		response.Error(c, response.ErrInternal, err)
		return
	}

	s.publishResourcesAdded(c.Request.Context(), convertResources(created))

	var leaf models.UserResource
	if err := s.db.Where("user_id = ? AND path = ?", uid, dirPath).First(&leaf).Error; err != nil {
		logger.FromContext(c).WithError(err).Error("error reading back created resource directory")
		response.Error(c, response.ErrInternal, err)
		return
	}

	response.Success(c, http.StatusOK, convertResource(leaf))
}

// createDirChain creates the segments of dirPath that have no row yet and
// returns them. Idempotent: an already-complete path creates nothing.
func (s *ResourceService) createDirChain(uid uint64, dirPath string) ([]models.UserResource, error) {
	tx := s.db.Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}

	created, _, _, err := ensureDestinationDirs(tx, uid, dirPath, false)
	if err != nil {
		tx.Rollback()

		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return created, nil
}

// ---- PUT /resources/move ---------------------------------------------------

// MoveResource moves or renames one or more resource files / directories.
//
// Single-source behaviour (exactly one unique source after dedup):
//
//	Destination is the exact target path, inheriting existing trailing-slash
//	and existing-directory semantics (unchanged from original behaviour).
//
// Multi-source behaviour (two or more unique sources after dedup):
//
//	Destination is always treated as a base directory. Each source lands at
//	destination/<basename>. All moves are executed in a single DB transaction
//	so the operation is atomic: either all succeed or none do.
//
// Response includes every Added (e.g. newly created parent directories) and
// every Updated (moved items) entry so the Apollo cache can be updated correctly.
//
// @Summary Move or rename resource(s) (files or directories)
// @Tags Resources
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body models.MoveResourceRequest true "move request"
// @Success 200 {object} response.successResp{data=models.ResourceList}
// @Failure 400 {object} response.errorResp
// @Failure 403 {object} response.errorResp
// @Failure 404 {object} response.errorResp
// @Failure 409 {object} response.errorResp
// @Failure 500 {object} response.errorResp
// @Router /resources/move [put]
func (s *ResourceService) MoveResource(c *gin.Context) {
	uid := c.GetUint64("uid")
	privs := c.GetStringSlice("prm")
	isAdmin := slices.Contains(privs, "resources.admin")

	if !slices.Contains(privs, "resources.edit") && !isAdmin {
		response.Error(c, response.ErrNotPermitted, nil)
		return
	}

	var req models.MoveResourceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, response.ErrResourcesInvalidRequest, err)
		return
	}

	// ── Merge + deduplicate source paths ─────────────────────────────────────
	rawSources := req.Sources
	if req.Source != "" {
		rawSources = append(rawSources, req.Source)
	}
	srcPaths, err := collectAndSanitizeResourcePaths(rawSources)
	if err != nil {
		response.Error(c, response.ErrResourcesInvalidRequest, fmt.Errorf("invalid source: %w", err))
		return
	}
	if len(srcPaths) == 0 {
		response.Error(c, response.ErrResourcesInvalidRequest,
			errors.New("at least one source path is required (use 'source' or 'sources')"))
		return
	}

	// ── Sanitize destination ──────────────────────────────────────────────────
	// Empty destination is allowed and means "move into the root directory".
	dstPath, err := resources.SanitizeResourceDir(req.Destination)
	if err != nil {
		response.Error(c, response.ErrResourcesInvalidRequest, fmt.Errorf("invalid destination: %w", err))
		return
	}

	// ── Route to single-source or multi-source path ───────────────────────────
	var result moveResourceResult
	if len(srcPaths) == 1 {
		srcPath := srcPaths[0]
		// When destination is root (""), the effective target is <root>/<src.Name>.
		// A source already living at root would land on itself — catch it early.
		dstIsDirHint := dstPath == "" || pathHasTrailingSeparator(req.Destination)
		if srcPath == dstPath || (dstPath == "" && !strings.Contains(srcPath, "/")) {
			response.Error(c, response.ErrResourcesInvalidRequest, errors.New("source and destination are the same"))
			return
		}
		src, found, findErr := rowAddressedByPath(s.db, uid, isAdmin, srcPath)
		if findErr != nil {
			logger.FromContext(c).WithError(findErr).Error("error finding source resource for move")
			response.Error(c, response.ErrInternal, findErr)
			return
		}
		if !found {
			response.Error(c, response.ErrResourcesNotFound, fmt.Errorf("resource %q not found", srcPath))
			return
		}
		result, err = s.moveResource(src.UserID, src, srcPath, dstPath, req.Force, dstIsDirHint)
	} else {
		owner, ownerErr := ownerOfAddressedPaths(s.db, uid, isAdmin, srcPaths)
		if ownerErr != nil {
			err = ownerErr
		} else {
			result, err = s.moveMultipleSources(owner, srcPaths, dstPath, req.Force)
		}
	}

	if err != nil {
		switch {
		case errors.Is(err, errResourceInvalid):
			response.Error(c, response.ErrResourcesInvalidRequest, err)
		case errors.Is(err, errResourceConflict):
			response.Error(c, response.ErrResourcesConflict, err)
		case errors.Is(err, errResourceNotFound):
			response.Error(c, response.ErrResourcesNotFound, err)
		default:
			logger.FromContext(c).WithError(err).Error("error moving resource(s)")
			response.Error(c, response.ErrInternal, err)
		}
		return
	}

	s.cleanupOrphanBlobs(c.Request.Context(), result.OrphanHashes)
	s.publishResourcesDeleted(c.Request.Context(), result.DeletedBefore)
	s.publishResourcesAdded(c.Request.Context(), result.Added)
	s.publishResourcesUpdated(c.Request.Context(), result.Updated)
	s.publishResourcesDeleted(c.Request.Context(), result.DeletedAfter)

	// Return Added + Updated: the frontend Apollo cache needs both newly created
	// parent directories (Added) and the moved items themselves (Updated) to
	// reflect the full post-move state.
	all := append(result.Added, result.Updated...)
	response.Success(c, http.StatusOK, models.ResourceList{Items: all, Total: uint64(len(all))})
}

// moveMultipleSources moves every source path to destination/<source.Name>
// inside a single DB transaction for full atomicity.
//
// Optimisations vs calling moveResource N times:
//   - One TX for all moves (atomic: all-or-nothing).
//   - Batch source lookup (one query regardless of N).
//   - Within-batch duplicate-basename detection before any DB writes.
func (s *ResourceService) moveMultipleSources(
	uid uint64,
	srcPaths []string,
	dstPath string,
	force bool,
) (moveResourceResult, error) {
	result := moveResourceResult{}

	tx := s.db.Begin()
	if tx.Error != nil {
		return result, tx.Error
	}

	// ── Batch-load all source records in one query ────────────────────────────
	srcs, err := findResourcesByPaths(tx, uid, srcPaths)
	if err != nil {
		tx.Rollback()
		return result, err
	}

	// Verify every source exists.
	if len(srcs) != len(srcPaths) {
		tx.Rollback()
		found := resourcesByPath(srcs)
		for _, sp := range srcPaths {
			if _, ok := found[sp]; !ok {
				return result, fmt.Errorf("%w: resource %q not found", errResourceNotFound, sp)
			}
		}
		return result, fmt.Errorf("%w: sources resolve to the same resource", errResourceInvalid)
	}

	// ── Within-batch target-basename conflict check ───────────────────────────
	// Two sources sharing a basename would both land at the same target path.
	basenameToSrc := make(map[string]string, len(srcs))
	for _, src := range srcs {
		if prev, conflict := basenameToSrc[src.Name]; conflict {
			tx.Rollback()
			return result, fmt.Errorf(
				"%w: sources %q and %q share the same base name %q",
				errResourceConflict, prev, src.Path, src.Name,
			)
		}
		basenameToSrc[src.Name] = src.Path
	}

	// Self-move guard: no source may be a directory that contains dstPath.
	for _, src := range srcs {
		if src.IsDir && resources.PathHasPrefix(dstPath, src.Path) {
			tx.Rollback()
			return result, fmt.Errorf(
				"%w: cannot move directory %q into itself", errResourceInvalid, src.Path,
			)
		}
		if resources.FilePath(dstPath, src.Name) == src.Path {
			tx.Rollback()
			return result, fmt.Errorf(
				"%w: source and destination are the same for %q", errResourceInvalid, src.Path,
			)
		}
	}

	// ── Ensure destination directory ──────────────────────────────────────────
	dest, destExists, err := findResourceByPath(tx, uid, dstPath)
	if err != nil {
		tx.Rollback()
		return result, err
	}
	if destExists && !dest.IsDir {
		tx.Rollback()
		return result, fmt.Errorf(
			"%w: destination %q is a file; cannot move multiple sources into it",
			errResourceConflict, dstPath,
		)
	}
	if !destExists {
		createdDirs, deletedDirs, orphanHashes, dirErr := ensureDestinationDirs(tx, uid, dstPath, force)
		if dirErr != nil {
			tx.Rollback()
			return result, dirErr
		}
		result.Added = append(result.Added, convertResources(createdDirs)...)
		result.DeletedBefore = append(result.DeletedBefore, convertResources(deletedDirs)...)
		result.OrphanHashes = append(result.OrphanHashes, orphanHashes...)
	}

	// ── Move each source into destination/<basename> ──────────────────────────
	// The inner move functions (moveFileResource / moveDirResource) call
	srcByPath := resourcesByPath(srcs)
	for _, sp := range srcPaths { // preserve original request order
		src, ok := srcByPath[sp]
		if !ok {
			continue // guarded by existence check above
		}
		targetPath := resources.FilePath(dstPath, src.Name)

		var moveResult moveResourceResult
		if !src.IsDir {
			// Pass dstIsDirHint=false: targetPath is the exact file destination.
			moveResult, err = s.moveFileResource(tx, uid, src, targetPath, force, false)
		} else {
			moveResult, err = s.moveDirResource(tx, uid, src.Path, targetPath, force)
		}
		if err != nil {
			tx.Rollback()
			return result, err
		}

		result.Added = append(result.Added, moveResult.Added...)
		result.Updated = append(result.Updated, moveResult.Updated...)
		result.DeletedBefore = append(result.DeletedBefore, moveResult.DeletedBefore...)
		result.DeletedAfter = append(result.DeletedAfter, moveResult.DeletedAfter...)
		result.OrphanHashes = append(result.OrphanHashes, moveResult.OrphanHashes...)
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return result, err
	}
	return result, nil
}

var errResourceConflict = errors.New("resource conflict")
var errResourceNotFound = errors.New("resource not found")
var errResourceInvalid = errors.New("invalid resource operation")

type moveResourceResult struct {
	Added         []models.ResourceEntry
	Updated       []models.ResourceEntry
	DeletedBefore []models.ResourceEntry
	DeletedAfter  []models.ResourceEntry
	OrphanHashes  []string
}

func (s *ResourceService) moveResource(
	uid uint64,
	src models.UserResource,
	srcPath, dstPath string,
	force bool,
	dstIsDirHint bool,
) (moveResourceResult, error) {
	tx := s.db.Begin()
	if tx.Error != nil {
		return moveResourceResult{}, tx.Error
	}

	if !src.IsDir {
		result, err := s.moveFileResource(tx, uid, src, dstPath, force, dstIsDirHint)
		if err != nil {
			tx.Rollback()
			return moveResourceResult{}, err
		}
		if err := tx.Commit().Error; err != nil {
			tx.Rollback()
			return moveResourceResult{}, err
		}
		return result, nil
	}

	// For directory sources: if the destination is an existing directory,
	// move the source INTO it (Unix mv semantics: "mv docs archive" where
	// archive exists → archive/docs), just like moveFileResource already does
	// for files.  When the destination is absent the source is renamed to it.
	dest, destExists, err := findResourceByPath(tx, uid, dstPath)
	if err != nil {
		tx.Rollback()
		return moveResourceResult{}, err
	}
	if destExists && dest.IsDir {
		dstPath = resources.FilePath(dstPath, src.Name)
	}

	result, err := s.moveDirResource(tx, uid, srcPath, dstPath, force)
	if err != nil {
		tx.Rollback()
		return moveResourceResult{}, err
	}
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return moveResourceResult{}, err
	}
	return result, nil
}

func (s *ResourceService) moveFileResource(
	tx *gorm.DB,
	uid uint64,
	src models.UserResource,
	dstPath string,
	force bool,
	dstIsDirHint bool,
) (moveResourceResult, error) {
	result := moveResourceResult{}
	targetPath := dstPath

	dest, exists, err := findResourceByPath(tx, uid, dstPath)
	if err != nil {
		return result, err
	}
	if exists && dest.IsDir {
		targetPath = resources.FilePath(dstPath, src.Name)
		dest, exists, err = findResourceByPath(tx, uid, targetPath)
		if err != nil {
			return result, err
		}
	} else if dstIsDirHint {
		createdDirs, deletedDirs, orphanHashes, err := ensureDestinationDirs(tx, uid, dstPath, force)
		if err != nil {
			return result, err
		}
		result.Added = append(result.Added, convertResources(createdDirs)...)
		result.DeletedBefore = append(result.DeletedBefore, convertResources(deletedDirs)...)
		result.OrphanHashes = append(result.OrphanHashes, orphanHashes...)
		targetPath = resources.FilePath(dstPath, src.Name)
		dest, exists, err = findResourceByPath(tx, uid, targetPath)
		if err != nil {
			return result, err
		}
	} else {
		createdDirs, deletedDirs, orphanHashes, err := ensureParentDirs(tx, uid, targetPath)
		if err != nil {
			return result, err
		}
		result.Added = append(result.Added, convertResources(createdDirs)...)
		result.DeletedBefore = append(result.DeletedBefore, convertResources(deletedDirs)...)
		result.OrphanHashes = append(result.OrphanHashes, orphanHashes...)
	}

	if exists {
		if dest.IsDir {
			return result, errResourceConflict
		}
		if !force {
			return result, errResourceConflict
		}
		if err := tx.Delete(&dest).Error; err != nil {
			return result, fmt.Errorf("failed to delete destination for overwrite: %w", err)
		}
		result.DeletedBefore = append(result.DeletedBefore, convertResource(dest))
		if dest.Hash != "" {
			result.OrphanHashes = append(result.OrphanHashes, dest.Hash)
		}
	}

	updated, err := updateMovedResource(tx, src, targetPath, time.Now())
	if err != nil {
		return result, err
	}
	result.Updated = append(result.Updated, convertResource(updated))
	return result, nil
}

func (s *ResourceService) moveDirResource(
	tx *gorm.DB,
	uid uint64,
	srcPath, dstPath string,
	force bool,
) (moveResourceResult, error) {
	result := moveResourceResult{}
	if resources.PathHasPrefix(dstPath, srcPath) {
		return result, fmt.Errorf("%w: cannot move directory into itself", errResourceInvalid)
	}
	// Moving a directory to root (dstPath=="") means placing it at root level
	if dstPath == "" {
		dstPath = path.Base(srcPath)
	}

	createdParents, deletedParents, orphanHashes, err := ensureParentDirs(tx, uid, dstPath)
	if err != nil {
		return result, err
	}
	result.Added = append(result.Added, convertResources(createdParents)...)
	result.DeletedBefore = append(result.DeletedBefore, convertResources(deletedParents)...)
	result.OrphanHashes = append(result.OrphanHashes, orphanHashes...)

	dest, destExists, err := findResourceByPath(tx, uid, dstPath)
	if err != nil {
		return result, err
	}
	if destExists && !dest.IsDir {
		if !force {
			return result, errResourceConflict
		}
		if err := tx.Delete(&dest).Error; err != nil {
			return result, fmt.Errorf("failed to delete destination for overwrite: %w", err)
		}
		result.DeletedBefore = append(result.DeletedBefore, convertResource(dest))
		if dest.Hash != "" {
			result.OrphanHashes = append(result.OrphanHashes, dest.Hash)
		}
		destExists = false
	}
	if destExists && !force {
		return result, errResourceConflict
	}

	entries, err := listResourceTree(tx, uid, srcPath)
	if err != nil {
		return result, err
	}

	now := time.Now()
	if !destExists {
		moved, err := moveDirResourceToAbsentDestination(tx, uid, entries, srcPath, dstPath, force, now)
		if err != nil {
			return result, err
		}
		result.Added = append(result.Added, moved.Added...)
		result.Updated = append(result.Updated, moved.Updated...)
		result.DeletedBefore = append(result.DeletedBefore, moved.DeletedBefore...)
		result.DeletedAfter = append(result.DeletedAfter, moved.DeletedAfter...)
		result.OrphanHashes = append(result.OrphanHashes, moved.OrphanHashes...)
		return result, nil
	}
	moved, err := moveDirResourceMerge(tx, uid, entries, srcPath, dstPath, now)
	if err != nil {
		return result, err
	}
	result.Added = append(result.Added, moved.Added...)
	result.Updated = append(result.Updated, moved.Updated...)
	result.DeletedBefore = append(result.DeletedBefore, moved.DeletedBefore...)
	result.DeletedAfter = append(result.DeletedAfter, moved.DeletedAfter...)
	result.OrphanHashes = append(result.OrphanHashes, moved.OrphanHashes...)
	return result, nil
}

func moveDirResourceToAbsentDestination(
	tx *gorm.DB,
	uid uint64,
	entries []models.UserResource,
	srcPath, dstPath string,
	force bool,
	now time.Time,
) (moveResourceResult, error) {
	result := moveResourceResult{}
	newPaths := make([]string, 0, len(entries))
	for _, entry := range entries {
		newPaths = append(newPaths, resources.ReplacePrefixPath(entry.Path, srcPath, dstPath))
	}

	existing, err := findResourcesByPaths(tx, uid, newPaths)
	if err != nil {
		return result, err
	}
	existingByPath := resourcesByPath(existing)
	if len(existingByPath) > 0 && !force {
		return result, errResourceConflict
	}

	for i, entry := range entries {
		newPath := newPaths[i]
		if dest, ok := existingByPath[newPath]; ok {
			if dest.IsDir != entry.IsDir {
				return result, errResourceConflict
			}
			if err := tx.Delete(&dest).Error; err != nil {
				return result, fmt.Errorf("failed to delete destination for overwrite: %w", err)
			}
			result.DeletedBefore = append(result.DeletedBefore, convertResource(dest))
			if dest.Hash != "" {
				result.OrphanHashes = append(result.OrphanHashes, dest.Hash)
			}
		}
		updated, err := updateMovedResource(tx, entry, newPath, now)
		if err != nil {
			return result, err
		}
		result.Updated = append(result.Updated, convertResource(updated))
	}

	return result, nil
}

func moveDirResourceMerge(
	tx *gorm.DB,
	uid uint64,
	entries []models.UserResource,
	srcPath, dstPath string,
	now time.Time,
) (moveResourceResult, error) {
	result := moveResourceResult{}
	mappedPaths := make([]string, 0, len(entries))

	for _, entry := range entries {
		if entry.Path == srcPath {
			continue
		}
		rel := strings.TrimPrefix(entry.Path, srcPath+"/")
		newPath := resources.FilePath(dstPath, rel)
		mappedPaths = append(mappedPaths, newPath)
	}

	existing, err := findResourcesByPaths(tx, uid, mappedPaths)
	if err != nil {
		return result, err
	}
	existingByPath := resourcesByPath(existing)

	var sourceRoot *models.UserResource
	var sourceDirsToDelete []models.UserResource
	for _, entry := range entries {
		if entry.Path == srcPath {
			entryCopy := entry
			sourceRoot = &entryCopy
			continue
		}

		rel := strings.TrimPrefix(entry.Path, srcPath+"/")
		newPath := resources.FilePath(dstPath, rel)
		if dest, ok := existingByPath[newPath]; ok {
			if dest.IsDir != entry.IsDir {
				return result, errResourceConflict
			}
			if entry.IsDir {
				sourceDirsToDelete = append(sourceDirsToDelete, entry)
				continue
			}
			if err := tx.Delete(&dest).Error; err != nil {
				return result, fmt.Errorf("failed to delete destination for overwrite: %w", err)
			}
			result.DeletedBefore = append(result.DeletedBefore, convertResource(dest))
			if dest.Hash != "" {
				result.OrphanHashes = append(result.OrphanHashes, dest.Hash)
			}
		}

		updated, err := updateMovedResource(tx, entry, newPath, now)
		if err != nil {
			return result, err
		}
		result.Updated = append(result.Updated, convertResource(updated))
	}

	for _, dir := range sourceDirsToDelete {
		if err := tx.Delete(&dir).Error; err != nil {
			return result, fmt.Errorf("failed to delete merged source directory %q: %w", dir.Path, err)
		}
		result.DeletedAfter = append(result.DeletedAfter, convertResource(dir))
	}
	if sourceRoot != nil {
		if err := tx.Delete(sourceRoot).Error; err != nil {
			return result, fmt.Errorf("failed to delete source directory %q: %w", sourceRoot.Path, err)
		}
		result.DeletedAfter = append(result.DeletedAfter, convertResource(*sourceRoot))
	}

	return result, nil
}

func listResourceTree(tx *gorm.DB, uid uint64, rootPath string) ([]models.UserResource, error) {
	escapedRoot := resources.EscapeLike(rootPath)
	var entries []models.UserResource
	if err := tx.Where(
		"user_id = ? AND (path = ? OR path LIKE ?)",
		uid, rootPath, escapedRoot+"/%",
	).Order("path ASC").Find(&entries).Error; err != nil {
		return nil, fmt.Errorf("failed to list source directory: %w", err)
	}
	return entries, nil
}

func findResourceByPath(tx *gorm.DB, uid uint64, resourcePath string) (models.UserResource, bool, error) {
	var rec models.UserResource
	err := tx.Where("user_id = ? AND path = ?", uid, resourcePath).First(&rec).Error
	if err == nil {
		return rec, true, nil
	}
	if gorm.IsRecordNotFoundError(err) {
		return models.UserResource{}, false, nil
	}
	return models.UserResource{}, false, err
}

func rowAddressedByPath(tx *gorm.DB, uid uint64, isAdmin bool, vPath string) (models.UserResource, bool, error) {
	rec, found, err := findResourceByPath(tx, uid, vPath)
	if err != nil || found || !isAdmin {
		return rec, found, err
	}

	err = tx.Where("path = ?", vPath).Order("updated_at DESC, name ASC, id ASC").First(&rec).Error
	if err == nil {
		return rec, true, nil
	}
	if gorm.IsRecordNotFoundError(err) {
		return models.UserResource{}, false, nil
	}

	return models.UserResource{}, false, err
}

func ownerOfAddressedPaths(tx *gorm.DB, uid uint64, isAdmin bool, vPaths []string) (uint64, error) {
	owner := uid
	for i, vPath := range vPaths {
		rec, found, err := rowAddressedByPath(tx, uid, isAdmin, vPath)
		if err != nil {
			return 0, err
		}
		if !found {
			return 0, fmt.Errorf("%w: resource %q not found", errResourceNotFound, vPath)
		}
		if i == 0 {
			owner = rec.UserID
			continue
		}
		if rec.UserID != owner {
			return 0, fmt.Errorf("%w: resources %q belong to different users", errResourceInvalid, vPath)
		}
	}

	return owner, nil
}

func findResourcesByPaths(tx *gorm.DB, uid uint64, paths []string) ([]models.UserResource, error) {
	if len(paths) == 0 {
		return nil, nil
	}

	var recs []models.UserResource
	if err := tx.Where("user_id = ? AND path IN (?)", uid, paths).Find(&recs).Error; err != nil {
		return nil, err
	}
	return recs, nil
}

func resourcesByPath(recs []models.UserResource) map[string]models.UserResource {
	byPath := make(map[string]models.UserResource, len(recs))
	for _, rec := range recs {
		byPath[rec.Path] = rec
	}
	return byPath
}

func updateMovedResource(tx *gorm.DB, rec models.UserResource, newPath string, now time.Time) (models.UserResource, error) {
	if err := tx.Model(&models.UserResource{}).
		Where("id = ?", rec.ID).
		Updates(map[string]interface{}{
			"path":       newPath,
			"name":       path.Base(newPath),
			"updated_at": now,
		}).Error; err != nil {
		return models.UserResource{}, fmt.Errorf("failed to move resource %q: %w", rec.Path, err)
	}
	rec.Path = newPath
	rec.Name = path.Base(newPath)
	rec.UpdatedAt = now
	return rec, nil
}

// ---- POST /resources/copy --------------------------------------------------

// CopyResource copies one or more resource files / directories to a new path.
//
// Single-source behaviour (exactly one unique source after dedup):
//
//	Destination is the exact target path, inheriting existing trailing-slash
//	and existing-directory semantics (unchanged from original behaviour).
//
// Multi-source behaviour (two or more unique sources after dedup):
//
//	Destination is always treated as a base directory. Each source lands at
//	destination/<basename>. All copies are executed in a single DB transaction
//	so the operation is atomic: either all succeed or none do.
//
// @Summary Copy resource(s) (files or directories)
// @Tags Resources
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body models.CopyResourceRequest true "copy request"
// @Success 200 {object} response.successResp{data=models.ResourceList}
// @Failure 400 {object} response.errorResp
// @Failure 403 {object} response.errorResp
// @Failure 404 {object} response.errorResp
// @Failure 409 {object} response.errorResp
// @Failure 500 {object} response.errorResp
// @Router /resources/copy [post]
func (s *ResourceService) CopyResource(c *gin.Context) {
	uid := c.GetUint64("uid")
	privs := c.GetStringSlice("prm")
	isAdmin := slices.Contains(privs, "resources.admin")

	if !slices.Contains(privs, "resources.edit") && !isAdmin {
		response.Error(c, response.ErrNotPermitted, nil)
		return
	}

	var req models.CopyResourceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, response.ErrResourcesInvalidRequest, err)
		return
	}

	// ── Merge + deduplicate source paths ─────────────────────────────────────
	rawSources := req.Sources
	if req.Source != "" {
		rawSources = append(rawSources, req.Source)
	}
	srcPaths, err := collectAndSanitizeResourcePaths(rawSources)
	if err != nil {
		response.Error(c, response.ErrResourcesInvalidRequest, fmt.Errorf("invalid source: %w", err))
		return
	}
	if len(srcPaths) == 0 {
		response.Error(c, response.ErrResourcesInvalidRequest,
			errors.New("at least one source path is required (use 'source' or 'sources')"))
		return
	}

	// ── Sanitize destination ──────────────────────────────────────────────────
	dstPath, err := resources.SanitizeResourcePath(req.Destination)
	if err != nil {
		response.Error(c, response.ErrResourcesInvalidRequest, fmt.Errorf("invalid destination: %w", err))
		return
	}

	// ── Route to single-source or multi-source path ───────────────────────────
	var result copyResourceResult
	if len(srcPaths) == 1 {
		srcPath := srcPaths[0]
		if srcPath == dstPath {
			response.Error(c, response.ErrResourcesInvalidRequest, errors.New("source and destination are the same"))
			return
		}
		src, found, findErr := rowAddressedByPath(s.db, uid, isAdmin, srcPath)
		if findErr != nil {
			logger.FromContext(c).WithError(findErr).Error("error finding source resource for copy")
			response.Error(c, response.ErrInternal, findErr)
			return
		}
		if !found {
			response.Error(c, response.ErrResourcesNotFound, fmt.Errorf("resource %q not found", srcPath))
			return
		}
		result, err = s.copyResource(src.UserID, src, srcPath, dstPath, req.Force, pathHasTrailingSeparator(req.Destination))
	} else {
		owner, ownerErr := ownerOfAddressedPaths(s.db, uid, isAdmin, srcPaths)
		if ownerErr != nil {
			err = ownerErr
		} else {
			result, err = s.copyMultipleSources(owner, srcPaths, dstPath, req.Force)
		}
	}

	if err != nil {
		switch {
		case errors.Is(err, errResourceInvalid):
			response.Error(c, response.ErrResourcesInvalidRequest, err)
		case errors.Is(err, errResourceConflict):
			response.Error(c, response.ErrResourcesConflict, err)
		case errors.Is(err, errResourceNotFound):
			response.Error(c, response.ErrResourcesNotFound, err)
		default:
			logger.FromContext(c).WithError(err).Error("error copying resource(s)")
			response.Error(c, response.ErrInternal, err)
		}
		return
	}

	s.cleanupOrphanBlobs(c.Request.Context(), result.OrphanHashes)
	s.publishResourcesDeleted(c.Request.Context(), result.Deleted)
	s.publishResourcesAdded(c.Request.Context(), result.Added)
	s.publishResourcesUpdated(c.Request.Context(), result.Updated)

	all := append(result.Added, result.Updated...)
	response.Success(c, http.StatusOK, models.ResourceList{Items: all, Total: uint64(len(all))})
}

// copyMultipleSources copies every source path to destination/<source.Name>
// inside a single DB transaction for full atomicity.
//
// Optimisations vs calling copyResource N times:
//   - One TX for all copies (atomic: all-or-nothing).
//   - Batch source lookup (one query regardless of N).
//   - Within-batch duplicate-basename detection before any DB writes.
func (s *ResourceService) copyMultipleSources(
	uid uint64,
	srcPaths []string,
	dstPath string,
	force bool,
) (copyResourceResult, error) {
	result := copyResourceResult{}

	tx := s.db.Begin()
	if tx.Error != nil {
		return result, tx.Error
	}

	// ── Batch-load all source records ─────────────────────────────────────────
	srcs, err := findResourcesByPaths(tx, uid, srcPaths)
	if err != nil {
		tx.Rollback()
		return result, err
	}

	// Verify every source exists.
	if len(srcs) != len(srcPaths) {
		tx.Rollback()
		found := resourcesByPath(srcs)
		for _, sp := range srcPaths {
			if _, ok := found[sp]; !ok {
				return result, fmt.Errorf("%w: resource %q not found", errResourceNotFound, sp)
			}
		}
		return result, fmt.Errorf("%w: sources resolve to the same resource", errResourceInvalid)
	}

	// ── Within-batch target-basename conflict check ───────────────────────────
	basenameToSrc := make(map[string]string, len(srcs))
	for _, src := range srcs {
		if prev, conflict := basenameToSrc[src.Name]; conflict {
			tx.Rollback()
			return result, fmt.Errorf(
				"%w: sources %q and %q share the same base name %q",
				errResourceConflict, prev, src.Path, src.Name,
			)
		}
		basenameToSrc[src.Name] = src.Path
	}

	// Self-copy guard: copying a directory into itself.
	for _, src := range srcs {
		if src.IsDir && resources.PathHasPrefix(dstPath, src.Path) {
			tx.Rollback()
			return result, fmt.Errorf(
				"%w: cannot copy directory %q into itself", errResourceInvalid, src.Path,
			)
		}
		if resources.FilePath(dstPath, src.Name) == src.Path {
			tx.Rollback()
			return result, fmt.Errorf(
				"%w: source and destination are the same for %q", errResourceInvalid, src.Path,
			)
		}
	}

	// ── Ensure destination directory ──────────────────────────────────────────
	dest, destExists, err := findResourceByPath(tx, uid, dstPath)
	if err != nil {
		tx.Rollback()
		return result, err
	}
	if destExists && !dest.IsDir {
		tx.Rollback()
		return result, fmt.Errorf(
			"%w: destination %q is a file; cannot copy multiple sources into it",
			errResourceConflict, dstPath,
		)
	}
	if !destExists {
		createdDirs, deletedDirs, orphanHashes, dirErr := ensureDestinationDirs(tx, uid, dstPath, force)
		if dirErr != nil {
			tx.Rollback()
			return result, dirErr
		}
		result.Added = append(result.Added, convertResources(createdDirs)...)
		result.Deleted = append(result.Deleted, convertResources(deletedDirs)...)
		result.OrphanHashes = append(result.OrphanHashes, orphanHashes...)
	}

	// ── Copy each source into destination/<basename> ──────────────────────────
	srcByPath := resourcesByPath(srcs)
	for _, sp := range srcPaths { // preserve original request order
		src, ok := srcByPath[sp]
		if !ok {
			continue
		}
		targetPath := resources.FilePath(dstPath, src.Name)

		var copyResult copyResourceResult
		if !src.IsDir {
			// Pass dstIsDirHint=false: targetPath is the exact file destination.
			copyResult, err = s.copyFileResource(tx, uid, src, targetPath, force, false)
		} else {
			copyResult, err = s.copyDirResource(tx, uid, src.Path, targetPath, force)
		}
		if err != nil {
			tx.Rollback()
			return result, err
		}

		result.Added = append(result.Added, copyResult.Added...)
		result.Updated = append(result.Updated, copyResult.Updated...)
		result.Deleted = append(result.Deleted, copyResult.Deleted...)
		result.OrphanHashes = append(result.OrphanHashes, copyResult.OrphanHashes...)
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return result, err
	}
	return result, nil
}

type copyResourceResult struct {
	Added        []models.ResourceEntry
	Updated      []models.ResourceEntry
	Deleted      []models.ResourceEntry
	OrphanHashes []string
}

func (s *ResourceService) copyResource(
	uid uint64,
	src models.UserResource,
	srcPath, dstPath string,
	force bool,
	dstIsDirHint bool,
) (copyResourceResult, error) {
	tx := s.db.Begin()
	if tx.Error != nil {
		return copyResourceResult{}, tx.Error
	}

	if !src.IsDir {
		result, err := s.copyFileResource(tx, uid, src, dstPath, force, dstIsDirHint)
		if err != nil {
			tx.Rollback()
			return copyResourceResult{}, err
		}
		if err := tx.Commit().Error; err != nil {
			tx.Rollback()
			return copyResourceResult{}, err
		}
		return result, nil
	}

	result, err := s.copyDirResource(tx, uid, srcPath, dstPath, force)
	if err != nil {
		tx.Rollback()
		return copyResourceResult{}, err
	}
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return copyResourceResult{}, err
	}
	return result, nil
}

func (s *ResourceService) copyFileResource(
	tx *gorm.DB,
	uid uint64,
	src models.UserResource,
	dstPath string,
	force bool,
	dstIsDirHint bool,
) (copyResourceResult, error) {
	result := copyResourceResult{}
	targetPath := dstPath

	dest, exists, err := findResourceByPath(tx, uid, dstPath)
	if err != nil {
		return result, err
	}
	if exists && dest.IsDir {
		targetPath = resources.FilePath(dstPath, src.Name)
		dest, exists, err = findResourceByPath(tx, uid, targetPath)
		if err != nil {
			return result, err
		}
	} else if dstIsDirHint {
		createdDirs, deletedDirs, orphanHashes, err := ensureDestinationDirs(tx, uid, dstPath, force)
		if err != nil {
			return result, err
		}
		result.Added = append(result.Added, convertResources(createdDirs)...)
		result.Deleted = append(result.Deleted, convertResources(deletedDirs)...)
		result.OrphanHashes = append(result.OrphanHashes, orphanHashes...)
		targetPath = resources.FilePath(dstPath, src.Name)
		dest, exists, err = findResourceByPath(tx, uid, targetPath)
		if err != nil {
			return result, err
		}
	} else {
		createdDirs, deletedDirs, orphanHashes, err := ensureParentDirs(tx, uid, targetPath)
		if err != nil {
			return result, err
		}
		result.Added = append(result.Added, convertResources(createdDirs)...)
		result.Deleted = append(result.Deleted, convertResources(deletedDirs)...)
		result.OrphanHashes = append(result.OrphanHashes, orphanHashes...)
	}

	if exists {
		if dest.IsDir {
			return result, errResourceConflict
		}
		if !force {
			return result, errResourceConflict
		}
		if err := tx.Delete(&dest).Error; err != nil {
			return result, fmt.Errorf("failed to delete destination for overwrite: %w", err)
		}
		result.Deleted = append(result.Deleted, convertResource(dest))
		if dest.Hash != "" {
			result.OrphanHashes = append(result.OrphanHashes, dest.Hash)
		}
	}

	newRec := models.UserResource{
		UserID: uid,
		Hash:   src.Hash,
		Name:   path.Base(targetPath),
		Path:   targetPath,
		Size:   src.Size,
		IsDir:  false,
	}
	if err := tx.Create(&newRec).Error; err != nil {
		return result, fmt.Errorf("failed to copy resource: %w", err)
	}

	entry := convertResource(newRec)
	if exists {
		result.Updated = append(result.Updated, entry)
	} else {
		result.Added = append(result.Added, entry)
	}
	return result, nil
}

func (s *ResourceService) copyDirResource(
	tx *gorm.DB,
	uid uint64,
	srcPath, dstPath string,
	force bool,
) (copyResourceResult, error) {
	result := copyResourceResult{}
	if resources.PathHasPrefix(dstPath, srcPath) {
		return result, fmt.Errorf("%w: cannot copy directory into itself", errResourceInvalid)
	}

	createdRoot, deletedRoot, orphanHashes, err := ensureDestinationDirs(tx, uid, dstPath, force)
	if err != nil {
		return result, err
	}
	result.Added = append(result.Added, convertResources(createdRoot)...)
	result.Deleted = append(result.Deleted, convertResources(deletedRoot)...)
	result.OrphanHashes = append(result.OrphanHashes, orphanHashes...)

	dest, destExists, err := findResourceByPath(tx, uid, dstPath)
	if err != nil {
		return result, err
	}
	if destExists && !dest.IsDir {
		return result, errResourceConflict
	}
	if destExists && !force && len(createdRoot) == 0 {
		return result, errResourceConflict
	}

	srcEntries, err := listResourceTree(tx, uid, srcPath)
	if err != nil {
		return result, err
	}

	newPaths := make([]string, 0, len(srcEntries))
	sourceByNewPath := make(map[string]models.UserResource, len(srcEntries))
	newPathBySourcePath := make(map[string]string, len(srcEntries))
	for _, e := range srcEntries {
		newPath := resources.ReplacePrefixPath(e.Path, srcPath, dstPath)
		if destExists {
			if e.Path == srcPath {
				continue
			}
			rel := strings.TrimPrefix(e.Path, srcPath+"/")
			newPath = resources.FilePath(dstPath, rel)
		}
		newPaths = append(newPaths, newPath)
		sourceByNewPath[newPath] = e
		newPathBySourcePath[e.Path] = newPath
	}

	existing, err := findResourcesByPaths(tx, uid, newPaths)
	if err != nil {
		return result, fmt.Errorf("failed to check destination conflicts: %w", err)
	}
	if len(existing) > 0 && !force {
		return result, errResourceConflict
	}

	existingSet := make(map[string]models.UserResource, len(existing))
	for _, e := range existing {
		srcEntry := sourceByNewPath[e.Path]
		if e.IsDir != srcEntry.IsDir {
			return result, errResourceConflict
		}
		existingSet[e.Path] = e
	}

	for _, e := range srcEntries {
		if destExists && e.Path == srcPath {
			continue
		}
		newPath := newPathBySourcePath[e.Path]
		if dest, wasExisting := existingSet[newPath]; wasExisting {
			if dest.IsDir {
				result.Updated = append(result.Updated, convertResource(dest))
				continue
			}
			if err := tx.Delete(&dest).Error; err != nil {
				return result, fmt.Errorf("failed to delete destination for overwrite: %w", err)
			}
			result.Deleted = append(result.Deleted, convertResource(dest))
			if dest.Hash != "" {
				result.OrphanHashes = append(result.OrphanHashes, dest.Hash)
			}
		}

		newRec := models.UserResource{
			UserID: uid,
			Hash:   e.Hash,
			Name:   path.Base(newPath),
			Path:   newPath,
			Size:   e.Size,
			IsDir:  e.IsDir,
		}
		if err := tx.Create(&newRec).Error; err != nil {
			return result, fmt.Errorf("failed to copy resource %q: %w", e.Path, err)
		}
		entry := convertResource(newRec)
		if _, wasExisting := existingSet[newPath]; wasExisting {
			result.Updated = append(result.Updated, entry)
		} else {
			result.Added = append(result.Added, entry)
		}
	}

	return result, nil
}

// ---- DELETE /resources/ ----------------------------------------------------

// DeleteResource deletes one or more files or directories (recursively) by virtual path.
// @Summary Delete a resource (file or directory)
// @Tags Resources
// @Produce json
// @Security BearerAuth
// @Param path query string false "virtual path to delete (may be combined with paths[])"
// @Param paths[] query []string false "additional virtual paths to delete (repeatable)"
// @Success 200 {object} response.successResp{data=models.ResourceList}
// @Failure 400 {object} response.errorResp
// @Failure 403 {object} response.errorResp
// @Failure 404 {object} response.errorResp
// @Failure 500 {object} response.errorResp
// @Router /resources/ [delete]
func (s *ResourceService) DeleteResource(c *gin.Context) {
	uid := c.GetUint64("uid")
	privs := c.GetStringSlice("prm")
	isAdmin := slices.Contains(privs, "resources.admin")

	if !slices.Contains(privs, "resources.delete") && !isAdmin {
		response.Error(c, response.ErrNotPermitted, nil)
		return
	}

	// Collect paths from both "path" and "paths[]" query params.
	rawPaths := c.QueryArray("paths[]")
	if singlePath := strings.TrimSpace(c.Query("path")); singlePath != "" {
		rawPaths = append(rawPaths, singlePath)
	}
	if len(rawPaths) == 0 {
		response.Error(c, response.ErrResourcesInvalidRequest,
			errors.New("at least one path is required (use 'path' or 'paths[]' query parameters)"))
		return
	}

	targetPaths, err := collectAndSanitizeResourcePaths(rawPaths)
	if err != nil {
		response.Error(c, response.ErrResourcesInvalidRequest, err)
		return
	}
	if len(targetPaths) == 0 {
		response.Error(c, response.ErrResourcesInvalidRequest,
			errors.New("at least one valid path is required"))
		return
	}

	// Phase 1: validate every path and collect all records to delete.
	// Validation is fail-fast so that no rows are removed if any path is missing.
	hashSet := make(map[string]struct{})
	idSet := make(map[uint64]struct{})
	var toDelete []models.UserResource

	for _, targetPath := range targetPaths {
		root, found, err := rowAddressedByPath(s.db, uid, isAdmin, targetPath)
		if err != nil {
			logger.FromContext(c).WithError(err).Error("error finding resource for delete")
			response.Error(c, response.ErrInternal, err)
			return
		}
		if !found {
			response.Error(c, response.ErrResourcesNotFound, fmt.Errorf("resource %q not found", targetPath))
			return
		}

		escapedTarget := resources.EscapeLike(targetPath)
		var descendants []models.UserResource
		if err := s.db.Where(
			"user_id = ? AND (path = ? OR path LIKE ?)",
			root.UserID, targetPath, escapedTarget+"/%",
		).Find(&descendants).Error; err != nil {
			logger.FromContext(c).WithError(err).Error("error listing resources for delete")
			response.Error(c, response.ErrInternal, err)
			return
		}

		for _, rec := range descendants {
			if _, dup := idSet[rec.ID]; !dup {
				idSet[rec.ID] = struct{}{}
				toDelete = append(toDelete, rec)
				if !rec.IsDir && rec.Hash != "" {
					hashSet[rec.Hash] = struct{}{}
				}
			}
		}
	}

	if len(toDelete) == 0 {
		response.Success(c, http.StatusOK, models.ResourceList{})
		return
	}

	ids := make([]uint64, 0, len(toDelete))
	deleted := make([]models.ResourceEntry, 0, len(toDelete))
	for _, rec := range toDelete {
		ids = append(ids, rec.ID)
		deleted = append(deleted, convertResource(rec))
	}

	if err := s.db.Where("id IN (?)", ids).Delete(&models.UserResource{}).Error; err != nil {
		logger.FromContext(c).WithError(err).Error("error deleting resources")
		response.Error(c, response.ErrInternal, err)
		return
	}

	if len(hashSet) > 0 {
		hashes := make([]string, 0, len(hashSet))
		for h := range hashSet {
			hashes = append(hashes, h)
		}
		s.cleanupOrphanBlobs(c.Request.Context(), hashes)
	}

	sort.Slice(deleted, func(i, j int) bool {
		if deleted[i].Path < deleted[j].Path {
			return true
		} else if deleted[i].Path > deleted[j].Path {
			return false
		} else if deleted[i].Name < deleted[j].Name {
			return true
		} else if deleted[i].Name > deleted[j].Name {
			return false
		}
		return deleted[i].CreatedAt.Before(deleted[j].CreatedAt)
	})

	s.publishResourcesDeleted(c.Request.Context(), deleted)
	response.Success(c, http.StatusOK, models.ResourceList{Items: deleted, Total: uint64(len(deleted))})
}

// ---- GET /resources/download -----------------------------------------------

// DownloadResource downloads one or more resources.
// Single regular file: served as a direct attachment with an explicit
// Content-Length header so that Swagger UI and browsers offer a download link.
// Single directory or multiple paths (any mix): packaged into a ZIP archive.
// @Summary Download a resource or resources
// @Tags Resources
// @Produce application/octet-stream,application/zip,json
// @Security BearerAuth
// @Param path query string false "virtual path to download (may be combined with paths[])"
// @Param paths[] query []string false "additional virtual paths to download (repeatable)"
// @Param id query integer false "resource id to download (may be combined with ids[])"
// @Param ids[] query []integer false "additional resource ids to download (repeatable)"
// @Success 200 {file} binary "file content, or ZIP archive for directories / multiple paths"
// @Failure 400 {object} response.errorResp "invalid resource request data"
// @Failure 403 {object} response.errorResp "downloading resource not permitted"
// @Failure 404 {object} response.errorResp "resource not found"
// @Failure 500 {object} response.errorResp "internal error on downloading resource"
// @Router /resources/download [get]
func (s *ResourceService) DownloadResource(c *gin.Context) {
	uid := c.GetUint64("uid")
	privs := c.GetStringSlice("prm")
	isAdmin := slices.Contains(privs, "resources.admin")

	if !isAdmin && !slices.Contains(privs, "resources.download") {
		response.Error(c, response.ErrNotPermitted, nil)
		return
	}

	// Collect paths from both "path" and "paths[]" query params, then deduplicate.
	rawPaths := c.QueryArray("paths[]")
	if singlePath := strings.TrimSpace(c.Query("path")); singlePath != "" {
		rawPaths = append(rawPaths, singlePath)
	}

	rawIDs := c.QueryArray("ids[]")
	if singleID := strings.TrimSpace(c.Query("id")); singleID != "" {
		rawIDs = append(rawIDs, singleID)
	}

	if len(rawPaths) == 0 && len(rawIDs) == 0 {
		response.Error(c, response.ErrResourcesInvalidRequest,
			errors.New("at least one path or id is required "+
				"(use 'path', 'paths[]', 'id' or 'ids[]' query parameters)"))
		return
	}

	targetIDs, err := collectResourceIDs(rawIDs)
	if err != nil {
		response.Error(c, response.ErrResourcesInvalidRequest, err)
		return
	}

	targetPaths, err := collectAndSanitizeResourcePaths(rawPaths)
	if err != nil {
		response.Error(c, response.ErrResourcesInvalidRequest, err)
		return
	}
	if len(targetPaths) == 0 && len(targetIDs) == 0 {
		response.Error(c, response.ErrResourcesInvalidRequest,
			errors.New("at least one valid path or id is required"))
		return
	}

	// Load each requested resource from the DB; fail-fast on missing.
	type resolvedEntry struct {
		rec    models.UserResource
		isByID bool
	}
	entries := make([]resolvedEntry, 0, len(targetPaths))
	seenTargets := make(map[string]struct{}, len(targetPaths))
	for _, targetPath := range targetPaths {
		if _, dup := seenTargets[targetPath]; dup {
			continue
		}
		seenTargets[targetPath] = struct{}{}

		var rec models.UserResource
		err := s.db.Where("user_id = ? AND path = ?", uid, targetPath).First(&rec).Error
		if isAdmin && gorm.IsRecordNotFoundError(err) {
			err = s.db.Where("path = ?", targetPath).Order("updated_at DESC, name ASC, id ASC").First(&rec).Error
		}
		if err != nil {
			if gorm.IsRecordNotFoundError(err) {
				response.Error(c, response.ErrResourcesNotFound, fmt.Errorf("resource %q not found", targetPath))
				return
			}
			logger.FromContext(c).WithError(err).Error("error finding resource for download")
			response.Error(c, response.ErrInternal, err)
			return
		}
		entries = append(entries, resolvedEntry{rec: rec})
	}

	for _, id := range targetIDs {
		var rec models.UserResource
		err := scopeToCaller(s.db.Where("id = ?", id), uid, isAdmin).First(&rec).Error
		if err != nil {
			if gorm.IsRecordNotFoundError(err) {
				response.Error(c, response.ErrResourcesNotFound, fmt.Errorf("resource %d not found", id))
				return
			}
			logger.FromContext(c).WithError(err).Error("error finding resource by id for download")
			response.Error(c, response.ErrInternal, err)
			return
		}
		entries = append(entries, resolvedEntry{rec: rec, isByID: true})
	}

	// Single regular file → serve as a direct attachment with explicit Content-Length.
	if len(entries) == 1 && !entries[0].rec.IsDir {
		e := entries[0]
		blobPath := resources.BlobPath(s.dataDir, e.rec.Hash)
		f, err := os.Open(blobPath)
		if err != nil {
			if os.IsNotExist(err) {
				response.Error(c, response.ErrResourcesNotFound,
					fmt.Errorf("blob for resource %q not found on disk", e.rec.Path))
				return
			}
			logger.FromContext(c).WithError(err).Error("error opening resource blob for download")
			response.Error(c, response.ErrInternal, err)
			return
		}
		defer f.Close()

		info, err := f.Stat()
		if err != nil {
			logger.FromContext(c).WithError(err).Error("error stating resource blob for download")
			response.Error(c, response.ErrInternal, err)
			return
		}
		c.DataFromReader(http.StatusOK, info.Size(), "application/octet-stream", f,
			map[string]string{
				"Content-Disposition": mime.FormatMediaType("attachment", map[string]string{
					"filename": e.rec.Name,
				}),
			})
		return
	}

	// Single directory → ZIP with paths relative to that directory (backward-compat).
	if len(entries) == 1 && entries[0].rec.IsDir {
		e := entries[0]
		escapedTarget := resources.EscapeLike(e.rec.Path)
		var fileRecs []models.UserResource
		if err := s.db.Where(
			"user_id = ? AND path LIKE ? AND is_dir = false",
			e.rec.UserID, escapedTarget+"/%",
		).Find(&fileRecs).Error; err != nil {
			logger.FromContext(c).WithError(err).Error("error listing resources for zip download")
			response.Error(c, response.ErrInternal, err)
			return
		}

		zipEntries := make([]resources.ZipEntry, 0, len(fileRecs))
		for _, fr := range fileRecs {
			relPath := strings.TrimPrefix(fr.Path, e.rec.Path+"/")
			if relPath == "" {
				relPath = fr.Name
			}
			zipEntries = append(zipEntries, resources.ZipEntry{
				BlobPath: resources.BlobPath(s.dataDir, fr.Hash),
				ZipPath:  relPath,
			})
		}

		dirName := path.Base(e.rec.Path)
		if err := streamZipArchive(c, dirName+".zip", func(w io.Writer) error {
			return resources.ZipResources(w, zipEntries)
		}); err != nil {
			logger.FromContext(c).WithError(err).Error("error streaming zip archive for download")
		}
		return
	}

	// Multiple paths (any mix of files and directories) → ZIP using the full
	// virtual path as each entry name so the caller sees the complete path context.
	var byPath, byID []models.UserResource
	for _, e := range entries {
		recs := []models.UserResource{e.rec}
		if e.rec.IsDir {
			// Directory: collect all file descendants using their full virtual paths.
			escapedTarget := resources.EscapeLike(e.rec.Path)
			recs = nil
			if err := s.db.Where(
				"user_id = ? AND path LIKE ? AND is_dir = false",
				e.rec.UserID, escapedTarget+"/%",
			).Find(&recs).Error; err != nil {
				logger.FromContext(c).WithError(err).Error("error listing resources for zip download")
				response.Error(c, response.ErrInternal, err)
				return
			}
		}
		if e.isByID {
			byID = append(byID, recs...)
		} else {
			byPath = append(byPath, recs...)
		}
	}

	picked := resources.DedupeByPathPreferringOwner(byPath, uid, resourceRecordOwner)
	holders := make(map[string]uint64, len(picked)+len(byID))
	for _, rec := range picked {
		holders[rec.Path] = rec.ID
	}
	for _, rec := range byID {
		if holder, isHeld := holders[rec.Path]; isHeld {
			if holder == rec.ID {
				continue
			}
			response.Error(c, response.ErrResourcesSharedPath,
				fmt.Errorf("resources %d and %d share the path %q", holder, rec.ID, rec.Path))
			return
		}
		holders[rec.Path] = rec.ID
		picked = append(picked, rec)
	}
	sort.Slice(picked, func(i, j int) bool { return picked[i].Path < picked[j].Path })

	zipEntries := make([]resources.ZipEntry, 0, len(picked))
	for _, rec := range picked {
		zipEntries = append(zipEntries, resources.ZipEntry{
			BlobPath: resources.BlobPath(s.dataDir, rec.Hash),
			ZipPath:  rec.Path,
		})
	}

	if err := streamZipArchive(c, "download.zip", func(w io.Writer) error {
		return resources.ZipResources(w, zipEntries)
	}); err != nil {
		logger.FromContext(c).WithError(err).Error("error streaming zip archive for download")
	}
}

// zipStreamWriter defers committing the ZIP response status and headers until the
// first byte is written. That lets streamZipArchive turn a build failure that
// happens before any output into a normal structured error response, while a
// mid-stream failure (after the 200 status is already on the wire) is aborted.
type zipStreamWriter struct {
	c        *gin.Context
	filename string
	started  bool
}

func (zw *zipStreamWriter) Write(p []byte) (int, error) {
	if !zw.started {
		zw.started = true
		zw.c.Header("Content-Type", "application/zip")
		zw.c.Header("Content-Disposition", mime.FormatMediaType("attachment", map[string]string{
			"filename": zw.filename,
		}))
		zw.c.Status(http.StatusOK)
	}
	return zw.c.Writer.Write(p)
}

// streamZipArchive streams a ZIP archive to the client as build writes it, instead
// of buffering the entire archive in memory before sending it. Memory stays
// proportional to a single file copy buffer rather than the full archive size.
//
// Response headers and the 200 status are committed lazily, on the first byte
// build writes (see zipStreamWriter). If build fails before writing anything
// (e.g. a blob that cannot be opened), nothing has been committed and a normal
// structured error response is returned. If it fails mid-stream the 200 status is
// already on the wire, so the partial response is aborted; either way the error is
// returned for the caller to log with its own context.
func streamZipArchive(c *gin.Context, filename string, build func(w io.Writer) error) error {
	if err := build(&zipStreamWriter{c: c, filename: filename}); err != nil {
		if c.Writer.Written() {
			c.Abort()
		} else {
			response.Error(c, response.ErrInternal, err)
		}
		return err
	}
	return nil
}

// ---- helper methods --------------------------------------------------------

func scopeToCaller(q *gorm.DB, uid uint64, isAdmin bool) *gorm.DB {
	if isAdmin {
		return q
	}
	return q.Where("user_id = ?", uid)
}

func (s *ResourceService) queryResources(
	uid uint64,
	isAdmin bool,
	dirPath string,
	recursive bool,
) ([]models.ResourceEntry, error) {
	q := scopeToCaller(s.db.Model(&models.UserResource{}).Order("updated_at DESC, name ASC, id ASC"), uid, isAdmin)

	if dirPath != "" {
		escaped := resources.EscapeLike(dirPath)
		if recursive {
			q = q.Where("path = ? OR path LIKE ?", dirPath, escaped+"/%")
		} else {
			q = q.Where(
				"(path = ? AND is_dir = true) OR (path LIKE ? AND path NOT LIKE ?)",
				dirPath,
				escaped+"/%",
				escaped+"/%/%",
			)
		}
	} else if !recursive {
		q = q.Where("path NOT LIKE ?", "%/%")
	}

	var recs []models.UserResource
	if err := q.Find(&recs).Error; err != nil {
		return nil, fmt.Errorf("failed to list resources: %w", err)
	}
	return convertResources(recs), nil
}

func (s *ResourceService) resourceExists(uid uint64, vPath string) (bool, error) {
	var count int64
	err := s.db.Model(&models.UserResource{}).
		Where("user_id = ? AND path = ?", uid, vPath).
		Count(&count).Error
	return count > 0, err
}

func ensureDestinationDirs(tx *gorm.DB, uid uint64, dstPath string, force bool) (
	[]models.UserResource,
	[]models.UserResource,
	[]string,
	error,
) {
	return ensureDirChain(tx, uid, dstPath, force)
}

func ensureParentDirs(tx *gorm.DB, uid uint64, targetPath string) (
	[]models.UserResource,
	[]models.UserResource,
	[]string,
	error,
) {
	return ensureDirChain(tx, uid, resources.ParentDir(targetPath), false)
}

func ensureDirChain(tx *gorm.DB, uid uint64, dirPath string, forceReplacesLeaf bool) (
	[]models.UserResource,
	[]models.UserResource,
	[]string,
	error,
) {
	if dirPath == "" {
		return nil, nil, nil, nil
	}

	parts := strings.Split(dirPath, "/")
	created := make([]models.UserResource, 0, len(parts))
	deleted := make([]models.UserResource, 0, len(parts))
	orphanHashes := make([]string, 0, len(parts))
	current := ""
	for i, part := range parts {
		current = resources.FilePath(current, part)

		existing, exists, err := findResourceByPath(tx, uid, current)
		if err != nil {
			return nil, nil, nil, err
		}
		if exists {
			if !existing.IsDir {
				if !forceReplacesLeaf || i != len(parts)-1 {
					return nil, nil, nil, fmt.Errorf("%w: resource %q already exists and is not a directory", errResourceConflict, current)
				}
				if err := tx.Delete(&existing).Error; err != nil {
					return nil, nil, nil, fmt.Errorf("failed to delete file blocking directory %q: %w", current, err)
				}
				deleted = append(deleted, existing)
				if existing.Hash != "" {
					orphanHashes = append(orphanHashes, existing.Hash)
				}
			} else {
				continue
			}
		}

		rec := models.UserResource{
			UserID: uid,
			Hash:   "",
			Name:   path.Base(current),
			Path:   current,
			Size:   0,
			IsDir:  true,
		}
		if err := tx.Create(&rec).Error; err != nil {
			if isUniqueViolation(err) {
				refetched, ok, refetchErr := findResourceByPath(tx, uid, current)
				if refetchErr != nil {
					return nil, nil, nil, refetchErr
				}
				if ok && refetched.IsDir {
					continue
				}
			}
			return nil, nil, nil, fmt.Errorf("failed to create resource directory %q: %w", current, err)
		}
		created = append(created, rec)
	}

	return created, deleted, orphanHashes, nil
}

// deleteOrphanBlob removes the .blob for hash if no DB row references it.
func (s *ResourceService) deleteOrphanBlob(_ context.Context, hash string) {
	if hash == "" {
		return
	}
	var count int64
	if err := s.db.Model(&models.UserResource{}).
		Where("hash = ?", hash).
		Count(&count).Error; err != nil || count > 0 {
		return
	}
	_ = resources.DeleteBlob(s.dataDir, hash)
}

// cleanupOrphanBlobs removes .blob files whose hashes are no longer referenced.
func (s *ResourceService) cleanupOrphanBlobs(_ context.Context, hashes []string) {
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
	for _, h := range hashes {
		if !refSet[h] {
			_ = resources.DeleteBlob(s.dataDir, h)
		}
	}
}

// cleanupResourceUploads removes tmp files left after a failed upload.
func cleanupResourceUploads(pending []pendingResourceUpload) {
	for _, p := range pending {
		if p.tmpPath != "" {
			os.Remove(p.tmpPath)
		}
	}
}

func pathHasTrailingSeparator(p string) bool {
	trimmed := strings.TrimSpace(p)
	return strings.HasSuffix(trimmed, "/") || strings.HasSuffix(trimmed, "\\")
}

// ---- conversion helpers ----------------------------------------------------

func resourceEntryOwner(e models.ResourceEntry) (string, uint64) { return e.Path, e.UserID }

func resourceRecordOwner(r models.UserResource) (string, uint64) { return r.Path, r.UserID }

func collectResourceIDs(raw []string) ([]uint64, error) {
	ids := make([]uint64, 0, len(raw))
	seen := make(map[uint64]struct{}, len(raw))
	for _, item := range raw {
		trimmed := strings.TrimSpace(item)
		if trimmed == "" {
			continue
		}
		id, err := strconv.ParseUint(trimmed, 10, 64)
		if err != nil || id == 0 {
			return nil, fmt.Errorf("invalid resource id %q", item)
		}
		if _, dup := seen[id]; dup {
			continue
		}
		seen[id] = struct{}{}
		ids = append(ids, id)
	}
	return ids, nil
}

func convertResource(r models.UserResource) models.ResourceEntry {
	return models.ResourceEntry{
		ID:        r.ID,
		UserID:    r.UserID,
		Name:      r.Name,
		Path:      r.Path,
		Size:      r.Size,
		IsDir:     r.IsDir,
		CreatedAt: r.CreatedAt,
		UpdatedAt: r.UpdatedAt,
	}
}

func convertResources(recs []models.UserResource) []models.ResourceEntry {
	entries := make([]models.ResourceEntry, 0, len(recs))
	for _, r := range recs {
		entries = append(entries, convertResource(r))
	}
	return entries
}

func convertResourceToModel(e models.ResourceEntry) *model.UserResource {
	return &model.UserResource{
		ID:        int64(e.ID),
		UserID:    int64(e.UserID),
		Name:      e.Name,
		Path:      e.Path,
		Size:      int(e.Size),
		IsDir:     e.IsDir,
		CreatedAt: e.CreatedAt,
		UpdatedAt: e.UpdatedAt,
	}
}

// ---- subscription publishing -----------------------------------------------

func (s *ResourceService) publishResourcesAdded(ctx context.Context, entries []models.ResourceEntry) {
	if s.ss == nil || len(entries) == 0 {
		return
	}
	for _, e := range entries {
		s.ss.NewResourcePublisher(int64(e.UserID)).ResourceAdded(ctx, convertResourceToModel(e))
	}
}

func (s *ResourceService) publishResourcesUpdated(ctx context.Context, entries []models.ResourceEntry) {
	if s.ss == nil || len(entries) == 0 {
		return
	}
	for _, e := range entries {
		s.ss.NewResourcePublisher(int64(e.UserID)).ResourceUpdated(ctx, convertResourceToModel(e))
	}
}

func (s *ResourceService) publishResourcesDeleted(ctx context.Context, entries []models.ResourceEntry) {
	if s.ss == nil || len(entries) == 0 {
		return
	}
	for _, e := range entries {
		s.ss.NewResourcePublisher(int64(e.UserID)).ResourceDeleted(ctx, convertResourceToModel(e))
	}
}

// ---- utility ---------------------------------------------------------------

// isUniqueViolation returns true if err is a unique-constraint violation. Matches Postgres
// ("duplicate key value violates unique constraint", SQLSTATE 23505) and SQLite ("UNIQUE
// constraint failed") case-insensitively.
func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "unique") ||
		strings.Contains(msg, "duplicate") ||
		strings.Contains(msg, "23505")
}
