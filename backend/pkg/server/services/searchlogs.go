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

type searchlogs struct {
	SearchLogs []models.Searchlog `json:"searchlogs"`
	Total      uint64             `json:"total"`
}

type searchlogsGrouped struct {
	Grouped []string `json:"grouped"`
	Total   uint64   `json:"total"`
}

var searchlogsSQLMappers = map[string]any{
	"id":         "{{table}}.id",
	"initiator":  "{{table}}.initiator",
	"executor":   "{{table}}.executor",
	"engine":     "{{table}}.engine",
	"query":      "{{table}}.query",
	"result":     "{{table}}.result",
	"flow_id":    "{{table}}.flow_id",
	"task_id":    "{{table}}.task_id",
	"subtask_id": "{{table}}.subtask_id",
	"created_at": "{{table}}.created_at",
	"data":       "({{table}}.query || ' ' || {{table}}.result)",
}

type SearchlogService struct {
	db *gorm.DB
}

func NewSearchlogService(db *gorm.DB) *SearchlogService {
	return &SearchlogService{
		db: db,
	}
}

// GetSearchlogs is a function to return searchlogs list
// @Summary Retrieve searchlogs list
// @Tags Searchlogs
// @Produce json
// @Security BearerAuth
// @Param request query rdb.TableQuery true "query table params"
// @Success 200 {object} response.successResp{data=searchlogs} "searchlogs list received successful"
// @Failure 400 {object} response.errorResp "invalid query request data"
// @Failure 403 {object} response.errorResp "getting searchlogs not permitted"
// @Failure 500 {object} response.errorResp "internal error on getting searchlogs"
// @Router /searchlogs/ [get]
func (s *SearchlogService) GetSearchlogs(c *gin.Context) {
	var (
		err   error
		query rdb.TableQuery
		resp  searchlogs
	)

	if err = c.ShouldBindQuery(&query); err != nil {
		logger.FromContext(c).WithError(err).Errorf("error binding query")
		response.Error(c, response.ErrSearchlogsInvalidRequest, err)
		return
	}

	uid := c.GetUint64("uid")
	privs := c.GetStringSlice("prm")
	var scope func(db *gorm.DB) *gorm.DB
	if slices.Contains(privs, "searchlogs.admin") {
		scope = func(db *gorm.DB) *gorm.DB {
			return db.
				Joins("INNER JOIN flows f ON f.id = flow_id")
		}
	} else if slices.Contains(privs, "searchlogs.view") {
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

	query.Init("searchlogs", searchlogsSQLMappers)

	if query.Group != "" {
		if _, ok := searchlogsSQLMappers[query.Group]; !ok {
			logger.FromContext(c).Errorf("error finding searchlogs grouped: group field not found")
			response.Error(c, response.ErrSearchlogsInvalidRequest, errors.New("group field not found"))
			return
		}

		var respGrouped searchlogsGrouped
		if respGrouped.Total, err = query.QueryGrouped(s.db, &respGrouped.Grouped, scope); err != nil {
			logger.FromContext(c).WithError(err).Errorf("error finding searchlogs grouped")
			response.Error(c, response.ErrInternal, err)
			return
		}

		response.Success(c, http.StatusOK, respGrouped)
		return
	}

	if resp.Total, err = query.Query(s.db, &resp.SearchLogs, scope); err != nil {
		logger.FromContext(c).WithError(err).Errorf("error finding searchlogs")
		response.Error(c, response.ErrInternal, err)
		return
	}

	for i := 0; i < len(resp.SearchLogs); i++ {
		if err = resp.SearchLogs[i].Valid(); err != nil {
			logger.FromContext(c).WithError(err).Errorf("error validating searchlog data '%d'", resp.SearchLogs[i].ID)
			response.Error(c, response.ErrSearchlogsInvalidData, err)
			return
		}
	}

	response.Success(c, http.StatusOK, resp)
}

// GetFlowSearchlogs is a function to return searchlogs list by flow id
// @Summary Retrieve searchlogs list by flow id
// @Tags Searchlogs
// @Produce json
// @Security BearerAuth
// @Param flowID path int true "flow id" minimum(0)
// @Param request query rdb.TableQuery true "query table params"
// @Success 200 {object} response.successResp{data=searchlogs} "searchlogs list received successful"
// @Failure 400 {object} response.errorResp "invalid query request data"
// @Failure 403 {object} response.errorResp "getting searchlogs not permitted"
// @Failure 500 {object} response.errorResp "internal error on getting searchlogs"
// @Router /flows/{flowID}/searchlogs/ [get]
func (s *SearchlogService) GetFlowSearchlogs(c *gin.Context) {
	var (
		err    error
		flowID uint64
		query  rdb.TableQuery
		resp   searchlogs
	)

	if flowID, err = strconv.ParseUint(c.Param("flowID"), 10, 64); err != nil {
		logger.FromContext(c).WithError(err).Errorf("error parsing flow id")
		response.Error(c, response.ErrSearchlogsInvalidRequest, err)
		return
	}

	if err = c.ShouldBindQuery(&query); err != nil {
		logger.FromContext(c).WithError(err).Errorf("error binding query")
		response.Error(c, response.ErrSearchlogsInvalidRequest, err)
		return
	}

	uid := c.GetUint64("uid")
	privs := c.GetStringSlice("prm")
	var scope func(db *gorm.DB) *gorm.DB
	if slices.Contains(privs, "searchlogs.admin") {
		scope = func(db *gorm.DB) *gorm.DB {
			return db.
				Joins("INNER JOIN flows f ON f.id = flow_id").
				Where("f.id = ?", flowID)
		}
	} else if slices.Contains(privs, "searchlogs.view") {
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

	query.Init("searchlogs", searchlogsSQLMappers)

	if query.Group != "" {
		if _, ok := searchlogsSQLMappers[query.Group]; !ok {
			logger.FromContext(c).Errorf("error finding searchlogs grouped: group field not found")
			response.Error(c, response.ErrSearchlogsInvalidRequest, errors.New("group field not found"))
			return
		}

		var respGrouped searchlogsGrouped
		if respGrouped.Total, err = query.QueryGrouped(s.db, &respGrouped.Grouped, scope); err != nil {
			logger.FromContext(c).WithError(err).Errorf("error finding searchlogs grouped")
			response.Error(c, response.ErrInternal, err)
			return
		}

		response.Success(c, http.StatusOK, respGrouped)
		return
	}

	if resp.Total, err = query.Query(s.db, &resp.SearchLogs, scope); err != nil {
		logger.FromContext(c).WithError(err).Errorf("error finding searchlogs")
		response.Error(c, response.ErrInternal, err)
		return
	}

	for i := 0; i < len(resp.SearchLogs); i++ {
		if err = resp.SearchLogs[i].Valid(); err != nil {
			logger.FromContext(c).WithError(err).Errorf("error validating searchlog data '%d'", resp.SearchLogs[i].ID)
			response.Error(c, response.ErrSearchlogsInvalidData, err)
			return
		}
	}

	response.Success(c, http.StatusOK, resp)
}
