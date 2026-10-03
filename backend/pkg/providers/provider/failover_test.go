package provider

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"pentagi/pkg/providers/pconfig"
	"pentagi/pkg/templates"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vxcontrol/langchaingo/llms"
	"github.com/vxcontrol/langchaingo/llms/streaming"
)

var errFailoverVendorDown = errors.New("vendor is down")

// fakeProvider embeds a nil Provider so an unstubbed method panics instead of answering zero values.
type fakeProvider struct {
	Provider

	name  ProviderName
	err   error
	calls int
}

func (f *fakeProvider) Name() ProviderName { return f.name }

func (f *fakeProvider) answer(input string) (*llms.ContentResponse, error) {
	f.calls++
	if f.err != nil {
		return nil, f.err
	}
	return &llms.ContentResponse{Choices: []*llms.ContentChoice{{Content: string(f.name) + ":" + input}}}, nil
}

func (f *fakeProvider) chainAnswer(chain []llms.MessageContent) (*llms.ContentResponse, error) {
	return f.answer(chain[0].Parts[0].(llms.TextContent).Text)
}

func (f *fakeProvider) Call(_ context.Context, _ pconfig.ProviderOptionsType, prompt string) (string, error) {
	resp, err := f.answer(prompt)
	if err != nil {
		return "", err
	}
	return resp.Choices[0].Content, nil
}

func (f *fakeProvider) CallEx(
	_ context.Context, _ pconfig.ProviderOptionsType, chain []llms.MessageContent, _ streaming.Callback,
) (*llms.ContentResponse, error) {
	return f.chainAnswer(chain)
}

func (f *fakeProvider) CallWithTools(
	_ context.Context, _ pconfig.ProviderOptionsType, chain []llms.MessageContent, _ []llms.Tool, _ streaming.Callback,
) (*llms.ContentResponse, error) {
	return f.chainAnswer(chain)
}

func (f *fakeProvider) CallWithExtraOptions(
	_ context.Context, _ pconfig.ProviderOptionsType, chain []llms.MessageContent, _ []llms.Tool,
	_ streaming.Callback, _ ...llms.CallOption,
) (*llms.ContentResponse, error) {
	return f.chainAnswer(chain)
}

// failoverPrompter marks the prompter a path hands in.
type failoverPrompter struct{ templates.Prompter }

func (f *fakeProvider) GetToolCallIDTemplate(ctx context.Context, prompter templates.Prompter) (string, error) {
	input := "no prompter"
	if _, isMarked := prompter.(failoverPrompter); isMarked {
		input = "hi"
	}

	return f.Call(ctx, pconfig.OptionsTypeSimple, input)
}

func failoverContent(resp *llms.ContentResponse, err error) (string, error) {
	if err != nil {
		return "", err
	}
	return resp.Choices[0].Content, nil
}

// failoverPaths drives every method that reaches the model with the same input.
var failoverPaths = []struct {
	name string
	call func(ctx context.Context, prv Provider) (string, error)
}{
	{"a single prompt", func(ctx context.Context, prv Provider) (string, error) {
		return prv.Call(ctx, pconfig.OptionsTypeSimple, "hi")
	}},
	{"a chain", func(ctx context.Context, prv Provider) (string, error) {
		return failoverContent(prv.CallEx(ctx, pconfig.OptionsTypeSimple,
			[]llms.MessageContent{llms.TextParts(llms.ChatMessageTypeHuman, "hi")}, nil))
	}},
	{"a chain with tools", func(ctx context.Context, prv Provider) (string, error) {
		return failoverContent(prv.CallWithTools(ctx, pconfig.OptionsTypeSimple,
			[]llms.MessageContent{llms.TextParts(llms.ChatMessageTypeHuman, "hi")}, nil, nil))
	}},
	{"a chain with extra options", func(ctx context.Context, prv Provider) (string, error) {
		return failoverContent(prv.CallWithExtraOptions(ctx, pconfig.OptionsTypeSimple,
			[]llms.MessageContent{llms.TextParts(llms.ChatMessageTypeHuman, "hi")}, nil, nil))
	}},
	{"the tool call id template", func(ctx context.Context, prv Provider) (string, error) {
		return prv.GetToolCallIDTemplate(ctx, failoverPrompter{})
	}},
}

// TestFailover_SwitchesToTheBackupOnlyWhenThePrimaryFails runs every case through each calling method.
func TestFailover_SwitchesToTheBackupOnlyWhenThePrimaryFails(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		primaryErr   error
		done         bool
		calls        int
		wantAnswer   string
		wantErr      error
		wantPrimary  int
		wantBackup   int
		wantSwitches []string
	}{
		{
			name:       "a failing primary hands every call to the backup, starting at the primary each time",
			primaryErr: errFailoverVendorDown, calls: 3, wantAnswer: "backup:hi", wantPrimary: 3, wantBackup: 3,
			wantSwitches: []string{
				"primary -> backup: vendor is down",
				"primary -> backup: vendor is down",
				"primary -> backup: vendor is down",
			},
		},
		{name: "a healthy primary keeps the call", calls: 1, wantAnswer: "primary:hi", wantPrimary: 1},
		{name: "a cancelled call does not switch", primaryErr: context.Canceled, calls: 1, wantErr: context.Canceled, wantPrimary: 1},
		{
			name: "a call whose deadline ran out does not switch", primaryErr: context.DeadlineExceeded, calls: 1,
			wantErr: context.DeadlineExceeded, wantPrimary: 1,
		},
		{
			name: "a failure on a context that is already done does not switch", primaryErr: errFailoverVendorDown,
			done: true, calls: 1, wantErr: errFailoverVendorDown, wantPrimary: 1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			for _, path := range failoverPaths {
				t.Run(path.name, func(t *testing.T) {
					primary := &fakeProvider{name: "primary", err: tc.primaryErr}
					backup := &fakeProvider{name: "backup"}

					var switches []string
					prv := WithFailover(primary, backup, func(_ context.Context, from, to ProviderName, err error) {
						switches = append(switches, fmt.Sprintf("%s -> %s: %v", from, to, err))
					})

					ctx, cancel := context.WithCancel(context.Background())
					defer cancel()
					if tc.done {
						cancel()
					}

					var (
						answer string
						err    error
					)
					for range tc.calls {
						answer, err = path.call(ctx, prv)
					}

					if tc.wantErr != nil {
						require.ErrorIs(t, err, tc.wantErr)
					} else {
						require.NoError(t, err)
						assert.Equal(t, tc.wantAnswer, answer)
					}
					assert.Equal(t, tc.wantPrimary, primary.calls)
					assert.Equal(t, tc.wantBackup, backup.calls)
					assert.Equal(t, tc.wantSwitches, switches)
				})
			}
		})
	}
}

func TestFailover_WithFailover_ReturnsThePrimaryWithoutAUsableBackup(t *testing.T) {
	t.Parallel()

	primary := &fakeProvider{name: "primary"}

	assert.Same(t, Provider(primary), WithFailover(primary, nil, nil),
		"no backup means no wrapper")
	assert.Same(t, Provider(primary), WithFailover(primary, &fakeProvider{name: "primary"}, nil),
		"a backup equal to the primary is not a backup")
}
