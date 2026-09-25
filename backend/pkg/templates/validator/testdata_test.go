package validator

import (
	"sort"
	"testing"

	"pentagi/pkg/templates"
)

func TestTestdata_CreateDummyTemplateData_CoversEveryPromptVariable(t *testing.T) {
	dummyData := CreateDummyTemplateData()

	var missing []string
	for _, variables := range templates.PromptVariables {
		for _, variable := range variables {
			if _, ok := dummyData[variable]; !ok {
				missing = append(missing, variable)
			}
		}
	}

	if len(missing) > 0 {
		sort.Strings(missing)
		t.Errorf("CreateDummyTemplateData is missing variables declared in PromptVariables: %v", missing)
	}
}
