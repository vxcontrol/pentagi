package anthropic

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"pentagi/pkg/config"
	"pentagi/pkg/providers/pconfig"
	"pentagi/pkg/providers/provider"

	"github.com/sirupsen/logrus"
	logtest "github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func federatedDoor(t *testing.T, cfg *config.Config) provider.Provider {
	t.Helper()

	providerConfig, err := DefaultProviderConfig()
	require.NoError(t, err)
	prov, err := New(cfg, provider.DefaultProviderNameAnthropic, providerConfig)
	require.NoError(t, err)
	return prov
}

func federatedCall(ctx context.Context, prov provider.Provider) error {
	_, err := prov.Call(ctx, pconfig.OptionsTypeSimple, "hi")
	return err
}

// federationStep advances the clock by after and makes one call; exchanges counts every exchange so far.
type federationStep struct {
	after      time.Duration
	down       bool
	wantBearer string
	wantErr    string
	exchanges  int
}

// runFederationSteps mints the first token at the stubbed epoch, then plays steps; rotate, when set,
// writes a fresh identity token before each step.
func runFederationSteps(
	t *testing.T, clk *fakeClock, upstream *anthropicUpstream, prov provider.Provider,
	rotate func(step int), steps []federationStep,
) {
	t.Helper()

	require.NoError(t, federatedCall(context.Background(), prov))

	for i, step := range steps {
		if rotate != nil {
			rotate(i + 1)
		}
		clk.Advance(step.after)
		if step.down {
			upstream.answerExchanges(503, `{"type":"error","error":{"type":"api_error","message":"exchange unavailable"}}`)
		} else {
			upstream.answerExchanges(0, "")
		}

		_, before := upstream.authentication()
		err := federatedCall(context.Background(), prov)
		exchanges, auth := upstream.authentication()

		if step.wantErr != "" {
			require.ErrorContains(t, err, step.wantErr, "step %d", i)
			assert.Len(t, auth, len(before), "step %d reached the API", i)
		} else {
			require.NoError(t, err, "step %d", i)
			require.Len(t, auth, len(before)+1, "step %d", i)
			assert.Equal(t, anthropicAuth{authorization: "Bearer " + step.wantBearer}, auth[len(auth)-1], "step %d", i)
		}
		assert.Len(t, exchanges, step.exchanges, "step %d", i)
	}
}

func TestFederation_AccessToken_DoorsOfOneIdentityShareOneExchange(t *testing.T) {
	shared := []string{"Bearer minted-1", "Bearer minted-1"}
	tests := []struct {
		name     string
		second   func(cfg *config.Config)
		wantAuth []string
	}{
		{name: "a second door of the same identity", wantAuth: shared},
		{
			name:     "the same server named with a trailing slash",
			second:   func(cfg *config.Config) { cfg.AnthropicServerURL += "/" },
			wantAuth: shared,
		},
		{
			name:     "another federation rule",
			second:   func(cfg *config.Config) { cfg.AnthropicFederationRuleID = "fdrl_2" },
			wantAuth: []string{"Bearer minted-1", "Bearer minted-2"},
		},
		{
			name:     "another organization",
			second:   func(cfg *config.Config) { cfg.AnthropicOrganizationID = "11111111-1111-1111-1111-111111111111" },
			wantAuth: []string{"Bearer minted-1", "Bearer minted-2"},
		},
		{
			name:     "another service account",
			second:   func(cfg *config.Config) { cfg.AnthropicServiceAccountID = "svac_2" },
			wantAuth: []string{"Bearer minted-1", "Bearer minted-2"},
		},
		{
			name:     "another workspace",
			second:   func(cfg *config.Config) { cfg.AnthropicWorkspaceID = "wrkspc_2" },
			wantAuth: []string{"Bearer minted-1", "Bearer minted-2"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			upstream := anthropicServe(t)

			first := federatedConfig(upstream.url)
			second := federatedConfig(upstream.url)
			if tc.second != nil {
				tc.second(second)
				second.AnthropicIdentityToken = "second.jwt.token"
			}

			for _, cfg := range []*config.Config{first, second} {
				require.NoError(t, federatedCall(context.Background(), federatedDoor(t, cfg)),
					"a door must reuse its identity's minted token, not present the identity token again")
			}

			_, auth := upstream.authentication()
			var authorization []string
			for _, a := range auth {
				assert.Empty(t, a.apiKey, "the placeholder key left the process")
				authorization = append(authorization, a.authorization)
			}
			assert.Equal(t, tc.wantAuth, authorization)
		})
	}
}

func TestFederation_AccessToken_ConcurrentCallsMakeOneExchange(t *testing.T) {
	const callers = 8

	upstream := anthropicServe(t)
	release := upstream.holdExchanges(t)
	cfg := federatedConfig(upstream.url)
	doors := []provider.Provider{federatedDoor(t, cfg), federatedDoor(t, cfg)}

	started := make(chan struct{}, callers)
	errs := make(chan error, callers)
	for i := range callers {
		go func() {
			started <- struct{}{}
			errs <- federatedCall(context.Background(), doors[i%len(doors)])
		}()
	}
	for range callers {
		select {
		case <-started:
		case <-time.After(10 * time.Second):
			t.Fatal("a caller never started")
		}
	}
	upstream.awaitExchange(t)

	select {
	case <-upstream.arrived:
		t.Fatal("a second exchange started while the first was in flight")
	case <-time.After(200 * time.Millisecond):
	}
	release()

	for range callers {
		select {
		case err := <-errs:
			require.NoError(t, err)
		case <-time.After(10 * time.Second):
			t.Fatal("a caller was still waiting 10 s after the exchange answered")
		}
	}

	exchanges, auth := upstream.authentication()
	assert.Len(t, exchanges, 1)
	require.Len(t, auth, callers)
	for _, a := range auth {
		assert.Equal(t, "Bearer minted-1", a.authorization)
	}
}

func TestFederation_AccessToken_RefreshesOnTheVendorSchedule(t *testing.T) {
	const unavailable = "anthropic federation: token exchange refused with status 503: exchange unavailable"

	tests := []struct {
		name      string
		expiresIn string
		steps     []federationStep
	}{
		{
			name:      "a token short of its advisory point is served as it is",
			expiresIn: "3600",
			steps:     []federationStep{{after: 3479 * time.Second, wantBearer: "minted-1", exchanges: 1}},
		},
		{
			name:      "an advisory refresh replaces the token",
			expiresIn: "3600",
			steps:     []federationStep{{after: 3481 * time.Second, wantBearer: "minted-2", exchanges: 2}},
		},
		{
			name:      "a failed advisory refresh keeps serving the cached token and waits before the next try",
			expiresIn: "3600",
			steps: []federationStep{
				{after: 3481 * time.Second, down: true, wantBearer: "minted-1", exchanges: 2},
				{after: 9 * time.Second, down: true, wantBearer: "minted-1", exchanges: 2},
				{after: time.Second, wantBearer: "minted-2", exchanges: 3},
			},
		},
		{
			name:      "a failed mandatory refresh fails the call",
			expiresIn: "3600",
			steps:     []federationStep{{after: 3571 * time.Second, down: true, wantErr: unavailable, exchanges: 2}},
		},
		{
			name:      "an expired token is replaced",
			expiresIn: "3600",
			steps:     []federationStep{{after: 3601 * time.Second, wantBearer: "minted-2", exchanges: 2}},
		},
		{
			name:      "a one-minute token serves the first half of its life without an exchange",
			expiresIn: "60",
			steps: []federationStep{
				{wantBearer: "minted-1", exchanges: 1},
				{after: 29 * time.Second, wantBearer: "minted-1", exchanges: 1},
				{after: 2 * time.Second, wantBearer: "minted-2", exchanges: 2},
			},
		},
		{
			name:      "a one-minute token outlives a failed refresh until its last quarter",
			expiresIn: "60",
			steps: []federationStep{
				{after: 44 * time.Second, down: true, wantBearer: "minted-1", exchanges: 2},
				{after: 2 * time.Second, down: true, wantErr: unavailable, exchanges: 3},
			},
		},
		{
			name: "an answer without expires_in lives a minute",
			steps: []federationStep{
				{after: 29 * time.Second, wantBearer: "minted-1", exchanges: 1},
				{after: 2 * time.Second, wantBearer: "minted-2", exchanges: 2},
			},
		},
		{
			name:      "a negative expires_in lives a minute",
			expiresIn: "-5",
			steps: []federationStep{
				{after: 29 * time.Second, wantBearer: "minted-1", exchanges: 1},
				{after: 2 * time.Second, wantBearer: "minted-2", exchanges: 2},
			},
		},
		{
			name:      "a fractional expires_in is read",
			expiresIn: "3600.0",
			steps: []federationStep{
				{after: 3479 * time.Second, wantBearer: "minted-1", exchanges: 1},
				{after: 2 * time.Second, wantBearer: "minted-2", exchanges: 2},
			},
		},
		{
			name:      "a quoted expires_in is read",
			expiresIn: `"3600"`,
			steps: []federationStep{
				{after: 3479 * time.Second, wantBearer: "minted-1", exchanges: 1},
				{after: 2 * time.Second, wantBearer: "minted-2", exchanges: 2},
			},
		},
		{
			name:      "an expires_in beyond a day lives a day",
			expiresIn: "1e30",
			steps: []federationStep{
				{after: 86279 * time.Second, wantBearer: "minted-1", exchanges: 1},
				{after: 2 * time.Second, wantBearer: "minted-2", exchanges: 2},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			clk := stubClock(t)
			upstream := anthropicServe(t)
			upstream.mintExpiringIn(tc.expiresIn)

			cfg := federatedConfig(upstream.url)
			cfg.AnthropicIdentityTokenFile = filepath.Join(t.TempDir(), "token")
			rotate := func(step int) {
				require.NoError(t, os.WriteFile(cfg.AnthropicIdentityTokenFile,
					fmt.Appendf(nil, "identity-%d.jwt.token", step), 0o600))
			}
			rotate(0)

			runFederationSteps(t, clk, upstream, federatedDoor(t, cfg), rotate, tc.steps)
		})
	}
}

func TestFederation_AccessToken_FailsWhenTheTokenExpiresDuringAFailedRefresh(t *testing.T) {
	clk := stubClock(t)
	upstream := anthropicServe(t)
	prov := federatedDoor(t, federatedConfig(upstream.url))
	require.NoError(t, federatedCall(context.Background(), prov))
	upstream.awaitExchange(t)

	clk.Advance(3481 * time.Second)
	upstream.answerExchanges(503, `{"type":"error","error":{"type":"api_error","message":"exchange unavailable"}}`)
	release := upstream.holdExchanges(t)
	done := make(chan error, 1)
	go func() { done <- federatedCall(context.Background(), prov) }()
	upstream.awaitExchange(t)
	clk.Advance(120 * time.Second)
	release()

	select {
	case err := <-done:
		require.ErrorContains(t, err, "anthropic federation: token exchange refused with status 503: exchange unavailable")
	case <-time.After(10 * time.Second):
		t.Fatal("the call never came back from the refresh")
	}
	_, auth := upstream.authentication()
	assert.Len(t, auth, 1, "an expired token was sent")
}

func TestFederation_AccessToken_ServesTheCachedTokenWhileAnAdvisoryRefreshRuns(t *testing.T) {
	clk := stubClock(t)
	upstream := anthropicServe(t)
	prov := federatedDoor(t, federatedConfig(upstream.url))
	require.NoError(t, federatedCall(context.Background(), prov))
	upstream.awaitExchange(t)

	clk.Advance(3481 * time.Second)
	release := upstream.holdExchanges(t)
	refreshing := make(chan error, 1)
	go func() { refreshing <- federatedCall(context.Background(), prov) }()
	upstream.awaitExchange(t)

	served := make(chan error, 1)
	go func() { served <- federatedCall(context.Background(), prov) }()
	select {
	case err := <-served:
		require.NoError(t, err)
	case <-time.After(10 * time.Second):
		t.Fatal("a caller waited for an advisory refresh while the cached token was still good")
	}
	release()
	select {
	case err := <-refreshing:
		require.NoError(t, err)
	case <-time.After(10 * time.Second):
		t.Fatal("the refreshing caller never came back")
	}

	exchanges, auth := upstream.authentication()
	assert.Len(t, exchanges, 2)
	assert.Equal(t, []anthropicAuth{
		{authorization: "Bearer minted-1"}, {authorization: "Bearer minted-1"}, {authorization: "Bearer minted-1"},
	}, auth)
}

// Not parallel: it swaps the hooks of the global logger.
func TestFederation_AccessToken_WarnsWhenAFailedRefreshKeepsTheCachedToken(t *testing.T) {
	hook := new(logtest.Hook)
	previous := logrus.StandardLogger().ReplaceHooks(logrus.LevelHooks{})
	logrus.AddHook(hook)
	t.Cleanup(func() { logrus.StandardLogger().ReplaceHooks(previous) })

	clk := stubClock(t)
	upstream := anthropicServe(t)
	dir := t.TempDir()
	cfg := federatedConfig(upstream.url)
	cfg.AnthropicIdentityTokenFile = filepath.Join(dir, "token")
	require.NoError(t, os.WriteFile(cfg.AnthropicIdentityTokenFile, []byte("unrotated.jwt.token"), 0o600))
	prov := federatedDoor(t, cfg)
	require.NoError(t, federatedCall(context.Background(), prov))

	clk.Advance(3481 * time.Second)
	require.NoError(t, federatedCall(context.Background(), prov))
	clk.Advance(90 * time.Second)
	require.ErrorContains(t, federatedCall(context.Background(), prov), "(jti_reused)")

	entries := hook.AllEntries()
	require.Len(t, entries, 1, "only the advisory failure keeps the cached token")
	assert.Equal(t, logrus.WarnLevel, entries[0].Level)
	assert.Equal(t, "anthropic federation: refreshing the access token failed; the cached one is used until "+
		"2026-09-24T12:59:30Z, and calls fail after that unless a refresh succeeds", entries[0].Message)
	logged, _ := entries[0].Data[logrus.ErrorKey].(error)
	assert.EqualError(t, logged, "anthropic federation: the token in "+dir+"/token was already exchanged "+
		"(jti_reused); the file must hold a new token before each refresh, so rotate it well within the minted "+
		"token's lifetime")
}

func TestFederation_AccessToken_ReadsTheTokenFileOnEveryExchange(t *testing.T) {
	clk := stubClock(t)
	upstream := anthropicServe(t)
	cfg := federatedConfig(upstream.url)
	cfg.AnthropicIdentityTokenFile = filepath.Join(t.TempDir(), "token")
	require.NoError(t, os.WriteFile(cfg.AnthropicIdentityTokenFile, []byte("first.jwt.token"), 0o600))
	prov := federatedDoor(t, cfg)

	require.NoError(t, federatedCall(context.Background(), prov))
	require.NoError(t, os.WriteFile(cfg.AnthropicIdentityTokenFile, []byte("second.jwt.token\n"), 0o600))
	clk.Advance(3481 * time.Second)
	require.NoError(t, federatedCall(context.Background(), prov))

	assert.Equal(t, []string{"first.jwt.token", "second.jwt.token"}, upstream.assertions())
	_, auth := upstream.authentication()
	assert.Equal(t, []anthropicAuth{{authorization: "Bearer minted-1"}, {authorization: "Bearer minted-2"}}, auth)
}

func TestFederation_AccessToken_RespectsTheLiteralTokensClaims(t *testing.T) {
	epoch := time.Date(2026, 9, 24, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name           string
		literal        string
		firstErr       string
		steps          []federationStep
		wantAssertions []string
	}{
		{
			name:    "an expired literal token is refused without being sent",
			literal: identityJWT(epoch.Add(-time.Second), ""),
			firstErr: "anthropic federation: ANTHROPIC_IDENTITY_TOKEN expired at 2026-09-24T11:59:59Z and is not sent; " +
				"point ANTHROPIC_IDENTITY_TOKEN_FILE at a token your platform rotates",
		},
		{
			name:           "a padded literal token is sent trimmed",
			literal:        " \tpadded.jwt.token\n",
			wantAssertions: []string{"padded.jwt.token"},
		},
		{
			name:    "a single-use literal token is exchanged once",
			literal: identityJWT(epoch.Add(2*time.Hour), "jti-1"),
			steps: []federationStep{
				{after: 3481 * time.Second, wantBearer: "minted-1", exchanges: 1},
				{after: 90 * time.Second, exchanges: 1, wantErr: "anthropic federation: the access token minted from " +
					"ANTHROPIC_IDENTITY_TOKEN cannot be renewed past 2026-09-24T13:00:00Z: the identity token carries " +
					"a jti claim, so it can be exchanged only once; point ANTHROPIC_IDENTITY_TOKEN_FILE at a token " +
					"your platform rotates"},
			},
		},
		{
			name:    "a reusable literal token is exchanged again at the advisory point",
			literal: identityJWT(epoch.Add(2*time.Hour), ""),
			steps:   []federationStep{{after: 3481 * time.Second, wantBearer: "minted-2", exchanges: 2}},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			clk := stubClock(t)
			upstream := anthropicServe(t)
			cfg := federatedConfig(upstream.url)
			cfg.AnthropicIdentityToken = tc.literal
			prov := federatedDoor(t, cfg)

			if tc.firstErr != "" {
				require.ErrorContains(t, federatedCall(context.Background(), prov), tc.firstErr)
				exchanges, auth := upstream.authentication()
				assert.Empty(t, exchanges)
				assert.Empty(t, auth)
				return
			}

			runFederationSteps(t, clk, upstream, prov, nil, tc.steps)
			if tc.wantAssertions != nil {
				assert.Equal(t, tc.wantAssertions, upstream.assertions())
			}
		})
	}
}

func TestFederation_AccessToken_ExplainsARefusedExchangeWithoutItsTokens(t *testing.T) {
	tests := []struct {
		name      string
		tokenFile bool
		reused    bool
		status    int
		body      string
		wantErr   string // {dir} stands for the directory holding the token file
	}{
		{
			name:      "a file token the endpoint already exchanged",
			tokenFile: true,
			reused:    true,
			wantErr: "anthropic federation: the token in {dir}/token was already exchanged (jti_reused); the file " +
				"must hold a new token before each refresh, so rotate it well within the minted token's lifetime",
		},
		{
			name:   "a literal token the endpoint already exchanged",
			reused: true,
			wantErr: "anthropic federation: ANTHROPIC_IDENTITY_TOKEN was already exchanged (jti_reused) and a " +
				"single-use token is accepted only once; point ANTHROPIC_IDENTITY_TOKEN_FILE at a token your platform rotates",
		},
		{
			name:      "a 400 naming jti_reused is read as a reused token",
			tokenFile: true,
			status:    400,
			body:      `{"error":"invalid_grant","error_description":"jti_reused"}`,
			wantErr: "anthropic federation: the token in {dir}/token was already exchanged (jti_reused); the file " +
				"must hold a new token before each refresh, so rotate it well within the minted token's lifetime",
		},
		{
			name:    "a refusal other than 400 or 401 is not read as a reused token",
			status:  403,
			body:    `{"error":"access_denied","error_description":"jti_reused is not checked here"}`,
			wantErr: "anthropic federation: token exchange refused with status 403: access_denied jti_reused is not checked here",
		},
		{
			name:    "the vendor's error names its message and hides the identity token",
			status:  400,
			body:    `{"type":"error","error":{"type":"invalid_request_error","message":"assertion literal.jwt.token is malformed"}}`,
			wantErr: "anthropic federation: token exchange refused with status 400: assertion [identity token] is malformed",
		},
		{
			name:    "an OAuth error names its code and description",
			status:  401,
			body:    `{"error":"invalid_grant","error_description":"issuer is not trusted"}`,
			wantErr: "anthropic federation: token exchange refused with status 401: invalid_grant issuer is not trusted",
		},
		{
			name:    "a body in neither shape is not quoted",
			status:  502,
			body:    `<html>literal.jwt.token rejected by the proxy</html>`,
			wantErr: "anthropic federation: token exchange refused with status 502: Bad Gateway",
		},
		{
			name:    "an answer without an access token",
			status:  200,
			body:    `{"token_type":"Bearer","expires_in":3600}`,
			wantErr: "anthropic federation: the token exchange answered without an access token",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			upstream := anthropicServe(t)
			dir := t.TempDir()
			cfg := federatedConfig(upstream.url)
			assertion := cfg.AnthropicIdentityToken
			if tc.tokenFile {
				assertion = "projected.jwt.token"
				cfg.AnthropicIdentityTokenFile = filepath.Join(dir, "token")
				require.NoError(t, os.WriteFile(cfg.AnthropicIdentityTokenFile, []byte(assertion), 0o600))
			}
			if tc.reused {
				upstream.markExchanged(assertion)
			}
			upstream.answerExchanges(tc.status, tc.body)

			err := federatedCall(context.Background(), federatedDoor(t, cfg))

			require.Error(t, err)
			want := strings.ReplaceAll(tc.wantErr, "{dir}", dir)
			assert.True(t, strings.HasSuffix(err.Error(), want), "%q does not end with %q", err, want)
			assert.NotContains(t, err.Error(), assertion)
		})
	}
}

func TestFederation_AccessToken_ACallerGivesUpWithoutStoppingTheExchange(t *testing.T) {
	upstream := anthropicServe(t)
	release := upstream.holdExchanges(t)
	prov := federatedDoor(t, federatedConfig(upstream.url))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	first := make(chan error, 1)
	go func() { first <- federatedCall(ctx, prov) }()
	upstream.awaitExchange(t)

	second := make(chan error, 1)
	go func() { second <- federatedCall(context.Background(), prov) }()
	cancel()

	select {
	case err := <-first:
		require.ErrorIs(t, err, context.Canceled)
	case <-time.After(10 * time.Second):
		t.Fatal("a cancelled caller kept waiting for the exchange")
	}

	release()
	select {
	case err := <-second:
		require.NoError(t, err, "the exchange died with the caller that started it")
	case <-time.After(10 * time.Second):
		t.Fatal("the remaining caller never got the exchanged token")
	}

	exchanges, auth := upstream.authentication()
	assert.Len(t, exchanges, 1)
	assert.Equal(t, []anthropicAuth{{authorization: "Bearer minted-1"}}, auth)
}

func TestFederation_AccessToken_BoundsTheExchangeByItsOwnTimeout(t *testing.T) {
	previous := exchangeTimeout
	exchangeTimeout = 50 * time.Millisecond
	t.Cleanup(func() { exchangeTimeout = previous })

	upstream := anthropicServe(t)
	upstream.holdExchanges(t)
	prov := federatedDoor(t, federatedConfig(upstream.url))

	done := make(chan error, 1)
	go func() { done <- federatedCall(context.Background(), prov) }()

	select {
	case err := <-done:
		require.ErrorContains(t, err, "anthropic federation: the token exchange did not answer within 50ms")
	case <-time.After(10 * time.Second):
		t.Fatal("the exchange outlived its own timeout")
	}
}

func TestFederation_FederatedBearer_ReturnsTheTokenTheDoorsShare(t *testing.T) {
	upstream := anthropicServe(t)
	cfg := federatedConfig(upstream.url)

	bearer, err := FederatedBearer(context.Background(), cfg, nil)
	require.NoError(t, err)
	assert.Equal(t, "minted-1", bearer)

	require.NoError(t, federatedCall(context.Background(), federatedDoor(t, cfg)))
	exchanges, auth := upstream.authentication()
	assert.Len(t, exchanges, 1)
	assert.Equal(t, []anthropicAuth{{authorization: "Bearer minted-1"}}, auth)
}

func TestFederation_FederatedBearer_RefusesAnIncompleteConfig(t *testing.T) {
	upstream := anthropicServe(t)
	cfg := federatedConfig(upstream.url)
	cfg.AnthropicServiceAccountID = ""

	_, err := FederatedBearer(context.Background(), cfg, nil)

	require.EqualError(t, err, "anthropic federation: the config is incomplete (missing ANTHROPIC_SERVICE_ACCOUNT_ID)")
	exchanges, _ := upstream.authentication()
	assert.Empty(t, exchanges)
}

func TestFederation_FederatedBearer_ExchangesAtTheVendorWithoutAServerURL(t *testing.T) {
	forgetSources(t, "https://api.anthropic.com/v1")

	var exchangedAt []string
	client := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		exchangedAt = append(exchangedAt, r.Method+" "+r.URL.String())
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body:       io.NopCloser(strings.NewReader(`{"access_token":"minted-1","expires_in":3600}`)),
		}, nil
	})}

	bearer, err := FederatedBearer(context.Background(), federatedConfig(""), client)

	require.NoError(t, err)
	assert.Equal(t, "minted-1", bearer)
	assert.Equal(t, []string{"POST https://api.anthropic.com/v1/oauth/token"}, exchangedAt)
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
