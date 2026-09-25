package providers

import (
	"net/http"
	"net/http/httptest"
	"slices"
	"sync"
	"testing"

	"pentagi/pkg/config"
	"pentagi/pkg/providers/pconfig"
	"pentagi/pkg/providers/provider"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func servedBy(t *testing.T, status int, body string) *httptest.Server {
	t.Helper()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)
	return srv
}

func TestRouting_CheckRouting_ReportsWhatTheDoorDoesNotServe(t *testing.T) {
	for _, tc := range []struct {
		name   string
		status int
		body   string
		cfg    func(url string) *config.Config
		// skipped is what an unchecked door must say about why; empty for a checked one.
		skipped     string
		wantMissing bool
		// served are catalogue names the endpoint serves, which must never come back missing.
		served       []string
		prefixedOnly []string
	}{
		{
			name:   "an endpoint serving two of the catalogue's models",
			status: http.StatusOK, body: `{"data":[{"id":"glm-4.7"},{"id":"zai/glm-4.6"}]}`,
			cfg: func(url string) *config.Config {
				return &config.Config{GLMServerURL: url, GLMAPIKey: "k"}
			},
			wantMissing:  true,
			served:       []string{"glm-4.7", "glm-4.6"},
			prefixedOnly: []string{"glm-4.6 -> zai/glm-4.6"},
		},
		{
			name:   "an endpoint answering unauthorized",
			status: http.StatusUnauthorized, body: `{"error":"nope"}`,
			cfg: func(url string) *config.Config {
				return &config.Config{GLMServerURL: url, GLMAPIKey: "k"}
			},
			skipped: "the endpoint did not answer",
		},
		{
			name:   "an endpoint answering not found",
			status: http.StatusNotFound, body: `not found`,
			cfg: func(url string) *config.Config {
				return &config.Config{GLMServerURL: url, GLMAPIKey: "k"}
			},
			skipped: "the endpoint did not answer",
		},
		{
			name:   "an endpoint with no key to ask it",
			status: http.StatusOK, body: `{"data":[{"id":"glm-4.7"}]}`,
			cfg: func(url string) *config.Config {
				return &config.Config{GLMServerURL: url}
			},
			skipped: "no key configured",
		},
		{
			// GLM_PROVIDER=zai is the prefix the door sends, so zai/<name> is callable, not prefixed-only.
			name:   "models served under the door's own prefix",
			status: http.StatusOK, body: `{"data":[{"id":"zai/glm-4.7"},{"id":"zai/glm-4.6"},{"id":"glm-4.5"}]}`,
			cfg: func(url string) *config.Config {
				return &config.Config{GLMServerURL: url, GLMAPIKey: "k", GLMProvider: "zai"}
			},
			wantMissing: true,
			served:      []string{"glm-4.7", "glm-4.6"},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			srv := servedBy(t, tc.status, tc.body)

			glm := checkFor(t, "glm", CheckRouting(tc.cfg(srv.URL), srv.Client()))

			if tc.skipped != "" {
				require.False(t, glm.Checked, "a door that refused to answer must not count as checked: %+v", glm)
				assert.Contains(t, glm.Skipped, tc.skipped, "an unchecked door must say why")
				assert.Empty(t, glm.Missing, "an unchecked door must report nothing missing")
			} else {
				require.True(t, glm.Checked, "the door answered but was not checked: %+v", glm)
			}
			assert.NotZero(t, glm.Shipped, "the catalogue size was not reported")
			if tc.wantMissing {
				assert.NotEmpty(t, glm.Missing, "the endpoint serves a few of the catalogue's models and nothing was reported missing")
			}
			for _, name := range tc.served {
				assert.NotContains(t, glm.Missing, name, "%q is served, yet reported missing", name)
			}
			assert.Equal(t, tc.prefixedOnly, glm.PrefixedOnly)
		})
	}
}

func TestRouting_CheckRouting_SaysWhyAnUnconfiguredDoorWasSkipped(t *testing.T) {
	checks := CheckRouting(&config.Config{}, http.DefaultClient)

	for _, check := range checks {
		if check.Checked {
			t.Errorf("door %q was checked with no configuration at all", check.Door)
		}
		if check.Skipped != "no server URL configured" {
			t.Errorf("door %q was skipped saying %q, not that it has no server URL", check.Door, check.Skipped)
		}
		if check.Shipped == 0 {
			t.Errorf("door %q reports no catalogue entries at all", check.Door)
		}
	}
}

// anthropicDoorServer answers the token exchange with exchangeStatus and returns
// the headers the model listing arrived with.
func anthropicDoorServer(t *testing.T, exchangeStatus int) (*httptest.Server, func() http.Header) {
	t.Helper()

	var (
		mu      sync.Mutex
		listing = http.Header{}
	)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/oauth/token" {
			w.WriteHeader(exchangeStatus)
			if exchangeStatus == http.StatusOK {
				_, _ = w.Write([]byte(`{"access_token":"minted-access-token","token_type":"Bearer","expires_in":3600}`))
			} else {
				_, _ = w.Write([]byte(`{"type":"error","error":{"type":"authentication_error","message":"invalid assertion"}}`))
			}
			return
		}
		mu.Lock()
		listing = r.Header.Clone()
		mu.Unlock()
		_, _ = w.Write([]byte(`{"data":[{"type":"model","id":"claude-haiku-4-5-20251001"}],"has_more":false}`))
	}))
	t.Cleanup(srv.Close)

	return srv, func() http.Header {
		mu.Lock()
		defer mu.Unlock()
		return listing
	}
}

func TestRouting_CheckRouting_AuthenticatesTheAnthropicDoorByKeyOrFederation(t *testing.T) {
	// Each row names its own rule, so a token minted for one row is never another's.
	federated := func(url, rule string) *config.Config {
		return &config.Config{
			AnthropicServerURL:        url,
			AnthropicFederationRuleID: rule,
			AnthropicOrganizationID:   "00000000-0000-0000-0000-000000000000",
			AnthropicServiceAccountID: "svac_01",
			AnthropicIdentityToken:    "eyJhbGciOiJSUzI1NiJ9.e30.sig",
		}
	}

	for _, tc := range []struct {
		name           string
		exchangeStatus int
		cfg            func(url string) *config.Config
		// skipped is what an unchecked door must say about why; empty for a checked one.
		skipped       string
		apiKey        string
		authorization string
	}{
		{
			name:           "a key is sent as x-api-key even beside a federation config",
			exchangeStatus: http.StatusOK,
			cfg: func(url string) *config.Config {
				cfg := federated(url, "fdrl_beside_key")
				cfg.AnthropicAPIKey = "sk-ant-api03-k"
				return cfg
			},
			apiKey: "sk-ant-api03-k",
		},
		{
			name:           "a federated door is listed with the token its identity was exchanged for",
			exchangeStatus: http.StatusOK,
			cfg:            func(url string) *config.Config { return federated(url, "fdrl_minted") },
			authorization:  "Bearer minted-access-token",
		},
		{
			name:           "a federated door whose identity token is refused says so",
			exchangeStatus: http.StatusUnauthorized,
			cfg:            func(url string) *config.Config { return federated(url, "fdrl_refused") },
			skipped:        "no federated token could be minted: anthropic federation: token exchange refused with status 401",
		},
		{
			name:           "a half-filled federation config is no credential",
			exchangeStatus: http.StatusOK,
			cfg: func(url string) *config.Config {
				cfg := federated(url, "fdrl_half")
				cfg.AnthropicServiceAccountID = ""
				return cfg
			},
			skipped: "no key configured",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			srv, listed := anthropicDoorServer(t, tc.exchangeStatus)

			door := checkFor(t, "anthropic", CheckRouting(tc.cfg(srv.URL), srv.Client()))

			if tc.skipped != "" {
				require.False(t, door.Checked, "the door had no usable credential yet was checked: %+v", door)
				assert.Contains(t, door.Skipped, tc.skipped, "an unchecked door must say why")
			} else {
				require.True(t, door.Checked, "the door answered but was not checked: %+v", door)
			}
			assert.Equal(t, tc.apiKey, listed().Get("x-api-key"), "x-api-key the listing was sent with")
			assert.Equal(t, tc.authorization, listed().Get("Authorization"), "Authorization the listing was sent with")
		})
	}
}

func modelNames(values ...string) pconfig.ModelsConfig {
	models := make(pconfig.ModelsConfig, 0, len(values))
	for _, value := range values {
		models = append(models, pconfig.ModelConfig{Name: value})
	}

	return models
}

func TestRouting_ClassifyShipped_SortsEveryShippedModelIntoCallablePrefixedOrMissing(t *testing.T) {
	tests := []struct {
		name         string
		shipped      pconfig.ModelsConfig
		served       pconfig.ModelsConfig
		missing      []string
		prefixedOnly []string
	}{
		{
			name:    "callable under its own name",
			shipped: modelNames("gpt-5.6"),
			served:  modelNames("gpt-5.6", "gpt-5.5"),
		},
		{
			name:         "listed only behind a vendor prefix",
			shipped:      modelNames("glm-5.3"),
			served:       modelNames("zai/glm-5.3"),
			prefixedOnly: []string{"glm-5.3 -> zai/glm-5.3"},
		},
		{
			name:    "absent in every form",
			shipped: modelNames("glm-4-32b-0414-128k"),
			served:  modelNames("zai/glm-5.3"),
			missing: []string{"glm-4-32b-0414-128k"},
		},
		{
			name:    "dated snapshot answers for the rolling alias",
			shipped: modelNames("gpt-4o"),
			served:  modelNames("gpt-4o-20240806"),
		},
		{
			name:         "every prefix that serves the name is listed, in a stable order",
			shipped:      modelNames("glm-5.3"),
			served:       modelNames("perplexity/perplexity/glm-5.3", "zai/glm-5.3", "hcnsec/glm-5.3"),
			prefixedOnly: []string{"glm-5.3 -> hcnsec/glm-5.3, perplexity/perplexity/glm-5.3, zai/glm-5.3"},
		},
		{
			name:         "dated snapshot behind a prefix is still only prefixed",
			shipped:      modelNames("gpt-4o"),
			served:       modelNames("openai/gpt-4o-20240806"),
			prefixedOnly: []string{"gpt-4o -> openai/gpt-4o-20240806"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			missing, prefixedOnly := classifyShipped(tt.shipped, tt.served)
			if !slices.Equal(missing, tt.missing) {
				t.Errorf("missing = %v, want %v", missing, tt.missing)
			}
			if !slices.Equal(prefixedOnly, tt.prefixedOnly) {
				t.Errorf("prefixedOnly = %v, want %v", prefixedOnly, tt.prefixedOnly)
			}
		})
	}
}

func TestRouting_AliasOfSnapshot_TreatsOnlyADateSuffixAsASnapshot(t *testing.T) {
	for _, tc := range []struct {
		name  string
		alias string
		ok    bool
	}{
		{"claude-haiku-4-5-20251001", "claude-haiku-4-5", true},
		{"gpt-4o-2024-08-06", "gpt-4o-2024-08", false},
		{"claude-opus-4-5-pro-0101", "", false},
		{"20251001", "", false},
	} {
		alias, ok := aliasOfSnapshot(tc.name)
		if ok != tc.ok {
			t.Errorf("aliasOfSnapshot(%q) ok = %v, want %v", tc.name, ok, tc.ok)
			continue
		}
		if ok && alias != tc.alias {
			t.Errorf("aliasOfSnapshot(%q) = %q, want %q", tc.name, alias, tc.alias)
		}
	}
}

func TestRouting_DoorEndpoints_CoversEveryBundledCatalogue(t *testing.T) {
	endpoints := doorEndpoints(&config.Config{})

	for _, catalog := range bundledCatalogs {
		if _, ok := endpoints[catalog.Pkg]; !ok {
			t.Errorf("%s ships a catalogue but has no endpoint, so the routing check reports it unconfigured forever", catalog.Pkg)
		}
	}
}

func TestRouting_ModelPrefix_IsThePrefixTheDoorSendsBeforeAModel(t *testing.T) {
	cfg := &config.Config{GLMProvider: "zai", KimiProvider: "moonshot", LLMServerProvider: "openrouter"}

	for prvtype, want := range map[provider.ProviderType]string{
		provider.ProviderGLM:    "zai",
		provider.ProviderKimi:   "moonshot",
		provider.ProviderCustom: "openrouter",
		provider.ProviderOpenAI: "",
		provider.ProviderOllama: "",
	} {
		if got := ModelPrefix(cfg, prvtype); got != want {
			t.Errorf("ModelPrefix(%s) = %q, want %q", prvtype, got, want)
		}
	}
}
