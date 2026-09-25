package models

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestTypes_RestoreModel_RestoresTheUpdateScreens(t *testing.T) {
	for _, model := range []tea.Model{(*InstallerUpdateModel)(nil), (*UpdateOverviewModel)(nil)} {
		if RestoreModel(model) == nil {
			t.Errorf("%T is missing its branch in RestoreModel", model)
		}
	}
}
