package vt

import (
	"unicode/utf8"

	uv "github.com/charmbracelet/ultraviolet"
	"github.com/charmbracelet/x/ansi"
	"github.com/mattn/go-runewidth"
	"github.com/rivo/uniseg"
)

// handlePrint handles printable characters.
func (t *Terminal) handlePrint(r rune) {
	if r >= ansi.SP && r < ansi.DEL {
		if len(t.grapheme) > 0 {
			// If we have a grapheme buffer, flush it before handling the ASCII character.
			t.flushGrapheme()
		}
		t.handleGrapheme(string(r), 1)
	} else {
		t.grapheme = append(t.grapheme, r)
	}
}

// flushGrapheme flushes the current grapheme buffer, if any, and handles the
// grapheme as a single unit.
func (t *Terminal) flushGrapheme() {
	if len(t.grapheme) == 0 {
		return
	}

	unicode := t.isModeSet(ansi.UnicodeCoreMode)
	gr := string(t.grapheme)

	var cl string
	var w int
	state := -1
	for len(gr) > 0 {
		cl, gr, w, state = uniseg.FirstGraphemeClusterInString(gr, state)
		if !unicode {
			//nolint:godox
			// TODO: Investigate this further, runewidth.StringWidth doesn't
			// report the correct width for some edge cases such as variation
			// selectors.
			w = 0
			for _, r := range cl {
				if r >= 0xFE00 && r <= 0xFE0F {
					// Variation Selectors 1 - 16
					continue
				}
				if r >= 0xE0100 && r <= 0xE01EF {
					// Variation Selectors 17-256
					continue
				}
				w += runewidth.RuneWidth(r)
			}
		}
		t.handleGrapheme(cl, w)
	}
	t.grapheme = t.grapheme[:0] // Reset the grapheme buffer.
}

// handleGrapheme handles UTF-8 graphemes.
func (t *Terminal) handleGrapheme(content string, width int) {
	awm := t.isModeSet(ansi.AutoWrapMode)
	cell := uv.Cell{
		Content: content,
		Width:   width,
		Style:   t.scr.cursorPen(),
		Link:    t.scr.cursorLink(),
	}

	x, y := t.scr.CursorPosition()
	if t.atPhantom && awm {
		// moves cursor down similar to [Terminal.linefeed] except it doesn't
		// respects [ansi.LNM] mode.
		// This will reset the phantom state i.e. pending wrap state.
		t.index()
		_, y = t.scr.CursorPosition()
		x = 0
	}

	// Handle character set mappings
	if len(content) == 1 { //nolint:nestif
		var charset CharSet
		c := content[0]
		if t.gsingle > 1 && t.gsingle < 4 {
			charset = t.charsets[t.gsingle]
			t.gsingle = 0
		} else if c < 128 {
			charset = t.charsets[t.gl]
		} else {
			charset = t.charsets[t.gr]
		}

		if charset != nil {
			if r, ok := charset[c]; ok {
				cell.Content = r
				cell.Width = 1
			}
		}
	}

	if cell.Width == 1 && len(content) == 1 {
		t.lastChar, _ = utf8.DecodeRuneInString(content)
	}

	t.scr.SetCell(x, y, &cell)

	// Handle phantom state at the end of the line
	t.atPhantom = awm && x >= t.scr.Width()-1
	if !t.atPhantom {
		x += cell.Width
	}

	// NOTE: We don't reset the phantom state here, we handle it up above.
	t.scr.setCursor(x, y, false)
}
