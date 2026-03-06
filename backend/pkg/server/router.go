package router

import (
	"encoding/gob"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"pentagi/pkg/config"
	"pentagi/pkg/controller"
	"pentagi/pkg/database"
	"pentagi/pkg/graph/subscriptions"
	"pentagi/pkg/providers"
	"pentagi/pkg/server/auth"
	"pentagi/pkg/server/logger"
	"pentagi/pkg/server/oauth"
	"pentagi/pkg/server/services"

	_ "pentagi/pkg/server/docs" // swagger docs

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"github.com/sirupsen/logrus"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/swaggo/gin-swagger/swaggerFiles"
)

const baseURL = "/api/v1"

const corsAllowGoogleOAuth = "https://accounts.google.com"

// frontendRoutes defines the list of URI prefixes that should be handled by the frontend SPA.
// Add new frontend base routes here if they are added in the frontend router (e.g., in App.tsx).
var frontendRoutes = []string{
	"/chat",
	"/oauth",
	"/login",
	"/flows",
	"/settings",
}

// @title PentAGI Swagger API
// @version 1.0
// @description Swagger API for Penetration Testing Advanced General Intelligence PentAGI.
// @termsOfService http://swagger.io/terms/

// @contact.url https://pentagi.com
// @contact.name PentAGI Development Team
// @contact.email team@pentagi.com

// @license.name MIT
// @license.url https://opensource.org/license/mit

// @query.collection.format multi

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

// @BasePath /api/v1
func NewRouter(
	db *database.Queries,
	orm *gorm.DB,
	cfg *config.Config,
	providers providers.ProviderController,
	controller controller.FlowController,
	subscriptions subscriptions.SubscriptionsController,
) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	if cfg.Debug {
		gin.SetMode(gin.DebugMode)
	}

	gob.Register([]string{})

	tokenCache := auth.NewTokenCache(orm)
	userCache := auth.NewUserCache(orm)
	authMiddleware := auth.NewAuthMiddleware(baseURL, cfg.CookieSigningSalt, tokenCache, userCache)
	oauthClients := make(map[string]oauth.OAuthClient)
	oauthLoginCallbackURL := "/auth/login-callback"

	publicURL, err := url.Parse(cfg.PublicURL)
	if err == nil {
		publicURL.Path = path.Join(baseURL, oauthLoginCallbackURL)
	}

	if publicURL != nil && cfg.OAuthGoogleClientID != "" && cfg.OAuthGoogleClientSecret != "" {
		googleClient := oauth.NewGoogleOAuthClient(
			cfg.OAuthGoogleClientID,
			cfg.OAuthGoogleClientSecret,
			publicURL.String(),
		)
		oauthClients[googleClient.ProviderName()] = googleClient
	}

	if publicURL != nil && cfg.OAuthGithubClientID != "" && cfg.OAuthGithubClientSecret != "" {
		githubClient := oauth.NewGithubOAuthClient(
			cfg.OAuthGithubClientID,
			cfg.OAuthGithubClientSecret,
			publicURL.String(),
		)
		oauthClients[githubClient.ProviderName()] = githubClient
	}

	// services
	authService := services.NewAuthService(
		services.AuthServiceConfig{
			BaseURL:          baseURL,
			LoginCallbackURL: oauthLoginCallbackURL,
			SessionTimeout:   4 * 60 * 60, // 4 hours
		},
		orm,
		oauthClients,
	)
	userService := services.NewUserService(orm, userCache)
	roleService := services.NewRoleService(orm)
	providerService := services.NewProviderService(providers)
	flowService := services.NewFlowService(orm, providers, controller, subscriptions)
	taskService := services.NewTaskService(orm)
	subtaskService := services.NewSubtaskService(orm)
	containerService := services.NewContainerService(orm)
	assistantService := services.NewAssistantService(orm, providers, controller, subscriptions)
	agentlogService := services.NewAgentlogService(orm)
	assistantlogService := services.NewAssistantlogService(orm)
	msglogService := services.NewMsglogService(orm)
	searchlogService := services.NewSearchlogService(orm)
	vecstorelogService := services.NewVecstorelogService(orm)
	termlogService := services.NewTermlogService(orm)
	screenshotService := services.NewScreenshotService(orm, cfg.DataDir)
	promptService := services.NewPromptService(orm)
	analyticsService := services.NewAnalyticsService(orm)
	tokenService := services.NewTokenService(orm, cfg.CookieSigningSalt, tokenCache, subscriptions)
	graphqlService := services.NewGraphqlService(
		db, cfg, baseURL, cfg.CorsOrigins, tokenCache, providers, controller, subscriptions,
	)

	router := gin.Default()

	// Configure CORS middleware
	config := cors.DefaultConfig()
	if !slices.Contains(cfg.CorsOrigins, "*") {
		config.AllowCredentials = true
	}
	config.AllowWildcard = true
	config.AllowWebSockets = true
	config.AllowPrivateNetwork = true

	// Add OAuth provider origins to CORS allowed origins
	allowedOrigins := make([]string, len(cfg.CorsOrigins))
	copy(allowedOrigins, cfg.CorsOrigins)

	// Google OAuth uses POST callback from accounts.google.com
	if cfg.OAuthGoogleClientID != "" && cfg.OAuthGoogleClientSecret != "" {
		if !slices.Contains(allowedOrigins, corsAllowGoogleOAuth) && !slices.Contains(cfg.CorsOrigins, "*") {
			allowedOrigins = append(allowedOrigins, corsAllowGoogleOAuth)
			logrus.Infof("Added %s to CORS allowed origins for Google OAuth", corsAllowGoogleOAuth)
		}
	}

	config.AllowOrigins = allowedOrigins
	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	if err := config.Validate(); err != nil {
		logrus.WithError(err).Error("failed to validate cors config")
	} else {
		router.Use(cors.New(config))
	}

	router.Use(gin.Recovery())
	router.Use(logger.WithGinLogger("pentagi-api"))

	cookieStore := cookie.NewStore(auth.MakeCookieStoreKey(cfg.CookieSigningSalt)...)
	router.Use(sessions.Sessions("auth", cookieStore))

	api := router.Group(baseURL)
	api.Use(noCacheMiddleware())

	// Special case for local user own password change
	changePasswordGroup := api.Group("/user")
	changePasswordGroup.Use(authMiddleware.AuthUserRequired)
	changePasswordGroup.Use(localUserRequired())
	changePasswordGroup.PUT("/password", userService.ChangePasswordCurrentUser)

	publicGroup := api.Group("/")
	publicGroup.Use(authMiddleware.TryAuth)
	{
		publicGroup.GET("/info", authService.Info)

		developerGroup := publicGroup.Group("/")
		{
			developerGroup.GET("/graphql/playground", graphqlService.ServeGraphqlPlayground)
			developerGroup.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
		}

		authGroup := publicGroup.Group("/auth")
		{
			authGroup.POST("/login", authService.AuthLogin)
			authGroup.GET("/logout", authService.AuthLogout)
			authGroup.GET("/authorize", authService.AuthAuthorize)
			authGroup.GET("/login-callback", authService.AuthLoginGetCallback)
			authGroup.POST("/login-callback", authService.AuthLoginPostCallback)
			authGroup.POST("/logout-callback", authService.AuthLogoutCallback)
		}
	}

	privateGroup := api.Group("/")
	privateGroup.Use(authMiddleware.AuthTokenRequired)
	{
		setGraphqlGroup(privateGroup, graphqlService)

		setProvidersGroup(privateGroup, providerService)
		setFlowsGroup(privateGroup, flowService)
		setTasksGroup(privateGroup, taskService)
		setSubtasksGroup(privateGroup, subtaskService)
		setContainersGroup(privateGroup, containerService)
		setAssistantsGroup(privateGroup, assistantService)
		setAgentlogsGroup(privateGroup, agentlogService)
		setAssistantlogsGroup(privateGroup, assistantlogService)
		setMsglogsGroup(privateGroup, msglogService)
		setTermlogsGroup(privateGroup, termlogService)
		setSearchlogsGroup(privateGroup, searchlogService)
		setVecstorelogsGroup(privateGroup, vecstorelogService)
		setScreenshotsGroup(privateGroup, screenshotService)
		setPromptsGroup(privateGroup, promptService)
		setAnalyticsGroup(privateGroup, analyticsService)
	}

	privateUserGroup := api.Group("/")
	privateUserGroup.Use(authMiddleware.AuthUserRequired)
	{
		setRolesGroup(privateGroup, roleService)
		setUsersGroup(privateGroup, userService)
		setTokensGroup(privateGroup, tokenService)
	}

	if cfg.StaticURL != nil && cfg.StaticURL.Scheme != "" && cfg.StaticURL.Host != "" {
		router.NoRoute(func() gin.HandlerFunc {
			return func(c *gin.Context) {
				director := func(req *http.Request) {
					*req = *c.Request
					req.URL.Scheme = cfg.StaticURL.Scheme
					req.URL.Host = cfg.StaticURL.Host
				}
				dialer := &net.Dialer{
					Timeout:   30 * time.Second,
					KeepAlive: 30 * time.Second,
				}
				httpTransport := &http.Transport{
					DialContext:           dialer.DialContext,
					ForceAttemptHTTP2:     true,
					MaxIdleConns:          20,
					IdleConnTimeout:       60 * time.Second,
					TLSHandshakeTimeout:   10 * time.Second,
					ExpectContinueTimeout: 1 * time.Second,
				}

				proxy := &httputil.ReverseProxy{
					Director:  director,
					Transport: httpTransport,
				}
				proxy.ServeHTTP(c.Writer, c.Request)
			}
		}())
	} else {
		router.Use(static.Serve("/", static.LocalFile(cfg.StaticDir, true)))

		indexExists := true
		indexPath := filepath.Join(cfg.StaticDir, "index.html")
		if _, err := os.Stat(indexPath); err != nil {
			indexExists = false
		}

		router.NoRoute(func(c *gin.Context) {
			if c.Request.Method == "GET" && !strings.HasPrefix(c.Request.URL.Path, baseURL) {
				isFrontendRoute := false
				path := c.Request.URL.Path
				for _, prefix := range frontendRoutes {
					if path == prefix || strings.HasPrefix(path, prefix+"/") {
						isFrontendRoute = true
						break
					}
				}

				if isFrontendRoute && indexExists {
					c.File(indexPath)
					return
				}
			}

			c.Redirect(http.StatusMovedPermanently, "/")
		})
	}

	return router
}

func setProvidersGroup(parent *gin.RouterGroup, svc *services.ProviderService) {
	providersGroup := parent.Group("/providers")
	{
		providersGroup.GET("/", svc.GetProviders)
	}
}

func setGraphqlGroup(parent *gin.RouterGroup, svc *services.GraphqlService) {
	graphqlGroup := parent.Group("/")
	{
		graphqlGroup.Any("/graphql", svc.ServeGraphql)
	}
}

func setSubtasksGroup(parent *gin.RouterGroup, svc *services.SubtaskService) {
	flowSubtasksViewGroup := parent.Group("/flows/:flowID/subtasks")
	{
		flowSubtasksViewGroup.GET("/", svc.GetFlowSubtasks)
	}

	flowTaskSubtasksViewGroup := parent.Group("/flows/:flowID/tasks/:taskID/subtasks")
	{
		flowTaskSubtasksViewGroup.GET("/", svc.GetFlowTaskSubtasks)
		flowTaskSubtasksViewGroup.GET("/:subtaskID", svc.GetFlowTaskSubtask)
	}
}

func setTasksGroup(parent *gin.RouterGroup, svc *services.TaskService) {
	flowTaskViewGroup := parent.Group("/flows/:flowID/tasks")
	{
		flowTaskViewGroup.GET("/", svc.GetFlowTasks)
		flowTaskViewGroup.GET("/:taskID", svc.GetFlowTask)
		flowTaskViewGroup.GET("/:taskID/graph", svc.GetFlowTaskGraph)
	}
}

func setFlowsGroup(parent *gin.RouterGroup, svc *services.FlowService) {
	flowCreateGroup := parent.Group("/flows")
	{
		flowCreateGroup.POST("/", svc.CreateFlow)
	}

	flowDeleteGroup := parent.Group("/flows")
	{
		flowDeleteGroup.DELETE("/:flowID", svc.DeleteFlow)
	}

	flowEditGroup := parent.Group("/flows")
	{
		flowEditGroup.PUT("/:flowID", svc.PatchFlow)
	}

	flowsViewGroup := parent.Group("/flows")
	{
		flowsViewGroup.GET("/", svc.GetFlows)
		flowsViewGroup.GET("/:flowID", svc.GetFlow)
		flowsViewGroup.GET("/:flowID/graph", svc.GetFlowGraph)
	}
}

func setContainersGroup(parent *gin.RouterGroup, svc *services.ContainerService) {
	containersViewGroup := parent.Group("/containers")
	{
		containersViewGroup.GET("/", svc.GetContainers)
	}

	flowContainersViewGroup := parent.Group("/flows/:flowID/containers")
	{
		flowContainersViewGroup.GET("/", svc.GetFlowContainers)
		flowContainersViewGroup.GET("/:containerID", svc.GetFlowContainer)
	}
}

func setAssistantsGroup(parent *gin.RouterGroup, svc *services.AssistantService) {
	flowCreateGroup := parent.Group("/flows/:flowID/assistants")
	{
		flowCreateGroup.POST("/", svc.CreateFlowAssistant)
	}

	flowDeleteGroup := parent.Group("/flows/:flowID/assistants")
	{
		flowDeleteGroup.DELETE("/:assistantID", svc.DeleteAssistant)
	}

	flowEditGroup := parent.Group("/flows/:flowID/assistants")
	{
		flowEditGroup.PUT("/:assistantID", svc.PatchAssistant)
	}

	flowsViewGroup := parent.Group("/flows/:flowID/assistants")
	{
		flowsViewGroup.GET("/", svc.GetFlowAssistants)
		flowsViewGroup.GET("/:assistantID", svc.GetFlowAssistant)
	}
}

func setAgentlogsGroup(parent *gin.RouterGroup, svc *services.AgentlogService) {
	agentlogsViewGroup := parent.Group("/agentlogs")
	{
		agentlogsViewGroup.GET("/", svc.GetAgentlogs)
	}

	flowAgentlogsViewGroup := parent.Group("/flows/:flowID/agentlogs")
	{
		flowAgentlogsViewGroup.GET("/", svc.GetFlowAgentlogs)
	}
}

func setAssistantlogsGroup(parent *gin.RouterGroup, svc *services.AssistantlogService) {
	assistantlogsViewGroup := parent.Group("/assistantlogs")
	{
		assistantlogsViewGroup.GET("/", svc.GetAssistantlogs)
	}

	flowAssistantlogsViewGroup := parent.Group("/flows/:flowID/assistantlogs")
	{
		flowAssistantlogsViewGroup.GET("/", svc.GetFlowAssistantlogs)
	}
}

func setMsglogsGroup(parent *gin.RouterGroup, svc *services.MsglogService) {
	msglogsViewGroup := parent.Group("/msglogs")
	{
		msglogsViewGroup.GET("/", svc.GetMsglogs)
	}

	flowMsglogsViewGroup := parent.Group("/flows/:flowID/msglogs")
	{
		flowMsglogsViewGroup.GET("/", svc.GetFlowMsglogs)
	}
}

func setSearchlogsGroup(parent *gin.RouterGroup, svc *services.SearchlogService) {
	searchlogsViewGroup := parent.Group("/searchlogs")
	{
		searchlogsViewGroup.GET("/", svc.GetSearchlogs)
	}

	flowSearchlogsViewGroup := parent.Group("/flows/:flowID/searchlogs")
	{
		flowSearchlogsViewGroup.GET("/", svc.GetFlowSearchlogs)
	}
}

func setTermlogsGroup(parent *gin.RouterGroup, svc *services.TermlogService) {
	termlogsViewGroup := parent.Group("/termlogs")
	{
		termlogsViewGroup.GET("/", svc.GetTermlogs)
	}

	flowTermlogsViewGroup := parent.Group("/flows/:flowID/termlogs")
	{
		flowTermlogsViewGroup.GET("/", svc.GetFlowTermlogs)
	}
}

func setVecstorelogsGroup(parent *gin.RouterGroup, svc *services.VecstorelogService) {
	vecstorelogsViewGroup := parent.Group("/vecstorelogs")
	{
		vecstorelogsViewGroup.GET("/", svc.GetVecstorelogs)
	}

	flowVecstorelogsViewGroup := parent.Group("/flows/:flowID/vecstorelogs")
	{
		flowVecstorelogsViewGroup.GET("/", svc.GetFlowVecstorelogs)
	}
}

func setScreenshotsGroup(parent *gin.RouterGroup, svc *services.ScreenshotService) {
	screenshotsViewGroup := parent.Group("/screenshots")
	{
		screenshotsViewGroup.GET("/", svc.GetScreenshots)
	}

	flowScreenshotsViewGroup := parent.Group("/flows/:flowID/screenshots")
	{
		flowScreenshotsViewGroup.GET("/", svc.GetFlowScreenshots)
		flowScreenshotsViewGroup.GET("/:screenshotID", svc.GetFlowScreenshot)
		flowScreenshotsViewGroup.GET("/:screenshotID/file", svc.GetFlowScreenshotFile)
	}
}

func setPromptsGroup(parent *gin.RouterGroup, svc *services.PromptService) {
	promptsViewGroup := parent.Group("/prompts")
	{
		promptsViewGroup.GET("/", svc.GetPrompts)
		promptsViewGroup.GET("/:promptType", svc.GetPrompt)
	}

	promptsEditGroup := parent.Group("/prompts")
	{
		promptsEditGroup.PUT("/:promptType", svc.PatchPrompt)
		promptsEditGroup.POST("/:promptType/default", svc.ResetPrompt)
		promptsEditGroup.DELETE("/:promptType", svc.DeletePrompt)
	}
}

func setRolesGroup(parent *gin.RouterGroup, svc *services.RoleService) {
	rolesViewGroup := parent.Group("/roles")
	{
		rolesViewGroup.GET("/", svc.GetRoles)
		rolesViewGroup.GET("/:roleID", svc.GetRole)
	}
}

func setUsersGroup(parent *gin.RouterGroup, svc *services.UserService) {
	usersCreateGroup := parent.Group("/users")
	{
		usersCreateGroup.POST("/", svc.CreateUser)
	}

	usersDeleteGroup := parent.Group("/users")
	{
		usersDeleteGroup.DELETE("/:hash", svc.DeleteUser)
	}

	usersEditGroup := parent.Group("/users")
	{
		usersEditGroup.PUT("/:hash", svc.PatchUser)
	}

	usersViewGroup := parent.Group("/users")
	{
		usersViewGroup.GET("/", svc.GetUsers)
		usersViewGroup.GET("/:hash", svc.GetUser)
	}

	userViewGroup := parent.Group("/user")
	{
		userViewGroup.GET("/", svc.GetCurrentUser)
	}
}

func setAnalyticsGroup(parent *gin.RouterGroup, svc *services.AnalyticsService) {
	// System-wide analytics
	usageViewGroup := parent.Group("/usage")
	{
		usageViewGroup.GET("/", svc.GetSystemUsage)
		usageViewGroup.GET("/:period", svc.GetPeriodUsage)
	}

	// Flow-specific analytics
	flowUsageViewGroup := parent.Group("/flows/:flowID/usage")
	{
		flowUsageViewGroup.GET("/", svc.GetFlowUsage)
	}
}

func setTokensGroup(parent *gin.RouterGroup, svc *services.TokenService) {
	tokensGroup := parent.Group("/tokens")
	{
		tokensGroup.POST("/", svc.CreateToken)
		tokensGroup.GET("/", svc.ListTokens)
		tokensGroup.GET("/:tokenID", svc.GetToken)
		tokensGroup.PUT("/:tokenID", svc.UpdateToken)
		tokensGroup.DELETE("/:tokenID", svc.DeleteToken)
	}
}
