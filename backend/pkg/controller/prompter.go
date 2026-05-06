package controller

import (
	"context"
	"encoding/json"
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
	return buildUserPrompter(templates.NewDefaultPrompter(), userPrompts)
}

// buildUserPrompter is the pure merge step extracted from newUserPrompter
// so it can be unit-tested without a database fake. It seeds the result
// with every default template and then overrides each prompt type the
// user has saved a non-empty body for.
func buildUserPrompter(defaults templates.Prompter, userPrompts []database.Prompt) (templates.Prompter, error) {
	blob, err := defaults.DumpTemplates()
	if err != nil {
		return nil, fmt.Errorf("failed to dump default templates: %w", err)
	}

	merged := templates.PromptsMap{}
	if err := json.Unmarshal(blob, &merged); err != nil {
		return nil, fmt.Errorf("failed to parse default templates: %w", err)
	}

	for _, p := range userPrompts {
		if p.Prompt == "" {
			// The Prompts UI uses delete (or reset, which writes the
			// default body back) to remove a customization, so an empty
			// body is unexpected. Skip it instead of clobbering the
			// default with an empty string that would later surface as
			// ErrTemplateNotFound deep inside agent rendering.
			continue
		}
		merged[templates.PromptType(p.Type)] = p.Prompt
	}

	return templates.NewFlowPrompter(merged), nil
}
