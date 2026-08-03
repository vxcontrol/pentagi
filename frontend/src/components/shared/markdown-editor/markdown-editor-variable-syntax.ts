// Pure Go-template `{{ … }}` matching — no ProseMirror/tiptap imports, so consumers that only need the
// syntax helpers (settings-prompt's variable panel: cycle + "used" count) can pull them without dragging the
// editor (and its tiptap chunk) into their eager bundle. The DOM-facing decoration side lives in
// markdown-editor-variable-highlight.ts, which reuses VARIABLE_RE / variableUseRegex from here.

// `[^{}]` keeps the scan linear (no catastrophic backtracking); Go actions never nest braces.
export const VARIABLE_RE = /\{\{[^{}]*\}\}/g;

const escapeRegExp = (value: string): string => value.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');

// Tests whether a single `{{ … }}` block references `variable` (`.Name` on a word boundary). Always used
// block-first — extract `{{ … }}` blocks with the linear VARIABLE_RE, THEN probe each — never a lazy
// `[^{}]*?` around the name, which backtracks O(n²) on an unclosed `{{`. Shared with settings-prompt.tsx
// (panel cycle + count) so the "used" badge and the editor cycle agree on what counts as a use.
export const variableUseRegex = (variable: string): RegExp => new RegExp(`\\.${escapeRegExp(variable)}\\b`);

export const findVariableUseRanges = (value: string, variable: string): { index: number; length: number }[] => {
    const probe = variableUseRegex(variable);
    const ranges: { index: number; length: number }[] = [];

    for (const match of value.matchAll(VARIABLE_RE)) {
        if (probe.test(match[0])) {
            ranges.push({ index: match.index, length: match[0].length });
        }
    }

    return ranges;
};

export interface VariableRange {
    end: number;
    start: number;
}

// Shared by BOTH variable panels (rich ProseMirror + raw textarea), so their cycling can't drift.
export function nextVariableRange(
    ranges: VariableRange[],
    selectionStart: number,
    selectionEnd: number,
): undefined | VariableRange {
    const current = ranges.findIndex((range) => range.start === selectionStart && range.end === selectionEnd);

    return current >= 0
        ? ranges[(current + 1) % ranges.length]
        : (ranges.find((range) => range.start >= selectionStart) ?? ranges[0]);
}
