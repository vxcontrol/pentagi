package services

import (
	"errors"
	"net/http"
	"time"

	"pentagi/pkg/database"
	"pentagi/pkg/graph/subscriptions"
	"pentagi/pkg/server/auth"
	"pentagi/pkg/server/logger"
	"pentagi/pkg/server/models"
	"pentagi/pkg/server/response"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

type tokens struct {
	Tokens []models.APIToken `json:"tokens"`
	Total  uint64            `json:"total"`
}

// TokenService handles API token management
type TokenService struct {
	db         *gorm.DB
	globalSalt string
	tokenCache *auth.TokenCache
	ss         subscriptions.SubscriptionsController
}

// NewTokenService creates a new TokenService instance
func NewTokenService(
	db *gorm.DB,
	globalSalt string,
	tokenCache *auth.TokenCache,
	ss subscriptions.SubscriptionsController,
) *TokenService {
	return &TokenService{
		db:         db,
		globalSalt: globalSalt,
		tokenCache: tokenCache,
		ss:         ss,
	}
}

// CreateToken creates a new API token
// @Summary Create new API token for automation
// @Tags Tokens
// @Accept json
// @Produce json
// @Param json body models.CreateAPITokenRequest true "Token creation request"
// @Success 201 {object} response.successResp{data=models.APITokenWithSecret} "token created successful"
// @Failure 400 {object} response.errorResp "invalid token request or default salt"
// @Failure 403 {object} response.errorResp "creating token not permitted"
// @Failure 500 {object} response.errorResp "internal error on creating token"
// @Router /tokens [post]
func (s *TokenService) CreateToken(c *gin.Context) {
	// check for default salt
	if s.globalSalt == "" || s.globalSalt == "salt" {
		logger.FromContext(c).Errorf("token creation attempted with default salt")
		response.Error(c, response.ErrTokenCreationDisabled, errors.New("token creation is disabled with default salt"))
		return
	}

	uid := c.GetUint64("uid")
	rid := c.GetUint64("rid")
	uhash := c.GetString("uhash")

	var req models.CreateAPITokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.FromContext(c).WithError(err).Errorf("error binding JSON")
		response.Error(c, response.ErrTokenInvalidRequest, err)
		return
	}
	if err := req.Valid(); err != nil {
		logger.FromContext(c).WithError(err).Errorf("error validating JSON")
		response.Error(c, response.ErrTokenInvalidRequest, err)
		return
	}

	// check if name is unique for this user (if provided)
	if req.Name != nil && *req.Name != "" {
		var existing models.APIToken
		err := s.db.
			Where("user_id = ? AND name = ? AND deleted_at IS NULL", uid, *req.Name).
			First(&existing).
			Error
		if err == nil {
			logger.FromContext(c).Errorf("token with name '%s' already exists for user %d", *req.Name, uid)
			response.Error(c, response.ErrTokenInvalidRequest, errors.New("token with this name already exists"))
			return
		}
	}

	// generate token_id
	tokenID, err := auth.GenerateTokenID()
	if err != nil {
		logger.FromContext(c).WithError(err).Errorf("error generating token ID")
		response.Error(c, response.ErrInternal, err)
		return
	}

	// create JWT claims
	claims := auth.MakeAPITokenClaims(tokenID, uhash, uid, rid, req.TTL)

	// sign token
	token, err := auth.MakeAPIToken(s.globalSalt, claims)
	if err != nil {
		logger.FromContext(c).WithError(err).Errorf("error signing token")
		response.Error(c, response.ErrInternal, err)
		return
	}

	// save to database
	apiToken := models.APIToken{
		TokenID: tokenID,
		UserID:  uid,
		RoleID:  rid,
		Name:    req.Name,
		TTL:     req.TTL,
		Status:  models.TokenStatusActive,
	}

	if err := s.db.Create(&apiToken).Error; err != nil {
		logger.FromContext(c).WithError(err).Errorf("error creating token in database")
		response.Error(c, response.ErrInternal, err)
		return
	}

	result := models.APITokenWithSecret{
		APIToken: apiToken,
		Token:    token,
	}

	// invalidate cache for negative caching results
	s.tokenCache.Invalidate(apiToken.TokenID)
	s.tokenCache.InvalidateUser(apiToken.UserID)

	if s.ss != nil {
		publisher := s.ss.NewFlowPublisher(int64(apiToken.UserID), 0)
		publisher.APITokenCreated(c, database.APITokenWithSecret{
			ApiToken: convertAPITokenToDatabase(apiToken),
			Token:    token,
		})
	}

	response.Success(c, http.StatusCreated, result)
}

// ListTokens returns a list of tokens (user sees only their own, admin sees all)
// @Summary List API tokens
// @Tags Tokens
// @Produce json
// @Success 200 {object} response.successResp{data=tokens} "tokens retrieved successful"
// @Failure 403 {object} response.errorResp "listing tokens not permitted"
// @Failure 500 {object} response.errorResp "internal error on listing tokens"
// @Router /tokens [get]
func (s *TokenService) ListTokens(c *gin.Context) {
	uid := c.GetUint64("uid")
	prms := c.GetStringSlice("prm")

	query := s.db.Where("deleted_at IS NULL")

	// check if user has admin privilege
	hasAdmin := auth.LookupPerm(prms, "settings.tokens.admin")
	if !hasAdmin {
		// regular user sees only their own tokens
		query = query.Where("user_id = ?", uid)
	}

	var tokenList []models.APIToken
	var total uint64

	if err := query.Order("created_at DESC").Find(&tokenList).Count(&total).Error; err != nil {
		logger.FromContext(c).WithError(err).Errorf("error finding tokens")
		response.Error(c, response.ErrInternal, err)
		return
	}

	for i := range tokenList {
		token := &tokenList[i]
		isExpired := token.CreatedAt.Add(time.Duration(token.TTL) * time.Second).Before(time.Now())
		if token.Status == models.TokenStatusActive && isExpired {
			token.Status = models.TokenStatusExpired
		}
	}

	result := tokens{
		Tokens: tokenList,
		Total:  total,
	}

	response.Success(c, http.StatusOK, result)
}

// GetToken returns information about a specific token
// @Summary Get API token details
// @Tags Tokens
// @Produce json
// @Param tokenID path string true "Token ID"
// @Success 200 {object} response.successResp{data=models.APIToken} "token retrieved successful"
// @Failure 403 {object} response.errorResp "accessing token not permitted"
// @Failure 404 {object} response.errorResp "token not found"
// @Failure 500 {object} response.errorResp "internal error on getting token"
// @Router /tokens/{tokenID} [get]
func (s *TokenService) GetToken(c *gin.Context) {
	uid := c.GetUint64("uid")
	prms := c.GetStringSlice("prm")
	tokenID := c.Param("tokenID")

	var token models.APIToken
	if err := s.db.Where("token_id = ? AND deleted_at IS NULL", tokenID).First(&token).Error; err != nil {
		logger.FromContext(c).WithError(err).Errorf("error finding token")
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Error(c, response.ErrTokenNotFound, err)
		} else {
			response.Error(c, response.ErrInternal, err)
		}
		return
	}

	// check authorization
	hasAdmin := auth.LookupPerm(prms, "settings.tokens.admin")
	if !hasAdmin && token.UserID != uid {
		logger.FromContext(c).Errorf("user %d attempted to access token of user %d", uid, token.UserID)
		response.Error(c, response.ErrTokenUnauthorized, errors.New("not authorized to access this token"))
		return
	}

	isExpired := token.CreatedAt.Add(time.Duration(token.TTL) * time.Second).Before(time.Now())
	if token.Status == models.TokenStatusActive && isExpired {
		token.Status = models.TokenStatusExpired
	}

	if err := token.Valid(); err != nil {
		logger.FromContext(c).WithError(err).Errorf("error validating token data")
		response.Error(c, response.ErrTokenInvalidData, err)
		return
	}

	response.Success(c, http.StatusOK, token)
}

// UpdateToken updates name and/or status of a token
// @Summary Update API token
// @Tags Tokens
// @Accept json
// @Produce json
// @Param tokenID path string true "Token ID"
// @Param json body models.UpdateAPITokenRequest true "Token update request"
// @Success 200 {object} response.successResp{data=models.APIToken} "token updated successful"
// @Failure 400 {object} response.errorResp "invalid update request"
// @Failure 403 {object} response.errorResp "updating token not permitted"
// @Failure 404 {object} response.errorResp "token not found"
// @Failure 500 {object} response.errorResp "internal error on updating token"
// @Router /tokens/{tokenID} [put]
func (s *TokenService) UpdateToken(c *gin.Context) {
	uid := c.GetUint64("uid")
	prms := c.GetStringSlice("prm")
	tokenID := c.Param("tokenID")

	var req models.UpdateAPITokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.FromContext(c).WithError(err).Errorf("error binding JSON")
		response.Error(c, response.ErrTokenInvalidRequest, err)
		return
	}
	if err := req.Valid(); err != nil {
		logger.FromContext(c).WithError(err).Errorf("error validating JSON")
		response.Error(c, response.ErrTokenInvalidRequest, err)
		return
	}

	var token models.APIToken
	if err := s.db.Where("token_id = ? AND deleted_at IS NULL", tokenID).First(&token).Error; err != nil {
		logger.FromContext(c).WithError(err).Errorf("error finding token")
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Error(c, response.ErrTokenNotFound, err)
		} else {
			response.Error(c, response.ErrInternal, err)
		}
		return
	}

	// check authorization
	hasAdmin := auth.LookupPerm(prms, "settings.tokens.admin")
	if !hasAdmin && token.UserID != uid {
		logger.FromContext(c).Errorf("user %d attempted to update token of user %d", uid, token.UserID)
		response.Error(c, response.ErrTokenUnauthorized, errors.New("not authorized to update this token"))
		return
	}

	// update fields
	updates := make(map[string]any)
	if req.Name != nil {
		// check uniqueness if name is changing
		if token.Name == nil || *token.Name != *req.Name {
			if *req.Name != "" {
				var existing models.APIToken
				err := s.db.
					Where("user_id = ? AND name = ? AND token_id != ? AND deleted_at IS NULL", token.UserID, *req.Name, tokenID).
					First(&existing).
					Error
				if err == nil {
					logger.FromContext(c).Errorf("token with name '%s' already exists for user %d", *req.Name, token.UserID)
					response.Error(c, response.ErrTokenInvalidRequest, errors.New("token with this name already exists"))
					return
				}
			}
		}
		updates["name"] = req.Name
	}
	switch req.Status {
	case models.TokenStatusActive:
		updates["status"] = models.TokenStatusActive
	case models.TokenStatusRevoked:
		updates["status"] = models.TokenStatusRevoked
	case models.TokenStatusExpired:
		updates["status"] = models.TokenStatusRevoked
	}

	if len(updates) > 0 {
		if err := s.db.Model(&token).Updates(updates).Error; err != nil {
			logger.FromContext(c).WithError(err).Errorf("error updating token")
			response.Error(c, response.ErrInternal, err)
			return
		}

		// invalidate cache if status changed
		if req.Status != "" {
			s.tokenCache.Invalidate(tokenID)
			// also invalidate all tokens for this user (in case of role change or security event)
			s.tokenCache.InvalidateUser(token.UserID)
		}

		// reload token
		if err := s.db.Where("token_id = ?", tokenID).First(&token).Error; err != nil {
			logger.FromContext(c).WithError(err).Errorf("error reloading token")
			response.Error(c, response.ErrInternal, err)
			return
		}
	}

	isExpired := token.CreatedAt.Add(time.Duration(token.TTL) * time.Second).Before(time.Now())
	if token.Status == models.TokenStatusActive && isExpired {
		token.Status = models.TokenStatusExpired
	}

	if s.ss != nil {
		publisher := s.ss.NewFlowPublisher(int64(token.UserID), 0)
		publisher.APITokenUpdated(c, convertAPITokenToDatabase(token))
	}

	response.Success(c, http.StatusOK, token)
}

// DeleteToken performs soft delete of a token
// @Summary Delete API token
// @Tags Tokens
// @Produce json
// @Param tokenID path string true "Token ID"
// @Success 200 {object} response.successResp "token deleted successful"
// @Failure 403 {object} response.errorResp "deleting token not permitted"
// @Failure 404 {object} response.errorResp "token not found"
// @Failure 500 {object} response.errorResp "internal error on deleting token"
// @Router /tokens/{tokenID} [delete]
func (s *TokenService) DeleteToken(c *gin.Context) {
	uid := c.GetUint64("uid")
	prms := c.GetStringSlice("prm")
	tokenID := c.Param("tokenID")

	var token models.APIToken
	if err := s.db.Where("token_id = ? AND deleted_at IS NULL", tokenID).First(&token).Error; err != nil {
		logger.FromContext(c).WithError(err).Errorf("error finding token")
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Error(c, response.ErrTokenNotFound, err)
		} else {
			response.Error(c, response.ErrInternal, err)
		}
		return
	}

	// check authorization
	hasAdmin := auth.LookupPerm(prms, "settings.tokens.admin")
	if !hasAdmin && token.UserID != uid {
		logger.FromContext(c).Errorf("user %d attempted to delete token of user %d", uid, token.UserID)
		response.Error(c, response.ErrTokenUnauthorized, errors.New("not authorized to delete this token"))
		return
	}

	// soft delete
	if err := s.db.Delete(&token).Error; err != nil {
		logger.FromContext(c).WithError(err).Errorf("error deleting token")
		response.Error(c, response.ErrInternal, err)
		return
	}

	// invalidate cache for this token and all user's tokens
	s.tokenCache.Invalidate(tokenID)
	s.tokenCache.InvalidateUser(token.UserID)

	if s.ss != nil {
		publisher := s.ss.NewFlowPublisher(int64(token.UserID), 0)
		publisher.APITokenDeleted(c, convertAPITokenToDatabase(token))
	}

	response.Success(c, http.StatusOK, gin.H{"message": "token deleted successfully"})
}

func convertAPITokenToDatabase(apiToken models.APIToken) database.ApiToken {
	return database.ApiToken{
		ID:        int64(apiToken.ID),
		TokenID:   apiToken.TokenID,
		UserID:    int64(apiToken.UserID),
		RoleID:    int64(apiToken.RoleID),
		Name:      database.StringToNullString(*apiToken.Name),
		Ttl:       int64(apiToken.TTL),
		Status:    database.TokenStatus(apiToken.Status),
		CreatedAt: database.TimeToNullTime(apiToken.CreatedAt),
		UpdatedAt: database.TimeToNullTime(apiToken.UpdatedAt),
		DeletedAt: database.PtrTimeToNullTime(apiToken.DeletedAt),
	}
}
