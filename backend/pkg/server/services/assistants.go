package services

import (
	"encoding/json"
	"errors"
	"net/http"
	"slices"
	"strconv"

	"pentagi/pkg/controller"
	"pentagi/pkg/database"
	"pentagi/pkg/graph/subscriptions"
	"pentagi/pkg/providers"
	"pentagi/pkg/providers/provider"
	"pentagi/pkg/server/logger"
	"pentagi/pkg/server/models"
	"pentagi/pkg/server/rdb"
	"pentagi/pkg/server/response"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

type assistants struct {
	Assistants []models.Assistant `json:"assistants"`
	Total      uint64             `json:"total"`
}

type assistantsGrouped struct {
	Grouped []string `json:"grouped"`
	Total   uint64   `json:"total"`
}

var assistantsSQLMappers = map[string]any{
	"id":                  "{{table}}.id",
	"status":              "{{table}}.status",
	"title":               "{{table}}.title",
	"model":               "{{table}}.model",
	"model_provider_name": "{{table}}.model_provider_name",
	"model_provider_type": "{{table}}.model_provider_type",
	"language":            "{{table}}.language",
	"flow_id":             "{{table}}.flow_id",
	"msgchain_id":         "{{table}}.msgchain_id",
	"created_at":          "{{table}}.created_at",
	"updated_at":          "{{table}}.updated_at",
	"data":                "({{table}}.status || ' ' || {{table}}.title || ' ' || {{table}}.flow_id)",
}

type AssistantService struct {
	db *gorm.DB
	pc providers.ProviderController
	fc controller.FlowController
	ss subscriptions.SubscriptionsController
}

func NewAssistantService(
	db *gorm.DB,
	pc providers.ProviderController,
	fc controller.FlowController,
	ss subscriptions.SubscriptionsController,
) *AssistantService {
	return &AssistantService{
		db: db,
		pc: pc,
		fc: fc,
		ss: ss,
	}
}

// GetAssistants is a function to return assistants list
// @Summary Retrieve assistants list
// @Tags Assistants
// @Produce json
// @Security BearerAuth
// @Param flowID path int true "flow id" minimum(0)
// @Param request query rdb.TableQuery true "query table params"
// @Success 200 {object} response.successResp{data=assistants} "assistants list received successful"
// @Failure 400 {object} response.errorResp "invalid query request data"
// @Failure 403 {object} response.errorResp "getting assistants not permitted"
// @Failure 500 {object} response.errorResp "internal error on getting assistants"
// @Router /flows/{flowID}/assistants/ [get]
func (s *AssistantService) GetFlowAssistants(c *gin.Context) {
	var (
		err    error
		flowID uint64
		query  rdb.TableQuery
		resp   assistants
	)

	if err = c.ShouldBindQuery(&query); err != nil {
		logger.FromContext(c).WithError(err).Errorf("error binding query")
		response.Error(c, response.ErrAssistantsInvalidRequest, err)
		return
	}

	if flowID, err = strconv.ParseUint(c.Param("flowID"), 10, 64); err != nil {
		logger.FromContext(c).WithError(err).Errorf("error parsing flow id")
		response.Error(c, response.ErrAssistantsInvalidRequest, err)
		return
	}

	uid := c.GetUint64("uid")
	privs := c.GetStringSlice("prm")
	var scope func(db *gorm.DB) *gorm.DB
	if slices.Contains(privs, "assistants.admin") {
		scope = func(db *gorm.DB) *gorm.DB {
			return db.
				Joins("INNER JOIN flows f ON f.id = assistants.flow_id").
				Where("f.id = ?", flowID)
		}
	} else if slices.Contains(privs, "assistants.view") {
		scope = func(db *gorm.DB) *gorm.DB {
			return db.
				Joins("INNER JOIN flows f ON f.id = assistants.flow_id").
				Where("f.id = ? AND f.user_id = ?", flowID, uid)
		}
	} else {
		logger.FromContext(c).Errorf("error filtering user role permissions: permission not found")
		response.Error(c, response.ErrNotPermitted, nil)
		return
	}

	query.Init("assistants", assistantsSQLMappers)

	if query.Group != "" {
		if _, ok := assistantsSQLMappers[query.Group]; !ok {
			logger.FromContext(c).Errorf("error finding assistants grouped: group field not found")
			response.Error(c, response.ErrAssistantsInvalidRequest, errors.New("group field not found"))
			return
		}

		var respGrouped assistantsGrouped
		if respGrouped.Total, err = query.QueryGrouped(s.db, &respGrouped.Grouped, scope); err != nil {
			logger.FromContext(c).WithError(err).Errorf("error finding assistants grouped")
			response.Error(c, response.ErrInternal, err)
			return
		}

		response.Success(c, http.StatusOK, respGrouped)
		return
	}

	if resp.Total, err = query.Query(s.db, &resp.Assistants, scope); err != nil {
		logger.FromContext(c).WithError(err).Errorf("error finding assistants")
		response.Error(c, response.ErrInternal, err)
		return
	}

	for i := 0; i < len(resp.Assistants); i++ {
		if err = resp.Assistants[i].Valid(); err != nil {
			logger.FromContext(c).WithError(err).Errorf("error validating assistant data '%d'", resp.Assistants[i].ID)
			response.Error(c, response.ErrAssistantsInvalidData, err)
			return
		}
	}

	response.Success(c, http.StatusOK, resp)
}

// GetFlowAssistant is a function to return flow assistant by id
// @Summary Retrieve flow assistant by id
// @Tags Assistants
// @Produce json
// @Security BearerAuth
// @Param flowID path int true "flow id" minimum(0)
// @Param assistantID path int true "assistant id" minimum(0)
// @Success 200 {object} response.successResp{data=models.Assistant} "flow assistant received successful"
// @Failure 403 {object} response.errorResp "getting flow assistant not permitted"
// @Failure 404 {object} response.errorResp "flow assistant not found"
// @Failure 500 {object} response.errorResp "internal error on getting flow assistant"
// @Router /flows/{flowID}/assistants/{assistantID} [get]
func (s *AssistantService) GetFlowAssistant(c *gin.Context) {
	var (
		err         error
		flowID      uint64
		assistantID uint64
		resp        models.Assistant
	)

	if flowID, err = strconv.ParseUint(c.Param("flowID"), 10, 64); err != nil {
		logger.FromContext(c).WithError(err).Errorf("error parsing flow id")
		response.Error(c, response.ErrAssistantsInvalidRequest, err)
		return
	}

	if assistantID, err = strconv.ParseUint(c.Param("assistantID"), 10, 64); err != nil {
		logger.FromContext(c).WithError(err).Errorf("error parsing assistant id")
		response.Error(c, response.ErrAssistantsInvalidRequest, err)
		return
	}

	uid := c.GetUint64("uid")
	privs := c.GetStringSlice("prm")
	var scope func(db *gorm.DB) *gorm.DB
	if slices.Contains(privs, "assistants.admin") {
		scope = func(db *gorm.DB) *gorm.DB {
			return db.
				Joins("INNER JOIN flows f ON f.id = assistants.flow_id").
				Where("f.id = ?", flowID)
		}
	} else if slices.Contains(privs, "assistants.view") {
		scope = func(db *gorm.DB) *gorm.DB {
			return db.
				Joins("INNER JOIN flows f ON f.id = assistants.flow_id").
				Where("f.id = ? AND f.user_id = ?", flowID, uid)
		}
	} else {
		logger.FromContext(c).Errorf("error filtering user role permissions: permission not found")
		response.Error(c, response.ErrNotPermitted, nil)
		return
	}

	err = s.db.Model(&resp).
		Scopes(scope).
		Where("assistants.id = ?", assistantID).
		Take(&resp).Error
	if err != nil {
		logger.FromContext(c).WithError(err).Errorf("error on getting flow assistant by id")
		if gorm.IsRecordNotFoundError(err) {
			response.Error(c, response.ErrAssistantsNotFound, err)
		} else {
			response.Error(c, response.ErrInternal, err)
		}
		return
	}

	response.Success(c, http.StatusOK, resp)
}

// CreateFlowAssistant is a function to create new assistant with custom functions
// @Summary Create new assistant with custom functions
// @Tags Assistants
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param flowID path int true "flow id" minimum(0)
// @Param json body models.CreateAssistant true "assistant model to create"
// @Success 201 {object} response.successResp{data=models.AssistantFlow} "assistant created successful"
// @Failure 400 {object} response.errorResp "invalid assistant request data"
// @Failure 403 {object} response.errorResp "creating assistant not permitted"
// @Failure 500 {object} response.errorResp "internal error on creating assistant"
// @Router /flows/{flowID}/assistants/ [post]
func (s *AssistantService) CreateFlowAssistant(c *gin.Context) {
	var (
		err             error
		flowID          uint64
		assistant       models.AssistantFlow
		createAssistant models.CreateAssistant
	)

	if err := c.ShouldBindJSON(&createAssistant); err != nil {
		logger.FromContext(c).WithError(err).Errorf("error binding JSON")
		response.Error(c, response.ErrAssistantsInvalidRequest, err)
		return
	}

	if err := createAssistant.Valid(); err != nil {
		logger.FromContext(c).WithError(err).Errorf("error validating assistant data")
		response.Error(c, response.ErrAssistantsInvalidData, err)
		return
	}

	if flowID, err = strconv.ParseUint(c.Param("flowID"), 10, 64); err != nil {
		logger.FromContext(c).WithError(err).Errorf("error parsing flow id")
		response.Error(c, response.ErrAssistantsInvalidRequest, err)
		return
	}

	privs := c.GetStringSlice("prm")
	if !slices.Contains(privs, "assistants.create") {
		logger.FromContext(c).Errorf("error filtering user role permissions: permission not found")
		response.Error(c, response.ErrNotPermitted, nil)
		return
	}

	uid := c.GetUint64("uid")
	prvname := provider.ProviderName(createAssistant.Provider)

	prv, err := s.pc.GetProvider(c, prvname, int64(uid))
	if err != nil {
		logger.FromContext(c).WithError(err).Errorf("error getting provider: not found")
		response.Error(c, response.ErrInternal, err)
		return
	}
	prvtype := prv.Type()

	aw, err := s.fc.CreateAssistant(
		c,
		int64(uid),
		int64(flowID),
		createAssistant.Input,
		createAssistant.UseAgents,
		prvname,
		prvtype,
		createAssistant.Functions,
	)
	if err != nil {
		logger.FromContext(c).WithError(err).Errorf("error creating assistant")
		response.Error(c, response.ErrInternal, err)
		return
	}

	err = s.db.Model(&assistant).Where("id = ?", aw.GetAssistantID()).Take(&assistant).Error
	if err != nil {
		logger.FromContext(c).WithError(err).Errorf("error getting assistant by id")
		response.Error(c, response.ErrInternal, err)
		return
	}

	response.Success(c, http.StatusCreated, assistant)
}

// PatchAssistant is a function to patch assistant
// @Summary Patch assistant
// @Tags Assistants
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param flowID path int true "flow id" minimum(0)
// @Param assistantID path int true "assistant id" minimum(0)
// @Param json body models.PatchAssistant true "assistant model to patch"
// @Success 200 {object} response.successResp{data=models.AssistantFlow} "assistant patched successful"
// @Failure 400 {object} response.errorResp "invalid assistant request data"
// @Failure 403 {object} response.errorResp "patching assistant not permitted"
// @Failure 500 {object} response.errorResp "internal error on patching assistant"
// @Router /flows/{flowID}/assistants/{assistantID} [put]
func (s *AssistantService) PatchAssistant(c *gin.Context) {
	var (
		err            error
		flowID         uint64
		assistant      models.AssistantFlow
		assistantID    uint64
		patchAssistant models.PatchAssistant
	)

	if err := c.ShouldBindJSON(&patchAssistant); err != nil {
		logger.FromContext(c).WithError(err).Errorf("error binding JSON")
		response.Error(c, response.ErrAssistantsInvalidRequest, err)
		return
	}

	if err := patchAssistant.Valid(); err != nil {
		logger.FromContext(c).WithError(err).Errorf("error validating assistant data")
		response.Error(c, response.ErrAssistantsInvalidData, err)
		return
	}

	flowID, err = strconv.ParseUint(c.Param("flowID"), 10, 64)
	if err != nil {
		logger.FromContext(c).WithError(err).Errorf("error parsing flow id")
		response.Error(c, response.ErrAssistantsInvalidRequest, err)
		return
	}

	assistantID, err = strconv.ParseUint(c.Param("assistantID"), 10, 64)
	if err != nil {
		logger.FromContext(c).WithError(err).Errorf("error parsing assistant id")
		response.Error(c, response.ErrAssistantsInvalidRequest, err)
		return
	}

	uid := c.GetUint64("uid")
	privs := c.GetStringSlice("prm")
	var scope func(db *gorm.DB) *gorm.DB
	if slices.Contains(privs, "assistants.admin") {
		scope = func(db *gorm.DB) *gorm.DB {
			return db.Where("id = ?", assistantID)
		}
	} else if slices.Contains(privs, "assistants.edit") {
		scope = func(db *gorm.DB) *gorm.DB {
			return db.Where("assistants.id = ? AND assistants.flow_id = ?", assistantID, flowID).
				Joins("INNER JOIN flows f ON f.id = assistants.flow_id").
				Where("f.user_id = ?", uid)
		}
	} else {
		logger.FromContext(c).Errorf("error filtering user role permissions: permission not found")
		response.Error(c, response.ErrNotPermitted, nil)
		return
	}

	if err = s.db.Model(&assistant).Scopes(scope).Take(&assistant).Error; err != nil {
		logger.FromContext(c).WithError(err).Errorf("error getting assistant by id")
		if gorm.IsRecordNotFoundError(err) {
			response.Error(c, response.ErrAssistantsNotFound, err)
		} else {
			response.Error(c, response.ErrInternal, err)
		}
		return
	}

	fw, err := s.fc.GetFlow(c, int64(flowID))
	if err != nil {
		logger.FromContext(c).WithError(err).Errorf("error getting flow by id in flow controller")
		response.Error(c, response.ErrInternal, err)
		return
	}

	aw, err := fw.GetAssistant(c, int64(assistantID))
	if err != nil {
		logger.FromContext(c).WithError(err).Errorf("error getting assistant by id in flow controller")
		response.Error(c, response.ErrInternal, err)
		return
	}

	switch patchAssistant.Action {
	case "stop":
		if err := aw.Stop(c); err != nil {
			logger.FromContext(c).WithError(err).Errorf("error stopping assistant")
			response.Error(c, response.ErrInternal, err)
			return
		}
	case "input":
		if patchAssistant.Input == nil || *patchAssistant.Input == "" {
			logger.FromContext(c).Errorf("error sending input to assistant: input is empty")
			response.Error(c, response.ErrAssistantsInvalidRequest, nil)
			return
		}

		if err := aw.PutInput(c, *patchAssistant.Input, patchAssistant.UseAgents); err != nil {
			logger.FromContext(c).WithError(err).Errorf("error sending input to assistant")
			response.Error(c, response.ErrInternal, err)
			return
		}
	default:
		logger.FromContext(c).Errorf("error filtering assistant action")
		response.Error(c, response.ErrAssistantsInvalidRequest, nil)
		return
	}

	if err = s.db.Model(&assistant).Scopes(scope).Take(&assistant).Error; err != nil {
		logger.FromContext(c).WithError(err).Errorf("error getting assistant by id")
		if gorm.IsRecordNotFoundError(err) {
			response.Error(c, response.ErrAssistantsNotFound, err)
		} else {
			response.Error(c, response.ErrInternal, err)
		}
		return
	}

	assistantDB, err := convertAssistantToDatabase(assistant.Assistant)
	if err != nil {
		logger.FromContext(c).WithError(err).Errorf("error converting assistant to database")
		response.Error(c, response.ErrInternal, err)
		return
	}

	if s.ss != nil {
		publisher := s.ss.NewFlowPublisher(int64(assistant.Flow.UserID), int64(assistant.FlowID))
		publisher.AssistantUpdated(c, assistantDB)
	}

	response.Success(c, http.StatusOK, assistant)
}

// DeleteAssistant is a function to delete assistant by id
// @Summary Delete assistant by id
// @Tags Assistants
// @Security BearerAuth
// @Param flowID path int true "flow id" minimum(0)
// @Param assistantID path int true "assistant id" minimum(0)
// @Success 200 {object} response.successResp{data=models.AssistantFlow} "assistant deleted successful"
// @Failure 403 {object} response.errorResp "deleting assistant not permitted"
// @Failure 404 {object} response.errorResp "assistant not found"
// @Failure 500 {object} response.errorResp "internal error on deleting assistant"
// @Router /flows/{flowID}/assistants/{assistantID} [delete]
func (s *AssistantService) DeleteAssistant(c *gin.Context) {
	var (
		err         error
		flowID      uint64
		assistant   models.AssistantFlow
		assistantID uint64
	)

	flowID, err = strconv.ParseUint(c.Param("flowID"), 10, 64)
	if err != nil {
		logger.FromContext(c).WithError(err).Errorf("error parsing flow id")
		response.Error(c, response.ErrAssistantsInvalidRequest, err)
		return
	}

	assistantID, err = strconv.ParseUint(c.Param("assistantID"), 10, 64)
	if err != nil {
		logger.FromContext(c).WithError(err).Errorf("error parsing assistant id")
		response.Error(c, response.ErrAssistantsInvalidRequest, err)
		return
	}

	uid := c.GetUint64("uid")
	privs := c.GetStringSlice("prm")
	var scope func(db *gorm.DB) *gorm.DB
	if slices.Contains(privs, "assistants.admin") {
		scope = func(db *gorm.DB) *gorm.DB {
			return db.Where("id = ?", assistantID)
		}
	} else if slices.Contains(privs, "assistants.delete") {
		scope = func(db *gorm.DB) *gorm.DB {
			return db.Where("assistants.id = ? AND assistants.flow_id = ?", assistantID, flowID).
				Joins("INNER JOIN flows f ON f.id = assistants.flow_id").
				Where("f.user_id = ?", uid)
		}
	} else {
		logger.FromContext(c).Errorf("error filtering user role permissions: permission not found")
		response.Error(c, response.ErrNotPermitted, nil)
		return
	}

	if err = s.db.Model(&assistant).Scopes(scope).Take(&assistant).Error; err != nil {
		logger.FromContext(c).WithError(err).Errorf("error getting assistant by id")
		if gorm.IsRecordNotFoundError(err) {
			response.Error(c, response.ErrAssistantsNotFound, err)
		} else {
			response.Error(c, response.ErrInternal, err)
		}
		return
	}

	fw, err := s.fc.GetFlow(c, int64(flowID))
	if err != nil {
		logger.FromContext(c).WithError(err).Errorf("error getting flow by id in flow controller")
		response.Error(c, response.ErrInternal, err)
		return
	}

	aw, err := fw.GetAssistant(c, int64(assistantID))
	if err != nil {
		logger.FromContext(c).WithError(err).Errorf("error getting assistant by id in flow controller")
		response.Error(c, response.ErrInternal, err)
		return
	}

	if err := aw.Finish(c); err != nil {
		logger.FromContext(c).WithError(err).Errorf("error stopping assistant")
		response.Error(c, response.ErrInternal, err)
		return
	}

	if err = s.db.Scopes(scope).Delete(&assistant).Error; err != nil {
		logger.FromContext(c).WithError(err).Errorf("error deleting assistant by id")
		if gorm.IsRecordNotFoundError(err) {
			response.Error(c, response.ErrAssistantsNotFound, err)
		} else {
			response.Error(c, response.ErrInternal, err)
		}
		return
	}

	assistantDB, err := convertAssistantToDatabase(assistant.Assistant)
	if err != nil {
		logger.FromContext(c).WithError(err).Errorf("error converting assistant to database")
		response.Error(c, response.ErrInternal, err)
		return
	}

	if s.ss != nil {
		publisher := s.ss.NewFlowPublisher(int64(assistant.Flow.UserID), int64(assistant.FlowID))
		publisher.AssistantDeleted(c, assistantDB)
	}

	response.Success(c, http.StatusOK, assistant)
}

func convertAssistantToDatabase(assistant models.Assistant) (database.Assistant, error) {
	functions, err := json.Marshal(assistant.Functions)
	if err != nil {
		return database.Assistant{}, err
	}

	return database.Assistant{
		ID:                 int64(assistant.ID),
		Status:             database.AssistantStatus(assistant.Status),
		Title:              assistant.Title,
		Model:              assistant.Model,
		ModelProviderName:  assistant.ModelProviderName,
		Language:           assistant.Language,
		Functions:          functions,
		TraceID:            database.PtrStringToNullString(assistant.TraceID),
		FlowID:             int64(assistant.FlowID),
		UseAgents:          assistant.UseAgents,
		MsgchainID:         database.Uint64ToNullInt64(assistant.MsgchainID),
		CreatedAt:          database.TimeToNullTime(assistant.CreatedAt),
		UpdatedAt:          database.TimeToNullTime(assistant.UpdatedAt),
		DeletedAt:          database.PtrTimeToNullTime(assistant.DeletedAt),
		ModelProviderType:  database.ProviderType(assistant.ModelProviderType),
		ToolCallIDTemplate: assistant.ToolCallIDTemplate,
	}, nil
}
