package templates_test

import (
	"strings"
	"testing"

	"pentagi/pkg/templates"
	"pentagi/pkg/templates/validator"
)

// sageTemplateData returns template variables populated for SAGE tests.
// It starts from the full dummy data set (which satisfies every template)
// and overrides the SAGE-specific flags as requested.
func sageTemplateData(sageEnabled bool) map[string]any {
	data := validator.CreateDummyTemplateData()
	data["SAGEEnabled"] = sageEnabled
	if sageEnabled {
		data["SageRecallToolName"] = "sage_recall"
		data["SageRememberToolName"] = "sage_remember"
	}
	return data
}

// renderTemplate renders a prompt template with the given data using the
// default prompter (which reads from the embedded prompts/*.tmpl files).
func renderTemplate(t *testing.T, promptType templates.PromptType, data map[string]any) string {
	t.Helper()
	prompter := templates.NewDefaultPrompter()
	out, err := prompter.RenderTemplate(promptType, data)
	if err != nil {
		t.Fatalf("failed to render template %s: %v", promptType, err)
	}
	return out
}

// --- Pentester template ---

func TestPentesterTemplate_SAGEEnabled(t *testing.T) {
	t.Parallel()

	out := renderTemplate(t, templates.PromptTypePentester, sageTemplateData(true))

	if !strings.Contains(out, "SAGE") {
		t.Error("expected pentester output to contain 'SAGE' when SAGEEnabled=true")
	}
	if !strings.Contains(out, "sage_recall") {
		t.Error("expected pentester output to contain 'sage_recall' when SAGEEnabled=true")
	}
	if !strings.Contains(out, "sage_remember") {
		t.Error("expected pentester output to contain 'sage_remember' when SAGEEnabled=true")
	}
}

func TestPentesterTemplate_SAGEDisabled(t *testing.T) {
	t.Parallel()

	data := sageTemplateData(false)
	out := renderTemplate(t, templates.PromptTypePentester, data)

	if strings.Contains(out, "SAGE") {
		t.Error("expected pentester output NOT to contain 'SAGE' when SAGEEnabled=false")
	}
	if strings.Contains(out, "sage_recall") {
		t.Error("expected pentester output NOT to contain 'sage_recall' when SAGEEnabled=false")
	}
}

// --- Coder template ---

func TestCoderTemplate_SAGEEnabled(t *testing.T) {
	t.Parallel()

	out := renderTemplate(t, templates.PromptTypeCoder, sageTemplateData(true))

	if !strings.Contains(out, "SAGE") {
		t.Error("expected coder output to contain 'SAGE' when SAGEEnabled=true")
	}
	if !strings.Contains(out, "sage_recall") {
		t.Error("expected coder output to contain 'sage_recall' when SAGEEnabled=true")
	}
	if !strings.Contains(out, "sage_remember") {
		t.Error("expected coder output to contain 'sage_remember' when SAGEEnabled=true")
	}
}

func TestCoderTemplate_SAGEDisabled(t *testing.T) {
	t.Parallel()

	out := renderTemplate(t, templates.PromptTypeCoder, sageTemplateData(false))

	if strings.Contains(out, "SAGE") {
		t.Error("expected coder output NOT to contain 'SAGE' when SAGEEnabled=false")
	}
	if strings.Contains(out, "sage_recall") {
		t.Error("expected coder output NOT to contain 'sage_recall' when SAGEEnabled=false")
	}
}

// --- Memorist template ---

func TestMemoristTemplate_SAGEEnabled(t *testing.T) {
	t.Parallel()

	out := renderTemplate(t, templates.PromptTypeMemorist, sageTemplateData(true))

	if !strings.Contains(out, "SAGE") {
		t.Error("expected memorist output to contain 'SAGE' when SAGEEnabled=true")
	}
	if !strings.Contains(out, "sage_recall") {
		t.Error("expected memorist output to contain 'sage_recall' when SAGEEnabled=true")
	}
}

func TestMemoristTemplate_SAGEDisabled(t *testing.T) {
	t.Parallel()

	out := renderTemplate(t, templates.PromptTypeMemorist, sageTemplateData(false))

	if strings.Contains(out, "SAGE") {
		t.Error("expected memorist output NOT to contain 'SAGE' when SAGEEnabled=false")
	}
	if strings.Contains(out, "sage_recall") {
		t.Error("expected memorist output NOT to contain 'sage_recall' when SAGEEnabled=false")
	}
}

// --- Enricher template ---

func TestEnricherTemplate_SAGEEnabled(t *testing.T) {
	t.Parallel()

	out := renderTemplate(t, templates.PromptTypeEnricher, sageTemplateData(true))

	if !strings.Contains(out, "SAGE") && !strings.Contains(out, "sage_recall") {
		t.Error("expected enricher output to contain SAGE references when SAGEEnabled=true")
	}
	if !strings.Contains(out, "sage_recall") {
		t.Error("expected enricher output to contain 'sage_recall' when SAGEEnabled=true")
	}
}

func TestEnricherTemplate_SAGEDisabled(t *testing.T) {
	t.Parallel()

	out := renderTemplate(t, templates.PromptTypeEnricher, sageTemplateData(false))

	if strings.Contains(out, "SAGE") {
		t.Error("expected enricher output NOT to contain 'SAGE' when SAGEEnabled=false")
	}
	if strings.Contains(out, "sage_recall") {
		t.Error("expected enricher output NOT to contain 'sage_recall' when SAGEEnabled=false")
	}
}
