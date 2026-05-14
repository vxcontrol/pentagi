package services

import (
	"errors"
	"net/http"
	"slices"
	"strconv"

	"pentagi/pkg/server/logger"
	"pentagi/pkg/server/models"
	"pentagi/pkg/server/rdb"
	"pentagi/pkg/server/response"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

type toolcalls struct {
	Toolcalls []models.Toolcall `json:"toolcalls"`
	Total     uint64            `json:"total"`
}

type toolcallsGrouped struct {
	Grouped []string `json:"grouped"`
	Total   uint64   `json:"total"`
}

var toolcallsSQLMappers = map[string]any{
	"id":               "{{table}}.id",
	"call_id":          "{{table}}.call_id",
	"status":           "{{table}}.status",
	"name":             "{{table}}.name",
	"args":             "{{table}}.args",
	"result":           "{{table}}.result",
	"duration_seconds": "{{table}}.duration_seconds",
	"flow_id":          "{{table}}.flow_id",
	"task_id":          "{{table}}.task_id",
	"subtask_id":       "{{table}}.subtask_id",
	"created_at":       "{{table}}.created_at",
	"updated_at":       "{{table}}.updated_at",
	"data":             "({{table}}.name || ' ' || {{table}}.call_id || ' ' || {{table}}.args::text || ' ' || {{table}}.result)",
}

type ToolcallService struct {
	db *gorm.DB
}

func NewToolcallService(db *gorm.DB) *ToolcallService {
	return &ToolcallService{
		db: db,
	}
}

// GetToolcalls is a function to return toolcalls list
// @Summary Retrieve toolcalls list
// @Tags Toolcalls
// @Produce json
// @Security BearerAuth
// @Param request query rdb.TableQuery true "query table params"
// @Success 200 {object} response.successResp{data=toolcalls} "toolcalls list received successful"
// @Failure 400 {object} response.errorResp "invalid query request data"
// @Failure 403 {object} response.errorResp "getting toolcalls not permitted"
// @Failure 500 {object} response.errorResp "internal error on getting toolcalls"
// @Router /toolcalls/ [get]
func (s *ToolcallService) GetToolcalls(c *gin.Context) {
	var (
		err   error
		query rdb.TableQuery
		resp  toolcalls
	)

	if err = c.ShouldBindQuery(&query); err != nil {
		logger.FromContext(c).WithError(err).Errorf("error binding query")
		response.Error(c, response.ErrToolcallsInvalidRequest, err)
		return
	}

	uid := c.GetUint64("uid")
	privs := c.GetStringSlice("prm")
	var scope func(db *gorm.DB) *gorm.DB
	if slices.Contains(privs, "toolcalls.admin") {
		scope = func(db *gorm.DB) *gorm.DB {
			return db.
				Joins("INNER JOIN flows f ON f.id = toolcalls.flow_id")
		}
	} else if slices.Contains(privs, "toolcalls.view") {
		scope = func(db *gorm.DB) *gorm.DB {
			return db.
				Joins("INNER JOIN flows f ON f.id = toolcalls.flow_id").
				Where("f.user_id = ?", uid)
		}
	} else {
		logger.FromContext(c).Errorf("error filtering user role permissions: permission not found")
		response.Error(c, response.ErrNotPermitted, nil)
		return
	}

	query.Init("toolcalls", toolcallsSQLMappers)

	if query.Group != "" {
		if _, ok := toolcallsSQLMappers[query.Group]; !ok {
			logger.FromContext(c).Errorf("error finding toolcalls grouped: group field not found")
			response.Error(c, response.ErrToolcallsInvalidRequest, errors.New("group field not found"))
			return
		}

		var respGrouped toolcallsGrouped
		if respGrouped.Total, err = query.QueryGrouped(s.db, &respGrouped.Grouped, scope); err != nil {
			logger.FromContext(c).WithError(err).Errorf("error finding toolcalls grouped")
			response.Error(c, response.ErrInternal, err)
			return
		}

		response.Success(c, http.StatusOK, respGrouped)
		return
	}

	if resp.Total, err = query.Query(s.db, &resp.Toolcalls, scope); err != nil {
		logger.FromContext(c).WithError(err).Errorf("error finding toolcalls")
		response.Error(c, response.ErrInternal, err)
		return
	}

	for i := range resp.Toolcalls {
		if err = resp.Toolcalls[i].Valid(); err != nil {
			logger.FromContext(c).WithError(err).Errorf("error validating toolcall data '%d'", resp.Toolcalls[i].ID)
			response.Error(c, response.ErrToolcallsInvalidData, err)
			return
		}
	}

	response.Success(c, http.StatusOK, resp)
}

// GetFlowToolcalls is a function to return toolcalls list by flow id
// @Summary Retrieve toolcalls list by flow id
// @Tags Toolcalls
// @Produce json
// @Security BearerAuth
// @Param flowID path int true "flow id" minimum(0)
// @Param request query rdb.TableQuery true "query table params"
// @Success 200 {object} response.successResp{data=toolcalls} "toolcalls list received successful"
// @Failure 400 {object} response.errorResp "invalid query request data"
// @Failure 403 {object} response.errorResp "getting toolcalls not permitted"
// @Failure 500 {object} response.errorResp "internal error on getting toolcalls"
// @Router /flows/{flowID}/toolcalls/ [get]
func (s *ToolcallService) GetFlowToolcalls(c *gin.Context) {
	var (
		err    error
		flowID uint64
		query  rdb.TableQuery
		resp   toolcalls
	)

	if flowID, err = strconv.ParseUint(c.Param("flowID"), 10, 64); err != nil {
		logger.FromContext(c).WithError(err).Errorf("error parsing flow id")
		response.Error(c, response.ErrToolcallsInvalidRequest, err)
		return
	}

	if err = c.ShouldBindQuery(&query); err != nil {
		logger.FromContext(c).WithError(err).Errorf("error binding query")
		response.Error(c, response.ErrToolcallsInvalidRequest, err)
		return
	}

	uid := c.GetUint64("uid")
	privs := c.GetStringSlice("prm")
	var scope func(db *gorm.DB) *gorm.DB
	if slices.Contains(privs, "toolcalls.admin") {
		scope = func(db *gorm.DB) *gorm.DB {
			return db.
				Joins("INNER JOIN flows f ON f.id = toolcalls.flow_id").
				Where("f.id = ?", flowID)
		}
	} else if slices.Contains(privs, "toolcalls.view") {
		scope = func(db *gorm.DB) *gorm.DB {
			return db.
				Joins("INNER JOIN flows f ON f.id = toolcalls.flow_id").
				Where("f.id = ? AND f.user_id = ?", flowID, uid)
		}
	} else {
		logger.FromContext(c).Errorf("error filtering user role permissions: permission not found")
		response.Error(c, response.ErrNotPermitted, nil)
		return
	}

	query.Init("toolcalls", toolcallsSQLMappers)

	if query.Group != "" {
		if _, ok := toolcallsSQLMappers[query.Group]; !ok {
			logger.FromContext(c).Errorf("error finding toolcalls grouped: group field not found")
			response.Error(c, response.ErrToolcallsInvalidRequest, errors.New("group field not found"))
			return
		}

		var respGrouped toolcallsGrouped
		if respGrouped.Total, err = query.QueryGrouped(s.db, &respGrouped.Grouped, scope); err != nil {
			logger.FromContext(c).WithError(err).Errorf("error finding toolcalls grouped")
			response.Error(c, response.ErrInternal, err)
			return
		}

		response.Success(c, http.StatusOK, respGrouped)
		return
	}

	if resp.Total, err = query.Query(s.db, &resp.Toolcalls, scope); err != nil {
		logger.FromContext(c).WithError(err).Errorf("error finding toolcalls")
		response.Error(c, response.ErrInternal, err)
		return
	}

	for i := range resp.Toolcalls {
		if err = resp.Toolcalls[i].Valid(); err != nil {
			logger.FromContext(c).WithError(err).Errorf("error validating toolcall data '%d'", resp.Toolcalls[i].ID)
			response.Error(c, response.ErrToolcallsInvalidData, err)
			return
		}
	}

	response.Success(c, http.StatusOK, resp)
}

// GetFlowToolcall is a function to return toolcall info by id and flow id
// @Summary Retrieve toolcall info by id and flow id
// @Tags Toolcalls
// @Produce json
// @Security BearerAuth
// @Param flowID path int true "flow id" minimum(0)
// @Param toolcallID path int true "toolcall id" minimum(0)
// @Success 200 {object} response.successResp{data=models.Toolcall} "toolcall info received successful"
// @Failure 400 {object} response.errorResp "invalid request data"
// @Failure 403 {object} response.errorResp "getting toolcall not permitted"
// @Failure 404 {object} response.errorResp "toolcall not found"
// @Failure 500 {object} response.errorResp "internal error on getting toolcall"
// @Router /flows/{flowID}/toolcalls/{toolcallID} [get]
func (s *ToolcallService) GetFlowToolcall(c *gin.Context) {
	var (
		err        error
		flowID     uint64
		toolcallID uint64
		resp       models.Toolcall
	)

	if flowID, err = strconv.ParseUint(c.Param("flowID"), 10, 64); err != nil {
		logger.FromContext(c).WithError(err).Errorf("error parsing flow id")
		response.Error(c, response.ErrToolcallsInvalidRequest, err)
		return
	}
	if toolcallID, err = strconv.ParseUint(c.Param("toolcallID"), 10, 64); err != nil {
		logger.FromContext(c).WithError(err).Errorf("error parsing toolcall id")
		response.Error(c, response.ErrToolcallsInvalidRequest, err)
		return
	}

	uid := c.GetUint64("uid")
	privs := c.GetStringSlice("prm")
	var scope func(db *gorm.DB) *gorm.DB
	if slices.Contains(privs, "toolcalls.admin") {
		scope = func(db *gorm.DB) *gorm.DB {
			return db.Where("f.id = ?", flowID)
		}
	} else if slices.Contains(privs, "toolcalls.view") {
		scope = func(db *gorm.DB) *gorm.DB {
			return db.Where("f.id = ? AND f.user_id = ?", flowID, uid)
		}
	} else {
		logger.FromContext(c).Errorf("error filtering user role permissions: permission not found")
		response.Error(c, response.ErrNotPermitted, nil)
		return
	}

	err = s.db.Model(&resp).
		Joins("INNER JOIN flows f ON f.id = toolcalls.flow_id").
		Scopes(scope).
		Where("toolcalls.id = ?", toolcallID).
		Take(&resp).Error
	if err != nil {
		logger.FromContext(c).WithError(err).Errorf("error on getting toolcall by id")
		if gorm.IsRecordNotFoundError(err) {
			response.Error(c, response.ErrToolcallsNotFound, err)
		} else {
			response.Error(c, response.ErrInternal, err)
		}
		return
	}

	response.Success(c, http.StatusOK, resp)
}
