package providers

import (
	"context"
	"sync"
	"testing"

	"pentagi/pkg/providers/provider"
	"pentagi/pkg/providers/tester/mock"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newSwitchTestFlowProvider builds the minimal flowProvider SetProvider needs:
// the lock it guards its state with, the currently installed provider and the
// tool call ID template resolved for it. Everything else stays nil — the mock
// provider ignores the prompter when resolving a template.
func newSwitchTestFlowProvider(current provider.Provider, tcIDTemplate string) *flowProvider {
	return &flowProvider{
		mx:           &sync.RWMutex{},
		tcIDTemplate: tcIDTemplate,
		Provider:     current,
	}
}

func TestFlowProviderSetProvider_SameNameAndConfigIsNoOp(t *testing.T) {
	current := mock.NewProvider(provider.ProviderQwen, "qwen", "qwen-model")
	incoming := mock.NewProvider(provider.ProviderQwen, "qwen", "qwen-model")

	fp := newSwitchTestFlowProvider(current, "call_{r:24:b}")

	changed, tcIDTemplate, err := fp.SetProvider(context.Background(), incoming)
	require.NoError(t, err)

	assert.False(t, changed, "an identical provider must not trigger a switch")
	assert.Equal(t, "call_{r:24:b}", tcIDTemplate, "the resolved template must be reported unchanged")
	assert.Same(t, current, fp.Provider, "the installed provider instance must be left alone")
}

// The regression this guards: a user provider may be named exactly like a
// built-in one (an intentional feature — it overrides the product default for
// that provider type). Deleting or renaming such an override leaves the name a
// flow refers to valid while the configuration behind it changes, so comparing
// names alone would silently keep the flow on the deleted configuration.
func TestFlowProviderSetProvider_SameNameDifferentConfigSwitches(t *testing.T) {
	override := mock.NewProvider(provider.ProviderOpenAI, "openai", "gpt-x")
	override.SetRawConfig([]byte(`{"base_url":"https://gateway.internal"}`))

	builtin := mock.NewProvider(provider.ProviderOpenAI, "openai", "gpt-x")
	builtin.SetRawConfig([]byte(`{"base_url":"https://api.openai.com"}`))

	fp := newSwitchTestFlowProvider(override, "stale_template")

	changed, tcIDTemplate, err := fp.SetProvider(context.Background(), builtin)
	require.NoError(t, err)

	assert.True(t, changed, "same name but different configuration must switch")
	assert.Equal(t, "toolu_{r:24:b}", tcIDTemplate, "the template must be re-resolved for the new provider")
	assert.Same(t, builtin, fp.Provider)
	assert.Equal(t, "toolu_{r:24:b}", fp.ToolCallIDTemplate(), "the stored template must be refreshed too")
}

func TestFlowProviderSetProvider_DifferentNameSwitches(t *testing.T) {
	current := mock.NewProvider(provider.ProviderQwen, "my-qwen", "qwen-model")
	incoming := mock.NewProvider(provider.ProviderQwen, "qwen", "qwen-model")

	fp := newSwitchTestFlowProvider(current, "stale_template")

	changed, tcIDTemplate, err := fp.SetProvider(context.Background(), incoming)
	require.NoError(t, err)

	assert.True(t, changed)
	assert.Equal(t, "toolu_{r:24:b}", tcIDTemplate)
	assert.Same(t, incoming, fp.Provider)
}

func TestFlowProviderSetProvider_NilProviderIsRejected(t *testing.T) {
	current := mock.NewProvider(provider.ProviderQwen, "qwen", "qwen-model")
	fp := newSwitchTestFlowProvider(current, "call_{r:24:b}")

	changed, _, err := fp.SetProvider(context.Background(), nil)

	require.Error(t, err)
	assert.False(t, changed)
	assert.Same(t, current, fp.Provider, "a rejected switch must leave the flow untouched")
}
