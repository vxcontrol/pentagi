package providers

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"unicode/utf8"

	"pentagi/pkg/templates"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// errRateLimited is shaped like the 429 the Gemini free tier returned while a flow was starting.
// alwaysFails makes fakeCallProvider fail every call a test can make.
const alwaysFails = 1 << 20

var errRateLimited = errors.New("googleapi: Error 429: You exceeded your current quota, RESOURCE_EXHAUSTED")

func taskTitleProvider(prv *fakeCallProvider) *flowProvider {
	fp := newFlowProvider()
	fp.Provider = prv
	fp.prompter = templates.NewDefaultPrompter()
	return fp
}

func TestTaskTitle_GetTaskTitle(t *testing.T) {
	t.Parallel()

	long := strings.Repeat("enumerate the open ports ", 20)
	emoji := strings.Repeat("🔍 сканировать 端口 ", 20)

	tests := []struct {
		name      string
		prv       fakeCallProvider
		input     string
		want      string
		wantCalls int
		why       string
	}{
		{
			name:      "a title the model returns is used as is",
			prv:       fakeCallProvider{result: "Scan Local Ports"},
			input:     "scan the local ports",
			want:      "Scan Local Ports",
			wantCalls: 1,
			why:       "the success path must not change",
		},
		{
			name:      "a single rate limit falls back at once",
			prv:       fakeCallProvider{failTimes: 1, err: errRateLimited, result: "Scan Local Ports"},
			input:     "scan the local ports",
			want:      "scan the local ports",
			wantCalls: 1,
			why:       "one 429 on a cosmetic call must neither fail nor delay the task",
		},
		{
			name:      "a model that keeps failing falls back to the input",
			prv:       fakeCallProvider{failTimes: alwaysFails, err: errRateLimited},
			input:     "scan the local ports",
			want:      "scan the local ports",
			wantCalls: 1,
			why:       "the task must start even when no title can be generated",
		},
		{
			name:      "an empty answer falls back to the input",
			prv:       fakeCallProvider{result: " \n\t "},
			input:     "scan the local ports",
			want:      "scan the local ports",
			wantCalls: 1,
			why:       "a blank title is useless in the UI",
		},
		{
			name:      "the fallback keeps the first line and collapses whitespace",
			prv:       fakeCallProvider{failTimes: alwaysFails, err: errRateLimited},
			input:     "\n\n  scan\tthe   local\tports  \nthen report back\n",
			want:      "scan the local ports",
			wantCalls: 1,
		},
		{
			name:      "an empty input falls back to a placeholder",
			prv:       fakeCallProvider{failTimes: alwaysFails, err: errRateLimited},
			input:     " \n\t ",
			want:      untitledTask,
			wantCalls: 1,
		},
		{
			name:      "a long input is cut at the title limit",
			prv:       fakeCallProvider{failTimes: alwaysFails, err: errRateLimited},
			input:     long,
			want:      string([]rune(strings.TrimSpace(long))[:taskTitleLimit-1]) + "…",
			wantCalls: 1,
		},
		{
			name:      "a multibyte input is cut on a rune boundary",
			prv:       fakeCallProvider{failTimes: alwaysFails, err: errRateLimited},
			input:     emoji,
			want:      string([]rune(strings.TrimSpace(emoji))[:taskTitleLimit-1]) + "…",
			wantCalls: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			prv := tt.prv
			got, err := taskTitleProvider(&prv).GetTaskTitle(t.Context(), tt.input)

			require.NoError(t, err, tt.why)
			assert.Equal(t, tt.want, got, tt.why)
			assert.True(t, utf8.ValidString(got), "the title must stay valid UTF-8")
			assert.LessOrEqual(t, utf8.RuneCountInString(got), taskTitleLimit)
			assert.Equal(t, tt.wantCalls, prv.callCount, tt.why)
		})
	}
}

func TestTaskTitle_GetTaskTitle_CanceledContextIsNotHiddenByTheFallback(t *testing.T) {
	t.Parallel()

	prv := &fakeCallProvider{failTimes: 1, err: fmt.Errorf("request aborted: %w", context.Canceled)}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	_, err := taskTitleProvider(prv).GetTaskTitle(ctx, "scan the local ports")

	require.ErrorIs(t, err, context.Canceled, "a stopped flow must not get a task with a fallback title")
	assert.Equal(t, 1, prv.callCount)
}
