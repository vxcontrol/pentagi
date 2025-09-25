package vt

import "github.com/charmbracelet/x/ansi"

// handleDcs handles a DCS escape sequence.
func (t *Terminal) handleDcs(cmd ansi.Cmd, params ansi.Params, data []byte) {
	t.flushGrapheme() // Flush any pending grapheme before handling DCS sequences.
	if !t.handlers.handleDcs(cmd, params, data) {
		t.logf("unhandled sequence: DCS %q %q", paramsString(cmd, params), data)
	}
}

// handleApc handles an APC escape sequence.
func (t *Terminal) handleApc(data []byte) {
	t.flushGrapheme() // Flush any pending grapheme before handling APC sequences.
	if !t.handlers.handleApc(data) {
		t.logf("unhandled sequence: APC %q", data)
	}
}

// handleSos handles an SOS escape sequence.
func (t *Terminal) handleSos(data []byte) {
	t.flushGrapheme() // Flush any pending grapheme before handling SOS sequences.
	if !t.handlers.handleSos(data) {
		t.logf("unhandled sequence: SOS %q", data)
	}
}

// handlePm handles a PM escape sequence.
func (t *Terminal) handlePm(data []byte) {
	t.flushGrapheme() // Flush any pending grapheme before handling PM sequences.
	if !t.handlers.handlePm(data) {
		t.logf("unhandled sequence: PM %q", data)
	}
}
