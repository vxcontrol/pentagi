package models

import (
	"errors"
	"testing"

	"pentagi/cmd/installer/wizard/styles"
	"pentagi/cmd/installer/wizard/window"

	tea "github.com/charmbracelet/bubbletea"
)

type saveHandler struct {
	*MockFormModel
	err error
}

func (h *saveHandler) HandleSave() error { return h.err }

func screenSavingWith(saveErr error) (*BaseScreen, *saveHandler) {
	handler := &saveHandler{
		MockFormModel: NewMockFormModel(nil, styles.New(), window.New(), "form", "Form", "A form"),
		err:           saveErr,
	}
	screen := NewBaseScreen(nil, styles.New(), window.New(), handler, nil)
	screen.SetHasChanges(true)
	screen.View()

	return screen, handler
}

func navigationOf(cmd tea.Cmd) (NavigationMsg, bool) {
	if cmd == nil {
		return NavigationMsg{}, false
	}
	nav, ok := cmd().(NavigationMsg)

	return nav, ok
}

func TestBaseScreen_Update_ShowsTheSaveErrorInTheStatusLine(t *testing.T) {
	for _, tc := range []struct {
		name string
		key  tea.KeyMsg
	}{
		{"ctrl+s", tea.KeyMsg{Type: tea.KeyCtrlS}},
		{"enter", tea.KeyMsg{Type: tea.KeyEnter}},
	} {
		t.Run(tc.name+" replaces the unsaved-changes warning with the error", func(t *testing.T) {
			screen, _ := screenSavingWith(errors.New("disk full"))

			screen.Update(tc.key)

			view := screen.View()
			contains(t, view, "Save failed: disk full", "the user must be told the save did not happen")
			absent(t, view, "Unsaved changes", "the error replaces the warning in the single status line")
			absent(t, view, "Configuration saved", "nothing was saved")
		})
	}
}

func TestBaseScreen_Update_EnterStaysOnTheScreenWhenSaveFails(t *testing.T) {
	screen, _ := screenSavingWith(errors.New("disk full"))

	cmd := screen.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if nav, ok := navigationOf(cmd); ok {
		t.Errorf("enter navigated with %+v although the save failed", nav)
	}
}

func TestBaseScreen_Update_ASuccessfulSaveClearsTheSaveError(t *testing.T) {
	t.Run("ctrl+s shows the configuration as saved", func(t *testing.T) {
		screen, handler := screenSavingWith(errors.New("disk full"))
		screen.Update(tea.KeyMsg{Type: tea.KeyCtrlS})
		handler.err = nil

		screen.Update(tea.KeyMsg{Type: tea.KeyCtrlS})

		view := screen.View()
		contains(t, view, "Configuration saved", "the save went through")
		absent(t, view, "Save failed", "the earlier error is stale")
	})

	t.Run("enter navigates back and no error is left on screen", func(t *testing.T) {
		screen, handler := screenSavingWith(errors.New("disk full"))
		screen.Update(tea.KeyMsg{Type: tea.KeyEnter})
		handler.err = nil

		nav, ok := navigationOf(screen.Update(tea.KeyMsg{Type: tea.KeyEnter}))

		if want := (NavigationMsg{GoBack: true}); !ok || nav != want {
			t.Errorf("enter produced %+v (navigation: %v), want %+v", nav, ok, want)
		}
		view := screen.View()
		contains(t, view, "Configuration saved", "the save went through")
		absent(t, view, "Save failed", "the earlier error is stale")
	})
}

func TestBaseScreen_Update_ResetDropsTheSaveError(t *testing.T) {
	screen, _ := screenSavingWith(errors.New("disk full"))
	screen.Update(tea.KeyMsg{Type: tea.KeyCtrlS})

	screen.Update(tea.KeyMsg{Type: tea.KeyCtrlR})

	view := screen.View()
	contains(t, view, "Configuration saved", "the reset form matches what is stored")
	absent(t, view, "Save failed", "the discarded edits are no longer the ones that failed to save")
}
