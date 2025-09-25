package vt

import (
	"github.com/charmbracelet/x/ansi"
)

// handleEsc handles an escape sequence.
func (t *Terminal) handleEsc(cmd ansi.Cmd) {
	t.flushGrapheme() // Flush any pending grapheme before handling ESC sequences.
	if !t.handlers.handleEsc(int(cmd)) {
		var str string
		if inter := cmd.Intermediate(); inter != 0 {
			str += string(inter) + " "
		}
		if final := cmd.Final(); final != 0 {
			str += string(final)
		}
		t.logf("unhandled sequence: ESC %q", str)
	}
}

// fullReset performs a full terminal reset as in [ansi.RIS].
func (t *Terminal) fullReset() {
	t.scrs[0].Reset()
	t.scrs[1].Reset()
	t.resetTabStops()
	t.resetModes()

	t.gl, t.gr = 0, 1
	t.gsingle = 0
	t.charsets = [4]CharSet{}
	t.atPhantom = false
}
