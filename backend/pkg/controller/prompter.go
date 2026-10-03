package controller

import (
	"context"
	"fmt"

	"pentagi/pkg/database"
	"pentagi/pkg/templates"
	"pentagi/pkg/templates/validator"

	"github.com/sirupsen/logrus"
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

// buildUserPrompter overlays each non-empty user override that still validates
// onto defaults and returns a Prompter backed by that map. It mutates defaults:
// callers must pass a fresh map (e.g., from templates.LoadDefaultPromptsMap)
// so the embedded defaults are not modified.
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

		// Saving validates against the variables declared at that time, so a stored
		// override can name one that was removed since; text/template would render
		// the missing key as "<no value>" without an error and point the agent at a
		// tool that does not exist.
		promptType := templates.PromptType(p.Type)
		if err := validator.ValidatePrompt(promptType, p.Prompt); err != nil {
			logrus.WithError(err).WithField("prompt_type", p.Type).
				Warn("custom prompt no longer validates, using the default")
			continue
		}
		defaults[promptType] = p.Prompt
	}

	return templates.NewFlowPrompter(defaults)
}
