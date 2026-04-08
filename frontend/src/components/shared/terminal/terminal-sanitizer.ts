const BINARY_SAMPLE_SIZE = 512;
const BINARY_CONTROL_RATIO = 0.1;
const MAX_CSI_LENGTH = 256;
const MAX_OSC_LENGTH = 2048;
const MAX_STRING_SEQUENCE_LENGTH = 100_000;

export const SAFE_PROTOCOLS = ['http://', 'https://', 'mailto:', 'ssh://', 'telnet://'];

/**
 * Quick scan to determine if a string needs full sanitization.
 * Returns false for strings that are purely ASCII printable + whitespace + valid Unicode.
 */
export function needsSanitization(input: string): boolean {
    if (!input) {
        return false;
    }

    for (let i = 0; i < input.length; i++) {
        const code = input.charCodeAt(i);

        if (code === 0x09 || code === 0x0a || code === 0x0d) {
            continue;
        }

        if (code >= 0x20 && code <= 0x7e) {
            continue;
        }

        if (code >= 0xa0) {
            if (code >= 0xd800 && code <= 0xdbff) {
                if (i + 1 < input.length) {
                    const next = input.charCodeAt(i + 1);

                    if (next >= 0xdc00 && next <= 0xdfff) {
                        i++;

                        continue;
                    }
                }

                return true;
            }

            if ((code >= 0xdc00 && code <= 0xdfff) || code === 0xfffd) {
                return true;
            }

            if (isBidiControl(code)) {
                return true;
            }

            continue;
        }

        return true;
    }

    return false;
}

export function processLog(log: string): string {
    return needsSanitization(log) ? sanitizeTerminalOutput(log) : log;
}

/**
 * Sanitizes terminal output for safe display in a read-only xterm.js terminal.
 *
 * Uses a slice-based approach: tracks ranges of safe characters and extracts
 * them via input.slice() instead of pushing individual characters, reducing
 * allocations from millions to thousands on large inputs.
 *
 * Preserves:
 *  - Valid Unicode text (Cyrillic, CJK, Latin Extended, emoji, etc.)
 *  - CSI SGR sequences (text color/style: ESC[...m)
 *  - OSC 8 hyperlinks with safe protocols (http, https, mailto, ssh, telnet)
 *
 * Strips:
 *  - C0/C1 control characters (except TAB, LF, CR)
 *  - All non-SGR CSI sequences (cursor movement, erase, scroll, DECSET)
 *  - OSC color/title/palette changes (OSC 0, 4, 10, 11, 12)
 *  - OSC 8 links with dangerous protocols (javascript:, data:, etc.)
 *  - All simple ESC sequences (cursor save/restore, index, charset)
 *  - DCS, APC, PM, SOS string sequences (body consumed until ST)
 *  - Binary content (replaced with placeholder)
 */
export function sanitizeTerminalOutput(input: string): string {
    if (!input) {
        return input;
    }

    if (isBinaryContent(input)) {
        return '[binary data]';
    }

    const result: string[] = [];
    let i = 0;
    let safeStart = 0;

    while (i < input.length) {
        const code = input.charCodeAt(i);

        // Fast path: ASCII printable (0x20-0x7E) and safe whitespace
        if ((code >= 0x20 && code <= 0x7e) || code === 0x09 || code === 0x0a || code === 0x0d) {
            i++;

            continue;
        }

        // Valid Unicode (>= 0xA0, excluding lone surrogates, U+FFFD, and bidi controls)
        if (code >= 0xa0) {
            // Valid surrogate pair (emoji, supplementary chars) — skip both units
            if (code >= 0xd800 && code <= 0xdbff) {
                if (i + 1 < input.length) {
                    const next = input.charCodeAt(i + 1);

                    if (next >= 0xdc00 && next <= 0xdfff) {
                        i += 2;

                        continue;
                    }
                }

                // Lone/invalid surrogate — falls through to replacement below
            } else if (code !== 0xfffd && !(code >= 0xdc00 && code <= 0xdfff) && !isBidiControl(code)) {
                i++;

                continue;
            }

            // Lone surrogates, U+FFFD, and bidi controls fall through to replacement below
        }

        // We hit a character that needs handling — flush the safe range
        if (i > safeStart) {
            result.push(input.slice(safeStart, i));
        }

        if (code === 0x1b) {
            i++;

            if (i >= input.length) {
                result.push('.');
                safeStart = i;

                continue;
            }

            const next = input.charCodeAt(i);

            // CSI sequence (ESC [) — only allow SGR (text style/color)
            if (next === 0x5b) {
                const parsed = parseCsiSequence(input, i + 1);

                if (parsed.end !== -1 && parsed.final === 'm') {
                    result.push(input.slice(i - 1, parsed.end));
                    i = parsed.end;
                    safeStart = i;

                    continue;
                }

                if (parsed.end !== -1) {
                    i = parsed.end;
                }

                result.push('.');
                safeStart = i;

                continue;
            }

            // OSC sequence (ESC ]) — only allow safe OSC 8 hyperlinks
            if (next === 0x5d) {
                const end = parseOscSequence(input, i + 1);

                if (end !== -1) {
                    const termLen = input.charCodeAt(end - 1) === 0x07 ? 1 : 2;
                    const oscContent = input.slice(i + 1, end - termLen);

                    if (isOscSafe(oscContent)) {
                        result.push(input.slice(i - 1, end));
                        i = end;
                        safeStart = i;

                        continue;
                    }

                    i = end;
                }

                result.push('.');
                safeStart = i;

                continue;
            }

            // DCS (ESC P), SOS (ESC X), PM (ESC ^), APC (ESC _) —
            // consume the entire body up to ST terminator
            if (next === 0x50 || next === 0x58 || next === 0x5e || next === 0x5f) {
                i = skipToST(input, i + 1);
                result.push('.');
                safeStart = i;

                continue;
            }

            // All other ESC sequences (cursor save/restore, index, charset, etc.)
            result.push('.');
            safeStart = i;

            continue;
        }

        // C0 controls, DEL, C1 controls, lone surrogates, U+FFFD
        result.push('.');
        i++;
        safeStart = i;
    }

    if (safeStart < input.length) {
        result.push(input.slice(safeStart));
    }

    return result.join('');
}

function isBidiControl(code: number): boolean {
    return (
        (code >= 0x200e && code <= 0x200f) || (code >= 0x202a && code <= 0x202e) || (code >= 0x2066 && code <= 0x2069)
    );
}

/**
 * Heuristic detection of binary content in a JS string.
 * Binary data (e.g. JPEG) interpreted as UTF-16 produces null bytes,
 * lone surrogates, and dense clusters of control characters.
 *
 * For short strings (< 32 chars), only checks for null bytes to avoid
 * false positives on terminal sequences containing BEL or other C0 codes.
 */
function isBinaryContent(input: string): boolean {
    // Short strings: only null byte is a reliable binary indicator
    if (input.length < 32) {
        return input.includes('\x00');
    }

    let controlCount = 0;
    const sampleSize = Math.min(input.length, BINARY_SAMPLE_SIZE);

    for (let i = 0; i < sampleSize; i++) {
        const code = input.charCodeAt(i);

        // Null byte — immediate binary indicator
        if (code === 0x00) {
            return true;
        }

        // Lone surrogates indicate corrupted/binary data
        if (code >= 0xd800 && code <= 0xdfff) {
            if (code <= 0xdbff && i + 1 < input.length) {
                const next = input.charCodeAt(i + 1);

                // Valid surrogate pair — skip
                if (next >= 0xdc00 && next <= 0xdfff) {
                    i++;

                    continue;
                }
            }

            return true;
        }

        // Count C0 controls (except TAB, LF, CR, ESC which are common in terminal data)
        if (code < 0x20 && code !== 0x09 && code !== 0x0a && code !== 0x0d && code !== 0x1b) {
            controlCount++;
        }

        // Count C1 controls (0x80-0x9F)
        if (code >= 0x80 && code <= 0x9f) {
            controlCount++;
        }
    }

    // High density of control characters indicates binary content
    return controlCount / sampleSize > BINARY_CONTROL_RATIO;
}

/**
 * Checks if a parsed OSC sequence content is safe for display.
 * Only allows OSC 8 hyperlinks with safe protocols (http, https, mailto).
 * Blocks OSC 0 (title), OSC 4/10/11/12 (color changes), and unsafe URIs.
 *
 * OSC 8 format: "8;params;uri" where params is optional key=value pairs.
 * Empty URI (close tag "8;;") is always safe.
 */
function isOscSafe(content: string): boolean {
    // Must be OSC 8 (hyperlink)
    if (!content.startsWith('8;')) {
        return false;
    }

    // Find the URI after the second semicolon: "8;params;uri"
    const secondSemicolon = content.indexOf(';', 2);

    if (secondSemicolon === -1) {
        return false;
    }

    const uri = content.slice(secondSemicolon + 1);

    // Empty URI = close tag (ESC]8;;BEL) — always safe
    if (!uri) {
        return true;
    }

    // Validate protocol against whitelist (blocks javascript:, data:, ftp:, file:, etc.)
    const uriLower = uri.toLowerCase();

    return SAFE_PROTOCOLS.some((protocol) => uriLower.startsWith(protocol));
}

/**
 * Parses a CSI sequence (ESC [). Returns the end index and final byte,
 * or end=-1 if invalid. Enforces a length limit to prevent unbounded scanning.
 *
 * CSI format: ESC [ [params] [intermediates] final_byte
 * - Parameter bytes: 0x30-0x3F (digits, semicolons, etc.)
 * - Intermediate bytes: 0x20-0x2F (space, !, ", etc.)
 * - Final byte: 0x40-0x7E (letters — 'm' for SGR, 'H' for CUP, etc.)
 */
function parseCsiSequence(input: string, start: number): { end: number; final: string } {
    let i = start;

    while (i < input.length && i - start < MAX_CSI_LENGTH) {
        const code = input.charCodeAt(i);

        if (code < 0x20 || code > 0x7e) {
            return { end: -1, final: '' };
        }

        i++;

        if (code >= 0x40 && code <= 0x7e) {
            return { end: i, final: String.fromCharCode(code) };
        }
    }

    return { end: -1, final: '' };
}

/**
 * Parses an OSC sequence (ESC ]). Terminated by BEL (0x07) or ST (ESC \).
 * Returns end index (after terminator), or -1 if invalid.
 *
 * OSC format: ESC ] content BEL  or  ESC ] content ESC \
 * Allows ASCII printables (0x20-0x7E) and Unicode >= 0xA0 per xterm.js parser
 * (which accepts "any codepoint greater than C1 as printable").
 * Blocks C0 controls (except BEL as terminator), DEL, and C1 controls (0x80-0x9F).
 */
function parseOscSequence(input: string, start: number): number {
    let i = start;

    while (i < input.length && i - start < MAX_OSC_LENGTH) {
        const code = input.charCodeAt(i);

        if (code === 0x07) {
            return i + 1;
        }

        if (code === 0x1b && i + 1 < input.length && input.charAt(i + 1) === '\\') {
            return i + 2;
        }

        if (code < 0x20 && code !== 0x07) {
            return -1;
        }

        if (code === 0x7f || (code >= 0x80 && code <= 0x9f)) {
            return -1;
        }

        i++;
    }

    return -1;
}

/**
 * Skips a string-type sequence body (DCS, SOS, PM, APC) until the ST terminator.
 * ST is either ESC \ (7-bit) or the C1 byte 0x9C (8-bit).
 * Returns the index after ST, or the end of input if unterminated.
 * Enforces a length limit to prevent unbounded scanning.
 */
function skipToST(input: string, start: number): number {
    let i = start;
    const limit = Math.min(input.length, start + MAX_STRING_SEQUENCE_LENGTH);

    while (i < limit) {
        const code = input.charCodeAt(i);

        if (code === 0x9c) {
            return i + 1;
        }

        if (code === 0x1b && i + 1 < input.length && input.charCodeAt(i + 1) === 0x5c) {
            return i + 2;
        }

        i++;
    }

    return i;
}
