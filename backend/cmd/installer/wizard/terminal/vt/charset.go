package vt

// CharSet represents a character set designator.
// This can be used to select a character set for G0 or G1 and others.
type CharSet map[byte]string

// Character sets.
var (
	UK = CharSet{
		'$': "£", // U+00A3
	}
	SpecialDrawing = CharSet{
		'`': "◆", // U+25C6
		'a': "▒", // U+2592
		'b': "␉", // U+2409
		'c': "␌", // U+240C
		'd': "␍", // U+240D
		'e': "␊", // U+240A
		'f': "°", // U+00B0
		'g': "±", // U+00B1
		'h': "␤", // U+2424
		'i': "␋", // U+240B
		'j': "┘", // U+2518
		'k': "┐", // U+2510
		'l': "┌", // U+250C
		'm': "└", // U+2514
		'n': "┼", // U+253C
		'o': "⎺", // U+23BA
		'p': "⎻", // U+23BB
		'q': "─", // U+2500
		'r': "⎼", // U+23BC
		's': "⎽", // U+23BD
		't': "├", // U+251C
		'u': "┤", // U+2524
		'v': "┴", // U+2534
		'w': "┬", // U+252C
		'x': "│", // U+2502
		'y': "⩽", // U+2A7D
		'z': "⩾", // U+2A7E
		'{': "π", // U+03C0
		'|': "≠", // U+2260
		'}': "£", // U+00A3
		'~': "·", // U+00B7
	}
)
