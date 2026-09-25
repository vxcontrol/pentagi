package models

import (
	"os"
	"regexp"
	"strings"
	"testing"

	"pentagi/pkg/providers/provider"
)

func TestLLMProviders_LoadItems_OffersEveryProductDoor(t *testing.T) {
	offered := map[LLMProviderID]bool{}
	for _, id := range offeredProviders(t) {
		offered[id] = true
	}

	for _, door := range provider.AllProviderTypes {
		if !offered[LLMProviderID(door)] {
			t.Errorf("the product has a %s provider the wizard does not offer", door)
		}
		delete(offered, LLMProviderID(door))
	}
	for id := range offered {
		t.Errorf("the wizard offers %s, which the product does not know", id)
	}
}

func TestLLMProviders_LoadItems_OffersOnlyDoorsWithARegisteredScreen(t *testing.T) {
	types, err := os.ReadFile("types.go")
	if err != nil {
		t.Fatal(err)
	}
	registry, err := os.ReadFile("../registry/registry.go")
	if err != nil {
		t.Fatal(err)
	}

	constants := map[string]string{}
	for _, m := range regexp.MustCompile(`(LLMProvider\w+Screen)\s+ScreenID = "`+providerFormPrefix+`(\w+)"`).
		FindAllStringSubmatch(string(types), -1) {
		constants[m[2]] = m[1]
	}

	for _, id := range offeredProviders(t) {
		constant, ok := constants[string(id)]
		if !ok {
			t.Errorf("%s has no screen constant in types.go", id)
			continue
		}
		if !strings.Contains(string(registry), "r.screens[models."+constant+"]") {
			t.Errorf("%s is offered but %s is never registered", id, constant)
		}
	}
}
