package services

import (
	"errors"
	"net/http"
	"slices"

	"pentagi/pkg/server/logger"
	"pentagi/pkg/server/models"
	"pentagi/pkg/server/rdb"
	"pentagi/pkg/server/response"
	"pentagi/pkg/templates"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

type prompts struct {
	Prompts []models.Prompt `json:"prompts"`
	Total   uint64          `json:"total"`
}

type promptsGrouped struct {
	Grouped []string `json:"grouped"`
	Total   uint64   `json:"total"`
}

var promptsSQLMappers = map[string]any{
	"type":       "{{table}}.type",
	"prompt":     "{{table}}.prompt",
	"created_at": "{{table}}.created_at",
	"updated_at": "{{table}}.updated_at",
	"data":       "({{table}}.type || ' ' || {{table}}.prompt)",
}

type PromptService struct {
	db       *gorm.DB
	prompter templates.Prompter
}

func NewPromptService(db *gorm.DB) *PromptService {
	return &PromptService{
		db:       db,
		prompter: templates.NewDefaultPrompter(),
	}
}

// GetPrompts is a function to return prompts list
// @Summary Retrieve prompts list
// @Tags Prompts
// @Produce json
// @Security BearerAuth
// @Param request query rdb.TableQuery true "query table params"
// @Success 200 {object} response.successResp{data=prompts} "prompts list received successful"
// @Failure 400 {object} response.errorResp "invalid query request data"
// @Failure 403 {object} response.errorResp "getting prompts not permitted"
// @Failure 500 {object} response.errorResp "internal error on getting prompts"
// @Router /prompts/ [get]
func (s *PromptService) GetPrompts(c *gin.Context) {
	var (
		err   error
		query rdb.TableQuery
		resp  prompts
	)

	if err = c.ShouldBindQuery(&query); err != nil {
		logger.FromContext(c).WithError(err).Errorf("error binding query")
		response.Error(c, response.ErrPromptsInvalidRequest, err)
		return
	}

	privs := c.GetStringSlice("prm")
	if !slices.Contains(privs, "settings.prompts.view") {
		logger.FromContext(c).Errorf("error filtering user role permissions: permission not found")
		response.Error(c, response.ErrNotPermitted, nil)
		return
	}

	uid := c.GetUint64("uid")
	scope := func(db *gorm.DB) *gorm.DB {
		return db.Where("user_id = ?", uid)
	}

	query.Init("prompts", promptsSQLMappers)

	if query.Group != "" {
		if _, ok := promptsSQLMappers[query.Group]; !ok {
			logger.FromContext(c).Errorf("error finding prompts grouped: group field not found")
			response.Error(c, response.ErrPromptsInvalidRequest, errors.New("group field not found"))
			return
		}

		var respGrouped promptsGrouped
		if respGrouped.Total, err = query.QueryGrouped(s.db, &respGrouped.Grouped, scope); err != nil {
			logger.FromContext(c).WithError(err).Errorf("error finding prompts grouped")
			response.Error(c, response.ErrInternal, err)
			return
		}

		response.Success(c, http.StatusOK, respGrouped)
		return
	}

	if resp.Total, err = query.Query(s.db, &resp.Prompts, scope); err != nil {
		logger.FromContext(c).WithError(err).Errorf("error finding prompts")
		response.Error(c, response.ErrInternal, err)
		return
	}

	for i := 0; i < len(resp.Prompts); i++ {
		if err = resp.Prompts[i].Valid(); err != nil {
			logger.FromContext(c).WithError(err).Errorf("error validating prompt data '%s'", resp.Prompts[i].Type)
			response.Error(c, response.ErrPromptsInvalidData, err)
			return
		}
	}

	response.Success(c, http.StatusOK, resp)
}

// GetPrompt is a function to return prompt by type
// @Summary Retrieve prompt by type
// @Tags Prompts
// @Produce json
// @Security BearerAuth
// @Param promptType path string true "prompt type"
// @Success 200 {object} response.successResp{data=models.Prompt} "prompt received successful"
// @Failure 400 {object} response.errorResp "invalid prompt request data"
// @Failure 403 {object} response.errorResp "getting prompt not permitted"
// @Failure 404 {object} response.errorResp "prompt not found"
// @Failure 500 {object} response.errorResp "internal error on getting prompt"
// @Router /prompts/{promptType} [get]
func (s *PromptService) GetPrompt(c *gin.Context) {
	var (
		err        error
		promptType models.PromptType = models.PromptType(c.Param("promptType"))
		resp       models.Prompt
	)

	if err = models.PromptType(promptType).Valid(); err != nil {
		logger.FromContext(c).WithError(err).Errorf("error validating prompt type '%s'", promptType)
		response.Error(c, response.ErrPromptsInvalidRequest, err)
		return
	}

	privs := c.GetStringSlice("prm")
	if !slices.Contains(privs, "settings.prompts.view") {
		logger.FromContext(c).Errorf("error filtering user role permissions: permission not found")
		response.Error(c, response.ErrNotPermitted, nil)
		return
	}

	uid := c.GetUint64("uid")
	scope := func(db *gorm.DB) *gorm.DB {
		return db.Where("type = ? AND user_id = ?", promptType, uid)
	}

	if err = s.db.Scopes(scope).Take(&resp).Error; err != nil {
		logger.FromContext(c).WithError(err).Errorf("error finding prompt by type")
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Error(c, response.ErrPromptsNotFound, err)
		} else {
			response.Error(c, response.ErrInternal, err)
		}
		return
	}
	if err = resp.Valid(); err != nil {
		logger.FromContext(c).WithError(err).Errorf("error validating prompt data '%s'", resp.Type)
		response.Error(c, response.ErrPromptsInvalidData, err)
		return
	}

	response.Success(c, http.StatusOK, resp)
}

// PatchPrompt is a function to update prompt by type
// @Summary Update prompt
// @Tags Prompts
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param promptType path string true "prompt type"
// @Param json body models.PatchPrompt true "prompt model to update"
// @Success 200 {object} response.successResp{data=models.Prompt} "prompt updated successful"
// @Success 201 {object} response.successResp{data=models.Prompt} "prompt created successful"
// @Failure 400 {object} response.errorResp "invalid prompt request data"
// @Failure 403 {object} response.errorResp "updating prompt not permitted"
// @Failure 404 {object} response.errorResp "prompt not found"
// @Failure 500 {object} response.errorResp "internal error on updating prompt"
// @Router /prompts/{promptType} [put]
func (s *PromptService) PatchPrompt(c *gin.Context) {
	var (
		err        error
		prompt     models.PatchPrompt
		promptType models.PromptType = models.PromptType(c.Param("promptType"))
		resp       models.Prompt
	)

	if err = c.ShouldBindJSON(&prompt); err != nil {
		logger.FromContext(c).WithError(err).Errorf("error binding JSON")
		response.Error(c, response.ErrPromptsInvalidRequest, err)
		return
	} else if err = prompt.Valid(); err != nil {
		logger.FromContext(c).WithError(err).Errorf("error validating prompt JSON")
		response.Error(c, response.ErrPromptsInvalidRequest, err)
		return
	} else if err = promptType.Valid(); err != nil {
		logger.FromContext(c).WithError(err).Errorf("error validating prompt type '%s'", promptType)
		response.Error(c, response.ErrPromptsInvalidRequest, err)
		return
	}

	privs := c.GetStringSlice("prm")
	if !slices.Contains(privs, "settings.prompts.edit") {
		logger.FromContext(c).Errorf("error filtering user role permissions: permission not found")
		response.Error(c, response.ErrNotPermitted, nil)
		return
	}

	uid := c.GetUint64("uid")
	scope := func(db *gorm.DB) *gorm.DB {
		return db.Where("type = ? AND user_id = ?", promptType, uid)
	}

	err = s.db.Scopes(scope).Take(&resp).Error
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		resp = models.Prompt{
			Type:   promptType,
			UserID: uid,
			Prompt: prompt.Prompt,
		}
		if err = s.db.Create(&resp).Error; err != nil {
			logger.FromContext(c).WithError(err).Errorf("error creating prompt by type '%s'", promptType)
			response.Error(c, response.ErrInternal, err)
			return
		}

		response.Success(c, http.StatusCreated, resp)
		return
	} else if err != nil {
		logger.FromContext(c).WithError(err).Errorf("error finding updated prompt by type '%s'", promptType)
		response.Error(c, response.ErrInternal, err)
		return
	}

	resp.Prompt = prompt.Prompt

	err = s.db.Scopes(scope).Save(&resp).Error
	if err != nil {
		logger.FromContext(c).WithError(err).Errorf("error updating prompt by type '%s'", promptType)
		response.Error(c, response.ErrInternal, err)
		return
	}

	response.Success(c, http.StatusOK, resp)
}

// ResetPrompt is a function to reset prompt by type to default value
// @Summary Reset prompt by type to default value
// @Tags Prompts
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param promptType path string true "prompt type"
// @Success 200 {object} response.successResp{data=models.Prompt} "prompt reset successful"
// @Success 201 {object} response.successResp{data=models.Prompt} "prompt created with default value successful"
// @Failure 400 {object} response.errorResp "invalid prompt request data"
// @Failure 403 {object} response.errorResp "updating prompt not permitted"
// @Failure 404 {object} response.errorResp "prompt not found"
// @Failure 500 {object} response.errorResp "internal error on resetting prompt"
// @Router /prompts/{promptType}/default [post]
func (s *PromptService) ResetPrompt(c *gin.Context) {
	var (
		err        error
		promptType models.PromptType = models.PromptType(c.Param("promptType"))
		resp       models.Prompt
	)

	if err = promptType.Valid(); err != nil {
		logger.FromContext(c).WithError(err).Errorf("error validating prompt type '%s'", promptType)
		response.Error(c, response.ErrPromptsInvalidRequest, err)
		return
	}

	privs := c.GetStringSlice("prm")
	if !slices.Contains(privs, "settings.prompts.edit") {
		logger.FromContext(c).Errorf("error filtering user role permissions: permission not found")
		response.Error(c, response.ErrNotPermitted, nil)
		return
	}

	uid := c.GetUint64("uid")
	scope := func(db *gorm.DB) *gorm.DB {
		return db.Where("type = ? AND user_id = ?", promptType, uid)
	}

	template, err := s.prompter.GetTemplate(templates.PromptType(promptType))
	if err != nil {
		logger.FromContext(c).WithError(err).Errorf("error getting template '%s'", promptType)
		response.Error(c, response.ErrPromptsInvalidRequest, err)
		return
	}

	err = s.db.Scopes(scope).Take(&resp).Error
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		resp = models.Prompt{
			Type:   promptType,
			UserID: uid,
			Prompt: template,
		}
		err = s.db.Create(&resp).Error
		if err != nil {
			logger.FromContext(c).WithError(err).Errorf("error creating default prompt by type '%s'", promptType)
			response.Error(c, response.ErrInternal, err)
		}

		response.Success(c, http.StatusCreated, resp)
		return
	} else if err != nil {
		logger.FromContext(c).WithError(err).Errorf("error finding updated prompt by type '%s'", promptType)
		response.Error(c, response.ErrInternal, err)
		return
	}

	resp.Prompt = template

	err = s.db.Scopes(scope).Save(&resp).Error
	if err != nil {
		logger.FromContext(c).WithError(err).Errorf("error resetting prompt by type '%s'", promptType)
		response.Error(c, response.ErrInternal, err)
		return
	}

	response.Success(c, http.StatusOK, resp)
}

// DeletePrompt is a function to delete prompt by type
// @Summary Delete prompt by type
// @Tags Prompts
// @Produce json
// @Security BearerAuth
// @Param promptType path string true "prompt type"
// @Success 200 {object} response.successResp "prompt deleted successful"
// @Failure 400 {object} response.errorResp "invalid prompt request data"
// @Failure 403 {object} response.errorResp "deleting prompt not permitted"
// @Failure 404 {object} response.errorResp "prompt not found"
// @Failure 500 {object} response.errorResp "internal error on deleting prompt"
// @Router /prompts/{promptType} [delete]
func (s *PromptService) DeletePrompt(c *gin.Context) {
	var (
		err        error
		promptType models.PromptType = models.PromptType(c.Param("promptType"))
		resp       models.Prompt
	)

	if err = promptType.Valid(); err != nil {
		logger.FromContext(c).WithError(err).Errorf("error validating prompt type '%s'", promptType)
		response.Error(c, response.ErrPromptsInvalidRequest, err)
		return
	}

	privs := c.GetStringSlice("prm")
	if !slices.Contains(privs, "settings.prompts.edit") {
		logger.FromContext(c).Errorf("error filtering user role permissions: permission not found")
		response.Error(c, response.ErrNotPermitted, nil)
		return
	}

	uid := c.GetUint64("uid")
	scope := func(db *gorm.DB) *gorm.DB {
		return db.Where("type = ? AND user_id = ?", promptType, uid)
	}

	if err = s.db.Scopes(scope).Take(&resp).Error; err != nil {
		logger.FromContext(c).WithError(err).Errorf("error finding prompt by type '%s'", promptType)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Error(c, response.ErrPromptsNotFound, err)
		} else {
			response.Error(c, response.ErrInternal, err)
		}
		return
	}

	if err = s.db.Scopes(scope).Delete(&resp).Error; err != nil {
		logger.FromContext(c).WithError(err).Errorf("error deleting prompt by type '%s'", promptType)
		response.Error(c, response.ErrInternal, err)
		return
	}

	response.Success(c, http.StatusOK, nil)
}
