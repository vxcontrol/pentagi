package router

import (
	"context"
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
	"pentagi/pkg/database/knowledge"
	"pentagi/pkg/database/knowledge/vectorstore"
	"pentagi/pkg/docker"
	"pentagi/pkg/graph/subscriptions"
	"pentagi/pkg/providers"
	"pentagi/pkg/server/auth"
	"pentagi/pkg/server/logger"
	"pentagi/pkg/server/oauth"
	"pentagi/pkg/server/services"
	"pentagi/pkg/server/update"
	"pentagi/pkg/timezone"

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
	"github.com/vxcontrol/cloud/anonymizer"
	"github.com/vxcontrol/cloud/anonymizer/patterns"
	"github.com/vxcontrol/langchaingo/vectorstores/pgvector"
)

const (
	baseURL = "/api/v1"

	// sessionTimeout is the life of an auth cookie, shared by the handler that
	// mints it and the one that re-stamps it after a password change.
	sessionTimeout = 4 * 60 * 60
)

const corsAllowGoogleOAuth = "https://accounts.google.com"

// frontendRoutes defines the list of URI prefixes that should be handled by the frontend SPA.
// Add new frontend base routes here if they are added in the frontend router (e.g., in App.tsx).
var frontendRoutes = []string{
	"/chat",
	"/oauth",
	"/login",
	"/flows",
	"/settings",
	"/templates",
	"/resources",
	"/knowledges",
	"/dashboard",
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
// Deadlines around a connection, not inside a request — RequestTimeout bounds
// the request itself. Read and write deadlines are deliberately absent: a body
// or a response may legitimately run for as long as an upload, a download or a
// subscription socket does.
const (
	ReadHeaderTimeout = 15 * time.Second
	IdleTimeout       = 2 * time.Minute
)

func newEngine(trustedProxies []string) *gin.Engine {
	engine := gin.Default()
	engine.ContextWithFallback = true

	// gin trusts every peer out of the box, so without this ClientIP returns
	// whatever X-Forwarded-For the caller wrote.
	if err := engine.SetTrustedProxies(trustedProxies); err != nil {
		logrus.WithError(err).Error("error setting trusted proxies, falling back to the peer address")
		_ = engine.SetTrustedProxies(nil)
	}

	return engine
}

func NewRouter(
	db *database.Queries,
	orm *gorm.DB,
	cfg *config.Config,
	providers providers.ProviderController,
	controller controller.FlowController,
	subscriptions subscriptions.SubscriptionsController,
	dockerClient docker.DockerClient,
	updates *update.Service,
) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	if cfg.Debug {
		gin.SetMode(gin.DebugMode)
	}

	gob.Register([]string{})

	tokenCache := auth.NewTokenCache(orm)
	userCache := auth.NewUserCache(orm)
	authMiddleware := auth.NewAuthMiddleware(baseURL, cfg.AuthSalt(), tokenCache, userCache)
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

	// ---- Knowledge (pgvector) store -----------------------------------------
	// Shared by both the GraphQL and REST layers.
	// Store and embedder are nil when no embedding provider is configured;
	// the knowledge store handles that gracefully (embedding-dependent ops error).
	embedder := providers.Embedder()
	var pgStore *pgvector.Store
	if embedder.IsAvailable() {
		if s, err := pgvector.New(context.Background(),
			vectorstore.Options(embedder, cfg.PgxPool, cfg.DatabaseURL)...); err == nil {
			pgStore = &s
		} else {
			logrus.WithError(err).Warn("failed to initialise pgvector store for knowledge API; embedding operations will be unavailable")
		}
	}
	var knowledgeStore = knowledge.NewKnowledgeStore(db, pgStore, embedder, subscriptions.NewKnowledgePublisher, cfg.EmbeddingMaxTextBytes)

	// ---- Anonymizer replacer ------------------------------------------------
	// Shared singleton used by the GraphQL anonymizeText mutation.
	// Falls back to a no-op nil replacer on failure so the rest of the server still starts correctly.
	var textReplacer anonymizer.Replacer
	if allPatterns, err := patterns.LoadPatterns(patterns.PatternListTypeAll); err != nil {
		logrus.WithError(err).Warn("failed to load anonymizer patterns; anonymizeText mutation will be unavailable")
	} else {
		allPatterns.Patterns = append(allPatterns.Patterns, cfg.GetSecretPatterns()...)

		if r, err := anonymizer.NewReplacer(allPatterns.Regexes(), allPatterns.Names()); err != nil {
			logrus.WithError(err).Warn("failed to create anonymizer replacer; anonymizeText mutation will be unavailable")
		} else {
			textReplacer = r
		}
	}

	// services
	authService := services.NewAuthService(
		services.AuthServiceConfig{
			BaseURL:          baseURL,
			LoginCallbackURL: oauthLoginCallbackURL,
			SessionTimeout:   sessionTimeout,
			CookiePrefix:     cfg.TenantPrefix(),
		},
		orm,
		oauthClients,
		userCache,
	)
	userService := services.NewUserService(orm, userCache, baseURL, sessionTimeout)
	roleService := services.NewRoleService(orm)
	providerService := services.NewProviderService(providers)
	settingsService := services.NewSettingsService(cfg)
	flowService := services.NewFlowService(orm, db, providers, controller, subscriptions)
	flowFileService := services.NewFlowFileService(orm, cfg.DataDir, cfg.TenantPrefix(), dockerClient, subscriptions)
	resourceService := services.NewResourceService(orm, cfg.DataDir, subscriptions)
	taskService := services.NewTaskService(orm)
	subtaskService := services.NewSubtaskService(orm)
	containerService := services.NewContainerService(orm)
	toolcallService := services.NewToolcallService(orm)
	assistantService := services.NewAssistantService(orm, providers, controller, subscriptions)
	agentlogService := services.NewAgentlogService(orm)
	assistantlogService := services.NewAssistantlogService(orm)
	msglogService := services.NewMsglogService(orm)
	searchlogService := services.NewSearchlogService(orm)
	vecstorelogService := services.NewVecstorelogService(orm)
	termlogService := services.NewTermlogService(orm)
	screenshotService := services.NewScreenshotService(orm, cfg.DataDir)
	promptService := services.NewPromptService(orm)
	timezones := timezone.NewCatalog(orm.DB())
	analyticsService := services.NewAnalyticsService(orm, timezones)
	tokenService := services.NewTokenService(orm, cfg.AuthSalt(), tokenCache, subscriptions)
	knowledgeService := services.NewKnowledgeService(orm, knowledgeStore)
	anonymizerService := services.NewAnonymizerService(textReplacer)
	graphqlService := services.NewGraphqlService(
		db, cfg, baseURL, cfg.CorsOrigins, tokenCache, providers, controller, subscriptions, knowledgeStore, textReplacer,
		updates, timezones,
	)

	router := newEngine(cfg.TrustedProxies)

	// Setup Cross-Origin Resource Sharing policy
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
	router.Use(compressionMiddleware())

	// AuthSalt mixes TENANT_ID into the key derivation so a session minted by one
	// instance is cryptographically invalid on another even when COOKIE_SIGNING_SALT
	// is shared. ScopedName separates the cookie itself, because cookies are scoped
	// by host and NOT by port — two instances on one host would otherwise overwrite
	// each other's sessions. Both are identity operations when TENANT_ID is empty.
	cookieStore := cookie.NewStore(auth.MakeCookieStoreKey(cfg.AuthSalt())...)
	router.Use(sessions.Sessions(cfg.ScopedName("auth"), cookieStore))

	api := router.Group(baseURL)
	api.Use(noCacheMiddleware())
	api.Use(requestDeadline(RequestTimeout))

	// Special case for local user own password change
	changePasswordGroup := api.Group("/user")
	changePasswordGroup.Use(authMiddleware.AuthUserRequired)
	changePasswordGroup.Use(localUserRequired())
	changePasswordGroup.PUT("/password", userService.ChangePasswordCurrentUser)
	changePasswordGroup.PUT("/email", userService.ChangeEmailCurrentUser)

	// Unlike password/email, the display name is editable by OAuth users too — no localUserRequired.
	changeNameGroup := api.Group("/user")
	changeNameGroup.Use(authMiddleware.AuthUserRequired)
	changeNameGroup.PUT("/name", userService.ChangeNameCurrentUser)

	publicGroup := api.Group("/")
	publicGroup.Use(authMiddleware.TryAuth)
	{
		publicGroup.GET("/info", authService.Info)

		developerGroup := publicGroup.Group("/")
		{
			developerGroup.GET("/graphql/playground", graphqlService.ServeGraphqlPlayground)
			developerGroup.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
		}

		registerAuthRoutes(publicGroup.Group("/auth"), authService)
	}

	privateGroup := api.Group("/")
	privateGroup.Use(authMiddleware.AuthTokenRequired)

	privateUserGroup := api.Group("/")
	privateUserGroup.Use(authMiddleware.AuthUserRequired)

	registerPrivateRoutes(privateGroup, privateUserGroup, privateServices{
		agentlog:     agentlogService,
		analytics:    analyticsService,
		anonymizer:   anonymizerService,
		assistant:    assistantService,
		assistantlog: assistantlogService,
		container:    containerService,
		flow:         flowService,
		flowFile:     flowFileService,
		graphql:      graphqlService,
		knowledge:    knowledgeService,
		msglog:       msglogService,
		prompt:       promptService,
		provider:     providerService,
		resource:     resourceService,
		role:         roleService,
		screenshot:   screenshotService,
		searchlog:    searchlogService,
		settings:     settingsService,
		subtask:      subtaskService,
		task:         taskService,
		termlog:      termlogService,
		token:        tokenService,
		toolcall:     toolcallService,
		user:         userService,
		vecstorelog:  vecstorelogService,
	})

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
		registerStaticFileServer(router, cfg.StaticDir)
	}

	return router
}

// registerStaticFileServer serves the locally-built SPA (used when no STATIC_URL
// upstream is set): cache headers, hashed assets via static.Serve, and an SPA
// fallback that serves index.html for client routes but returns 404 for a missing
// /assets/* — so the module loader fails cleanly and the app can reload to recover
// instead of getting index.html and a MIME error. Split out to be unit-testable.
func registerStaticFileServer(router *gin.Engine, staticDir string) {
	router.Use(staticCacheMiddleware())
	router.Use(static.Serve("/", static.LocalFile(staticDir, true)))

	indexExists := true
	indexPath := filepath.Join(staticDir, "index.html")
	if _, err := os.Stat(indexPath); err != nil {
		indexExists = false
	}

	router.NoRoute(func(c *gin.Context) {
		if c.Request.Method == http.MethodGet && !strings.HasPrefix(c.Request.URL.Path, baseURL) {
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

			if strings.HasPrefix(path, "/assets/") {
				// A missing hashed asset may be a transient rolling-deploy
				// race, so override the immutable directive set above —
				// never cache this 404 as a permanent negative.
				c.Header("Cache-Control", "no-store")
				c.Status(http.StatusNotFound)
				return
			}
		}

		c.Redirect(http.StatusMovedPermanently, "/")
	})
}

func setKnowledgeGroup(parent *gin.RouterGroup, svc *services.KnowledgeService) {
	kg := parent.Group("/knowledge")
	{
		kg.GET("/", auth.PrivilegesRequired("knowledge.view"), svc.ListDocuments)
		kg.GET("/:id", auth.PrivilegesRequired("knowledge.view"), svc.GetDocument)
		kg.POST("/", auth.PrivilegesRequired("knowledge.create"), svc.CreateDocument)
		kg.POST("/search", auth.PrivilegesRequired("knowledge.search"), svc.SearchDocuments)
		kg.PUT("/:id", auth.PrivilegesRequired("knowledge.edit"), svc.UpdateDocument)
		kg.DELETE("/:id", auth.PrivilegesRequired("knowledge.delete"), svc.DeleteDocument)
	}
}

func setProvidersGroup(parent *gin.RouterGroup, svc *services.ProviderService) {
	providersGroup := parent.Group("/providers")
	{
		providersGroup.GET("/", svc.GetProviders)
	}
}

func setSettingsGroup(parent *gin.RouterGroup, svc *services.SettingsService) {
	settingsGroup := parent.Group("/settings")
	{
		settingsGroup.GET("/", svc.GetSettings)
	}
}

type privateServices struct {
	agentlog     *services.AgentlogService
	analytics    *services.AnalyticsService
	anonymizer   *services.AnonymizerService
	assistant    *services.AssistantService
	assistantlog *services.AssistantlogService
	container    *services.ContainerService
	flow         *services.FlowService
	flowFile     *services.FlowFileService
	graphql      *services.GraphqlService
	knowledge    *services.KnowledgeService
	msglog       *services.MsglogService
	prompt       *services.PromptService
	provider     *services.ProviderService
	resource     *services.ResourceService
	role         *services.RoleService
	screenshot   *services.ScreenshotService
	searchlog    *services.SearchlogService
	settings     *services.SettingsService
	subtask      *services.SubtaskService
	task         *services.TaskService
	termlog      *services.TermlogService
	token        *services.TokenService
	toolcall     *services.ToolcallService
	user         *services.UserService
	vecstorelog  *services.VecstorelogService
}

func registerAuthRoutes(authGroup *gin.RouterGroup, authService *services.AuthService) {
	authGroup.POST("/login", authService.AuthLogin)
	authGroup.POST("/logout", authService.AuthLogout)
	authGroup.GET("/authorize", authService.AuthAuthorize)
	authGroup.GET("/login-callback", authService.AuthLoginGetCallback)
	authGroup.POST("/login-callback", authService.AuthLoginPostCallback)
	authGroup.POST("/logout-callback", authService.AuthLogoutCallback)
}

// Anything that administers accounts or credentials belongs on userTier: a bearer
// token reaching /users or /tokens can mint an administrator and escape its scope.
func registerPrivateRoutes(tokenTier, userTier *gin.RouterGroup, svc privateServices) {
	setGraphqlGroup(tokenTier, svc.graphql)

	setKnowledgeGroup(tokenTier, svc.knowledge)
	setProvidersGroup(tokenTier, svc.provider)
	setSettingsGroup(tokenTier, svc.settings)
	setFlowsGroup(tokenTier, svc.flow)
	setFlowFilesGroup(tokenTier, svc.flowFile)
	setResourcesGroup(tokenTier, svc.resource)
	setTasksGroup(tokenTier, svc.task)
	setSubtasksGroup(tokenTier, svc.subtask)
	setContainersGroup(tokenTier, svc.container)
	setToolcallsGroup(tokenTier, svc.toolcall)
	setAssistantsGroup(tokenTier, svc.assistant)
	setAgentlogsGroup(tokenTier, svc.agentlog)
	setAssistantlogsGroup(tokenTier, svc.assistantlog)
	setMsglogsGroup(tokenTier, svc.msglog)
	setTermlogsGroup(tokenTier, svc.termlog)
	setSearchlogsGroup(tokenTier, svc.searchlog)
	setVecstorelogsGroup(tokenTier, svc.vecstorelog)
	setScreenshotsGroup(tokenTier, svc.screenshot)
	setPromptsGroup(tokenTier, svc.prompt)
	setAnonymizeGroup(tokenTier, svc.anonymizer)
	setAnalyticsGroup(tokenTier, svc.analytics)

	setCurrentUserGroup(tokenTier, svc.user)

	setRolesGroup(userTier, svc.role)
	setUsersGroup(userTier, svc.user)
	setTokensGroup(userTier, svc.token)
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

func setFlowFilesGroup(parent *gin.RouterGroup, svc *services.FlowFileService) {
	flowFilesGroup := parent.Group("/flows/:flowID/files")
	{
		flowFilesGroup.GET("/", svc.GetFlowFiles)
		flowFilesGroup.GET("/container", svc.GetFlowContainerFiles)
		flowFilesGroup.POST("/", svc.UploadFlowFiles)
		flowFilesGroup.DELETE("/", svc.DeleteFlowFile)
		flowFilesGroup.GET("/download", svc.DownloadFlowFile)
		flowFilesGroup.POST("/pull", svc.PullFlowFiles)
		flowFilesGroup.POST("/resources", svc.AddResourcesToFlow)
		flowFilesGroup.POST("/to-resources", svc.AddResourceFromFlow)
	}
}

func setResourcesGroup(parent *gin.RouterGroup, svc *services.ResourceService) {
	rg := parent.Group("/resources")
	{
		rg.GET("/", svc.ListResources)
		rg.POST("/", svc.UploadResources)
		rg.POST("/mkdir", svc.MkdirResource)
		rg.PUT("/move", svc.MoveResource)
		rg.POST("/copy", svc.CopyResource)
		rg.DELETE("/", svc.DeleteResource)
		rg.GET("/download", svc.DownloadResource)
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

func setToolcallsGroup(parent *gin.RouterGroup, svc *services.ToolcallService) {
	toolcallsViewGroup := parent.Group("/toolcalls")
	{
		toolcallsViewGroup.GET("/", svc.GetToolcalls)
	}

	flowToolcallsViewGroup := parent.Group("/flows/:flowID/toolcalls")
	{
		flowToolcallsViewGroup.GET("/", svc.GetFlowToolcalls)
		flowToolcallsViewGroup.GET("/:toolcallID", svc.GetFlowToolcall)
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

func setAnonymizeGroup(parent *gin.RouterGroup, svc *services.AnonymizerService) {
	group := parent.Group("/anonymize")
	{
		group.POST("/text", svc.AnonymizeText)
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

}

func setCurrentUserGroup(parent *gin.RouterGroup, svc *services.UserService) {
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
