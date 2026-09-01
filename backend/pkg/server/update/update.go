// Package update asks the update service what this installation should be running.
//
// An update check made by the installer says which artefacts are on the machine
// and nothing about the server itself, and it only happens when somebody is
// installing something. This asks directly and on a schedule: it sends the build
// that is running and a summary of how the server is set up, and is told which
// build to move to — as a container image to pull, and, where one is published, as
// an executable that can replace this process's own binary.
//
// What it sends about the installation is produced by database.GetInstanceSummary
// and carried through untouched: quantities and closed vocabularies, nothing that
// identifies a person and nothing anybody wrote. The summary is versioned inside
// itself, so its shape can grow without this package, or anything in between,
// knowing what it holds.
//
// Nothing here ACTS on the answer. The verdict is kept, as a Status the rest of the
// server can read and show — the version badge in the UI is what it exists for —
// but applying an update is a decision for whoever operates the installation, and
// this package's whole contract with the rest of the server is that it cannot fail
// it and cannot slow it down.
package update

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"pentagi/pkg/config"
	"pentagi/pkg/database"
	"pentagi/pkg/providers"
	"pentagi/pkg/providers/embeddings"
	"pentagi/pkg/version"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/vxcontrol/cloud/models"
	"github.com/vxcontrol/cloud/sdk"
)

const (
	// The route name IS the endpoint code the service rate-limits, quotas and
	// authorises by — a name of our own invention is a name it does not recognise,
	// and the failure is a 403 or a quota rejection rather than "no such endpoint".
	// Neither this nor the path is ours to choose.
	routeProductUpdate = "pentagi_updates_check"
	pathProductUpdate  = "/api/v1/proxy/updates/pentagi/check"

	// requestTimeout bounds one exchange. The transport fetches a proof-of-work
	// ticket first, so the budget covers two round trips and the work between.
	requestTimeout = 30 * time.Second

	// startupDelay staggers the first exchange.
	//
	// Every installation in the fleet starts its clock at its own boot, but a
	// power cut across a datacentre or a fleet-wide upgrade restarts many at once,
	// and a call that fires immediately would turn that into a thundering herd
	// against one endpoint. A random wait inside this window costs nothing and
	// spreads them.
	startupDelay = 2 * time.Minute
)

// State is what the last exchange with the update service established about the
// build this process runs.
type State string

const (
	// StateDisabled says update checks are switched off by configuration, or were
	// never configured well enough to run.
	StateDisabled State = "disabled"
	// StatePending says no exchange has completed since this process started.
	StatePending State = "pending"
	// StateUnreachable says every exchange so far failed before an answer arrived.
	StateUnreachable State = "unreachable"
	// StateUnknown says the service answered but could not say whether this build
	// has an update: it did not recognise the build, named no version to move to,
	// or publishes nothing for this platform.
	StateUnknown State = "unknown"
	// StateUpToDate says this build is the newest offered on its channel.
	StateUpToDate State = "up_to_date"
	// StateUpdateAvailable says a newer version is offered; Status.Latest names it.
	StateUpdateAvailable State = "update_available"
)

// Status is what this installation knows about updates to itself.
//
// A verdict, once reached, outlives the failures that follow it: an update that
// was available an hour ago is still available, and the build this process runs
// cannot change underneath it — a replaced binary is a new process, which starts
// from pending. What a failure changes is how old the verdict is, and FailedAt
// says so.
type Status struct {
	// Current is the build this process runs.
	Current string
	// State is the verdict of the most recent answer, or why there is none.
	State State
	// Latest is the newest version the service offered, when it named one. It is
	// kept beside StateUnknown too, when the service named the newest published
	// build without being able to say whether this one is behind it.
	Latest string
	// Strategy is the channel the answer is computed under.
	Strategy string
	// CheckedAt is when the answer behind State arrived; zero until one has.
	CheckedAt time.Time
	// FailedAt is when the most recent exchange failed; zero when the last one
	// produced an answer.
	FailedAt time.Time
}

// DisabledStatus describes an installation whose checks never run — because the
// service could not be built, typically over a missing or malformed installation
// id. To whoever reads the badge that is the same thing as checks switched off,
// and it still names the build, which is the half of the badge that never depends
// on the service.
func DisabledStatus(cfg *config.Config) Status {
	return Status{
		Current:  version.GetBinaryVersion(),
		State:    StateDisabled,
		Strategy: string(effectiveStrategy(cfg)),
	}
}

// Service sends this installation's state to the update service periodically and
// keeps what it answers.
type Service struct {
	cfg     *config.Config
	queries *database.Queries
	client  *client
	logger  *logrus.Entry

	mu     sync.RWMutex
	status Status
}

// New builds the service, or explains why there will not be one.
//
// A missing or malformed installation id is the only hard stop: a message that
// cannot say which installation it describes says nothing, and inventing an
// identity would create an installation that never existed.
func New(cfg *config.Config, queries *database.Queries) (*Service, error) {
	if cfg == nil || queries == nil {
		return nil, fmt.Errorf("update: configuration and database are both required")
	}

	built, err := newClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("update: %w", err)
	}

	return &Service{
		cfg:     cfg,
		queries: queries,
		client:  built,
		logger:  logrus.WithField("component", "update"),
		status: Status{
			Current:  version.GetBinaryVersion(),
			State:    StatePending,
			Strategy: string(effectiveStrategy(cfg)),
		},
	}, nil
}

// Status is a copy of what is currently known. Safe to call from any goroutine,
// and cheap enough to answer on every page load.
func (r *Service) Status() Status {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.status
}

// Run keeps going until the context ends.
//
// Nothing here is allowed to interrupt the product. Every failure is a debug line
// and the next tick tries again; an endpoint that is down, a licence that has
// expired, a network that is not there are all ordinary states of the world for
// something that talks to a service it does not depend on.
func (r *Service) Run(ctx context.Context) {
	interval := r.cfg.UpdateCheckInterval
	if interval <= 0 {
		r.logger.Debug("disabled by configuration")
		r.mu.Lock()
		r.status.State = StateDisabled
		r.mu.Unlock()
		return
	}

	select {
	case <-ctx.Done():
		return
	case <-time.After(time.Duration(rand.Int63n(int64(startupDelay)))):
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		r.sendOnce(ctx)

		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

// sendOnce gathers the current state and sends it.
func (r *Service) sendOnce(ctx context.Context) {
	ctx, cancel := context.WithTimeout(ctx, requestTimeout)
	defer cancel()

	summary := r.queries.GetInstanceSummary(ctx, database.SummaryOptions{
		Features: database.InstanceFeatures{
			Debug:              r.cfg.Debug,
			AskUser:            r.cfg.AskUser,
			DockerInside:       r.cfg.DockerInside,
			AssistantUseAgents: r.cfg.AssistantUseAgents,
		},
		OAuth:             configuredOAuthProviders(r.cfg),
		DefaultProviders:  providers.EnabledDefaultProviderTypes(r.cfg),
		EmbeddingProvider: embeddings.ResolvedProviderType(r.cfg),
	})

	document, err := json.Marshal(summary)
	if err != nil {
		r.recordFailure(fmt.Errorf("cannot render the state summary: %w", err))
		return
	}
	if len(document) > models.MaxProductInfoBytes {
		// Refusing here rather than being refused there. The service would reject
		// the whole request, and the reason would arrive as a validation error
		// about a field nobody can see from this side.
		r.recordFailure(fmt.Errorf("the state summary is %d bytes, larger than the service accepts",
			len(document)))
		return
	}

	request := models.ProductUpdateRequest{
		Schema:  uint16(summary.Schema),
		Version: normalizeVersion(summary.Version),
		// The platform of THIS PROCESS, which for a container is the platform of its
		// image — so it selects both the image the service matches us against and the
		// executable it offers beside it.
		//
		// Read out of the summary rather than computed here, exactly like the schema
		// and the version above. The document carries all three as well, because a
		// description forwarded by somebody else has to stand on its own, and two
		// places computing the same fact is how the envelope and the document come to
		// disagree about which build is running.
		OS:         models.OSType(summary.OS),
		Arch:       models.ArchType(summary.Arch),
		BinaryHash: summary.BinaryHash,
		Strategy:   updateStrategy(r.cfg),
		Info:       json.RawMessage(document),
	}

	answer, err := r.client.send(ctx, request)
	if err != nil {
		r.recordFailure(err)
		return
	}

	r.record(answer)
}

// record keeps what the service answered, and says so in the log. Nothing acts
// on it: a server that started pulling images or replacing its own binary on a
// timer would be doing something nobody asked it to.
func (r *Service) record(answer *models.ProductUpdateResponse) {
	verdict := interpret(answer)

	r.mu.Lock()
	r.status.State = verdict.state
	r.status.Latest = verdict.latest
	r.status.CheckedAt = time.Now()
	r.status.FailedAt = time.Time{}
	r.mu.Unlock()

	entry := r.logger.WithField("state", verdict.state)
	if answer != nil && answer.Update != nil {
		entry = entry.WithField("resolution", answer.Update.Resolution)
	}
	if verdict.latest != "" {
		entry = entry.WithField("offered", verdict.latest)
	}
	if answer != nil && answer.Binary != nil {
		entry = entry.WithField("executable", answer.Binary.Version)
	}
	entry.Debug(verdict.message)
}

// recordFailure notes that an exchange produced no answer.
//
// A verdict already reached stays — see Status — and only an installation that
// has never heard anything becomes unreachable: nothing is known, and the reason
// is that nothing could be asked.
func (r *Service) recordFailure(err error) {
	r.logger.WithError(err).Debug("not delivered")

	r.mu.Lock()
	defer r.mu.Unlock()
	r.status.FailedAt = time.Now()
	if r.status.State == StatePending {
		r.status.State = StateUnreachable
	}
}

// verdict is what one answer says, reduced to what the badge and the log need.
type verdict struct {
	state   State
	latest  string
	message string
}

// interpret reads an answer.
//
// `identified` is why this is more than one branch. When the service could not
// match the running build to an image it published — a development build, a
// private rebuild — the answer describes the newest published build rather than a
// step from this one, and reading it as "an update is available" would announce
// an update on every single check, forever. The version it names is still kept:
// "the newest published is 2.4.0" is worth showing beside "this build could not be
// matched".
//
// A `has_update` that names no version is not an update either. There is nothing
// to tell the operator to move to, and a badge pointing at nothing cannot be acted
// on — which is what the operator would try to do with it.
func interpret(answer *models.ProductUpdateResponse) verdict {
	if answer == nil || answer.Update == nil {
		return verdict{
			state:   StateUnknown,
			message: "delivered; no update information for this platform",
		}
	}

	latest := offeredVersion(answer)
	switch {
	case !answer.Identified:
		return verdict{
			state:  StateUnknown,
			latest: latest,
			message: "delivered; this build is not one we published, so the answer names " +
				"the newest available rather than a step from here",
		}
	case !answer.Update.HasUpdate:
		return verdict{state: StateUpToDate, latest: latest, message: "delivered; up to date"}
	case latest == "":
		return verdict{
			state:   StateUnknown,
			message: "delivered; an update is offered but no version is named for it",
		}
	default:
		return verdict{state: StateUpdateAvailable, latest: latest, message: "delivered; an update is available"}
	}
}

// offeredVersion is the version the answer points at, from whichever field
// carries it: the offered image's version first, the stack's latest release when
// the image declares none.
func offeredVersion(answer *models.ProductUpdateResponse) string {
	if answer.Version != nil && *answer.Version != "" {
		return *answer.Version
	}
	if answer.Update != nil && answer.Update.LatestVersion != nil {
		return *answer.Update.LatestVersion
	}
	return ""
}

// updateStrategy is which channel to be answered from.
//
// Sent rather than left to the service's default so that one installation reads the
// same channel everywhere: UPDATE_STRATEGY is what the installer applies to the
// images it pulls, and an answer here computed under a different one would disagree
// with what the operator configured. An unrecognised value is dropped rather than
// forwarded — the field is validated on both sides, and a whole check refused over
// a typo in a variable is worse than a check answered under the default.
func updateStrategy(cfg *config.Config) models.UpdateStrategy {
	if cfg == nil {
		return ""
	}
	strategy := models.UpdateStrategy(strings.TrimSpace(cfg.UpdateStrategy))
	if strategy.Valid() != nil {
		return ""
	}
	return strategy
}

// effectiveStrategy is the channel the service actually answers under: the
// configured one when it is forwarded, otherwise the default both sides share.
func effectiveStrategy(cfg *config.Config) models.UpdateStrategy {
	if strategy := updateStrategy(cfg); strategy != "" {
		return strategy
	}
	return models.UpdateStrategyPreview
}

// normalizeVersion converts a build version into one the service accepts.
//
// A development build is named after its branch rather than a release, and the
// service validates this field as a version. Sending the branch name would have
// the whole request refused over a field that is decoration; "0.0.0" says "some
// build, older than anything published", which is exactly true.
func normalizeVersion(raw string) string {
	if err := models.GetValidator().Var(raw, "semver"); err != nil {
		return "0.0.0"
	}
	return raw
}

// configuredOAuthProviders names the sign-in providers this server offers — the
// providers, never a user.
func configuredOAuthProviders(cfg *config.Config) []string {
	providers := []string{}
	if cfg.PublicURL == "" {
		return providers
	}
	if cfg.OAuthGoogleClientID != "" && cfg.OAuthGoogleClientSecret != "" {
		providers = append(providers, "google")
	}
	if cfg.OAuthGithubClientID != "" && cfg.OAuthGithubClientSecret != "" {
		providers = append(providers, "github")
	}
	return providers
}

// client is the transport half: one SDK call, built once.
type client struct {
	call sdk.CallReqBytesRespBytes
}

func newClient(cfg *config.Config) (*client, error) {
	buildVersion := version.GetBinaryVersion()

	installationID, err := uuid.Parse(cfg.InstallationID)
	if err != nil {
		return nil, fmt.Errorf("update: installation id %q is not a uuid: %w", cfg.InstallationID, err)
	}

	transport := sdk.DefaultTransport()
	if value := strings.TrimSpace(cfg.ProxyURL); value != "" {
		parsed, err := url.Parse(value)
		if err != nil {
			return nil, fmt.Errorf("proxy %q is not a valid URL: %w", cfg.ProxyURL, err)
		}
		if parsed.Host == "" {
			return nil, fmt.Errorf("proxy %q has no host", cfg.ProxyURL)
		}
		transport.Proxy = http.ProxyURL(parsed)
	}

	built := &client{}

	options := []sdk.Option{
		sdk.WithTransport(transport),
		sdk.WithInstallationID(installationID),
		sdk.WithClient("PentAGI", buildVersion),
	}

	// Validated before it is handed over: the SDK drops a key it cannot decode
	// without saying so, and the call then runs anonymously — which looks from
	// the outside exactly like a licence that is not being honoured.
	if key := strings.TrimSpace(cfg.LicenseKey); key != "" {
		if _, err := sdk.IntrospectLicenseKey(key); err == nil {
			options = append(options, sdk.WithLicenseKey(key))
		}
	}

	configs := []sdk.CallConfig{
		{
			Calls:  []any{&built.call},
			Host:   cfg.UpdateServerHost,
			Name:   routeProductUpdate,
			Path:   pathProductUpdate,
			Method: sdk.CallMethodPOST,
		},
	}

	if err := sdk.Build(configs, options...); err != nil {
		return nil, fmt.Errorf("failed to build the client: %w", err)
	}

	return built, nil
}

func (c *client) send(
	ctx context.Context, request models.ProductUpdateRequest,
) (*models.ProductUpdateResponse, error) {
	if c == nil || c.call == nil {
		return nil, fmt.Errorf("the client is not initialised")
	}
	if err := request.Valid(); err != nil {
		return nil, fmt.Errorf("the request does not satisfy the contract: %w", err)
	}

	body, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to encode the request: %w", err)
	}

	response, err := c.call(ctx, body)
	if err != nil {
		return nil, err
	}

	// Through the envelope, never straight into the payload. Every answer is
	// wrapped in {"status": …, "data": …}: decoding the body directly into the
	// payload type SUCCEEDS and yields a zero value, so a client that skips this
	// step reads every answer as "nothing to do" and never finds out.
	answer, err := models.ParseEnvelope[models.ProductUpdateResponse](response)
	if err != nil {
		return nil, fmt.Errorf("failed to decode the answer: %w", err)
	}

	return &answer, nil
}
