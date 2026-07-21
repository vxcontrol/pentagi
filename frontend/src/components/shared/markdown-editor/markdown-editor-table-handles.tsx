import type { Editor } from '@tiptap/react';

import { useEditorState } from '@tiptap/react';
import {
    AlignLeft,
    ArrowDown,
    ArrowLeft,
    ArrowRight,
    ArrowUp,
    Delete,
    Eraser,
    GripHorizontal,
    GripVertical,
    PanelTop,
} from 'lucide-react';
import { useEffect, useRef, useState } from 'react';
import { createPortal } from 'react-dom';

import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuItem,
    DropdownMenuSeparator,
    DropdownMenuSub,
    DropdownMenuSubContent,
    DropdownMenuSubTrigger,
    DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { Switch } from '@/components/ui/switch';

import { getEditorScrollParent } from './markdown-editor-styles';
import { ALIGN_OPTIONS, clearLineContents, hasHeaderRow, setColumnAlign } from './markdown-editor-table-commands';

interface HoverTarget {
    cellPos: number;
    colLeft: number;
    colWidth: number;
    rowHeight: number;
    rowTop: number;
    tableLeft: number;
    tableTop: number;
}

type OpenMenu = 'column' | 'row' | null;

const GRIP = 18;
const GRIP_CLASS =
    'bg-muted hover:bg-accent text-muted-foreground hover:text-accent-foreground flex items-center justify-center rounded border shadow-sm transition-colors';

interface TableHandlesController {
    onMenuChange: (menu: 'column' | 'row') => (isOpen: boolean) => void;
    open: OpenMenu;
    target: HoverTarget | null;
}

// Notion-style hover handles: a grip appears above the hovered column and left of the hovered row; clicking it
// opens a shadcn menu of markdown-safe row/column operations. Overlay-only — it never mutates the document until
// a menu item runs — so it stays out of the byte round-trip. Merged cells / colour / header-column are omitted
// (not GFM-representable); the toolbar Table dropdown carries the same ops for keyboard/no-hover users.
export function TableHandles({ editor }: { editor: Editor }) {
    const { onMenuChange, open, target } = useTableHandles(editor);

    if (!target) {
        return null;
    }

    const focusTarget = () => editor.chain().focus().setTextSelection(target.cellPos);

    return createPortal(
        <>
            <DropdownMenu
                onOpenChange={onMenuChange('column')}
                open={open === 'column'}
            >
                <DropdownMenuTrigger asChild>
                    <button
                        aria-label="Column actions"
                        className={GRIP_CLASS}
                        data-table-grip=""
                        style={{
                            height: GRIP,
                            left: target.colLeft + target.colWidth / 2 - GRIP / 2,
                            position: 'fixed',
                            top: target.tableTop - GRIP / 2,
                            width: GRIP,
                            zIndex: 40,
                        }}
                        type="button"
                    >
                        <GripHorizontal className="size-3.5" />
                    </button>
                </DropdownMenuTrigger>
                <DropdownMenuContent
                    align="start"
                    className="min-w-[176px]"
                    onCloseAutoFocus={(event) => {
                        event.preventDefault();
                        editor.commands.focus();
                    }}
                >
                    <DropdownMenuItem onSelect={() => focusTarget().addColumnBefore().run()}>
                        <ArrowLeft className="text-muted-foreground size-4 shrink-0" />
                        Insert left
                    </DropdownMenuItem>
                    <DropdownMenuItem onSelect={() => focusTarget().addColumnAfter().run()}>
                        <ArrowRight className="text-muted-foreground size-4 shrink-0" />
                        Insert right
                    </DropdownMenuItem>
                    <DropdownMenuSub>
                        <DropdownMenuSubTrigger>
                            <AlignLeft className="text-muted-foreground size-4 shrink-0" />
                            Align column
                        </DropdownMenuSubTrigger>
                        <DropdownMenuSubContent>
                            {ALIGN_OPTIONS.map((option) => (
                                <DropdownMenuItem
                                    key={option.value}
                                    onSelect={() => {
                                        focusTarget().run();
                                        setColumnAlign(editor, option.value);
                                    }}
                                >
                                    <option.icon className="text-muted-foreground size-4 shrink-0" />
                                    {option.label}
                                </DropdownMenuItem>
                            ))}
                        </DropdownMenuSubContent>
                    </DropdownMenuSub>
                    <DropdownMenuItem
                        onSelect={() => {
                            focusTarget().run();
                            clearLineContents(editor, 'column');
                        }}
                    >
                        <Eraser className="text-muted-foreground size-4 shrink-0" />
                        Clear contents
                    </DropdownMenuItem>
                    <DropdownMenuSeparator />
                    <DropdownMenuItem onSelect={() => focusTarget().deleteColumn().run()}>
                        <Delete className="text-muted-foreground size-4 shrink-0" />
                        Delete column
                    </DropdownMenuItem>
                </DropdownMenuContent>
            </DropdownMenu>

            <DropdownMenu
                onOpenChange={onMenuChange('row')}
                open={open === 'row'}
            >
                <DropdownMenuTrigger asChild>
                    <button
                        aria-label="Row actions"
                        className={GRIP_CLASS}
                        data-table-grip=""
                        style={{
                            height: GRIP,
                            left: target.tableLeft - GRIP / 2,
                            position: 'fixed',
                            top: target.rowTop + target.rowHeight / 2 - GRIP / 2,
                            width: GRIP,
                            zIndex: 40,
                        }}
                        type="button"
                    >
                        <GripVertical className="size-3.5" />
                    </button>
                </DropdownMenuTrigger>
                <DropdownMenuContent
                    align="start"
                    className="min-w-[176px]"
                    onCloseAutoFocus={(event) => {
                        event.preventDefault();
                        editor.commands.focus();
                    }}
                >
                    <RowHeaderToggleItem
                        cellPos={target.cellPos}
                        editor={editor}
                    />
                    <DropdownMenuSeparator />
                    <DropdownMenuItem onSelect={() => focusTarget().addRowBefore().run()}>
                        <ArrowUp className="text-muted-foreground size-4 shrink-0" />
                        Insert above
                    </DropdownMenuItem>
                    <DropdownMenuItem onSelect={() => focusTarget().addRowAfter().run()}>
                        <ArrowDown className="text-muted-foreground size-4 shrink-0" />
                        Insert below
                    </DropdownMenuItem>
                    <DropdownMenuItem
                        onSelect={() => {
                            focusTarget().run();
                            clearLineContents(editor, 'row');
                        }}
                    >
                        <Eraser className="text-muted-foreground size-4 shrink-0" />
                        Clear contents
                    </DropdownMenuItem>
                    <DropdownMenuSeparator />
                    <DropdownMenuItem onSelect={() => focusTarget().deleteRow().run()}>
                        <Delete className="text-muted-foreground size-4 shrink-0 -rotate-90" />
                        Delete row
                    </DropdownMenuItem>
                </DropdownMenuContent>
            </DropdownMenu>
        </>,
        document.body,
    );
}

// Its useEditorState subscription must stay inside the row-grip DropdownMenuContent: Radix Presence unmounts it
// on close, so the deliberately non-reactive useTableHandles hover path pays no per-transaction cost while idle.
// Reads header state by the hovered cell's position, not the caret — the grip opens without moving the selection.
function RowHeaderToggleItem({ cellPos, editor }: { cellPos: number; editor: Editor }) {
    const isHeaderRow = useEditorState({
        editor,
        selector: ({ editor }) => hasHeaderRow(editor, cellPos),
    });

    return (
        <DropdownMenuItem
            aria-checked={isHeaderRow}
            onSelect={(event) => {
                // preventDefault keeps the menu open so the switch flips in place; setTextSelection (not .focus())
                // seats the toggle in the hovered table without pulling DOM focus off the open menu.
                event.preventDefault();
                editor.chain().setTextSelection(cellPos).toggleHeaderRow().run();
            }}
            role="menuitemcheckbox"
        >
            <PanelTop className="text-muted-foreground size-4 shrink-0" />
            <span>Header row</span>
            <Switch
                checked={isHeaderRow}
                className="pointer-events-none ml-auto"
                tabIndex={-1}
            />
        </DropdownMenuItem>
    );
}

function useTableHandles(editor: Editor): TableHandlesController {
    const [target, setTarget] = useState<HoverTarget | null>(null);
    const [open, setOpen] = useState<OpenMenu>(null);
    const clearTimer = useRef<null | ReturnType<typeof setTimeout>>(null);
    const openRef = useRef<OpenMenu>(null);
    const cellPosRef = useRef<null | number>(null);

    useEffect(() => {
        const dom = editor.view.dom;

        const cancelClear = () => {
            if (clearTimer.current) {
                clearTimeout(clearTimer.current);
                clearTimer.current = null;
            }
        };

        const clearSoon = () => {
            cancelClear();
            clearTimer.current = setTimeout(() => {
                if (!openRef.current) {
                    cellPosRef.current = null;
                    setTarget(null);
                }
            }, 120);
        };

        // One document-level listener rather than the editor's own: the grip straddles the table border, so the
        // editor's mouseleave fires the instant the cursor crosses onto it and it would flicker away. Here we
        // classify each move — over a grip (keep), over a cell (reposition), over neither (hide) — so travelling
        // from a cell onto the grip never clears. `cellPosRef` skips re-renders while the cursor stays in one cell.
        const handleMove = (event: MouseEvent) => {
            if (openRef.current) {
                return;
            }

            const el = event.target as HTMLElement | null;

            if (el?.closest?.('[data-table-grip]')) {
                cancelClear();

                return;
            }

            const cell = el?.closest?.('td, th');

            if (!(cell instanceof HTMLElement) || !dom.contains(cell)) {
                if (cellPosRef.current !== null) {
                    clearSoon();
                }

                return;
            }

            cancelClear();

            let cellPos: number;

            try {
                cellPos = editor.view.posAtDOM(cell, 0);
            } catch {
                return;
            }

            if (cellPos === cellPosRef.current) {
                return;
            }

            const row = cell.closest('tr');
            const table = cell.closest('table');

            if (!row || !table) {
                return;
            }

            cellPosRef.current = cellPos;

            const cellRect = cell.getBoundingClientRect();
            const rowRect = row.getBoundingClientRect();
            const tableRect = table.getBoundingClientRect();

            setTarget({
                cellPos,
                colLeft: cellRect.left,
                colWidth: cellRect.width,
                rowHeight: rowRect.height,
                rowTop: rowRect.top,
                tableLeft: tableRect.left,
                tableTop: tableRect.top,
            });
        };

        document.addEventListener('mousemove', handleMove);

        // A grip captured before a scroll, resize, or edit points at a now-shifted position: scroll/resize move it
        // off the table visually, and a doc edit shifts every position after it, so a menu action would resolve
        // the stale cellPos against the current doc (wrong row/column, or an out-of-bounds RangeError in the
        // header selector). Drop the target in every case — it reappears on the next hover.
        const scrollParent = getEditorScrollParent(dom);

        const dropStaleTarget = () => {
            if (!openRef.current) {
                cellPosRef.current = null;
                setTarget(null);
            }
        };

        scrollParent.addEventListener('scroll', dropStaleTarget, { passive: true });
        window.addEventListener('resize', dropStaleTarget);
        editor.on('update', dropStaleTarget);

        return () => {
            cancelClear();
            document.removeEventListener('mousemove', handleMove);
            scrollParent.removeEventListener('scroll', dropStaleTarget);
            window.removeEventListener('resize', dropStaleTarget);
            editor.off('update', dropStaleTarget);
        };
    }, [editor]);

    const onMenuChange = (menu: 'column' | 'row') => (isOpen: boolean) => {
        const next = isOpen ? menu : null;

        // Write the ref synchronously, before setOpen. handleMove/dropStaleTarget are continuous-event handlers
        // reading openRef.current; mirroring `open` through a passive effect flushes a tick late, so a mousemove
        // in that window would retarget the grip to a neighbouring cell just as a menu opens on this one.
        openRef.current = next;
        setOpen(next);

        if (!isOpen) {
            cellPosRef.current = null;
            setTarget(null);
        }
    };

    return { onMenuChange, open, target };
}
