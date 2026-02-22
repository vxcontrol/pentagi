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

type vecstorelogs struct {
	VecstoreLogs []models.Vecstorelog `json:"vecstorelogs"`
	Total        uint64               `json:"total"`
}

type vecstorelogsGrouped struct {
	Grouped []string `json:"grouped"`
	Total   uint64   `json:"total"`
}

var vecstorelogsSQLMappers = map[string]any{
	"id":         "{{table}}.id",
	"initiator":  "{{table}}.initiator",
	"executor":   "{{table}}.executor",
	"filter":     "{{table}}.filter",
	"query":      "{{table}}.query",
	"action":     "{{table}}.action",
	"result":     "{{table}}.result",
	"flow_id":    "{{table}}.flow_id",
	"task_id":    "{{table}}.task_id",
	"subtask_id": "{{table}}.subtask_id",
	"created_at": "{{table}}.created_at",
	"data":       "({{table}}.filter || ' ' || {{table}}.query || ' ' || {{table}}.result)",
}

type VecstorelogService struct {
	db *gorm.DB
}

func NewVecstorelogService(db *gorm.DB) *VecstorelogService {
	return &VecstorelogService{
		db: db,
	}
}

// GetVecstorelogs is a function to return vecstorelogs list
// @Summary Retrieve vecstorelogs list
// @Tags Vecstorelogs
// @Produce json
// @Security BearerAuth
// @Param request query rdb.TableQuery true "query table params"
// @Success 200 {object} response.successResp{data=vecstorelogs} "vecstorelogs list received successful"
// @Failure 400 {object} response.errorResp "invalid query request data"
// @Failure 403 {object} response.errorResp "getting vecstorelogs not permitted"
// @Failure 500 {object} response.errorResp "internal error on getting vecstorelogs"
// @Router /vecstorelogs/ [get]
func (s *VecstorelogService) GetVecstorelogs(c *gin.Context) {
	var (
		err   error
		query rdb.TableQuery
		resp  vecstorelogs
	)

	if err = c.ShouldBindQuery(&query); err != nil {
		logger.FromContext(c).WithError(err).Errorf("error binding query")
		response.Error(c, response.ErrVecstorelogsInvalidRequest, err)
		return
	}

	uid := c.GetUint64("uid")
	privs := c.GetStringSlice("prm")
	var scope func(db *gorm.DB) *gorm.DB
	if slices.Contains(privs, "vecstorelogs.admin") {
		scope = func(db *gorm.DB) *gorm.DB {
			return db.
				Joins("INNER JOIN flows f ON f.id = flow_id")
		}
	} else if slices.Contains(privs, "vecstorelogs.view") {
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

	query.Init("vecstorelogs", vecstorelogsSQLMappers)

	if query.Group != "" {
		if _, ok := vecstorelogsSQLMappers[query.Group]; !ok {
			logger.FromContext(c).Errorf("error finding vecstorelogs grouped: group field not found")
			response.Error(c, response.ErrVecstorelogsInvalidRequest, errors.New("group field not found"))
			return
		}

		var respGrouped vecstorelogsGrouped
		if respGrouped.Total, err = query.QueryGrouped(s.db, &respGrouped.Grouped, scope); err != nil {
			logger.FromContext(c).WithError(err).Errorf("error finding vecstorelogs grouped")
			response.Error(c, response.ErrInternal, err)
			return
		}

		response.Success(c, http.StatusOK, respGrouped)
		return
	}

	if resp.Total, err = query.Query(s.db, &resp.VecstoreLogs, scope); err != nil {
		logger.FromContext(c).WithError(err).Errorf("error finding vecstorelogs")
		response.Error(c, response.ErrInternal, err)
		return
	}

	for i := 0; i < len(resp.VecstoreLogs); i++ {
		if err = resp.VecstoreLogs[i].Valid(); err != nil {
			logger.FromContext(c).WithError(err).Errorf("error validating vecstorelog data '%d'", resp.VecstoreLogs[i].ID)
			response.Error(c, response.ErrVecstorelogsInvalidData, err)
			return
		}
	}

	response.Success(c, http.StatusOK, resp)
}

// GetFlowVecstorelogs is a function to return vecstorelogs list by flow id
// @Summary Retrieve vecstorelogs list by flow id
// @Tags Vecstorelogs
// @Produce json
// @Security BearerAuth
// @Param flowID path int true "flow id" minimum(0)
// @Param request query rdb.TableQuery true "query table params"
// @Success 200 {object} response.successResp{data=vecstorelogs} "vecstorelogs list received successful"
// @Failure 400 {object} response.errorResp "invalid query request data"
// @Failure 403 {object} response.errorResp "getting vecstorelogs not permitted"
// @Failure 500 {object} response.errorResp "internal error on getting vecstorelogs"
// @Router /flows/{flowID}/vecstorelogs/ [get]
func (s *VecstorelogService) GetFlowVecstorelogs(c *gin.Context) {
	var (
		err    error
		flowID uint64
		query  rdb.TableQuery
		resp   vecstorelogs
	)

	if flowID, err = strconv.ParseUint(c.Param("flowID"), 10, 64); err != nil {
		logger.FromContext(c).WithError(err).Errorf("error parsing flow id")
		response.Error(c, response.ErrVecstorelogsInvalidRequest, err)
		return
	}

	if err = c.ShouldBindQuery(&query); err != nil {
		logger.FromContext(c).WithError(err).Errorf("error binding query")
		response.Error(c, response.ErrVecstorelogsInvalidRequest, err)
		return
	}

	uid := c.GetUint64("uid")
	privs := c.GetStringSlice("prm")
	var scope func(db *gorm.DB) *gorm.DB
	if slices.Contains(privs, "vecstorelogs.admin") {
		scope = func(db *gorm.DB) *gorm.DB {
			return db.
				Joins("INNER JOIN flows f ON f.id = flow_id").
				Where("f.id = ?", flowID)
		}
	} else if slices.Contains(privs, "vecstorelogs.view") {
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

	query.Init("vecstorelogs", vecstorelogsSQLMappers)

	if query.Group != "" {
		if _, ok := vecstorelogsSQLMappers[query.Group]; !ok {
			logger.FromContext(c).Errorf("error finding vecstorelogs grouped: group field not found")
			response.Error(c, response.ErrVecstorelogsInvalidRequest, errors.New("group field not found"))
			return
		}

		var respGrouped vecstorelogsGrouped
		if respGrouped.Total, err = query.QueryGrouped(s.db, &respGrouped.Grouped, scope); err != nil {
			logger.FromContext(c).WithError(err).Errorf("error finding vecstorelogs grouped")
			response.Error(c, response.ErrInternal, err)
			return
		}

		response.Success(c, http.StatusOK, respGrouped)
		return
	}

	if resp.Total, err = query.Query(s.db, &resp.VecstoreLogs, scope); err != nil {
		logger.FromContext(c).WithError(err).Errorf("error finding vecstorelogs")
		response.Error(c, response.ErrInternal, err)
		return
	}

	for i := 0; i < len(resp.VecstoreLogs); i++ {
		if err = resp.VecstoreLogs[i].Valid(); err != nil {
			logger.FromContext(c).WithError(err).Errorf("error validating vecstorelog data '%d'", resp.VecstoreLogs[i].ID)
			response.Error(c, response.ErrVecstorelogsInvalidData, err)
			return
		}
	}

	response.Success(c, http.StatusOK, resp)
}
