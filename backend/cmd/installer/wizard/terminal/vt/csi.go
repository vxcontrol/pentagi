package vt

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/x/ansi"
)

func (t *Terminal) handleCsi(cmd ansi.Cmd, params ansi.Params) {
	t.flushGrapheme() // Flush any pending grapheme before handling CSI sequences.
	if !t.handlers.handleCsi(cmd, params) {
		t.logf("unhandled sequence: CSI %q", paramsString(cmd, params))
	}
}

func (t *Terminal) handleRequestMode(params ansi.Params, isAnsi bool) {
	n, _, ok := params.Param(0, 0)
	if !ok || n == 0 {
		return
	}

	var mode ansi.Mode = ansi.DECMode(n)
	if isAnsi {
		mode = ansi.ANSIMode(n)
	}

	setting := t.modes[mode]
	_, _ = io.WriteString(t.pw, ansi.ReportMode(mode, setting))
}

func paramsString(cmd ansi.Cmd, params ansi.Params) string {
	var s strings.Builder
	if mark := cmd.Prefix(); mark != 0 {
		s.WriteByte(mark)
	}
	params.ForEach(-1, func(i, p int, more bool) {
		s.WriteString(fmt.Sprintf("%d", p))
		if i < len(params)-1 {
			if more {
				s.WriteByte(':')
			} else {
				s.WriteByte(';')
			}
		}
	})
	if inter := cmd.Intermediate(); inter != 0 {
		s.WriteByte(inter)
	}
	if final := cmd.Final(); final != 0 {
		s.WriteByte(final)
	}
	return s.String()
}
