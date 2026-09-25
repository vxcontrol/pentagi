package mistral

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"pentagi/pkg/config"
	"pentagi/pkg/providers/pconfig"
	"pentagi/pkg/providers/provider"

	"github.com/vxcontrol/langchaingo/llms"
)

func mistralWireBody(t *testing.T, opt pconfig.ProviderOptionsType) map[string]any {
	t.Helper()

	var body map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &body)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"x","object":"chat.completion","choices":[{"index":0,` +
			`"message":{"role":"assistant","content":"ok"},"finish_reason":"stop"}]}`))
	}))
	defer srv.Close()

	cfg := &config.Config{MistralAPIKey: "k", MistralServerURL: srv.URL, MistralProvider: "mistral"}
	pc, err := DefaultProviderConfig()
	if err != nil {
		t.Fatalf("config: %v", err)
	}
	prov, err := New(cfg, provider.DefaultProviderNameMistral, pc)
	if err != nil {
		t.Fatalf("provider: %v", err)
	}
	if _, err = prov.CallEx(context.Background(), opt,
		[]llms.MessageContent{llms.TextParts(llms.ChatMessageTypeHuman, "hi")}, nil); err != nil {
		t.Fatalf("call: %v", err)
	}
	if body == nil {
		t.Fatal("no request reached the wire")
	}
	return body
}

func TestMistral_DefaultProviderConfig_SendsEffortHighOnlyFromReasoningAgents(t *testing.T) {
	for _, tc := range []struct {
		effort string // empty: reasoning_effort must stay off the wire
		agents []pconfig.ProviderOptionsType
	}{
		{effort: "high", agents: []pconfig.ProviderOptionsType{
			pconfig.OptionsTypeGenerator, pconfig.OptionsTypeRefiner, pconfig.OptionsTypeAdviser,
			pconfig.OptionsTypeInstaller,
		}},
		{agents: []pconfig.ProviderOptionsType{
			pconfig.OptionsTypeSimple, pconfig.OptionsTypeSimpleJSON,
			pconfig.OptionsTypePrimaryAgent, pconfig.OptionsTypeAssistant,
			pconfig.OptionsTypeReflector, pconfig.OptionsTypeSearcher, pconfig.OptionsTypeEnricher,
			pconfig.OptionsTypeCoder, pconfig.OptionsTypePentester,
		}},
	} {
		for _, opt := range tc.agents {
			t.Run("the shipped "+string(opt)+" agent", func(t *testing.T) {
				body := mistralWireBody(t, opt)
				got, sent := body["reasoning_effort"]
				if tc.effort == "" {
					if sent {
						t.Errorf("reasoning_effort reached the wire (%v) for a utility agent", got)
					}
					return
				}
				if got != tc.effort {
					t.Errorf("reasoning_effort = %v, want %q; body=%v", got, tc.effort, body)
				}
			})
		}
	}
}
