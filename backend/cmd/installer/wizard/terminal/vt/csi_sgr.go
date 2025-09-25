package vt

import (
	uv "github.com/charmbracelet/ultraviolet"
	"github.com/charmbracelet/x/ansi"
)

// handleSgr handles SGR escape sequences.
// handleSgr handles Select Graphic Rendition (SGR) escape sequences.
func (t *Terminal) handleSgr(params ansi.Params) {
	uv.ReadStyle(params, &t.scr.cur.Pen)
}
