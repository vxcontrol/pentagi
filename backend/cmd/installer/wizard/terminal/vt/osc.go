// Package vt provides a virtual terminal implementation.
// SKIP: Fix typecheck errors - function signature mismatches and undefined types
package vt

import (
	"bytes"
	"image/color"
	"io"

	"github.com/charmbracelet/x/ansi"
)

// handleOsc handles an OSC escape sequence.
func (t *Terminal) handleOsc(cmd int, data []byte) {
	t.flushGrapheme() // Flush any pending grapheme before handling OSC sequences.
	if !t.handlers.handleOsc(cmd, data) {
		t.logf("unhandled sequence: OSC %q", data)
	}
}

func (t *Terminal) handleTitle(cmd int, data []byte) {
	parts := bytes.Split(data, []byte{';'})
	if len(parts) != 2 {
		// Invalid, ignore
		return
	}
	switch cmd {
	case 0: // Set window title and icon name
		name := string(parts[1])
		t.iconName, t.title = name, name
		if t.cb.Title != nil {
			t.cb.Title(name)
		}
		if t.cb.IconName != nil {
			t.cb.IconName(name)
		}
	case 1: // Set icon name
		name := string(parts[1])
		t.iconName = name
		if t.cb.IconName != nil {
			t.cb.IconName(name)
		}
	case 2: // Set window title
		name := string(parts[1])
		t.title = name
		if t.cb.Title != nil {
			t.cb.Title(name)
		}
	}
}

func (t *Terminal) handleDefaultColor(cmd int, data []byte) {
	if cmd != 10 && cmd != 11 && cmd != 12 &&
		cmd != 110 && cmd != 111 && cmd != 112 {
		// Invalid, ignore
		return
	}

	parts := bytes.Split(data, []byte{';'})
	if len(parts) == 0 {
		// Invalid, ignore
		return
	}

	cb := func(c color.Color) {
		switch cmd {
		case 10, 110: // Foreground color
			t.SetForegroundColor(c)
		case 11, 111: // Background color
			t.SetBackgroundColor(c)
		case 12, 112: // Cursor color
			t.SetCursorColor(c)
		}
	}

	switch len(parts) {
	case 1: // Reset color
		cb(nil)
	case 2: // Set/Query color
		arg := string(parts[1])
		if arg == "?" {
			var xrgb ansi.XRGBColor
			switch cmd {
			case 10: // Query foreground color
				xrgb.Color = t.ForegroundColor()
				if xrgb.Color != nil {
					io.WriteString(t.pw, ansi.SetForegroundColor(xrgb.String())) //nolint:errcheck,gosec
				}
			case 11: // Query background color
				xrgb.Color = t.BackgroundColor()
				if xrgb.Color != nil {
					io.WriteString(t.pw, ansi.SetBackgroundColor(xrgb.String())) //nolint:errcheck,gosec
				}
			case 12: // Query cursor color
				xrgb.Color = t.CursorColor()
				if xrgb.Color != nil {
					io.WriteString(t.pw, ansi.SetCursorColor(xrgb.String())) //nolint:errcheck,gosec
				}
			}
		} else if c := ansi.XParseColor(arg); c != nil {
			cb(c)
		}
	}
}

func (t *Terminal) handleWorkingDirectory(cmd int, data []byte) {
	if cmd != 7 {
		// Invalid, ignore
		return
	}

	// The data is the working directory path.
	parts := bytes.Split(data, []byte{';'})
	if len(parts) != 2 {
		// Invalid, ignore
		return
	}

	path := string(parts[1])
	t.cwd = path

	if t.cb.WorkingDirectory != nil {
		t.cb.WorkingDirectory(path)
	}
}

func (t *Terminal) handleHyperlink(cmd int, data []byte) {
	parts := bytes.Split(data, []byte{';'})
	if len(parts) != 3 || cmd != 8 {
		// Invalid, ignore
		return
	}

	t.scr.cur.Link.URL = string(parts[1])
	t.scr.cur.Link.Params = string(parts[2])
}
