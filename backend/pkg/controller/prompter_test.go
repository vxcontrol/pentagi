package controller

import (
	"context"
	"errors"
	"testing"

	"pentagi/pkg/database"
	"pentagi/pkg/templates"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type prompterFakeQuerier struct {
	database.Querier

	prompts []database.Prompt
	err     error
}

func (f *prompterFakeQuerier) GetUserPrompts(context.Context, int64) ([]database.Prompt, error) {
	return f.prompts, f.err
}

func TestPrompter_NewUserPrompter_OverlaysTheUsersPromptsOnTheDefaults(t *testing.T) {
	dbErr := errors.New("db connection lost")
	defaults := templates.NewDefaultPrompter()

	for _, tc := range []struct {
		name        string
		prompts     []database.Prompt
		dbErr       error
		wantCustom  map[templates.PromptType]string
		wantDefault []templates.PromptType
	}{
		{
			name:        "no user prompts",
			wantDefault: []templates.PromptType{templates.PromptTypePrimaryAgent, templates.PromptTypeAssistant},
		},
		{
			name:        "one override",
			prompts:     []database.Prompt{{Type: database.PromptTypePrimaryAgent, Prompt: "custom primary"}},
			wantCustom:  map[templates.PromptType]string{templates.PromptTypePrimaryAgent: "custom primary"},
			wantDefault: []templates.PromptType{templates.PromptTypeAssistant, templates.PromptTypePentester},
		},
		{
			name: "several overrides",
			prompts: []database.Prompt{
				{Type: database.PromptTypePrimaryAgent, Prompt: "custom primary"},
				{Type: database.PromptTypeCoder, Prompt: "custom coder"},
				{Type: database.PromptTypeReporter, Prompt: "custom reporter"},
			},
			wantCustom: map[templates.PromptType]string{
				templates.PromptTypePrimaryAgent: "custom primary",
				templates.PromptTypeCoder:        "custom coder",
				templates.PromptTypeReporter:     "custom reporter",
			},
			wantDefault: []templates.PromptType{
				templates.PromptTypeAssistant, templates.PromptTypePentester, templates.PromptTypeSearcher,
			},
		},
		{
			name:        "an empty body, which would surface deep in agent rendering",
			prompts:     []database.Prompt{{Type: database.PromptTypePrimaryAgent, Prompt: ""}},
			wantDefault: []templates.PromptType{templates.PromptTypePrimaryAgent},
		},
		{
			name: "an override using a variable the backend no longer provides",
			prompts: []database.Prompt{
				{Type: database.PromptTypeSearcher, Prompt: "Search with {{.GoogleToolName}}"},
				{Type: database.PromptTypeCoder, Prompt: "custom coder"},
			},
			wantCustom:  map[templates.PromptType]string{templates.PromptTypeCoder: "custom coder"},
			wantDefault: []templates.PromptType{templates.PromptTypeSearcher},
		},
		{name: "a database that cannot be read", dbErr: dbErr},
	} {
		t.Run(tc.name, func(t *testing.T) {
			prompter, err := newUserPrompter(context.Background(), &prompterFakeQuerier{prompts: tc.prompts, err: tc.dbErr}, 42)

			if tc.dbErr != nil {
				assert.ErrorIs(t, err, tc.dbErr, "a session is not built on defaults the user did not choose")
				assert.Nil(t, prompter)

				return
			}
			require.NoError(t, err)
			for pt, want := range tc.wantCustom {
				got, err := prompter.GetTemplate(pt)
				require.NoError(t, err)
				assert.Equal(t, want, got, "prompt %s", pt)
			}
			for _, pt := range tc.wantDefault {
				got, err := prompter.GetTemplate(pt)
				require.NoError(t, err)
				want, err := defaults.GetTemplate(pt)
				require.NoError(t, err)
				assert.Equal(t, want, got, "prompt %s keeps its default", pt)
			}
		})
	}
}
