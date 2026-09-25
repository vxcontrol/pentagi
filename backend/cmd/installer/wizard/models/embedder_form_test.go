package models

import (
	"cmp"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"pentagi/cmd/installer/checker"
	"pentagi/cmd/installer/files"
	"pentagi/cmd/installer/state"
	"pentagi/cmd/installer/wizard/controller"
	"pentagi/cmd/installer/wizard/locale"
	"pentagi/cmd/installer/wizard/styles"
	"pentagi/cmd/installer/wizard/window"
)

func embedderFormOver(t *testing.T, vars map[string]string) (*EmbedderFormModel, controller.Controller) {
	t.Helper()

	example, err := os.ReadFile(filepath.Join("..", "..", "..", "..", "..", ".env.example"))
	if err != nil {
		t.Fatal(err)
	}
	envPath := filepath.Join(t.TempDir(), ".env")
	if err := os.WriteFile(envPath, example, 0o600); err != nil {
		t.Fatal(err)
	}

	st, err := state.NewState(envPath)
	if err != nil {
		t.Fatal(err)
	}
	c := controller.NewController(st, files.NewFiles(), checker.CheckResult{})
	for name, value := range vars {
		if err := c.SetVar(name, value); err != nil {
			t.Fatal(err)
		}
	}

	return NewEmbedderFormModel(c, styles.New(), window.New()), c
}

func TestEmbedderForm_HandleSave_ClearsAValueTheServerStillReads(t *testing.T) {
	for _, tc := range []struct {
		provider, variable, field, leftover string
	}{
		{"", "EMBEDDING_MODEL", "model", "mistral-embed"},
		{"mistral", "EMBEDDING_MODEL", "model", "text-embedding-3-large"},
		{"voyageai", "EMBEDDING_URL", "url", "https://llm.example/openai/v1"},
	} {
		name := cmp.Or(tc.provider, locale.EmbedderProviderIDDefault)
		t.Run("the "+name+" form clears its "+tc.field+" field", func(t *testing.T) {
			form, c := embedderFormOver(t, map[string]string{
				"EMBEDDING_PROVIDER": tc.provider,
				tc.variable:          tc.leftover,
			})
			form.BuildForm()

			i := slices.IndexFunc(form.GetFormFields(), func(f FormField) bool { return f.Key == tc.field })
			if i < 0 {
				t.Fatalf("the %s form hides %s, so a value left from another provider is sent and cannot be cleared",
					name, tc.variable)
			}

			form.fields[i].Input.SetValue("")
			if err := form.HandleSave(); err != nil {
				t.Fatal(err)
			}

			if saved, _ := c.GetVar(tc.variable); saved.Value != "" {
				t.Errorf("clearing %s in the %s form kept %q", tc.variable, name, saved.Value)
			}
		})
	}
}

func TestEmbedderForm_InitEmbeddingProviders_ShowsTheOpenAIKeyRuleInBothOpenAIChoices(t *testing.T) {
	providers := initEmbeddingProviders()

	for _, id := range []string{locale.EmbedderProviderIDDefault, locale.EmbedderProviderIDOpenAI} {
		contains(t, providers[id].HelpText, "When the embedder has no API Key, it uses the OpenAI key from LLM Providers, "+
			"and only with the OpenAI server set there: its API Endpoint URL must be empty or name that server. "+
			"An embedder API Key goes to the API Endpoint URL, or to api.openai.com when the URL is empty. "+
			"Any other endpoint needs its own API Key; a local server that checks no key accepts any non-empty value.",
			"the help panel of the "+id+" embedder does not say which key goes to which server")
	}
}
