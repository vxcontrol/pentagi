package models

import (
	"testing"

	"pentagi/cmd/installer/checker"
	"pentagi/cmd/installer/processor"
	"pentagi/cmd/installer/wizard/styles"
	"pentagi/cmd/installer/wizard/window"
)

func TestProcessorOperationForm_GetHelpContent_WarnsThatInstallReusesALeftoverDatabase(t *testing.T) {
	const warning = "The existing PentAGI database will be reused, with its accounts, passwords and flows. Purge PentAGI or Factory Reset removes it"

	for _, tc := range []struct {
		name      string
		operation processor.ProcessorOperation
		result    checker.CheckResult
		warned    bool
	}{
		{"volumes left behind by a removed container", processor.ProcessorOperationInstall, checker.CheckResult{PentagiVolumesExist: true}, true},
		{"volumes of an installed container", processor.ProcessorOperationInstall, checker.CheckResult{PentagiVolumesExist: true, PentagiInstalled: true}, false},
		{"no volumes", processor.ProcessorOperationInstall, checker.CheckResult{}, false},
		{"purge of volumes left behind", processor.ProcessorOperationPurge, checker.CheckResult{PentagiVolumesExist: true}, false},
		{"factory reset of volumes left behind", processor.ProcessorOperationFactoryReset, checker.CheckResult{PentagiVolumesExist: true}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			m := NewProcessorOperationFormModel(newTestController(t, tc.result), styles.New(), window.New(),
				nil, processor.ProductStackAll, tc.operation)

			help := m.GetHelpContent()

			if tc.warned {
				contains(t, help, warning, "the install screen must say the old database will be reused")
			} else {
				absent(t, help, warning, "only an install reuses the old database")
			}
		})
	}
}
