package vt

import (
	"io"

	uv "github.com/charmbracelet/ultraviolet"
	"github.com/charmbracelet/x/ansi"
)

// KeyMod represents a key modifier.
type KeyMod = uv.KeyMod

// Modifier keys.
const (
	ModShift = uv.ModShift
	ModAlt   = uv.ModAlt
	ModCtrl  = uv.ModCtrl
	ModMeta  = uv.ModMeta
)

// KeyPressEvent represents a key press event.
type KeyPressEvent = uv.KeyPressEvent

// SendKey returns the default key map.
func (t *Terminal) SendKey(k uv.KeyEvent) {
	var seq string

	ack := t.isModeSet(ansi.CursorKeysMode)    // Application cursor keys mode
	akk := t.isModeSet(ansi.NumericKeypadMode) // Application keypad keys mode

	//nolint:godox
	// TODO: Support Kitty, CSI u, and XTerm modifyOtherKeys.
	switch key := k.(type) {
	case KeyPressEvent:
		if key.Mod&ModAlt != 0 {
			// Handle alt-modified keys
			seq = "\x1b" + seq
			key.Mod &^= ModAlt // Remove the Alt modifier for easier matching
		}

		//nolint:godox
		// FIXME: We remove any Base and Shifted codes to properly handle
		// comparison. This is a workaround for the fact that we don't support
		// extended keys yet.
		key.BaseCode = 0
		key.ShiftedCode = 0

		switch key {
		// Control keys
		case KeyPressEvent{Code: KeySpace, Mod: ModCtrl}:
			seq += "\x00"
		case KeyPressEvent{Code: 'a', Mod: ModCtrl}:
			seq += "\x01"
		case KeyPressEvent{Code: 'b', Mod: ModCtrl}:
			seq += "\x02"
		case KeyPressEvent{Code: 'c', Mod: ModCtrl}:
			seq += "\x03"
		case KeyPressEvent{Code: 'd', Mod: ModCtrl}:
			seq += "\x04"
		case KeyPressEvent{Code: 'e', Mod: ModCtrl}:
			seq += "\x05"
		case KeyPressEvent{Code: 'f', Mod: ModCtrl}:
			seq += "\x06"
		case KeyPressEvent{Code: 'g', Mod: ModCtrl}:
			seq += "\x07"
		case KeyPressEvent{Code: 'h', Mod: ModCtrl}:
			seq += "\x08"
		case KeyPressEvent{Code: 'i', Mod: ModCtrl}:
			seq += "\x09"
		case KeyPressEvent{Code: 'j', Mod: ModCtrl}:
			seq += "\x0a"
		case KeyPressEvent{Code: 'k', Mod: ModCtrl}:
			seq += "\x0b"
		case KeyPressEvent{Code: 'l', Mod: ModCtrl}:
			seq += "\x0c"
		case KeyPressEvent{Code: 'm', Mod: ModCtrl}:
			seq += "\x0d"
		case KeyPressEvent{Code: 'n', Mod: ModCtrl}:
			seq += "\x0e"
		case KeyPressEvent{Code: 'o', Mod: ModCtrl}:
			seq += "\x0f"
		case KeyPressEvent{Code: 'p', Mod: ModCtrl}:
			seq += "\x10"
		case KeyPressEvent{Code: 'q', Mod: ModCtrl}:
			seq += "\x11"
		case KeyPressEvent{Code: 'r', Mod: ModCtrl}:
			seq += "\x12"
		case KeyPressEvent{Code: 's', Mod: ModCtrl}:
			seq += "\x13"
		case KeyPressEvent{Code: 't', Mod: ModCtrl}:
			seq += "\x14"
		case KeyPressEvent{Code: 'u', Mod: ModCtrl}:
			seq += "\x15"
		case KeyPressEvent{Code: 'v', Mod: ModCtrl}:
			seq += "\x16"
		case KeyPressEvent{Code: 'w', Mod: ModCtrl}:
			seq += "\x17"
		case KeyPressEvent{Code: 'x', Mod: ModCtrl}:
			seq += "\x18"
		case KeyPressEvent{Code: 'y', Mod: ModCtrl}:
			seq += "\x19"
		case KeyPressEvent{Code: 'z', Mod: ModCtrl}:
			seq += "\x1a"
		case KeyPressEvent{Code: '[', Mod: ModCtrl}:
			seq += "\x1b"
		case KeyPressEvent{Code: '\\', Mod: ModCtrl}:
			seq += "\x1c"
		case KeyPressEvent{Code: ']', Mod: ModCtrl}:
			seq += "\x1d"
		case KeyPressEvent{Code: '^', Mod: ModCtrl}:
			seq += "\x1e"
		case KeyPressEvent{Code: '_', Mod: ModCtrl}:
			seq += "\x1f"

		case KeyPressEvent{Code: KeyEnter}:
			seq += "\r"
		case KeyPressEvent{Code: KeyTab}:
			seq += "\t"
		case KeyPressEvent{Code: KeyBackspace}:
			seq += "\x7f"
		case KeyPressEvent{Code: KeyEscape}:
			seq += "\x1b"

		case KeyPressEvent{Code: KeyUp}:
			if ack {
				seq += "\x1bOA"
			} else {
				seq += "\x1b[A"
			}
		case KeyPressEvent{Code: KeyDown}:
			if ack {
				seq += "\x1bOB"
			} else {
				seq += "\x1b[B"
			}
		case KeyPressEvent{Code: KeyRight}:
			if ack {
				seq += "\x1bOC"
			} else {
				seq += "\x1b[C"
			}
		case KeyPressEvent{Code: KeyLeft}:
			if ack {
				seq += "\x1bOD"
			} else {
				seq += "\x1b[D"
			}

		case KeyPressEvent{Code: KeyInsert}:
			seq += "\x1b[2~"
		case KeyPressEvent{Code: KeyDelete}:
			seq += "\x1b[3~"
		case KeyPressEvent{Code: KeyHome}:
			seq += "\x1b[H"
		case KeyPressEvent{Code: KeyEnd}:
			seq += "\x1b[F"
		case KeyPressEvent{Code: KeyPgUp}:
			seq += "\x1b[5~"
		case KeyPressEvent{Code: KeyPgDown}:
			seq += "\x1b[6~"

		case KeyPressEvent{Code: KeyF1}:
			seq += "\x1bOP"
		case KeyPressEvent{Code: KeyF2}:
			seq += "\x1bOQ"
		case KeyPressEvent{Code: KeyF3}:
			seq += "\x1bOR"
		case KeyPressEvent{Code: KeyF4}:
			seq += "\x1bOS"
		case KeyPressEvent{Code: KeyF5}:
			seq += "\x1b[15~"
		case KeyPressEvent{Code: KeyF6}:
			seq += "\x1b[17~"
		case KeyPressEvent{Code: KeyF7}:
			seq += "\x1b[18~"
		case KeyPressEvent{Code: KeyF8}:
			seq += "\x1b[19~"
		case KeyPressEvent{Code: KeyF9}:
			seq += "\x1b[20~"
		case KeyPressEvent{Code: KeyF10}:
			seq += "\x1b[21~"
		case KeyPressEvent{Code: KeyF11}:
			seq += "\x1b[23~"
		case KeyPressEvent{Code: KeyF12}:
			seq += "\x1b[24~"

		case KeyPressEvent{Code: KeyKp0}:
			if akk {
				seq += "\x1bOp"
			} else {
				seq += "0"
			}
		case KeyPressEvent{Code: KeyKp1}:
			if akk {
				seq += "\x1bOq"
			} else {
				seq += "1"
			}
		case KeyPressEvent{Code: KeyKp2}:
			if akk {
				seq += "\x1bOr"
			} else {
				seq += "2"
			}
		case KeyPressEvent{Code: KeyKp3}:
			if akk {
				seq += "\x1bOs"
			} else {
				seq += "3"
			}
		case KeyPressEvent{Code: KeyKp4}:
			if akk {
				seq += "\x1bOt"
			} else {
				seq += "4"
			}
		case KeyPressEvent{Code: KeyKp5}:
			if akk {
				seq += "\x1bOu"
			} else {
				seq += "5"
			}
		case KeyPressEvent{Code: KeyKp6}:
			if akk {
				seq += "\x1bOv"
			} else {
				seq += "6"
			}
		case KeyPressEvent{Code: KeyKp7}:
			if akk {
				seq += "\x1bOw"
			} else {
				seq += "7"
			}
		case KeyPressEvent{Code: KeyKp8}:
			if akk {
				seq += "\x1bOx"
			} else {
				seq = "8"
			}
		case KeyPressEvent{Code: KeyKp9}:
			if akk {
				seq += "\x1bOy"
			} else {
				seq += "9"
			}
		case KeyPressEvent{Code: KeyKpEnter}:
			if akk {
				seq += "\x1bOM"
			} else {
				seq += "\r"
			}
		case KeyPressEvent{Code: KeyKpEqual}:
			if akk {
				seq += "\x1bOX"
			} else {
				seq += "="
			}
		case KeyPressEvent{Code: KeyKpMultiply}:
			if akk {
				seq += "\x1bOj"
			} else {
				seq += "*"
			}
		case KeyPressEvent{Code: KeyKpPlus}:
			if akk {
				seq += "\x1bOk"
			} else {
				seq += "+"
			}
		case KeyPressEvent{Code: KeyKpComma}:
			if akk {
				seq += "\x1bOl"
			} else {
				seq += ","
			}
		case KeyPressEvent{Code: KeyKpMinus}:
			if akk {
				seq += "\x1bOm"
			} else {
				seq += "-"
			}
		case KeyPressEvent{Code: KeyKpDecimal}:
			if akk {
				seq += "\x1bOn"
			} else {
				seq += "."
			}

		case KeyPressEvent{Code: KeyTab, Mod: ModShift}:
			seq += "\x1b[Z"

		default:
			// Handle the rest of the keys.
			if key.Mod == 0 {
				seq += string(key.Code)
			}
		}

		io.WriteString(t.pw, seq) //nolint:errcheck,gosec
	}
}

// Key codes.
const (
	KeyExtended         = uv.KeyExtended
	KeyUp               = uv.KeyUp
	KeyDown             = uv.KeyDown
	KeyRight            = uv.KeyRight
	KeyLeft             = uv.KeyLeft
	KeyBegin            = uv.KeyBegin
	KeyFind             = uv.KeyFind
	KeyInsert           = uv.KeyInsert
	KeyDelete           = uv.KeyDelete
	KeySelect           = uv.KeySelect
	KeyPgUp             = uv.KeyPgUp
	KeyPgDown           = uv.KeyPgDown
	KeyHome             = uv.KeyHome
	KeyEnd              = uv.KeyEnd
	KeyKpEnter          = uv.KeyKpEnter
	KeyKpEqual          = uv.KeyKpEqual
	KeyKpMultiply       = uv.KeyKpMultiply
	KeyKpPlus           = uv.KeyKpPlus
	KeyKpComma          = uv.KeyKpComma
	KeyKpMinus          = uv.KeyKpMinus
	KeyKpDecimal        = uv.KeyKpDecimal
	KeyKpDivide         = uv.KeyKpDivide
	KeyKp0              = uv.KeyKp0
	KeyKp1              = uv.KeyKp1
	KeyKp2              = uv.KeyKp2
	KeyKp3              = uv.KeyKp3
	KeyKp4              = uv.KeyKp4
	KeyKp5              = uv.KeyKp5
	KeyKp6              = uv.KeyKp6
	KeyKp7              = uv.KeyKp7
	KeyKp8              = uv.KeyKp8
	KeyKp9              = uv.KeyKp9
	KeyKpSep            = uv.KeyKpSep
	KeyKpUp             = uv.KeyKpUp
	KeyKpDown           = uv.KeyKpDown
	KeyKpLeft           = uv.KeyKpLeft
	KeyKpRight          = uv.KeyKpRight
	KeyKpPgUp           = uv.KeyKpPgUp
	KeyKpPgDown         = uv.KeyKpPgDown
	KeyKpHome           = uv.KeyKpHome
	KeyKpEnd            = uv.KeyKpEnd
	KeyKpInsert         = uv.KeyKpInsert
	KeyKpDelete         = uv.KeyKpDelete
	KeyKpBegin          = uv.KeyKpBegin
	KeyF1               = uv.KeyF1
	KeyF2               = uv.KeyF2
	KeyF3               = uv.KeyF3
	KeyF4               = uv.KeyF4
	KeyF5               = uv.KeyF5
	KeyF6               = uv.KeyF6
	KeyF7               = uv.KeyF7
	KeyF8               = uv.KeyF8
	KeyF9               = uv.KeyF9
	KeyF10              = uv.KeyF10
	KeyF11              = uv.KeyF11
	KeyF12              = uv.KeyF12
	KeyF13              = uv.KeyF13
	KeyF14              = uv.KeyF14
	KeyF15              = uv.KeyF15
	KeyF16              = uv.KeyF16
	KeyF17              = uv.KeyF17
	KeyF18              = uv.KeyF18
	KeyF19              = uv.KeyF19
	KeyF20              = uv.KeyF20
	KeyF21              = uv.KeyF21
	KeyF22              = uv.KeyF22
	KeyF23              = uv.KeyF23
	KeyF24              = uv.KeyF24
	KeyF25              = uv.KeyF25
	KeyF26              = uv.KeyF26
	KeyF27              = uv.KeyF27
	KeyF28              = uv.KeyF28
	KeyF29              = uv.KeyF29
	KeyF30              = uv.KeyF30
	KeyF31              = uv.KeyF31
	KeyF32              = uv.KeyF32
	KeyF33              = uv.KeyF33
	KeyF34              = uv.KeyF34
	KeyF35              = uv.KeyF35
	KeyF36              = uv.KeyF36
	KeyF37              = uv.KeyF37
	KeyF38              = uv.KeyF38
	KeyF39              = uv.KeyF39
	KeyF40              = uv.KeyF40
	KeyF41              = uv.KeyF41
	KeyF42              = uv.KeyF42
	KeyF43              = uv.KeyF43
	KeyF44              = uv.KeyF44
	KeyF45              = uv.KeyF45
	KeyF46              = uv.KeyF46
	KeyF47              = uv.KeyF47
	KeyF48              = uv.KeyF48
	KeyF49              = uv.KeyF49
	KeyF50              = uv.KeyF50
	KeyF51              = uv.KeyF51
	KeyF52              = uv.KeyF52
	KeyF53              = uv.KeyF53
	KeyF54              = uv.KeyF54
	KeyF55              = uv.KeyF55
	KeyF56              = uv.KeyF56
	KeyF57              = uv.KeyF57
	KeyF58              = uv.KeyF58
	KeyF59              = uv.KeyF59
	KeyF60              = uv.KeyF60
	KeyF61              = uv.KeyF61
	KeyF62              = uv.KeyF62
	KeyF63              = uv.KeyF63
	KeyCapsLock         = uv.KeyCapsLock
	KeyScrollLock       = uv.KeyScrollLock
	KeyNumLock          = uv.KeyNumLock
	KeyPrintScreen      = uv.KeyPrintScreen
	KeyPause            = uv.KeyPause
	KeyMenu             = uv.KeyMenu
	KeyMediaPlay        = uv.KeyMediaPlay
	KeyMediaPause       = uv.KeyMediaPause
	KeyMediaPlayPause   = uv.KeyMediaPlayPause
	KeyMediaReverse     = uv.KeyMediaReverse
	KeyMediaStop        = uv.KeyMediaStop
	KeyMediaFastForward = uv.KeyMediaFastForward
	KeyMediaRewind      = uv.KeyMediaRewind
	KeyMediaNext        = uv.KeyMediaNext
	KeyMediaPrev        = uv.KeyMediaPrev
	KeyMediaRecord      = uv.KeyMediaRecord
	KeyLowerVol         = uv.KeyLowerVol
	KeyRaiseVol         = uv.KeyRaiseVol
	KeyMute             = uv.KeyMute
	KeyLeftShift        = uv.KeyLeftShift
	KeyLeftAlt          = uv.KeyLeftAlt
	KeyLeftCtrl         = uv.KeyLeftCtrl
	KeyLeftSuper        = uv.KeyLeftSuper
	KeyLeftHyper        = uv.KeyLeftHyper
	KeyLeftMeta         = uv.KeyLeftMeta
	KeyRightShift       = uv.KeyRightShift
	KeyRightAlt         = uv.KeyRightAlt
	KeyRightCtrl        = uv.KeyRightCtrl
	KeyRightSuper       = uv.KeyRightSuper
	KeyRightHyper       = uv.KeyRightHyper
	KeyRightMeta        = uv.KeyRightMeta
	KeyIsoLevel3Shift   = uv.KeyIsoLevel3Shift
	KeyIsoLevel5Shift   = uv.KeyIsoLevel5Shift
	KeyBackspace        = uv.KeyBackspace
	KeyTab              = uv.KeyTab
	KeyEnter            = uv.KeyEnter
	KeyReturn           = uv.KeyReturn
	KeyEscape           = uv.KeyEscape
	KeyEsc              = uv.KeyEsc
	KeySpace            = uv.KeySpace
)
