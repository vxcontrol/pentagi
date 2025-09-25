package vt

import (
	"testing"

	uv "github.com/charmbracelet/ultraviolet"
)

// testLogger wraps a testing.TB to implement the Logger interface.
type testLogger struct {
	t testing.TB
}

// Printf implements the Logger interface.
func (l *testLogger) Printf(format string, v ...any) {
	l.t.Logf(format, v...)
}

// newTestTerminal creates a new test terminal.
func newTestTerminal(t testing.TB, width, height int) *Terminal {
	term := NewTerminal(width, height, nil)
	term.SetLogger(&testLogger{t})
	return term
}

var cases = []struct {
	name  string
	w, h  int
	input []string
	want  []string
	pos   uv.Position
}{
	// Cursor Backward Tabulation [ansi.CBT]
	{
		name: "CBT Left Beyond First Column",
		w:    10, h: 1,
		input: []string{
			"\x1b[?W", // reset tab stops
			"\x1b[10Z",
			"A",
		},
		want: []string{"A         "},
		pos:  uv.Pos(1, 0),
	},
	{
		name: "CBT Left Starting After Tab Stop",
		w:    11, h: 1,
		input: []string{
			"\x1b[?W", // reset tab stops
			"\x1b[1;10H",
			"X",
			"\x1b[Z",
			"A",
		},
		want: []string{"        AX "},
		pos:  uv.Pos(9, 0),
	},
	{
		name: "CBT Left Starting on Tabstop",
		w:    10, h: 1,
		input: []string{
			"\x1b[?W", // reset tab stops
			"\x1b[1;9H",
			"X",
			"\x1b[1;9H",
			"\x1b[Z",
			"A",
		},
		want: []string{"A       X "},
		pos:  uv.Pos(1, 0),
	},
	{
		name: "CBT Left Margin with Origin Mode",
		w:    10, h: 1,
		input: []string{
			"\x1b[1;1H", // move to top left
			"\x1b[2J",   // clear screen
			"\x1b[?W",   // reset tab stops
			"\x1b[?6h",  // origin mode
			"\x1b[?69h", // left margin mode
			"\x1b[3;6s", // scroll region left/right
			"\x1b[1;2H",
			"X",
			"\x1b[Z",
			"A",
		},
		want: []string{"  AX      "},
		pos:  uv.Pos(3, 0),
	},

	// Cursor Horizontal Tabulation [ansi.CHT]
	{
		name: "CHT Right Beyond Last Column",
		w:    10, h: 1,
		input: []string{
			"\x1b[?W",   // reset tab stops
			"\x1b[100I", // move right 100 tab stops
			"A",
		},
		want: []string{"         A"},
		pos:  uv.Pos(9, 0),
	},
	{
		name: "CHT Right From Before Tabstop",
		w:    10, h: 1,
		input: []string{
			"\x1b[?W",   // reset tab stops
			"\x1b[1;2H", // move to column 2
			"A",
			"\x1b[I", // move right one tab stop
			"X",
		},
		want: []string{" A      X "},
		pos:  uv.Pos(9, 0),
	},
	{
		name: "CHT Right Margin",
		w:    10, h: 1,
		input: []string{
			"\x1b[1;1H", // move to top-left
			"\x1b[2J",   // clear screen
			"\x1b[?W",   // reset tab stops
			"\x1b[?69h", // enable left/right margins
			"\x1b[3;6s", // scroll region left/right
			"\x1b[1;1H", // move cursor in region
			"X",
			"\x1b[I", // move right one tab stop
			"A",
		},
		want: []string{"X    A    "},
		pos:  uv.Pos(6, 0),
	},

	// Carriage Return [ansi.CR]
	{
		name: "CR Pending Wrap is Unset",
		w:    10, h: 2,
		input: []string{
			"\x1b[10G", // move to last column
			"A",        // set pending wrap state
			"\r",       // carriage return
			"X",
		},
		want: []string{
			"X        A",
			"          ",
		},
		pos: uv.Pos(1, 0),
	},
	{
		name: "CR Left Margin",
		w:    10, h: 1,
		input: []string{
			"\x1b[1;1H", // move to top-left
			"\x1b[2J",   // clear screen
			"\x1b[?69h", // enable left/right margin mode
			"\x1b[2;5s", // set left/right margin
			"\x1b[4G",   // move to column 4
			"A",
			"\r",
			"X",
		},
		want: []string{" X A      "},
		pos:  uv.Pos(2, 0),
	},
	{
		name: "CR Left of Left Margin",
		w:    10, h: 1,
		input: []string{
			"\x1b[1;1H", // move to top-left
			"\x1b[2J",   // clear screen
			"\x1b[?69h", // enable left/right margin mode
			"\x1b[2;5s", // set left/right margin
			"\x1b[4G",   // move to column 4
			"A",
			"\x1b[1G",
			"\r",
			"X",
		},
		want: []string{"X  A      "},
		pos:  uv.Pos(1, 0),
	},
	{
		name: "CR Left Margin with Origin Mode",
		w:    10, h: 1,
		input: []string{
			"\x1b[1;1H", // move to top-left
			"\x1b[2J",   // clear screen
			"\x1b[?6h",  // enable origin mode
			"\x1b[?69h", // enable left/right margin mode
			"\x1b[2;5s", // set left/right margin
			"\x1b[4G",   // move to column 4
			"A",
			"\x1b[1G",
			"\r",
			"X",
		},
		want: []string{" X A      "},
		pos:  uv.Pos(2, 0),
	},

	// Cursor Backward [ansi.CUB]
	{
		name: "CUB Pending Wrap is Unset",
		w:    10, h: 2,
		input: []string{
			"\x1b[10G", // move to last column
			"A",        // set pending wrap state
			"\x1b[D",   // move back one
			"XYZ",
		},
		want: []string{
			"        XY",
			"Z         ",
		},
		pos: uv.Pos(1, 1),
	},
	{
		name: "CUB Leftmost Boundary with Reverse Wrap Disabled",
		w:    10, h: 2,
		input: []string{
			"\x1b[?45l", // disable reverse wrap
			"A\n",
			"\x1b[10D", // back
			"B",
		},
		want: []string{
			"A         ",
			"B         ",
		},
		pos: uv.Pos(1, 1),
	},
	{
		name: "CUB Reverse Wrap",
		w:    10, h: 2,
		input: []string{
			"\x1b[?7h",  // enable wraparound
			"\x1b[?45h", // enable reverse wrap
			"\x1b[1;1H", // move to top-left
			"\x1b[2J",   // clear screen
			"\x1b[10G",  // move to end of line
			"AB",        // write and wrap
			"\x1b[D",    // move back one
			"X",
		},
		want: []string{
			"         A",
			"X         ",
		},
		pos: uv.Pos(1, 1),
	},

	// Cursor Down [ansi.CUD]
	{
		name: "CUD Cursor Down",
		w:    10, h: 3,
		input: []string{
			"A",
			"\x1b[2B", // cursor down 2 lines
			"X",
		},
		want: []string{
			"A         ",
			"          ",
			" X        ",
		},
		pos: uv.Pos(2, 2),
	},
	{
		name: "CUD Cursor Down Above Bottom Margin",
		w:    10, h: 4,
		input: []string{
			"\x1b[1;1H", // move to top-left
			"\x1b[2J",   // clear screen
			"\n\n\n\n",  // move down 4 lines
			"\x1b[1;3r", // set scrolling region
			"A",
			"\x1b[5B", // cursor down 5 lines
			"X",
		},
		want: []string{
			"A         ",
			"          ",
			" X        ",
			"          ",
		},
		pos: uv.Pos(2, 2),
	},
	{
		name: "CUD Cursor Down Below Bottom Margin",
		w:    10, h: 5,
		input: []string{
			"\x1b[1;1H",  // move to top-left
			"\x1b[2J",    // clear screen
			"\n\n\n\n\n", // move down 5 lines
			"\x1b[1;3r",  // set scrolling region
			"A",
			"\x1b[4;1H", // move below region
			"\x1b[5B",   // cursor down 5 lines
			"X",
		},
		want: []string{
			"A         ",
			"          ",
			"          ",
			"          ",
			"X         ",
		},
		pos: uv.Pos(1, 4),
	},

	// Cursor Position [ansi.CUP]
	{
		name: "CUP Normal Usage",
		w:    10, h: 2,
		input: []string{
			"\x1b[1;1H", // move to top-left
			"\x1b[2J",   // clear screen
			"\x1b[2;3H", // move to row 2, col 3
			"A",
		},
		want: []string{
			"          ",
			"  A       ",
		},
		pos: uv.Pos(3, 1),
	},
	{
		name: "CUP Off the Screen",
		w:    10, h: 3,
		input: []string{
			"\x1b[1;1H",     // move to top-left
			"\x1b[2J",       // clear screen
			"\x1b[500;500H", // move way off screen
			"A",
		},
		want: []string{
			"          ",
			"          ",
			"         A",
		},
		pos: uv.Pos(9, 2),
	},
	{
		name: "CUP Relative to Origin",
		w:    10, h: 2,
		input: []string{
			"\x1b[1;1H", // move to top-left
			"\x1b[2J",   // clear screen
			"\x1b[2;3r", // scroll region top/bottom
			"\x1b[?6h",  // origin mode
			"\x1b[1;1H", // move to top-left
			"X",
		},
		want: []string{
			"          ",
			"X         ",
		},
		pos: uv.Pos(1, 1),
	},
	{
		name: "CUP Relative to Origin with Margins",
		w:    10, h: 2,
		input: []string{
			"\x1b[1;1H", // move to top-left
			"\x1b[2J",   // clear screen
			"\x1b[?69h", // enable left/right margins
			"\x1b[3;5s", // scroll region left/right
			"\x1b[2;3r", // scroll region top/bottom
			"\x1b[?6h",  // origin mode
			"\x1b[1;1H", // move to top-left
			"X",
		},
		want: []string{
			"          ",
			"  X       ",
		},
		pos: uv.Pos(3, 1),
	},
	{
		name: "CUP Limits with Scroll Region and Origin Mode",
		w:    10, h: 3,
		input: []string{
			"\x1b[1;1H",     // move to top-left
			"\x1b[2J",       // clear screen
			"\x1b[?69h",     // enable left/right margins
			"\x1b[3;5s",     // scroll region left/right
			"\x1b[2;3r",     // scroll region top/bottom
			"\x1b[?6h",      // origin mode
			"\x1b[500;500H", // move way off screen
			"X",
		},
		want: []string{
			"          ",
			"          ",
			"    X     ",
		},
		pos: uv.Pos(5, 2),
	},
	{
		name: "CUP Pending Wrap is Unset",
		w:    10, h: 1,
		input: []string{
			"\x1b[10G", // move to last column
			"A",        // set pending wrap state
			"\x1b[1;1H",
			"X",
		},
		want: []string{
			"X        A",
		},
		pos: uv.Pos(1, 0),
	},

	// Cursor Forward [ansi.CUF]
	{
		name: "CUF Pending Wrap is Unset",
		w:    10, h: 2,
		input: []string{
			"\x1b[10G", // move to last column
			"A",        // set pending wrap state
			"\x1b[C",   // move forward one
			"XYZ",
		},
		want: []string{
			"         X",
			"YZ        ",
		},
		pos: uv.Pos(2, 1),
	},
	{
		name: "CUF Rightmost Boundary",
		w:    10, h: 1,
		input: []string{
			"A",
			"\x1b[500C", // forward larger than screen width
			"B",
		},
		want: []string{
			"A        B",
		},
		pos: uv.Pos(9, 0),
	},
	{
		name: "CUF Left of Right Margin",
		w:    10, h: 1,
		input: []string{
			"\x1b[1;1H", // move to top-left
			"\x1b[2J",   // clear screen
			"\x1b[?69h", // enable left/right margins
			"\x1b[3;5s", // scroll region left/right
			"\x1b[1G",   // move to left
			"\x1b[500C", // forward larger than screen width
			"X",
		},
		want: []string{
			"    X     ",
		},
		pos: uv.Pos(5, 0),
	},
	{
		name: "CUF Right of Right Margin",
		w:    10, h: 1,
		input: []string{
			"\x1b[1;1H", // move to top-left
			"\x1b[2J",   // clear screen
			"\x1b[?69h", // enable left/right margins
			"\x1b[3;5s", // scroll region left/right
			"\x1b[6G",   // move to right of margin
			"\x1b[500C", // forward larger than screen width
			"X",
		},
		want: []string{
			"         X",
		},
		pos: uv.Pos(9, 0),
	},

	// Cursor Up [ansi.CUU]
	{
		name: "CUU Normal Usage",
		w:    10, h: 3,
		input: []string{
			"\x1b[1;1H", // move to top-left
			"\x1b[2J",   // clear screen
			"\x1b[3;1H", // move to row 3
			"A",
			"\x1b[2A", // cursor up 2
			"X",
		},
		want: []string{
			" X        ",
			"          ",
			"A         ",
		},
		pos: uv.Pos(2, 0),
	},
	{
		name: "CUU Below Top Margin",
		w:    10, h: 4,
		input: []string{
			"\x1b[1;1H", // move to top-left
			"\x1b[2J",   // clear screen
			"\x1b[2;4r", // set scrolling region
			"\x1b[3;1H", // move to row 3
			"A",
			"\x1b[5A", // cursor up 5
			"X",
		},
		want: []string{
			"          ",
			" X        ",
			"A         ",
			"          ",
		},
		pos: uv.Pos(2, 1),
	},
	{
		name: "CUU Above Top Margin",
		w:    10, h: 5,
		input: []string{
			"\x1b[1;1H", // move to top-left
			"\x1b[2J",   // clear screen
			"\x1b[3;5r", // set scrolling region
			"\x1b[3;1H", // move to row 3
			"A",
			"\x1b[2;1H", // move above region
			"\x1b[5A",   // cursor up 5
			"X",
		},
		want: []string{
			"X         ",
			"          ",
			"A         ",
			"          ",
			"          ",
		},
		pos: uv.Pos(1, 0),
	},

	// Delete Line [ansi.DL]
	{
		name: "DL Simple Delete Line",
		w:    8, h: 3,
		input: []string{
			"\x1b[1;1H", // move to top-left
			"\x1b[2J",   // clear screen
			"ABC\r\n",
			"DEF\r\n",
			"GHI",
			"\x1b[2;2H",
			"\x1b[M",
		},
		want: []string{
			"ABC     ",
			"GHI     ",
			"        ",
		},
		pos: uv.Pos(0, 1),
	},
	{
		name: "DL Cursor Outside Scroll Region",
		w:    8, h: 3,
		input: []string{
			"\x1b[1;1H", // move to top-left
			"\x1b[2J",   // clear screen
			"ABC\r\n",
			"DEF\r\n",
			"GHI",
			"\x1b[3;4r", // scroll region top/bottom
			"\x1b[2;2H",
			"\x1b[M",
		},
		want: []string{
			"ABC     ",
			"DEF     ",
			"GHI     ",
		},
		pos: uv.Pos(1, 1),
	},
	{
		name: "DL With Top/Bottom Scroll Regions",
		w:    8, h: 4,
		input: []string{
			"\x1b[1;1H", // move to top-left
			"\x1b[2J",   // clear screen
			"ABC\r\n",
			"DEF\r\n",
			"GHI\r\n",
			"123",
			"\x1b[1;3r", // scroll region top/bottom
			"\x1b[2;2H",
			"\x1b[M",
		},
		want: []string{
			"ABC     ",
			"GHI     ",
			"        ",
			"123     ",
		},
		pos: uv.Pos(0, 1),
	},
	{
		name: "DL With Left/Right Scroll Regions",
		w:    8, h: 3,
		input: []string{
			"\x1b[1;1H", // move to top-left
			"\x1b[2J",   // clear screen
			"ABC123\r\n",
			"DEF456\r\n",
			"GHI789",
			"\x1b[?69h", // enable left/right margins
			"\x1b[2;4s", // scroll region left/right
			"\x1b[2;2H",
			"\x1b[M",
		},
		want: []string{
			"ABC123  ",
			"DHI756  ",
			"G   89  ",
		},
		pos: uv.Pos(1, 1),
	},

	// Insert Line [ansi.IL]
	{
		name: "IL Simple Insert Line",
		w:    8, h: 4,
		input: []string{
			"\x1b[1;1H", // move to top-left
			"\x1b[2J",   // clear screen
			"ABC\r\n",
			"DEF\r\n",
			"GHI",
			"\x1b[2;2H",
			"\x1b[L",
		},
		want: []string{
			"ABC     ",
			"        ",
			"DEF     ",
			"GHI     ",
		},
		pos: uv.Pos(0, 1),
	},
	{
		name: "IL Cursor Outside Scroll Region",
		w:    8, h: 3,
		input: []string{
			"\x1b[1;1H", // move to top-left
			"\x1b[2J",   // clear screen
			"ABC\r\n",
			"DEF\r\n",
			"GHI",
			"\x1b[3;4r", // scroll region top/bottom
			"\x1b[2;2H",
			"\x1b[L",
		},
		want: []string{
			"ABC     ",
			"DEF     ",
			"GHI     ",
		},
		pos: uv.Pos(1, 1),
	},
	{
		name: "IL With Top/Bottom Scroll Regions",
		w:    8, h: 4,
		input: []string{
			"\x1b[1;1H", // move to top-left
			"\x1b[2J",   // clear screen
			"ABC\r\n",
			"DEF\r\n",
			"GHI\r\n",
			"123",
			"\x1b[1;3r", // scroll region top/bottom
			"\x1b[2;2H",
			"\x1b[L",
		},
		want: []string{
			"ABC     ",
			"        ",
			"DEF     ",
			"123     ",
		},
		pos: uv.Pos(0, 1),
	},
	{
		name: "IL With Left/Right Scroll Regions",
		w:    8, h: 4,
		input: []string{
			"\x1b[1;1H", // move to top-left
			"\x1b[2J",   // clear screen
			"ABC123\r\n",
			"DEF456\r\n",
			"GHI789",
			"\x1b[?69h", // enable left/right margins
			"\x1b[2;4s", // scroll region left/right
			"\x1b[2;2H",
			"\x1b[L",
		},
		want: []string{
			"ABC123  ",
			"D   56  ",
			"GEF489  ",
			" HI7    ",
		},
		pos: uv.Pos(1, 1),
	},

	// Delete Character [ansi.DCH]
	{
		name: "DCH Simple Delete Character",
		w:    8, h: 1,
		input: []string{
			"ABC123",
			"\x1b[3G",
			"\x1b[2P",
		},
		want: []string{"AB23    "},
		pos:  uv.Pos(2, 0),
	},
	{
		name: "DCH with SGR State",
		w:    8, h: 1,
		input: []string{
			"ABC123",
			"\x1b[3G",
			"\x1b[41m",
			"\x1b[2P",
		},
		want: []string{"AB23    "},
		pos:  uv.Pos(2, 0),
	},
	{
		name: "DCH Outside Left/Right Scroll Region",
		w:    8, h: 1,
		input: []string{
			"\x1b[1;1H", // move to top-left
			"\x1b[2J",   // clear screen
			"ABC123",
			"\x1b[?69h", // enable left/right margins
			"\x1b[3;5s", // scroll region left/right
			"\x1b[2G",
			"\x1b[P",
		},
		want: []string{"ABC123  "},
		pos:  uv.Pos(1, 0),
	},
	{
		name: "DCH Inside Left/Right Scroll Region",
		w:    8, h: 1,
		input: []string{
			"\x1b[1;1H", // move to top-left
			"\x1b[2J",   // clear screen
			"ABC123",
			"\x1b[?69h", // enable left/right margins
			"\x1b[3;5s", // scroll region left/right
			"\x1b[4G",
			"\x1b[P",
		},
		want: []string{"ABC2 3  "},
		pos:  uv.Pos(3, 0),
	},
	{
		name: "DCH Split Wide Character",
		w:    10, h: 1,
		input: []string{
			"\x1b[1;1H", // move to top-left
			"\x1b[2J",   // clear screen
			"A橋123",
			"\x1b[3G",
			"\x1b[P",
		},
		want: []string{"A 123     "},
		pos:  uv.Pos(2, 0),
	},

	// Set Top and Bottom Margins [ansi.DECSTBM]
	{
		name: "DECSTBM Full Screen Scroll Up",
		w:    8, h: 4,
		input: []string{
			"\x1b[1;1H", // move to top-left
			"\x1b[2J",   // clear screen
			"ABC\r\n",
			"DEF\r\n",
			"GHI",
			"\x1b[r", // set full screen scroll region
			"\x1b[T", // scroll up
		},
		want: []string{
			"        ",
			"ABC     ",
			"DEF     ",
			"GHI     ",
		},
		pos: uv.Pos(0, 0),
	},
	{
		name: "DECSTBM Top Only Scroll Up",
		w:    8, h: 4,
		input: []string{
			"\x1b[1;1H", // move to top-left
			"\x1b[2J",   // clear screen
			"ABC\r\n",
			"DEF\r\n",
			"GHI",
			"\x1b[2r", // set scroll region from line 2
			"\x1b[T",  // scroll up
		},
		want: []string{
			"ABC     ",
			"        ",
			"DEF     ",
			"GHI     ",
		},
		pos: uv.Pos(0, 0),
	},
	{
		name: "DECSTBM Top and Bottom Scroll Up",
		w:    8, h: 4,
		input: []string{
			"\x1b[1;1H", // move to top-left
			"\x1b[2J",   // clear screen
			"ABC\r\n",
			"DEF\r\n",
			"GHI",
			"\x1b[1;2r", // set scroll region from line 1 to 2
			"\x1b[T",    // scroll up
		},
		want: []string{
			"        ",
			"ABC     ",
			"GHI     ",
			"        ",
		},
		pos: uv.Pos(0, 0),
	},
	{
		name: "DECSTBM Top Equal Bottom Scroll Up",
		w:    8, h: 4,
		input: []string{
			"\x1b[1;1H", // move to top-left
			"\x1b[2J",   // clear screen
			"ABC\r\n",
			"DEF\r\n",
			"GHI",
			"\x1b[2;2r", // set scroll region at line 2 only
			"\x1b[T",    // scroll up
		},
		want: []string{
			"        ",
			"ABC     ",
			"DEF     ",
			"GHI     ",
		},
		pos: uv.Pos(3, 2),
	},

	// Set Left/Right Margins [ansi.DECSLRM]
	{
		name: "DECSLRM Full Screen",
		w:    8, h: 3,
		input: []string{
			"\x1b[1;1H", // move to top-left
			"\x1b[2J",   // clear screen
			"ABC\r\n",
			"DEF\r\n",
			"GHI",
			"\x1b[?69h", // enable left/right margins
			"\x1b[s",    // scroll region left/right
			"\x1b[X",
		},
		want: []string{
			" BC     ",
			"DEF     ",
			"GHI     ",
		},
		pos: uv.Pos(0, 0),
	},
	{
		name: "DECSLRM Left Only",
		w:    8, h: 4,
		input: []string{
			"\x1b[1;1H", // move to top-left
			"\x1b[2J",   // clear screen
			"ABC\r\n",
			"DEF\r\n",
			"GHI",
			"\x1b[?69h", // enable left/right margins
			"\x1b[2s",   // scroll region left/right
			"\x1b[2G",   // move cursor to column 2
			"\x1b[L",
		},
		want: []string{
			"A       ",
			"DBC     ",
			"GEF     ",
			" HI     ",
		},
		pos: uv.Pos(1, 0),
	},
	{
		name: "DECSLRM Left And Right",
		w:    8, h: 4,
		input: []string{
			"\x1b[1;1H", // move to top-left
			"\x1b[2J",   // clear screen
			"ABC\r\n",
			"DEF\r\n",
			"GHI",
			"\x1b[?69h", // enable left/right margins
			"\x1b[1;2s", // scroll region left/right
			"\x1b[2G",   // move cursor to column 2
			"\x1b[L",
		},
		want: []string{
			"  C     ",
			"ABF     ",
			"DEI     ",
			"GH      ",
		},
		pos: uv.Pos(0, 0),
	},
	{
		name: "DECSLRM Left Equal to Right",
		w:    8, h: 3,
		input: []string{
			"\x1b[1;1H", // move to top-left
			"\x1b[2J",   // clear screen
			"ABC\r\n",
			"DEF\r\n",
			"GHI",
			"\x1b[?69h", // enable left/right margins
			"\x1b[2;2s", // scroll region left/right
			"\x1b[X",
		},
		want: []string{
			"ABC     ",
			"DEF     ",
			"GHI     ",
		},
		pos: uv.Pos(3, 2),
	},

	// Erase Character [ansi.ECH]
	{
		name: "ECH Simple Operation",
		w:    8, h: 1,
		input: []string{
			"ABC",
			"\x1b[1G",
			"\x1b[2X",
		},
		want: []string{"  C     "},
		pos:  uv.Pos(0, 0),
	},
	{
		name: "ECH Erasing Beyond Edge of Screen",
		w:    8, h: 1,
		input: []string{
			"\x1b[8G",
			"\x1b[2D",
			"ABC",
			"\x1b[D",
			"\x1b[10X",
		},
		want: []string{"     A  "},
		pos:  uv.Pos(6, 0),
	},
	{
		name: "ECH Reset Pending Wrap State",
		w:    8, h: 1,
		input: []string{
			"\x1b[8G", // move to last column
			"A",       // set pending wrap state
			"\x1b[X",  // erase one char
			"X",       // write X
		},
		want: []string{"       X"},
		pos:  uv.Pos(7, 0),
	},
	{
		name: "ECH with SGR State",
		w:    8, h: 1,
		input: []string{
			"ABC",
			"\x1b[1G",
			"\x1b[41m", // set red background
			"\x1b[2X",
		},
		want: []string{"  C     "},
		pos:  uv.Pos(0, 0),
	},
	{
		name: "ECH Multi-cell Character",
		w:    8, h: 1,
		input: []string{
			"橋BC",
			"\x1b[1G",
			"\x1b[X",
			"X",
		},
		want: []string{"X BC    "},
		pos:  uv.Pos(1, 0),
	},
	{
		name: "ECH Left/Right Scroll Region Ignored",
		w:    10, h: 1,
		input: []string{
			"\x1b[1;1H", // move to top-left
			"\x1b[2J",   // clear screen
			"\x1b[?69h", // enable left/right margins
			"\x1b[1;3s", // scroll region left/right
			"\x1b[4G",
			"ABC",
			"\x1b[1G",
			"\x1b[4X",
		},
		want: []string{"    BC    "},
		pos:  uv.Pos(0, 0),
	},
	// XXX: Support DECSCA
	// {
	// 	name: "ECH Protected Attributes Ignored with DECSCA",
	// 	w:    8, h: 1,
	// 	input: []string{
	// 		"\x1bV",
	// 		"ABC",
	// 		"\x1b[1\"q",
	// 		"\x1b[0\"q",
	// 		"\x1b[1G",
	// 		"\x1b[2X",
	// 	},
	// 	want: []string{"  C     "},
	// 	pos:  uv.Pos(0, 0),
	// },
	// {
	// 	name: "ECH Protected Attributes Respected without DECSCA",
	// 	w:    8, h: 1,
	// 	input: []string{
	// 		"\x1b[1\"q",
	// 		"ABC",
	// 		"\x1bV",
	// 		"\x1b[1G",
	// 		"\x1b[2X",
	// 	},
	// 	want: []string{"ABC     "},
	// 	pos:  uv.Pos(0, 0),
	// },

	// Erase Line [ansi.EL]
	{
		name: "EL Simple Erase Right",
		w:    8, h: 1,
		input: []string{
			"ABCDE",
			"\x1b[3G",
			"\x1b[0K",
		},
		want: []string{"AB      "},
		pos:  uv.Pos(2, 0),
	},
	{
		name: "EL Erase Right Resets Pending Wrap",
		w:    8, h: 1,
		input: []string{
			"\x1b[8G", // move to last column
			"A",       // set pending wrap state
			"\x1b[0K", // erase right
			"X",
		},
		want: []string{"       X"},
		pos:  uv.Pos(7, 0),
	},
	{
		name: "EL Erase Right with SGR State",
		w:    8, h: 1,
		input: []string{
			"ABC",
			"\x1b[2G",
			"\x1b[41m", // set red background
			"\x1b[0K",
		},
		want: []string{"A       "},
		pos:  uv.Pos(1, 0),
	},
	{
		name: "EL Erase Right Multi-cell Character",
		w:    8, h: 1,
		input: []string{
			"AB橋DE",
			"\x1b[4G",
			"\x1b[0K",
		},
		want: []string{"AB      "},
		pos:  uv.Pos(3, 0),
	},
	{
		name: "EL Erase Right with Left/Right Margins",
		w:    10, h: 1,
		input: []string{
			"\x1b[1;1H", // move to top-left
			"\x1b[2J",   // clear screen
			"ABCDE",
			"\x1b[?69h", // enable left/right margins
			"\x1b[1;3s", // scroll region left/right
			"\x1b[2G",
			"\x1b[0K",
		},
		want: []string{"A         "},
		pos:  uv.Pos(1, 0),
	},
	{
		name: "EL Simple Erase Left",
		w:    8, h: 1,
		input: []string{
			"ABCDE",
			"\x1b[3G",
			"\x1b[1K",
		},
		want: []string{"   DE   "},
		pos:  uv.Pos(2, 0),
	},
	{
		name: "EL Erase Left with SGR State",
		w:    8, h: 1,
		input: []string{
			"ABC",
			"\x1b[2G",
			"\x1b[41m", // set red background
			"\x1b[1K",
		},
		want: []string{"  C     "},
		pos:  uv.Pos(1, 0),
	},
	{
		name: "EL Erase Left Multi-cell Character",
		w:    8, h: 1,
		input: []string{
			"AB橋DE",
			"\x1b[3G",
			"\x1b[1K",
		},
		want: []string{"    DE  "},
		pos:  uv.Pos(2, 0),
	},
	// XXX: Support DECSCA
	// {
	// 	name: "EL Erase Left Protected Attributes Ignored with DECSCA",
	// 	w:    8, h: 1,
	// 	input: []string{
	// 		"\x1bV",
	// 		"ABCDE",
	// 		"\x1b[1\"q",
	// 		"\x1b[0\"q",
	// 		"\x1b[2G",
	// 		"\x1b[1K",
	// 	},
	// 	want: []string{"  CDE   "},
	// 	pos:  uv.Pos(1, 0),
	// },
	{
		name: "EL Simple Erase Complete Line",
		w:    8, h: 1,
		input: []string{
			"ABCDE",
			"\x1b[3G",
			"\x1b[2K",
		},
		want: []string{"        "},
		pos:  uv.Pos(2, 0),
	},
	{
		name: "EL Erase Complete with SGR State",
		w:    8, h: 1,
		input: []string{
			"ABC",
			"\x1b[2G",
			"\x1b[41m", // set red background
			"\x1b[2K",
		},
		want: []string{"        "},
		pos:  uv.Pos(1, 0),
	},

	// Index [ansi.IND]
	{
		name: "IND No Scroll Region Top of Screen",
		w:    10, h: 2,
		input: []string{
			"\x1b[1;1H", // move to top-left
			"\x1b[2J",   // clear screen
			"A",
			"\x1bD", // index
			"X",
		},
		want: []string{
			"A         ",
			" X        ",
		},
		pos: uv.Pos(2, 1),
	},
	{
		name: "IND Bottom of Primary Screen",
		w:    10, h: 2,
		input: []string{
			"\x1b[1;1H", // move to top-left
			"\x1b[2J",   // clear screen
			"\x1b[2;1H", // move to bottom-left
			"A",
			"\x1bD", // index
			"X",
		},
		want: []string{
			"A         ",
			" X        ",
		},
		pos: uv.Pos(2, 1),
	},
	{
		name: "IND Inside Scroll Region",
		w:    10, h: 2,
		input: []string{
			"\x1b[1;1H", // move to top-left
			"\x1b[2J",   // clear screen
			"\x1b[1;3r", // scroll region
			"A",
			"\x1bD", // index
			"X",
		},
		want: []string{
			"A         ",
			" X        ",
		},
		pos: uv.Pos(2, 1),
	},
	{
		name: "IND Bottom of Scroll Region",
		w:    10, h: 4,
		input: []string{
			"\x1b[1;1H", // move to top-left
			"\x1b[2J",   // clear screen
			"\x1b[1;3r", // scroll region
			"\x1b[4;1H", // below scroll region
			"B",
			"\x1b[3;1H", // move to last row of region
			"A",
			"\x1bD", // index
			"X",
		},
		want: []string{
			"          ",
			"A         ",
			" X        ",
			"B         ",
		},
		pos: uv.Pos(2, 2),
	},
	{
		name: "IND Bottom of Primary Screen with Scroll Region",
		w:    10, h: 5,
		input: []string{
			"\x1b[1;1H", // move to top-left
			"\x1b[2J",   // clear screen
			"\x1b[1;3r", // scroll region
			"\x1b[3;1H", // move to last row of region
			"A",
			"\x1b[5;1H", // move to bottom-left
			"\x1bD",     // index
			"X",
		},
		want: []string{
			"          ",
			"          ",
			"A         ",
			"          ",
			"X         ",
		},
		pos: uv.Pos(1, 4),
	},
	{
		name: "IND Outside of Left/Right Scroll Region",
		w:    10, h: 3,
		input: []string{
			"\x1b[1;1H", // move to top-left
			"\x1b[2J",   // clear screen
			"\x1b[?69h", // enable left/right margins
			"\x1b[1;3r", // scroll region top/bottom
			"\x1b[3;5s", // scroll region left/right
			"\x1b[3;3H",
			"A",
			"\x1b[3;1H",
			"\x1bD", // index
			"X",
		},
		want: []string{
			"          ",
			"          ",
			"X A       ",
		},
		pos: uv.Pos(1, 2),
	},
	{
		name: "IND Inside of Left/Right Scroll Region",
		w:    10, h: 3,
		input: []string{
			"\x1b[1;1H", // move to top-left
			"\x1b[2J",   // clear screen
			"AAAAAA\r\n",
			"AAAAAA\r\n",
			"AAAAAA",
			"\x1b[?69h", // enable left/right margins
			"\x1b[1;3s", // set scroll region left/right
			"\x1b[1;3r", // set scroll region top/bottom
			"\x1b[3;1H", // Move to bottom left
			"\x1bD",     // index
		},
		want: []string{
			"AAAAAA    ",
			"AAAAAA    ",
			"   AAA    ",
		},
		pos: uv.Pos(0, 2),
	},

	// Erase Display [ansi.ED]
	{
		name: "ED Simple Erase Below",
		w:    8, h: 3,
		input: []string{
			"\x1b[1;1H", // move to top-left
			"\x1b[2J",   // clear screen
			"ABC\r\n",
			"DEF\r\n",
			"GHI",
			"\x1b[2;2H",
			"\x1b[0J",
		},
		want: []string{
			"ABC     ",
			"D       ",
			"        ",
		},
		pos: uv.Pos(1, 1),
	},
	{
		name: "ED Erase Below with SGR State",
		w:    8, h: 3,
		input: []string{
			"\x1b[1;1H", // move to top-left
			"\x1b[0J",   // clear screen
			"ABC\r\n",
			"DEF\r\n",
			"GHI",
			"\x1b[2;2H",
			"\x1b[41m", // set red background
			"\x1b[0J",
		},
		want: []string{
			"ABC     ",
			"D       ",
			"        ",
		},
		pos: uv.Pos(1, 1),
	},
	{
		name: "ED Erase Below with Multi-Cell Character",
		w:    8, h: 3,
		input: []string{
			"\x1b[1;1H", // move to top-left
			"\x1b[2J",   // clear screen
			"AB橋C\r\n",
			"DE橋F\r\n",
			"GH橋I",
			"\x1b[2;3H", // move to 2nd row 3rd column
			"\x1b[0J",
		},
		want: []string{
			"AB橋C   ",
			"DE      ",
			"        ",
		},
		pos: uv.Pos(2, 1),
	},
	{
		name: "ED Simple Erase Above",
		w:    8, h: 3,
		input: []string{
			"\x1b[1;1H", // move to top-left
			"\x1b[2J",   // clear screen
			"ABC\r\n",
			"DEF\r\n",
			"GHI",
			"\x1b[2;2H",
			"\x1b[1J",
		},
		want: []string{
			"        ",
			"        ",
			"GHI     ",
		},
		pos: uv.Pos(1, 1),
	},
	{
		name: "ED Simple Erase Complete",
		w:    8, h: 3,
		input: []string{
			"\x1b[1;1H", // move to top-left
			"\x1b[2J",   // clear screen
			"ABC\r\n",
			"DEF\r\n",
			"GHI",
			"\x1b[2;2H",
			"\x1b[2J",
		},
		want: []string{
			"        ",
			"        ",
			"        ",
		},
		pos: uv.Pos(1, 1),
	},

	// Reverse Index [ansi.RI]
	{
		name: "RI No Scroll Region Top of Screen",
		w:    10, h: 4,
		input: []string{
			"\x1b[1;1H", // move to top-left
			"\x1b[2J",   // clear screen
			"A\r\n",
			"B\r\n",
			"C\r\n",
			"\x1b[1;1H", // move to top-left
			"\x1bM",     // reverse index
			"X",
		},
		want: []string{
			"X         ",
			"A         ",
			"B         ",
			"C         ",
		},
		pos: uv.Pos(1, 0),
	},
	{
		name: "RI No Scroll Region Not Top of Screen",
		w:    10, h: 3,
		input: []string{
			"\x1b[1;1H", // move to top-left
			"\x1b[2J",   // clear screen
			"A\r\n",
			"B\r\n",
			"C",
			"\x1b[2;1H",
			"\x1bM", // reverse index
			"X",
		},
		want: []string{
			"X         ",
			"B         ",
			"C         ",
		},
		pos: uv.Pos(1, 0),
	},
	{
		name: "RI Top/Bottom Scroll Region",
		w:    10, h: 3,
		input: []string{
			"\x1b[1;1H", // move to top-left
			"\x1b[2J",   // clear screen
			"A\r\n",
			"B\r\n",
			"C",
			"\x1b[2;3r", // scroll region
			"\x1b[2;1H",
			"\x1bM", // reverse index
			"X",
		},
		want: []string{
			"A         ",
			"X         ",
			"B         ",
		},
		pos: uv.Pos(1, 1),
	},
	{
		name: "RI Outside of Top/Bottom Scroll Region",
		w:    10, h: 3,
		input: []string{
			"\x1b[1;1H", // move to top-left
			"\x1b[2J",   // clear screen
			"A\r\n",
			"B\r\n",
			"C",
			"\x1b[2;3r", // scroll region
			"\x1b[1;1H",
			"\x1bM", // reverse index
		},
		want: []string{
			"A         ",
			"B         ",
			"C         ",
		},
		pos: uv.Pos(0, 0),
	},
	{
		name: "RI Left/Right Scroll Region",
		w:    10, h: 4,
		input: []string{
			"\x1b[1;1H", // move to top-left
			"\x1b[2J",   // clear screen
			"ABC\r\n",
			"DEF\r\n",
			"GHI",
			"\x1b[?69h", // enable left/right margins
			"\x1b[2;3s", // scroll region left/right
			"\x1b[1;2H",
			"\x1bM",
		},
		want: []string{
			"A         ",
			"DBC       ",
			"GEF       ",
			" HI       ",
		},
		pos: uv.Pos(1, 0),
	},
	{
		name: "RI Outside Left/Right Scroll Region",
		w:    10, h: 3,
		input: []string{
			"\x1b[1;1H", // move to top-left
			"\x1b[2J",   // clear screen
			"ABC\r\n",
			"DEF\r\n",
			"GHI",
			"\x1b[?69h", // enable left/right margins
			"\x1b[2;3s", // scroll region left/right
			"\x1b[2;1H",
			"\x1bM",
		},
		want: []string{
			"ABC       ",
			"DEF       ",
			"GHI       ",
		},
		pos: uv.Pos(0, 0),
	},

	// Scroll Down [ansi.SD]
	{
		name: "SD Outside of Top/Bottom Scroll Region",
		w:    10, h: 4,
		input: []string{
			"\x1b[1;1H", // move to top-left
			"\x1b[2J",   // clear screen
			"ABC\r\n",
			"DEF\r\n",
			"GHI",
			"\x1b[3;4r", // scroll region top/bottom
			"\x1b[2;2H", // move cursor outside region
			"\x1b[T",    // scroll down
		},
		want: []string{
			"ABC       ",
			"DEF       ",
			"          ",
			"GHI       ",
		},
		pos: uv.Pos(1, 1),
	},

	// Scroll Up [ansi.SU]
	{
		name: "SU Simple Usage",
		w:    10, h: 3,
		input: []string{
			"\x1b[1;1H", // move to top-left
			"\x1b[2J",   // clear screen
			"ABC\r\n",
			"DEF\r\n",
			"GHI",
			"\x1b[2;2H",
			"\x1b[S",
		},
		want: []string{
			"DEF       ",
			"GHI       ",
			"          ",
		},
		pos: uv.Pos(1, 1),
	},
	{
		name: "SU Top/Bottom Scroll Region",
		w:    10, h: 3,
		input: []string{
			"\x1b[1;1H", // move to top-left
			"\x1b[2J",   // clear screen
			"ABC\r\n",
			"DEF\r\n",
			"GHI",
			"\x1b[2;3r", // scroll region top/bottom
			"\x1b[1;1H",
			"\x1b[S",
		},
		want: []string{
			"ABC       ",
			"GHI       ",
			"          ",
		},
		pos: uv.Pos(0, 0),
	},
	{
		name: "SU Left/Right Scroll Regions",
		w:    10, h: 3,
		input: []string{
			"\x1b[1;1H", // move to top-left
			"\x1b[2J",   // clear screen
			"ABC123\r\n",
			"DEF456\r\n",
			"GHI789",
			"\x1b[?69h", // enable left/right margins
			"\x1b[2;4s", // scroll region left/right
			"\x1b[2;2H",
			"\x1b[S",
		},
		want: []string{
			"AEF423    ",
			"DHI756    ",
			"G   89    ",
		},
		pos: uv.Pos(1, 1),
	},
	{
		name: "SU Preserves Pending Wrap",
		w:    10, h: 4,
		input: []string{
			"\x1b[1;10H", // move to top-right
			"\x1b[2J",    // clear screen
			"A",
			"\x1b[2;10H",
			"B",
			"\x1b[3;10H",
			"C",
			"\x1b[S",
			"X",
		},
		want: []string{
			"         B",
			"         C",
			"          ",
			"X         ",
		},
		pos: uv.Pos(1, 3),
	},
	{
		name: "SU Scroll Full Top/Bottom Scroll Region",
		w:    10, h: 5,
		input: []string{
			"\x1b[1;1H", // move to top-left
			"\x1b[2J",   // clear screen
			"top",
			"\x1b[5;1H",
			"ABCDEF",
			"\x1b[2;5r", // scroll region top/bottom
			"\x1b[4S",
		},
		want: []string{
			"top       ",
			"          ",
			"          ",
			"          ",
			"          ",
		},
		pos: uv.Pos(0, 0),
	},

	// Tab Clear [ansi.TBC]
	{
		name: "TBC Clear Single Tab Stop",
		w:    23, h: 1,
		input: []string{
			"\x1b[1;1H", // move to top-left
			"\x1b[2J",   // clear screen
			"\x1b[?W",   // reset tabs
			"\t",        // tab to first stop
			"\x1b[g",    // clear current tab stop
			"\x1b[1G",   // move back to start
			"\t",        // tab again - should go to next stop
		},
		want: []string{"                       "},
		pos:  uv.Pos(16, 0),
	},
	{
		name: "TBC Clear All Tab Stops",
		w:    23, h: 1,
		input: []string{
			"\x1b[1;1H", // move to top-left
			"\x1b[2J",   // clear screen
			"\x1b[?W",   // reset tabs
			"\x1b[3g",   // clear all tab stops
			"\x1b[1G",   // move back to start
			"\t",        // tab - should go to end since no stops
		},
		want: []string{"                       "},
		pos:  uv.Pos(22, 0),
	},
}

// TestTerminal tests the terminal.
func TestTerminal(t *testing.T) {
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			term := newTestTerminal(t, tt.w, tt.h)
			for _, in := range tt.input {
				term.Write([]byte(in))
			}
			got := termText(term)
			if len(got) != len(tt.want) {
				t.Errorf("output length doesn't match: want %d, got %d", len(tt.want), len(got))
			}
			for i := 0; i < len(got) && i < len(tt.want); i++ {
				if got[i] != tt.want[i] {
					t.Errorf("line %d doesn't match:\nwant: %q\ngot:  %q", i+1, tt.want[i], got[i])
				}
			}
			pos := term.CursorPosition()
			if pos != tt.pos {
				t.Errorf("cursor position doesn't match: want %v, got %v", tt.pos, pos)
			}
		})
	}
}

func termText(term *Terminal) []string {
	var lines []string
	for y := range term.Height() {
		var line string
		for x := 0; x < term.Width(); x++ {
			cell := term.CellAt(x, y)
			if cell == nil {
				continue
			}
			line += cell.String()
			x += cell.Width - 1
		}
		lines = append(lines, line)
	}
	return lines
}
