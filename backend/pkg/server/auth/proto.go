package auth

import (
	"fmt"
	"net/http"
	"time"

	"pentagi/pkg/server/logger"
	"pentagi/pkg/server/models"
	"pentagi/pkg/server/response"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type ProtoService struct {
	globalSalt string
}

func NewProtoService(globalSalt string) *ProtoService {
	return &ProtoService{
		globalSalt: globalSalt,
	}
}

// CreateAuthToken is a function to create new JWT token to authorize automation requests
// @Summary Create new JWT token to use it into automation connections
// @Tags Proto
// @Accept json
// @Produce json
// @Param json body models.ProtoAuthTokenRequest true "Proto auth token request JSON data"
// @Success 201 {object} response.successResp{data=models.ProtoAuthToken} "token created successful"
// @Failure 400 {object} response.errorResp "invalid requested token info"
// @Failure 403 {object} response.errorResp "creating token not permitted"
// @Failure 500 {object} response.errorResp "internal error on creating token"
// @Router /token [post]
func (p *ProtoService) CreateAuthToken(c *gin.Context) {
	var req models.ProtoAuthTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.FromContext(c).WithError(err).Errorf("error binding JSON")
		response.Error(c, response.ErrProtoInvalidRequest, err)
		return
	}
	if err := req.Valid(); err != nil {
		logger.FromContext(c).WithError(err).Errorf("error validating JSON")
		response.Error(c, response.ErrProtoInvalidRequest, err)
		return
	}

	token, err := p.MakeToken(c, &req)
	if err != nil {
		logger.FromContext(c).WithError(err).Errorf("error on making token")
		response.Error(c, response.ErrProtoCreateTokenFail, err)
		return
	}
	if _, err = ValidateToken(token, p.globalSalt); err != nil {
		logger.FromContext(c).WithError(err).Errorf("error on validating token")
		response.Error(c, response.ErrProtoInvalidToken, err)
		return
	}

	pat := models.ProtoAuthToken{
		Token:       token,
		TTL:         req.TTL,
		CreatedDate: time.Now(),
	}
	response.Success(c, http.StatusCreated, pat)
}

func (p *ProtoService) MakeToken(c *gin.Context, req *models.ProtoAuthTokenRequest) (string, error) {
	claims, err := p.makeTokenClaims(c, req.Type)
	if err != nil {
		return "", fmt.Errorf("failed to get token claims: %w", err)
	}

	now := time.Now()
	claims.RegisteredClaims = jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(now.Add(time.Duration(req.TTL) * time.Second)),
		IssuedAt:  jwt.NewNumericDate(now),
		Subject:   "automation",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(MakeCookieStoreKey(p.globalSalt)[1])
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}
	return tokenString, nil
}

func (p *ProtoService) makeTokenClaims(c *gin.Context, cpt string) (*models.ProtoAuthTokenClaims, error) {
	rid := c.GetUint64("rid")
	if rid == 0 {
		return nil, fmt.Errorf("input RID invalid %d", rid)
	}

	uid := c.GetUint64("uid")
	if uid == 0 {
		return nil, fmt.Errorf("input UID invalid %d", uid)
	}

	tid := c.GetString("tid")
	if tid == "" {
		return nil, fmt.Errorf("input TID invalid %s", tid)
	}

	uhash := c.GetString("uhash")
	if uid == 0 {
		return nil, fmt.Errorf("input UHASH invalid %d", uid)
	}

	return &models.ProtoAuthTokenClaims{
		RID:   rid,
		UID:   uid,
		TID:   tid,
		UHASH: uhash,
		CPT:   cpt,
	}, nil
}
