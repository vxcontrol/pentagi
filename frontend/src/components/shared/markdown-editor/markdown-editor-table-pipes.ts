import { Lexer } from 'marked';

// marked's GFM table tokenizer splits every row on raw `|` BEFORE inline tokenization, so a pipe inside a
// code span (`` `x | y` ``), a Go-template action (`{{.X | upper}}`), or a URL (`[x](http://a|b)`, a bare
// `http://a|b`, an image src) in a cell creates a phantom column and the trailing cells are silently DROPPED
// on load. The splitter does honor `\|` — and unescapes it in the cell text — so pre-escaping those pipes
// before marked lexes protects the load side the same way escapeCellPipes in cellText protects save. The
// Lexer asks for a stand-in character instead of `\|` (see TABLE_PIPE): it is not a pipe to the splitter
// either, it is put back as a pipe wherever it lands, and it tells a pipe that was written bare from one the
// author escaped.
//
// Scope must match EXACTLY the rows marked itself treats as the table — no more, no less:
//   • the header/delimiter pair must already parse as a table (matching cell counts, delimiter has |/:) —
//     escaping pipes in a NON-table line would surface a literal `\|` (the escape tokenizer is neutralised,
//     so `\|` outside a real table row does NOT unescape) and could even turn a setext heading into a table;
//   • the leading/trailing pipe is OPTIONAL in GFM, so the header row AND every body row are transformed
//     (marked's own gfmTable body capture — `(?!blank|hr|heading|blockquote|code|fences|list).*` — has no
//     leading-pipe requirement; gating on a leading `|` dropped cells for the common no-outer-pipe style);
//   • a table's rows run until a blank line or the start of another block (isTableBodyEnd);
//   • fenced code blocks are never touched.

// Capture the full fence run (not a fixed 3) plus the trailing text: renderTunedCodeBlock widens a fence to
// 4+ backticks when its content holds a ``` line, and CommonMark closes a fence only with a run of the same
// char, length >= the opener, and no info string — so length- and char-aware tracking is load-bearing.
// The backtick branch carries marked's own lookahead: a backtick anywhere in the rest of the line means this
// is prose holding an inline code span, not a fence. Without it the scanner opens a fence marked never opened
// and the two run out of step in BOTH directions — tables after it lose pipe protection, and once the phantom
// closes the parity is inverted and pipes get escaped inside real code. Tilde fences carry no such rule.
const FENCE_LINE = /^ {0,3}(`{3,}(?=[^`\n]*$)|~{3,})(.*)$/;
// Written to be linear: the trailing `(?: *\|)? *$` (not `\|? *$`) plus per-cell spacing keep any two space
// runs from competing for the same characters, so a crafted delimiter-looking line can't force O(n²) backtracking.
export const TABLE_DELIMITER_LINE = /^ {0,3}\|? *:?-+:?(?: *\| *:?-+:?)*(?: *\|)? *$/;
// A Go template action. A lone brace inside it — `{{ printf "{%s}" .X }}` — is part of it, and so is `}}`
// inside a string, a rune or a comment. Every character has one reading and none of them runs over `{{`, so a
// scan that fails stops at the next one, which keeps it linear.
export const TEMPLATE_ACTION =
    /\{\{(?:\/\*(?:[^*{]|\*(?!\/)|\{(?!\{))*\*\/|"(?:\\.|[^"\\\n{]|\{(?!\{))*"|`(?:[^`{]|\{(?!\{))*`|'(?:\\.|[^'\\\n{}])'|'(?!(?:\\.|[^'\\\n{}])')|\/(?!\*)|[^{}"`'/]|\{(?!\{)|\}(?!\}))*\}\}/g;
// A line that opens with a control action is the template's own line: `{{ range .Rows }}` and `{{ end }}`
// around the rows of a table are not rows of it, and read as rows they are padded into cells.
const TEMPLATE_CONTROL_LINE = /^ {0,3}\{\{-?[ \t]*(?:range|if|else|end|with|define|template|block|break|continue)\b/;
// A pipe inside a Go action is the template's pipeline operator, never a column separator — marked's splitter
// does not know that, so the scanner masks those pipes before counting. Without this the serializer cannot
// stop escaping them: an unescaped pipeline in the HEADER row would count one cell too many, detection would
// bail, and marked would degrade the whole table to a paragraph.
const maskActionPipes = (row: string): string => row.replace(TEMPLATE_ACTION, (action) => action.replace(/\|/g, ''));
// One blockquote marker (`>` + an optional space or tab, as marked takes it). The line scanner only recognizes top-level tables, but marked
// strips this prefix and re-lexes the inner content, so a table inside a blockquote needs the prefix removed
// before detection — see the recursive handling in escapeTablePipes.
const BLOCKQUOTE_PREFIX = /^ {0,3}>[ \t]?/;
// A list item whose content actually starts on the same line. CommonMark puts that content at the column after
// the marker plus its following spaces, except that 5+ spaces mean the item opens with indented code and the
// content column is one past the marker.
const LIST_ITEM_START = /^( {0,3})([*+-]|\d{1,9}[.)])( +)(?=\S)/;

// A table's body ends at a blank line or the first line that starts a different block. Which lines those are
// is read off marked's own table rule — its body rows are `(?:(?!…).*(?:\n|$))*` — so the two cannot drift
// apart: a standard HTML block tag ends a table too. The writer asks the same question before it puts text
// on the line right under a table. A marked whose rule has another shape leaves only the blank line.
const BODY_ROWS = /\(\?:\\n\(\(\?:\(\?!(.*)\)\.\*\(\?:\\n\|\$\)\)\*\)\\n\*\|\$\)$/;
const ENDS_TABLE_BODY = new RegExp(`^(?:${BODY_ROWS.exec(Lexer.rules.block.gfm.table.source)?.[1] ?? ' *\\n'})`);

// The rule reads a line with the line break that ends it.
export const isTableBodyEnd = (line: string): boolean => ENDS_TABLE_BODY.test(`${line}\n`);

const isEscapedAt = (text: string, offset: number): boolean => {
    let isEscaped = false;
    let index = offset;

    while (--index >= 0 && text[index] === '\\') {
        isEscaped = !isEscaped;
    }

    return isEscaped;
};

// What a protected pipe is written as while marked lexes the document: a private-use character that the Lexer
// turns back into a pipe in every token. A scan that takes a line for a row which marked does not then leaves
// nothing behind, where `\|` would stay in the text as a backslash.
export const TABLE_PIPE = '\uE002';

const escapeUnescapedPipes = (text: string, escaped: string): string =>
    text.replace(/\|/g, (pipe, offset: number, source: string) => (isEscapedAt(source, offset) ? pipe : escaped));

// Mirrors marked's splitCells: split on unescaped pipes, drop a blank leading/trailing cell.
const countCells = (row: string): number => {
    const cells = maskActionPipes(row)
        .replace(/\|/g, (pipe, offset: number, source: string) => (isEscapedAt(source, offset) ? pipe : ' |'))
        .split(/ \|/);

    if (cells.length > 0 && !cells[0]!.trim()) {
        cells.shift();
    }

    if (cells.length > 0 && !cells.at(-1)!.trim()) {
        cells.pop();
    }

    return cells.length;
};

export const findCodeSpanCloser = (row: string, from: number, runLength: number): number => {
    for (let index = from; index < row.length; index++) {
        if (row[index] !== '`') {
            continue;
        }

        let runEnd = index;

        while (runEnd < row.length && row[runEnd] === '`') {
            runEnd++;
        }

        if (runEnd - index === runLength) {
            return index;
        }

        index = runEnd - 1;
    }

    return -1;
};

const escapeRowPipes = (row: string, expectedCells: number, escaped: string): string => {
    let result = '';
    let index = 0;

    while (index < row.length) {
        const char = row[index]!;

        if (char !== '`') {
            result += char;
            index++;
            continue;
        }

        let runEnd = index;

        while (runEnd < row.length && row[runEnd] === '`') {
            runEnd++;
        }

        const run = row.slice(index, runEnd);
        const closerStart = findCodeSpanCloser(row, runEnd, run.length);

        // An unclosed backtick run is literal content, not a code span.
        if (closerStart < 0) {
            result += run;
            index = runEnd;
            continue;
        }

        result += run + escapeUnescapedPipes(row.slice(runEnd, closerStart), escaped) + run;
        index = closerStart + run.length;
    }

    result = result.replace(TEMPLATE_ACTION, (action) => escapeUnescapedPipes(action, escaped));

    // A pipe inside a URL (bare or linked) is content, but a spaceless `|` between two cells is a separator —
    // the two are indistinguishable in isolation. Only escape scheme-run pipes when the row otherwise splits
    // into MORE cells than the header expects; a row already at the right count keeps its structural pipes.
    // This guard also skips the scan on ordinary rows, so a long non-URL token can't drive it superlinearly.
    if (countCells(result) <= expectedCells) {
        return result;
    }

    return result.replace(/\S+/g, (token) => (token.includes('://') ? escapeUnescapedPipes(token, escaped) : token));
};

// A line marked would take for a row and that is not one: the template's own line — it opens with a control
// action, or holds nothing but actions — and a line with more cells than the header, whose extra cells marked
// drops without a trace. The table ends above it and the line is the text it was written as.
export const isNotARow = (line: string, columns: number): boolean => {
    if (TEMPLATE_CONTROL_LINE.test(line)) {
        return true;
    }

    const rest = line.replace(TEMPLATE_ACTION, '');

    return (rest !== line && !rest.trim()) || countCells(escapeRowPipes(line, columns, TABLE_PIPE)) > columns;
};

export const countRowCells = (row: string): number => countCells(row);

// Whether a row reads as its cells only because this scan protects a pipe in it before marked splits it.
export const holdsProtectedPipe = (row: string, columns: number): boolean =>
    escapeRowPipes(row, columns, TABLE_PIPE) !== row;

export const escapeTablePipes = (markdown: string, escaped = '\\|'): string => {
    if (!markdown.includes('|')) {
        return markdown;
    }

    // marked normalizes CRLF itself, but this pre-pass runs BEFORE it: a trailing `\r` left by split('\n')
    // defeats the `$`-anchored TABLE_DELIMITER_LINE, so a CRLF document's tables would go unprotected here and
    // lose cells. Normalize first; the return below keeps the original bytes when nothing was escaped.
    const source = markdown.includes('\r') ? markdown.replace(/\r\n?/g, '\n') : markdown;
    const lines = source.split('\n');
    let openFence: null | { char: string; length: number } = null;
    let isChanged = false;

    const escapeRow = (row: number, expectedCells: number): void => {
        const protectedRow = escapeRowPipes(lines[row]!, expectedCells, escaped);

        if (protectedRow !== lines[row]) {
            lines[row] = protectedRow;
            isChanged = true;
        }
    };

    for (let index = 0; index < lines.length; index++) {
        const line = lines[index]!;
        const fence = FENCE_LINE.exec(line);

        if (fence) {
            const marker = fence[1]!;

            if (openFence === null) {
                openFence = { char: marker[0]!, length: marker.length };
            } else if (marker[0] === openFence.char && marker.length >= openFence.length && !fence[2]!.trim()) {
                openFence = null;
            }

            continue;
        }

        if (openFence !== null) {
            continue;
        }

        // A blockquote (or nested `> >`) hides its table from the top-level scan. Strip the prefix off the whole
        // run, recurse on the inner content (which reuses every rule here, incl. nested blockquotes and fences),
        // then re-apply each line's original prefix — escaping never touches the prefix, so this stays byte-exact.
        if (BLOCKQUOTE_PREFIX.test(line)) {
            let end = index + 1;

            while (end < lines.length && BLOCKQUOTE_PREFIX.test(lines[end]!)) {
                end++;
            }

            const prefixes = lines.slice(index, end).map((row) => BLOCKQUOTE_PREFIX.exec(row)![0]);
            const inner = lines.slice(index, end).map((row, offset) => row.slice(prefixes[offset]!.length));
            const inside = escapeTablePipes(inner.join('\n'), escaped).split('\n');

            for (let row = index; row < end; row++) {
                const rebuilt = prefixes[row - index]! + inside[row - index]!;

                if (rebuilt !== lines[row]) {
                    lines[row] = rebuilt;
                    isChanged = true;
                }
            }

            index = end - 1;

            continue;
        }

        // A list item indents its content to its own column, and once that column reaches 4 the top-level scan
        // reads the table as indented code and skips it — while marked dedents the item and re-lexes, where the
        // table is live. Mirror that: dedent by the content column, recurse, re-indent. The indented-code rule
        // then applies RELATIVE to the item, so a line four columns past the content start stays unprotected —
        // which is why this recurses instead of relaxing TABLE_DELIMITER_LINE's leading-space cap.
        const item = LIST_ITEM_START.exec(line);

        if (item) {
            const [, indent = '', marker = '', spaces = ''] = item;
            const contentColumn = indent.length + marker.length + (spaces.length >= 5 ? 1 : spaces.length);
            let end = index + 1;

            while (end < lines.length && (!lines[end]!.trim() || /^ */.exec(lines[end]!)![0].length >= contentColumn)) {
                end++;
            }

            while (end > index + 1 && !lines[end - 1]!.trim()) {
                end--;
            }

            const run = lines.slice(index, end);

            // The run is the item's, with or without a table in it: walked line by line here, a fence the
            // item opens on its marker line is not seen, the line that closes it opens one in this scan,
            // and every table below would go unprotected.
            if (end > index + 1) {
                if (run.some((row) => row.includes('|'))) {
                    const inside = escapeTablePipes(
                        run.map((row) => row.slice(contentColumn)).join('\n'),
                        escaped,
                    ).split('\n');

                    for (let row = index; row < end; row++) {
                        const rebuilt = lines[row]!.slice(0, contentColumn) + inside[row - index]!;

                        if (rebuilt !== lines[row]) {
                            lines[row] = rebuilt;
                            isChanged = true;
                        }
                    }
                }

                index = end - 1;

                continue;
            }
        }

        // A header may be written without a pipe; the delimiter row under it is what makes the table.
        const delimiter = lines[index + 1];

        if (delimiter === undefined || !TABLE_DELIMITER_LINE.test(delimiter) || !/[|:]/.test(delimiter)) {
            continue;
        }

        const expected = countCells(delimiter);

        if (countCells(line) !== expected) {
            continue;
        }

        // Header + every body row up to the block boundary; the delimiter row (index + 1) holds no cell
        // content. escapeRowPipes leaves structural pipes alone, so escaping the header can't skew its cell count.
        escapeRow(index, expected);

        for (let row = index + 2; row < lines.length; row++) {
            if (isTableBodyEnd(lines[row]!) || isNotARow(lines[row]!, expected)) {
                break;
            }

            escapeRow(row, expected);
        }
    }

    return isChanged ? lines.join('\n') : markdown;
};
