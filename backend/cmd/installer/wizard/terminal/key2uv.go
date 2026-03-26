package terminal

import (
	"pentagi/cmd/installer/wizard/terminal/vt"

	tea "github.com/charmbracelet/bubbletea"
	uv "github.com/charmbracelet/ultraviolet"
)

// teaKeyToUVKey converts a BubbleTea KeyMsg to an Ultraviolet KeyEvent
func teaKeyToUVKey(msg tea.KeyMsg) uv.KeyEvent {
	key := tea.Key(msg)

	// Build modifiers
	var mod vt.KeyMod
	if key.Alt {
		mod |= vt.ModAlt
	}

	// Map special keys
	switch key.Type {
	case tea.KeyEnter:
		return vt.KeyPressEvent{Code: vt.KeyEnter, Mod: mod}
	case tea.KeyTab:
		return vt.KeyPressEvent{Code: vt.KeyTab, Mod: mod}
	case tea.KeyBackspace:
		return vt.KeyPressEvent{Code: vt.KeyBackspace, Mod: mod}
	case tea.KeyEscape:
		return vt.KeyPressEvent{Code: vt.KeyEscape, Mod: mod}
	case tea.KeySpace:
		return vt.KeyPressEvent{Code: vt.KeySpace, Mod: mod}

	// Arrow keys
	case tea.KeyUp:
		return vt.KeyPressEvent{Code: vt.KeyUp, Mod: mod}
	case tea.KeyDown:
		return vt.KeyPressEvent{Code: vt.KeyDown, Mod: mod}
	case tea.KeyLeft:
		return vt.KeyPressEvent{Code: vt.KeyLeft, Mod: mod}
	case tea.KeyRight:
		return vt.KeyPressEvent{Code: vt.KeyRight, Mod: mod}

	// Navigation keys
	case tea.KeyHome:
		return vt.KeyPressEvent{Code: vt.KeyHome, Mod: mod}
	case tea.KeyEnd:
		return vt.KeyPressEvent{Code: vt.KeyEnd, Mod: mod}
	case tea.KeyPgUp:
		return vt.KeyPressEvent{Code: vt.KeyPgUp, Mod: mod}
	case tea.KeyPgDown:
		return vt.KeyPressEvent{Code: vt.KeyPgDown, Mod: mod}
	case tea.KeyDelete:
		return vt.KeyPressEvent{Code: vt.KeyDelete, Mod: mod}
	case tea.KeyInsert:
		return vt.KeyPressEvent{Code: vt.KeyInsert, Mod: mod}

	// Function keys
	case tea.KeyF1:
		return vt.KeyPressEvent{Code: vt.KeyF1, Mod: mod}
	case tea.KeyF2:
		return vt.KeyPressEvent{Code: vt.KeyF2, Mod: mod}
	case tea.KeyF3:
		return vt.KeyPressEvent{Code: vt.KeyF3, Mod: mod}
	case tea.KeyF4:
		return vt.KeyPressEvent{Code: vt.KeyF4, Mod: mod}
	case tea.KeyF5:
		return vt.KeyPressEvent{Code: vt.KeyF5, Mod: mod}
	case tea.KeyF6:
		return vt.KeyPressEvent{Code: vt.KeyF6, Mod: mod}
	case tea.KeyF7:
		return vt.KeyPressEvent{Code: vt.KeyF7, Mod: mod}
	case tea.KeyF8:
		return vt.KeyPressEvent{Code: vt.KeyF8, Mod: mod}
	case tea.KeyF9:
		return vt.KeyPressEvent{Code: vt.KeyF9, Mod: mod}
	case tea.KeyF10:
		return vt.KeyPressEvent{Code: vt.KeyF10, Mod: mod}
	case tea.KeyF11:
		return vt.KeyPressEvent{Code: vt.KeyF11, Mod: mod}
	case tea.KeyF12:
		return vt.KeyPressEvent{Code: vt.KeyF12, Mod: mod}
	case tea.KeyF13:
		return vt.KeyPressEvent{Code: vt.KeyF13, Mod: mod}
	case tea.KeyF14:
		return vt.KeyPressEvent{Code: vt.KeyF14, Mod: mod}
	case tea.KeyF15:
		return vt.KeyPressEvent{Code: vt.KeyF15, Mod: mod}
	case tea.KeyF16:
		return vt.KeyPressEvent{Code: vt.KeyF16, Mod: mod}
	case tea.KeyF17:
		return vt.KeyPressEvent{Code: vt.KeyF17, Mod: mod}
	case tea.KeyF18:
		return vt.KeyPressEvent{Code: vt.KeyF18, Mod: mod}
	case tea.KeyF19:
		return vt.KeyPressEvent{Code: vt.KeyF19, Mod: mod}
	case tea.KeyF20:
		return vt.KeyPressEvent{Code: vt.KeyF20, Mod: mod}

	// Control keys - map to ASCII control codes
	case tea.KeyCtrlA:
		return vt.KeyPressEvent{Code: 'a', Mod: mod | vt.ModCtrl}
	case tea.KeyCtrlB:
		return vt.KeyPressEvent{Code: 'b', Mod: mod | vt.ModCtrl}
	case tea.KeyCtrlC:
		return vt.KeyPressEvent{Code: 'c', Mod: mod | vt.ModCtrl}
	case tea.KeyCtrlD:
		return vt.KeyPressEvent{Code: 'd', Mod: mod | vt.ModCtrl}
	case tea.KeyCtrlE:
		return vt.KeyPressEvent{Code: 'e', Mod: mod | vt.ModCtrl}
	case tea.KeyCtrlF:
		return vt.KeyPressEvent{Code: 'f', Mod: mod | vt.ModCtrl}
	case tea.KeyCtrlG:
		return vt.KeyPressEvent{Code: 'g', Mod: mod | vt.ModCtrl}
	case tea.KeyCtrlH:
		return vt.KeyPressEvent{Code: 'h', Mod: mod | vt.ModCtrl}
	// tea.KeyCtrlI == tea.KeyTab, handled above
	case tea.KeyCtrlJ:
		return vt.KeyPressEvent{Code: 'j', Mod: mod | vt.ModCtrl}
	case tea.KeyCtrlK:
		return vt.KeyPressEvent{Code: 'k', Mod: mod | vt.ModCtrl}
	case tea.KeyCtrlL:
		return vt.KeyPressEvent{Code: 'l', Mod: mod | vt.ModCtrl}
	// tea.KeyCtrlM == tea.KeyEnter, handled above
	case tea.KeyCtrlN:
		return vt.KeyPressEvent{Code: 'n', Mod: mod | vt.ModCtrl}
	case tea.KeyCtrlO:
		return vt.KeyPressEvent{Code: 'o', Mod: mod | vt.ModCtrl}
	case tea.KeyCtrlP:
		return vt.KeyPressEvent{Code: 'p', Mod: mod | vt.ModCtrl}
	case tea.KeyCtrlQ:
		return vt.KeyPressEvent{Code: 'q', Mod: mod | vt.ModCtrl}
	case tea.KeyCtrlR:
		return vt.KeyPressEvent{Code: 'r', Mod: mod | vt.ModCtrl}
	case tea.KeyCtrlS:
		return vt.KeyPressEvent{Code: 's', Mod: mod | vt.ModCtrl}
	case tea.KeyCtrlT:
		return vt.KeyPressEvent{Code: 't', Mod: mod | vt.ModCtrl}
	case tea.KeyCtrlU:
		return vt.KeyPressEvent{Code: 'u', Mod: mod | vt.ModCtrl}
	case tea.KeyCtrlV:
		return vt.KeyPressEvent{Code: 'v', Mod: mod | vt.ModCtrl}
	case tea.KeyCtrlW:
		return vt.KeyPressEvent{Code: 'w', Mod: mod | vt.ModCtrl}
	case tea.KeyCtrlX:
		return vt.KeyPressEvent{Code: 'x', Mod: mod | vt.ModCtrl}
	case tea.KeyCtrlY:
		return vt.KeyPressEvent{Code: 'y', Mod: mod | vt.ModCtrl}
	case tea.KeyCtrlZ:
		return vt.KeyPressEvent{Code: 'z', Mod: mod | vt.ModCtrl}

	// Shift+Tab
	case tea.KeyShiftTab:
		return vt.KeyPressEvent{Code: vt.KeyTab, Mod: mod | vt.ModShift}

	// Arrow keys with modifiers
	case tea.KeyShiftUp:
		return vt.KeyPressEvent{Code: vt.KeyUp, Mod: mod | vt.ModShift}
	case tea.KeyShiftDown:
		return vt.KeyPressEvent{Code: vt.KeyDown, Mod: mod | vt.ModShift}
	case tea.KeyShiftLeft:
		return vt.KeyPressEvent{Code: vt.KeyLeft, Mod: mod | vt.ModShift}
	case tea.KeyShiftRight:
		return vt.KeyPressEvent{Code: vt.KeyRight, Mod: mod | vt.ModShift}

	case tea.KeyCtrlUp:
		return vt.KeyPressEvent{Code: vt.KeyUp, Mod: mod | vt.ModCtrl}
	case tea.KeyCtrlDown:
		return vt.KeyPressEvent{Code: vt.KeyDown, Mod: mod | vt.ModCtrl}
	case tea.KeyCtrlLeft:
		return vt.KeyPressEvent{Code: vt.KeyLeft, Mod: mod | vt.ModCtrl}
	case tea.KeyCtrlRight:
		return vt.KeyPressEvent{Code: vt.KeyRight, Mod: mod | vt.ModCtrl}

	case tea.KeyCtrlShiftUp:
		return vt.KeyPressEvent{Code: vt.KeyUp, Mod: mod | vt.ModCtrl | vt.ModShift}
	case tea.KeyCtrlShiftDown:
		return vt.KeyPressEvent{Code: vt.KeyDown, Mod: mod | vt.ModCtrl | vt.ModShift}
	case tea.KeyCtrlShiftLeft:
		return vt.KeyPressEvent{Code: vt.KeyLeft, Mod: mod | vt.ModCtrl | vt.ModShift}
	case tea.KeyCtrlShiftRight:
		return vt.KeyPressEvent{Code: vt.KeyRight, Mod: mod | vt.ModCtrl | vt.ModShift}

	// Home/End with modifiers
	case tea.KeyShiftHome:
		return vt.KeyPressEvent{Code: vt.KeyHome, Mod: mod | vt.ModShift}
	case tea.KeyShiftEnd:
		return vt.KeyPressEvent{Code: vt.KeyEnd, Mod: mod | vt.ModShift}
	case tea.KeyCtrlHome:
		return vt.KeyPressEvent{Code: vt.KeyHome, Mod: mod | vt.ModCtrl}
	case tea.KeyCtrlEnd:
		return vt.KeyPressEvent{Code: vt.KeyEnd, Mod: mod | vt.ModCtrl}
	case tea.KeyCtrlShiftHome:
		return vt.KeyPressEvent{Code: vt.KeyHome, Mod: mod | vt.ModCtrl | vt.ModShift}
	case tea.KeyCtrlShiftEnd:
		return vt.KeyPressEvent{Code: vt.KeyEnd, Mod: mod | vt.ModCtrl | vt.ModShift}

	// Page Up/Down with modifiers
	case tea.KeyCtrlPgUp:
		return vt.KeyPressEvent{Code: vt.KeyPgUp, Mod: mod | vt.ModCtrl}
	case tea.KeyCtrlPgDown:
		return vt.KeyPressEvent{Code: vt.KeyPgDown, Mod: mod | vt.ModCtrl}

	// Handle regular character input (runes)
	case tea.KeyRunes:
		if len(key.Runes) > 0 {
			return vt.KeyPressEvent{Code: key.Runes[0], Mod: mod}
		}
		return nil

	default:
		// For any unmapped keys, return nil
		return nil
	}
}
