package services

import (
	"net/http"
	"slices"

	"pentagi/pkg/server/logger"
	"pentagi/pkg/server/response"

	"github.com/gin-gonic/gin"
	"github.com/vxcontrol/cloud/anonymizer"
)

type anonymizeTextRequest struct {
	Text string `json:"text" binding:"required"`
}

type anonymizeTextResponse struct {
	Text string `json:"text"`
}

// AnonymizerService handles text anonymization via the REST API.
type AnonymizerService struct {
	replacer anonymizer.Replacer
}

func NewAnonymizerService(replacer anonymizer.Replacer) *AnonymizerService {
	return &AnonymizerService{replacer: replacer}
}

// AnonymizeText replaces sensitive data patterns in the provided text.
// @Summary Anonymize text
// @Description Replace sensitive data patterns (credentials, keys, PII, etc.) in the provided text with safe placeholders.
// @Tags Anonymize
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param json body anonymizeTextRequest true "Text to anonymize"
// @Success 200 {object} response.successResp{data=anonymizeTextResponse} "anonymized text returned successfully"
// @Failure 400 {object} response.errorResp "invalid request body"
// @Failure 403 {object} response.errorResp "anonymize.call permission required"
// @Failure 503 {object} response.errorResp "anonymizer is not configured"
// @Router /anonymize/text [post]
func (s *AnonymizerService) AnonymizeText(c *gin.Context) {
	privs := c.GetStringSlice("prm")
	if !slices.Contains(privs, "anonymize.call") {
		logger.FromContext(c).Errorf("error filtering user role permissions: permission not found")
		response.Error(c, response.ErrNotPermitted, nil)
		return
	}

	if s.replacer == nil {
		response.Error(c, response.ErrAnonymizeUnavailable, nil)
		return
	}

	var req anonymizeTextRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.FromContext(c).WithError(err).Error("error binding anonymize request")
		response.Error(c, response.ErrAnonymizeInvalidRequest, err)
		return
	}

	response.Success(c, http.StatusOK, anonymizeTextResponse{
		Text: s.replacer.ReplaceString(req.Text),
	})
}
