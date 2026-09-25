package models

import (
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
	"testing"

	"pentagi/cmd/installer/wizard/locale"
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
