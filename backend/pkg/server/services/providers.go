package services

import (
	"net/http"
	"slices"

	"pentagi/pkg/providers"
	"pentagi/pkg/providers/pconfig"
	"pentagi/pkg/providers/provider"
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
// @Success 200 {object} response.successResp{data=[]models.ProviderInfo} "providers list received successful"
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
		prv := providers[name]
		providerInfos[i] = models.ProviderInfo{
			Name:         name.String(),
			Type:         models.ProviderType(prv.Type()),
			DefaultModel: prv.Model(pconfig.OptionsTypePrimaryAgent),
			Models:       buildModelInfos(prv),
		}
	}

	response.Success(c, http.StatusOK, providerInfos)
}

func buildModelInfos(prv provider.Provider) []models.ModelInfo {
	modelsConfig := prv.GetModels()

	// Build lookup: model name -> price from ModelsConfig (models.yml)
	modelConfigPrice := make(map[string]*pconfig.PriceInfo, len(modelsConfig))
	for _, mc := range modelsConfig {
		if mc.Price != nil {
			modelConfigPrice[mc.Name] = mc.Price
		}
	}

	// Collect unique models actually in use across all agent types.
	// For price priority: AgentConfig.Price > ModelsConfig price.
	type entry struct {
		price      *pconfig.PriceInfo
		agentTypes []string
	}
	seen := make(map[string]*entry)
	order := make([]string, 0)

	for _, optType := range pconfig.AllAgentTypes {
		modelName := prv.Model(optType)
		if modelName == "" {
			continue
		}

		e, exists := seen[modelName]
		if !exists {
			price := prv.GetPriceInfo(optType)
			if price == nil {
				price = modelConfigPrice[modelName]
			}
			e = &entry{price: price}
			seen[modelName] = e
			order = append(order, modelName)
		}

		e.agentTypes = append(e.agentTypes, string(optType))
	}

	modelInfos := make([]models.ModelInfo, 0, len(order))
	for _, name := range order {
		e := seen[name]
		mi := models.ModelInfo{
			Name:       name,
			AgentTypes: e.agentTypes,
		}
		if e.price != nil {
			mi.PriceInfo = &models.ModelPriceInfo{
				Input:      e.price.Input,
				Output:     e.price.Output,
				CacheRead:  e.price.CacheRead,
				CacheWrite: e.price.CacheWrite,
			}
		}
		modelInfos = append(modelInfos, mi)
	}

	return modelInfos
}
