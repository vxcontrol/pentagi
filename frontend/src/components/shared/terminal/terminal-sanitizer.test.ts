import { describe, expect, it } from 'vitest';

import { needsSanitization, processLog, sanitizeTerminalOutput } from './terminal-sanitizer';

// Helper: check that no C1 control bytes (0x80-0x9F) remain in output
function hasC1(s: string): boolean {
    for (let i = 0; i < s.length; i++) {
        const c = s.charCodeAt(i);

        if (c >= 0x80 && c <= 0x9f) {
            return true;
        }
    }

    return false;
}

// Helper: check that no ESC byte (0x1B) remains in output
function hasEsc(s: string): boolean {
    return s.includes('\x1b');
}

describe('needsSanitization', () => {
    it('returns false for pure ASCII', () => {
        expect(needsSanitization('Hello World 123 !@#')).toBe(false);
    });

    it('returns false for empty/null', () => {
        expect(needsSanitization('')).toBe(false);
        expect(needsSanitization(null as unknown as string)).toBe(false);
    });

    it('returns false for whitespace (TAB, LF, CR)', () => {
        expect(needsSanitization('line1\nline2\ttab\rcarriage')).toBe(false);
    });

    it('returns false for valid Unicode (Cyrillic, CJK, Latin Extended)', () => {
        expect(needsSanitization('Привет мир')).toBe(false);
        expect(needsSanitization('你好世界')).toBe(false);
        expect(needsSanitization('café résumé')).toBe(false);
    });

    it('returns false for box-drawing characters', () => {
        expect(needsSanitization('━━━┃━━━')).toBe(false);
    });

    it('returns false for valid emoji (surrogate pairs)', () => {
        expect(needsSanitization('Hello 🌍 World')).toBe(false);
        expect(needsSanitization('😀😁😂')).toBe(false);
        expect(needsSanitization('🇺🇸')).toBe(false);
    });

    it('returns true for ESC sequences', () => {
        expect(needsSanitization('\x1b[31m')).toBe(true);
        expect(needsSanitization('text\x1b')).toBe(true);
    });

    it('returns true for C0 control characters', () => {
        expect(needsSanitization('\x00')).toBe(true);
        expect(needsSanitization('\x07')).toBe(true);
        expect(needsSanitization('text\x01more')).toBe(true);
    });

    it('returns true for C1 control characters', () => {
        expect(needsSanitization('\x80')).toBe(true);
        expect(needsSanitization('\x90')).toBe(true);
        expect(needsSanitization('\x9b')).toBe(true);
    });

    it('returns true for DEL', () => {
        expect(needsSanitization('\x7f')).toBe(true);
    });

    it('returns true for U+FFFD', () => {
        expect(needsSanitization('\ufffd')).toBe(true);
    });

    it('returns true for lone surrogates', () => {
        expect(needsSanitization('\uD83C')).toBe(true);
        expect(needsSanitization('\uDFFF')).toBe(true);
        expect(needsSanitization('text\uD800')).toBe(true);
    });
});

describe('sanitizeTerminalOutput', () => {
    describe('CSI SGR (colors/styles) — preserved', () => {
        it('preserves SGR red', () => {
            expect(processLog('\x1b[31m')).toBe('\x1b[31m');
        });

        it('preserves SGR reset', () => {
            expect(processLog('\x1b[0m')).toBe('\x1b[0m');
        });

        it('preserves SGR bold+red', () => {
            expect(processLog('\x1b[1;31m')).toBe('\x1b[1;31m');
        });

        it('preserves SGR 256-color', () => {
            expect(processLog('\x1b[38;5;196m')).toBe('\x1b[38;5;196m');
        });

        it('preserves SGR RGB', () => {
            expect(processLog('\x1b[38;2;255;0;0m')).toBe('\x1b[38;2;255;0;0m');
        });

        it('preserves SGR in context with surrounding text', () => {
            const input = 'hello \x1b[31mred\x1b[0m normal';

            expect(processLog(input)).toBe(input);
        });

        it('preserves multiple SGR sequences', () => {
            const input = '\x1b[1m\x1b[31mBOLD RED\x1b[0m';

            expect(processLog(input)).toBe(input);
        });
    });

    describe('CSI non-SGR — blocked', () => {
        it('strips ED: erase display', () => {
            expect(hasEsc(processLog('a\x1b[2Jb'))).toBe(false);
        });

        it('strips ED: erase scrollback', () => {
            expect(hasEsc(processLog('a\x1b[3Jb'))).toBe(false);
        });

        it('strips CUP: cursor home', () => {
            expect(hasEsc(processLog('a\x1b[Hb'))).toBe(false);
        });

        it('strips SU: scroll up', () => {
            expect(hasEsc(processLog('a\x1b[999Sb'))).toBe(false);
        });

        it('strips SD: scroll down', () => {
            expect(hasEsc(processLog('a\x1b[999Tb'))).toBe(false);
        });

        it('strips DECSET: alternate buffer', () => {
            expect(hasEsc(processLog('a\x1b[?1049hb'))).toBe(false);
        });

        it('strips DECSET: hide cursor', () => {
            expect(hasEsc(processLog('a\x1b[?25lb'))).toBe(false);
        });

        it('strips DL: delete lines', () => {
            expect(hasEsc(processLog('a\x1b[10Mb'))).toBe(false);
        });

        it('strips IL: insert lines', () => {
            expect(hasEsc(processLog('a\x1b[10Lb'))).toBe(false);
        });

        it('strips ICH: insert characters', () => {
            expect(hasEsc(processLog('a\x1b[10@b'))).toBe(false);
        });

        it('preserves surrounding text when stripping CSI', () => {
            const out = processLog('before\x1b[2Jafter');

            expect(out).toContain('before');
            expect(out).toContain('after');
        });
    });

    describe('CSI length limit', () => {
        it('strips CSI exceeding MAX_CSI_LENGTH', () => {
            const input = '\x1b[' + '1;'.repeat(1000) + 'm';

            expect(processLog(input)).toMatch(/^\./);
            expect(processLog(input)).not.toContain('\x1b');
        });
    });

    describe('OSC 8 hyperlinks — safe protocols preserved', () => {
        it('preserves https link', () => {
            const input = '\x1b]8;;https://example.com\x07Click\x1b]8;;\x07';

            expect(processLog(input)).toContain('\x1b]8;;https://example.com\x07');
        });

        it('preserves http link', () => {
            const input = '\x1b]8;;http://example.com\x07Click\x1b]8;;\x07';

            expect(processLog(input)).toContain('\x1b]8;;http://');
        });

        it('preserves mailto link', () => {
            const input = '\x1b]8;;mailto:user@example.com\x07Click\x1b]8;;\x07';

            expect(processLog(input)).toContain('\x1b]8;;mailto:');
        });

        it('preserves close tag (BEL terminated)', () => {
            expect(processLog('\x1b]8;;\x07')).toBe('\x1b]8;;\x07');
        });

        it('preserves close tag (ST terminated)', () => {
            expect(processLog('\x1b]8;;\x1b\\')).toBe('\x1b]8;;\x1b\\');
        });

        it('preserves OSC 8 with id param', () => {
            const input = '\x1b]8;id=foo;https://example.com\x07Click\x1b]8;;\x07';

            expect(processLog(input)).toContain('\x1b]8;id=foo;https://example.com');
        });

        it('preserves ssh link', () => {
            const input = '\x1b]8;;ssh://user@host.com\x07Click\x1b]8;;\x07';

            expect(processLog(input)).toContain('\x1b]8;;ssh://');
        });

        it('preserves telnet link', () => {
            const input = '\x1b]8;;telnet://host:23\x07Click\x1b]8;;\x07';

            expect(processLog(input)).toContain('\x1b]8;;telnet://');
        });

        it('preserves full link with ST terminator', () => {
            const input = '\x1b]8;;https://example.com\x1b\\Click\x1b]8;;\x1b\\';

            expect(processLog(input)).toContain('\x1b]8;;https://example.com\x1b\\');
        });

        it('preserves long URL up to MAX_OSC_LENGTH (2048)', () => {
            const longPath = '/path/' + 'a'.repeat(500);
            const input = '\x1b]8;;https://example.com' + longPath + '\x07Click\x1b]8;;\x07';

            expect(processLog(input)).toContain('https://example.com' + longPath);
        });
    });

    describe('OSC 8 hyperlinks — dangerous protocols blocked', () => {
        it('strips javascript: URI', () => {
            const input = '\x1b]8;;javascript:alert(1)\x07Click\x1b]8;;\x07';
            const out = processLog(input);

            expect(out).not.toContain('\x1b]8;;javascript:');
        });

        it('strips data: URI', () => {
            const input = '\x1b]8;;data:text/html,<h1>XSS</h1>\x07Click\x1b]8;;\x07';
            const out = processLog(input);

            expect(out).not.toContain('\x1b]8;;data:');
        });

        it('strips ftp: URI', () => {
            const input = '\x1b]8;;ftp://evil.com/shell\x07Click\x1b]8;;\x07';
            const out = processLog(input);

            expect(out).not.toContain('\x1b]8;;ftp:');
        });

        it('strips file: URI', () => {
            const input = '\x1b]8;;file:///etc/passwd\x07Click\x1b]8;;\x07';
            const out = processLog(input);

            expect(out).not.toContain('\x1b]8;;file:');
        });

        it('preserves safe close tag even when open tag is stripped', () => {
            const input = '\x1b]8;;javascript:x\x07Click\x1b]8;;\x07';
            const out = processLog(input);

            expect(out).toContain('\x1b]8;;\x07');
        });

        it('strips OSC 8 with missing second semicolon', () => {
            const input = '\x1b]8;https://example.com\x07Click\x1b]8;;\x07';
            const out = processLog(input);

            expect(out).not.toContain('\x1b]8;https:');
        });

        it('strips OSC 8 with unknown protocol', () => {
            const input = '\x1b]8;;custom://something\x07Click\x1b]8;;\x07';
            const out = processLog(input);

            expect(out).not.toContain('\x1b]8;;custom:');
        });
    });

    describe('OSC non-hyperlink — blocked', () => {
        it('strips OSC 0: window title', () => {
            expect(hasEsc(processLog('\x1b]0;HACKED\x07'))).toBe(false);
        });

        it('strips OSC 2: window title', () => {
            expect(hasEsc(processLog('\x1b]2;HACKED\x07'))).toBe(false);
        });

        it('strips OSC 10: foreground color', () => {
            expect(hasEsc(processLog('\x1b]10;#ff0000\x07'))).toBe(false);
        });

        it('strips OSC 11: background color', () => {
            expect(hasEsc(processLog('\x1b]11;#00ff00\x07'))).toBe(false);
        });

        it('strips OSC 12: cursor color', () => {
            expect(hasEsc(processLog('\x1b]12;#0000ff\x07'))).toBe(false);
        });

        it('strips OSC 4: palette color', () => {
            expect(hasEsc(processLog('\x1b]4;1;#ff0000\x07'))).toBe(false);
        });
    });

    describe('ESC simple sequences — all blocked', () => {
        it('strips ESC 7 (save cursor)', () => {
            expect(hasEsc(processLog('\x1b7'))).toBe(false);
        });

        it('strips ESC 8 (restore cursor)', () => {
            expect(hasEsc(processLog('\x1b8'))).toBe(false);
        });

        it('strips ESC D (index)', () => {
            expect(hasEsc(processLog('\x1bD'))).toBe(false);
        });

        it('strips ESC E (next line)', () => {
            expect(hasEsc(processLog('\x1bE'))).toBe(false);
        });

        it('strips ESC M (reverse index)', () => {
            expect(hasEsc(processLog('\x1bM'))).toBe(false);
        });

        it('strips ESC _ (APC) with body consumed', () => {
            const out = processLog('\x1b_payload\x1b\\');

            expect(hasEsc(out)).toBe(false);
            expect(out).not.toContain('payload');
        });

        it('strips ESC ^ (PM) with body consumed', () => {
            const out = processLog('\x1b^payload\x1b\\');

            expect(hasEsc(out)).toBe(false);
            expect(out).not.toContain('payload');
        });

        it('strips ESC X (SOS) with body consumed', () => {
            const out = processLog('\x1bXpayload\x1b\\');

            expect(hasEsc(out)).toBe(false);
            expect(out).not.toContain('payload');
        });

        it('strips ESC P (DCS) with body consumed', () => {
            const out = processLog('\x1bP0;1|data\x1b\\');

            expect(hasEsc(out)).toBe(false);
            expect(out).not.toContain('data');
        });

        it('strips ESC ( (charset G0)', () => {
            expect(hasEsc(processLog('\x1b(B'))).toBe(false);
        });

        it('strips truncated ESC at end of input', () => {
            expect(processLog('text\x1b')).toBe('text.');
        });
    });

    describe('C0 control characters', () => {
        it('preserves TAB', () => {
            expect(processLog('a\tb')).toBe('a\tb');
        });

        it('preserves LF', () => {
            expect(processLog('a\nb')).toBe('a\nb');
        });

        it('preserves CR', () => {
            expect(processLog('a\rb')).toBe('a\rb');
        });

        it('strips NUL via binary detection', () => {
            expect(processLog('x\x00y')).toBe('[binary data]');
        });

        it('strips BEL', () => {
            expect(processLog('before\x07after')).toBe('before.after');
        });

        it('strips BS', () => {
            expect(processLog('before\x08after')).toBe('before.after');
        });

        it('strips SOH, STX, ETX', () => {
            expect(processLog('a\x01\x02\x03b')).toBe('a...b');
        });
    });

    describe('C1 control characters (0x80-0x9F)', () => {
        it('strips all C1 bytes', () => {
            const padding = 'A'.repeat(400);
            const c1Chars = String.fromCharCode(...Array.from({ length: 32 }, (_, i) => 0x80 + i));
            const input = padding + c1Chars + 'end';
            const out = processLog(input);

            expect(hasC1(out)).toBe(false);
            expect(out).toContain('end');
        });

        it('strips C1 DCS (0x90)', () => {
            expect(processLog('a\x90b')).toBe('a.b');
        });

        it('strips C1 CSI (0x9B)', () => {
            expect(processLog('a\x9bb')).toBe('a.b');
        });

        it('strips C1 OSC (0x9D)', () => {
            expect(processLog('a\x9db')).toBe('a.b');
        });
    });

    describe('DEL and U+FFFD', () => {
        it('strips DEL (0x7F)', () => {
            expect(processLog('a\x7fb')).toBe('a.b');
        });

        it('strips U+FFFD', () => {
            expect(processLog('a\ufffdb')).toBe('a.b');
        });
    });

    describe('Unicode preservation', () => {
        it('preserves Cyrillic', () => {
            expect(processLog('Привет мир')).toBe('Привет мир');
        });

        it('preserves CJK', () => {
            expect(processLog('你好世界')).toBe('你好世界');
        });

        it('preserves Latin Extended', () => {
            expect(processLog('café résumé')).toBe('café résumé');
        });

        it('preserves emoji', () => {
            expect(processLog('Hello 🌍 World')).toBe('Hello 🌍 World');
        });

        it('preserves box-drawing', () => {
            expect(processLog('━━━┃━━━')).toBe('━━━┃━━━');
        });

        it('preserves mixed Unicode with SGR', () => {
            const input = '\x1b[31mПривет\x1b[0m 🌍';

            expect(processLog(input)).toBe(input);
        });
    });

    describe('binary content detection', () => {
        it('detects null byte as binary', () => {
            expect(processLog('text\x00more')).toBe('[binary data]');
        });

        it('detects lone surrogates as binary', () => {
            const input = 'A'.repeat(100) + '\uD800' + 'B'.repeat(100);

            expect(processLog(input)).toBe('[binary data]');
        });

        it('detects high control char density as binary', () => {
            const input = Array.from({ length: 100 }, (_, i) => String.fromCharCode(i < 15 ? 0x01 + i : 0x41)).join('');

            expect(processLog(input)).toBe('[binary data]');
        });

        it('does not false-positive on short strings with BEL', () => {
            expect(processLog('\x1b]8;;\x07')).not.toBe('[binary data]');
        });

        it('does not false-positive on short strings without null', () => {
            expect(processLog('\x07\x08\x01')).not.toBe('[binary data]');
        });

        it('passes through valid content after 512-byte sample window', () => {
            const input = 'A'.repeat(600) + '\x90';
            const out = processLog(input);

            expect(out).not.toBe('[binary data]');
            expect(hasC1(out)).toBe(false);
        });
    });

    describe('edge cases', () => {
        it('handles empty string', () => {
            expect(sanitizeTerminalOutput('')).toBe('');
        });

        it('handles null', () => {
            expect(sanitizeTerminalOutput(null as unknown as string)).toBe(null);
        });

        it('handles double ESC', () => {
            expect(processLog('\x1b\x1b')).toBe('..');
        });

        it('handles ESC followed by non-printable', () => {
            expect(processLog('\x1b\x01')).toBe('..');
        });

        it('handles ESC [ without final byte', () => {
            const out = processLog('\x1b[');

            expect(out).not.toContain('\x1b');
        });

        it('handles very long input without stack overflow', () => {
            const input = '\x1b[31m' + 'A'.repeat(1_000_000) + '\x1b[0m';
            const out = processLog(input);

            expect(out).toContain('\x1b[31m');
            expect(out.length).toBe(input.length);
        });
    });

    describe('processLog fast path', () => {
        it('skips sanitization for clean ASCII', () => {
            const input = 'Hello World';

            expect(processLog(input)).toBe(input);
        });

        it('skips sanitization for clean Unicode', () => {
            const input = 'Привет мир 🌍';

            expect(processLog(input)).toBe(input);
        });

        it('runs sanitization for strings with ESC', () => {
            const input = '\x1b[31mred\x1b[0m';

            expect(processLog(input)).toBe(input);
        });
    });

    describe('string sequence body consumption (DCS/APC/PM/SOS)', () => {
        it('consumes DCS body: ESC P body ESC \\ → single dot', () => {
            expect(processLog('before\x1bPbody\x1b\\after')).toBe('before.after');
        });

        it('consumes APC body: ESC _ body ESC \\ → single dot', () => {
            expect(processLog('before\x1b_body\x1b\\after')).toBe('before.after');
        });

        it('consumes PM body: ESC ^ body ESC \\ → single dot', () => {
            expect(processLog('before\x1b^body\x1b\\after')).toBe('before.after');
        });

        it('consumes SOS body: ESC X body ESC \\ → single dot', () => {
            expect(processLog('before\x1bXbody\x1b\\after')).toBe('before.after');
        });

        it('consumes DCS with C1 ST terminator (0x9C)', () => {
            expect(processLog('before\x1bPbody\x9cafter')).toBe('before.after');
        });

        it('handles unterminated DCS — consumes until end of input', () => {
            const out = processLog('before\x1bPbody');

            expect(out).toBe('before.');
            expect(out).not.toContain('body');
        });

        it('handles unterminated APC — consumes until end of input', () => {
            const out = processLog('before\x1b_body');

            expect(out).toBe('before.');
            expect(out).not.toContain('body');
        });

        it('consumes DCS with ESC sequences inside body', () => {
            const out = processLog('before\x1bP\x1b[31m\x1b\\after');

            expect(out).toBe('before.after');
        });

        it('consumes long DCS body without hanging', () => {
            const longBody = 'X'.repeat(50_000);
            const out = processLog('a\x1bP' + longBody + '\x1b\\b');

            expect(out).toBe('a.b');
        });
    });

    describe('performance', () => {
        it('processes 1M chars in under 500ms', () => {
            const input = '\x1b[31m' + 'A'.repeat(100_000) + '\x1b[0m';
            const start = performance.now();

            for (let i = 0; i < 100; i++) {
                sanitizeTerminalOutput(input);
            }

            const duration = performance.now() - start;

            expect(duration).toBeLessThan(500);
        });

        it('handles 1M ESC bytes without ReDoS', () => {
            const input = '\x1b'.repeat(1_000_000);
            const start = performance.now();
            sanitizeTerminalOutput(input);
            const duration = performance.now() - start;

            expect(duration).toBeLessThan(1000);
        });
    });
});
