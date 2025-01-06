package langfuse

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"net/http"
	"slices"
	"strings"
	"time"

	"pentagi/pkg/observability/langfuse/api"
	"pentagi/pkg/observability/langfuse/api/client"
	"pentagi/pkg/observability/langfuse/api/core"
	"pentagi/pkg/observability/langfuse/api/option"
)

const InstrumentationVersion = "1.0.0"

type CommentsClient interface {
	Create(ctx context.Context, request *api.CreateCommentRequest, opts ...core.RequestOption) (*api.CreateCommentResponse, error)
	Get(ctx context.Context, request *api.CommentsGetRequest, opts ...core.RequestOption) (*api.GetCommentsResponse, error)
	GetById(ctx context.Context, commentId string, opts ...core.RequestOption) (*api.Comment, error)
}

type DatasetitemsClient interface {
	Create(ctx context.Context, request *api.CreateDatasetItemRequest, opts ...core.RequestOption) (*api.DatasetItem, error)
	Get(ctx context.Context, id string, opts ...core.RequestOption) (*api.DatasetItem, error)
	List(ctx context.Context, request *api.DatasetItemsListRequest, opts ...core.RequestOption) (*api.PaginatedDatasetItems, error)
}

type DatasetrunitemsClient interface {
	Create(ctx context.Context, request *api.CreateDatasetRunItemRequest, opts ...core.RequestOption) (*api.DatasetRunItem, error)
}

type DatasetsClient interface {
	Create(ctx context.Context, request *api.CreateDatasetRequest, opts ...core.RequestOption) (*api.Dataset, error)
	Get(ctx context.Context, datasetName string, opts ...core.RequestOption) (*api.Dataset, error)
	Getrun(ctx context.Context, datasetName string, runName string, opts ...core.RequestOption) (*api.DatasetRunWithItems, error)
	Getruns(ctx context.Context, datasetName string, request *api.DatasetsGetRunsRequest, opts ...core.RequestOption) (*api.PaginatedDatasetRuns, error)
	List(ctx context.Context, request *api.DatasetsListRequest, opts ...core.RequestOption) (*api.PaginatedDatasets, error)
}

type HealthClient interface {
	Health(ctx context.Context, opts ...core.RequestOption) (*api.HealthResponse, error)
}

type IngestionClient interface {
	Batch(ctx context.Context, request *api.IngestionBatchRequest, opts ...core.RequestOption) (*api.IngestionResponse, error)
}

type MediaClient interface {
	Get(ctx context.Context, mediaId string, opts ...core.RequestOption) (*api.GetMediaResponse, error)
	Getuploadurl(ctx context.Context, request *api.GetMediaUploadUrlRequest, opts ...core.RequestOption) (*api.GetMediaUploadUrlResponse, error)
	Patch(ctx context.Context, mediaId string, request *api.PatchMediaBody, opts ...core.RequestOption) error
}

type MetricsClient interface {
	Daily(ctx context.Context, request *api.MetricsDailyRequest, opts ...core.RequestOption) (*api.DailyMetrics, error)
}

type ModelsClient interface {
	Create(ctx context.Context, request *api.CreateModelRequest, opts ...core.RequestOption) (*api.Model, error)
	Delete(ctx context.Context, id string, opts ...core.RequestOption) error
	Get(ctx context.Context, id string, opts ...core.RequestOption) (*api.Model, error)
	List(ctx context.Context, request *api.ModelsListRequest, opts ...core.RequestOption) (*api.PaginatedModels, error)
}

type ObservationsClient interface {
	Get(ctx context.Context, observationId string, opts ...core.RequestOption) (*api.ObservationsView, error)
	Getmany(ctx context.Context, request *api.ObservationsGetManyRequest, opts ...core.RequestOption) (*api.ObservationsViews, error)
}

type ProjectsClient interface {
	Get(ctx context.Context, opts ...core.RequestOption) (*api.Projects, error)
}

type PromptsClient interface {
	Create(ctx context.Context, request *api.CreatePromptRequest, opts ...core.RequestOption) (*api.Prompt, error)
	Get(ctx context.Context, promptName string, request *api.PromptsGetRequest, opts ...core.RequestOption) (*api.Prompt, error)
	List(ctx context.Context, request *api.PromptsListRequest, opts ...core.RequestOption) (*api.PromptMetaListResponse, error)
}

type ScoreconfigsClient interface {
	Create(ctx context.Context, request *api.CreateScoreConfigRequest, opts ...core.RequestOption) (*api.ScoreConfig, error)
	Get(ctx context.Context, request *api.ScoreConfigsGetRequest, opts ...core.RequestOption) (*api.ScoreConfigs, error)
	GetById(ctx context.Context, configId string, opts ...core.RequestOption) (*api.ScoreConfig, error)
}

type ScoreClient interface {
	Create(ctx context.Context, request *api.CreateScoreRequest, opts ...core.RequestOption) (*api.CreateScoreResponse, error)
	Delete(ctx context.Context, scoreId string, opts ...core.RequestOption) error
	Get(ctx context.Context, request *api.ScoreGetRequest, opts ...core.RequestOption) (*api.GetScoresResponse, error)
	GetById(ctx context.Context, scoreId string, opts ...core.RequestOption) (*api.Score, error)
}

type SessionsClient interface {
	Get(ctx context.Context, sessionId string, opts ...core.RequestOption) (*api.SessionWithTraces, error)
	List(ctx context.Context, request *api.SessionsListRequest, opts ...core.RequestOption) (*api.PaginatedSessions, error)
}

type TraceClient interface {
	Get(ctx context.Context, traceId string, opts ...core.RequestOption) (*api.TraceWithFullDetails, error)
	List(ctx context.Context, request *api.TraceListRequest, opts ...core.RequestOption) (*api.Traces, error)
}

type Client struct {
	Comments        CommentsClient
	Datasetitems    DatasetitemsClient
	Datasetrunitems DatasetrunitemsClient
	Datasets        DatasetsClient
	Health          HealthClient
	Ingestion       IngestionClient
	Media           MediaClient
	Metrics         MetricsClient
	Models          ModelsClient
	Observations    ObservationsClient
	Projects        ProjectsClient
	Prompts         PromptsClient
	Scoreconfigs    ScoreconfigsClient
	Score           ScoreClient
	Sessions        SessionsClient
	Trace           TraceClient

	publicKey string
	projectID string
}

func (c *Client) PublicKey() string {
	return c.publicKey
}

func (c *Client) ProjectID() string {
	return c.projectID
}

type ClientContext struct {
	BaseURL     string
	PublicKey   string
	SecretKey   string
	ProjectID   string
	HTTPClient  *http.Client
	MaxAttempts int
}

func (c *ClientContext) Validate() error {
	if c.BaseURL == "" {
		return fmt.Errorf("base url is required")
	}
	if c.PublicKey == "" {
		return fmt.Errorf("public key is required")
	}
	if c.SecretKey == "" {
		return fmt.Errorf("secret key is required")
	}
	if c.ProjectID == "" {
		return fmt.Errorf("project id is required")
	}
	if c.HTTPClient == nil {
		return fmt.Errorf("http client is required")
	}
	return nil
}

type ClientContextOption func(*ClientContext)

func WithBaseURL(baseURL string) ClientContextOption {
	return func(c *ClientContext) {
		c.BaseURL = baseURL
	}
}

func WithPublicKey(publicKey string) ClientContextOption {
	return func(c *ClientContext) {
		c.PublicKey = publicKey
	}
}

func WithSecretKey(secretKey string) ClientContextOption {
	return func(c *ClientContext) {
		c.SecretKey = secretKey
	}
}

func WithProjectID(projectID string) ClientContextOption {
	return func(c *ClientContext) {
		c.ProjectID = projectID
	}
}

func WithHTTPClient(httpClient *http.Client) ClientContextOption {
	return func(c *ClientContext) {
		c.HTTPClient = httpClient
	}
}

func WithMaxAttempts(maxAttempts int) ClientContextOption {
	return func(c *ClientContext) {
		c.MaxAttempts = maxAttempts
	}
}

func NewClient(opts ...ClientContextOption) (*Client, error) {
	clientCtx := ClientContext{
		HTTPClient: &http.Client{
			Timeout: defaultTimeout,
			Transport: &http.Transport{
				MaxIdleConns:        5,
				IdleConnTimeout:     30 * time.Second,
				TLSHandshakeTimeout: 10 * time.Second,
				TLSClientConfig:     &tls.Config{InsecureSkipVerify: true},
			},
		},
	}
	for _, opt := range opts {
		opt(&clientCtx)
	}

	if err := clientCtx.Validate(); err != nil {
		return nil, err
	}

	publicKey := strings.TrimSpace(clientCtx.PublicKey)
	secretKey := strings.TrimSpace(clientCtx.SecretKey)
	authToken := base64.StdEncoding.EncodeToString([]byte(publicKey + ":" + secretKey))
	options := []option.RequestOption{
		option.WithBaseURL(clientCtx.BaseURL),
		option.WithHTTPClient(clientCtx.HTTPClient),
		option.WithHTTPHeader(http.Header{
			"User-Agent":             []string{"langfuse golang sdk"},
			"Authorization":          []string{"Basic " + authToken},
			"x_fern_language":        []string{"golang"},
			"x_langfuse_sdk_name":    []string{"langfuse-observability-client-go"},
			"x_langfuse_sdk_version": []string{InstrumentationVersion},
			"x_langfuse_public_key":  []string{clientCtx.PublicKey},
			"x_langfuse_project_id":  []string{clientCtx.ProjectID},
		}),
	}

	if clientCtx.MaxAttempts > 0 {
		options = append(options, option.WithMaxAttempts(uint(clientCtx.MaxAttempts)))
	}

	client := client.NewClient(options...)

	resp, err := client.Projects.Get(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to get projects list: %w", err)
	}

	idxProject := slices.IndexFunc(resp.Data, func(p *api.Project) bool {
		return p.Id == clientCtx.ProjectID
	})
	if idxProject == -1 {
		return nil, fmt.Errorf("project not found")
	}

	return &Client{
		Comments:        client.Comments,
		Datasetitems:    client.Datasetitems,
		Datasetrunitems: client.Datasetrunitems,
		Datasets:        client.Datasets,
		Health:          client.Health,
		Ingestion:       client.Ingestion,
		Media:           client.Media,
		Metrics:         client.Metrics,
		Models:          client.Models,
		Observations:    client.Observations,
		Projects:        client.Projects,
		Prompts:         client.Prompts,
		Scoreconfigs:    client.Scoreconfigs,
		Score:           client.Score,
		Sessions:        client.Sessions,
		Trace:           client.Trace,
		publicKey:       clientCtx.PublicKey,
		projectID:       clientCtx.ProjectID,
	}, nil
}
