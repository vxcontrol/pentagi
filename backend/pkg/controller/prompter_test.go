package controller

import (
	"context"
	"errors"
	"testing"

	"pentagi/pkg/database"
	"pentagi/pkg/templates"
)

// fakeQuerier satisfies database.Querier by embedding the interface
// (so the unused methods stay nil) and overrides only GetUserPrompts.
// Calling any other method would panic, which is exactly what we want
// for a unit test that should not touch unrelated DB code.
type fakeQuerier struct {
	database.Querier
	prompts []database.Prompt
	err     error
}

func (f *fakeQuerier) GetUserPrompts(ctx context.Context, userID int64) ([]database.Prompt, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.prompts, nil
}

// getDefaultTemplate returns the compiled default body for a prompt
// type, used to compare overridden vs. preserved prompts in tests.
func getDefaultTemplate(t *testing.T, pt templates.PromptType) string {
	t.Helper()
	body, err := templates.NewDefaultPrompter().GetTemplate(pt)
	if err != nil {
		t.Fatalf("default prompter has no template for %q: %v", pt, err)
	}
	return body
}

// loadDefaults returns a fresh PromptsMap of the embedded default templates so
// each test mutates its own map without leaking state between cases.
func loadDefaults(t *testing.T) templates.PromptsMap {
	t.Helper()
	defaults, err := templates.LoadDefaultPromptsMap()
	if err != nil {
		t.Fatalf("LoadDefaultPromptsMap error: %v", err)
	}
	return defaults
}

func TestBuildUserPrompter_NoUserPrompts(t *testing.T) {
	prompter := buildUserPrompter(loadDefaults(t), nil)

	for _, pt := range []templates.PromptType{
		templates.PromptTypePrimaryAgent,
		templates.PromptTypeAssistant,
	} {
		got, err := prompter.GetTemplate(pt)
		if err != nil {
			t.Fatalf("GetTemplate(%q) error: %v", pt, err)
		}
		want := getDefaultTemplate(t, pt)
		if got != want {
			t.Errorf("prompt %q: expected default body when user has no overrides, got divergent body", pt)
		}
	}
}

func TestBuildUserPrompter_SingleOverride(t *testing.T) {
	const customBody = "custom primary agent prompt body"
	userPrompts := []database.Prompt{
		{Type: database.PromptTypePrimaryAgent, Prompt: customBody},
	}

	prompter := buildUserPrompter(loadDefaults(t), userPrompts)

	got, err := prompter.GetTemplate(templates.PromptTypePrimaryAgent)
	if err != nil {
		t.Fatalf("GetTemplate(primary_agent) error: %v", err)
	}
	if got != customBody {
		t.Errorf("primary_agent: expected custom body %q, got %q", customBody, got)
	}

	// Spot-check that an unrelated type still resolves to the default.
	for _, pt := range []templates.PromptType{
		templates.PromptTypeAssistant,
		templates.PromptTypePentester,
	} {
		got, err := prompter.GetTemplate(pt)
		if err != nil {
			t.Fatalf("GetTemplate(%q) error: %v", pt, err)
		}
		if got != getDefaultTemplate(t, pt) {
			t.Errorf("prompt %q should still match default after a single unrelated override", pt)
		}
	}
}

func TestBuildUserPrompter_PartialOverridesPreserveDefaults(t *testing.T) {
	overrides := map[database.PromptType]string{
		database.PromptTypePrimaryAgent: "custom primary",
		database.PromptTypeCoder:        "custom coder",
		database.PromptTypeReporter:     "custom reporter",
	}

	userPrompts := make([]database.Prompt, 0, len(overrides))
	for pt, body := range overrides {
		userPrompts = append(userPrompts, database.Prompt{Type: pt, Prompt: body})
	}

	prompter := buildUserPrompter(loadDefaults(t), userPrompts)

	for pt, want := range overrides {
		got, err := prompter.GetTemplate(templates.PromptType(pt))
		if err != nil {
			t.Fatalf("GetTemplate(%q) error: %v", pt, err)
		}
		if got != want {
			t.Errorf("prompt %q: expected custom body %q, got %q", pt, want, got)
		}
	}

	// Types that were not overridden must still match the defaults.
	for _, pt := range []templates.PromptType{
		templates.PromptTypeAssistant,
		templates.PromptTypePentester,
		templates.PromptTypeSearcher,
	} {
		got, err := prompter.GetTemplate(pt)
		if err != nil {
			t.Fatalf("GetTemplate(%q) error: %v", pt, err)
		}
		if got != getDefaultTemplate(t, pt) {
			t.Errorf("prompt %q should still match default after partial overrides", pt)
		}
	}
}

func TestBuildUserPrompter_EmptyBodyIsIgnored(t *testing.T) {
	userPrompts := []database.Prompt{
		{Type: database.PromptTypePrimaryAgent, Prompt: ""},
	}

	prompter := buildUserPrompter(loadDefaults(t), userPrompts)

	got, err := prompter.GetTemplate(templates.PromptTypePrimaryAgent)
	if err != nil {
		t.Fatalf("GetTemplate(primary_agent) error: %v", err)
	}
	if got != getDefaultTemplate(t, templates.PromptTypePrimaryAgent) {
		t.Errorf("primary_agent: empty user body must not clobber default")
	}
}

func TestNewUserPrompter_HappyPath(t *testing.T) {
	const customAssistant = "custom assistant prompt"
	const customCoder = "custom coder prompt"

	db := &fakeQuerier{
		prompts: []database.Prompt{
			{Type: database.PromptTypeAssistant, Prompt: customAssistant},
			{Type: database.PromptTypeCoder, Prompt: customCoder},
		},
	}

	prompter, err := newUserPrompter(context.Background(), db, 42)
	if err != nil {
		t.Fatalf("newUserPrompter error: %v", err)
	}

	for pt, want := range map[templates.PromptType]string{
		templates.PromptTypeAssistant: customAssistant,
		templates.PromptTypeCoder:     customCoder,
	} {
		got, err := prompter.GetTemplate(pt)
		if err != nil {
			t.Fatalf("GetTemplate(%q) error: %v", pt, err)
		}
		if got != want {
			t.Errorf("prompt %q: expected custom body %q, got %q", pt, want, got)
		}
	}

	// Sanity check that an un-customized type still falls back.
	got, err := prompter.GetTemplate(templates.PromptTypePentester)
	if err != nil {
		t.Fatalf("GetTemplate(pentester) error: %v", err)
	}
	if got != getDefaultTemplate(t, templates.PromptTypePentester) {
		t.Errorf("pentester: expected default body when user has no override")
	}
}

func TestNewUserPrompter_DBErrorPropagates(t *testing.T) {
	sentinel := errors.New("db connection lost")
	db := &fakeQuerier{err: sentinel}

	prompter, err := newUserPrompter(context.Background(), db, 42)
	if err == nil {
		t.Fatalf("expected error, got nil prompter=%v", prompter)
	}
	if prompter != nil {
		t.Errorf("expected nil prompter on DB error, got %v", prompter)
	}
	if !errors.Is(err, sentinel) {
		t.Errorf("expected wrapped sentinel error, got %v", err)
	}
}
