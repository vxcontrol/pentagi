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

type msglogs struct {
	MsgLogs []models.Msglog `json:"msglogs"`
	Total   uint64          `json:"total"`
}

type msglogsGrouped struct {
	Grouped []string `json:"grouped"`
	Total   uint64   `json:"total"`
}

var msglogsSQLMappers = map[string]any{
	"id":            "{{table}}.id",
	"type":          "{{table}}.type",
	"message":       "{{table}}.message",
	"thinking":      "{{table}}.thinking",
	"result":        "{{table}}.result",
	"result_format": "{{table}}.result_format",
	"flow_id":       "{{table}}.flow_id",
	"task_id":       "{{table}}.task_id",
	"subtask_id":    "{{table}}.subtask_id",
	"created_at":    "{{table}}.created_at",
	"data":          "({{table}}.type || ' ' || {{table}}.message || ' ' || {{table}}.thinking || ' ' || {{table}}.result)",
}

type MsglogService struct {
	db *gorm.DB
}

func NewMsglogService(db *gorm.DB) *MsglogService {
	return &MsglogService{
		db: db,
	}
}

// GetMsglogs is a function to return msglogs list
// @Summary Retrieve msglogs list
// @Tags Msglogs
// @Produce json
// @Security BearerAuth
// @Param request query rdb.TableQuery true "query table params"
// @Success 200 {object} response.successResp{data=msglogs} "msglogs list received successful"
// @Failure 400 {object} response.errorResp "invalid query request data"
// @Failure 403 {object} response.errorResp "getting msglogs not permitted"
// @Failure 500 {object} response.errorResp "internal error on getting msglogs"
// @Router /msglogs/ [get]
func (s *MsglogService) GetMsglogs(c *gin.Context) {
	var (
		err   error
		query rdb.TableQuery
		resp  msglogs
	)

	if err = c.ShouldBindQuery(&query); err != nil {
		logger.FromContext(c).WithError(err).Errorf("error binding query")
		response.Error(c, response.ErrMsglogsInvalidRequest, err)
		return
	}

	uid := c.GetUint64("uid")
	privs := c.GetStringSlice("prm")
	var scope func(db *gorm.DB) *gorm.DB
	if slices.Contains(privs, "msglogs.admin") {
		scope = func(db *gorm.DB) *gorm.DB {
			return db.
				Joins("INNER JOIN flows f ON f.id = msglogs.flow_id")
		}
	} else if slices.Contains(privs, "msglogs.view") {
		scope = func(db *gorm.DB) *gorm.DB {
			return db.
				Joins("INNER JOIN flows f ON f.id = msglogs.flow_id").
				Where("f.user_id = ?", uid)
		}
	} else {
		logger.FromContext(c).Errorf("error filtering user role permissions: permission not found")
		response.Error(c, response.ErrNotPermitted, nil)
		return
	}

	query.Init("msglogs", msglogsSQLMappers)

	if query.Group != "" {
		if _, ok := msglogsSQLMappers[query.Group]; !ok {
			logger.FromContext(c).Errorf("error finding msglogs grouped: group field not found")
			response.Error(c, response.ErrMsglogsInvalidRequest, errors.New("group field not found"))
			return
		}

		var respGrouped msglogsGrouped
		if respGrouped.Total, err = query.QueryGrouped(s.db, &respGrouped.Grouped, scope); err != nil {
			logger.FromContext(c).WithError(err).Errorf("error finding msglogs grouped")
			response.Error(c, response.ErrInternal, err)
			return
		}

		response.Success(c, http.StatusOK, respGrouped)
		return
	}

	if resp.Total, err = query.Query(s.db, &resp.MsgLogs, scope); err != nil {
		logger.FromContext(c).WithError(err).Errorf("error finding msglogs")
		response.Error(c, response.ErrInternal, err)
		return
	}

	for i := 0; i < len(resp.MsgLogs); i++ {
		if err = resp.MsgLogs[i].Valid(); err != nil {
			logger.FromContext(c).WithError(err).Errorf("error validating msglog data '%d'", resp.MsgLogs[i].ID)
			response.Error(c, response.ErrMsglogsInvalidData, err)
			return
		}
	}

	response.Success(c, http.StatusOK, resp)
}

// GetFlowMsglogs is a function to return msglogs list by flow id
// @Summary Retrieve msglogs list by flow id
// @Tags Msglogs
// @Produce json
// @Security BearerAuth
// @Param flowID path int true "flow id" minimum(0)
// @Param request query rdb.TableQuery true "query table params"
// @Success 200 {object} response.successResp{data=msglogs} "msglogs list received successful"
// @Failure 400 {object} response.errorResp "invalid query request data"
// @Failure 403 {object} response.errorResp "getting msglogs not permitted"
// @Failure 500 {object} response.errorResp "internal error on getting msglogs"
// @Router /flows/{flowID}/msglogs/ [get]
func (s *MsglogService) GetFlowMsglogs(c *gin.Context) {
	var (
		err    error
		flowID uint64
		query  rdb.TableQuery
		resp   msglogs
	)

	if flowID, err = strconv.ParseUint(c.Param("flowID"), 10, 64); err != nil {
		logger.FromContext(c).WithError(err).Errorf("error parsing flow id")
		response.Error(c, response.ErrMsglogsInvalidRequest, err)
		return
	}

	if err = c.ShouldBindQuery(&query); err != nil {
		logger.FromContext(c).WithError(err).Errorf("error binding query")
		response.Error(c, response.ErrMsglogsInvalidRequest, err)
		return
	}

	uid := c.GetUint64("uid")
	privs := c.GetStringSlice("prm")
	var scope func(db *gorm.DB) *gorm.DB
	if slices.Contains(privs, "msglogs.admin") {
		scope = func(db *gorm.DB) *gorm.DB {
			return db.
				Joins("INNER JOIN flows f ON f.id = msglogs.flow_id").
				Where("f.id = ?", flowID)
		}
	} else if slices.Contains(privs, "msglogs.view") {
		scope = func(db *gorm.DB) *gorm.DB {
			return db.
				Joins("INNER JOIN flows f ON f.id = msglogs.flow_id").
				Where("f.id = ? AND f.user_id = ?", flowID, uid)
		}
	} else {
		logger.FromContext(c).Errorf("error filtering user role permissions: permission not found")
		response.Error(c, response.ErrNotPermitted, nil)
		return
	}

	query.Init("msglogs", msglogsSQLMappers)

	if query.Group != "" {
		if _, ok := msglogsSQLMappers[query.Group]; !ok {
			logger.FromContext(c).Errorf("error finding msglogs grouped: group field not found")
			response.Error(c, response.ErrMsglogsInvalidRequest, errors.New("group field not found"))
			return
		}

		var respGrouped msglogsGrouped
		if respGrouped.Total, err = query.QueryGrouped(s.db, &respGrouped.Grouped, scope); err != nil {
			logger.FromContext(c).WithError(err).Errorf("error finding msglogs grouped")
			response.Error(c, response.ErrInternal, err)
			return
		}

		response.Success(c, http.StatusOK, respGrouped)
		return
	}

	if resp.Total, err = query.Query(s.db, &resp.MsgLogs, scope); err != nil {
		logger.FromContext(c).WithError(err).Errorf("error finding msglogs")
		response.Error(c, response.ErrInternal, err)
		return
	}

	for i := 0; i < len(resp.MsgLogs); i++ {
		if err = resp.MsgLogs[i].Valid(); err != nil {
			logger.FromContext(c).WithError(err).Errorf("error validating msglog data '%d'", resp.MsgLogs[i].ID)
			response.Error(c, response.ErrMsglogsInvalidData, err)
			return
		}
	}

	response.Success(c, http.StatusOK, resp)
}
