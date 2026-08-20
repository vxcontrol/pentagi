package services

import (
	"net/http"
	"slices"

	"pentagi/pkg/config"
	"pentagi/pkg/server/logger"
	"pentagi/pkg/server/models"
	"pentagi/pkg/server/response"
	"pentagi/pkg/version"

	"github.com/gin-gonic/gin"
)

type SettingsService struct {
	cfg *config.Config
}

func NewSettingsService(cfg *config.Config) *SettingsService {
	return &SettingsService{cfg: cfg}
}

// GetSettings is a function to return settings
// @Summary Retrieve settings
// @Tags Settings
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.successResp{data=models.Settings} "settings received successful"
// @Failure 403 {object} response.errorResp "getting settings not permitted"
// @Router /settings/ [get]
func (s *SettingsService) GetSettings(c *gin.Context) {
	privs := c.GetStringSlice("prm")
	if !slices.Contains(privs, "settings.view") {
		logger.FromContext(c).Errorf("error filtering user role permissions: permission not found")
		response.Error(c, response.ErrNotPermitted, nil)
		return
	}

	settings := models.Settings{
		Debug:              s.cfg.Debug,
		AskUser:            s.cfg.AskUser,
		Version:            version.GetBinaryVersion(),
		DockerInside:       s.cfg.DockerInside,
		IsDevelopMode:      version.IsDevelopMode(),
		AssistantUseAgents: s.cfg.AssistantUseAgents,
	}

	response.Success(c, http.StatusOK, settings)
}

func (s *SettingsService) deafGuardConfig() models.DeafGuardConfig {
	mode := s.cfg.DeafGuardMode
	if mode == "" {
		mode = "log"
	}
	return models.DeafGuardConfig{
		Enabled: s.cfg.DeafGuardEnabled,
		Mode:    mode,
	}
}

// GetDeafGuardConfig returns the process-wide Deaf Guard configuration.
// @Summary Retrieve Deaf Guard configuration
// @Tags Settings
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.successResp{data=models.DeafGuardConfig} "deaf guard config received successful"
// @Failure 403 {object} response.errorResp "getting deaf guard config not permitted"
// @Router /deafguard/config [get]
func (s *SettingsService) GetDeafGuardConfig(c *gin.Context) {
	privs := c.GetStringSlice("prm")
	if !slices.Contains(privs, "settings.view") {
		logger.FromContext(c).Errorf("error filtering user role permissions: permission not found")
		response.Error(c, response.ErrNotPermitted, nil)
		return
	}

	response.Success(c, http.StatusOK, s.deafGuardConfig())
}

// PutDeafGuardConfig updates the process-wide Deaf Guard configuration.
// Changes take effect on the next flow; already-running flows keep their snapshot.
// @Summary Update Deaf Guard configuration
// @Tags Settings
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.successResp{data=models.DeafGuardConfig} "deaf guard config updated successful"
// @Failure 400 {object} response.errorResp "invalid deaf guard request"
// @Failure 403 {object} response.errorResp "updating deaf guard config not permitted"
// @Router /deafguard/config [put]
func (s *SettingsService) PutDeafGuardConfig(c *gin.Context) {
	privs := c.GetStringSlice("prm")
	if !slices.Contains(privs, "settings.view") {
		logger.FromContext(c).Errorf("error filtering user role permissions: permission not found")
		response.Error(c, response.ErrNotPermitted, nil)
		return
	}

	var req models.DeafGuardConfigUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.FromContext(c).WithError(err).Errorf("error binding JSON")
		response.Error(c, response.ErrDeafGuardInvalidRequest, err)
		return
	}

	if req.Enabled != nil {
		s.cfg.DeafGuardEnabled = *req.Enabled
	}
	if req.Mode != nil {
		switch *req.Mode {
		case "log", "warn", "enforce":
			s.cfg.DeafGuardMode = *req.Mode
		default:
			response.Error(c, response.ErrDeafGuardInvalidRequest, nil)
			return
		}
	}

	response.Success(c, http.StatusOK, s.deafGuardConfig())
}
