package services

import (
	"net/http"
	"slices"

	"pentagi/pkg/providers"
	"pentagi/pkg/server/logger"
	"pentagi/pkg/server/models"
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
// @Security BearerAuth
// @Success 200 {object} response.successResp{data=models.ProviderInfo} "providers list received successful"
// @Failure 403 {object} response.errorResp "getting providers not permitted"
// @Router /providers/ [get]
func (s *ProviderService) GetProviders(c *gin.Context) {
	privs := c.GetStringSlice("prm")
	if !slices.Contains(privs, "providers.view") {
		logger.FromContext(c).Errorf("error filtering user role permissions: permission not found")
		response.Error(c, response.ErrNotPermitted, nil)
		return
	}

	providers, err := s.providers.GetProviders(c, int64(c.GetUint64("uid")))
	if err != nil {
		logger.FromContext(c).Errorf("error getting providers: %v", err)
		response.Error(c, response.ErrInternal, nil)
		return
	}

	providerInfos := make([]models.ProviderInfo, len(providers))
	for i, name := range providers.ListNames() {
		providerInfos[i] = models.ProviderInfo{
			Name: name.String(),
			Type: models.ProviderType(providers[name].Type()),
		}
	}

	response.Success(c, http.StatusOK, providerInfos)
}
