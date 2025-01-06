package auth

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"pentagi/pkg/server/models"
	"pentagi/pkg/server/response"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
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
}

func NewAuthMiddleware(baseURL, globalSalt string) *AuthMiddleware {
	return &AuthMiddleware{
		globalSalt: globalSalt,
	}
}

func (p *AuthMiddleware) AuthRequired(c *gin.Context) {
	p.tryAuth(c, true, p.tryUserCookieAuthentication)
}

func (p *AuthMiddleware) AuthTokenProtoRequired(c *gin.Context) {
	p.tryAuth(c, true, p.tryProtoTokenAuthentication, p.tryUserCookieAuthentication)
}

func (p *AuthMiddleware) TryAuth(c *gin.Context) {
	p.tryAuth(c, false, p.tryUserCookieAuthentication)
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

	for _, attr := range []interface{}{uid, rid, prm, exp, gtm, uname, tid} {
		if attr == nil {
			return authResultFail, errors.New("token claim invalid")
		}
	}

	prms, ok := prm.([]string)
	if !ok {
		return authResultFail, errors.New("no pemissions granted")
	}

	c.Set("prm", prms)
	c.Set("uid", uid.(uint64))
	c.Set("uhash", uhash.(string))
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

	claims, err := ValidateToken(token, p.globalSalt)
	if err != nil {
		return authResultFail, errors.New("token is invalid")
	}

	c.Set("uid", claims.UID)
	c.Set("tid", claims.TID)
	c.Set("uhash", claims.UHASH)
	c.Set("rid", claims.RID)
	c.Set("cpt", claims.CPT)
	c.Set("prm", []string{PrivilegeAutomation})

	c.Next()

	return authResultOk, nil
}

func ValidateToken(tokenString, globalSalt string) (*models.ProtoAuthTokenClaims, error) {
	var claims models.ProtoAuthTokenClaims
	token, err := jwt.ParseWithClaims(tokenString, &claims, func(token *jwt.Token) (interface{}, error) {
		return MakeCookieStoreKey(globalSalt)[1], nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenMalformed) {
			return nil, fmt.Errorf("token is malformed")
		} else if errors.Is(err, jwt.ErrTokenExpired) || errors.Is(err, jwt.ErrTokenNotValidYet) {
			return nil, fmt.Errorf("token is either expired or not active yet")
		} else {
			return nil, fmt.Errorf("token invalid: %w", err)
		}
	}

	if !token.Valid {
		return nil, fmt.Errorf("token is invalid")
	}

	return &claims, nil
}
