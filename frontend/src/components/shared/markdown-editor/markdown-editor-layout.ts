import type { AnyExtension, JSONContent, MarkdownParseResult } from '@tiptap/core';
import type { Node as PMNode } from '@tiptap/pm/model';

import { Extension } from '@tiptap/core';
import { Plugin } from '@tiptap/pm/state';
import { ReplaceStep } from '@tiptap/pm/transform';
import { Lexer } from 'marked';

import {
    holdsProtectedPipe,
    isNotARow,
    isTableBodyEnd,
    TABLE_DELIMITER_LINE,
    TABLE_PIPE,
} from './markdown-editor-table-pipes';

// A document's ProseMirror structure says which blocks it has, not how they were written: whether a blank line
// stood between two of them, which character a list used, how a table was padded. Agents write the document
// and a person edits a line of it, so the rest is saved the way it was written. The facts are read off the
// marked tokens here, travel on the nodes as attributes that are never rendered, and every writer falls back to
// the editor's own layout wherever the stored one would change how the text parses.

type LayoutToken = {
    header?: LayoutToken[];
    items?: LayoutToken[];
    layout?: Record<string, unknown>;
    ordered?: boolean;
    raw: string;
    rows?: LayoutToken[][];
    task?: boolean;
    text?: string;
    tokens?: LayoutToken[];
    type: string;
};

const note = (token: LayoutToken, facts: Record<string, unknown>): void => {
    token.layout = { ...token.layout, ...facts };
};

const newlines = (text: string): number => text.split('\n').length - 1;

// `/\n+$/` starts over at every line break of a run that is not the last one.
export const lineBreaksAtEnd = (text: string): string => {
    let end = text.length;

    while (end > 0 && text[end - 1] === '\n') {
        end--;
    }

    return text.slice(end);
};

// The newline that ends the last line of content is not a blank line; every one after it is.
const blankLinesAfter = (text: string): number => {
    let count = 0;

    for (let at = text.length - 1; at >= 0 && ' \t\r\n'.includes(text[at]!); at--) {
        count += text[at] === '\n' ? 1 : 0;
    }

    return Math.max(0, count - 1);
};

const ITEM_MARKER = /^ {0,3}(?:([-*+])|(\d{1,9})([.)]))([ \t]*)(.*)$/;

// A row written `| a | b |`: one space of padding, an empty cell as one space or two. This only picks a style
// — a wrong pick is still a table — but the row marked hands over holds a stand-in for each pipe of its code
// spans and Go actions, or `\|` where the author escaped one, and a split that took either for a separator
// would call every such table padded.
const COMPACT_CELL = /^(?: {1,2}| \S(?:.*\S)? )$/;

const cellsOf = (row: string): string[] =>
    row
        .replace(/\\[\\|]/g, 'xx')
        .replaceAll(TABLE_PIPE, 'x')
        .split('|');

const isCompactRow = (row: string): boolean => {
    const cells = cellsOf(row);

    return (
        cells.length > 2 && !cells[0] && !cells.at(-1) && cells.slice(1, -1).every((cell) => COMPACT_CELL.test(cell))
    );
};

// A line of text right under a table is a row to marked, written without a single pipe. Beside the style of
// the table, every row is kept as it was written, next to the cells marked read in it: a row that still holds
// those cells is saved as that line (withWrittenRows).
const annotateTable = (table: LayoutToken): void => {
    const lines = table.raw.slice(0, table.raw.length - lineBreaksAtEnd(table.raw).length).split('\n');
    const [header = '', delimiterRow = '', ...rows] = lines.map((line) => line.trim());
    const entry = (line = '', cells: LayoutToken[] = []): string[] => [line, ...cells.map((cell) => cell.text ?? '')];
    // How an empty cell was written, for a row that has to be written anew.
    const empty = [header, ...rows].flatMap((row) => cellsOf(row).slice(1, -1)).filter((cell) => !cell.trim());

    note(table, {
        written: [
            entry(lines[0], table.header),
            [lines[1] ?? ''],
            ...(table.rows ?? []).map((row, at) => entry(lines[at + 2], row)),
        ],
        ...(isCompactRow(header) && rows.every((row) => !row.includes('|') || isCompactRow(row))
            ? { delimiterRow, ...(empty.length && empty.every((cell) => cell === '  ') ? { emptyCell: 2 } : {}) }
            : {}),
    });
};

// A fence that nothing closes runs to the end of the document. Closing it adds a line the author did not
// write, so the document renderer leaves the last block open again. Every block keeps its lines as they stood
// (`written`), without the line break marked hands over with them when a block follows on the next line.
const annotateCode = (code: LayoutToken): void => {
    const lines = code.raw.split('\n');

    while (lines.length && !lines.at(-1)!.trim()) {
        lines.pop();
    }

    const fence = /^ {0,3}(`{3,}|~{3,})/.exec(lines[0] ?? '')?.[1];
    // Each run is of one character: a line of a thousand backticks is matched once, not from every length.
    const [, ticks, tildes] =
        (lines.length > 1 ? /^ {0,3}(?:(`{3,})(?:~[~`]*)?|(~{3,})(?:`[~`]*)?) *$/.exec(lines.at(-1)!) : null) ?? [];

    note(code, {
        ...(fence === undefined || fence === '```' ? {} : { fence }),
        ...(fence === undefined || (ticks ?? tildes)?.startsWith(fence) ? {} : { unclosed: true }),
        written: lines.join('\n'),
    });
};

// A line of an item's first paragraph written left of the content column — a closing `</tag>` or a sentence
// straight under the list — belongs to that paragraph as much as an indented one. marked hands such a line
// over untouched, where it strips the indentation off every other, and that is the test. `outer[at]` counts
// the enclosing items the line is left of as well: an item's lines are its parent's lines one for one, so the
// counts add up. A paragraph whose lines marked has merged or split is left alone, its lines no longer being
// the ones counted here.
const annotateItem = (item: LayoutToken, outer: number[], base: number, margin: string): void => {
    const written = item.raw.split('\n');
    const read = (item.text ?? '').split('\n');
    // A line of nothing but spaces, as it stood past the margin its list is written at.
    const blanks = written.map((line) =>
        /^ +$/.test(line) && line.startsWith(margin) ? line.slice(margin.length) : '',
    );
    const depths = read.map((line, at) =>
        at > 0 && line.trim() && line === written[at] ? 1 + (outer[base + at] ?? 0) : 0,
    );
    const tokens = item.tokens ?? [];
    const paragraph = tokens.find((token) => token.type !== 'checkbox');

    if (paragraph?.type === 'text' || paragraph?.type === 'paragraph') {
        const lines = newlines(paragraph.raw.trimEnd()) + 1;
        const from = depths.findIndex((depth, at) => depth > 0 && at < lines && /^\S/.test(read[at]!));

        if (from > 0 && newlines(paragraph.text ?? '') + 1 === lines) {
            note(item, { lazyDepth: depths[from], lazyFrom: from });
        }
    }

    // An empty line of a code block in the item, written with spaces that marked took for the item's
    // indentation: with the same ones on every such line of the block, they are what it is saved with.
    let line = 0;

    for (const token of tokens) {
        if (token.type === 'code') {
            const empty = blanks
                .slice(line, line + newlines(token.raw.trimEnd()) + 1)
                .filter((_, at) => read[line + at] === '');

            if (empty[0] && empty.every((blank) => blank === empty[0])) {
                note(token, { blankLine: empty[0] });
            }
        }

        line += newlines(token.raw);
    }

    annotateBlocks(tokens, depths, 0, true, blanks);
};

const annotateList = (list: LayoutToken, outer: number[], base: number): void => {
    const items = list.items ?? [];
    const markers = items.map((item) => ITEM_MARKER.exec(item.raw.split('\n', 1)[0]!));

    if (list.ordered) {
        const numbers = markers.map((marker) => Number(marker?.[2]));

        if (markers[0]?.[3] === ')') {
            note(list, { delimiter: ')' });
        }

        // A list numbered the way the editor numbers it stores nothing and renumbers itself when an item is
        // added or removed. `1. 1. 1.` is a fact about the list. Any other numbering is kept on the items that
        // break the count — and only a list that was read with such numbers writes them, so an item carried
        // into another list by a drag, a Tab or a list toggle does not bring its number along.
        if (numbers.every(Number.isInteger) && numbers.some((number, at) => number !== numbers[0]! + at)) {
            const isRepeated = numbers.every((number) => number === numbers[0]);

            note(list, { numbering: isRepeated ? 'same' : 'own' });
            items.forEach((item, at) => {
                if (!isRepeated && at && numbers[at] !== numbers[at - 1]! + 1) {
                    note(item, { number: numbers[at] });
                }
            });
        }
    } else if (markers[0]?.[1] && markers[0][1] !== '-') {
        note(list, { marker: markers[0][1] });
    }

    // Up to three spaces before every marker, which marked does not count as nesting — unless a line of the
    // list stands left of them, where it would not stay once the list is written that far in.
    const indents = items.map((item) => /^ {0,3}/.exec(item.raw)![0].length);
    const margin = ' '.repeat(indents[0] ?? 0);

    // Nor in a bullet list with a task item: those are rendered upstream, at the margin.
    const isIndented =
        indents[0] &&
        indents.every((indent) => indent === indents[0]) &&
        (list.ordered || !items.some((item) => item.task)) &&
        list.raw.split('\n').every((row) => !row.trim() || row.startsWith(margin));

    if (isIndented) {
        note(list, { indent: indents[0] });
    }

    let line = 0;

    items.forEach((item, at) => {
        if (at) {
            // The one blank line between two items, when it was written with spaces in it.
            const above = items[at - 1]!.raw;
            const gap = blankLinesAfter(above);
            const blank = above.slice(0, -1).split('\n').at(-1) ?? '';
            const written = isIndented && blank.startsWith(margin) ? blank.slice(margin.length) : blank;

            note(item, { gap, ...(/^ +$/.test(written) ? { blank: written } : {}) });
        }

        if (markers[at]?.[5] && /^ {2,4}$/.test(markers[at][4]!)) {
            note(item, { spaces: markers[at][4]!.length });
        }

        annotateItem(item, outer, base + line, isIndented ? margin : '');
        line += newlines(item.raw);
    });
};

const OWNS_ITS_LAST_LINES = new Set(['paragraph', 'table']);

// @tiptap/markdown makes a list of each kind out of a list that mixes plain and task items, by this very test,
// and hands every one of them the facts of the whole: what stood above the list is then above each of them.
const TASK_ITEM = /^(\s*)[-+*]\s+\[([ xX])\]\s+/;

const isSplitByTasks = (list: LayoutToken): boolean => {
    const items = list.items ?? [];
    const tasks = items.filter((item) => TASK_ITEM.test(item.raw)).length;

    return tasks > 0 && tasks < items.length;
};

const annotateBlocks = (
    tokens: LayoutToken[],
    outer: number[] = [],
    base = 0,
    isInItem = false,
    blanks: string[] = [],
): void => {
    let above: null | string = null;
    let line = 0;

    for (const [at, token] of tokens.entries()) {
        if (token.type === 'space') {
            above = above === null ? null : above + token.raw;
        } else if (token.type !== 'checkbox') {
            if (above !== null) {
                // The one blank line above, when it is written with spaces or tabs in it. In a list item
                // marked has taken the item's indentation off its lines, and what is left of that line is
                // not what was written: there it is read off the item's own lines (`blanks`).
                const gap = blankLinesAfter(above);
                const written = isInItem ? (blanks[line - 1] ?? '') : (above.slice(0, -1).split('\n').at(-1) ?? '');
                const blank = gap === 1 ? written : '';

                // Of the facts of a list that is split, only that a blank line stood above it is true of
                // every part.
                note(
                    token,
                    token.type === 'list' && isSplitByTasks(token)
                        ? { gap: Math.min(gap, 1) }
                        : { gap, ...(blank ? { blank } : {}) },
                );
            }

            if (token.type === 'list') {
                // marked ends a list at the last character of its text: the spaces its last line ended with
                // open the space under it.
                const [trail = ''] = tokens[at + 1]?.type === 'space' ? tokens[at + 1]!.raw.split('\n', 1) : [];

                note(token, trail && !isSplitByTasks(token) ? { trail } : {});
                annotateList(token, outer, base + line);
            } else if (token.type === 'blockquote') {
                annotateBlocks(token.tokens ?? [], [], 0, isInItem);
            } else if (token.type === 'table') {
                annotateTable(token);
            } else if (token.type === 'code') {
                annotateCode(token);
            } else if (token.type === 'hr' && token.raw.trim() !== '---') {
                note(token, { rule: token.raw.trim() });
            } else if (token.type === 'heading') {
                // A setext heading: the row of `=` or `-` under its text. Any other is one line, and keeps the
                // spaces that line ended with.
                const [, underline] = /\n {0,3}(=+|-+) *$/.exec(token.raw.trimEnd()) ?? [];
                const [first = ''] = token.raw.split('\n', 1);
                const trail = first.slice(first.trimEnd().length);

                note(token, underline ? { underline } : trail ? { trail } : {});
            }

            // A line of nothing but a tab right under text or under a table is a line of the text and a row of
            // the table, not a blank line under them.
            above = OWNS_ITS_LAST_LINES.has(token.type) ? `x${lineBreaksAtEnd(token.raw)}` : token.raw;
        }

        line += newlines(token.raw);
    }
};

// The line breaks a document ends with are part of no block, so its last block carries them.
export const annotateLayout = (tokens: unknown[], source: string): void => {
    const blocks = (tokens as LayoutToken[]).filter((token) => token.type !== 'space');
    const eof = lineBreaksAtEnd(source);

    annotateBlocks(tokens as LayoutToken[]);

    if (blocks.length && eof) {
        note(blocks.at(-1)!, { eof });
    }
};

const withLayout = (parsed: MarkdownParseResult, layout?: Record<string, unknown>): MarkdownParseResult => {
    const [first, ...rest] = Array.isArray(parsed) ? parsed : [parsed];

    if (!layout || !first || 'mark' in first) {
        return parsed;
    }

    const noted = { ...first, attrs: { ...first.attrs, ...layout } };

    return Array.isArray(parsed) ? [noted, ...rest] : noted;
};

const BLOCKS = [
    'blockquote',
    'bulletList',
    'codeBlock',
    'heading',
    'horizontalRule',
    'image',
    'listItem',
    'orderedList',
    'paragraph',
    'table',
    'taskList',
];

type ParseMarkdown = (this: unknown, token: LayoutToken, helpers: unknown) => MarkdownParseResult;

// The Markdown extension parses the initial content in the hook that creates its manager, so there is no
// moment to reach into the manager first: each block's own handler hands the facts over instead.
export const withLayoutFacts = <T extends AnyExtension>(extension: T): T => {
    const parse = (extension.config as { parseMarkdown?: ParseMarkdown }).parseMarkdown;

    if (!parse) {
        return extension;
    }

    return extension.extend({
        parseMarkdown(token: LayoutToken, helpers: unknown) {
            return withLayout(parse.call(this, token, helpers), token.layout);
        },
    }) as T;
};

// keepOnSplit: Enter in a paragraph makes a new block, and what was written about the old one is not true of it.
// parseHTML: without one tiptap reads the attribute off pasted HTML, where nothing has written it.
const fact = { default: null, keepOnSplit: false, parseHTML: () => null, rendered: false };

const endOf = (doc: PMNode): null | { at: number; eof: string } => {
    let end: null | { at: number; eof: string } = null;

    doc.forEach((block, at) => {
        end ??= typeof block.attrs.eof === 'string' ? { at, eof: block.attrs.eof } : null;
    });

    return end;
};

export const MarkdownLayout = Extension.create({
    addGlobalAttributes() {
        return [
            { attributes: { blank: fact, eof: fact, gap: fact }, types: BLOCKS },
            { attributes: { indent: fact, trail: fact }, types: ['bulletList', 'orderedList'] },
            { attributes: { marker: fact }, types: ['bulletList'] },
            { attributes: { delimiter: fact, numbering: fact }, types: ['orderedList'] },
            { attributes: { trail: fact, underline: fact }, types: ['heading'] },
            { attributes: { lazyDepth: fact, lazyFrom: fact, number: fact, spaces: fact }, types: ['listItem'] },
            { attributes: { blankLine: fact, fence: fact, unclosed: fact, written: fact }, types: ['codeBlock'] },
            { attributes: { rule: fact }, types: ['horizontalRule'] },
            { attributes: { form: fact }, types: ['link'] },
            { attributes: { fence: fact }, types: ['code'] },
            { attributes: { link: fact }, types: ['image'] },
            { attributes: { delimiterRow: fact, emptyCell: fact, written: fact }, types: ['table'] },
        ];
    },
    // The block that carries how the document ends can be joined into the one above, deleted, wrapped in a
    // list or retyped by a list command, none of which is an edit of the document's end: the fact moves to
    // the last block. Replacing the whole document is one.
    addProseMirrorPlugins() {
        return [
            new Plugin({
                appendTransaction(transactions, { doc: before }, { doc, tr }) {
                    const end = endOf(before);
                    const isReplaced = transactions.some(({ steps }) =>
                        steps.some(
                            (step) => step instanceof ReplaceStep && !step.from && step.to >= before.content.size,
                        ),
                    );

                    if (!end || endOf(doc) || isReplaced || !transactions.some(({ docChanged }) => docChanged)) {
                        return null;
                    }

                    const at = doc.content.size - (doc.lastChild?.nodeSize ?? 0);

                    return doc.nodeAt(at)?.type.spec.attrs?.eof ? tr.setNodeAttribute(at, 'eof', end.eof) : null;
                },
            }),
        ];
    },
    name: 'markdownLayout',
});

const CLOSED_BLOCKS = new Set(['codeBlock', 'heading', 'horizontalRule']);
const LISTS = new Set(['bulletList', 'orderedList', 'taskList']);
const ENDS_AT_A_TAB_LINE = new Set([...CLOSED_BLOCKS, ...LISTS]);
const OPENS_LIST_UNDER_TEXT = /^(?:[-*+]|1[.)])[ \t]+\S/;
const OPENS_LIST = /^(?:([-*+])|\d{1,9}([.)]))[ \t]+\S/;
const OPENS_DEFINITION = /^ {0,3}\[(?:\\.|[^[\]\\])+\]: *$/;
const UNDERLINE = /^ {0,3}(?:=+|-+) *$/;
const RULE_LINE = String.raw`(?:- *){3,}$|(?:_ *){3,}$|(?:\* *){3,}$`;
const TAG_START = String.raw`<(?:[a-z].*>|!--)`;
const OPENS_TAG = new RegExp(`^${TAG_START}`, 'i');

const RULE = new RegExp(`^(?:${RULE_LINE})`);

export const isRule = (line: string): boolean => RULE.test(line);

const markerKindOf = (list: JSONContent): string =>
    list.type === 'orderedList' ? (list.attrs?.delimiter === ')' ? ')' : '.') : bulletOf(list);

// marked's own rule for a paragraph: it stops at a line that opens a block, a line of a known HTML tag among
// them, though here that tag is text. Two lines are handed over: a line over a row of dashes and pipes ends
// the text above it by the look of a table, whether it turns out to be one or not.
const endsText = (lines: string): boolean => Lexer.rules.block.gfm.paragraph.exec(`x\n${lines}`)?.[1] === 'x';

const columnsOf = (table: JSONContent): number => table.content?.[0]?.content?.length ?? 0;

// Whether `next`, written as `written`, can stand on the line right under `previous` and still be read as a
// block of its own — each pair was run through marked, with every kind of content the editor writes. Text
// continues any block a blank line would have closed, except that outside a list item a line marked takes
// for the start of a block ends the text above it, and any `<tag>` ends a list. `---` under a line of text,
// an image included, is a setext underline, and a row of dashes and pipes under one is a table's delimiter
// row. A quote or a table swallows a second one of its kind, and a table the list under it and any text that
// is not a line where it ends by itself or one the loader takes for no row of it (isNotARow). A list opens
// under text only with a first item that has text, and a numbered one only from 1; under another list it is
// a list of its own only with a different bullet or delimiter.
const canSitUnder = (previous: JSONContent, next: JSONContent, written: string, isInItem: boolean): boolean => {
    const above = previous.type ?? '';
    const [line = ''] = written.split('\n', 1);
    const isUnderText = above === 'paragraph' && !TABLE_DELIMITER_LINE.test(line);

    switch (next.type) {
        case 'blockquote':
            return above !== 'blockquote';
        case 'bulletList':
        case 'orderedList': {
            if (above === 'bulletList' || above === 'orderedList') {
                const [, bullet, delimiter] = OPENS_LIST.exec(written) ?? [];

                return (bullet ?? delimiter ?? markerKindOf(previous)) !== markerKindOf(previous);
            }

            return CLOSED_BLOCKS.has(above) || (isUnderText && OPENS_LIST_UNDER_TEXT.test(written));
        }

        case 'codeBlock':
            return true;
        case 'heading':
            return CLOSED_BLOCKS.has(above) || /^#{1,6}( [^\n]*)?$/.test(written);
        case 'horizontalRule':
            return (
                !/^-/.test(written) ||
                CLOSED_BLOCKS.has(above) ||
                LISTS.has(above) ||
                above === 'blockquote' ||
                above === 'table'
            );
        case 'paragraph':
            return (
                CLOSED_BLOCKS.has(above) ||
                (above === 'table' && (isTableBodyEnd(line) || isNotARow(line, columnsOf(previous)))) ||
                (!isInItem && above === 'paragraph' && endsText(written.split('\n', 2).join('\n'))) ||
                (!isInItem && (above === 'bulletList' || above === 'orderedList') && OPENS_TAG.test(line))
            );
        case 'table':
            return CLOSED_BLOCKS.has(above) || isUnderText;
        case 'taskList':
            return CLOSED_BLOCKS.has(above);
        default:
            return false;
    }
};

type Render = (block: JSONContent, index: number) => string;

const endsWithText = (list: JSONContent): boolean => {
    const block = list.content?.at(-1)?.content?.at(-1);

    return LISTS.has(block?.type ?? '')
        ? endsWithText(block!)
        : block?.type === 'paragraph' && Boolean(block.content?.length);
};

// A code block as it was written, while those lines still read as the block it is now: its code and its
// language, closed by its own last line. One written by indentation and edited since is written with every
// line four spaces in, where that reads as its code — not with a language, a blank first line or a line break
// at its end. Anything else is left to the fenced form.
const writtenCode = (block: JSONContent): null | { isIndented: boolean; text: string } => {
    const { language, written } = block.attrs ?? {};
    const code = (block.content ?? []).map((node) => node.text ?? '').join('');

    if (block.type !== 'codeBlock' || typeof written !== 'string') {
        return null;
    }

    const isIndented = /^(?: {4}| {0,3}\t)/.test(written);

    // A line of an indented block that looks like a fence is one to the scan that shields the tabs of fenced
    // code on the next load: between fences longer than it, it is a line of code to every reader.
    if (isIndented && /^[ \t>]*(?:`{3,}|~{3,})/m.test(code)) {
        return null;
    }

    // So is a closing fence that scan does not take for one: a run of one character and nothing after it.
    if (!isIndented && !/^ {0,3}(?:`+|~+) *$/.test(written.slice(written.lastIndexOf('\n') + 1))) {
        return null;
    }

    const typed = code
        .split('\n')
        .map((line) => (line ? `    ${line}` : line))
        .join('\n');

    const reads = (text: string): boolean => {
        const [token, next] = new Lexer({ gfm: true }).blockTokens(`${text}\nx`) as (LayoutToken & {
            lang?: string;
        })[];
        const read = token?.text ?? '';

        // An indented block is read without the line break marked leaves at its end, as the loader takes it.
        return (
            token?.type === 'code' &&
            next?.type === 'paragraph' &&
            next.raw === 'x' &&
            (isIndented ? read.slice(0, read.length - lineBreaksAtEnd(read).length) : read) === code &&
            (token.lang ?? '') === (typeof language === 'string' ? language : '')
        );
    };

    const text = [written, ...(isIndented ? [typed] : [])].find(reads);

    return text === undefined ? null : { isIndented, text };
};

// Two lists of one kind next to each other, which only an edit leaves behind, are one list to every reader:
// they are written as one, numbered through, with no blank line that would make it a loose one. Loose task
// items are read as lists of their own, and the blank line between those was written.
const withListsJoined = (blocks: JSONContent[]): JSONContent[] =>
    blocks.reduce<JSONContent[]>((joined, block) => {
        const previous = joined.at(-1);
        const isWrittenApart = block.type === 'taskList' && Number(block.attrs?.gap) > 0;

        if (
            previous &&
            LISTS.has(block.type ?? '') &&
            previous.type === block.type &&
            markerKindOf(previous) === markerKindOf(block) &&
            !isWrittenApart
        ) {
            joined[joined.length - 1] = {
                ...previous,
                content: [...(previous.content ?? []), ...(block.content ?? [])],
            };
        } else {
            joined.push(block);
        }

        return joined;
    }, []);

// `gap` is the number of blank lines the block was written under: none keeps it on the next line where that
// is safe, any keeps the blank line. A block the user made has no gap and takes `isTightByDefault`, which is
// how the blocks of a list item are joined and those of the document are not. A block that writes nothing —
// an empty paragraph — is a blank line of its own, and one more on each side of it is what brings it back on
// the next load. Two things an edit can leave reach past the block above: a `[label]:` line takes the line
// under it for its target, and marked reads text, a `***` rule right under it and then an underline as one
// setext heading.
export const joinBlocks = (
    siblings: JSONContent[],
    render: Render,
    isTightByDefault: boolean,
    place: (written: string, index: number, block: JSONContent) => string = (written) => written,
): string => {
    const blocks = withListsJoined(siblings);
    let above = '';
    let aboveType = '';
    let empty = 0;
    let isAboveIndented = false;
    // A list above takes a line written as far in as the content of its last item for that item's, whatever
    // blank lines stand between: the list as it is written, until a line at the margin has ended it.
    let listAbove: JSONContent | undefined;
    let openList = '';
    let isDashOpen = false;
    let isTextOpen = false;

    return blocks.reduce((output, block, index) => {
        // The lines a code block was written as, where they stand for themselves: at the margin of the
        // document or of a quote. In a list item — and in a quote inside one, where marked turns tabs into
        // spaces — indentation is the item's own, so is that of a line under a list, and two indented blocks
        // one under the other are one.
        const kept = itemDepth > 0 ? null : writtenCode(block);
        const isKept =
            kept !== null && !(openList && /^[ \t]/.test(kept.text)) && !(kept.isIndented && isAboveIndented);
        const indented = isKept && kept.isIndented ? kept.text : null;
        const rendered = isKept ? kept.text : render(block, index);
        // Nor do the lines of a table keep the spaces they open with there, or text its own where the list
        // would take it: a task item takes any line that does not start at the margin, and for any other
        // list marked is asked about a word written that far in.
        const [lead] = /^[ \t]*/.exec(rendered)!;
        const isTaken =
            block.type === 'paragraph' &&
            openList !== '' &&
            lead !== '' &&
            (listAbove?.type === 'taskList' ||
                new Lexer({ gfm: true }).blockTokens(`${openList}\n\n${lead}x`).at(-1)?.raw !== `${lead}x`);
        const written =
            block.type === 'table' && (openList || itemDepth > 0)
                ? rendered
                      .split('\n')
                      .map((line) => line.trimStart())
                      .join('\n')
                : isTaken
                  ? rendered.trimStart()
                  : rendered;
        const previous = blocks[index - 1];
        const gap: unknown = block.attrs?.gap;
        // The underline of a heading is its last line; any other block reaches the text above only with its first.
        const [first = ''] = written.split('\n', 1);
        const underline = block.type === 'heading' ? written.slice(written.lastIndexOf('\n') + 1) : first;
        const isClaimed =
            OPENS_DEFINITION.test(above.slice(above.lastIndexOf('\n') + 1)) ||
            (isTextOpen && previous?.type === 'horizontalRule' && UNDERLINE.test(underline));
        const isTight =
            above !== '' &&
            written !== '' &&
            !isClaimed &&
            (typeof gap === 'number' ? gap === 0 : isTightByDefault) &&
            (indented === null
                ? canSitUnder(blocks[index - 1]!, block, written, isTightByDefault)
                : CLOSED_BLOCKS.has(aboveType));
        const { blank, indent } = block.attrs ?? {};
        // A blank line is written with its spaces between two blocks that write something; with a tab in it
        // only under a block a tab does not continue — it is a row to a table and a line to text.
        const isBlankWritten =
            above !== '' &&
            written !== '' &&
            typeof blank === 'string' &&
            /^[ \t]+$/.test(blank) &&
            (!blank.includes('\t') || ENDS_AT_A_TAB_LINE.has(previous?.type ?? ''));
        // Every two blank lines past the first are read as an empty paragraph, which writes two: of an even
        // number one is left over, and is written while the empty paragraphs above the block are the ones it
        // was read with — half an odd number is no count of them.
        // Not between a list of `-` bullets and the task lists under it, which are one list to marked: there
        // the line is an empty paragraph of an item.
        const isDashList = block.type === 'taskList' || (block.type === 'bulletList' && bulletOf(block) === '-');
        const isOneList = isDashOpen && isDashList && openList !== '';
        const isOneMore = typeof gap === 'number' && gap > 1 && empty === gap / 2 - 1 && !isOneList;
        // Under a task list one of the blank lines is read as part of the list: an odd count above two comes
        // back two short there, and any count one short where an empty paragraph stands between.
        const isTwoMore =
            typeof gap === 'number' && gap > 2 && gap % 2 === 1 && empty === (gap - 1) / 2 - 1 && !isOneList;
        const more =
            aboveType === 'taskList' && written !== ''
                ? isTwoMore
                    ? 2
                    : isOneMore || empty > 0
                      ? 1
                      : 0
                : Number(isOneMore);
        const separator = index ? (isTight ? '\n' : isBlankWritten ? `\n${blank}\n` : `\n\n${'\n'.repeat(more)}`) : '';
        // A list keeps the spaces its markers were written after, except where they reach the column an item
        // of the list above — the nearest block above that writes anything — has its content at, which would
        // nest it there, and with a line left of its items, which those spaces would move. A task item takes
        // any line that does not start at the margin.
        const contentColumn =
            aboveType === 'orderedList' ? 3 : aboveType === 'taskList' ? 1 : LISTS.has(aboveType) ? 2 : 4;
        const pad =
            (block.type === 'bulletList' || block.type === 'orderedList') &&
            (indent === 1 || indent === 2 || indent === 3) &&
            indent < contentColumn &&
            !/"lazyFrom":\d/.test(JSON.stringify(block))
                ? indent
                : 0;
        // The spaces a list ended with are written after the text of its last item, and only above a block
        // that is not a list: at the end of the document nothing reads them back, and above a list they are
        // text of the item.
        const { trail } = block.attrs ?? {};
        const hasTrail =
            typeof trail === 'string' && /^ +$/.test(trail) && LISTS.has(block.type ?? '') && endsWithText(block);
        const under = hasTrail ? blocks.slice(index + 1).find((later) => !isWrittenEmpty(later)) : undefined;
        const isTrailed = under !== undefined && !LISTS.has(under.type ?? '');
        const placed = (pad ? indentList(written, pad) : written) + (isTrailed ? trail : '');

        isTextOpen =
            block.type === 'paragraph' || block.type === 'image'
                ? written !== ''
                : block.type === 'horizontalRule' && isTextOpen && isTight;
        above = written;
        aboveType = written === '' ? aboveType : (block.type ?? '');
        empty = written === '' ? empty + 1 : 0;
        isAboveIndented = written.trim() === '' ? isAboveIndented : indented !== null;

        if (LISTS.has(block.type ?? '') && written !== '') {
            isDashOpen =
                block.type === 'taskList' ? isDashOpen : block.type === 'bulletList' && bulletOf(block) === '-';
            listAbove = block;
            openList = placed;
        } else if (written.trim() !== '' && (block.type === 'paragraph' || /^\S/.test(written))) {
            // Text that kept its spaces is what marked did not give to the list; white space alone is a blank
            // line to it.
            isDashOpen = false;
            openList = '';
        }

        return output + separator + place(placed, index, block);
    }, '');
};

// A line of an item's paragraph that is written left of the content column carries one mark for every
// enclosing item that must leave it there; an item takes a mark off instead of indenting.
const LAZY_LINE = '\uE001';
const LAZY_MARKS = /^\uE001+/;

const indentList = (list: string, indent: number): string =>
    list
        .split('\n')
        .map((line) => (line ? ' '.repeat(indent) + line : line))
        .join('\n');

// A document that holds the mark itself, in its text or in an attribute: it is written with every line
// indented, and nothing in it is taken for a mark.
let hasOwnMarks = false;

// marked keeps an under-indented line in an item only as the continuation of a line of text: not under a
// blank line, not under one whose remainder past the content column opens a block or indented code, and not
// if the line itself opens one — `<tag>` and `#word` included, which end an item though they are text
// anywhere else.
const BLOCK_START = new RegExp(
    `^ {0,3}(?:#|>|${TAG_START}|\`{3}|~{3}|(?:[*+-]|\\d{1,9}[.)])(?:[ \\t]|$)|${RULE_LINE})`,
    'i',
);
const ENDS_LAZINESS = new RegExp(`^ {0,3}(?:#|\`{3}|~{3}|${RULE_LINE})`);

const mayStayLazy = (line: string, above: string): boolean =>
    /^\S/.test(line) &&
    !BLOCK_START.test(line) &&
    above.trim() !== '' &&
    !above.startsWith('    ') &&
    !ENDS_LAZINESS.test(above);

// Indents the lines of one block of an item to the item's content column, from line `from` on. A line from
// `lazyFrom` on, or one a nested item has marked, stays where it is while marked would still read it as a
// continuation; the first that would not is indented, and so is every such line right under it. An empty
// line is written as `blank`.
export const indentItemBlock = (
    block: string,
    pad: string,
    from = 0,
    lazyFrom = Infinity,
    lazyDepth = 1,
    blank = '',
): string => {
    let above = '';
    let isRefused = false;

    return block
        .split('\n')
        .map((line, index) => {
            const text = hasOwnMarks ? line : line.replace(LAZY_MARKS, '');
            const isLazy = !hasOwnMarks && (text.length < line.length || index >= lazyFrom);

            if (index < from || !line) {
                above = text;
                isRefused = false;

                return line || blank;
            }

            if (isLazy && !isRefused && mayStayLazy(text, above)) {
                above = text.replaceAll('\t', '    ').slice(pad.length);

                return text.length < line.length ? line.slice(1) : LAZY_LINE.repeat(lazyDepth - 1) + text;
            }

            above = text;
            isRefused = isLazy;

            return pad + text;
        })
        .join('\n');
};

// The spaces the empty lines of a code block were written with in its item, while they still fit under the
// item's marker: past its content column they would be read as the code's own.
export const blankLineOf = (block: JSONContent, pad: string): string => {
    const blank: unknown = block.attrs?.blankLine;

    return typeof blank === 'string' && /^ +$/.test(blank) && blank.length <= pad.length ? blank : '';
};

export const isWrittenEmpty = (block: JSONContent | undefined): boolean =>
    block?.type === 'paragraph' && !block.content?.length;

// The blocks of the document, and around them what belongs to no block: the fence an unclosed code block
// never had — an opening fence with nothing under it, not even a line break, is text of the block above —
// and the line breaks the document ended with. Those stand for the empty paragraphs at its end —
// the ones they were read as, and the one TrailingNode puts under a last block that is not a paragraph as
// soon as anything is edited. A heading without text, or text that is only spaces outside a code span,
// writes nothing either, and left at the end it would put two line breaks there that nobody wrote. Under an
// unclosed block all but the last of them are lines of its code.
export const renderDocument = (doc: JSONContent, render: Render): string => {
    const all = doc.content ?? [];
    // A heading is written trimmed; in a paragraph only spaces and tabs are nothing to marked.
    const isNothing = (text: string, type?: string) => (type === 'heading' ? !text.trim() : /^[ \t]*$/.test(text));
    const writesNothing = (block: JSONContent) =>
        (block.type === 'heading' || block.type === 'paragraph') &&
        (block.content ?? []).every(
            (node) =>
                node.type === 'hardBreak' ||
                (node.type === 'text' &&
                    isNothing(node.text ?? '', block.type) &&
                    !node.marks?.some((mark) => mark.type === 'code')),
        );
    const blocks = all.slice(0, all.findLastIndex((block) => !writesNothing(block)) + 1);
    const last = blocks.at(-1);
    const eof: unknown = all.find((block) => typeof block.attrs?.eof === 'string')?.attrs?.eof;
    const hasLineBreak = typeof eof === 'string' && eof !== '' && lineBreaksAtEnd(eof) === eof;
    const code = last?.content?.[0]?.text ?? '';
    const isUnclosed = last?.type === 'codeBlock' && last.attrs?.unclosed === true && (code !== '' || hasLineBreak);

    hasOwnMarks = JSON.stringify(blocks).includes(LAZY_LINE);

    let written = joinBlocks(blocks, render, false);

    written = hasOwnMarks ? written : written.replaceAll(LAZY_LINE, '');

    if (isUnclosed) {
        written = written.slice(0, written.lastIndexOf('\n') - (code ? 0 : 1));
    }

    if (typeof eof === 'string' && hasLineBreak) {
        return written + (isUnclosed ? '\n' : eof);
    }

    // The empty line an open block ends with is read as the document's line break, unless one more follows.
    return isUnclosed && code.endsWith('\n') ? `${written}\n` : written;
};

const itemMarkers = new WeakMap<JSONContent, string>();

export const itemMarker = (item: JSONContent): string => itemMarkers.get(item) ?? '- ';

const bulletOf = (list: JSONContent): string => {
    const marker: unknown = list.attrs?.marker;

    return marker === '*' || marker === '+' ? marker : '-';
};

const isListNumber = (value: unknown): value is number =>
    typeof value === 'number' && Number.isInteger(value) && value >= 0 && value <= 999_999_999;

// A list writes its items' markers, because a number depends on the items above it.
export const renderList = (list: JSONContent, render: (item: JSONContent, index: number) => string): string => {
    const { numbering, start } = list.attrs ?? {};
    const first = isListNumber(start) ? start : 1;
    let number = first - 1;

    const items = list.content ?? [];

    return items.reduce((output, item, index) => {
        const { gap, number: written, spaces } = item.attrs ?? {};

        number = numbering === 'own' && isListNumber(written) ? written : numbering === 'same' ? first : number + 1;
        itemMarkers.set(
            item,
            (list.type === 'orderedList' ? `${number}${markerKindOf(list)}` : bulletOf(list)) +
                ' '.repeat(typeof spaces === 'number' && spaces > 1 && spaces < 5 ? spaces : 1),
        );

        // Two blank lines between items load as an empty paragraph closing the upper item, which writes one
        // of them itself.
        const isLoose = typeof gap === 'number' && gap > 0 && !isWrittenEmpty(items[index - 1]?.content?.at(-1));
        // The spaces the one blank line above the item was written with.
        const blank = gap === 1 && typeof item.attrs?.blank === 'string' && /^ +$/.test(item.attrs.blank);

        return output + (index ? (isLoose ? `\n${blank ? item.attrs?.blank : ''}\n` : '\n') : '') + render(item, index);
    }, '');
};

type Alignment = 'center' | 'left' | 'right' | null;

const alignmentOf = (cell: string): Alignment | undefined => {
    const [, left, right] = /^(:?)-+(:?)$/.exec(cell.trim()) ?? [];

    if (left === undefined) {
        return undefined;
    }

    return left && right ? 'center' : left ? 'left' : right ? 'right' : null;
};

const DELIMITER_CELL: Record<string, string> = { center: ':---:', left: ':---', right: '---:' };

const alignmentsOf = (table: JSONContent): Alignment[] => {
    const rows = (table.content ?? []).map((row) => row.content ?? []);
    const width = rows.reduce((max, row) => Math.max(max, row.length), 0);

    return Array.from(
        { length: width },
        (_, column): Alignment =>
            (rows.map((row) => row[column]?.attrs?.align).find(Boolean) as Alignment | undefined) ?? null,
    );
};

// Whether a delimiter row still describes the columns: their number and their alignment.
const describes = (delimiterRow: string, alignments: Alignment[]): boolean => {
    const written = delimiterRow
        .trim()
        .replace(/^\||\|$/g, '')
        .split('|')
        .map(alignmentOf);

    return (
        delimiterRow.includes('|') &&
        written.length === alignments.length &&
        written.every((alignment, column) => alignment === alignments[column])
    );
};

// The table as rows of cell text, each cell one line, with one space around it and an empty cell as the one
// or two spaces the table was written with. The delimiter row is the written one while it still describes the
// columns, and a plain one once the table has been reshaped.
export const renderCompactTable = (table: JSONContent, delimiterRow: string, emptyCell: unknown): string => {
    const alignments = alignmentsOf(table);
    const line = (cells: string[]) =>
        `|${alignments.map((_, column) => (cells[column] ? ` ${cells[column]} ` : emptyCell === 2 ? '  ' : ' ')).join('|')}|`;
    const texts = (table.content ?? []).map((row) =>
        (row.content ?? []).map((cell) => (cell.content?.[0]?.text ?? '').trim()),
    );

    return [
        line(texts[0] ?? []),
        describes(delimiterRow, alignments)
            ? delimiterRow
            : line(alignments.map((alignment) => DELIMITER_CELL[alignment ?? ''] ?? '---')),
        ...texts.slice(1).map(line),
    ].join('\n');
};

// How many list items the writer is inside of.
let itemDepth = 0;

export const withinItem = <T>(write: () => T): T => {
    itemDepth += 1;

    try {
        return write();
    } finally {
        itemDepth -= 1;
    }
};

// A line that holds something: one of nothing but white space is a row only until a quote or a list item is
// written around it.
const isLine = (value: unknown): value is string =>
    typeof value === 'string' && value.trim() !== '' && !value.includes('\n');

// The lines of a table as it is written now, each replaced by the line it was written as while that line
// still says the same: a row by the cells marked read in it — the writer's own line holds the same ones once
// marked has trimmed them and taken the backslash off an escaped pipe — and the delimiter row by the columns
// it describes. Rows are matched by what they hold, not by where they stand, so a row added, removed or moved
// costs no other row its line; the header only by the header, which a line without a pipe cannot be. In a
// list item a line that needs a pipe protected is not kept: the scan that protects it on the next load does
// not find every table there.
export const withWrittenRows = (lines: string[], table: JSONContent, written: unknown): string => {
    const entries = Array.isArray(written) ? written.filter((entry): entry is unknown[] => Array.isArray(entry)) : [];
    const [header, delimiter, ...body] = entries;
    const keyOf = (cells: unknown[]): string => JSON.stringify(cells);
    const rows = (table.content ?? []).map((row) =>
        keyOf((row.content ?? []).map((cell) => (cell.content?.[0]?.text ?? '').trim().replace(/\\\|/g, '|'))),
    );
    const columns = alignmentsOf(table).length;
    // A line that would end the table once another stands under it — a fence, a tag — is a row only as the
    // last line of its container.
    const isKept = (raw: unknown): raw is string =>
        isLine(raw) && !isTableBodyEnd(raw) && !(itemDepth > 0 && holdsProtectedPipe(raw, columns));
    // Rows that hold the same cells are handed out in the order they were written in.
    const kept = new Map<string, { next: number; raws: string[] }>();

    for (const [raw, ...cells] of body) {
        if (isKept(raw)) {
            const key = keyOf(cells);
            const same = kept.get(key) ?? { next: 0, raws: [] };

            same.raws.push(raw);
            kept.set(key, same);
        }
    }

    return lines
        .map((line, at) => {
            if (at === 1) {
                return isLine(delimiter?.[0]) && describes(delimiter[0], alignmentsOf(table)) ? delimiter[0] : line;
            }

            if (!at) {
                return header && isKept(header[0]) && keyOf(header.slice(1)) === rows[0] ? header[0] : line;
            }

            const same = kept.get(rows[at - 1] ?? '');

            return same?.raws[same.next++] ?? line;
        })
        .join('\n');
};
