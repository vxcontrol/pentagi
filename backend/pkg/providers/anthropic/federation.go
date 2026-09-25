package anthropic

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"pentagi/pkg/config"

	"github.com/sirupsen/logrus"
)

// The vendor SDKs' schedule: within advisoryRefresh of expiry a failed exchange keeps the cached token,
// within mandatoryRefresh it fails the call; both shrink for a short-lived token. advisoryRetryDelay keeps
// a burst of calls from presenting the same single-use token over and over.
const (
	advisoryRefresh    = 120 * time.Second
	mandatoryRefresh   = 30 * time.Second
	advisoryRetryDelay = 10 * time.Second
)

const (
	minMintedLifetime  = 60 * time.Second
	maxMintedLifetime  = 24 * time.Hour
	exchangeBodyLimit  = 64 << 10
	defaultBaseURL     = "https://api.anthropic.com/v1"
	jwtBearerGrantType = "urn:ietf:params:oauth:grant-type:jwt-bearer"
	rotateTokenFile    = "point ANTHROPIC_IDENTITY_TOKEN_FILE at a token your platform rotates"
)

var (
	clock           = time.Now
	exchangeTimeout = 30 * time.Second

	// One source per identity for the process: pentagi builds many doors over one identity, and a
	// single-use identity token is refused on its second exchange.
	sourcesMu sync.Mutex
	sources   = map[federationIdentity]*tokenSource{}
)

type federationIdentity struct {
	ruleID, organizationID, serviceAccountID, workspaceID, baseURL string
}

// A file is re-read on every exchange; a literal token's exp and jti are decoded once, unverified.
type identityToken struct {
	file    string
	literal string
	expiry  time.Time
	jti     bool
}

func literalIdentityToken(token string) identityToken {
	identity := identityToken{literal: strings.TrimSpace(token)}

	parts := strings.Split(identity.literal, ".")
	if len(parts) != 3 {
		return identity
	}
	payload, err := base64.RawURLEncoding.DecodeString(strings.TrimRight(parts[1], "="))
	if err != nil {
		return identity
	}
	var claims struct {
		Exp float64 `json:"exp"`
		JTI string  `json:"jti"`
	}
	if json.Unmarshal(payload, &claims) != nil {
		return identity
	}

	if claims.Exp > 0 {
		identity.expiry = time.Unix(int64(claims.Exp), 0).UTC()
	}
	identity.jti = claims.JTI != ""
	return identity
}

func (it identityToken) read() (string, error) {
	if it.file == "" {
		return it.literal, nil
	}

	body, err := os.ReadFile(it.file)
	if err != nil {
		return "", fmt.Errorf("read ANTHROPIC_IDENTITY_TOKEN_FILE: %w", err)
	}
	token := strings.TrimSpace(string(body))
	if token == "" {
		return "", fmt.Errorf("ANTHROPIC_IDENTITY_TOKEN_FILE %s holds no token", it.file)
	}
	return token, nil
}

func (it identityToken) notice(now time.Time) string {
	if it.file != "" {
		if _, err := it.read(); err != nil {
			return "anthropic: " + err.Error() + "; the provider stays enabled and reads the file again on every exchange"
		}
		return ""
	}

	var notice string
	switch {
	case it.expiry.IsZero():
		notice = "anthropic: ANTHROPIC_IDENTITY_TOKEN cannot be refreshed once it expires"
	case now.Before(it.expiry):
		notice = "anthropic: ANTHROPIC_IDENTITY_TOKEN cannot be refreshed once it expires at " +
			it.expiry.Format(time.RFC3339)
	default:
		notice = "anthropic: ANTHROPIC_IDENTITY_TOKEN expired at " + it.expiry.Format(time.RFC3339) +
			" and will not be exchanged"
	}
	if it.jti {
		notice += "; it carries a jti claim, so it can be exchanged only once"
	}
	return notice + "; " + rotateTokenFile + " to run longer than its lifetime"
}

type tokenSource struct {
	id       federationIdentity
	identity identityToken
	now      func() time.Time
	timeout  time.Duration

	mu        sync.Mutex
	token     string
	expiresAt time.Time
	lifetime  time.Duration
	spent     bool // a literal token carrying jti has been exchanged
	retryAt   time.Time
	inflight  *exchangeCall
}

type exchangeCall struct {
	done chan struct{}
	err  error
}

func sharedSource(id federationIdentity, identity identityToken) *tokenSource {
	sourcesMu.Lock()
	defer sourcesMu.Unlock()

	if source, ok := sources[id]; ok {
		return source
	}
	source := &tokenSource{id: id, identity: identity, now: clock, timeout: exchangeTimeout}
	sources[id] = source
	return source
}

func (s *tokenSource) accessToken(ctx context.Context, client *http.Client) (string, error) {
	s.mu.Lock()
	now := s.now()
	left := s.expiresAt.Sub(now)
	advisory, mandatory := s.refreshPoints()
	optional := s.token != "" && left > mandatory

	if optional && (left > advisory || s.inflight != nil || now.Before(s.retryAt) || s.spent) {
		token := s.token
		s.mu.Unlock()
		return token, nil
	}
	if s.spent {
		s.mu.Unlock()
		return "", fmt.Errorf("anthropic federation: the access token minted from ANTHROPIC_IDENTITY_TOKEN "+
			"cannot be renewed past %s: the identity token carries a jti claim, so it can be exchanged only once; %s",
			s.expiresAt.UTC().Format(time.RFC3339), rotateTokenFile)
	}

	call := s.inflight
	if call == nil {
		call = s.start(ctx, client)
	}
	s.mu.Unlock()

	select {
	case <-call.done:
	case <-ctx.Done():
		return "", ctx.Err()
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if call.err != nil && !(optional && s.now().Before(s.expiresAt)) {
		return "", call.err
	}
	return s.token, nil
}

// refreshPoints must be called with s.mu held.
func (s *tokenSource) refreshPoints() (advisory, mandatory time.Duration) {
	return min(advisoryRefresh, s.lifetime/2), min(mandatoryRefresh, s.lifetime/4)
}

// start must be called with s.mu held.
func (s *tokenSource) start(ctx context.Context, client *http.Client) *exchangeCall {
	call := &exchangeCall{done: make(chan struct{})}
	s.inflight = call

	go func() {
		// Detached from its caller: the endpoint may consume a single-use identity token after
		// the caller has gone, and only this answer can use it.
		ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), s.timeout)
		defer cancel()

		sent := s.now()
		token, lifetime, err := s.exchange(ctx, client)

		var keptUntil time.Time
		s.mu.Lock()
		s.inflight = nil
		if err != nil {
			call.err = err
			s.retryAt = s.now().Add(advisoryRetryDelay)
			if _, mandatory := s.refreshPoints(); s.token != "" && s.expiresAt.Sub(s.now()) > mandatory {
				keptUntil = s.expiresAt.Add(-mandatory)
			}
		} else {
			s.token, s.lifetime, s.expiresAt = token, lifetime, sent.Add(lifetime)
			s.spent = s.identity.jti
		}
		s.mu.Unlock()

		// Unlocked, as a log hook may block; before the waiters wake, so it precedes the calls it explains.
		if !keptUntil.IsZero() {
			logrus.WithError(err).Warnf("anthropic federation: refreshing the access token failed; the cached one "+
				"is used until %s, and calls fail after that unless a refresh succeeds", keptUntil.UTC().Format(time.RFC3339))
		}
		close(call.done)
	}()

	return call
}

func (s *tokenSource) exchange(ctx context.Context, client *http.Client) (string, time.Duration, error) {
	if !s.identity.expiry.IsZero() && !s.now().Before(s.identity.expiry) {
		return "", 0, fmt.Errorf("anthropic federation: ANTHROPIC_IDENTITY_TOKEN expired at %s and is not sent; %s",
			s.identity.expiry.Format(time.RFC3339), rotateTokenFile)
	}

	assertion, err := s.identity.read()
	if err != nil {
		return "", 0, fmt.Errorf("anthropic federation: %w", err)
	}

	payload, err := json.Marshal(struct {
		GrantType        string `json:"grant_type"`
		Assertion        string `json:"assertion"`
		FederationRuleID string `json:"federation_rule_id"`
		OrganizationID   string `json:"organization_id"`
		ServiceAccountID string `json:"service_account_id"`
		WorkspaceID      string `json:"workspace_id,omitempty"`
	}{jwtBearerGrantType, assertion, s.id.ruleID, s.id.organizationID, s.id.serviceAccountID, s.id.workspaceID})
	if err != nil {
		return "", 0, fmt.Errorf("anthropic federation: marshal the token exchange: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.id.baseURL+"/oauth/token", bytes.NewReader(payload))
	if err != nil {
		return "", 0, fmt.Errorf("anthropic federation: build the token exchange: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		if ctx.Err() != nil {
			return "", 0, fmt.Errorf("anthropic federation: the token exchange did not answer within %s: %w", s.timeout, err)
		}
		return "", 0, fmt.Errorf("anthropic federation: token exchange: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, exchangeBodyLimit))
	if err != nil {
		return "", 0, fmt.Errorf("anthropic federation: read the token exchange answer: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return "", 0, s.refused(resp.StatusCode, body, assertion)
	}

	var minted struct {
		AccessToken string          `json:"access_token"`
		ExpiresIn   json.RawMessage `json:"expires_in"`
	}
	if json.Unmarshal(body, &minted) != nil || minted.AccessToken == "" {
		return "", 0, errors.New("anthropic federation: the token exchange answered without an access token")
	}

	// The minted token is kept whatever its expires_in, as the identity token behind it may be spent. An
	// absent or unreadable value reads as 0; every value is clamped to the vendor's 60 s to 24 h in
	// seconds, so an absurd one cannot overflow a Duration.
	var seconds float64
	_ = json.Unmarshal(bytes.Trim(minted.ExpiresIn, `"`), &seconds)
	seconds = min(max(seconds, minMintedLifetime.Seconds()), maxMintedLifetime.Seconds())
	return minted.AccessToken, time.Duration(seconds * float64(time.Second)), nil
}

// Only the vendor's and RFC 6749's error shapes are quoted: a proxy page could carry anything. The vendor
// names the jti_reused reason but not its status, and RFC 6749 answers an invalid grant with 400.
func (s *tokenSource) refused(status int, body []byte, assertion string) error {
	if (status == http.StatusBadRequest || status == http.StatusUnauthorized) && bytes.Contains(body, []byte("jti_reused")) {
		if s.identity.file == "" {
			return fmt.Errorf("anthropic federation: ANTHROPIC_IDENTITY_TOKEN was already exchanged (jti_reused) "+
				"and a single-use token is accepted only once; %s", rotateTokenFile)
		}
		return fmt.Errorf("anthropic federation: the token in %s was already exchanged (jti_reused); the file must "+
			"hold a new token before each refresh, so rotate it well within the minted token's lifetime", s.identity.file)
	}

	reason := http.StatusText(status)
	var answer struct {
		Error       json.RawMessage `json:"error"`
		Description string          `json:"error_description"`
	}
	if json.Unmarshal(body, &answer) == nil {
		var vendor struct {
			Message string `json:"message"`
		}
		var code string
		switch {
		case json.Unmarshal(answer.Error, &vendor) == nil && vendor.Message != "":
			reason = vendor.Message
		case json.Unmarshal(answer.Error, &code) == nil && code != "":
			reason = strings.TrimSpace(code + " " + answer.Description)
		}
	}

	return fmt.Errorf("anthropic federation: token exchange refused with status %d: %s",
		status, strings.ReplaceAll(reason, assertion, "[identity token]"))
}

type federatedDoer struct {
	client *http.Client
	source *tokenSource
}

func (d federatedDoer) Do(req *http.Request) (*http.Response, error) {
	token, err := d.source.accessToken(req.Context(), d.client)
	if err != nil {
		return nil, err
	}

	req = req.Clone(req.Context())
	req.Header.Del("X-Api-Key")
	req.Header.Set("Authorization", "Bearer "+token)
	return d.client.Do(req)
}

// FederatedBearer returns the access token every federated Anthropic door of cfg shares, exchanging the
// identity token through client when none is cached or it is due for refresh. A nil client is
// http.DefaultClient.
func FederatedBearer(ctx context.Context, cfg *config.Config, client *http.Client) (string, error) {
	id, identity, missing := federation(cfg)
	if len(missing) > 0 {
		return "", fmt.Errorf("anthropic federation: the config is incomplete (missing %s)", strings.Join(missing, ", "))
	}
	if client == nil {
		client = http.DefaultClient
	}

	return sharedSource(id, identity).accessToken(ctx, client)
}
