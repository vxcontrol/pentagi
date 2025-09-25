package vt

import (
	uv "github.com/charmbracelet/ultraviolet"
)

// eraseCharacter erases n characters starting from the cursor position. It
// does not move the cursor. This is equivalent to [ansi.ECH].
func (t *Terminal) eraseCharacter(n int) {
	if n <= 0 {
		n = 1
	}
	x, y := t.scr.CursorPosition()
	rect := uv.Rect(x, y, n, 1)
	t.scr.FillArea(t.scr.blankCell(), rect)
	t.atPhantom = false
	// ECH does not move the cursor.
}
