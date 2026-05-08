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
