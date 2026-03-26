package vt

import "github.com/charmbracelet/x/ansi"

// resetModes resets all modes to their default values.
func (t *Terminal) resetModes() {
	t.modes = ansi.Modes{
		// Recognized modes and their default values.
		ansi.CursorKeysMode:          ansi.ModeReset, // ?1
		ansi.OriginMode:              ansi.ModeReset, // ?6
		ansi.AutoWrapMode:            ansi.ModeSet,   // ?7
		ansi.X10MouseMode:            ansi.ModeReset, // ?9
		ansi.LineFeedNewLineMode:     ansi.ModeReset, // ?20
		ansi.TextCursorEnableMode:    ansi.ModeSet,   // ?25
		ansi.NumericKeypadMode:       ansi.ModeReset, // ?66
		ansi.LeftRightMarginMode:     ansi.ModeReset, // ?69
		ansi.NormalMouseMode:         ansi.ModeReset, // ?1000
		ansi.HighlightMouseMode:      ansi.ModeReset, // ?1001
		ansi.ButtonEventMouseMode:    ansi.ModeReset, // ?1002
		ansi.AnyEventMouseMode:       ansi.ModeReset, // ?1003
		ansi.FocusEventMode:          ansi.ModeReset, // ?1004
		ansi.SgrExtMouseMode:         ansi.ModeReset, // ?1006
		ansi.AltScreenMode:           ansi.ModeReset, // ?1047
		ansi.SaveCursorMode:          ansi.ModeReset, // ?1048
		ansi.AltScreenSaveCursorMode: ansi.ModeReset, // ?1049
		ansi.BracketedPasteMode:      ansi.ModeReset, // ?2004
	}

	// Set mode effects.
	for mode, setting := range t.modes {
		t.setMode(mode, setting)
	}
}
