import type { Editor } from '@tiptap/react';
import type { LucideIcon } from 'lucide-react';

import { AlignCenter, AlignLeft, AlignRight } from 'lucide-react';

export type ColumnAlign = 'center' | 'left' | 'right';

export const ALIGN_OPTIONS: { icon: LucideIcon; label: string; value: ColumnAlign }[] = [
    { icon: AlignLeft, label: 'Left', value: 'left' },
    { icon: AlignCenter, label: 'Center', value: 'center' },
    { icon: AlignRight, label: 'Right', value: 'right' },
];

export function clearLineContents(editor: Editor, axis: 'column' | 'row'): void {
    editor
        .chain()
        .focus()
        .command(({ dispatch, editor: instance, tr }) => {
            const paragraphType = instance.schema.nodes.paragraph;

            if (dispatch && paragraphType) {
                // Descending order so each replacement leaves earlier cell positions valid.
                for (const pos of cellPositionsInLine(editor, axis).reverse()) {
                    const cell = tr.doc.nodeAt(pos);

                    if (cell) {
                        tr.replaceWith(pos + 1, pos + cell.nodeSize - 1, paragraphType.create());
                    }
                }
            }

            return true;
        })
        .run();
}

// Whether a table uses a header row (its first row holds `tableHeader` cells). Defaults to the table around the
// caret; pass `pos` to read a specific cell's table instead — the hover grips open without moving the selection,
// so they must resolve the hovered table by position rather than the caret's last table.
export function hasHeaderRow(editor: Editor, pos?: number): boolean {
    const { doc, selection } = editor.state;

    // A hover grip can outlive the edit that shrank the doc below its captured cellPos. doc.resolve throws a
    // RangeError on an out-of-bounds position, and this runs inside a useEditorState selector, so an unguarded
    // throw takes down the whole editor render tree. A stale position has no header row to report.
    if (pos != null && pos > doc.content.size) {
        return false;
    }

    const $pos = pos == null ? selection.$from : doc.resolve(pos);

    for (let depth = $pos.depth; depth > 0; depth--) {
        if ($pos.node(depth).type.name === 'table') {
            return $pos.node(depth).firstChild?.firstChild?.type.name === 'tableHeader';
        }
    }

    return false;
}

// GFM aligns whole COLUMNS (the delimiter row), stored by @tiptap/extension-table as an `align` attr on every
// cell — setCellAttribute only touches the caret cell, so set it on the whole column or only one cell aligns
// until the doc reloads.
export function setColumnAlign(editor: Editor, align: ColumnAlign): void {
    editor
        .chain()
        .focus()
        .command(({ dispatch, tr }) => {
            if (dispatch) {
                for (const pos of cellPositionsInLine(editor, 'column')) {
                    tr.setNodeAttribute(pos, 'align', align);
                }
            }

            return true;
        })
        .run();
}

// Walks the table containing the caret and yields the doc position of every cell in the caret's column (or row).
// colspan is always 1 here — GFM tables have no merged cells — so a cell's index within its row IS its column.
function cellPositionsInLine(editor: Editor, axis: 'column' | 'row'): number[] {
    const { $from } = editor.state.selection;
    let depth = $from.depth;

    while (depth > 0 && !['tableCell', 'tableHeader'].includes($from.node(depth).type.name)) {
        depth--;
    }

    if (depth === 0) {
        return [];
    }

    const targetIndex = axis === 'column' ? $from.index(depth - 1) : $from.index(depth - 2);
    const tableDepth = depth - 2;
    const tablePos = $from.before(tableDepth);
    const positions: number[] = [];

    $from.node(tableDepth).forEach((row, rowOffset, rowIndex) => {
        row.forEach((cell, cellOffset, cellIndex) => {
            const matches = axis === 'column' ? cellIndex === targetIndex : rowIndex === targetIndex;

            if (matches) {
                positions.push(tablePos + 1 + rowOffset + 1 + cellOffset);
            }
        });
    });

    return positions;
}
