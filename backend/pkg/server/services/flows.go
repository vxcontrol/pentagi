package services

import (
	"errors"
	"net/http"
	"slices"
	"strconv"

	"pentagi/pkg/controller"
	"pentagi/pkg/providers"
	"pentagi/pkg/providers/provider"
	"pentagi/pkg/server/logger"
	"pentagi/pkg/server/models"
	"pentagi/pkg/server/rdb"
	"pentagi/pkg/server/response"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

type flows struct {
	Flows []models.Flow `json:"flows"`
	Total uint64        `json:"total"`
}

type flowsGrouped struct {
	Grouped []string `json:"grouped"`
	Total   uint64   `json:"total"`
}

var flowsSQLMappers = map[string]any{
	"id":                  "{{table}}.id",
	"status":              "{{table}}.status",
	"title":               "{{table}}.title",
	"model":               "{{table}}.model",
	"model_provider_name": "{{table}}.model_provider_name",
	"model_provider_type": "{{table}}.model_provider_type",
	"language":            "{{table}}.language",
	"created_at":          "{{table}}.created_at",
	"updated_at":          "{{table}}.updated_at",
	"data":                "({{table}}.status || ' ' || {{table}}.title || ' ' || {{table}}.model || ' ' || {{table}}.model_provider || ' ' || {{table}}.language)",
}

type FlowService struct {
	db *gorm.DB
	pc providers.ProviderController
	fc controller.FlowController
}

func NewFlowService(db *gorm.DB, pc providers.ProviderController, fc controller.FlowController) *FlowService {
	return &FlowService{
		db: db,
		pc: pc,
		fc: fc,
	}
}

// GetFlows is a function to return flows list
// @Summary Retrieve flows list
// @Tags Flows
// @Produce json
// @Security BearerAuth
// @Param request query rdb.TableQuery true "query table params"
// @Success 200 {object} response.successResp{data=flows} "flows list received successful"
// @Failure 400 {object} response.errorResp "invalid query request data"
// @Failure 403 {object} response.errorResp "getting flows not permitted"
// @Failure 500 {object} response.errorResp "internal error on getting flows"
// @Router /flows/ [get]
func (s *FlowService) GetFlows(c *gin.Context) {
	var (
		err   error
		query rdb.TableQuery
		resp  flows
	)

	if err = c.ShouldBindQuery(&query); err != nil {
		logger.FromContext(c).WithError(err).Errorf("error binding query")
		response.Error(c, response.ErrFlowsInvalidRequest, err)
		return
	}

	uid := c.GetUint64("uid")
	privs := c.GetStringSlice("prm")
	var scope func(db *gorm.DB) *gorm.DB
	if slices.Contains(privs, "flows.admin") {
		scope = func(db *gorm.DB) *gorm.DB {
			return db
		}
	} else if slices.Contains(privs, "flows.view") {
		scope = func(db *gorm.DB) *gorm.DB {
			return db.Where("user_id = ?", uid)
		}
	} else {
		logger.FromContext(c).Errorf("error filtering user role permissions: permission not found")
		response.Error(c, response.ErrNotPermitted, nil)
		return
	}

	query.Init("flows", flowsSQLMappers)

	if query.Group != "" {
		if _, ok := flowsSQLMappers[query.Group]; !ok {
			logger.FromContext(c).Errorf("error finding flows grouped: group field not found")
			response.Error(c, response.ErrFlowsInvalidRequest, errors.New("group field not found"))
			return
		}

		var respGrouped flowsGrouped
		if respGrouped.Total, err = query.QueryGrouped(s.db, &respGrouped.Grouped, scope); err != nil {
			logger.FromContext(c).WithError(err).Errorf("error finding flows grouped")
			response.Error(c, response.ErrInternal, err)
			return
		}

		response.Success(c, http.StatusOK, respGrouped)
		return
	}

	if resp.Total, err = query.Query(s.db, &resp.Flows, scope); err != nil {
		logger.FromContext(c).WithError(err).Errorf("error finding flows")
		response.Error(c, response.ErrInternal, err)
		return
	}

	for i := 0; i < len(resp.Flows); i++ {
		if err = resp.Flows[i].Valid(); err != nil {
			logger.FromContext(c).WithError(err).Errorf("error validating flow data '%d'", resp.Flows[i].ID)
			response.Error(c, response.ErrFlowsInvalidData, err)
			return
		}
	}

	response.Success(c, http.StatusOK, resp)
}

// GetFlow is a function to return flow by id
// @Summary Retrieve flow by id
// @Tags Flows
// @Produce json
// @Security BearerAuth
// @Param flowID path int true "flow id" minimum(0)
// @Success 200 {object} response.successResp{data=models.Flow} "flow received successful"
// @Failure 403 {object} response.errorResp "getting flow not permitted"
// @Failure 404 {object} response.errorResp "flow not found"
// @Failure 500 {object} response.errorResp "internal error on getting flow"
// @Router /flows/{flowID} [get]
func (s *FlowService) GetFlow(c *gin.Context) {
	var (
		err    error
		flowID uint64
		resp   models.Flow
	)

	if flowID, err = strconv.ParseUint(c.Param("flowID"), 10, 64); err != nil {
		logger.FromContext(c).WithError(err).Errorf("error parsing flow id")
		response.Error(c, response.ErrFlowsInvalidRequest, err)
		return
	}

	uid := c.GetUint64("uid")
	privs := c.GetStringSlice("prm")
	var scope func(db *gorm.DB) *gorm.DB
	if slices.Contains(privs, "flows.admin") {
		scope = func(db *gorm.DB) *gorm.DB {
			return db.Where("id = ?", flowID)
		}
	} else if slices.Contains(privs, "flows.view") {
		scope = func(db *gorm.DB) *gorm.DB {
			return db.Where("id = ? AND user_id = ?", flowID, uid)
		}
	} else {
		logger.FromContext(c).Errorf("error filtering user role permissions: permission not found")
		response.Error(c, response.ErrNotPermitted, nil)
		return
	}

	if err = s.db.Model(&resp).Scopes(scope).Take(&resp).Error; err != nil {
		logger.FromContext(c).WithError(err).Errorf("error on getting flow by id")
		if gorm.IsRecordNotFoundError(err) {
			response.Error(c, response.ErrFlowsNotFound, err)
		} else {
			response.Error(c, response.ErrInternal, err)
		}
		return
	}

	response.Success(c, http.StatusOK, resp)
}

// GetFlowGraph is a function to return flow graph by id
// @Summary Retrieve flow graph by id
// @Tags Flows
// @Produce json
// @Security BearerAuth
// @Param flowID path int true "flow id" minimum(0)
// @Success 200 {object} response.successResp{data=models.FlowTasksSubtasks} "flow graph received successful"
// @Failure 403 {object} response.errorResp "getting flow graph not permitted"
// @Failure 404 {object} response.errorResp "flow graph not found"
// @Failure 500 {object} response.errorResp "internal error on getting flow graph"
// @Router /flows/{flowID}/graph [get]
func (s *FlowService) GetFlowGraph(c *gin.Context) {
	var (
		err    error
		flowID uint64
		resp   models.FlowTasksSubtasks
		tids   []uint64
	)

	if flowID, err = strconv.ParseUint(c.Param("flowID"), 10, 64); err != nil {
		logger.FromContext(c).WithError(err).Errorf("error parsing flow id")
		response.Error(c, response.ErrFlowsInvalidRequest, err)
		return
	}

	uid := c.GetUint64("uid")
	privs := c.GetStringSlice("prm")
	var scope func(db *gorm.DB) *gorm.DB
	if slices.Contains(privs, "flows.admin") {
		scope = func(db *gorm.DB) *gorm.DB {
			return db.Where("id = ?", flowID)
		}
	} else if slices.Contains(privs, "flows.view") {
		scope = func(db *gorm.DB) *gorm.DB {
			return db.Where("id = ? AND user_id = ?", flowID, uid)
		}
	} else {
		logger.FromContext(c).Errorf("error filtering user role permissions: permission not found")
		response.Error(c, response.ErrNotPermitted, nil)
		return
	}

	err = s.db.Model(&resp).
		Scopes(scope).
		Take(&resp).Error
	if err != nil {
		logger.FromContext(c).WithError(err).Errorf("error on getting flow by id")
		if gorm.IsRecordNotFoundError(err) {
			response.Error(c, response.ErrFlowsNotFound, err)
		} else {
			response.Error(c, response.ErrInternal, err)
		}
		return
	}

	isTasksAdmin := slices.Contains(privs, "tasks.admin")
	isTasksView := slices.Contains(privs, "tasks.view")
	if !(resp.UserID == uid && isTasksView) && !(resp.UserID != uid && isTasksAdmin) {
		response.Success(c, http.StatusOK, resp)
		return
	}

	if resp.UserID != uid && !slices.Contains(privs, "tasks.admin") {
		response.Success(c, http.StatusOK, resp)
		return
	}

	err = s.db.Model(&resp).Association("tasks").Find(&resp.Tasks).Error
	if err != nil {
		logger.FromContext(c).WithError(err).Errorf("error on getting flow tasks")
		response.Error(c, response.ErrInternal, err)
		return
	}

	isSubtasksAdmin := slices.Contains(privs, "subtasks.admin")
	isSubtasksView := slices.Contains(privs, "subtasks.view")
	if !(resp.UserID == uid && isSubtasksView) && !(resp.UserID != uid && isSubtasksAdmin) {
		response.Success(c, http.StatusOK, resp)
		return
	}

	for _, task := range resp.Tasks {
		tids = append(tids, task.ID)
	}

	var subtasks []models.Subtask
	err = s.db.Model(&subtasks).Where("task_id IN (?)", tids).Find(&subtasks).Error
	if err != nil {
		logger.FromContext(c).WithError(err).Errorf("error on getting flow subtasks")
		response.Error(c, response.ErrInternal, err)
		return
	}

	tasksSubtasks := map[uint64][]models.Subtask{}
	for _, subtask := range subtasks {
		tasksSubtasks[subtask.TaskID] = append(tasksSubtasks[subtask.TaskID], subtask)
	}

	for i := range resp.Tasks {
		resp.Tasks[i].Subtasks = tasksSubtasks[resp.Tasks[i].ID]
	}

	if err = resp.Valid(); err != nil {
		logger.FromContext(c).WithError(err).Errorf("error validating flow data '%d'", flowID)
		response.Error(c, response.ErrFlowsInvalidData, err)
		return
	}

	response.Success(c, http.StatusOK, resp)
}

// CreateFlow is a function to create new flow with custom functions
// @Summary Create new flow with custom functions
// @Tags Flows
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param json body models.CreateFlow true "flow model to create"
// @Success 201 {object} response.successResp{data=models.Flow} "flow created successful"
// @Failure 400 {object} response.errorResp "invalid flow request data"
// @Failure 403 {object} response.errorResp "creating flow not permitted"
// @Failure 500 {object} response.errorResp "internal error on creating flow"
// @Router /flows/ [post]
func (s *FlowService) CreateFlow(c *gin.Context) {
	var (
		err        error
		flow       models.Flow
		createFlow models.CreateFlow
	)

	if err := c.ShouldBindJSON(&createFlow); err != nil {
		logger.FromContext(c).WithError(err).Errorf("error binding JSON")
		response.Error(c, response.ErrFlowsInvalidRequest, err)
		return
	}

	privs := c.GetStringSlice("prm")
	if !slices.Contains(privs, "flows.create") {
		logger.FromContext(c).Errorf("error filtering user role permissions: permission not found")
		response.Error(c, response.ErrNotPermitted, nil)
		return
	}

	if err := createFlow.Valid(); err != nil {
		logger.FromContext(c).WithError(err).Errorf("error validating flow data")
		response.Error(c, response.ErrFlowsInvalidData, err)
		return
	}

	uid := c.GetUint64("uid")
	prvname := provider.ProviderName(createFlow.Provider)

	prv, err := s.pc.GetProvider(c, prvname, int64(uid))
	if err != nil {
		logger.FromContext(c).WithError(err).Errorf("error getting provider: not found")
		response.Error(c, response.ErrInternal, err)
		return
	}
	prvtype := prv.Type()

	fw, err := s.fc.CreateFlow(c, int64(uid), createFlow.Input, prvname, prvtype, createFlow.Functions)
	if err != nil {
		logger.FromContext(c).WithError(err).Errorf("error creating flow")
		response.Error(c, response.ErrInternal, err)
		return
	}

	err = s.db.Model(&flow).Where("id = ?", fw.GetFlowID()).Take(&flow).Error
	if err != nil {
		logger.FromContext(c).WithError(err).Errorf("error getting flow by id")
		response.Error(c, response.ErrInternal, err)
		return
	}

	response.Success(c, http.StatusCreated, flow)
}

// PatchFlow is a function to patch flow
// @Summary Patch flow
// @Tags Flows
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param flowID path int true "flow id" minimum(0)
// @Param json body models.PatchFlow true "flow model to patch"
// @Success 200 {object} response.successResp{data=models.Flow} "flow patched successful"
// @Failure 400 {object} response.errorResp "invalid flow request data"
// @Failure 403 {object} response.errorResp "patching flow not permitted"
// @Failure 500 {object} response.errorResp "internal error on patching flow"
// @Router /flows/{flowID} [put]
func (s *FlowService) PatchFlow(c *gin.Context) {
	var (
		err       error
		flow      models.Flow
		flowID    uint64
		patchFlow models.PatchFlow
	)

	if err := c.ShouldBindJSON(&patchFlow); err != nil {
		logger.FromContext(c).WithError(err).Errorf("error binding JSON")
		response.Error(c, response.ErrFlowsInvalidRequest, err)
		return
	}

	if err := patchFlow.Valid(); err != nil {
		logger.FromContext(c).WithError(err).Errorf("error validating flow data")
		response.Error(c, response.ErrFlowsInvalidData, err)
		return
	}

	flowID, err = strconv.ParseUint(c.Param("flowID"), 10, 64)
	if err != nil {
		logger.FromContext(c).WithError(err).Errorf("error parsing flow id")
		response.Error(c, response.ErrFlowsInvalidRequest, err)
		return
	}

	uid := c.GetUint64("uid")
	privs := c.GetStringSlice("prm")
	var scope func(db *gorm.DB) *gorm.DB
	if slices.Contains(privs, "flows.admin") {
		scope = func(db *gorm.DB) *gorm.DB {
			return db.Where("id = ?", flowID)
		}
	} else if slices.Contains(privs, "flows.edit") {
		scope = func(db *gorm.DB) *gorm.DB {
			return db.Where("id = ? AND user_id = ?", flowID, uid)
		}
	} else {
		logger.FromContext(c).Errorf("error filtering user role permissions: permission not found")
		response.Error(c, response.ErrNotPermitted, nil)
		return
	}

	if err = s.db.Model(&flow).Scopes(scope).Take(&flow).Error; err != nil {
		logger.FromContext(c).WithError(err).Errorf("error getting flow by id")
		if gorm.IsRecordNotFoundError(err) {
			response.Error(c, response.ErrFlowsNotFound, err)
		} else {
			response.Error(c, response.ErrInternal, err)
		}
		return
	}

	fw, err := s.fc.GetFlow(c, int64(flow.ID))
	if err != nil {
		logger.FromContext(c).WithError(err).Errorf("error getting flow by id in flow controller")
		response.Error(c, response.ErrInternal, err)
		return
	}

	switch patchFlow.Action {
	case "stop":
		if err := fw.Stop(c); err != nil {
			logger.FromContext(c).WithError(err).Errorf("error stopping flow")
			response.Error(c, response.ErrInternal, err)
			return
		}
	case "finish":
		if err := fw.Finish(c); err != nil {
			logger.FromContext(c).WithError(err).Errorf("error finishing flow")
			response.Error(c, response.ErrInternal, err)
			return
		}
	case "input":
		if patchFlow.Input == nil || *patchFlow.Input == "" {
			logger.FromContext(c).Errorf("error sending input to flow: input is empty")
			response.Error(c, response.ErrFlowsInvalidRequest, nil)
			return
		}

		if err := fw.PutInput(c, *patchFlow.Input); err != nil {
			logger.FromContext(c).WithError(err).Errorf("error sending input to flow")
			response.Error(c, response.ErrInternal, err)
			return
		}
	default:
		logger.FromContext(c).Errorf("error filtering flow action")
		response.Error(c, response.ErrFlowsInvalidRequest, nil)
		return
	}

	if err = s.db.Model(&flow).Scopes(scope).Take(&flow).Error; err != nil {
		logger.FromContext(c).WithError(err).Errorf("error getting flow by id")
		if gorm.IsRecordNotFoundError(err) {
			response.Error(c, response.ErrFlowsNotFound, err)
		} else {
			response.Error(c, response.ErrInternal, err)
		}
		return
	}

	response.Success(c, http.StatusOK, flow)
}

// DeleteFlow is a function to delete flow by id
// @Summary Delete flow by id
// @Tags Flows
// @Security BearerAuth
// @Param flowID path int true "flow id" minimum(0)
// @Success 200 {object} response.successResp{data=models.Flow} "flow deleted successful"
// @Failure 403 {object} response.errorResp "deleting flow not permitted"
// @Failure 404 {object} response.errorResp "flow not found"
// @Failure 500 {object} response.errorResp "internal error on deleting flow"
// @Router /flows/{flowID} [delete]
func (s *FlowService) DeleteFlow(c *gin.Context) {
	var (
		err    error
		flow   models.Flow
		flowID uint64
	)

	flowID, err = strconv.ParseUint(c.Param("flowID"), 10, 64)
	if err != nil {
		logger.FromContext(c).WithError(err).Errorf("error parsing flow id")
		response.Error(c, response.ErrFlowsInvalidRequest, err)
		return
	}

	uid := c.GetUint64("uid")
	privs := c.GetStringSlice("prm")
	var scope func(db *gorm.DB) *gorm.DB
	if slices.Contains(privs, "flows.admin") {
		scope = func(db *gorm.DB) *gorm.DB {
			return db.Where("id = ?", flowID)
		}
	} else if slices.Contains(privs, "flows.delete") {
		scope = func(db *gorm.DB) *gorm.DB {
			return db.Where("id = ? AND user_id = ?", flowID, uid)
		}
	} else {
		logger.FromContext(c).Errorf("error filtering user role permissions: permission not found")
		response.Error(c, response.ErrNotPermitted, nil)
		return
	}

	if err = s.db.Model(&flow).Scopes(scope).Take(&flow).Error; err != nil {
		logger.FromContext(c).WithError(err).Errorf("error getting flow by id")
		if gorm.IsRecordNotFoundError(err) {
			response.Error(c, response.ErrFlowsNotFound, err)
		} else {
			response.Error(c, response.ErrInternal, err)
		}
		return
	}

	if err := s.fc.FinishFlow(c, int64(flow.ID)); err != nil {
		logger.FromContext(c).WithError(err).Errorf("error stopping flow")
		response.Error(c, response.ErrInternal, err)
		return
	}

	if err = s.db.Scopes(scope).Delete(&flow).Error; err != nil {
		logger.FromContext(c).WithError(err).Errorf("error deleting flow by id")
		if gorm.IsRecordNotFoundError(err) {
			response.Error(c, response.ErrFlowsNotFound, err)
		} else {
			response.Error(c, response.ErrInternal, err)
		}
		return
	}

	response.Success(c, http.StatusOK, flow)
}
