package terminal

import (
	"testing"

	"pentagi/cmd/installer/wizard/terminal/vt"

	tea "github.com/charmbracelet/bubbletea"
)

func TestKey2uv_TeaKeyToUVKey_MapsEachKeyToItsVTPress(t *testing.T) {
	for _, tc := range []struct {
		name string
		key  tea.KeyMsg
		want vt.KeyPressEvent
	}{
		{"the up arrow", tea.KeyMsg{Type: tea.KeyUp}, vt.KeyPressEvent{Code: vt.KeyUp}},
		{"the down arrow", tea.KeyMsg{Type: tea.KeyDown}, vt.KeyPressEvent{Code: vt.KeyDown}},
		{"the left arrow", tea.KeyMsg{Type: tea.KeyLeft}, vt.KeyPressEvent{Code: vt.KeyLeft}},
		{"the right arrow", tea.KeyMsg{Type: tea.KeyRight}, vt.KeyPressEvent{Code: vt.KeyRight}},
		{"the enter key", tea.KeyMsg{Type: tea.KeyEnter}, vt.KeyPressEvent{Code: vt.KeyEnter}},
		{"the tab key", tea.KeyMsg{Type: tea.KeyTab}, vt.KeyPressEvent{Code: vt.KeyTab}},
		{"the space bar", tea.KeyMsg{Type: tea.KeySpace}, vt.KeyPressEvent{Code: vt.KeySpace}},
		{"a typed letter", tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}}, vt.KeyPressEvent{Code: 'a'}},
		{"ctrl+c as ctrl and c", tea.KeyMsg{Type: tea.KeyCtrlC}, vt.KeyPressEvent{Code: 'c', Mod: vt.ModCtrl}},
		{"ctrl+d as ctrl and d", tea.KeyMsg{Type: tea.KeyCtrlD}, vt.KeyPressEvent{Code: 'd', Mod: vt.ModCtrl}},
		{"alt with an arrow keeps alt", tea.KeyMsg{Type: tea.KeyUp, Alt: true}, vt.KeyPressEvent{Code: vt.KeyUp, Mod: vt.ModAlt}},
		{"shift+tab as shift and tab", tea.KeyMsg{Type: tea.KeyShiftTab}, vt.KeyPressEvent{Code: vt.KeyTab, Mod: vt.ModShift}},
		{"a function key", tea.KeyMsg{Type: tea.KeyF1}, vt.KeyPressEvent{Code: vt.KeyF1}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			result := teaKeyToUVKey(tc.key)
			got, ok := result.(vt.KeyPressEvent)
			if !ok {
				t.Fatalf("teaKeyToUVKey(%+v) = %T, want a key press", tc.key, result)
			}
			if got.Code != tc.want.Code || got.Mod != tc.want.Mod {
				t.Errorf("teaKeyToUVKey(%+v) = {Code: %v, Mod: %v}, want {Code: %v, Mod: %v}",
					tc.key, got.Code, got.Mod, tc.want.Code, tc.want.Mod)
			}
		})
	}
}

func BenchmarkKeySequenceConversion(b *testing.B) {
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = teaKeyToUVKey(msg)
	}
}
