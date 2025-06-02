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

type tasks struct {
	Tasks []models.Task `json:"tasks"`
	Total uint64        `json:"total"`
}

type tasksGrouped struct {
	Grouped []string `json:"grouped"`
	Total   uint64   `json:"total"`
}

var tasksSQLMappers = map[string]interface{}{
	"id":      "{{table}}.id",
	"status":  "{{table}}.status",
	"title":   "{{table}}.title",
	"input":   "{{table}}.input",
	"result":  "{{table}}.result",
	"flow_id": "{{table}}.flow_id",
	"data":    "({{table}}.status || ' ' || {{table}}.title || ' ' || {{table}}.input || ' ' || {{table}}.result)",
}

type TaskService struct {
	db *gorm.DB
}

func NewTaskService(db *gorm.DB) *TaskService {
	return &TaskService{
		db: db,
	}
}

// GetFlowTasks is a function to return flow tasks list
// @Summary Retrieve flow tasks list
// @Tags Tasks
// @Produce json
// @Param flowID path int true "flow id" minimum(0)
// @Param request query rdb.TableQuery true "query table params"
// @Success 200 {object} response.successResp{data=tasks} "flow tasks list received successful"
// @Failure 400 {object} response.errorResp "invalid query request data"
// @Failure 403 {object} response.errorResp "getting flow tasks not permitted"
// @Failure 500 {object} response.errorResp "internal error on getting flow tasks"
// @Router /flows/{flowID}/tasks/ [get]
func (s *TaskService) GetFlowTasks(c *gin.Context) {
	var (
		err    error
		flowID uint64
		query  rdb.TableQuery
		resp   tasks
	)

	if err = c.ShouldBindQuery(&query); err != nil {
		logger.FromContext(c).WithError(err).Errorf("error binding query")
		response.Error(c, response.ErrTasksInvalidRequest, err)
		return
	}

	if flowID, err = strconv.ParseUint(c.Param("flowID"), 10, 64); err != nil {
		logger.FromContext(c).WithError(err).Errorf("error parsing flow id")
		response.Error(c, response.ErrTasksInvalidRequest, err)
		return
	}

	uid := c.GetUint64("uid")
	privs := c.GetStringSlice("prm")
	var scope func(db *gorm.DB) *gorm.DB
	if slices.Contains(privs, "tasks.admin") {
		scope = func(db *gorm.DB) *gorm.DB {
			return db.
				Joins("INNER JOIN flows f ON f.id = flow_id").
				Where("f.id = ?", flowID)
		}
	} else if slices.Contains(privs, "tasks.view") {
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

	query.Init("tasks", tasksSQLMappers)

	if query.Group != "" {
		var respGrouped tasksGrouped
		if respGrouped.Total, err = query.QueryGrouped(s.db, &respGrouped.Grouped, scope); err != nil {
			logger.FromContext(c).WithError(err).Errorf("error finding tasks grouped")
			response.Error(c, response.ErrInternal, err)
			return
		}

		response.Success(c, http.StatusOK, respGrouped)
		return
	}

	if resp.Total, err = query.Query(s.db, &resp.Tasks, scope); err != nil {
		logger.FromContext(c).WithError(err).Errorf("error finding tasks")
		response.Error(c, response.ErrInternal, err)
		return
	}

	for i := 0; i < len(resp.Tasks); i++ {
		if err = resp.Tasks[i].Valid(); err != nil {
			logger.FromContext(c).WithError(err).Errorf("error validating task data '%d'", resp.Tasks[i].ID)
			response.Error(c, response.ErrTasksInvalidData, err)
			return
		}
	}

	response.Success(c, http.StatusOK, resp)
}

// GetFlowTask is a function to return flow task by id
// @Summary Retrieve flow task by id
// @Tags Tasks
// @Produce json
// @Param flowID path int true "flow id" minimum(0)
// @Param taskID path int true "task id" minimum(0)
// @Success 200 {object} response.successResp{data=models.Task} "flow task received successful"
// @Failure 403 {object} response.errorResp "getting flow task not permitted"
// @Failure 404 {object} response.errorResp "flow task not found"
// @Failure 500 {object} response.errorResp "internal error on getting flow task"
// @Router /flows/{flowID}/tasks/{taskID} [get]
func (s *TaskService) GetFlowTask(c *gin.Context) {
	var (
		err    error
		flowID uint64
		taskID uint64
		resp   models.Task
	)

	if flowID, err = strconv.ParseUint(c.Param("flowID"), 10, 64); err != nil {
		logger.FromContext(c).WithError(err).Errorf("error parsing flow id")
		response.Error(c, response.ErrTasksInvalidRequest, err)
		return
	}

	if taskID, err = strconv.ParseUint(c.Param("taskID"), 10, 64); err != nil {
		logger.FromContext(c).WithError(err).Errorf("error parsing task id")
		response.Error(c, response.ErrTasksInvalidRequest, err)
		return
	}

	uid := c.GetUint64("uid")
	privs := c.GetStringSlice("prm")
	var scope func(db *gorm.DB) *gorm.DB
	if slices.Contains(privs, "tasks.admin") {
		scope = func(db *gorm.DB) *gorm.DB {
			return db.
				Joins("INNER JOIN flows f ON f.id = flow_id").
				Where("f.id = ?", flowID)
		}
	} else if slices.Contains(privs, "tasks.view") {
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

	err = s.db.Model(&resp).
		Scopes(scope).
		Where("id = ?", taskID).
		Take(&resp).Error
	if err != nil {
		logger.FromContext(c).WithError(err).Errorf("error on getting flow task by id")
		if gorm.IsRecordNotFoundError(err) {
			response.Error(c, response.ErrTasksNotFound, err)
		} else {
			response.Error(c, response.ErrInternal, err)
		}
		return
	}

	response.Success(c, http.StatusOK, resp)
}

// GetFlowTaskGraph is a function to return flow task graph by id
// @Summary Retrieve flow task graph by id
// @Tags Tasks
// @Produce json
// @Param flowID path int true "flow id" minimum(0)
// @Param taskID path int true "task id" minimum(0)
// @Success 200 {object} response.successResp{data=models.FlowTasksSubtasks} "flow task graph received successful"
// @Failure 403 {object} response.errorResp "getting flow task graph not permitted"
// @Failure 404 {object} response.errorResp "flow task graph not found"
// @Failure 500 {object} response.errorResp "internal error on getting flow task graph"
// @Router /flows/{flowID}/tasks/{taskID}/graph [get]
func (s *TaskService) GetFlowTaskGraph(c *gin.Context) {
	var (
		err    error
		flow   models.Flow
		flowID uint64
		taskID uint64
		resp   models.TaskSubtasks
	)

	if flowID, err = strconv.ParseUint(c.Param("flowID"), 10, 64); err != nil {
		logger.FromContext(c).WithError(err).Errorf("error parsing flow id")
		response.Error(c, response.ErrTasksInvalidRequest, err)
		return
	}

	if taskID, err = strconv.ParseUint(c.Param("taskID"), 10, 64); err != nil {
		logger.FromContext(c).WithError(err).Errorf("error parsing task id")
		response.Error(c, response.ErrTasksInvalidRequest, err)
		return
	}

	uid := c.GetUint64("uid")
	privs := c.GetStringSlice("prm")
	var scope func(db *gorm.DB) *gorm.DB
	if slices.Contains(privs, "tasks.admin") {
		scope = func(db *gorm.DB) *gorm.DB {
			return db.Where("f.id = ?", flowID)
		}
	} else if slices.Contains(privs, "tasks.view") {
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
		Where("id = ?", taskID).
		Take(&resp).Related(&flow).Error
	if err != nil {
		logger.FromContext(c).WithError(err).Errorf("error on getting flow task by id")
		if gorm.IsRecordNotFoundError(err) {
			response.Error(c, response.ErrTasksNotFound, err)
		} else {
			response.Error(c, response.ErrInternal, err)
		}
		return
	}

	isSubtasksAdmin := slices.Contains(privs, "subtasks.admin")
	isSubtasksView := slices.Contains(privs, "subtasks.view")
	if !(flow.UserID == uid && isSubtasksView) && !(flow.UserID != uid && isSubtasksAdmin) {
		response.Success(c, http.StatusOK, resp)
		return
	}

	err = s.db.Model(&resp).Association("subtasks").Find(&resp.Subtasks).Error
	if err != nil {
		logger.FromContext(c).WithError(err).Errorf("error on getting task subtasks")
		response.Error(c, response.ErrInternal, err)
		return
	}

	if err = resp.Valid(); err != nil {
		logger.FromContext(c).WithError(err).Errorf("error validating task data '%d'", taskID)
		response.Error(c, response.ErrTasksInvalidData, err)
		return
	}

	response.Success(c, http.StatusOK, resp)
}
