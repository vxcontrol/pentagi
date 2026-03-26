package vt

import (
	"io"

	uv "github.com/charmbracelet/ultraviolet"
	"github.com/charmbracelet/x/ansi"
)

// DcsHandler is a function that handles a DCS escape sequence.
type DcsHandler func(params ansi.Params, data []byte) bool

// CsiHandler is a function that handles a CSI escape sequence.
type CsiHandler func(params ansi.Params) bool

// OscHandler is a function that handles an OSC escape sequence.
type OscHandler func(data []byte) bool

// ApcHandler is a function that handles an APC escape sequence.
type ApcHandler func(data []byte) bool

// SosHandler is a function that handles an SOS escape sequence.
type SosHandler func(data []byte) bool

// PmHandler is a function that handles a PM escape sequence.
type PmHandler func(data []byte) bool

// EscHandler is a function that handles an ESC escape sequence.
type EscHandler func() bool

// CcHandler is a function that handles a control character.
type CcHandler func() bool

// handlers contains the terminal's escape sequence handlers.
type handlers struct {
	ccHandlers  map[byte][]CcHandler
	dcsHandlers map[int][]DcsHandler
	csiHandlers map[int][]CsiHandler
	oscHandlers map[int][]OscHandler
	escHandler  map[int][]EscHandler
	apcHandlers []ApcHandler
	sosHandlers []SosHandler
	pmHandlers  []PmHandler
}

// RegisterDcsHandler registers a DCS escape sequence handler.
func (h *handlers) RegisterDcsHandler(cmd int, handler DcsHandler) {
	if h.dcsHandlers == nil {
		h.dcsHandlers = make(map[int][]DcsHandler)
	}
	h.dcsHandlers[cmd] = append(h.dcsHandlers[cmd], handler)
}

// RegisterCsiHandler registers a CSI escape sequence handler.
func (h *handlers) RegisterCsiHandler(cmd int, handler CsiHandler) {
	if h.csiHandlers == nil {
		h.csiHandlers = make(map[int][]CsiHandler)
	}
	h.csiHandlers[cmd] = append(h.csiHandlers[cmd], handler)
}

// RegisterOscHandler registers an OSC escape sequence handler.
func (h *handlers) RegisterOscHandler(cmd int, handler OscHandler) {
	if h.oscHandlers == nil {
		h.oscHandlers = make(map[int][]OscHandler)
	}
	h.oscHandlers[cmd] = append(h.oscHandlers[cmd], handler)
}

// RegisterApcHandler registers an APC escape sequence handler.
func (h *handlers) RegisterApcHandler(handler ApcHandler) {
	h.apcHandlers = append(h.apcHandlers, handler)
}

// RegisterSosHandler registers an SOS escape sequence handler.
func (h *handlers) RegisterSosHandler(handler SosHandler) {
	h.sosHandlers = append(h.sosHandlers, handler)
}

// RegisterPmHandler registers a PM escape sequence handler.
func (h *handlers) RegisterPmHandler(handler PmHandler) {
	h.pmHandlers = append(h.pmHandlers, handler)
}

// RegisterEscHandler registers an ESC escape sequence handler.
func (h *handlers) RegisterEscHandler(cmd int, handler EscHandler) {
	if h.escHandler == nil {
		h.escHandler = make(map[int][]EscHandler)
	}
	h.escHandler[cmd] = append(h.escHandler[cmd], handler)
}

// registerCcHandler registers a control character handler.
func (h *handlers) registerCcHandler(r byte, handler CcHandler) {
	if h.ccHandlers == nil {
		h.ccHandlers = make(map[byte][]CcHandler)
	}
	h.ccHandlers[r] = append(h.ccHandlers[r], handler)
}

// handleCc handles a control character.
// It returns true if the control character was handled.
func (h *handlers) handleCc(r byte) bool {
	// Reverse iterate over the handlers so that the last registered handler
	// is the first to be called.
	for i := len(h.ccHandlers[r]) - 1; i >= 0; i-- {
		if h.ccHandlers[r][i]() {
			return true
		}
	}
	return false
}

// handleDcs handles a DCS escape sequence.
// It returns true if the sequence was handled.
func (h *handlers) handleDcs(cmd ansi.Cmd, params ansi.Params, data []byte) bool {
	// Reverse iterate over the handlers so that the last registered handler
	// is the first to be called.
	if handlers, ok := h.dcsHandlers[int(cmd)]; ok {
		for i := len(handlers) - 1; i >= 0; i-- {
			if handlers[i](params, data) {
				return true
			}
		}
	}
	return false
}

// handleCsi handles a CSI escape sequence.
// It returns true if the sequence was handled.
func (h *handlers) handleCsi(cmd ansi.Cmd, params ansi.Params) bool {
	// Reverse iterate over the handlers so that the last registered handler
	// is the first to be called.
	if handlers, ok := h.csiHandlers[int(cmd)]; ok {
		for i := len(handlers) - 1; i >= 0; i-- {
			if handlers[i](params) {
				return true
			}
		}
	}
	return false
}

// handleOsc handles an OSC escape sequence.
// It returns true if the sequence was handled.
func (h *handlers) handleOsc(cmd int, data []byte) bool {
	// Reverse iterate over the handlers so that the last registered handler
	// is the first to be called.
	if handlers, ok := h.oscHandlers[cmd]; ok {
		for i := len(handlers) - 1; i >= 0; i-- {
			if handlers[i](data) {
				return true
			}
		}
	}
	return false
}

// handleApc handles an APC escape sequence.
// It returns true if the sequence was handled.
func (h *handlers) handleApc(data []byte) bool {
	// Reverse iterate over the handlers so that the last registered handler
	// is the first to be called.
	for i := len(h.apcHandlers) - 1; i >= 0; i-- {
		if h.apcHandlers[i](data) {
			return true
		}
	}
	return false
}

// handleSos handles an SOS escape sequence.
// It returns true if the sequence was handled.
func (h *handlers) handleSos(data []byte) bool {
	// Reverse iterate over the handlers so that the last registered handler
	// is the first to be called.
	for i := len(h.sosHandlers) - 1; i >= 0; i-- {
		if h.sosHandlers[i](data) {
			return true
		}
	}
	return false
}

// handlePm handles a PM escape sequence.
// It returns true if the sequence was handled.
func (h *handlers) handlePm(data []byte) bool {
	// Reverse iterate over the handlers so that the last registered handler
	// is the first to be called.
	for i := len(h.pmHandlers) - 1; i >= 0; i-- {
		if h.pmHandlers[i](data) {
			return true
		}
	}
	return false
}

// handleEsc handles an ESC escape sequence.
// It returns true if the sequence was handled.
func (h *handlers) handleEsc(cmd int) bool {
	// Reverse iterate over the handlers so that the last registered handler
	// is the first to be called.
	if handlers, ok := h.escHandler[cmd]; ok {
		for i := len(handlers) - 1; i >= 0; i-- {
			if handlers[i]() {
				return true
			}
		}
	}
	return false
}

// registerDefaultHandlers registers the default escape sequence handlers.
func (t *Terminal) registerDefaultHandlers() {
	t.registerDefaultCcHandlers()
	t.registerDefaultCsiHandlers()
	t.registerDefaultEscHandlers()
	t.registerDefaultOscHandlers()
}

// registerDefaultCcHandlers registers the default control character handlers.
func (t *Terminal) registerDefaultCcHandlers() {
	for i := byte(ansi.NUL); i <= ansi.US; i++ {
		switch i {
		case ansi.NUL: // Null [ansi.NUL]
			// Ignored
			t.registerCcHandler(i, func() bool {
				return true
			})
		case ansi.BEL: // Bell [ansi.BEL]
			t.registerCcHandler(i, func() bool {
				if t.cb.Bell != nil {
					t.cb.Bell()
				}
				return true
			})
		case ansi.BS: // Backspace [ansi.BS]
			t.registerCcHandler(i, func() bool {
				t.backspace()
				return true
			})
		case ansi.HT: // Horizontal Tab [ansi.HT]
			t.registerCcHandler(i, func() bool {
				t.nextTab(1)
				return true
			})
		case ansi.LF, ansi.VT, ansi.FF:
			// Line Feed [ansi.LF]
			// Vertical Tab [ansi.VT]
			// Form Feed [ansi.FF]
			t.registerCcHandler(i, func() bool {
				t.linefeed()
				return true
			})
		case ansi.CR: // Carriage Return [ansi.CR]
			t.registerCcHandler(i, func() bool {
				t.carriageReturn()
				return true
			})
		}
	}

	for i := byte(ansi.PAD); i <= byte(ansi.APC); i++ {
		switch i {
		case ansi.HTS: // Horizontal Tab Set [ansi.HTS]
			t.registerCcHandler(i, func() bool {
				t.horizontalTabSet()
				return true
			})
		case ansi.RI: // Reverse Index [ansi.RI]
			t.registerCcHandler(i, func() bool {
				t.reverseIndex()
				return true
			})
		case ansi.SO: // Shift Out [ansi.SO]
			t.registerCcHandler(i, func() bool {
				t.gl = 1
				return true
			})
		case ansi.SI: // Shift In [ansi.SI]
			t.registerCcHandler(i, func() bool {
				t.gl = 0
				return true
			})
		case ansi.IND: // Index [ansi.IND]
			t.registerCcHandler(i, func() bool {
				t.index()
				return true
			})
		case ansi.SS2: // Single Shift 2 [ansi.SS2]
			t.registerCcHandler(i, func() bool {
				t.gsingle = 2
				return true
			})
		case ansi.SS3: // Single Shift 3 [ansi.SS3]
			t.registerCcHandler(i, func() bool {
				t.gsingle = 3
				return true
			})
		}
	}
}

// registerDefaultOscHandlers registers the default OSC escape sequence handlers.
func (t *Terminal) registerDefaultOscHandlers() {
	for _, cmd := range []int{
		0, // Set window title and icon name
		1, // Set icon name
		2, // Set window title
	} {
		t.RegisterOscHandler(cmd, func(data []byte) bool {
			t.handleTitle(cmd, data)
			return true
		})
	}

	t.RegisterOscHandler(7, func(data []byte) bool {
		// Report the shell current working directory
		// [ansi.NotifyWorkingDirectory].
		t.handleWorkingDirectory(7, data)
		return true
	})

	t.RegisterOscHandler(8, func(data []byte) bool {
		// Set/Query Hyperlink [ansi.SetHyperlink]
		t.handleHyperlink(8, data)
		return true
	})

	for _, cmd := range []int{
		10,  // Set/Query foreground color
		11,  // Set/Query background color
		12,  // Set/Query cursor color
		110, // Reset foreground color
		111, // Reset background color
		112, // Reset cursor color
	} {
		t.RegisterOscHandler(cmd, func(data []byte) bool {
			t.handleDefaultColor(cmd, data)
			return true
		})
	}
}

// registerDefaultEscHandlers registers the default ESC escape sequence handlers.
func (t *Terminal) registerDefaultEscHandlers() {
	t.RegisterEscHandler('=', func() bool {
		// Keypad Application Mode [ansi.DECKPAM]
		t.setMode(ansi.NumericKeypadMode, ansi.ModeSet)
		return true
	})

	t.RegisterEscHandler('>', func() bool {
		// Keypad Numeric Mode [ansi.DECKPNM]
		t.setMode(ansi.NumericKeypadMode, ansi.ModeReset)
		return true
	})

	t.RegisterEscHandler('7', func() bool {
		// Save Cursor [ansi.DECSC]
		t.scr.SaveCursor()
		return true
	})

	t.RegisterEscHandler('8', func() bool {
		// Restore Cursor [ansi.DECRC]
		t.scr.RestoreCursor()
		return true
	})

	for _, cmd := range []int{
		ansi.Command(0, '(', 'A'), // UK G0
		ansi.Command(0, ')', 'A'), // UK G1
		ansi.Command(0, '*', 'A'), // UK G2
		ansi.Command(0, '+', 'A'), // UK G3
		ansi.Command(0, '(', 'B'), // USASCII G0
		ansi.Command(0, ')', 'B'), // USASCII G1
		ansi.Command(0, '*', 'B'), // USASCII G2
		ansi.Command(0, '+', 'B'), // USASCII G3
		ansi.Command(0, '(', '0'), // Special G0
		ansi.Command(0, ')', '0'), // Special G1
		ansi.Command(0, '*', '0'), // Special G2
		ansi.Command(0, '+', '0'), // Special G3
	} {
		t.RegisterEscHandler(cmd, func() bool {
			// Select Character Set [ansi.SCS]
			c := ansi.Cmd(cmd)
			set := c.Intermediate() - '('
			switch c.Final() {
			case 'A': // UK Character Set
				t.charsets[set] = UK
			case 'B': // USASCII Character Set
				t.charsets[set] = nil // USASCII is the default
			case '0': // Special Drawing Character Set
				t.charsets[set] = SpecialDrawing
			default:
				return false
			}
			return true
		})
	}

	t.RegisterEscHandler('D', func() bool {
		// Index [ansi.IND]
		t.index()
		return true
	})

	t.RegisterEscHandler('H', func() bool {
		// Horizontal Tab Set [ansi.HTS]
		t.horizontalTabSet()
		return true
	})

	t.RegisterEscHandler('M', func() bool {
		// Reverse Index [ansi.RI]
		t.reverseIndex()
		return true
	})

	t.RegisterEscHandler('c', func() bool {
		// Reset Initial State [ansi.RIS]
		t.fullReset()
		return true
	})

	t.RegisterEscHandler('n', func() bool {
		// Locking Shift G2 [ansi.LS2]
		t.gl = 2
		return true
	})

	t.RegisterEscHandler('o', func() bool {
		// Locking Shift G3 [ansi.LS3]
		t.gl = 3
		return true
	})

	t.RegisterEscHandler('|', func() bool {
		// Locking Shift 3 Right [ansi.LS3R]
		t.gr = 3
		return true
	})

	t.RegisterEscHandler('}', func() bool {
		// Locking Shift 2 Right [ansi.LS2R]
		t.gr = 2
		return true
	})

	t.RegisterEscHandler('~', func() bool {
		// Locking Shift 1 Right [ansi.LS1R]
		t.gr = 1
		return true
	})
}

// registerDefaultCsiHandlers registers the default CSI escape sequence handlers.
func (t *Terminal) registerDefaultCsiHandlers() {
	t.RegisterCsiHandler('@', func(params ansi.Params) bool {
		// Insert Character [ansi.ICH]
		n, _, _ := params.Param(0, 1)
		t.scr.InsertCell(n)
		return true
	})

	t.RegisterCsiHandler('A', func(params ansi.Params) bool {
		// Cursor Up [ansi.CUU]
		n, _, _ := params.Param(0, 1)
		t.moveCursor(0, -n)
		return true
	})

	t.RegisterCsiHandler('B', func(params ansi.Params) bool {
		// Cursor Down [ansi.CUD]
		n, _, _ := params.Param(0, 1)
		t.moveCursor(0, n)
		return true
	})

	t.RegisterCsiHandler('C', func(params ansi.Params) bool {
		// Cursor Forward [ansi.CUF]
		n, _, _ := params.Param(0, 1)
		t.moveCursor(n, 0)
		return true
	})

	t.RegisterCsiHandler('D', func(params ansi.Params) bool {
		// Cursor Backward [ansi.CUB]
		n, _, _ := params.Param(0, 1)
		t.moveCursor(-n, 0)
		return true
	})

	t.RegisterCsiHandler('E', func(params ansi.Params) bool {
		// Cursor Next Line [ansi.CNL]
		n, _, _ := params.Param(0, 1)
		t.moveCursor(0, n)
		t.carriageReturn()
		return true
	})

	t.RegisterCsiHandler('F', func(params ansi.Params) bool {
		// Cursor Previous Line [ansi.CPL]
		n, _, _ := params.Param(0, 1)
		t.moveCursor(0, -n)
		t.carriageReturn()
		return true
	})

	t.RegisterCsiHandler('G', func(params ansi.Params) bool {
		// Cursor Horizontal Absolute [ansi.CHA]
		n, _, _ := params.Param(0, 1)
		_, y := t.scr.CursorPosition()
		t.setCursor(n-1, y)
		return true
	})

	t.RegisterCsiHandler('H', func(params ansi.Params) bool {
		// Cursor Position [ansi.CUP]
		width, height := t.Width(), t.Height()
		row, _, _ := params.Param(0, 1)
		col, _, _ := params.Param(1, 1)
		if row < 1 {
			row = 1
		}
		if col < 1 {
			col = 1
		}
		y := min(height-1, row-1)
		x := min(width-1, col-1)
		t.setCursorPosition(x, y)
		return true
	})

	t.RegisterCsiHandler('I', func(params ansi.Params) bool {
		// Cursor Horizontal Tabulation [ansi.CHT]
		n, _, _ := params.Param(0, 1)
		t.nextTab(n)
		return true
	})

	t.RegisterCsiHandler('J', func(params ansi.Params) bool {
		// Erase in Display [ansi.ED]
		n, _, _ := params.Param(0, 0)
		width, height := t.Width(), t.Height()
		x, y := t.scr.CursorPosition()
		switch n {
		case 0: // Erase screen below (from after cursor position)
			rect1 := uv.Rect(x, y, width, 1)            // cursor to end of line
			rect2 := uv.Rect(0, y+1, width, height-y-1) // next line onwards
			t.scr.FillArea(t.scr.blankCell(), rect1)
			t.scr.FillArea(t.scr.blankCell(), rect2)
		case 1: // Erase screen above (including cursor)
			rect := uv.Rect(0, 0, width, y+1)
			t.scr.FillArea(t.scr.blankCell(), rect)
		case 2: // erase screen
			fallthrough
		case 3: // erase display
			//nolint:godox
			// TODO: Scrollback buffer support?
			t.scr.Clear()
		default:
			return false
		}
		return true
	})

	t.RegisterCsiHandler('K', func(params ansi.Params) bool {
		// Erase in Line [ansi.EL]
		n, _, _ := params.Param(0, 0)
		// NOTE: Erase Line (EL) erases all character attributes but not cell
		// bg color.
		x, y := t.scr.CursorPosition()
		w := t.scr.Width()

		switch n {
		case 0: // Erase from cursor to end of line
			t.eraseCharacter(w - x)
		case 1: // Erase from start of line to cursor
			rect := uv.Rect(0, y, x+1, 1)
			t.scr.FillArea(t.scr.blankCell(), rect)
		case 2: // Erase entire line
			rect := uv.Rect(0, y, w, 1)
			t.scr.FillArea(t.scr.blankCell(), rect)
		default:
			return false
		}
		return true
	})

	t.RegisterCsiHandler('L', func(params ansi.Params) bool {
		// Insert Line [ansi.IL]
		n, _, _ := params.Param(0, 1)
		if t.scr.InsertLine(n) {
			// Move the cursor to the left margin.
			t.scr.setCursorX(0, true)
		}
		return true
	})

	t.RegisterCsiHandler('M', func(params ansi.Params) bool {
		// Delete Line [ansi.DL]
		n, _, _ := params.Param(0, 1)
		if t.scr.DeleteLine(n) {
			// If the line was deleted successfully, move the cursor to the
			// left.
			// Move the cursor to the left margin.
			t.scr.setCursorX(0, true)
		}
		return true
	})

	t.RegisterCsiHandler('P', func(params ansi.Params) bool {
		// Delete Character [ansi.DCH]
		n, _, _ := params.Param(0, 1)
		t.scr.DeleteCell(n)
		return true
	})

	t.RegisterCsiHandler('S', func(params ansi.Params) bool {
		// Scroll Up [ansi.SU]
		n, _, _ := params.Param(0, 1)
		t.scr.ScrollUp(n)
		return true
	})

	t.RegisterCsiHandler('T', func(params ansi.Params) bool {
		// Scroll Down [ansi.SD]
		n, _, _ := params.Param(0, 1)
		t.scr.ScrollDown(n)
		return true
	})

	t.RegisterCsiHandler(ansi.Command('?', 0, 'W'), func(params ansi.Params) bool {
		// Set Tab at Every 8 Columns [ansi.DECST8C]
		if len(params) == 1 && params[0] == 5 {
			t.resetTabStops()
			return true
		}
		return false
	})

	t.RegisterCsiHandler('X', func(params ansi.Params) bool {
		// Erase Character [ansi.ECH]
		n, _, _ := params.Param(0, 1)
		t.eraseCharacter(n)
		return true
	})

	t.RegisterCsiHandler('Z', func(params ansi.Params) bool {
		// Cursor Backward Tabulation [ansi.CBT]
		n, _, _ := params.Param(0, 1)
		t.prevTab(n)
		return true
	})

	t.RegisterCsiHandler('`', func(params ansi.Params) bool {
		// Horizontal Position Absolute [ansi.HPA]
		n, _, _ := params.Param(0, 1)
		width := t.Width()
		_, y := t.scr.CursorPosition()
		t.setCursorPosition(min(width-1, n-1), y)
		return true
	})

	t.RegisterCsiHandler('a', func(params ansi.Params) bool {
		// Horizontal Position Relative [ansi.HPR]
		n, _, _ := params.Param(0, 1)
		width := t.Width()
		x, y := t.scr.CursorPosition()
		t.setCursorPosition(min(width-1, x+n), y)
		return true
	})

	t.RegisterCsiHandler('b', func(params ansi.Params) bool {
		// Repeat Previous Character [ansi.REP]
		n, _, _ := params.Param(0, 1)
		t.repeatPreviousCharacter(n)
		return true
	})

	t.RegisterCsiHandler('c', func(params ansi.Params) bool {
		// Primary Device Attributes [ansi.DA1]
		n, _, _ := params.Param(0, 0)
		if n != 0 {
			return false
		}

		// Do we fully support VT220?
		_, _ = io.WriteString(t.pw, ansi.PrimaryDeviceAttributes(
			62, // VT220
			1,  // 132 columns
			6,  // Selective Erase
			22, // ANSI color
		))
		return true
	})

	t.RegisterCsiHandler(ansi.Command('>', 0, 'c'), func(params ansi.Params) bool {
		// Secondary Device Attributes [ansi.DA2]
		n, _, _ := params.Param(0, 0)
		if n != 0 {
			return false
		}

		// Do we fully support VT220?
		_, _ = io.WriteString(t.pw, ansi.SecondaryDeviceAttributes(
			1,  // VT220
			10, // Version 1.0
			0,  // ROM Cartridge is always zero
		))
		return true
	})

	t.RegisterCsiHandler('d', func(params ansi.Params) bool {
		// Vertical Position Absolute [ansi.VPA]
		n, _, _ := params.Param(0, 1)
		height := t.Height()
		x, _ := t.scr.CursorPosition()
		t.setCursorPosition(x, min(height-1, n-1))
		return true
	})

	t.RegisterCsiHandler('e', func(params ansi.Params) bool {
		// Vertical Position Relative [ansi.VPR]
		n, _, _ := params.Param(0, 1)
		height := t.Height()
		x, y := t.scr.CursorPosition()
		t.setCursorPosition(x, min(height-1, y+n))
		return true
	})

	t.RegisterCsiHandler('f', func(params ansi.Params) bool {
		// Horizontal and Vertical Position [ansi.HVP]
		width, height := t.Width(), t.Height()
		row, _, _ := params.Param(0, 1)
		col, _, _ := params.Param(1, 1)
		y := min(height-1, row-1)
		x := min(width-1, col-1)
		t.setCursor(x, y)
		return true
	})

	t.RegisterCsiHandler('g', func(params ansi.Params) bool {
		// Tab Clear [ansi.TBC]
		value, _, _ := params.Param(0, 0)
		switch value {
		case 0:
			x, _ := t.scr.CursorPosition()
			t.tabstops.Reset(x)
		case 3:
			t.tabstops.Clear()
		default:
			return false
		}

		return true
	})

	t.RegisterCsiHandler('h', func(params ansi.Params) bool {
		// Set Mode [ansi.SM] - ANSI
		t.handleMode(params, true, true)
		return true
	})

	t.RegisterCsiHandler(ansi.Command('?', 0, 'h'), func(params ansi.Params) bool {
		// Set Mode [ansi.SM] - DEC
		t.handleMode(params, true, false)
		return true
	})

	t.RegisterCsiHandler('l', func(params ansi.Params) bool {
		// Reset Mode [ansi.RM] - ANSI
		t.handleMode(params, false, true)
		return true
	})

	t.RegisterCsiHandler(ansi.Command('?', 0, 'l'), func(params ansi.Params) bool {
		// Reset Mode [ansi.RM] - DEC
		t.handleMode(params, false, false)
		return true
	})

	t.RegisterCsiHandler('m', func(params ansi.Params) bool {
		// Select Graphic Rendition [ansi.SGR]
		t.handleSgr(params)
		return true
	})

	t.RegisterCsiHandler('n', func(params ansi.Params) bool {
		// Device Status Report [ansi.DSR]
		n, _, ok := params.Param(0, 1)
		if !ok || n == 0 {
			return false
		}

		switch n {
		case 5: // Operating Status
			// We're always ready ;)
			// See: https://vt100.net/docs/vt510-rm/DSR-OS.html
			_, _ = io.WriteString(t.pw, ansi.DeviceStatusReport(ansi.DECStatusReport(0)))
		case 6: // Cursor Position Report [ansi.CPR]
			x, y := t.scr.CursorPosition()
			_, _ = io.WriteString(t.pw, ansi.CursorPositionReport(x+1, y+1))
		default:
			return false
		}

		return true
	})

	t.RegisterCsiHandler(ansi.Command('?', 0, 'n'), func(params ansi.Params) bool {
		n, _, ok := params.Param(0, 1)
		if !ok || n == 0 {
			return false
		}

		switch n {
		case 6: // Extended Cursor Position Report [ansi.DECXCPR]
			x, y := t.scr.CursorPosition()
			_, _ = io.WriteString(t.pw, ansi.ExtendedCursorPositionReport(x+1, y+1, 0)) // We don't support page numbers //nolint:errcheck
		default:
			return false
		}

		return true
	})

	t.RegisterCsiHandler(ansi.Command(0, '$', 'p'), func(params ansi.Params) bool {
		// Request Mode [ansi.DECRQM] - ANSI
		t.handleRequestMode(params, true)
		return true
	})

	t.RegisterCsiHandler(ansi.Command('?', '$', 'p'), func(params ansi.Params) bool {
		// Request Mode [ansi.DECRQM] - DEC
		t.handleRequestMode(params, false)
		return true
	})

	t.RegisterCsiHandler(ansi.Command(0, ' ', 'q'), func(params ansi.Params) bool {
		// Set Cursor Style [ansi.DECSCUSR]
		n := 1
		if param, _, ok := params.Param(0, 0); ok && param > n {
			n = param
		}
		blink := n == 0 || n%2 == 1
		style := n / 2
		if !blink {
			style--
		}
		t.scr.setCursorStyle(CursorStyle(style), blink)
		return true
	})

	t.RegisterCsiHandler('r', func(params ansi.Params) bool {
		// Set Top and Bottom Margins [ansi.DECSTBM]
		top, _, _ := params.Param(0, 1)
		if top < 1 {
			top = 1
		}

		height := t.Height()
		bottom, _ := t.parser.Param(1, height)
		if bottom < 1 {
			bottom = height
		}

		if top >= bottom {
			return false
		}

		// Rect is [x, y) which means y is exclusive. So the top margin
		// is the top of the screen minus one.
		t.scr.setVerticalMargins(top-1, bottom)

		// Move the cursor to the top-left of the screen or scroll region
		// depending on [ansi.DECOM].
		t.setCursorPosition(0, 0)
		return true
	})

	t.RegisterCsiHandler('s', func(params ansi.Params) bool {
		// Set Left and Right Margins [ansi.DECSLRM]
		// These conflict with each other. When [ansi.DECSLRM] is set, the we
		// set the left and right margins. Otherwise, we save the cursor
		// position.

		if t.isModeSet(ansi.LeftRightMarginMode) {
			// Set Left Right Margins [ansi.DECSLRM]
			left, _, _ := params.Param(0, 1)
			if left < 1 {
				left = 1
			}

			width := t.Width()
			right, _, _ := params.Param(1, width)
			if right < 1 {
				right = width
			}

			if left >= right {
				return false
			}

			t.scr.setHorizontalMargins(left-1, right)

			// Move the cursor to the top-left of the screen or scroll region
			// depending on [ansi.DECOM].
			t.setCursorPosition(0, 0)
		} else {
			// Save Current Cursor Position [ansi.SCOSC]
			t.scr.SaveCursor()
		}

		return true
	})
}
