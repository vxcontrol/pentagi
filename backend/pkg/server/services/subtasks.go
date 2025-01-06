package services

import (
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

type subtasks struct {
	Subtasks []models.Subtask `json:"subtasks"`
	Total    uint64           `json:"total"`
}

type subtasksGrouped struct {
	Grouped []string `json:"grouped"`
	Total   uint64   `json:"total"`
}

var subtasksSQLMappers = map[string]interface{}{
	"id":          "{{table}}.id",
	"status":      "{{table}}.status",
	"title":       "{{table}}.title",
	"description": "{{table}}.description",
	"result":      "{{table}}.result",
	"task_id":     "{{table}}.task_id",
	"data":        "({{table}}.status || ' ' || {{table}}.title || ' ' || {{table}}.description || ' ' || {{table}}.result)",
}

type SubtaskService struct {
	db *gorm.DB
}

func NewSubtaskService(db *gorm.DB) *SubtaskService {
	return &SubtaskService{
		db: db,
	}
}

// GetFlowSubtasks is a function to return flow subtasks list
// @Summary Retrieve flow subtasks list
// @Tags Subtasks
// @Produce json
// @Param request query rdb.TableQuery true "query table params"
// @Success 200 {object} response.successResp{data=subtasks} "flow subtasks list received successful"
// @Failure 400 {object} response.errorResp "invalid query request data"
// @Failure 403 {object} response.errorResp "getting flow subtasks not permitted"
// @Failure 500 {object} response.errorResp "internal error on getting flow subtasks"
// @Router /flows/{flowID}/subtasks/ [get]
func (s *SubtaskService) GetFlowSubtasks(c *gin.Context) {
	var (
		err    error
		flowID uint64
		query  rdb.TableQuery
		resp   subtasks
	)

	if err = c.ShouldBindQuery(&query); err != nil {
		logger.FromContext(c).WithError(err).Errorf("error binding query")
		response.Error(c, response.ErrSubtasksInvalidRequest, err)
		return
	}

	if flowID, err = strconv.ParseUint(c.Param("flowID"), 10, 64); err != nil {
		logger.FromContext(c).WithError(err).Errorf("error parsing flow id")
		response.Error(c, response.ErrSubtasksInvalidRequest, err)
		return
	}

	uid := c.GetUint64("uid")
	privs := c.GetStringSlice("prm")
	var scope func(db *gorm.DB) *gorm.DB
	if slices.Contains(privs, "subtasks.admin") {
		scope = func(db *gorm.DB) *gorm.DB {
			return db.
				Joins("INNER JOIN tasks t ON t.id = task_id").
				Joins("INNER JOIN flows f ON f.id = t.flow_id").
				Where("f.id = ?", flowID)
		}
	} else if slices.Contains(privs, "subtasks.view") {
		scope = func(db *gorm.DB) *gorm.DB {
			return db.
				Joins("INNER JOIN tasks t ON t.id = task_id").
				Joins("INNER JOIN flows f ON f.id = t.flow_id").
				Where("f.id = ? AND f.user_id = ?", flowID, uid)
		}
	} else {
		logger.FromContext(c).Errorf("error filtering user role permissions: permission not found")
		response.Error(c, response.ErrNotPermitted, nil)
		return
	}

	query.Init("subtasks", subtasksSQLMappers)

	if query.Group != "" {
		var respGrouped subtasksGrouped
		if respGrouped.Total, err = query.QueryGrouped(s.db, &respGrouped.Grouped, scope); err != nil {
			logger.FromContext(c).WithError(err).Errorf("error finding subtasks grouped")
			response.Error(c, response.ErrInternal, err)
			return
		}

		response.Success(c, http.StatusOK, respGrouped)
		return
	}

	if resp.Total, err = query.Query(s.db, &resp.Subtasks, scope); err != nil {
		logger.FromContext(c).WithError(err).Errorf("error finding subtasks")
		response.Error(c, response.ErrInternal, err)
		return
	}

	for i := 0; i < len(resp.Subtasks); i++ {
		if err = resp.Subtasks[i].Valid(); err != nil {
			logger.FromContext(c).WithError(err).Errorf("error validating subtask data '%d'", resp.Subtasks[i].ID)
			response.Error(c, response.ErrSubtasksInvalidData, err)
			return
		}
	}

	response.Success(c, http.StatusOK, resp)
}

// GetFlowTaskSubtasks is a function to return flow task subtasks list
// @Summary Retrieve flow task subtasks list
// @Tags Subtasks
// @Produce json
// @Param request query rdb.TableQuery true "query table params"
// @Success 200 {object} response.successResp{data=subtasks} "flow task subtasks list received successful"
// @Failure 400 {object} response.errorResp "invalid query request data"
// @Failure 403 {object} response.errorResp "getting flow task subtasks not permitted"
// @Failure 500 {object} response.errorResp "internal error on getting flow subtasks"
// @Router /flows/{flowID}/tasks/{taskID}/subtasks/ [get]
func (s *SubtaskService) GetFlowTaskSubtasks(c *gin.Context) {
	var (
		err    error
		flowID uint64
		taskID uint64
		query  rdb.TableQuery
		resp   subtasks
	)

	if err = c.ShouldBindQuery(&query); err != nil {
		logger.FromContext(c).WithError(err).Errorf("error binding query")
		response.Error(c, response.ErrSubtasksInvalidRequest, err)
		return
	}

	if flowID, err = strconv.ParseUint(c.Param("flowID"), 10, 64); err != nil {
		logger.FromContext(c).WithError(err).Errorf("error parsing flow id")
		response.Error(c, response.ErrSubtasksInvalidRequest, err)
		return
	}

	if taskID, err = strconv.ParseUint(c.Param("taskID"), 10, 64); err != nil {
		logger.FromContext(c).WithError(err).Errorf("error parsing task id")
		response.Error(c, response.ErrSubtasksInvalidRequest, err)
		return
	}

	uid := c.GetUint64("uid")
	privs := c.GetStringSlice("prm")
	var scope func(db *gorm.DB) *gorm.DB
	if slices.Contains(privs, "subtasks.admin") {
		scope = func(db *gorm.DB) *gorm.DB {
			return db.
				Joins("INNER JOIN tasks t ON t.id = task_id").
				Joins("INNER JOIN flows f ON f.id = t.flow_id").
				Where("f.id = ? AND t.id = ?", flowID, taskID)
		}
	} else if slices.Contains(privs, "subtasks.view") {
		scope = func(db *gorm.DB) *gorm.DB {
			return db.
				Joins("INNER JOIN tasks t ON t.id = task_id").
				Joins("INNER JOIN flows f ON f.id = t.flow_id").
				Where("f.id = ? AND f.user_id = ? AND t.id = ?", flowID, uid, taskID)
		}
	} else {
		logger.FromContext(c).Errorf("error filtering user role permissions: permission not found")
		response.Error(c, response.ErrNotPermitted, nil)
		return
	}

	query.Init("subtasks", subtasksSQLMappers)

	if query.Group != "" {
		var respGrouped subtasksGrouped
		if respGrouped.Total, err = query.QueryGrouped(s.db, &respGrouped.Grouped, scope); err != nil {
			logger.FromContext(c).WithError(err).Errorf("error finding subtasks grouped")
			response.Error(c, response.ErrInternal, err)
			return
		}

		response.Success(c, http.StatusOK, respGrouped)
		return
	}

	if resp.Total, err = query.Query(s.db, &resp.Subtasks, scope); err != nil {
		logger.FromContext(c).WithError(err).Errorf("error finding subtasks")
		response.Error(c, response.ErrInternal, err)
		return
	}

	for i := 0; i < len(resp.Subtasks); i++ {
		if err = resp.Subtasks[i].Valid(); err != nil {
			logger.FromContext(c).WithError(err).Errorf("error validating subtask data '%d'", resp.Subtasks[i].ID)
			response.Error(c, response.ErrSubtasksInvalidData, err)
			return
		}
	}

	response.Success(c, http.StatusOK, resp)
}

// GetFlowTaskSubtask is a function to return flow task subtask by id
// @Summary Retrieve flow task subtask by id
// @Tags Subtasks
// @Produce json
// @Param flowID path int true "flow id" minimum(0)
// @Param taskID path int true "task id" minimum(0)
// @Param subtaskID path int true "subtask id" minimum(0)
// @Success 200 {object} response.successResp{data=models.Subtask} "flow task subtask received successful"
// @Failure 403 {object} response.errorResp "getting flow task subtask not permitted"
// @Failure 404 {object} response.errorResp "flow task subtask not found"
// @Failure 500 {object} response.errorResp "internal error on getting flow task subtask"
// @Router /flows/{flowID}/tasks/{taskID}/subtasks/{subtaskID} [get]
func (s *SubtaskService) GetFlowTaskSubtask(c *gin.Context) {
	var (
		err       error
		flowID    uint64
		taskID    uint64
		subtaskID uint64
		resp      models.Task
	)

	if flowID, err = strconv.ParseUint(c.Param("flowID"), 10, 64); err != nil {
		logger.FromContext(c).WithError(err).Errorf("error parsing flow id")
		response.Error(c, response.ErrSubtasksInvalidRequest, err)
		return
	}

	if taskID, err = strconv.ParseUint(c.Param("taskID"), 10, 64); err != nil {
		logger.FromContext(c).WithError(err).Errorf("error parsing task id")
		response.Error(c, response.ErrSubtasksInvalidRequest, err)
		return
	}

	if subtaskID, err = strconv.ParseUint(c.Param("subtaskID"), 10, 64); err != nil {
		logger.FromContext(c).WithError(err).Errorf("error parsing subtask id")
		response.Error(c, response.ErrSubtasksInvalidRequest, err)
		return
	}

	uid := c.GetUint64("uid")
	privs := c.GetStringSlice("prm")
	var scope func(db *gorm.DB) *gorm.DB
	if slices.Contains(privs, "subtasks.admin") {
		scope = func(db *gorm.DB) *gorm.DB {
			return db.
				Joins("INNER JOIN tasks t ON t.id = task_id").
				Joins("INNER JOIN flows f ON f.id = t.flow_id").
				Where("f.id = ? AND t.id = ?", flowID, taskID)
		}
	} else if slices.Contains(privs, "subtasks.view") {
		scope = func(db *gorm.DB) *gorm.DB {
			return db.
				Joins("INNER JOIN tasks t ON t.id = task_id").
				Joins("INNER JOIN flows f ON f.id = t.flow_id").
				Where("f.id = ? AND f.user_id = ? AND t.id = ?", flowID, uid, taskID)
		}
	} else {
		logger.FromContext(c).Errorf("error filtering user role permissions: permission not found")
		response.Error(c, response.ErrNotPermitted, nil)
		return
	}

	err = s.db.Model(&resp).
		Scopes(scope).
		Where("id = ?", subtaskID).
		Take(&resp).Error
	if err != nil {
		logger.FromContext(c).WithError(err).Errorf("error on getting flow task subtask by id")
		if gorm.IsRecordNotFoundError(err) {
			response.Error(c, response.ErrSubtasksNotFound, err)
		} else {
			response.Error(c, response.ErrInternal, err)
		}
		return
	}

	response.Success(c, http.StatusOK, resp)
}
