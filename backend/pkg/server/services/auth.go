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

	"pentagi/pkg/server/auth"
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

	// CookiePrefix namespaces the OAuth CSRF cookies for this instance. Cookies
	// are scoped by host and not by port, so two instances on one host would
	// otherwise clobber each other's in-flight OAuth handshakes. Empty in
	// single-instance mode, which keeps the cookie names exactly "state"/"nonce".
	CookiePrefix string

	LoginPairAttemptLimit int
	LoginAddrAttemptLimit int
	LoginFailureWindow    time.Duration
	LoginLockout          time.Duration
}

// stateCookieName returns the tenant-scoped name of the OAuth CSRF state cookie.
func (s *AuthService) stateCookieName() string {
	return s.cfg.CookiePrefix + authStateCookieName
}

// nonceCookieName returns the tenant-scoped name of the OIDC nonce cookie.
func (s *AuthService) nonceCookieName() string {
	return s.cfg.CookiePrefix + authNonceCookieName
}

type AuthService struct {
	cfg       AuthServiceConfig
	db        *gorm.DB
	key       []byte
	oauth     map[string]oauth.OAuthClient
	userCache *auth.UserCache
	pairGuard *auth.LoginGuard
	addrGuard *auth.LoginGuard
	dummyHash []byte
}

func NewAuthService(
	cfg AuthServiceConfig,
	db *gorm.DB,
	oauth map[string]oauth.OAuthClient,
	userCache *auth.UserCache,
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

	if cfg.LoginPairAttemptLimit == 0 {
		cfg.LoginPairAttemptLimit = defaultLoginPairAttemptLimit
	}
	if cfg.LoginAddrAttemptLimit == 0 {
		cfg.LoginAddrAttemptLimit = defaultLoginAddrAttemptLimit
	}
	if cfg.LoginFailureWindow == 0 {
		cfg.LoginFailureWindow = defaultLoginFailureWindow
	}
	if cfg.LoginLockout == 0 {
		cfg.LoginLockout = defaultLoginLockout
	}

	return &AuthService{
		cfg:       cfg,
		db:        db,
		key:       key,
		oauth:     oauth,
		userCache: userCache,
		pairGuard: auth.NewLoginGuard(cfg.LoginPairAttemptLimit, cfg.LoginFailureWindow, cfg.LoginLockout),
		addrGuard: auth.NewLoginGuard(cfg.LoginAddrAttemptLimit, cfg.LoginFailureWindow, cfg.LoginLockout),
		dummyHash: newDummyPasswordHash(),
	}
}

const (
	defaultLoginPairAttemptLimit = 10
	defaultLoginAddrAttemptLimit = 30
	defaultLoginFailureWindow    = 5 * time.Minute
	defaultLoginLockout          = 15 * time.Minute
)

func newDummyPasswordHash() []byte {
	secret, err := randBytes(32)
	if err != nil {
		logrus.WithError(err).Errorf("error generating dummy password source")
		secret = make([]byte, 32)
	}

	hash, err := bcrypt.GenerateFromPassword(secret, bcrypt.DefaultCost)
	if err != nil {
		logrus.WithError(err).Errorf("error generating dummy password hash")
		return nil
	}

	return hash
}

var errInvalidCredentials = errors.New("invalid login or password")

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
// @Failure 429 {object} response.errorResp "too many login attempts"
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

	addrKey := c.ClientIP()
	pairKey := strings.ToLower(data.Mail) + "|" + addrKey

	retryAfter, allowed := s.pairGuard.Take(pairKey)
	if allowed {
		if retryAfter, allowed = s.addrGuard.Take(addrKey); !allowed {
			s.pairGuard.Refund(pairKey)
		}
	}
	if !allowed {
		c.Header("Retry-After", strconv.Itoa(int(retryAfter.Round(time.Second).Seconds())))
		logger.FromContext(c).Errorf("too many login attempts")
		response.Error(c, response.ErrAuthTooManyAttempts, fmt.Errorf("too many login attempts"))
		return
	}

	var user models.UserPassword
	if err := s.db.Take(&user, "mail = ? AND password IS NOT NULL", data.Mail).Error; err != nil {
		_ = bcrypt.CompareHashAndPassword(s.dummyHash, []byte(data.Password))
		logrus.WithError(err).Errorf("error getting user by mail '%s'", data.Mail)
		response.Error(c, response.ErrAuthInvalidCredentials, errInvalidCredentials)
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
		logger.FromContext(c).WithError(err).Errorf("error matching user input password")
		response.Error(c, response.ErrAuthInvalidCredentials, errInvalidCredentials)
		return
	}

	if user.Status != "active" {
		logger.FromContext(c).Errorf("error checking active state for user '%s'", user.Status)
		response.Error(c, response.ErrAuthInactiveUser, fmt.Errorf("user is inactive"))
		return
	}

	s.pairGuard.Reset(pairKey)
	s.addrGuard.Relax(addrKey)

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
	session.Set("sgn", user.SessionGeneration)
	session.Options(sessionOptions(c.Request, s.cfg.BaseURL, int(expires)))
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
	session.Set("sgn", resp.User.SessionGeneration)
	session.Options(sessionOptions(c.Request, s.cfg.BaseURL, expires))
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
	s.setCallbackCookie(c.Writer, c.Request, s.stateCookieName(), state, maxAge, sameSiteMode)
	s.setCallbackCookie(c.Writer, c.Request, s.nonceCookieName(), nonce, maxAge, sameSiteMode)

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

	state, err := c.Request.Cookie(s.stateCookieName())
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

	state, err := c.Request.Cookie(s.stateCookieName())
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
// @Failure 500 {object} response.errorResp "internal error on ending the user's sessions"
// @Router /auth/logout-callback [post]
func (s *AuthService) AuthLogoutCallback(c *gin.Context) {
	err := s.revokeCallerSessions(c)
	s.resetSession(c)
	if err != nil {
		logger.FromContext(c).WithError(err).Errorf("error revoking sessions on logout callback")
		response.Error(c, response.ErrInternal, err)
		return
	}

	http.Redirect(c.Writer, c.Request, "/", http.StatusSeeOther)
}

// AuthLogout is function to logout current user
// @Summary Logout current user and end all of their sessions
// @Tags Public
// @Produce json
// @Success 200 {object} response.successResp "logout successful"
// @Failure 500 {object} response.errorResp "internal error on ending the user's sessions"
// @Router /auth/logout [post]
func (s *AuthService) AuthLogout(c *gin.Context) {
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

	err := s.revokeCallerSessions(c)
	s.resetSession(c)
	if err != nil {
		logger.FromContext(c).WithError(err).Errorf("error revoking sessions on logout")
		response.Error(c, response.ErrInternal, err)
		return
	}

	response.Success(c, http.StatusOK, struct{}{})
}

func (s *AuthService) revokeCallerSessions(c *gin.Context) error {
	if _, isSession := c.Get("sgn"); !isSession {
		return auth.BackendFailure(c)
	}

	_, err := auth.RevokeSessions(s.db, s.userCache, c.GetUint64("uid"))
	if gorm.IsRecordNotFoundError(err) {
		return nil
	}

	return err
}

func (s *AuthService) authLoginCallback(c *gin.Context, stateData map[string]string, code string) {
	var (
		privs []string
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

	nonce, err := c.Request.Cookie(s.nonceCookieName())
	if err != nil {
		logger.FromContext(c).WithError(err).Errorf("error getting nonce from cookie")
		response.Error(c, response.ErrAuthInvalidAuthorizationNonce, err)
		return
	}

	email, verified, err := oauthClient.ResolveEmail(ctx, nonce.Value, oauth2Token)
	if err != nil {
		logger.FromContext(c).WithError(err).Errorf("failed to resolve email")
		response.Error(c, response.ErrAuthInvalidUserData, err)
		return
	}

	if !verified {
		logger.FromContext(c).Errorf("provider returned an unverified email '%s'", email)
		response.Error(c, response.ErrAuthInvalidUserData, fmt.Errorf("email not verified by provider"))
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

	if err = s.db.Take(&user, "mail = ?", email).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			user = models.User{
				Hash:     rdb.MakeUserHash(email),
				Mail:     email,
				Name:     username,
				RoleID:   models.RoleUser,
				Status:   "active",
				Type:     models.UserTypeOAuth,
				Provider: &provider,
			}

			tx := s.db.Begin()
			if tx.Error != nil {
				logger.FromContext(c).WithError(tx.Error).Errorf("error starting transaction")
				response.Error(c, response.ErrInternal, tx.Error)
				return
			}

			if err = tx.Create(&user).Error; err != nil {
				tx.Rollback()
				if !isUniqueViolation(err) {
					logger.FromContext(c).WithError(err).Errorf("error creating user")
					response.Error(c, response.ErrInternal, err)
					return
				}
				if err = s.db.Take(&user, "mail = ?", email).Error; err != nil {
					logger.FromContext(c).WithError(err).Errorf("error loading concurrently created user '%s'", email)
					response.Error(c, response.ErrInternal, err)
					return
				}
			} else {
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
	} else if user.Provider == nil || *user.Provider != provider {
		if err = s.db.Model(&user).Update("provider", provider).Error; err != nil {
			logger.FromContext(c).WithError(err).Errorf("error updating user provider '%s'", user.Hash)
		}
	}

	if user.Status != "active" {
		logger.FromContext(c).Errorf("error checking active state for user '%s'", user.Status)
		response.Error(c, response.ErrAuthInactiveUser, fmt.Errorf("user is inactive"))
		return
	}

	err = s.db.Table("privileges").
		Where("role_id = ?", user.RoleID).
		Pluck("name", &privs).Error
	if err != nil {
		logger.FromContext(c).WithError(err).Errorf("error getting user privileges list '%s'", user.Hash)
		response.Error(c, response.ErrAuthInvalidServiceData, err)
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
	session.Set("sgn", user.SessionGeneration)
	session.Options(sessionOptions(c.Request, s.cfg.BaseURL, expires))
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
	s.setCallbackCookie(c.Writer, c.Request, s.stateCookieName(), "", 0, sameSiteMode)
	s.setCallbackCookie(c.Writer, c.Request, s.nonceCookieName(), "", 0, sameSiteMode)

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
	c := &http.Cookie{
		Name:     name,
		Value:    value,
		HttpOnly: true,
		Secure:   isHTTPS(r),
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
	session.Options(sessionOptions(c.Request, s.cfg.BaseURL, -1))
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
// @Failure 503 {object} response.errorResp "the session could not be checked right now"
// @Router /info [get]
func (s *AuthService) Info(c *gin.Context) {
	var resp info

	if err := auth.BackendFailure(c); err != nil {
		response.Error(c, response.ErrAuthUnavailable, err)
		return
	}

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

	answerGuest := func() {
		resp.Type = "guest"
		resp.Privs = []string{}
		resp.User, resp.Role = models.User{}, models.Role{}
		resp.IssuedAt, resp.ExpiresAt = time.Unix(0, 0).UTC(), time.Unix(0, 0).UTC()
		response.Success(c, http.StatusOK, resp)
	}

	if uhash == "" || exp == 0 || gtm == 0 || now > exp {
		answerGuest()
		return
	}

	err := s.db.Take(&resp.User, "id = ?", uid).Related(&resp.Role).Error
	if gorm.IsRecordNotFoundError(err) {
		response.Error(c, response.ErrInfoUserNotFound, err)
		return
	} else if err != nil {
		response.Error(c, response.ErrAuthUnavailable, err)
		return
	} else if err = resp.User.Valid(); err != nil {
		logger.FromContext(c).WithError(err).Errorf("error validating user data '%s'", resp.User.Hash)
		response.Error(c, response.ErrInfoInvalidUserData, err)
		return
	}

	if _, isSession := c.Get("sgn"); isSession && c.GetUint64("sgn") != resp.User.SessionGeneration {
		s.userCache.Invalidate(uid)
		answerGuest()
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
