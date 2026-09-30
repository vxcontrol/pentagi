package models

import (
	"os"
	"path/filepath"
	"slices"
	"testing"

	"pentagi/cmd/installer/checker"
	"pentagi/cmd/installer/files"
	"pentagi/cmd/installer/state"
	"pentagi/cmd/installer/wizard/controller"
	"pentagi/cmd/installer/wizard/styles"
	"pentagi/cmd/installer/wizard/window"
)

func TestMainMenu_LoadItems_HighlightsMaintenanceWhileAnUpdateWaits(t *testing.T) {
	upToDate := checker.CheckResult{
		WorkerImageExists: true, WorkerIsUpToDate: true,
		PentagiInstalled: true, PentagiIsUpToDate: true,
		InstallerIsUpToDate: true, UpdateServerAccessible: true,
	}

	for _, tc := range []struct {
		name   string
		update func(*checker.CheckResult)
		want   bool
	}{
		{"everything is up to date", func(*checker.CheckResult) {}, false},
		{"a newer worker image", func(c *checker.CheckResult) { c.WorkerIsUpToDate = false }, true},
		{"a newer stack", func(c *checker.CheckResult) { c.PentagiIsUpToDate = false }, true},
		{"a newer installer", func(c *checker.CheckResult) { c.InstallerIsUpToDate = false }, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			envPath := filepath.Join(t.TempDir(), ".env")
			if err := os.WriteFile(envPath, nil, 0o600); err != nil {
				t.Fatal(err)
			}
			st, err := state.NewState(envPath)
			if err != nil {
				t.Fatal(err)
			}
			result := upToDate
			tc.update(&result)
			c := controller.NewController(st, files.NewFiles(), result)

			items := NewMainMenuHandler(c, styles.New(), window.New()).LoadItems()

			i := slices.IndexFunc(items, func(item ListItem) bool { return item.ID == MaintenanceScreen })
			if i < 0 {
				t.Fatal("maintenance is not offered")
			}
			if items[i].Highlighted != tc.want {
				t.Errorf("maintenance highlighted = %v, want %v", items[i].Highlighted, tc.want)
			}
		})
	}
}
