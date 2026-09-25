package providers

import (
	"context"
	"fmt"
	"net/http"
	"slices"
	"sort"
	"strings"

	"pentagi/pkg/config"
	"pentagi/pkg/providers/anthropic"
	"pentagi/pkg/providers/pconfig"
	"pentagi/pkg/providers/provider"
)

// Unchecked is not clean: a door whose endpoint refused to answer proves
// nothing about its catalogue.
type RoutingCheck struct {
	Door         string   `json:"door"`
	Endpoint     string   `json:"endpoint"`
	Checked      bool     `json:"checked"`
	Skipped      string   `json:"skipped,omitempty"`
	Prefix       string   `json:"prefix,omitempty"`
	Shipped      int      `json:"shipped"`
	Missing      []string `json:"missing,omitempty"`
	PrefixedOnly []string `json:"prefixed_only,omitempty"`
}

type doorEndpoint struct {
	url    string
	key    string
	prefix string
}

func doorEndpoints(cfg *config.Config) map[string]doorEndpoint {
	return map[string]doorEndpoint{
		"anthropic": {cfg.AnthropicServerURL, cfg.AnthropicAPIKey, ""},
		"bedrock":   {cfg.BedrockServerURL, cfg.BedrockBearerToken, ""},
		"deepseek":  {cfg.DeepSeekServerURL, cfg.DeepSeekAPIKey, cfg.DeepSeekProvider},
		"gemini":    {cfg.GeminiServerURL, cfg.GeminiAPIKey, ""},
		"glm":       {cfg.GLMServerURL, cfg.GLMAPIKey, cfg.GLMProvider},
		"kimi":      {cfg.KimiServerURL, cfg.KimiAPIKey, cfg.KimiProvider},
		"minimax":   {cfg.MiniMaxServerURL, cfg.MiniMaxAPIKey, cfg.MiniMaxProvider},
		"mistral":   {cfg.MistralServerURL, cfg.MistralAPIKey, cfg.MistralProvider},
		"openai":    {cfg.OpenAIServerURL, cfg.OpenAIKey, ""},
		"qwen":      {cfg.QwenServerURL, cfg.QwenAPIKey, cfg.QwenProvider},
		"xai":       {cfg.XAIServerURL, cfg.XAIAPIKey, cfg.XAIProvider},
	}
}

// federatedDoors are the doors configured without a key that authenticate with a
// bearer token minted from the workload's identity.
func federatedDoors(cfg *config.Config) map[string]func(client *http.Client) (string, error) {
	doors := map[string]func(client *http.Client) (string, error){}
	if cfg.AnthropicAPIKey == "" && anthropic.Configured(cfg) {
		doors["anthropic"] = func(client *http.Client) (string, error) {
			ctx, cancel := context.WithTimeout(context.Background(), doorListingTimeout)
			defer cancel()
			return anthropic.FederatedBearer(ctx, cfg, client)
		}
	}
	return doors
}

func ModelPrefix(cfg *config.Config, prvtype provider.ProviderType) string {
	if prvtype == provider.ProviderCustom {
		return cfg.LLMServerProvider
	}
	return doorEndpoints(cfg)[string(prvtype)].prefix
}

func CheckRouting(cfg *config.Config, client *http.Client) []RoutingCheck {
	endpoints := doorEndpoints(cfg)
	federated := federatedDoors(cfg)

	checks := make([]RoutingCheck, 0, len(bundledCatalogs))
	for _, catalog := range bundledCatalogs {
		endpoint, mint := endpoints[catalog.Pkg], federated[catalog.Pkg]
		check := RoutingCheck{Door: catalog.Pkg, Endpoint: endpoint.url, Prefix: endpoint.prefix}

		shipped, err := catalog.Load()
		if err != nil {
			check.Skipped = fmt.Sprintf("catalogue did not load: %v", err)
			checks = append(checks, check)
			continue
		}
		check.Shipped = len(shipped)

		switch {
		case endpoint.url == "":
			check.Skipped = "no server URL configured"
		case endpoint.key == "" && mint == nil:
			check.Skipped = "no key configured"
		default:
			doorClient := client
			if mint != nil {
				token, err := mint(client)
				if err != nil {
					check.Skipped = fmt.Sprintf("no federated token could be minted: %v", err)
					break
				}
				endpoint.key, doorClient = token, withBearer(client, token)
			}
			served, err := servedModels(catalog.Pkg, endpoint, doorClient)
			if err != nil {
				check.Skipped = fmt.Sprintf("the endpoint did not answer: %v", err)
				break
			}
			check.Checked = true
			check.Missing, check.PrefixedOnly = classifyShipped(shipped, served)
		}

		checks = append(checks, check)
	}

	return checks
}

// withBearer sends token in the Authorization header, where a federated token
// belongs, in place of the x-api-key the library sets from its key.
func withBearer(client *http.Client, token string) *http.Client {
	var wrapped http.Client
	if client != nil {
		wrapped = *client
	}
	next := wrapped.Transport
	if next == nil {
		next = http.DefaultTransport
	}
	wrapped.Transport = bearerTransport{next: next, token: token}
	return &wrapped
}

type bearerTransport struct {
	next  http.RoundTripper
	token string
}

func (t bearerTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req = req.Clone(req.Context())
	req.Header.Del("x-api-key")
	req.Header.Set("Authorization", "Bearer "+t.token)
	return t.next.RoundTrip(req)
}

func classifyShipped(shipped, served pconfig.ModelsConfig) (missing, prefixedOnly []string) {
	callable := make(map[string]bool, len(served)*2)
	behindPrefix := make(map[string][]string, len(served))
	for _, model := range served {
		callable[model.Name] = true
		if alias, ok := aliasOfSnapshot(model.Name); ok {
			callable[alias] = true
		}

		cut := strings.LastIndex(model.Name, "/")
		if cut < 0 {
			continue
		}

		bare := model.Name[cut+1:]
		behindPrefix[bare] = append(behindPrefix[bare], model.Name)
		if alias, ok := aliasOfSnapshot(bare); ok {
			behindPrefix[alias] = append(behindPrefix[alias], model.Name)
		}
	}

	for _, model := range shipped {
		switch {
		case callable[model.Name]:
		case len(behindPrefix[model.Name]) > 0:
			routes := slices.Clone(behindPrefix[model.Name])
			slices.Sort(routes)
			prefixedOnly = append(prefixedOnly, model.Name+" -> "+strings.Join(slices.Compact(routes), ", "))
		default:
			missing = append(missing, model.Name)
		}
	}
	sort.Strings(missing)
	sort.Strings(prefixedOnly)

	return missing, prefixedOnly
}

func aliasOfSnapshot(name string) (string, bool) {
	const suffix = len("-20060102")
	if len(name) <= suffix || name[len(name)-suffix] != '-' {
		return "", false
	}
	for _, c := range name[len(name)-suffix+1:] {
		if c < '0' || c > '9' {
			return "", false
		}
	}
	return name[:len(name)-suffix], true
}
