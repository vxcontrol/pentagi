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

type containers struct {
	Containers []models.Container `json:"containers"`
	Total      uint64             `json:"total"`
}

type containersGrouped struct {
	Grouped []string `json:"grouped"`
	Total   uint64   `json:"total"`
}

var containersSQLMappers = map[string]any{
	"id":         "{{table}}.id",
	"type":       "{{table}}.type",
	"name":       "{{table}}.name",
	"image":      "{{table}}.image",
	"status":     "{{table}}.status",
	"local_id":   "{{table}}.local_id",
	"local_dir":  "{{table}}.local_dir",
	"flow_id":    "{{table}}.flow_id",
	"created_at": "{{table}}.created_at",
	"updated_at": "{{table}}.updated_at",
	"data":       "({{table}}.type || ' ' || {{table}}.name || ' ' || {{table}}.status || ' ' || {{table}}.local_id || ' ' || {{table}}.local_dir)",
}

type ContainerService struct {
	db *gorm.DB
}

func NewContainerService(db *gorm.DB) *ContainerService {
	return &ContainerService{
		db: db,
	}
}

// GetContainers is a function to return containers list
// @Summary Retrieve containers list
// @Tags Containers
// @Produce json
// @Security BearerAuth
// @Param request query rdb.TableQuery true "query table params"
// @Success 200 {object} response.successResp{data=containers} "containers list received successful"
// @Failure 400 {object} response.errorResp "invalid query request data"
// @Failure 403 {object} response.errorResp "getting containers not permitted"
// @Failure 500 {object} response.errorResp "internal error on getting containers"
// @Router /containers/ [get]
func (s *ContainerService) GetContainers(c *gin.Context) {
	var (
		err   error
		query rdb.TableQuery
		resp  containers
	)

	if err = c.ShouldBindQuery(&query); err != nil {
		logger.FromContext(c).WithError(err).Errorf("error binding query")
		response.Error(c, response.ErrContainersInvalidRequest, err)
		return
	}

	uid := c.GetUint64("uid")
	privs := c.GetStringSlice("prm")
	var scope func(db *gorm.DB) *gorm.DB
	if slices.Contains(privs, "containers.admin") {
		scope = func(db *gorm.DB) *gorm.DB {
			return db.
				Joins("INNER JOIN flows f ON f.id = containers.flow_id")
		}
	} else if slices.Contains(privs, "containers.view") {
		scope = func(db *gorm.DB) *gorm.DB {
			return db.
				Joins("INNER JOIN flows f ON f.id = containers.flow_id").
				Where("f.user_id = ?", uid)
		}
	} else {
		logger.FromContext(c).Errorf("error filtering user role permissions: permission not found")
		response.Error(c, response.ErrNotPermitted, nil)
		return
	}

	query.Init("containers", containersSQLMappers)

	if query.Group != "" {
		if _, ok := containersSQLMappers[query.Group]; !ok {
			logger.FromContext(c).Errorf("error finding containers grouped: group field not found")
			response.Error(c, response.ErrContainersInvalidRequest, errors.New("group field not found"))
			return
		}

		var respGrouped containersGrouped
		if respGrouped.Total, err = query.QueryGrouped(s.db, &respGrouped.Grouped, scope); err != nil {
			logger.FromContext(c).WithError(err).Errorf("error finding containers grouped")
			response.Error(c, response.ErrInternal, err)
			return
		}

		response.Success(c, http.StatusOK, respGrouped)
		return
	}

	if resp.Total, err = query.Query(s.db, &resp.Containers, scope); err != nil {
		logger.FromContext(c).WithError(err).Errorf("error finding containers")
		response.Error(c, response.ErrInternal, err)
		return
	}

	for i := 0; i < len(resp.Containers); i++ {
		if err = resp.Containers[i].Valid(); err != nil {
			logger.FromContext(c).WithError(err).Errorf("error validating container data '%d'", resp.Containers[i].ID)
			response.Error(c, response.ErrContainersInvalidData, err)
			return
		}
	}

	response.Success(c, http.StatusOK, resp)
}

// GetFlowContainers is a function to return containers list by flow id
// @Summary Retrieve containers list by flow id
// @Tags Containers
// @Produce json
// @Security BearerAuth
// @Param flowID path int true "flow id" minimum(0)
// @Param request query rdb.TableQuery true "query table params"
// @Success 200 {object} response.successResp{data=containers} "containers list received successful"
// @Failure 400 {object} response.errorResp "invalid query request data"
// @Failure 403 {object} response.errorResp "getting containers not permitted"
// @Failure 500 {object} response.errorResp "internal error on getting containers"
// @Router /flows/{flowID}/containers/ [get]
func (s *ContainerService) GetFlowContainers(c *gin.Context) {
	var (
		err    error
		flowID uint64
		query  rdb.TableQuery
		resp   containers
	)

	if flowID, err = strconv.ParseUint(c.Param("flowID"), 10, 64); err != nil {
		logger.FromContext(c).WithError(err).Errorf("error parsing flow id")
		response.Error(c, response.ErrContainersInvalidRequest, err)
		return
	}

	if err = c.ShouldBindQuery(&query); err != nil {
		logger.FromContext(c).WithError(err).Errorf("error binding query")
		response.Error(c, response.ErrContainersInvalidRequest, err)
		return
	}

	uid := c.GetUint64("uid")
	privs := c.GetStringSlice("prm")
	var scope func(db *gorm.DB) *gorm.DB
	if slices.Contains(privs, "containers.admin") {
		scope = func(db *gorm.DB) *gorm.DB {
			return db.
				Joins("INNER JOIN flows f ON f.id = containers.flow_id").
				Where("f.id = ?", flowID)
		}
	} else if slices.Contains(privs, "containers.view") {
		scope = func(db *gorm.DB) *gorm.DB {
			return db.
				Joins("INNER JOIN flows f ON f.id = containers.flow_id").
				Where("f.id = ? AND f.user_id = ?", flowID, uid)
		}
	} else {
		logger.FromContext(c).Errorf("error filtering user role permissions: permission not found")
		response.Error(c, response.ErrNotPermitted, nil)
		return
	}

	query.Init("containers", containersSQLMappers)

	if query.Group != "" {
		if _, ok := containersSQLMappers[query.Group]; !ok {
			logger.FromContext(c).Errorf("error finding containers grouped: group field not found")
			response.Error(c, response.ErrContainersInvalidRequest, errors.New("group field not found"))
			return
		}

		var respGrouped containersGrouped
		if respGrouped.Total, err = query.QueryGrouped(s.db, &respGrouped.Grouped, scope); err != nil {
			logger.FromContext(c).WithError(err).Errorf("error finding containers grouped")
			response.Error(c, response.ErrInternal, err)
			return
		}

		response.Success(c, http.StatusOK, respGrouped)
		return
	}

	if resp.Total, err = query.Query(s.db, &resp.Containers, scope); err != nil {
		logger.FromContext(c).WithError(err).Errorf("error finding containers")
		response.Error(c, response.ErrInternal, err)
		return
	}

	for i := 0; i < len(resp.Containers); i++ {
		if err = resp.Containers[i].Valid(); err != nil {
			logger.FromContext(c).WithError(err).Errorf("error validating container data '%d'", resp.Containers[i].ID)
			response.Error(c, response.ErrContainersInvalidData, err)
			return
		}
	}

	response.Success(c, http.StatusOK, resp)
}

// GetFlowContainer is a function to return container info by id and flow id
// @Summary Retrieve container info by id and flow id
// @Tags Containers
// @Produce json
// @Security BearerAuth
// @Param flowID path int true "flow id" minimum(0)
// @Param containerID path int true "container id" minimum(0)
// @Success 200 {object} response.successResp{data=models.Container} "container info received successful"
// @Failure 403 {object} response.errorResp "getting container not permitted"
// @Failure 404 {object} response.errorResp "container not found"
// @Failure 500 {object} response.errorResp "internal error on getting container"
// @Router /flows/{flowID}/containers/{containerID} [get]
func (s *ContainerService) GetFlowContainer(c *gin.Context) {
	var (
		err         error
		containerID uint64
		flowID      uint64
		resp        models.Container
	)

	if flowID, err = strconv.ParseUint(c.Param("flowID"), 10, 64); err != nil {
		logger.FromContext(c).WithError(err).Errorf("error parsing flow id")
		response.Error(c, response.ErrContainersInvalidRequest, err)
		return
	}
	if containerID, err = strconv.ParseUint(c.Param("containerID"), 10, 64); err != nil {
		logger.FromContext(c).WithError(err).Errorf("error parsing container id")
		response.Error(c, response.ErrContainersInvalidRequest, err)
		return
	}

	uid := c.GetUint64("uid")
	privs := c.GetStringSlice("prm")
	var scope func(db *gorm.DB) *gorm.DB
	if slices.Contains(privs, "containers.admin") {
		scope = func(db *gorm.DB) *gorm.DB {
			return db.Where("f.id = ?", flowID)
		}
	} else if slices.Contains(privs, "containers.view") {
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
		Where("containers.id = ?", containerID).
		Take(&resp).Error
	if err != nil {
		logger.FromContext(c).WithError(err).Errorf("error on getting container by id")
		if gorm.IsRecordNotFoundError(err) {
			response.Error(c, response.ErrContainersNotFound, err)
		} else {
			response.Error(c, response.ErrInternal, err)
		}
		return
	}

	response.Success(c, http.StatusOK, resp)
}
