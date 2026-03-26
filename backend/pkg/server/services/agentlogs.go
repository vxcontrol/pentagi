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

type agentlogs struct {
	AgentLogs []models.Agentlog `json:"agentlogs"`
	Total     uint64            `json:"total"`
}

type agentlogsGrouped struct {
	Grouped []string `json:"grouped"`
	Total   uint64   `json:"total"`
}

var agentlogsSQLMappers = map[string]any{
	"id":         "{{table}}.id",
	"initiator":  "{{table}}.initiator",
	"executor":   "{{table}}.executor",
	"task":       "{{table}}.task",
	"result":     "{{table}}.result",
	"flow_id":    "{{table}}.flow_id",
	"task_id":    "{{table}}.task_id",
	"subtask_id": "{{table}}.subtask_id",
	"created_at": "{{table}}.created_at",
	"data":       "({{table}}.task || ' ' || {{table}}.result)",
}

type AgentlogService struct {
	db *gorm.DB
}

func NewAgentlogService(db *gorm.DB) *AgentlogService {
	return &AgentlogService{
		db: db,
	}
}

// GetAgentlogs is a function to return agentlogs list
// @Summary Retrieve agentlogs list
// @Tags Agentlogs
// @Produce json
// @Security BearerAuth
// @Param request query rdb.TableQuery true "query table params"
// @Success 200 {object} response.successResp{data=agentlogs} "agentlogs list received successful"
// @Failure 400 {object} response.errorResp "invalid query request data"
// @Failure 403 {object} response.errorResp "getting agentlogs not permitted"
// @Failure 500 {object} response.errorResp "internal error on getting agentlogs"
// @Router /agentlogs/ [get]
func (s *AgentlogService) GetAgentlogs(c *gin.Context) {
	var (
		err   error
		query rdb.TableQuery
		resp  agentlogs
	)

	if err = c.ShouldBindQuery(&query); err != nil {
		logger.FromContext(c).WithError(err).Errorf("error binding query")
		response.Error(c, response.ErrAgentlogsInvalidRequest, err)
		return
	}

	uid := c.GetUint64("uid")
	privs := c.GetStringSlice("prm")
	var scope func(db *gorm.DB) *gorm.DB
	if slices.Contains(privs, "agentlogs.admin") {
		scope = func(db *gorm.DB) *gorm.DB {
			return db.
				Joins("INNER JOIN flows f ON f.id = flow_id")
		}
	} else if slices.Contains(privs, "agentlogs.view") {
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

	query.Init("agentlogs", agentlogsSQLMappers)

	if query.Group != "" {
		if _, ok := agentlogsSQLMappers[query.Group]; !ok {
			logger.FromContext(c).Errorf("error finding agentlogs grouped: group field not found")
			response.Error(c, response.ErrAgentlogsInvalidRequest, errors.New("group field not found"))
			return
		}

		var respGrouped agentlogsGrouped
		if respGrouped.Total, err = query.QueryGrouped(s.db, &respGrouped.Grouped, scope); err != nil {
			logger.FromContext(c).WithError(err).Errorf("error finding agentlogs grouped")
			response.Error(c, response.ErrInternal, err)
			return
		}

		response.Success(c, http.StatusOK, respGrouped)
		return
	}

	if resp.Total, err = query.Query(s.db, &resp.AgentLogs, scope); err != nil {
		logger.FromContext(c).WithError(err).Errorf("error finding agentlogs")
		response.Error(c, response.ErrInternal, err)
		return
	}

	for i := 0; i < len(resp.AgentLogs); i++ {
		if err = resp.AgentLogs[i].Valid(); err != nil {
			logger.FromContext(c).WithError(err).Errorf("error validating agentlog data '%d'", resp.AgentLogs[i].ID)
			response.Error(c, response.ErrAgentlogsInvalidData, err)
			return
		}
	}

	response.Success(c, http.StatusOK, resp)
}

// GetFlowAgentlogs is a function to return agentlogs list by flow id
// @Summary Retrieve agentlogs list by flow id
// @Tags Agentlogs
// @Produce json
// @Security BearerAuth
// @Param flowID path int true "flow id" minimum(0)
// @Param request query rdb.TableQuery true "query table params"
// @Success 200 {object} response.successResp{data=agentlogs} "agentlogs list received successful"
// @Failure 400 {object} response.errorResp "invalid query request data"
// @Failure 403 {object} response.errorResp "getting agentlogs not permitted"
// @Failure 500 {object} response.errorResp "internal error on getting agentlogs"
// @Router /flows/{flowID}/agentlogs/ [get]
func (s *AgentlogService) GetFlowAgentlogs(c *gin.Context) {
	var (
		err    error
		flowID uint64
		query  rdb.TableQuery
		resp   agentlogs
	)

	if flowID, err = strconv.ParseUint(c.Param("flowID"), 10, 64); err != nil {
		logger.FromContext(c).WithError(err).Errorf("error parsing flow id")
		response.Error(c, response.ErrAgentlogsInvalidRequest, err)
		return
	}

	if err = c.ShouldBindQuery(&query); err != nil {
		logger.FromContext(c).WithError(err).Errorf("error binding query")
		response.Error(c, response.ErrAgentlogsInvalidRequest, err)
		return
	}

	uid := c.GetUint64("uid")
	privs := c.GetStringSlice("prm")
	var scope func(db *gorm.DB) *gorm.DB
	if slices.Contains(privs, "agentlogs.admin") {
		scope = func(db *gorm.DB) *gorm.DB {
			return db.
				Joins("INNER JOIN flows f ON f.id = flow_id").
				Where("f.id = ?", flowID)
		}
	} else if slices.Contains(privs, "agentlogs.view") {
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

	query.Init("agentlogs", agentlogsSQLMappers)

	if query.Group != "" {
		if _, ok := agentlogsSQLMappers[query.Group]; !ok {
			logger.FromContext(c).Errorf("error finding agentlogs grouped: group field not found")
			response.Error(c, response.ErrAgentlogsInvalidRequest, errors.New("group field not found"))
			return
		}

		var respGrouped agentlogsGrouped
		if respGrouped.Total, err = query.QueryGrouped(s.db, &respGrouped.Grouped, scope); err != nil {
			logger.FromContext(c).WithError(err).Errorf("error finding agentlogs grouped")
			response.Error(c, response.ErrInternal, err)
			return
		}

		response.Success(c, http.StatusOK, respGrouped)
		return
	}

	if resp.Total, err = query.Query(s.db, &resp.AgentLogs, scope); err != nil {
		logger.FromContext(c).WithError(err).Errorf("error finding agentlogs")
		response.Error(c, response.ErrInternal, err)
		return
	}

	for i := 0; i < len(resp.AgentLogs); i++ {
		if err = resp.AgentLogs[i].Valid(); err != nil {
			logger.FromContext(c).WithError(err).Errorf("error validating agentlog data '%d'", resp.AgentLogs[i].ID)
			response.Error(c, response.ErrAgentlogsInvalidData, err)
			return
		}
	}

	response.Success(c, http.StatusOK, resp)
}
