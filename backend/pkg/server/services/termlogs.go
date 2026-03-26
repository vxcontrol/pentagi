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

type termlogs struct {
	TermLogs []models.Termlog `json:"termlogs"`
	Total    uint64           `json:"total"`
}

type termlogsGrouped struct {
	Grouped []string `json:"grouped"`
	Total   uint64   `json:"total"`
}

var termlogsSQLMappers = map[string]any{
	"id":           "{{table}}.id",
	"type":         "{{table}}.type",
	"text":         "{{table}}.text",
	"container_id": "{{table}}.container_id",
	"flow_id":      "{{table}}.flow_id",
	"task_id":      "{{table}}.task_id",
	"subtask_id":   "{{table}}.subtask_id",
	"created_at":   "{{table}}.created_at",
	"data":         "({{table}}.type || ' ' || {{table}}.text)",
}

type TermlogService struct {
	db *gorm.DB
}

func NewTermlogService(db *gorm.DB) *TermlogService {
	return &TermlogService{
		db: db,
	}
}

// GetTermlogs is a function to return termlogs list
// @Summary Retrieve termlogs list
// @Tags Termlogs
// @Produce json
// @Security BearerAuth
// @Param request query rdb.TableQuery true "query table params"
// @Success 200 {object} response.successResp{data=termlogs} "termlogs list received successful"
// @Failure 400 {object} response.errorResp "invalid query request data"
// @Failure 403 {object} response.errorResp "getting termlogs not permitted"
// @Failure 500 {object} response.errorResp "internal error on getting termlogs"
// @Router /termlogs/ [get]
func (s *TermlogService) GetTermlogs(c *gin.Context) {
	var (
		err   error
		query rdb.TableQuery
		resp  termlogs
	)

	if err = c.ShouldBindQuery(&query); err != nil {
		logger.FromContext(c).WithError(err).Errorf("error binding query")
		response.Error(c, response.ErrTermlogsInvalidRequest, err)
		return
	}

	uid := c.GetUint64("uid")
	privs := c.GetStringSlice("prm")
	var scope func(db *gorm.DB) *gorm.DB
	if slices.Contains(privs, "termlogs.admin") {
		scope = func(db *gorm.DB) *gorm.DB {
			return db.
				Joins("INNER JOIN flows f ON f.id = flow_id")
		}
	} else if slices.Contains(privs, "termlogs.view") {
		scope = func(db *gorm.DB) *gorm.DB {
			return db.
				Joins("INNER JOIN flows f ON f.id = flow_id").
				Where("f.user_id = ?", uid)
		}
	} else {
		logger.FromContext(c).Errorf("error filtering user role permissions: permission not found")
		response.Error(c, response.ErrNotPermitted, nil)
		return
	}

	query.Init("termlogs", termlogsSQLMappers)

	if query.Group != "" {
		if _, ok := termlogsSQLMappers[query.Group]; !ok {
			logger.FromContext(c).Errorf("error finding termlogs grouped: group field not found")
			response.Error(c, response.ErrTermlogsInvalidRequest, errors.New("group field not found"))
			return
		}

		var respGrouped termlogsGrouped
		if respGrouped.Total, err = query.QueryGrouped(s.db, &respGrouped.Grouped, scope); err != nil {
			logger.FromContext(c).WithError(err).Errorf("error finding termlogs grouped")
			response.Error(c, response.ErrInternal, err)
			return
		}

		response.Success(c, http.StatusOK, respGrouped)
		return
	}

	if resp.Total, err = query.Query(s.db, &resp.TermLogs, scope); err != nil {
		logger.FromContext(c).WithError(err).Errorf("error finding termlogs")
		response.Error(c, response.ErrInternal, err)
		return
	}

	for i := 0; i < len(resp.TermLogs); i++ {
		if err = resp.TermLogs[i].Valid(); err != nil {
			logger.FromContext(c).WithError(err).Errorf("error validating termlog data '%d'", resp.TermLogs[i].ID)
			response.Error(c, response.ErrTermlogsInvalidData, err)
			return
		}
	}

	response.Success(c, http.StatusOK, resp)
}

// GetFlowTermlogs is a function to return termlogs list by flow id
// @Summary Retrieve termlogs list by flow id
// @Tags Termlogs
// @Produce json
// @Security BearerAuth
// @Param flowID path int true "flow id" minimum(0)
// @Param request query rdb.TableQuery true "query table params"
// @Success 200 {object} response.successResp{data=termlogs} "termlogs list received successful"
// @Failure 400 {object} response.errorResp "invalid query request data"
// @Failure 403 {object} response.errorResp "getting termlogs not permitted"
// @Failure 500 {object} response.errorResp "internal error on getting termlogs"
// @Router /flows/{flowID}/termlogs/ [get]
func (s *TermlogService) GetFlowTermlogs(c *gin.Context) {
	var (
		err    error
		flowID uint64
		query  rdb.TableQuery
		resp   termlogs
	)

	if flowID, err = strconv.ParseUint(c.Param("flowID"), 10, 64); err != nil {
		logger.FromContext(c).WithError(err).Errorf("error parsing flow id")
		response.Error(c, response.ErrTermlogsInvalidRequest, err)
		return
	}

	if err = c.ShouldBindQuery(&query); err != nil {
		logger.FromContext(c).WithError(err).Errorf("error binding query")
		response.Error(c, response.ErrTermlogsInvalidRequest, err)
		return
	}

	uid := c.GetUint64("uid")
	privs := c.GetStringSlice("prm")
	var scope func(db *gorm.DB) *gorm.DB
	if slices.Contains(privs, "termlogs.admin") {
		scope = func(db *gorm.DB) *gorm.DB {
			return db.
				Joins("INNER JOIN flows f ON f.id = flow_id").
				Where("f.id = ?", flowID)
		}
	} else if slices.Contains(privs, "termlogs.view") {
		scope = func(db *gorm.DB) *gorm.DB {
			return db.
				Joins("INNER JOIN flows f ON f.id = flow_id").
				Where("f.id = ? AND f.user_id = ?", flowID, uid)
		}
	} else {
		logger.FromContext(c).Errorf("error filtering user role permissions: permission not found")
		response.Error(c, response.ErrNotPermitted, nil)
		return
	}

	query.Init("termlogs", termlogsSQLMappers)

	if query.Group != "" {
		if _, ok := termlogsSQLMappers[query.Group]; !ok {
			logger.FromContext(c).Errorf("error finding termlogs grouped: group field not found")
			response.Error(c, response.ErrTermlogsInvalidRequest, errors.New("group field not found"))
			return
		}

		var respGrouped termlogsGrouped
		if respGrouped.Total, err = query.QueryGrouped(s.db, &respGrouped.Grouped, scope); err != nil {
			logger.FromContext(c).WithError(err).Errorf("error finding termlogs grouped")
			response.Error(c, response.ErrInternal, err)
			return
		}

		response.Success(c, http.StatusOK, respGrouped)
		return
	}

	if resp.Total, err = query.Query(s.db, &resp.TermLogs, scope); err != nil {
		logger.FromContext(c).WithError(err).Errorf("error finding termlogs")
		response.Error(c, response.ErrInternal, err)
		return
	}

	for i := 0; i < len(resp.TermLogs); i++ {
		if err = resp.TermLogs[i].Valid(); err != nil {
			logger.FromContext(c).WithError(err).Errorf("error validating termlog data '%d'", resp.TermLogs[i].ID)
			response.Error(c, response.ErrTermlogsInvalidData, err)
			return
		}
	}

	response.Success(c, http.StatusOK, resp)
}
