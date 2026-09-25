// Test doubles and helpers shared by more than one test file; a single file's helper stays in that file.

package providers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	"pentagi/pkg/cast"
	"pentagi/pkg/config"
	"pentagi/pkg/providers/pconfig"
	"pentagi/pkg/providers/provider"

	"github.com/vxcontrol/langchaingo/llms"
)

func newFlowProvider() *flowProvider {
	return &flowProvider{
		mx:              &sync.RWMutex{},
		cfg:             &config.Config{DataDir: "testdata"},
		summarizerCache: newSummarizerCache(),
		callCounter:     &atomic.Int64{},
		maxGACallsLimit: maxGeneralAgentChainIterations,
		maxLACallsLimit: maxLimitedAgentChainIterations,
		buildMonitor: func() *executionMonitor {
			return &executionMonitor{
				enabled: false,
			}
		},
	}
}

// fakeCallProvider stands in for the provider callWithSetupRetries calls; only Call is implemented.
type fakeCallProvider struct {
	provider.Provider
	callCount int
	failTimes int
	err       error
	result    string
}

func (f *fakeCallProvider) Call(ctx context.Context, opt pconfig.ProviderOptionsType, prompt string) (string, error) {
	f.callCount++
	if f.callCount <= f.failTimes {
		return "", f.err
	}
	return f.result, nil
}

func checkFor(t *testing.T, door string, checks []RoutingCheck) RoutingCheck {
	t.Helper()

	for _, check := range checks {
		if check.Door == door {
			return check
		}
	}
	t.Fatalf("no check for door %q", door)
	return RoutingCheck{}
}

// customDoor builds door against a gateway serving model, with block as the simple section of the empty config.
func customDoor(
	t *testing.T, model string, block map[string]any, door registryEntry,
) (provider.Provider, *map[string]any) {
	t.Helper()

	srv, body := gatewayServing(t, model)
	cfg := &config.Config{LLMServerKey: "k", LLMServerURL: srv.URL, LLMServerModel: model}

	var asMap map[string]any
	if err := json.Unmarshal([]byte(pconfig.EmptyProviderConfigRaw), &asMap); err != nil {
		t.Fatalf("template: %v", err)
	}
	asMap[string(pconfig.OptionsTypeSimple)] = block
	patched, _ := json.Marshal(asMap)

	providerConfig, err := door.BuildConfig(cfg, patched)
	if err != nil {
		t.Fatalf("config: %v", err)
	}

	prov, err := door.New(cfg, provider.DefaultProviderNameCustom, providerConfig)
	if err != nil {
		t.Fatalf("provider: %v", err)
	}

	return prov, body
}

func gatewayServing(t *testing.T, name string) (*httptest.Server, *map[string]any) {
	t.Helper()

	var body map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/models" {
			fmt.Fprintf(w, `{"data":[{"id":%q,"supported_parameters":["tools","reasoning"]}]}`, name)
			return
		}
		raw, _ := io.ReadAll(r.Body)
		body = nil
		_ = json.Unmarshal(raw, &body)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"id":"x","object":"chat.completion","choices":[{"index":0,`+
			`"message":{"role":"assistant","content":"ok"},"finish_reason":"stop"}]}`)
	}))
	t.Cleanup(srv.Close)

	return srv, &body
}

func windowCompactionChain(turns, bytesPerTurn int) []llms.MessageContent {
	chain := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeSystem, "you are a tester"),
	}
	for range turns {
		chain = append(chain,
			llms.TextParts(llms.ChatMessageTypeHuman, strings.Repeat("q", bytesPerTurn)),
			llms.TextParts(llms.ChatMessageTypeAI, strings.Repeat("a", bytesPerTurn)),
		)
	}

	return chain
}

func windowCompactionBytes(chain []llms.MessageContent) int {
	size := 0
	for idx := range chain {
		size += cast.CalculateMessageSize(&chain[idx])
	}

	return size
}

// windowCompactionTools is one tool whose schema marshals to description+100 bytes.
func windowCompactionTools(description int) []llms.Tool {
	return []llms.Tool{{
		Type: "function",
		Function: &llms.FunctionDefinition{
			Name:        "terminal",
			Description: strings.Repeat("d", description),
			Parameters:  map[string]any{"type": "object"},
		},
	}}
}
