package controller

import (
	"context"
	"fmt"

	"pentagi/pkg/database"
	"pentagi/pkg/templates"
)

// newUserPrompter loads the user's custom prompts from the database and
// overlays them onto the compiled default templates. Prompt types that
// the user has not customized continue to use the defaults. A database
// error is returned to the caller so that session creation fails
// explicitly instead of silently falling back to defaults.
func newUserPrompter(ctx context.Context, db database.Querier, userID int64) (templates.Prompter, error) {
	userPrompts, err := db.GetUserPrompts(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to load user prompts: %w", err)
	}

	defaults, err := templates.LoadDefaultPromptsMap()
	if err != nil {
		return nil, fmt.Errorf("failed to load default templates: %w", err)
	}

	return buildUserPrompter(defaults, userPrompts), nil
}

// buildUserPrompter is the pure merge step extracted from newUserPrompter so
// it can be unit-tested without a database fake or filesystem access. It
// mutates the supplied defaults map by overlaying each non-empty user
// override on top, then returns a Prompter backed by that map. Callers must
// pass a fresh map (e.g., from templates.LoadDefaultPromptsMap) so the
// embedded defaults are not modified.
func buildUserPrompter(defaults templates.PromptsMap, userPrompts []database.Prompt) templates.Prompter {
	for _, p := range userPrompts {
		if p.Prompt == "" {
			// The Prompts UI uses delete (or reset, which writes the
			// default body back) to remove a customization, so an empty
			// body is unexpected. Skip it instead of clobbering the
			// default with an empty string that would later surface as
			// ErrTemplateNotFound deep inside agent rendering.
			continue
		}
		defaults[templates.PromptType(p.Type)] = p.Prompt
	}

	return templates.NewFlowPrompter(defaults)
}
