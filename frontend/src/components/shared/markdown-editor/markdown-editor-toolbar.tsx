import type { Editor } from '@tiptap/react';

import { useEditorState } from '@tiptap/react';
import {
    Bold,
    Code,
    Italic,
    Minus,
    Quote,
    Redo,
    RemoveFormatting,
    SquareCode,
    Strikethrough,
    Undo,
} from 'lucide-react';
import { memo, useEffect, useRef } from 'react';

import { Separator } from '@/components/ui/separator';
import { TooltipProvider } from '@/components/ui/tooltip';
import { cn } from '@/lib/utils';

import type { HeadingLevel } from './markdown-editor-toolbar-heading';
import type { ListType } from './markdown-editor-toolbar-list';
import type { ColumnAlign } from './markdown-editor-toolbar-table';

import { hasHeaderRow } from './markdown-editor-table-commands';
import { ToolbarButton, ToolbarToggle } from './markdown-editor-toolbar-button';
import { HeadingMenu } from './markdown-editor-toolbar-heading';
import { ImagePopover } from './markdown-editor-toolbar-image';
import { LinkPopover } from './markdown-editor-toolbar-link';
import { ListMenu } from './markdown-editor-toolbar-list';
import { TableMenu } from './markdown-editor-toolbar-table';

interface MarkdownEditorToolbarProps {
    disabled?: boolean;
    editor: Editor;
}

const HEADING_LEVELS: HeadingLevel[] = [1, 2, 3, 4, 5, 6];

// A mouse wheel emits deltaY, which the browser applies to the nearest VERTICAL scroller — never to this
// overflow-x strip, so a wheel-mouse user can't reach overflowed controls (trackpads emit deltaX and already
// work). Translate a vertical-dominant wheel into horizontal scroll, releasing at the ends so the page can
// still scroll past the toolbar. Attached non-passively — React's onWheel is passive, so it can't preventDefault.
function useHorizontalWheelScroll(ref: React.RefObject<HTMLDivElement | null>) {
    useEffect(() => {
        const strip = ref.current;

        if (!strip) {
            return;
        }

        const handleWheel = (event: WheelEvent) => {
            if (strip.scrollWidth <= strip.clientWidth || Math.abs(event.deltaX) >= Math.abs(event.deltaY)) {
                return;
            }

            const atStart = strip.scrollLeft <= 0 && event.deltaY < 0;
            // Sub-pixel tolerance: at fractional zoom scrollLeft never exactly reaches scrollWidth - clientWidth,
            // so an exact `>=` would never see the end and would keep eating the wheel, locking page scroll.
            const atEnd = strip.scrollLeft + strip.clientWidth >= strip.scrollWidth - 1 && event.deltaY > 0;

            if (atStart || atEnd) {
                return;
            }

            strip.scrollLeft += event.deltaY;
            event.preventDefault();
        };

        strip.addEventListener('wheel', handleWheel, { passive: false });

        return () => strip.removeEventListener('wheel', handleWheel);
    }, [ref]);
}

// WAI-ARIA toolbar pattern: one Tab stop for the whole bar, Arrow/Home/End move between controls. Managed
// imperatively on `[data-toolbar-item]` so each control stays a dumb button. The item set is static, but a
// control's `disabled` toggling changes which items are in the roving set, so a MutationObserver on the
// `disabled` attribute re-seeds the single tab stop (menus swap their content in a body portal, not here, so
// no childList watch is needed). When a popover/dropdown is open its focus lives in that portal outside the
// bar, so activeElement isn't an item and Arrow keys fall through to the menu instead of being hijacked here.
function useToolbarRovingFocus(ref: React.RefObject<HTMLDivElement | null>) {
    useEffect(() => {
        const toolbar = ref.current;

        if (!toolbar) {
            return;
        }

        const enabledItems = () =>
            Array.from(toolbar.querySelectorAll<HTMLElement>('[data-toolbar-item]')).filter(
                (item) => !item.hasAttribute('disabled'),
            );

        const seedTabStop = () => {
            const items = enabledItems();
            const active = items.find((item) => item.tabIndex === 0) ?? items[0] ?? null;

            for (const item of toolbar.querySelectorAll<HTMLElement>('[data-toolbar-item]')) {
                item.tabIndex = item === active ? 0 : -1;
            }
        };

        seedTabStop();

        const observer = new MutationObserver(seedTabStop);
        observer.observe(toolbar, { attributeFilter: ['disabled'], subtree: true });

        const handleKeyDown = (event: KeyboardEvent) => {
            if (!['ArrowLeft', 'ArrowRight', 'End', 'Home'].includes(event.key)) {
                return;
            }

            const items = enabledItems();
            const index = items.indexOf(document.activeElement as HTMLElement);

            if (index === -1) {
                return;
            }

            event.preventDefault();

            const nextIndex =
                event.key === 'Home'
                    ? 0
                    : event.key === 'End'
                      ? items.length - 1
                      : event.key === 'ArrowRight'
                        ? (index + 1) % items.length
                        : (index - 1 + items.length) % items.length;
            const next = items[nextIndex];

            if (!next) {
                return;
            }

            for (const item of items) {
                item.tabIndex = item === next ? 0 : -1;
            }

            next.focus();
        };

        toolbar.addEventListener('keydown', handleKeyDown);

        return () => {
            toolbar.removeEventListener('keydown', handleKeyDown);
            observer.disconnect();
        };
    }, [ref]);
}

// memo: every keystroke re-renders the RHF-controlled parent; without it all Toggle/Button subtrees re-render
// per keystroke for referentially-stable props.
export const MarkdownEditorToolbar = memo(function MarkdownEditorToolbar({
    disabled,
    editor,
}: MarkdownEditorToolbarProps) {
    // tiptap v3's useEditor does NOT re-render on every transaction, so reading editor.isActive/can() inline
    // would leave the toolbar stale on selection-only moves (click into a bold word → Bold stays unlit).
    // useEditorState re-runs this selector per transaction and re-renders only when a value flips.
    const state = useEditorState({
        editor,
        selector: ({ editor }) => ({
            activeListType: (editor.isActive('bulletList')
                ? 'bullet'
                : editor.isActive('orderedList')
                  ? 'ordered'
                  : editor.isActive('taskList')
                    ? 'task'
                    : null) as ListType | null,
            canRedo: editor.can().redo(),
            canUndo: editor.can().undo(),
            columnAlign: (editor.getAttributes('tableHeader').align ??
                editor.getAttributes('tableCell').align ??
                null) as ColumnAlign | null,
            headingLevel: (HEADING_LEVELS.find((level) => editor.isActive('heading', { level })) ?? 0) as
                | 0
                | HeadingLevel,
            isBlockquote: editor.isActive('blockquote'),
            isBold: editor.isActive('bold'),
            isCode: editor.isActive('code'),
            isCodeBlock: editor.isActive('codeBlock'),
            isHeaderRow: hasHeaderRow(editor),
            isItalic: editor.isActive('italic'),
            isLink: editor.isActive('link'),
            isStrike: editor.isActive('strike'),
            isTable: editor.isActive('table'),
        }),
    });

    const toolbarRef = useRef<HTMLDivElement>(null);
    const stripRef = useRef<HTMLDivElement>(null);

    useToolbarRovingFocus(toolbarRef);
    useHorizontalWheelScroll(stripRef);

    return (
        <TooltipProvider delayDuration={400}>
            <div
                aria-label="Formatting"
                className={cn(
                    'bg-muted/40 order-last flex items-center border-t px-1 py-1',
                    'md:order-first md:border-t-0 md:border-b',
                    disabled && 'opacity-60',
                )}
                data-slot="markdown-editor-toolbar"
                ref={toolbarRef}
                role="toolbar"
            >
                {/* Scroll strip: controls never wrap — they scroll horizontally. min-w-0 lets this flex child
                    shrink below its content so the overflow actually engages; Undo/Redo live OUTSIDE it (below) so
                    they stay pinned right instead of scrolling away (ml-auto collapses to 0 once a flex row overflows). */}
                <div
                    className="flex min-w-0 flex-1 [scrollbar-width:none] items-center gap-0.5 overflow-x-auto overscroll-x-contain [&::-webkit-scrollbar]:hidden [&>*]:shrink-0"
                    ref={stripRef}
                >
                    <HeadingMenu
                        activeLevel={state.headingLevel}
                        disabled={disabled}
                        editor={editor}
                    />

                    <Separator
                        className="mx-1 h-5"
                        orientation="vertical"
                    />

                    <div
                        className="flex items-center gap-0.5"
                        role="group"
                    >
                        <ToolbarToggle
                            disabled={disabled}
                            label="Bold"
                            onPressedChange={() => editor.chain().focus().toggleBold().run()}
                            pressed={state.isBold}
                            shortcut="⌘B"
                        >
                            <Bold />
                        </ToolbarToggle>
                        <ToolbarToggle
                            disabled={disabled}
                            label="Italic"
                            onPressedChange={() => editor.chain().focus().toggleItalic().run()}
                            pressed={state.isItalic}
                            shortcut="⌘I"
                        >
                            <Italic />
                        </ToolbarToggle>
                        <ToolbarToggle
                            disabled={disabled}
                            label="Strikethrough"
                            onPressedChange={() => editor.chain().focus().toggleStrike().run()}
                            pressed={state.isStrike}
                        >
                            <Strikethrough />
                        </ToolbarToggle>
                        <ToolbarToggle
                            disabled={disabled}
                            label="Inline code"
                            onPressedChange={() => editor.chain().focus().toggleCode().run()}
                            pressed={state.isCode}
                        >
                            <Code />
                        </ToolbarToggle>
                        <LinkPopover
                            disabled={disabled}
                            editor={editor}
                            isActive={state.isLink}
                        />
                    </div>

                    <Separator
                        className="mx-1 h-5"
                        orientation="vertical"
                    />

                    <ListMenu
                        activeType={state.activeListType}
                        disabled={disabled}
                        editor={editor}
                    />

                    <Separator
                        className="mx-1 h-5"
                        orientation="vertical"
                    />

                    <TableMenu
                        columnAlign={state.columnAlign}
                        disabled={disabled}
                        editor={editor}
                        isActive={state.isTable}
                        isHeaderRow={state.isHeaderRow}
                    />

                    <Separator
                        className="mx-1 h-5"
                        orientation="vertical"
                    />

                    <div
                        className="flex items-center gap-0.5"
                        role="group"
                    >
                        <ToolbarToggle
                            disabled={disabled}
                            label="Blockquote"
                            onPressedChange={() => editor.chain().focus().toggleBlockquote().run()}
                            pressed={state.isBlockquote}
                        >
                            <Quote />
                        </ToolbarToggle>
                        <ToolbarToggle
                            disabled={disabled}
                            label="Code block"
                            onPressedChange={() => editor.chain().focus().toggleCodeBlock().run()}
                            pressed={state.isCodeBlock}
                        >
                            <SquareCode />
                        </ToolbarToggle>
                        <ImagePopover
                            disabled={disabled}
                            editor={editor}
                        />
                        <ToolbarButton
                            disabled={disabled}
                            label="Horizontal rule"
                            onClick={() => editor.chain().focus().setHorizontalRule().run()}
                        >
                            <Minus />
                        </ToolbarButton>
                    </div>

                    <Separator
                        className="mx-1 h-5"
                        orientation="vertical"
                    />

                    <ToolbarButton
                        disabled={disabled}
                        label="Clear formatting"
                        onClick={() => editor.chain().focus().unsetAllMarks().clearNodes().run()}
                    >
                        <RemoveFormatting />
                    </ToolbarButton>
                </div>

                <Separator
                    className="mx-1 h-5"
                    orientation="vertical"
                />

                <div
                    className="flex shrink-0 items-center gap-0.5"
                    role="group"
                >
                    <ToolbarButton
                        disabled={disabled || !state.canUndo}
                        label="Undo"
                        onClick={() => editor.chain().focus().undo().run()}
                        shortcut="⌘Z"
                    >
                        <Undo />
                    </ToolbarButton>
                    <ToolbarButton
                        disabled={disabled || !state.canRedo}
                        label="Redo"
                        onClick={() => editor.chain().focus().redo().run()}
                        shortcut="⇧⌘Z"
                    >
                        <Redo />
                    </ToolbarButton>
                </div>
            </div>
        </TooltipProvider>
    );
});
