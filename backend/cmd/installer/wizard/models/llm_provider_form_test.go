package models

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strings"
	"testing"

	"pentagi/cmd/installer/checker"
	"pentagi/cmd/installer/files"
	"pentagi/cmd/installer/state"
	"pentagi/cmd/installer/wizard/controller"
	"pentagi/cmd/installer/wizard/locale"
	"pentagi/cmd/installer/wizard/styles"
	"pentagi/cmd/installer/wizard/window"
)

// A field the save switch does not name still renders, and is dropped on save without an error.
func TestLLMProviderForm_HandleSave_ReadsEveryFormFieldKey(t *testing.T) {
	source, err := os.ReadFile("llm_provider_form.go")
	if err != nil {
		t.Fatal(err)
	}
	text := string(source)

	keys := map[string]bool{}
	for _, m := range regexp.MustCompile(`Key:\s*"([a-z_]+)"`).FindAllStringSubmatch(text, -1) {
		keys[m[1]] = true
	}
	if len(keys) < 10 {
		t.Fatalf("found only %d field keys, the pattern stopped matching", len(keys))
	}

	start := strings.Index(text, "func (m *LLMProviderFormModel) HandleSave()")
	if start < 0 {
		t.Fatal("HandleSave is gone or renamed")
	}
	end := strings.Index(text[start+1:], "\nfunc ")
	if end < 0 {
		t.Fatal("could not find the end of HandleSave")
	}
	save := text[start : start+1+end]

	handled := map[string]bool{}
	for _, m := range regexp.MustCompile(`case "([a-z_]+)":`).FindAllStringSubmatch(save, -1) {
		handled[m[1]] = true
	}

	var missing []string
	for key := range keys {
		if !handled[key] {
			missing = append(missing, key)
		}
	}
	sort.Strings(missing)
	if len(missing) > 0 {
		t.Errorf("HandleSave never reads %v, so those fields are discarded on save", missing)
	}
}

func TestLLMProviderForm_GivesEveryOfferedDoorItsOwnNameDescriptionAndURL(t *testing.T) {
	for _, id := range offeredProviders(t) {
		form := &LLMProviderFormModel{providerID: id}

		if form.GetFormName() == fmt.Sprintf(locale.LLMProviderFormName, "") {
			t.Errorf("%s has no form name", id)
		}
		if form.GetFormDescription() == locale.LLMProviderFormDescription {
			t.Errorf("%s has no form description", id)
		}
		if id != LLMProviderBedrock && form.getDefaultBaseURL() == "" {
			t.Errorf("%s has no default server URL", id)
		}
	}
}

// The env file has no PENTAGI_BEDROCK_CONFIG_PATH line, as on a host whose .env predates the variable.
func TestLLMProviderForm_HandleSave_MountsABedrockConfigTypedAsAFileOnTheHost(t *testing.T) {
	onHost := filepath.Join(t.TempDir(), "my-bedrock.provider.yml")
	if err := os.WriteFile(onHost, []byte("simple:\n  model: zai.glm-4.7-flash\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	mountedAt, shipped := "/opt/pentagi/conf/bedrock.provider.yml", "/opt/pentagi/conf/bedrock-glm-flash.provider.yml"
	typed := func(path string) *string { return &path }
	for _, tc := range []struct {
		name                    string
		container               string  // BEDROCK_CONFIG_PATH in the env file
		typed                   *string // nil: the form is saved as it was shown
		wantShown               string
		wantContainer, wantHost string
		wantErr                 string
	}{
		{name: "a file on the host", typed: typed(onHost), wantContainer: mountedAt, wantHost: onHost},
		{name: "a config the image ships", typed: typed(shipped), wantContainer: shipped},
		{name: "the path the default example is mounted at", typed: typed(mountedAt), wantContainer: mountedAt},
		{name: "a path that is on neither", typed: typed("/opt/pentagi/conf/bedrock.yml"),
			wantErr: "config file does not exist: /opt/pentagi/conf/bedrock.yml"},
		{name: "the default example in use is shown and kept by an untouched save",
			container: mountedAt, wantShown: mountedAt, wantContainer: mountedAt},
		{name: "a config the image ships in use is shown and kept by an untouched save",
			container: shipped, wantShown: shipped, wantContainer: shipped},
	} {
		t.Run(tc.name, func(t *testing.T) {
			example, err := os.ReadFile(filepath.Join("..", "..", "..", "..", "..", ".env.example"))
			if err != nil {
				t.Fatal(err)
			}
			env := strings.Replace(string(example), "\nPENTAGI_BEDROCK_CONFIG_PATH=\n", "\n", 1)
			env = strings.Replace(env, "\nBEDROCK_CONFIG_PATH=\n", "\nBEDROCK_CONFIG_PATH="+tc.container+"\n", 1)
			if env == string(example) || strings.Contains(env, "PENTAGI_BEDROCK_CONFIG_PATH") {
				t.Fatal("the example env no longer holds the two bedrock config lines as empty assignments")
			}
			envPath := filepath.Join(t.TempDir(), ".env")
			if err := os.WriteFile(envPath, []byte(env), 0o600); err != nil {
				t.Fatal(err)
			}
			st, err := state.NewState(envPath)
			if err != nil {
				t.Fatal(err)
			}
			c := controller.NewController(st, files.NewFiles(), checker.CheckResult{})
			form := NewLLMProviderFormModel(c, styles.New(), window.New(), LLMProviderBedrock)
			form.BuildForm()

			i := slices.IndexFunc(form.GetFormFields(), func(f FormField) bool { return f.Key == "config_path" })
			if i < 0 {
				t.Fatal("the bedrock form has no config path field")
			}
			if got := form.fields[i].Input.Value(); got != tc.wantShown {
				t.Errorf("the field shows %q, want %q", got, tc.wantShown)
			}
			if tc.typed != nil {
				form.fields[i].Input.SetValue(*tc.typed)
			}

			err = form.HandleSave()
			if tc.wantErr != "" {
				if err == nil || err.Error() != tc.wantErr {
					t.Fatalf("HandleSave() = %v, want %q", err, tc.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if got, _ := c.GetVar("BEDROCK_CONFIG_PATH"); got.Value != tc.wantContainer {
				t.Errorf("BEDROCK_CONFIG_PATH = %q, want %q", got.Value, tc.wantContainer)
			}
			if got, _ := c.GetVar("PENTAGI_BEDROCK_CONFIG_PATH"); got.Value != tc.wantHost {
				t.Errorf("PENTAGI_BEDROCK_CONFIG_PATH = %q, want %q", got.Value, tc.wantHost)
			}
		})
	}
}
