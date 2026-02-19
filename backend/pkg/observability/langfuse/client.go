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
	"pentagi/pkg/observability/langfuse/api/option"
)

const InstrumentationVersion = "2.0.0"

type AnnotationQueuesClient interface {
	Createqueue(ctx context.Context, request *api.CreateAnnotationQueueRequest, opts ...option.RequestOption) (*api.AnnotationQueue, error)
	Createqueueassignment(ctx context.Context, request *api.AnnotationQueuesCreateQueueAssignmentRequest, opts ...option.RequestOption) (*api.CreateAnnotationQueueAssignmentResponse, error)
	Createqueueitem(ctx context.Context, request *api.CreateAnnotationQueueItemRequest, opts ...option.RequestOption) (*api.AnnotationQueueItem, error)
	Deletequeueassignment(ctx context.Context, request *api.AnnotationQueuesDeleteQueueAssignmentRequest, opts ...option.RequestOption) (*api.DeleteAnnotationQueueAssignmentResponse, error)
	Deletequeueitem(ctx context.Context, request *api.AnnotationQueuesDeleteQueueItemRequest, opts ...option.RequestOption) (*api.DeleteAnnotationQueueItemResponse, error)
	Getqueue(ctx context.Context, request *api.AnnotationQueuesGetQueueRequest, opts ...option.RequestOption) (*api.AnnotationQueue, error)
	Getqueueitem(ctx context.Context, request *api.AnnotationQueuesGetQueueItemRequest, opts ...option.RequestOption) (*api.AnnotationQueueItem, error)
	Listqueueitems(ctx context.Context, request *api.AnnotationQueuesListQueueItemsRequest, opts ...option.RequestOption) (*api.PaginatedAnnotationQueueItems, error)
	Listqueues(ctx context.Context, request *api.AnnotationQueuesListQueuesRequest, opts ...option.RequestOption) (*api.PaginatedAnnotationQueues, error)
	Updatequeueitem(ctx context.Context, request *api.UpdateAnnotationQueueItemRequest, opts ...option.RequestOption) (*api.AnnotationQueueItem, error)
}

type BlobStorageIntegrationsClient interface {
	Deleteblobstorageintegration(ctx context.Context, request *api.BlobStorageIntegrationsDeleteBlobStorageIntegrationRequest, opts ...option.RequestOption) (*api.BlobStorageIntegrationDeletionResponse, error)
	Getblobstorageintegrations(ctx context.Context, opts ...option.RequestOption) (*api.BlobStorageIntegrationsResponse, error)
	Upsertblobstorageintegration(ctx context.Context, request *api.CreateBlobStorageIntegrationRequest, opts ...option.RequestOption) (*api.BlobStorageIntegrationResponse, error)
}

type CommentsClient interface {
	Create(ctx context.Context, request *api.CreateCommentRequest, opts ...option.RequestOption) (*api.CreateCommentResponse, error)
	Get(ctx context.Context, request *api.CommentsGetRequest, opts ...option.RequestOption) (*api.GetCommentsResponse, error)
	GetByID(ctx context.Context, request *api.CommentsGetByIDRequest, opts ...option.RequestOption) (*api.Comment, error)
}

type DatasetitemsClient interface {
	Create(ctx context.Context, request *api.CreateDatasetItemRequest, opts ...option.RequestOption) (*api.DatasetItem, error)
	Delete(ctx context.Context, request *api.DatasetItemsDeleteRequest, opts ...option.RequestOption) (*api.DeleteDatasetItemResponse, error)
	Get(ctx context.Context, request *api.DatasetItemsGetRequest, opts ...option.RequestOption) (*api.DatasetItem, error)
	List(ctx context.Context, request *api.DatasetItemsListRequest, opts ...option.RequestOption) (*api.PaginatedDatasetItems, error)
}

type DatasetrunitemsClient interface {
	Create(ctx context.Context, request *api.CreateDatasetRunItemRequest, opts ...option.RequestOption) (*api.DatasetRunItem, error)
	List(ctx context.Context, request *api.DatasetRunItemsListRequest, opts ...option.RequestOption) (*api.PaginatedDatasetRunItems, error)
}

type DatasetsClient interface {
	Create(ctx context.Context, request *api.CreateDatasetRequest, opts ...option.RequestOption) (*api.Dataset, error)
	Deleterun(ctx context.Context, request *api.DatasetsDeleteRunRequest, opts ...option.RequestOption) (*api.DeleteDatasetRunResponse, error)
	Get(ctx context.Context, request *api.DatasetsGetRequest, opts ...option.RequestOption) (*api.Dataset, error)
	Getrun(ctx context.Context, request *api.DatasetsGetRunRequest, opts ...option.RequestOption) (*api.DatasetRunWithItems, error)
	Getruns(ctx context.Context, request *api.DatasetsGetRunsRequest, opts ...option.RequestOption) (*api.PaginatedDatasetRuns, error)
	List(ctx context.Context, request *api.DatasetsListRequest, opts ...option.RequestOption) (*api.PaginatedDatasets, error)
}

type HealthClient interface {
	Health(ctx context.Context, opts ...option.RequestOption) (*api.HealthResponse, error)
}

type IngestionClient interface {
	Batch(ctx context.Context, request *api.IngestionBatchRequest, opts ...option.RequestOption) (*api.IngestionResponse, error)
}

type MediaClient interface {
	Get(ctx context.Context, request *api.MediaGetRequest, opts ...option.RequestOption) (*api.GetMediaResponse, error)
	Getuploadurl(ctx context.Context, request *api.GetMediaUploadURLRequest, opts ...option.RequestOption) (*api.GetMediaUploadURLResponse, error)
	Patch(ctx context.Context, request *api.PatchMediaBody, opts ...option.RequestOption) error
}

type MetricsV2Client interface {
	Metrics(ctx context.Context, request *api.MetricsV2MetricsRequest, opts ...option.RequestOption) (*api.MetricsV2Response, error)
}

type MetricsClient interface {
	Metrics(ctx context.Context, request *api.MetricsMetricsRequest, opts ...option.RequestOption) (*api.MetricsResponse, error)
}

type ModelsClient interface {
	Create(ctx context.Context, request *api.CreateModelRequest, opts ...option.RequestOption) (*api.Model, error)
	Delete(ctx context.Context, request *api.ModelsDeleteRequest, opts ...option.RequestOption) error
	Get(ctx context.Context, request *api.ModelsGetRequest, opts ...option.RequestOption) (*api.Model, error)
	List(ctx context.Context, request *api.ModelsListRequest, opts ...option.RequestOption) (*api.PaginatedModels, error)
}

type ObservationsV2Client interface {
	Getmany(ctx context.Context, request *api.ObservationsV2GetManyRequest, opts ...option.RequestOption) (*api.ObservationsV2Response, error)
}

type ObservationsClient interface {
	Get(ctx context.Context, request *api.ObservationsGetRequest, opts ...option.RequestOption) (*api.ObservationsView, error)
	Getmany(ctx context.Context, request *api.ObservationsGetManyRequest, opts ...option.RequestOption) (*api.ObservationsViews, error)
}

type OpentelemetryClient interface {
	Exporttraces(ctx context.Context, request *api.OpentelemetryExportTracesRequest, opts ...option.RequestOption) (*api.OtelTraceResponse, error)
}

type OrganizationsClient interface {
	Deleteorganizationmembership(ctx context.Context, request *api.DeleteMembershipRequest, opts ...option.RequestOption) (*api.MembershipDeletionResponse, error)
	Deleteprojectmembership(ctx context.Context, request *api.OrganizationsDeleteProjectMembershipRequest, opts ...option.RequestOption) (*api.MembershipDeletionResponse, error)
	Getorganizationapikeys(ctx context.Context, opts ...option.RequestOption) (*api.OrganizationAPIKeysResponse, error)
	Getorganizationmemberships(ctx context.Context, opts ...option.RequestOption) (*api.MembershipsResponse, error)
	Getorganizationprojects(ctx context.Context, opts ...option.RequestOption) (*api.OrganizationProjectsResponse, error)
	Getprojectmemberships(ctx context.Context, request *api.OrganizationsGetProjectMembershipsRequest, opts ...option.RequestOption) (*api.MembershipsResponse, error)
	Updateorganizationmembership(ctx context.Context, request *api.MembershipRequest, opts ...option.RequestOption) (*api.MembershipResponse, error)
	Updateprojectmembership(ctx context.Context, request *api.OrganizationsUpdateProjectMembershipRequest, opts ...option.RequestOption) (*api.MembershipResponse, error)
}

type ProjectsClient interface {
	Create(ctx context.Context, request *api.ProjectsCreateRequest, opts ...option.RequestOption) (*api.Project, error)
	Createapikey(ctx context.Context, request *api.ProjectsCreateAPIKeyRequest, opts ...option.RequestOption) (*api.APIKeyResponse, error)
	Delete(ctx context.Context, request *api.ProjectsDeleteRequest, opts ...option.RequestOption) (*api.ProjectDeletionResponse, error)
	Deleteapikey(ctx context.Context, request *api.ProjectsDeleteAPIKeyRequest, opts ...option.RequestOption) (*api.APIKeyDeletionResponse, error)
	Get(ctx context.Context, opts ...option.RequestOption) (*api.Projects, error)
	Getapikeys(ctx context.Context, request *api.ProjectsGetAPIKeysRequest, opts ...option.RequestOption) (*api.APIKeyList, error)
	Update(ctx context.Context, request *api.ProjectsUpdateRequest, opts ...option.RequestOption) (*api.Project, error)
}

type PromptversionClient interface {
	Update(ctx context.Context, request *api.PromptVersionUpdateRequest, opts ...option.RequestOption) (*api.Prompt, error)
}

type PromptsClient interface {
	Create(ctx context.Context, request *api.CreatePromptRequest, opts ...option.RequestOption) (*api.Prompt, error)
	Delete(ctx context.Context, request *api.PromptsDeleteRequest, opts ...option.RequestOption) error
	Get(ctx context.Context, request *api.PromptsGetRequest, opts ...option.RequestOption) (*api.Prompt, error)
	List(ctx context.Context, request *api.PromptsListRequest, opts ...option.RequestOption) (*api.PromptMetaListResponse, error)
}

type SCIMClient interface {
	Createuser(ctx context.Context, request *api.SCIMCreateUserRequest, opts ...option.RequestOption) (*api.SCIMUser, error)
	Deleteuser(ctx context.Context, request *api.SCIMDeleteUserRequest, opts ...option.RequestOption) (*api.EmptyResponse, error)
	Getresourcetypes(ctx context.Context, opts ...option.RequestOption) (*api.ResourceTypesResponse, error)
	Getschemas(ctx context.Context, opts ...option.RequestOption) (*api.SchemasResponse, error)
	Getserviceproviderconfig(ctx context.Context, opts ...option.RequestOption) (*api.ServiceProviderConfig, error)
	Getuser(ctx context.Context, request *api.SCIMGetUserRequest, opts ...option.RequestOption) (*api.SCIMUser, error)
	Listusers(ctx context.Context, request *api.SCIMListUsersRequest, opts ...option.RequestOption) (*api.SCIMUsersListResponse, error)
}

type ScoreconfigsClient interface {
	Create(ctx context.Context, request *api.CreateScoreConfigRequest, opts ...option.RequestOption) (*api.ScoreConfig, error)
	Get(ctx context.Context, request *api.ScoreConfigsGetRequest, opts ...option.RequestOption) (*api.ScoreConfigs, error)
	GetByID(ctx context.Context, request *api.ScoreConfigsGetByIDRequest, opts ...option.RequestOption) (*api.ScoreConfig, error)
	Update(ctx context.Context, request *api.UpdateScoreConfigRequest, opts ...option.RequestOption) (*api.ScoreConfig, error)
}

type ScoreV2Client interface {
	Get(ctx context.Context, request *api.ScoreV2GetRequest, opts ...option.RequestOption) (*api.GetScoresResponse, error)
	GetByID(ctx context.Context, request *api.ScoreV2GetByIDRequest, opts ...option.RequestOption) (*api.Score, error)
}

type ScoreClient interface {
	Create(ctx context.Context, request *api.CreateScoreRequest, opts ...option.RequestOption) (*api.CreateScoreResponse, error)
	Delete(ctx context.Context, request *api.ScoreDeleteRequest, opts ...option.RequestOption) error
}

type SessionsClient interface {
	Get(ctx context.Context, request *api.SessionsGetRequest, opts ...option.RequestOption) (*api.SessionWithTraces, error)
	List(ctx context.Context, request *api.SessionsListRequest, opts ...option.RequestOption) (*api.PaginatedSessions, error)
}

type TraceClient interface {
	Delete(ctx context.Context, request *api.TraceDeleteRequest, opts ...option.RequestOption) (*api.DeleteTraceResponse, error)
	Deletemultiple(ctx context.Context, request *api.TraceDeleteMultipleRequest, opts ...option.RequestOption) (*api.DeleteTraceResponse, error)
	Get(ctx context.Context, request *api.TraceGetRequest, opts ...option.RequestOption) (*api.TraceWithFullDetails, error)
	List(ctx context.Context, request *api.TraceListRequest, opts ...option.RequestOption) (*api.Traces, error)
}

type Client struct {
	AnnotationQueues        AnnotationQueuesClient
	BlobStorageIntegrations BlobStorageIntegrationsClient
	Comments                CommentsClient
	Datasetitems            DatasetitemsClient
	Datasetrunitems         DatasetrunitemsClient
	Datasets                DatasetsClient
	Health                  HealthClient
	Ingestion               IngestionClient
	Media                   MediaClient
	MetricsV2               MetricsV2Client
	Metrics                 MetricsClient
	Models                  ModelsClient
	ObservationsV2          ObservationsV2Client
	Observations            ObservationsClient
	Opentelemetry           OpentelemetryClient
	Organizations           OrganizationsClient
	Projects                ProjectsClient
	Promptversion           PromptversionClient
	Prompts                 PromptsClient
	SCIM                    SCIMClient
	Scoreconfigs            ScoreconfigsClient
	ScoreV2                 ScoreV2Client
	Score                   ScoreClient
	Sessions                SessionsClient
	Trace                   TraceClient

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
		return p.ID == clientCtx.ProjectID
	})
	if idxProject == -1 {
		return nil, fmt.Errorf("project not found")
	}

	return &Client{
		AnnotationQueues:        client.Annotationqueues,
		BlobStorageIntegrations: client.Blobstorageintegrations,
		Comments:                client.Comments,
		Datasetitems:            client.Datasetitems,
		Datasetrunitems:         client.Datasetrunitems,
		Datasets:                client.Datasets,
		Health:                  client.Health,
		Ingestion:               client.Ingestion,
		Media:                   client.Media,
		MetricsV2:               client.Metricsv2,
		Metrics:                 client.Metrics,
		Models:                  client.Models,
		ObservationsV2:          client.Observationsv2,
		Observations:            client.Observations,
		Opentelemetry:           client.Opentelemetry,
		Organizations:           client.Organizations,
		Projects:                client.Projects,
		Promptversion:           client.Promptversion,
		Prompts:                 client.Prompts,
		SCIM:                    client.SCIM,
		Scoreconfigs:            client.Scoreconfigs,
		ScoreV2:                 client.Scorev2,
		Score:                   client.Score,
		Sessions:                client.Sessions,
		Trace:                   client.Trace,
		publicKey:               clientCtx.PublicKey,
		projectID:               clientCtx.ProjectID,
	}, nil
}
