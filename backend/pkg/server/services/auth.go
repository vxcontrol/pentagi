package services

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"slices"
	"strconv"
	"strings"
	"time"

	"pentagi/pkg/server/logger"
	"pentagi/pkg/server/models"
	"pentagi/pkg/server/oauth"
	"pentagi/pkg/server/rdb"
	"pentagi/pkg/server/response"
	"pentagi/pkg/version"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/oauth2"
)

const (
	authStateCookieName = "state"
	authNonceCookieName = "nonce"
	authStateRequestTTL = 5 * time.Minute
)

type AuthServiceConfig struct {
	BaseURL          string
	LoginCallbackURL string
	SessionTimeout   int // in seconds
}

type AuthService struct {
	cfg   AuthServiceConfig
	db    *gorm.DB
	key   []byte
	oauth map[string]oauth.OAuthClient
}

func NewAuthService(
	cfg AuthServiceConfig,
	db *gorm.DB,
	oauth map[string]oauth.OAuthClient,
) *AuthService {
	var count int
	err := db.Model(&models.User{}).Where("type = 'local'").Count(&count).Error
	if err != nil {
		logrus.WithError(err).Errorf("error getting local users count")
	}

	key, err := randBytes(32)
	if err != nil {
		logrus.WithError(err).Errorf("error generating key")
	}

	return &AuthService{
		cfg:   cfg,
		db:    db,
		key:   key,
		oauth: oauth,
	}
}

// AuthLogin is function to login user in the system
// @Summary Login user into system
// @Tags Public
// @Accept json
// @Produce json
// @Param json body models.Login true "Login form JSON data"
// @Success 200 {object} response.successResp "login successful"
// @Failure 400 {object} response.errorResp "invalid login data"
// @Failure 401 {object} response.errorResp "invalid login or password"
// @Failure 403 {object} response.errorResp "login not permitted"
// @Failure 500 {object} response.errorResp "internal error on login"
// @Router /auth/login [post]
func (s *AuthService) AuthLogin(c *gin.Context) {
	var data models.Login
	if err := c.ShouldBindJSON(&data); err != nil || data.Valid() != nil {
		if err == nil {
			err = data.Valid()
		}
		logger.FromContext(c).WithError(err).Errorf("error validating request data")
		response.Error(c, response.ErrAuthInvalidLoginRequest, err)
		return
	}

	var user models.UserPassword
	if err := s.db.Take(&user, "mail = ? AND password IS NOT NULL", data.Mail).Error; err != nil {
		logrus.WithError(err).Errorf("error getting user by mail '%s'", data.Mail)
		response.Error(c, response.ErrAuthInvalidCredentials, err)
		return
	} else if err = user.Valid(); err != nil {
		logger.FromContext(c).WithError(err).Errorf("error validating user data '%s'", user.Hash)
		response.Error(c, response.ErrAuthInvalidUserData, err)
		return
	} else if user.RoleID == 100 {
		logger.FromContext(c).WithError(err).Errorf("can't authorize external user '%s'", user.Hash)
		response.Error(c, response.ErrAuthInvalidUserData, fmt.Errorf("user is external"))
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(data.Password)); err != nil {
		logger.FromContext(c).Errorf("error matching user input password")
		response.Error(c, response.ErrAuthInvalidCredentials, err)
		return
	}

	if user.Status != "active" {
		logger.FromContext(c).Errorf("error checking active state for user '%s'", user.Status)
		response.Error(c, response.ErrAuthInactiveUser, fmt.Errorf("user is inactive"))
		return
	}

	var privs []string
	err := s.db.Table("privileges").
		Where("role_id = ?", user.RoleID).
		Pluck("name", &privs).Error
	if err != nil {
		logger.FromContext(c).WithError(err).Errorf("error getting user privileges list '%s'", user.Hash)
		response.Error(c, response.ErrAuthInvalidServiceData, err)
		return
	}

	uuid, err := rdb.MakeUuidStrFromHash(user.Hash)
	if err != nil {
		logger.FromContext(c).WithError(err).Errorf("error validating user data '%s'", user.Hash)
		response.Error(c, response.ErrAuthInvalidUserData, err)
		return
	}

	expires := s.cfg.SessionTimeout
	session := sessions.Default(c)
	session.Set("uid", user.ID)
	session.Set("uhash", user.Hash)
	session.Set("rid", user.RoleID)
	session.Set("tid", models.UserTypeLocal.String())
	session.Set("prm", privs)
	session.Set("gtm", time.Now().Unix())
	session.Set("exp", time.Now().Add(time.Duration(expires)*time.Second).Unix())
	session.Set("uuid", uuid)
	session.Set("uname", user.Name)
	session.Options(sessions.Options{
		HttpOnly: true,
		Secure:   c.Request.TLS != nil,
		SameSite: http.SameSiteLaxMode,
		Path:     s.cfg.BaseURL,
		MaxAge:   int(expires),
	})
	if err := session.Save(); err != nil {
		logger.FromContext(c).WithError(err).Errorf("error saving session")
		response.Error(c, response.ErrInternal, err)
		return
	}

	logger.FromContext(c).
		WithFields(logrus.Fields{
			"age":   expires,
			"uid":   user.ID,
			"uhash": user.Hash,
			"rid":   user.RoleID,
			"tid":   session.Get("tid"),
			"gtm":   session.Get("gtm"),
			"exp":   session.Get("exp"),
			"prm":   session.Get("prm"),
		}).
		Infof("user made successful local login for '%s'", data.Mail)

	response.Success(c, http.StatusOK, struct{}{})
}

func (s *AuthService) refreshCookie(c *gin.Context, resp *info, privs []string) error {
	session := sessions.Default(c)
	expires := int(s.cfg.SessionTimeout)
	session.Set("prm", privs)
	session.Set("gtm", time.Now().Unix())
	session.Set("exp", time.Now().Add(time.Duration(expires)*time.Second).Unix())
	resp.Privs = privs

	session.Set("uid", resp.User.ID)
	session.Set("uhash", resp.User.Hash)
	session.Set("rid", resp.User.RoleID)
	session.Set("tid", resp.User.Type.String())
	session.Options(sessions.Options{
		HttpOnly: true,
		Secure:   c.Request.TLS != nil,
		SameSite: http.SameSiteLaxMode,
		Path:     s.cfg.BaseURL,
		MaxAge:   expires,
	})
	if err := session.Save(); err != nil {
		logger.FromContext(c).WithError(err).Errorf("error saving session")
		return err
	}

	logger.FromContext(c).
		WithFields(logrus.Fields{
			"age":   expires,
			"uid":   resp.User.ID,
			"uhash": resp.User.Hash,
			"rid":   resp.User.RoleID,
			"tid":   session.Get("tid"),
			"gtm":   session.Get("gtm"),
			"exp":   session.Get("exp"),
			"prm":   session.Get("prm"),
		}).
		Infof("session was refreshed for '%s' '%s'", resp.User.Mail, resp.User.Name)

	return nil
}

// AuthAuthorize is function to login user in OAuth2 external system
// @Summary Login user into OAuth2 external system via HTTP redirect
// @Tags Public
// @Produce json
// @Param return_uri query string false "URI to redirect user there after login" default(/)
// @Param provider query string false "OAuth provider name (google, github, etc.)" default(google) enums:"google,github"
// @Success 307 "redirect to SSO login page"
// @Failure 400 {object} response.errorResp "invalid autorizarion query"
// @Failure 403 {object} response.errorResp "authorize not permitted"
// @Failure 500 {object} response.errorResp "internal error on autorizarion"
// @Router /auth/authorize [get]
func (s *AuthService) AuthAuthorize(c *gin.Context) {
	stateData := map[string]string{
		"exp": strconv.FormatInt(time.Now().Add(authStateRequestTTL).Unix(), 10),
	}

	queryReturnURI := c.Query("return_uri")
	if queryReturnURI != "" {
		returnURL, err := url.Parse(queryReturnURI)
		if err != nil {
			logger.FromContext(c).WithError(err).Errorf("failed to parse return url argument '%s'", queryReturnURI)
			response.Error(c, response.ErrAuthInvalidAuthorizeQuery, err)
			return
		}
		returnURL.Path = path.Clean(path.Join("/", returnURL.Path))
		stateData["return_uri"] = returnURL.RequestURI()
	}

	provider := c.Query("provider")
	oauthClient, ok := s.oauth[provider]
	if !ok {
		logger.FromContext(c).Errorf("external OAuth2 provider '%s' is not initialized", provider)
		err := fmt.Errorf("provider not initialized")
		response.Error(c, response.ErrNotPermitted, err)
		return
	}
	stateData["provider"] = provider

	stateUniq, err := randBase64String(16)
	if err != nil {
		logger.FromContext(c).WithError(err).Errorf("failed to generate state random data")
		response.Error(c, response.ErrInternal, err)
		return
	}
	stateData["uniq"] = stateUniq

	nonce, err := randBase64String(16)
	if err != nil {
		logger.FromContext(c).WithError(err).Errorf("failed to generate nonce random data")
		response.Error(c, response.ErrInternal, err)
		return
	}

	stateJSON, err := json.Marshal(stateData)
	if err != nil {
		logger.FromContext(c).WithError(err).Errorf("failed to marshal state json data")
		response.Error(c, response.ErrInternal, err)
		return
	}
	mac := hmac.New(sha256.New, s.key)
	mac.Write(stateJSON)
	signature := mac.Sum(nil)

	signedStateJSON := append(signature, stateJSON...)
	state := base64.RawURLEncoding.EncodeToString(signedStateJSON)

	// Google OAuth uses POST callback which requires SameSite=None for cross-site requests
	// GitHub and other providers use GET callback which works with SameSite=Lax
	sameSiteMode := http.SameSiteLaxMode
	if provider == "google" {
		sameSiteMode = http.SameSiteNoneMode
	}

	maxAge := int(authStateRequestTTL / time.Second)
	s.setCallbackCookie(c.Writer, c.Request, authStateCookieName, state, maxAge, sameSiteMode)
	s.setCallbackCookie(c.Writer, c.Request, authNonceCookieName, nonce, maxAge, sameSiteMode)

	authOpts := []oauth2.AuthCodeOption{
		oauth2.SetAuthURLParam("nonce", nonce),
		oauth2.SetAuthURLParam("response_mode", "form_post"),
		oauth2.SetAuthURLParam("response_type", "code id_token"),
	}
	http.Redirect(c.Writer, c.Request,
		oauthClient.AuthCodeURL(state, authOpts...),
		http.StatusTemporaryRedirect)
}

// AuthLoginGetCallback is function to catch login callback from OAuth application with code only
// @Summary Login user from external OAuth application
// @Tags Public
// @Accept json
// @Produce json
// @Param code query string false "Auth code from OAuth provider to exchange token"
// @Success 303 "redirect to registered return_uri path in the state"
// @Failure 400 {object} response.errorResp "invalid login data"
// @Failure 401 {object} response.errorResp "invalid login or password"
// @Failure 403 {object} response.errorResp "login not permitted"
// @Failure 500 {object} response.errorResp "internal error on login"
// @Router /auth/login-callback [get]
func (s *AuthService) AuthLoginGetCallback(c *gin.Context) {
	code := c.Query("code")
	if code == "" {
		response.Error(c, response.ErrAuthInvalidLoginCallbackRequest, fmt.Errorf("code is required"))
		return
	}

	state, err := c.Request.Cookie(authStateCookieName)
	if err != nil {
		logger.FromContext(c).WithError(err).Errorf("error getting state from cookie")
		response.Error(c, response.ErrAuthInvalidAuthorizationState, err)
		return
	}

	queryState := c.Query("state")
	if queryState == "" {
		logger.FromContext(c).Errorf("error missing state parameter in OAuth callback")
		response.Error(c, response.ErrAuthInvalidAuthorizationState, fmt.Errorf("state parameter is required"))
		return
	}

	if queryState != state.Value {
		logger.FromContext(c).Errorf("error matching received state to stored one")
		response.Error(c, response.ErrAuthInvalidAuthorizationState, nil)
		return
	}

	stateData, err := s.parseState(c, state.Value)
	if err != nil {
		return
	}

	s.authLoginCallback(c, stateData, code)
}

// AuthLoginPostCallback is function to catch login callback from OAuth application
// @Summary Login user from external OAuth application
// @Tags Public
// @Accept json
// @Produce json
// @Param json body models.AuthCallback true "Auth form JSON data"
// @Success 303 "redirect to registered return_uri path in the state"
// @Failure 400 {object} response.errorResp "invalid login data"
// @Failure 401 {object} response.errorResp "invalid login or password"
// @Failure 403 {object} response.errorResp "login not permitted"
// @Failure 500 {object} response.errorResp "internal error on login"
// @Router /auth/login-callback [post]
func (s *AuthService) AuthLoginPostCallback(c *gin.Context) {
	var (
		data models.AuthCallback
		err  error
	)

	if err = c.ShouldBind(&data); err != nil || data.Valid() != nil {
		if err == nil {
			err = data.Valid()
		}
		logger.FromContext(c).WithError(err).Errorf("error validating request data")
		response.Error(c, response.ErrAuthInvalidLoginCallbackRequest, err)
		return
	}

	state, err := c.Request.Cookie(authStateCookieName)
	if err != nil {
		logger.FromContext(c).WithError(err).Errorf("error getting state from cookie")
		response.Error(c, response.ErrAuthInvalidAuthorizationState, err)
		return
	}

	if data.State != state.Value {
		logger.FromContext(c).Errorf("error matching received state to stored one")
		response.Error(c, response.ErrAuthInvalidAuthorizationState, nil)
		return
	}

	stateData, err := s.parseState(c, state.Value)
	if err != nil {
		return
	}

	s.authLoginCallback(c, stateData, data.Code)
}

// AuthLogoutCallback is function to catch logout callback from OAuth application
// @Summary Logout current user from external OAuth application
// @Tags Public
// @Accept json
// @Success 303 {object} response.successResp "logout successful"
// @Router /auth/logout-callback [post]
func (s *AuthService) AuthLogoutCallback(c *gin.Context) {
	s.resetSession(c)
	http.Redirect(c.Writer, c.Request, "/", http.StatusSeeOther)
}

// AuthLogout is function to logout current user
// @Summary Logout current user via HTTP redirect
// @Tags Public
// @Produce json
// @Param return_uri query string false "URI to redirect user there after logout" default(/)
// @Success 307 "redirect to input return_uri path"
// @Router /auth/logout [get]
func (s *AuthService) AuthLogout(c *gin.Context) {
	returnURI := "/"
	if returnURL, err := url.Parse(c.Query("return_uri")); err == nil {
		if uri := returnURL.RequestURI(); uri != "" {
			returnURI = path.Clean(path.Join("/", uri))
		}
	}

	session := sessions.Default(c)
	logger.FromContext(c).
		WithFields(logrus.Fields{
			"uid":   session.Get("uid"),
			"uhash": session.Get("uhash"),
			"rid":   session.Get("rid"),
			"tid":   session.Get("tid"),
			"gtm":   session.Get("gtm"),
			"exp":   session.Get("exp"),
			"prm":   session.Get("prm"),
		}).
		Info("user made successful logout")

	s.resetSession(c)
	http.Redirect(c.Writer, c.Request, returnURI, http.StatusTemporaryRedirect)
}

func (s *AuthService) authLoginCallback(c *gin.Context, stateData map[string]string, code string) {
	var (
		privs []string
		role  models.Role
		user  models.User
	)

	provider := stateData["provider"]
	oauthClient, ok := s.oauth[provider]
	if !ok {
		logger.FromContext(c).Errorf("external OAuth2 provider '%s' is not initialized", provider)
		response.Error(c, response.ErrNotPermitted, fmt.Errorf("provider not initialized"))
		return
	}

	ctx := c.Request.Context()

	oauth2Token, err := oauthClient.Exchange(ctx, code)
	if err != nil {
		logger.FromContext(c).WithError(err).Errorf("failed to exchange token")
		response.Error(c, response.ErrAuthExchangeTokenFail, err)
		return
	}

	if !oauth2Token.Valid() {
		logger.FromContext(c).Errorf("failed to validate OAuth2 token")
		response.Error(c, response.ErrAuthVerificationTokenFail, nil)
		return
	}

	nonce, err := c.Request.Cookie(authNonceCookieName)
	if err != nil {
		logger.FromContext(c).WithError(err).Errorf("error getting nonce from cookie")
		response.Error(c, response.ErrAuthInvalidAuthorizationNonce, err)
		return
	}

	email, err := oauthClient.ResolveEmail(ctx, nonce.Value, oauth2Token)
	if err != nil {
		logger.FromContext(c).WithError(err).Errorf("failed to resolve email")
		response.Error(c, response.ErrAuthInvalidUserData, err)
		return
	}

	if !strings.Contains(email, "@") {
		logger.FromContext(c).Errorf("invalid email format '%s'", email)
		response.Error(c, response.ErrAuthInvalidUserData, fmt.Errorf("invalid email format"))
		return
	}

	username := strings.Split(email, "@")[0]
	if username == "" {
		logger.FromContext(c).Errorf("empty username from email '%s'", email)
		response.Error(c, response.ErrAuthInvalidUserData, fmt.Errorf("empty username"))
		return
	}

	err = s.db.Take(&role, "id = ?", models.RoleUser).Error
	if err != nil {
		logger.FromContext(c).WithError(err).Errorf("error getting user role '%d'", models.RoleUser)
		response.Error(c, response.ErrAuthInvalidServiceData, err)
		return
	}

	err = s.db.Table("privileges").
		Where("role_id = ?", models.RoleUser).
		Pluck("name", &privs).Error
	if err != nil {
		logger.FromContext(c).WithError(err).Errorf("error getting user privileges list '%s'", user.Hash)
		response.Error(c, response.ErrAuthInvalidServiceData, err)
		return
	}

	filterQuery := "mail = ? AND type = ?"
	if err = s.db.Take(&user, filterQuery, email, models.UserTypeOAuth).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			user = models.User{
				Hash:   rdb.MakeUserHash(email),
				Mail:   email,
				Name:   username,
				RoleID: models.RoleUser,
				Status: "active",
				Type:   models.UserTypeOAuth,
			}

			tx := s.db.Begin()
			if tx.Error != nil {
				logger.FromContext(c).WithError(tx.Error).Errorf("error starting transaction")
				response.Error(c, response.ErrInternal, tx.Error)
				return
			}

			if err = tx.Create(&user).Error; err != nil {
				tx.Rollback()
				logger.FromContext(c).WithError(err).Errorf("error creating user")
				response.Error(c, response.ErrInternal, err)
				return
			}

			preferences := models.NewUserPreferences(user.ID)
			if err = tx.Create(preferences).Error; err != nil {
				tx.Rollback()
				logger.FromContext(c).WithError(err).Errorf("error creating user preferences")
				response.Error(c, response.ErrInternal, err)
				return
			}

			if err = tx.Commit().Error; err != nil {
				logger.FromContext(c).WithError(err).Errorf("error committing transaction")
				response.Error(c, response.ErrInternal, err)
				return
			}
		} else {
			logger.FromContext(c).WithError(err).Errorf("error searching user by email '%s'", email)
			response.Error(c, response.ErrInternal, err)
			return
		}
	} else if err = user.Valid(); err != nil {
		logger.FromContext(c).WithError(err).Errorf("error validating user data '%s'", user.Hash)
		response.Error(c, response.ErrAuthInvalidUserData, err)
		return
	}

	if user.Status != "active" {
		logger.FromContext(c).Errorf("error checking active state for user '%s'", user.Status)
		response.Error(c, response.ErrAuthInactiveUser, fmt.Errorf("user is inactive"))
		return
	}

	expires := s.cfg.SessionTimeout
	gtm := time.Now().Unix()
	exp := time.Now().Add(time.Duration(expires) * time.Second).Unix()
	session := sessions.Default(c)
	session.Set("uid", user.ID)
	session.Set("uhash", user.Hash)
	session.Set("rid", user.RoleID)
	session.Set("tid", user.Type.String())
	session.Set("prm", privs)
	session.Set("gtm", gtm)
	session.Set("exp", exp)
	session.Set("uuid", user.Mail)
	session.Set("uname", user.Name)
	session.Options(sessions.Options{
		HttpOnly: true,
		Secure:   c.Request.TLS != nil,
		SameSite: http.SameSiteLaxMode,
		Path:     s.cfg.BaseURL,
		MaxAge:   expires,
	})
	if err := session.Save(); err != nil {
		logger.FromContext(c).WithError(err).Errorf("error saving session")
		response.Error(c, response.ErrInternal, err)
		return
	}

	// delete temporary cookies
	// Google OAuth uses POST callback which requires SameSite=None for cross-site requests
	// GitHub and other providers use GET callback which works with SameSite=Lax
	sameSiteMode := http.SameSiteLaxMode
	if stateData["provider"] == "google" {
		sameSiteMode = http.SameSiteNoneMode
	}
	s.setCallbackCookie(c.Writer, c.Request, authStateCookieName, "", 0, sameSiteMode)
	s.setCallbackCookie(c.Writer, c.Request, authNonceCookieName, "", 0, sameSiteMode)

	logger.FromContext(c).
		WithFields(logrus.Fields{
			"age":   expires,
			"uid":   user.ID,
			"uhash": user.Hash,
			"rid":   user.RoleID,
			"tid":   user.Type,
			"gtm":   session.Get("gtm"),
			"exp":   session.Get("exp"),
			"prm":   session.Get("prm"),
		}).
		Infof("user made successful SSO login for '%s' '%s'", user.Mail, user.Name)

	if returnURI := stateData["return_uri"]; returnURI == "" {
		response.Success(c, http.StatusOK, nil)
	} else {
		u, err := url.Parse(returnURI)
		if err != nil {
			response.Success(c, http.StatusOK, nil)
			return
		}
		query := u.Query()
		query.Add("status", "success")
		u.RawQuery = query.Encode()
		http.Redirect(c.Writer, c.Request, u.RequestURI(), http.StatusSeeOther)
	}
}

func (s *AuthService) parseState(c *gin.Context, state string) (map[string]string, error) {
	var stateData map[string]string

	stateJSON, err := base64.RawURLEncoding.DecodeString(state)
	if err != nil {
		logger.FromContext(c).WithError(err).Errorf("error on getting state as a base64")
		response.Error(c, response.ErrAuthInvalidAuthorizationState, err)
		return nil, err
	}

	signatureLen := 32
	if len(stateJSON) <= signatureLen {
		logger.FromContext(c).Errorf("error on parsing state from json data")
		err := fmt.Errorf("unexpected state length")
		response.Error(c, response.ErrAuthInvalidAuthorizationState, err)
		return nil, err
	}
	stateSignature := stateJSON[:signatureLen]
	stateJSON = stateJSON[signatureLen:]

	mac := hmac.New(sha256.New, s.key)
	mac.Write(stateJSON)
	signature := mac.Sum(nil)

	if !hmac.Equal(stateSignature, signature) {
		logger.FromContext(c).Errorf("error on matching signature")
		err := fmt.Errorf("mismatch state signature")
		response.Error(c, response.ErrAuthInvalidAuthorizationState, err)
		return nil, err
	}

	if err := json.Unmarshal(stateJSON, &stateData); err != nil {
		logger.FromContext(c).WithError(err).Errorf("error on parsing state from json data")
		response.Error(c, response.ErrAuthInvalidAuthorizationState, err)
		return nil, err
	}

	expStr, ok := stateData["exp"]
	if !ok || expStr == "" {
		err := fmt.Errorf("missing required field: exp")
		logger.FromContext(c).WithError(err).Errorf("error on validating state data")
		response.Error(c, response.ErrAuthInvalidAuthorizationState, err)
		return nil, err
	}

	if _, ok := stateData["provider"]; !ok {
		err := fmt.Errorf("missing required field: provider")
		logger.FromContext(c).WithError(err).Errorf("error on validating state data")
		response.Error(c, response.ErrAuthInvalidAuthorizationState, err)
		return nil, err
	}

	exp, err := strconv.ParseInt(expStr, 10, 64)
	if err != nil {
		logger.FromContext(c).WithError(err).Errorf("error on parsing expiration time")
		response.Error(c, response.ErrAuthInvalidAuthorizationState, err)
		return nil, err
	}

	if time.Now().Unix() > exp {
		logger.FromContext(c).Errorf("error on checking expiration time")
		err := fmt.Errorf("state signature expired")
		response.Error(c, response.ErrAuthTokenExpired, err)
		return nil, err
	}

	return stateData, nil
}

func (s *AuthService) setCallbackCookie(
	w http.ResponseWriter, r *http.Request,
	name, value string, maxAge int,
	sameSite http.SameSite,
) {
	// Check both direct TLS and X-Forwarded-Proto header (for reverse proxy setups)
	useTLS := r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https"

	c := &http.Cookie{
		Name:     name,
		Value:    value,
		HttpOnly: true,
		Secure:   useTLS,
		SameSite: sameSite,
		Path:     path.Join(s.cfg.BaseURL, s.cfg.LoginCallbackURL),
		MaxAge:   maxAge,
	}
	http.SetCookie(w, c)
}

func (s *AuthService) resetSession(c *gin.Context) {
	now := time.Now().Add(-1 * time.Second)
	session := sessions.Default(c)
	session.Set("gtm", now.Unix())
	session.Set("exp", now.Unix())
	session.Options(sessions.Options{
		HttpOnly: true,
		Secure:   c.Request.TLS != nil,
		SameSite: http.SameSiteLaxMode,
		Path:     s.cfg.BaseURL,
		MaxAge:   -1,
	})
	if err := session.Save(); err != nil {
		logger.FromContext(c).WithError(err).Errorf("error resetting session")
	}
}

// randBase64String is function to generate random base64 with set length (bytes)
func randBase64String(nByte int) (string, error) {
	b := make([]byte, nByte)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// randBytes is function to generate random bytes with set length (bytes)
func randBytes(nByte int) ([]byte, error) {
	b := make([]byte, nByte)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return nil, err
	}
	return b, nil
}

type info struct {
	Type      string      `json:"type"`
	Develop   bool        `json:"develop"`
	User      models.User `json:"user"`
	Role      models.Role `json:"role"`
	Providers []string    `json:"providers"`
	Privs     []string    `json:"privileges"`
	OAuth     bool        `json:"oauth"`
	IssuedAt  time.Time   `json:"issued_at"`
	ExpiresAt time.Time   `json:"expires_at"`
}

// Info is function to return settings and current information about system and config
// @Summary Retrieve current user and system settings
// @Tags Public
// @Produce json
// @Security BearerAuth
// @Param refresh_cookie query boolean false "boolean arg to refresh current cookie, use explicit false"
// @Success 200 {object} response.successResp{data=info} "info received successful"
// @Failure 403 {object} response.errorResp "getting info not permitted"
// @Failure 404 {object} response.errorResp "user not found"
// @Failure 500 {object} response.errorResp "internal error on getting information about system and config"
// @Router /info [get]
func (s *AuthService) Info(c *gin.Context) {
	var resp info

	logger.FromContext(c).WithFields(logrus.Fields(c.Keys)).Trace("AuthService.Info")
	now := time.Now().Unix()
	uhash := c.GetString("uhash")
	uid := c.GetUint64("uid")
	tid := c.GetString("tid")
	exp := c.GetInt64("exp")
	gtm := c.GetInt64("gtm")
	cpt := c.GetString("cpt")
	privs := c.GetStringSlice("prm")

	resp.Privs = privs
	resp.IssuedAt = time.Unix(gtm, 0).UTC()
	resp.ExpiresAt = time.Unix(exp, 0).UTC()
	resp.Develop = version.IsDevelopMode()
	resp.OAuth = tid == models.UserTypeOAuth.String()
	for name := range s.oauth {
		resp.Providers = append(resp.Providers, name)
	}

	logger.FromContext(c).WithFields(logrus.Fields(
		map[string]any{
			"exp":   exp,
			"gtm":   gtm,
			"uhash": uhash,
			"now":   now,
			"cpt":   cpt,
			"uid":   uid,
			"tid":   tid,
		},
	)).Trace("AuthService.Info")

	if uhash == "" || exp == 0 || gtm == 0 || now > exp {
		resp.Type = "guest"
		resp.Privs = []string{}
		response.Success(c, http.StatusOK, resp)
		return
	}

	err := s.db.Take(&resp.User, "id = ?", uid).Related(&resp.Role).Error
	if err != nil {
		response.Error(c, response.ErrInfoUserNotFound, err)
		return
	} else if err = resp.User.Valid(); err != nil {
		logger.FromContext(c).WithError(err).Errorf("error validating user data '%s'", resp.User.Hash)
		response.Error(c, response.ErrInfoInvalidUserData, err)
		return
	}
	if err = s.db.Table("privileges").Where("role_id = ?", resp.User.RoleID).Pluck("name", &privs).Error; err != nil {
		logger.FromContext(c).WithError(err).Errorf("error getting user privileges list '%s'", resp.User.Hash)
		response.Error(c, response.ErrInfoInvalidUserData, err)
		return
	}

	if cpt == "automation" {
		resp.Type = models.UserTypeAPI.String()
		// filter out privileges that are not supported for API tokens
		privs = slices.DeleteFunc(privs, func(priv string) bool {
			return strings.HasPrefix(priv, "users.") ||
				strings.HasPrefix(priv, "roles.") ||
				strings.HasPrefix(priv, "settings.user.") ||
				strings.HasPrefix(priv, "settings.tokens.")
		})
		resp.Privs = privs
		response.Success(c, http.StatusOK, resp)
		return
	}

	resp.Type = "user"

	// check 5 minutes timeout to refresh current token
	var fiveMins int64 = 5 * 60
	if now >= gtm+fiveMins && c.Query("refresh_cookie") != "false" {
		if err = s.refreshCookie(c, &resp, privs); err != nil {
			logger.FromContext(c).WithError(err).Errorf("failed to refresh token")
			// raise error when there is elapsing last five minutes
			if now >= gtm+int64(s.cfg.SessionTimeout)-fiveMins {
				response.Error(c, response.ErrAuthRequired, err)
				return
			}
		}
	}

	// raise error when user has no permissions in the session auth cookie
	if resp.Type != "guest" && resp.Privs == nil {
		logger.FromContext(c).
			WithFields(logrus.Fields{
				"uid": resp.User.ID,
				"rid": resp.User.RoleID,
				"tid": resp.User.Type,
			}).
			Errorf("failed to get user privileges for '%s' '%s'", resp.User.Mail, resp.User.Name)
		response.Error(c, response.ErrInternal, err)
		return
	}

	response.Success(c, http.StatusOK, resp)
}
