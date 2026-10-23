package models

import (
	"testing"

	"pentagi/cmd/installer/checker"
	"pentagi/cmd/installer/wizard/styles"
	"pentagi/cmd/installer/wizard/window"
)

func TestApplyChanges_GetHelpContent_WarnsThatInstallReusesALeftoverDatabase(t *testing.T) {
	const warning = "The existing PentAGI database will be reused, with its accounts, passwords and flows. Purge PentAGI or Factory Reset removes it"

	for _, tc := range []struct {
		name   string
		result checker.CheckResult
		warned bool
	}{
		{"volumes left behind by a removed container", checker.CheckResult{PentagiVolumesExist: true}, true},
		{"volumes of an installed container", checker.CheckResult{PentagiVolumesExist: true, PentagiInstalled: true}, false},
		{"no volumes", checker.CheckResult{}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			m := NewApplyChangesFormModel(newTestController(t, tc.result), styles.New(), window.New(), nil)

			help := m.GetHelpContent()

			if tc.warned {
				contains(t, help, warning, "apply changes must say the installation it performs reuses the old database")
				absent(t, help, "\n\n\n", "the warning must be set off by a single blank line")
			} else {
				absent(t, help, warning, "apply changes must not warn about a database it will not reuse")
			}
		})
	}
}
