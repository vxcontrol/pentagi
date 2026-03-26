package services

import (
	"context"
	"fmt"
	"net/http"
	"slices"
	"strings"
	"time"

	"pentagi/pkg/config"
	"pentagi/pkg/controller"
	"pentagi/pkg/database"
	"pentagi/pkg/graph"
	"pentagi/pkg/graph/subscriptions"
	"pentagi/pkg/providers"
	"pentagi/pkg/server/auth"
	"pentagi/pkg/server/logger"
	"pentagi/pkg/templates"

	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"github.com/vektah/gqlparser/v2/ast"
)

var (
	_ graphql.RawParams // @lint:ignore U1000
	_ graphql.Response  // @lint:ignore U1000
)

type GraphqlService struct {
	srv  *handler.Server
	play http.HandlerFunc
}

type originValidator struct {
	allowAll  bool
	allowed   []string
	wildcards [][]string
	wrappers  []string
}

func NewGraphqlService(
	db *database.Queries,
	cfg *config.Config,
	baseURL string,
	origins []string,
	tokenCache *auth.TokenCache,
	providers providers.ProviderController,
	controller controller.FlowController,
	subscriptions subscriptions.SubscriptionsController,
) *GraphqlService {
	srv := handler.New(graph.NewExecutableSchema(graph.Config{Resolvers: &graph.Resolver{
		DB:              db,
		Config:          cfg,
		Logger:          logrus.StandardLogger().WithField("component", "pentagi-gql-bl"),
		TokenCache:      tokenCache,
		DefaultPrompter: templates.NewDefaultPrompter(),
		ProvidersCtrl:   providers,
		Controller:      controller,
		Subscriptions:   subscriptions,
	}}))

	component := "pentagi-gql"
	srv.AroundResponses(logger.WithGqlLogger(component))
	logger := logrus.WithField("component", component)

	srv.AddTransport(transport.Options{})
	srv.AddTransport(transport.GET{})
	srv.AddTransport(transport.POST{})
	srv.AddTransport(transport.MultipartForm{
		MaxMemory: 32 << 20, // 32MB
	})

	srv.SetQueryCache(lru.New[*ast.QueryDocument](1000))

	srv.Use(extension.Introspection{})
	srv.Use(extension.AutomaticPersistedQuery{
		Cache: lru.New[string](100),
	})
	srv.Use(extension.FixedComplexityLimit(20000))

	ov := newOriginValidator(origins)

	// Add transport to support GraphQL subscriptions
	srv.AddTransport(&transport.Websocket{
		KeepAlivePingInterval: 10 * time.Second,
		InitFunc: func(ctx context.Context, initPayload transport.InitPayload) (context.Context, *transport.InitPayload, error) {
			uid, err := graph.GetUserID(ctx)
			if err != nil {
				return nil, nil, fmt.Errorf("unauthorized: invalid user: %v", err)
			}
			logger.WithField("uid", uid).Info("graphql websocket upgrade")

			return ctx, &initPayload, nil
		},
		Upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) (allowed bool) {
				return ov.validateOrigin(r.Header.Get("Origin"), r.Host)
			},
			EnableCompression: true,
		},
	})

	return &GraphqlService{
		srv:  srv,
		play: playground.Handler("GraphQL", baseURL+"/graphql"),
	}
}

// ServeGraphql is a function to perform graphql requests
// @Summary Perform graphql requests
// @Tags GraphQL
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param json body graphql.RawParams true "graphql request"
// @Success 200 {object} graphql.Response "graphql response"
// @Failure 400 {object} graphql.Response "invalid graphql request data"
// @Failure 403 {object} graphql.Response "unauthorized"
// @Failure 500 {object} graphql.Response "internal error on graphql request"
// @Router /graphql [post]
func (s *GraphqlService) ServeGraphql(c *gin.Context) {
	uid := c.GetUint64("uid")
	tid := c.GetString("tid")
	privs := c.GetStringSlice("prm")

	savedCtx := c.Request.Context()
	defer func() {
		c.Request = c.Request.WithContext(savedCtx)
	}()

	ctx := savedCtx
	ctx = graph.SetUserID(ctx, uid)
	ctx = graph.SetUserType(ctx, tid)
	ctx = graph.SetUserPermissions(ctx, privs)
	c.Request = c.Request.WithContext(ctx)

	s.srv.ServeHTTP(c.Writer, c.Request)
}

func (s *GraphqlService) ServeGraphqlPlayground(c *gin.Context) {
	s.play.ServeHTTP(c.Writer, c.Request)
}

func newOriginValidator(origins []string) *originValidator {
	var wRules [][]string

	for _, o := range origins {
		if !strings.Contains(o, "*") {
			continue
		}

		if c := strings.Count(o, "*"); c > 1 {
			continue
		}

		i := strings.Index(o, "*")
		if i == 0 {
			wRules = append(wRules, []string{"*", o[1:]})
			continue
		}
		if i == (len(o) - 1) {
			wRules = append(wRules, []string{o[:i], "*"})
			continue
		}

		wRules = append(wRules, []string{o[:i], o[i+1:]})
	}

	return &originValidator{
		allowAll:  slices.Contains(origins, "*"),
		allowed:   origins,
		wildcards: wRules,
		wrappers:  []string{"http://", "https://", "ws://", "wss://"},
	}
}

func (ov *originValidator) validateOrigin(origin, host string) bool {
	if ov.allowAll {
		// CORS for origin '*' is allowed
		return true
	}

	if len(origin) == 0 {
		// request is not a CORS request
		return true
	}

	for _, wrapper := range ov.wrappers {
		if origin == wrapper+host {
			// request is not a CORS request but have origin header
			return true
		}
	}

	if slices.Contains(ov.allowed, origin) {
		return true
	}

	for _, w := range ov.wildcards {
		if w[0] == "*" && strings.HasSuffix(origin, w[1]) {
			return true
		}
		if w[1] == "*" && strings.HasPrefix(origin, w[0]) {
			return true
		}
		if strings.HasPrefix(origin, w[0]) && strings.HasSuffix(origin, w[1]) {
			return true
		}
	}

	return false
}
