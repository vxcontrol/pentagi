package providers

import (
	"context"
	"errors"
	"sync"
	"testing"

	"pentagi/pkg/providers/provider"
	"pentagi/pkg/providers/tester/mock"
	"pentagi/pkg/templates"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// providerTemplateStub resolves the tool call ID template to template, or fails with err.
type providerTemplateStub struct {
	*mock.Provider

	template string
	err      error
}

func (p *providerTemplateStub) GetToolCallIDTemplate(context.Context, templates.Prompter) (string, error) {
	return p.template, p.err
}

func TestProvider_SetProvider_SwitchesOnlyWhenTheNameOrConfigurationChanges(t *testing.T) {
	withConfig := func(prv *mock.Provider, raw string) *mock.Provider {
		prv.SetRawConfig([]byte(raw))
		return prv
	}

	tests := []struct {
		name         string
		current      provider.Provider
		incoming     provider.Provider
		wantSwitch   bool
		wantTemplate string
		wantErr      string
	}{
		{
			name:         "the same name and configuration keeps the installed provider",
			current:      mock.NewProvider(provider.ProviderQwen, "qwen", "qwen-model"),
			incoming:     mock.NewProvider(provider.ProviderQwen, "qwen", "qwen-model"),
			wantTemplate: "call_{r:24:b}",
		},
		{
			name:         "the same name with another configuration switches",
			current:      withConfig(mock.NewProvider(provider.ProviderOpenAI, "openai", "gpt-x"), `{"base_url":"https://gateway.internal"}`),
			incoming:     withConfig(mock.NewProvider(provider.ProviderOpenAI, "openai", "gpt-x"), `{"base_url":"https://api.openai.com"}`),
			wantSwitch:   true,
			wantTemplate: "toolu_{r:24:b}",
		},
		{
			name:         "another name switches",
			current:      mock.NewProvider(provider.ProviderQwen, "my-qwen", "qwen-model"),
			incoming:     mock.NewProvider(provider.ProviderQwen, "qwen", "qwen-model"),
			wantSwitch:   true,
			wantTemplate: "toolu_{r:24:b}",
		},
		{
			name:         "a provider that resolves no template gets the default one",
			current:      mock.NewProvider(provider.ProviderQwen, "my-qwen", "qwen-model"),
			incoming:     &providerTemplateStub{Provider: mock.NewProvider(provider.ProviderQwen, "qwen", "qwen-model")},
			wantSwitch:   true,
			wantTemplate: "call_{r:24:x}",
		},
		{
			name:    "a template that cannot be resolved is refused",
			current: mock.NewProvider(provider.ProviderQwen, "my-qwen", "qwen-model"),
			incoming: &providerTemplateStub{
				Provider: mock.NewProvider(provider.ProviderQwen, "qwen", "qwen-model"),
				err:      errors.New("model unreachable"),
			},
			wantErr: "failed to get tool call ID template: model unreachable",
		},
		{
			name:    "no provider is refused",
			current: mock.NewProvider(provider.ProviderQwen, "qwen", "qwen-model"),
			wantErr: "new provider is nil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fp := &flowProvider{mx: &sync.RWMutex{}, tcIDTemplate: "call_{r:24:b}", Provider: tt.current}

			changed, template, err := fp.SetProvider(context.Background(), tt.incoming)

			installed, stored := tt.current, "call_{r:24:b}"
			if tt.wantSwitch {
				installed, stored = tt.incoming, tt.wantTemplate
			}
			if tt.wantErr != "" {
				assert.EqualError(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
			}
			assert.Equal(t, tt.wantSwitch, changed)
			assert.Equal(t, tt.wantTemplate, template, "the reported template")
			assert.Same(t, installed, fp.Provider, "the installed provider")
			assert.Equal(t, stored, fp.ToolCallIDTemplate(), "the stored template")
		})
	}
}
