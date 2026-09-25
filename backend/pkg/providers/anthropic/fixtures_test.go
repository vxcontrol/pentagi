package anthropic

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"pentagi/pkg/config"
)

// anthropicUpstream answers the token exchange and every other request with a short text turn, and keeps
// what each request carried. Like the vendor, it exchanges a single-use identity token only once: a JWT
// carrying jti, and here any token that is not a JWT.
type anthropicUpstream struct {
	url     string
	arrived chan struct{}

	mu        sync.Mutex
	bodies    []map[string]any
	auth      []anthropicAuth
	exchanges []map[string]any
	exchanged map[string]bool
	minted    int
	expiresIn string          // raw JSON; empty leaves the field out
	answer    *exchangeAnswer // every exchange gets this instead of a token
	hold      chan struct{}   // every exchange waits for it to close
}

type anthropicAuth struct{ authorization, apiKey string }

type exchangeAnswer struct {
	status int
	body   string
}

func anthropicServe(t *testing.T) *anthropicUpstream {
	t.Helper()

	upstream := &anthropicUpstream{arrived: make(chan struct{}, 64), exchanged: map[string]bool{}, expiresIn: "3600"}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if strings.HasSuffix(r.URL.Path, "/oauth/token") {
			upstream.exchange(w, r, body)
			return
		}

		upstream.mu.Lock()
		defer upstream.mu.Unlock()

		upstream.bodies = append(upstream.bodies, body)
		upstream.auth = append(upstream.auth,
			anthropicAuth{authorization: r.Header.Get("Authorization"), apiKey: r.Header.Get("x-api-key")})
		fmt.Fprint(w, `{"id":"msg_1","type":"message","role":"assistant","model":"claude-sonnet-5",`+
			`"content":[{"type":"text","text":"ok"}],"stop_reason":"end_turn",`+
			`"usage":{"input_tokens":1,"output_tokens":1}}`)
	}))
	t.Cleanup(srv.Close)
	forgetSources(t, srv.URL)
	upstream.url = srv.URL

	return upstream
}

func (u *anthropicUpstream) exchange(w http.ResponseWriter, r *http.Request, body map[string]any) {
	u.mu.Lock()
	u.exchanges = append(u.exchanges, body)
	hold := u.hold
	u.mu.Unlock()

	select {
	case u.arrived <- struct{}{}:
	default:
	}
	if hold != nil {
		select {
		case <-hold:
		case <-r.Context().Done():
			return
		}
	}

	u.mu.Lock()
	defer u.mu.Unlock()

	assertion, _ := body["assertion"].(string)
	switch {
	case u.answer != nil:
		w.WriteHeader(u.answer.status)
		fmt.Fprint(w, u.answer.body)
	case u.exchanged[assertion] && singleUse(assertion):
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(w, `{"error":"invalid_grant","error_description":"jti_reused"}`)
	default:
		u.exchanged[assertion] = true
		u.minted++
		answer := fmt.Sprintf(`{"access_token":"minted-%d","token_type":"Bearer","scope":"user:inference"`, u.minted)
		if u.expiresIn != "" {
			answer += `,"expires_in":` + u.expiresIn
		}
		fmt.Fprint(w, answer+"}")
	}
}

func singleUse(assertion string) bool {
	parts := strings.Split(assertion, ".")
	if len(parts) != 3 {
		return true
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	var claims map[string]any
	if err != nil || json.Unmarshal(payload, &claims) != nil {
		return true
	}
	_, jti := claims["jti"]
	return jti
}

func (u *anthropicUpstream) take() []map[string]any {
	u.mu.Lock()
	defer u.mu.Unlock()

	bodies := u.bodies
	u.bodies = nil
	return bodies
}

func (u *anthropicUpstream) authentication() ([]map[string]any, []anthropicAuth) {
	u.mu.Lock()
	defer u.mu.Unlock()

	return u.exchanges, u.auth
}

func (u *anthropicUpstream) assertions() []string {
	u.mu.Lock()
	defer u.mu.Unlock()

	var assertions []string
	for _, exchange := range u.exchanges {
		assertion, _ := exchange["assertion"].(string)
		assertions = append(assertions, assertion)
	}
	return assertions
}

func (u *anthropicUpstream) mintExpiringIn(expiresIn string) {
	u.mu.Lock()
	defer u.mu.Unlock()

	u.expiresIn = expiresIn
}

// answerExchanges makes every exchange get status and body; a zero status mints tokens again.
func (u *anthropicUpstream) answerExchanges(status int, body string) {
	u.mu.Lock()
	defer u.mu.Unlock()

	u.answer = nil
	if status != 0 {
		u.answer = &exchangeAnswer{status: status, body: body}
	}
}

func (u *anthropicUpstream) markExchanged(assertion string) {
	u.mu.Lock()
	defer u.mu.Unlock()

	u.exchanged[assertion] = true
}

// holdExchanges keeps every exchange waiting until the returned release is called.
func (u *anthropicUpstream) holdExchanges(t *testing.T) (release func()) {
	t.Helper()

	hold := make(chan struct{})
	u.mu.Lock()
	u.hold = hold
	u.mu.Unlock()

	var once sync.Once
	release = func() { once.Do(func() { close(hold) }) }
	t.Cleanup(release)
	return release
}

func (u *anthropicUpstream) awaitExchange(t *testing.T) {
	t.Helper()

	select {
	case <-u.arrived:
	case <-time.After(10 * time.Second):
		t.Fatal("no token exchange reached the upstream within 10 s")
	}
}

// forgetSources drops the token sources of baseURL when the test ends, so a later test that is handed
// the same port, or a rerun under -count, starts without a cached token.
func forgetSources(t *testing.T, baseURL string) {
	t.Helper()

	t.Cleanup(func() {
		sourcesMu.Lock()
		defer sourcesMu.Unlock()

		for id := range sources {
			if id.baseURL == baseURL {
				delete(sources, id)
			}
		}
	})
}

func federatedConfig(serverURL string) *config.Config {
	return &config.Config{
		AnthropicServerURL:        serverURL,
		AnthropicFederationRuleID: "fdrl_1",
		AnthropicOrganizationID:   "00000000-0000-0000-0000-000000000000",
		AnthropicServiceAccountID: "svac_1",
		AnthropicWorkspaceID:      "wrkspc_1",
		AnthropicIdentityToken:    "literal.jwt.token",
	}
}

type fakeClock struct {
	mu  sync.Mutex
	now time.Time
}

func (c *fakeClock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.now
}

func (c *fakeClock) Advance(d time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.now = c.now.Add(d)
}

// stubClock starts at 2026-09-24 12:00 UTC and is read by CredentialsNotice and by every token source
// created after it.
func stubClock(t *testing.T) *fakeClock {
	t.Helper()

	c := &fakeClock{now: time.Date(2026, 9, 24, 12, 0, 0, 0, time.UTC)}
	previous := clock
	clock = c.Now
	t.Cleanup(func() { clock = previous })
	return c
}

// identityJWT is an unsigned identity token with an exp claim, and a jti claim unless jti is empty.
func identityJWT(exp time.Time, jti string) string {
	claims := map[string]any{"sub": "system:serviceaccount:pentagi:agent", "exp": exp.Unix()}
	if jti != "" {
		claims["jti"] = jti
	}
	payload, _ := json.Marshal(claims)
	return "eyJhbGciOiJSUzI1NiJ9." + base64.RawURLEncoding.EncodeToString(payload) + ".c2lnbmF0dXJl"
}
