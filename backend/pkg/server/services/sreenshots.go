package services

import (
	"errors"
	"fmt"
	"net/http"
	"path/filepath"
	"slices"
	"strconv"

	"pentagi/pkg/server/logger"
	"pentagi/pkg/server/models"
	"pentagi/pkg/server/rdb"
	"pentagi/pkg/server/response"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

type screenshots struct {
	Screenshots []models.Screenshot `json:"screenshots"`
	Total       uint64              `json:"total"`
}

type screenshotsGrouped struct {
	Grouped []string `json:"grouped"`
	Total   uint64   `json:"total"`
}

var screenshotsSQLMappers = map[string]any{
	"id":         "{{table}}.id",
	"name":       "{{table}}.name",
	"url":        "{{table}}.url",
	"flow_id":    "{{table}}.flow_id",
	"task_id":    "{{table}}.task_id",
	"subtask_id": "{{table}}.subtask_id",
	"created_at": "{{table}}.created_at",
	"data":       "({{table}}.name || ' ' || {{table}}.url)",
}

type ScreenshotService struct {
	db      *gorm.DB
	dataDir string
}

func NewScreenshotService(db *gorm.DB, dataDir string) *ScreenshotService {
	return &ScreenshotService{
		db:      db,
		dataDir: dataDir,
	}
}

// GetScreenshots is a function to return screenshots list
// @Summary Retrieve screenshots list
// @Tags Screenshots
// @Produce json
// @Security BearerAuth
// @Param request query rdb.TableQuery true "query table params"
// @Success 200 {object} response.successResp{data=screenshots} "screenshots list received successful"
// @Failure 400 {object} response.errorResp "invalid query request data"
// @Failure 403 {object} response.errorResp "getting screenshots not permitted"
// @Failure 500 {object} response.errorResp "internal error on getting screenshots"
// @Router /screenshots/ [get]
func (s *ScreenshotService) GetScreenshots(c *gin.Context) {
	var (
		err   error
		query rdb.TableQuery
		resp  screenshots
	)

	if err = c.ShouldBindQuery(&query); err != nil {
		logger.FromContext(c).WithError(err).Errorf("error binding query")
		response.Error(c, response.ErrScreenshotsInvalidRequest, err)
		return
	}

	uid := c.GetUint64("uid")
	privs := c.GetStringSlice("prm")
	var scope func(db *gorm.DB) *gorm.DB
	if slices.Contains(privs, "screenshots.admin") {
		scope = func(db *gorm.DB) *gorm.DB {
			return db.
				Joins("INNER JOIN flows f ON f.id = screenshots.flow_id")
		}
	} else if slices.Contains(privs, "screenshots.view") {
		scope = func(db *gorm.DB) *gorm.DB {
			return db.
				Joins("INNER JOIN flows f ON f.id = screenshots.flow_id").
				Where("f.user_id = ?", uid)
		}
	} else {
		logger.FromContext(c).Errorf("error filtering user role permissions: permission not found")
		response.Error(c, response.ErrNotPermitted, nil)
		return
	}

	query.Init("screenshots", screenshotsSQLMappers)

	if query.Group != "" {
		if _, ok := screenshotsSQLMappers[query.Group]; !ok {
			logger.FromContext(c).Errorf("error finding screenshots grouped: group field not found")
			response.Error(c, response.ErrScreenshotsInvalidRequest, errors.New("group field not found"))
			return
		}

		var respGrouped screenshotsGrouped
		if respGrouped.Total, err = query.QueryGrouped(s.db, &respGrouped.Grouped, scope); err != nil {
			logger.FromContext(c).WithError(err).Errorf("error finding screenshots grouped")
			response.Error(c, response.ErrInternal, err)
			return
		}

		response.Success(c, http.StatusOK, respGrouped)
		return
	}

	if resp.Total, err = query.Query(s.db, &resp.Screenshots, scope); err != nil {
		logger.FromContext(c).WithError(err).Errorf("error finding screenshots")
		response.Error(c, response.ErrInternal, err)
		return
	}

	for i := 0; i < len(resp.Screenshots); i++ {
		if err = resp.Screenshots[i].Valid(); err != nil {
			logger.FromContext(c).WithError(err).Errorf("error validating screenshot data '%d'", resp.Screenshots[i].ID)
			response.Error(c, response.ErrScreenshotsInvalidData, err)
			return
		}
	}

	response.Success(c, http.StatusOK, resp)
}

// GetFlowScreenshots is a function to return screenshots list by flow id
// @Summary Retrieve screenshots list by flow id
// @Tags Screenshots
// @Produce json
// @Security BearerAuth
// @Param flowID path int true "flow id" minimum(0)
// @Param request query rdb.TableQuery true "query table params"
// @Success 200 {object} response.successResp{data=screenshots} "screenshots list received successful"
// @Failure 400 {object} response.errorResp "invalid query request data"
// @Failure 403 {object} response.errorResp "getting screenshots not permitted"
// @Failure 500 {object} response.errorResp "internal error on getting screenshots"
// @Router /flows/{flowID}/screenshots/ [get]
func (s *ScreenshotService) GetFlowScreenshots(c *gin.Context) {
	var (
		err    error
		flowID uint64
		query  rdb.TableQuery
		resp   screenshots
	)

	if flowID, err = strconv.ParseUint(c.Param("flowID"), 10, 64); err != nil {
		logger.FromContext(c).WithError(err).Errorf("error parsing flow id")
		response.Error(c, response.ErrScreenshotsInvalidRequest, err)
		return
	}

	if err = c.ShouldBindQuery(&query); err != nil {
		logger.FromContext(c).WithError(err).Errorf("error binding query")
		response.Error(c, response.ErrScreenshotsInvalidRequest, err)
		return
	}

	uid := c.GetUint64("uid")
	privs := c.GetStringSlice("prm")
	var scope func(db *gorm.DB) *gorm.DB
	if slices.Contains(privs, "screenshots.admin") {
		scope = func(db *gorm.DB) *gorm.DB {
			return db.
				Joins("INNER JOIN flows f ON f.id = screenshots.flow_id").
				Where("f.id = ?", flowID)
		}
	} else if slices.Contains(privs, "screenshots.view") {
		scope = func(db *gorm.DB) *gorm.DB {
			return db.
				Joins("INNER JOIN flows f ON f.id = screenshots.flow_id").
				Where("f.id = ? AND f.user_id = ?", flowID, uid)
		}
	} else {
		logger.FromContext(c).Errorf("error filtering user role permissions: permission not found")
		response.Error(c, response.ErrNotPermitted, nil)
		return
	}

	query.Init("screenshots", screenshotsSQLMappers)

	if query.Group != "" {
		if _, ok := screenshotsSQLMappers[query.Group]; !ok {
			logger.FromContext(c).Errorf("error finding screenshots grouped: group field not found")
			response.Error(c, response.ErrScreenshotsInvalidRequest, errors.New("group field not found"))
			return
		}

		var respGrouped screenshotsGrouped
		if respGrouped.Total, err = query.QueryGrouped(s.db, &respGrouped.Grouped, scope); err != nil {
			logger.FromContext(c).WithError(err).Errorf("error finding screenshots grouped")
			response.Error(c, response.ErrInternal, err)
			return
		}

		response.Success(c, http.StatusOK, respGrouped)
		return
	}

	if resp.Total, err = query.Query(s.db, &resp.Screenshots, scope); err != nil {
		logger.FromContext(c).WithError(err).Errorf("error finding screenshots")
		response.Error(c, response.ErrInternal, err)
		return
	}

	for i := 0; i < len(resp.Screenshots); i++ {
		if err = resp.Screenshots[i].Valid(); err != nil {
			logger.FromContext(c).WithError(err).Errorf("error validating screenshot data '%d'", resp.Screenshots[i].ID)
			response.Error(c, response.ErrScreenshotsInvalidData, err)
			return
		}
	}

	response.Success(c, http.StatusOK, resp)
}

// GetFlowScreenshot is a function to return screenshot info by id and flow id
// @Summary Retrieve screenshot info by id and flow id
// @Tags Screenshots
// @Produce json
// @Security BearerAuth
// @Param flowID path int true "flow id" minimum(0)
// @Param screenshotID path int true "screenshot id" minimum(0)
// @Success 200 {object} response.successResp{data=models.Screenshot} "screenshot info received successful"
// @Failure 403 {object} response.errorResp "getting screenshot not permitted"
// @Failure 404 {object} response.errorResp "screenshot not found"
// @Failure 500 {object} response.errorResp "internal error on getting screenshot"
// @Router /flows/{flowID}/screenshots/{screenshotID} [get]
func (s *ScreenshotService) GetFlowScreenshot(c *gin.Context) {
	var (
		err          error
		flowID       uint64
		screenshotID uint64
		resp         models.Screenshot
	)

	if flowID, err = strconv.ParseUint(c.Param("flowID"), 10, 64); err != nil {
		logger.FromContext(c).WithError(err).Errorf("error parsing flow id")
		response.Error(c, response.ErrScreenshotsInvalidRequest, err)
		return
	}
	if screenshotID, err = strconv.ParseUint(c.Param("screenshotID"), 10, 64); err != nil {
		logger.FromContext(c).WithError(err).Errorf("error parsing screenshot id")
		response.Error(c, response.ErrScreenshotsInvalidRequest, err)
		return
	}

	uid := c.GetUint64("uid")
	privs := c.GetStringSlice("prm")
	var scope func(db *gorm.DB) *gorm.DB
	if slices.Contains(privs, "screenshots.admin") {
		scope = func(db *gorm.DB) *gorm.DB {
			return db.Where("f.id = ?", flowID)
		}
	} else if slices.Contains(privs, "screenshots.view") {
		scope = func(db *gorm.DB) *gorm.DB {
			return db.Where("f.id = ? AND f.user_id = ?", flowID, uid)
		}
	} else {
		logger.FromContext(c).Errorf("error filtering user role permissions: permission not found")
		response.Error(c, response.ErrNotPermitted, nil)
		return
	}

	err = s.db.Model(&resp).
		Joins("INNER JOIN flows f ON f.id = flow_id").
		Scopes(scope).
		Where("screenshots.id = ?", screenshotID).
		Take(&resp).Error
	if err != nil {
		logger.FromContext(c).WithError(err).Errorf("error on getting screenshot by id")
		if gorm.IsRecordNotFoundError(err) {
			response.Error(c, response.ErrScreenshotsNotFound, err)
		} else {
			response.Error(c, response.ErrInternal, err)
		}
		return
	}

	response.Success(c, http.StatusOK, resp)
}

// GetFlowScreenshotFile is a function to return screenshot file by id and flow id
// @Summary Retrieve screenshot file by id and flow id
// @Tags Screenshots
// @Produce png,json
// @Security BearerAuth
// @Param flowID path int true "flow id" minimum(0)
// @Param screenshotID path int true "screenshot id" minimum(0)
// @Success 200 {file} file "screenshot file"
// @Failure 403 {object} response.errorResp "getting screenshot not permitted"
// @Failure 500 {object} response.errorResp "internal error on getting screenshot"
// @Router /flows/{flowID}/screenshots/{screenshotID}/file [get]
func (s *ScreenshotService) GetFlowScreenshotFile(c *gin.Context) {
	var (
		err          error
		flowID       uint64
		screenshotID uint64
		resp         models.Screenshot
	)

	if flowID, err = strconv.ParseUint(c.Param("flowID"), 10, 64); err != nil {
		logger.FromContext(c).WithError(err).Errorf("error parsing flow id")
		response.Error(c, response.ErrScreenshotsInvalidRequest, err)
		return
	}
	if screenshotID, err = strconv.ParseUint(c.Param("screenshotID"), 10, 64); err != nil {
		logger.FromContext(c).WithError(err).Errorf("error parsing screenshot id")
		response.Error(c, response.ErrScreenshotsInvalidRequest, err)
		return
	}

	uid := c.GetUint64("uid")
	privs := c.GetStringSlice("prm")
	var scope func(db *gorm.DB) *gorm.DB
	if slices.Contains(privs, "screenshots.admin") {
		scope = func(db *gorm.DB) *gorm.DB {
			return db.Where("f.id = ?", flowID)
		}
	} else if slices.Contains(privs, "screenshots.download") {
		scope = func(db *gorm.DB) *gorm.DB {
			return db.Where("f.id = ? AND f.user_id = ?", flowID, uid)
		}
	} else {
		logger.FromContext(c).Errorf("error filtering user role permissions: permission not found")
		response.Error(c, response.ErrNotPermitted, nil)
		return
	}

	err = s.db.Model(&resp).
		Joins("INNER JOIN flows f ON f.id = flow_id").
		Scopes(scope).
		Where("screenshots.id = ?", screenshotID).
		Take(&resp).Error
	if err != nil {
		logger.FromContext(c).WithError(err).Errorf("error on getting screenshot by id")
		if gorm.IsRecordNotFoundError(err) {
			response.Error(c, response.ErrScreenshotsNotFound, err)
		} else {
			response.Error(c, response.ErrInternal, err)
		}
		return
	}

	flowDirName := fmt.Sprintf("flow-%d", resp.FlowID)
	c.File(filepath.Join(s.dataDir, "screenshots", flowDirName, resp.Name))
}
