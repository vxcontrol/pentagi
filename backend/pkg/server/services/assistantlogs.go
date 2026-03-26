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

type assistantlogs struct {
	AssistantLogs []models.Assistantlog `json:"assistantlogs"`
	Total         uint64                `json:"total"`
}

type assistantlogsGrouped struct {
	Grouped []string `json:"grouped"`
	Total   uint64   `json:"total"`
}

var assistantlogsSQLMappers = map[string]any{
	"id":            "{{table}}.id",
	"type":          "{{table}}.type",
	"message":       "{{table}}.message",
	"result":        "{{table}}.result",
	"result_format": "{{table}}.result_format",
	"flow_id":       "{{table}}.flow_id",
	"assistant_id":  "{{table}}.assistant_id",
	"created_at":    "{{table}}.created_at",
	"data":          "({{table}}.type || ' ' || {{table}}.message || ' ' || {{table}}.result)",
}

type AssistantlogService struct {
	db *gorm.DB
}

func NewAssistantlogService(db *gorm.DB) *AssistantlogService {
	return &AssistantlogService{
		db: db,
	}
}

// GetAssistantlogs is a function to return assistantlogs list
// @Summary Retrieve assistantlogs list
// @Tags Assistantlogs
// @Produce json
// @Security BearerAuth
// @Param request query rdb.TableQuery true "query table params"
// @Success 200 {object} response.successResp{data=assistantlogs} "assistantlogs list received successful"
// @Failure 400 {object} response.errorResp "invalid query request data"
// @Failure 403 {object} response.errorResp "getting assistantlogs not permitted"
// @Failure 500 {object} response.errorResp "internal error on getting assistantlogs"
// @Router /assistantlogs/ [get]
func (s *AssistantlogService) GetAssistantlogs(c *gin.Context) {
	var (
		err   error
		query rdb.TableQuery
		resp  assistantlogs
	)

	if err = c.ShouldBindQuery(&query); err != nil {
		logger.FromContext(c).WithError(err).Errorf("error binding query")
		response.Error(c, response.ErrAssistantlogsInvalidRequest, err)
		return
	}

	uid := c.GetUint64("uid")
	privs := c.GetStringSlice("prm")
	var scope func(db *gorm.DB) *gorm.DB
	if slices.Contains(privs, "assistantlogs.admin") {
		scope = func(db *gorm.DB) *gorm.DB {
			return db.
				Joins("INNER JOIN flows f ON f.id = flow_id")
		}
	} else if slices.Contains(privs, "assistantlogs.view") {
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

	query.Init("assistantlogs", assistantlogsSQLMappers)

	if query.Group != "" {
		if _, ok := assistantlogsSQLMappers[query.Group]; !ok {
			logger.FromContext(c).Errorf("error finding assistantlogs grouped: group field not found")
			response.Error(c, response.ErrAssistantlogsInvalidRequest, errors.New("group field not found"))
			return
		}

		var respGrouped assistantlogsGrouped
		if respGrouped.Total, err = query.QueryGrouped(s.db, &respGrouped.Grouped, scope); err != nil {
			logger.FromContext(c).WithError(err).Errorf("error finding assistantlogs grouped")
			response.Error(c, response.ErrInternal, err)
			return
		}

		response.Success(c, http.StatusOK, respGrouped)
		return
	}

	if resp.Total, err = query.Query(s.db, &resp.AssistantLogs, scope); err != nil {
		logger.FromContext(c).WithError(err).Errorf("error finding assistantlogs")
		response.Error(c, response.ErrInternal, err)
		return
	}

	for i := 0; i < len(resp.AssistantLogs); i++ {
		if err = resp.AssistantLogs[i].Valid(); err != nil {
			logger.FromContext(c).WithError(err).Errorf("error validating assistantlog data '%d'", resp.AssistantLogs[i].ID)
			response.Error(c, response.ErrAssistantlogsInvalidData, err)
			return
		}
	}

	response.Success(c, http.StatusOK, resp)
}

// GetFlowAssistantlogs is a function to return assistantlogs list by flow id
// @Summary Retrieve assistantlogs list by flow id
// @Tags Assistantlogs
// @Produce json
// @Security BearerAuth
// @Param flowID path int true "flow id" minimum(0)
// @Param request query rdb.TableQuery true "query table params"
// @Success 200 {object} response.successResp{data=assistantlogs} "assistantlogs list received successful"
// @Failure 400 {object} response.errorResp "invalid query request data"
// @Failure 403 {object} response.errorResp "getting assistantlogs not permitted"
// @Failure 500 {object} response.errorResp "internal error on getting assistantlogs"
// @Router /flows/{flowID}/assistantlogs/ [get]
func (s *AssistantlogService) GetFlowAssistantlogs(c *gin.Context) {
	var (
		err    error
		flowID uint64
		query  rdb.TableQuery
		resp   assistantlogs
	)

	if flowID, err = strconv.ParseUint(c.Param("flowID"), 10, 64); err != nil {
		logger.FromContext(c).WithError(err).Errorf("error parsing flow id")
		response.Error(c, response.ErrAssistantlogsInvalidRequest, err)
		return
	}

	if err = c.ShouldBindQuery(&query); err != nil {
		logger.FromContext(c).WithError(err).Errorf("error binding query")
		response.Error(c, response.ErrAssistantlogsInvalidRequest, err)
		return
	}

	uid := c.GetUint64("uid")
	privs := c.GetStringSlice("prm")
	var scope func(db *gorm.DB) *gorm.DB
	if slices.Contains(privs, "assistantlogs.admin") {
		scope = func(db *gorm.DB) *gorm.DB {
			return db.
				Joins("INNER JOIN flows f ON f.id = flow_id").
				Where("f.id = ?", flowID)
		}
	} else if slices.Contains(privs, "assistantlogs.view") {
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

	query.Init("assistantlogs", assistantlogsSQLMappers)

	if query.Group != "" {
		if _, ok := assistantlogsSQLMappers[query.Group]; !ok {
			logger.FromContext(c).Errorf("error finding assistantlogs grouped: group field not found")
			response.Error(c, response.ErrAssistantlogsInvalidRequest, errors.New("group field not found"))
			return
		}

		var respGrouped assistantlogsGrouped
		if respGrouped.Total, err = query.QueryGrouped(s.db, &respGrouped.Grouped, scope); err != nil {
			logger.FromContext(c).WithError(err).Errorf("error finding assistantlogs grouped")
			response.Error(c, response.ErrInternal, err)
			return
		}

		response.Success(c, http.StatusOK, respGrouped)
		return
	}

	if resp.Total, err = query.Query(s.db, &resp.AssistantLogs, scope); err != nil {
		logger.FromContext(c).WithError(err).Errorf("error finding assistantlogs")
		response.Error(c, response.ErrInternal, err)
		return
	}

	for i := 0; i < len(resp.AssistantLogs); i++ {
		if err = resp.AssistantLogs[i].Valid(); err != nil {
			logger.FromContext(c).WithError(err).Errorf("error validating assistantlog data '%d'", resp.AssistantLogs[i].ID)
			response.Error(c, response.ErrAssistantlogsInvalidData, err)
			return
		}
	}

	response.Success(c, http.StatusOK, resp)
}
