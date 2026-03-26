package auth

import (
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"pentagi/pkg/server/models"
	"pentagi/pkg/server/rdb"
	"pentagi/pkg/server/response"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

type authResult int

const (
	authResultOk authResult = iota
	authResultSkip
	authResultFail
	authResultAbort
)

type AuthMiddleware struct {
	globalSalt string
	tokenCache *TokenCache
	userCache  *UserCache
}

func NewAuthMiddleware(baseURL, globalSalt string, tokenCache *TokenCache, userCache *UserCache) *AuthMiddleware {
	return &AuthMiddleware{
		globalSalt: globalSalt,
		tokenCache: tokenCache,
		userCache:  userCache,
	}
}

func (p *AuthMiddleware) AuthUserRequired(c *gin.Context) {
	p.tryAuth(c, true, p.tryUserCookieAuthentication)
}

func (p *AuthMiddleware) AuthTokenRequired(c *gin.Context) {
	p.tryAuth(c, true, p.tryProtoTokenAuthentication, p.tryUserCookieAuthentication)
}

func (p *AuthMiddleware) TryAuth(c *gin.Context) {
	p.tryAuth(c, false, p.tryProtoTokenAuthentication, p.tryUserCookieAuthentication)
}

func (p *AuthMiddleware) tryAuth(
	c *gin.Context,
	withFail bool,
	authMethods ...func(c *gin.Context) (authResult, error),
) {
	if c.IsAborted() {
		return
	}

	result := authResultSkip
	var authErr error
	for _, authMethod := range authMethods {
		result, authErr = authMethod(c)
		if c.IsAborted() || result == authResultAbort {
			return
		}
		if result != authResultSkip {
			break
		}
	}

	if withFail && result != authResultOk {
		response.Error(c, response.ErrAuthRequired, authErr)
		return
	}
	c.Next()
}

func (p *AuthMiddleware) tryUserCookieAuthentication(c *gin.Context) (authResult, error) {
	sessionObject, exists := c.Get(sessions.DefaultKey)
	if !exists {
		return authResultSkip, errors.New("can't find session object")
	}

	session, ok := sessionObject.(sessions.Session)
	if !ok {
		return authResultFail, errors.New("not a session object")
	}

	uid := session.Get("uid")
	uhash := session.Get("uhash")
	rid := session.Get("rid")
	prm := session.Get("prm")
	exp := session.Get("exp")
	gtm := session.Get("gtm")
	tid := session.Get("tid")
	uname := session.Get("uname")

	for _, attr := range []any{uid, rid, prm, exp, gtm, uname, uhash, tid} {
		if attr == nil {
			return authResultFail, errors.New("cookie claim invalid")
		}
	}

	prms, ok := prm.([]string)
	if !ok {
		return authResultFail, errors.New("no permissions granted")
	}

	// Verify session expiration
	expVal, ok := exp.(int64)
	if !ok {
		return authResultFail, errors.New("token claim invalid")
	}
	if time.Now().Unix() > expVal {
		return authResultFail, errors.New("session expired")
	}

	// Verify user hash matches database
	userID := uid.(uint64)
	sessionHash := uhash.(string)

	dbHash, userStatus, err := p.userCache.GetUserHash(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return authResultFail, errors.New("user has been deleted")
		}
		return authResultFail, fmt.Errorf("error checking user status: %w", err)
	}

	switch userStatus {
	case models.UserStatusBlocked:
		return authResultFail, errors.New("user has been blocked")
	case models.UserStatusCreated:
		return authResultFail, errors.New("user is not ready")
	case models.UserStatusActive:
	}

	if dbHash != sessionHash {
		return authResultFail, errors.New("user hash mismatch - session invalid for this installation")
	}

	c.Set("prm", prms)
	c.Set("uid", userID)
	c.Set("uhash", sessionHash)
	c.Set("rid", rid.(uint64))
	c.Set("exp", exp.(int64))
	c.Set("gtm", gtm.(int64))
	c.Set("tid", tid.(string))
	c.Set("uname", uname.(string))

	if slices.Contains(prms, PrivilegeAutomation) {
		c.Set("cpt", "automation")
	}

	return authResultOk, nil
}

const PrivilegeAutomation = "pentagi.automation"

func (p *AuthMiddleware) tryProtoTokenAuthentication(c *gin.Context) (authResult, error) {
	authHeader := c.Request.Header.Get("Authorization")
	if authHeader == "" {
		return authResultSkip, errors.New("token required")
	}

	if !strings.HasPrefix(authHeader, "Bearer ") {
		return authResultSkip, errors.New("bearer scheme must be used")
	}
	token := authHeader[7:]
	if token == "" {
		return authResultSkip, errors.New("token can't be empty")
	}

	// skip validation if using default salt (for backward compatibility)
	if p.globalSalt == "" || p.globalSalt == "salt" {
		return authResultSkip, errors.New("token validation disabled with default salt")
	}

	// try to validate as API token first (new format with JWT signing key)
	apiClaims, apiErr := ValidateAPIToken(token, p.globalSalt)
	if apiErr != nil {
		return authResultFail, errors.New("token is invalid")
	}

	// check token status and get privileges through cache
	status, privileges, err := p.tokenCache.GetStatus(apiClaims.TokenID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return authResultFail, errors.New("token not found in database")
		}
		return authResultFail, fmt.Errorf("error checking token status: %w", err)
	}
	if status != models.TokenStatusActive {
		return authResultFail, errors.New("token has been revoked")
	}

	// Verify user hash matches database
	dbHash, userStatus, err := p.userCache.GetUserHash(apiClaims.UID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return authResultFail, errors.New("user has been deleted")
		}
		return authResultFail, fmt.Errorf("error checking user status: %w", err)
	}

	if userStatus == models.UserStatusBlocked {
		return authResultFail, errors.New("user has been blocked")
	}

	if dbHash != apiClaims.UHASH {
		return authResultFail, errors.New("user hash mismatch - token invalid for this installation")
	}

	// generate UUID from user hash (fallback to empty string if hash is invalid)
	uuid, err := rdb.MakeUuidStrFromHash(apiClaims.UHASH)
	if err != nil {
		// Use empty UUID for invalid hashes (e.g., in tests)
		uuid = ""
	}

	// set session fields similar to regular login
	c.Set("uid", apiClaims.UID)
	c.Set("uhash", apiClaims.UHASH)
	c.Set("rid", apiClaims.RID)
	c.Set("tid", models.UserTypeAPI.String())
	c.Set("prm", privileges)
	c.Set("gtm", time.Now().Unix())
	c.Set("exp", apiClaims.ExpiresAt.Unix())
	c.Set("uuid", uuid)
	c.Set("cpt", "automation")

	return authResultOk, nil
}
