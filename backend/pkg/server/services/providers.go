package services

import (
	"net/http"
	"slices"

	"pentagi/pkg/providers"
	"pentagi/pkg/providers/provider"
	"pentagi/pkg/server/logger"
	"pentagi/pkg/server/response"

	"github.com/gin-gonic/gin"
)

type ProviderService struct {
	providers providers.ProviderController
}

func NewProviderService(providers providers.ProviderController) *ProviderService {
	return &ProviderService{
		providers: providers,
	}
}

// GetProviders is a function to return providers list
// @Summary Retrieve providers list
// @Tags Providers
// @Produce json
// @Success 200 {object} response.successResp{data=provider.ProvidersList} "providers list received successful"
// @Failure 403 {object} response.errorResp "getting providers not permitted"
// @Router /providers/ [get]
func (s *ProviderService) GetProviders(c *gin.Context) {
	privs := c.GetStringSlice("prm")
	if !slices.Contains(privs, "providers.view") {
		logger.FromContext(c).Errorf("error filtering user role permissions: permission not found")
		response.Error(c, response.ErrNotPermitted, nil)
		return
	}

	var list provider.ProvidersList
	list = s.providers.List()

	response.Success(c, http.StatusOK, list)
}
