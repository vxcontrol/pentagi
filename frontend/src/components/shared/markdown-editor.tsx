import type { Editor } from '@tiptap/react';

import { Placeholder } from '@tiptap/extensions';
import { history } from '@tiptap/pm/history';
import { EditorState } from '@tiptap/pm/state';
import { EditorContent, useEditor } from '@tiptap/react';
import StarterKit from '@tiptap/starter-kit';
import {
    Bold,
    Code,
    Code2,
    Heading1,
    Heading2,
    Heading3,
    Italic,
    Link as LinkIcon,
    List,
    ListOrdered,
    Minus,
    Quote,
    Redo,
    Strikethrough,
    Undo,
} from 'lucide-react';
import { type Ref, useCallback, useEffect, useImperativeHandle, useRef } from 'react';
import { Markdown } from 'tiptap-markdown';

import { Separator } from '@/components/ui/separator';
import { Toggle } from '@/components/ui/toggle';
import { cn } from '@/lib/utils';

export interface MarkdownEditorHandle {
    focus: () => void;
    getEditor: () => Editor | null;
}

export interface MarkdownEditorProps {
    autoFocus?: boolean;
    className?: string;
    contentClassName?: string;
    disabled?: boolean;
    onBlur?: () => void;
    onChange: (value: string) => void;
    placeholder?: string;
    showToolbar?: boolean;
    value: string;
}

// Replaces the editor's ProseMirror history plugin with a fresh instance,
// effectively clearing the undo/redo stack. We call this on initial mount
// (to discard the construction-time transactions from extensions like
// `trailingNode`) and after every external content sync (`setContent`
// itself is a transaction that lands in history). Crucially, we MUST NOT
// call this on every value change — user edits also propagate through
// `value`, and wiping then would break Ctrl+Z.
const resetUndoHistory = (editor: Editor): void => {
    const { state, view } = editor;
    const historyIndex = state.plugins.findIndex((plugin) => {
        const pluginKey = plugin.spec.key as undefined | { key?: string };

        return pluginKey?.key?.startsWith('history$') ?? false;
    });

    if (historyIndex < 0) {
        return;
    }

    const plugins = [...state.plugins];

    plugins[historyIndex] = history();

    // We MUST create a fresh `EditorState` rather than calling
    // `state.reconfigure({ plugins })`. PM's `history()` factory uses a
    // module-level singleton `PluginKey`, so every invocation returns a
    // plugin keyed `history$`. `reconfigure` sees the same key in the old
    // and new plugin arrays, decides "same plugin", and KEEPS the old
    // state — meaning the undo stack stays populated even after our swap.
    // `EditorState.create` has no prior state to carry over and initializes
    // every plugin from scratch, which is exactly what we want here.
    const newState = EditorState.create({
        doc: state.doc,
        plugins,
        schema: state.schema,
        selection: state.selection,
        storedMarks: state.storedMarks,
    });

    view.updateState(newState);

    // `view.updateState` bypasses PM's dispatch pipeline, so tiptap's
    // `transaction` subscription doesn't fire — the toolbar would keep
    // showing the stale `canUndo: true` value. Dispatch an empty
    // transaction to wake the React subscription. The transaction is
    // marked as non-historable so the freshly-empty stack stays empty.
    view.dispatch(newState.tr.setMeta('addToHistory', false));
};

interface MarkdownEditorToolbarProps {
    disabled?: boolean;
    editor: Editor;
}

function MarkdownEditor({
    autoFocus,
    className,
    contentClassName,
    disabled,
    onBlur,
    onChange,
    placeholder = 'Write something…',
    ref,
    showToolbar = true,
    value,
}: MarkdownEditorProps & { ref?: Ref<MarkdownEditorHandle> }) {
    const onChangeRef = useRef(onChange);
    const onBlurRef = useRef(onBlur);
    // Tracks the last markdown the editor reported externally. We compare
    // against this to suppress echo updates: tiptap-markdown can re-serialize
    // content slightly differently than the input string (whitespace/list
    // markers/hard breaks/etc.), and we don't want to flag those
    // normalizations as user edits — that would falsely flip RHF's
    // `isDirty` flag.
    //
    // The baseline is the editor's own serialized markdown, NOT the raw
    // input `value`. Comparing user edits against the canonical
    // (post-normalization) form is what makes the comparison correct.
    const lastEmittedRef = useRef<string>(value);

    // Tiptap dispatches transactions for the initial content during view
    // construction. Those `onUpdate` calls are echoes of the initial
    // parse, not real user edits — forwarding them to RHF would mark
    // the form dirty on mount. We start forwarding edits only after
    // `onCreate` has captured the canonical baseline.
    const isInitializedRef = useRef(false);

    // Tracks whether we have already cleared the undo stack on initial
    // mount. After the first sync useEffect run we set this to true; on
    // subsequent renders we only clear the stack when an external value
    // arrived (see `shouldExternalSync` below).
    const hasResetInitialHistoryRef = useRef(false);

    useEffect(() => {
        onChangeRef.current = onChange;
        onBlurRef.current = onBlur;
    }, [onChange, onBlur]);

    const editor = useEditor({
        content: value,
        editable: !disabled,
        extensions: [
            StarterKit.configure({
                codeBlock: { HTMLAttributes: { class: 'hljs' } },
            }),
            Placeholder.configure({
                emptyEditorClass: 'is-editor-empty',
                placeholder,
            }),
            Markdown.configure({
                breaks: true,
                html: false,
                linkify: true,
                tightLists: true,
                transformCopiedText: true,
                // Plain text pasted from the OS clipboard is left as-is.
                // With `transformPastedText: true`, a leading "- " (or
                // "1. ", "> ", etc.) would be parsed as markdown and
                // turn the paste into a list/blockquote — almost never
                // what the user wants for knowledge documents.
                transformPastedText: false,
            }),
        ],
        immediatelyRender: false,
        onBlur: () => onBlurRef.current?.(),
        onCreate: ({ editor: instance }) => {
            lastEmittedRef.current = instance.storage.markdown.getMarkdown();
            isInitializedRef.current = true;
        },
        onUpdate: ({ editor: instance }) => {
            if (!isInitializedRef.current) {
                return;
            }

            const next = instance.storage.markdown.getMarkdown();

            if (next === lastEmittedRef.current) {
                return;
            }

            lastEmittedRef.current = next;
            onChangeRef.current?.(next);
        },
    });

    useImperativeHandle(
        ref,
        () => ({
            focus: () => editor?.commands.focus(),
            getEditor: () => editor,
        }),
        [editor],
    );

    // Keep external value in sync (e.g. on form reset). Avoid resetting if
    // the editor already reflects the same markdown to keep cursor stable.
    useEffect(() => {
        if (!editor) {
            return;
        }

        const current = editor.storage.markdown.getMarkdown();
        const shouldExternalSync = current !== value;

        if (shouldExternalSync) {
            editor.commands.setContent(value, { emitUpdate: false });
        }

        // The editor's own serialized markdown is the canonical form —
        // subsequent user edits will be compared against this
        // representation, not the (possibly non-normalized) input
        // `value`. Storing the raw `value` here would cause the first
        // user keystroke to also re-emit the round-trip normalization.
        lastEmittedRef.current = editor.storage.markdown.getMarkdown();

        // Clear the undo stack:
        //   - on initial mount, to discard the construction-time
        //     transactions dispatched by extensions like `trailingNode`
        //     plus the initial `setContent` we just ran above;
        //   - on every external `setContent` (e.g. parent calls
        //     form.reset(serverDocument)), because that transaction is
        //     not a user edit either.
        //
        // We MUST NOT reset on every render: when the user types,
        // `value` changes too, and resetting then would erase the
        // user's own undo history.
        if (!hasResetInitialHistoryRef.current || shouldExternalSync) {
            hasResetInitialHistoryRef.current = true;
            resetUndoHistory(editor);
        }
    }, [editor, value]);

    useEffect(() => {
        if (!editor) {
            return;
        }

        editor.setEditable(!disabled);
    }, [editor, disabled]);

    useEffect(() => {
        if (autoFocus && editor) {
            editor.commands.focus('end');
        }
        // `autoFocus` is intentionally omitted from the deps — it's a
        // mount-time prop, equivalent to the native `<input autoFocus>`
        // attribute. We don't want a parent flipping `autoFocus` later
        // to steal focus back into the editor.
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [editor]);

    // While tiptap is initializing (`useEditor` returns `null` on the
    // first render with `immediatelyRender: false`), render a placeholder
    // with the same outer classes so the bounding box is already correct
    // and the parent layout doesn't jump when the editor mounts.
    if (!editor) {
        return (
            <div
                aria-busy="true"
                className={cn(
                    'border-input dark:bg-input/30 group/markdown-editor flex w-full flex-col overflow-hidden rounded-md border shadow-2xs outline-hidden transition-[color,box-shadow]',
                    disabled && 'pointer-events-none opacity-60',
                    className,
                )}
                data-slot="markdown-editor"
            />
        );
    }

    return (
        <div
            className={cn(
                'border-input dark:bg-input/30 group/markdown-editor flex w-full flex-col overflow-hidden rounded-md border shadow-2xs outline-hidden transition-[color,box-shadow]',
                'focus-within:ring-ring focus-within:ring-1',
                disabled && 'pointer-events-none opacity-60',
                className,
            )}
            data-slot="markdown-editor"
        >
            {showToolbar ? (
                <MarkdownEditorToolbar
                    disabled={disabled}
                    editor={editor}
                />
            ) : null}
            <EditorContent
                className={cn(
                    'prose prose-sm dark:prose-invert tiptap-content max-w-none min-w-0 flex-1 overflow-auto px-3 py-2',
                    '[&_.ProseMirror]:min-h-full [&_.ProseMirror]:outline-none',
                    contentClassName,
                )}
                editor={editor}
            />
        </div>
    );
}

function MarkdownEditorToolbar({ disabled, editor }: MarkdownEditorToolbarProps) {
    const handleSetLink = useCallback(() => {
        const previousUrl = editor.getAttributes('link').href as string | undefined;
        const url = window.prompt('URL', previousUrl ?? '');

        if (url === null) {
            return;
        }

        if (url === '') {
            editor.chain().focus().extendMarkRange('link').unsetLink().run();

            return;
        }

        editor.chain().focus().extendMarkRange('link').setLink({ href: url }).run();
    }, [editor]);

    return (
        <div
            className={cn(
                'bg-muted/40 flex flex-wrap items-center gap-0.5 border-b px-1 py-1',
                disabled && 'pointer-events-none opacity-60',
            )}
            data-slot="markdown-editor-toolbar"
        >
            <Toggle
                aria-label="Bold"
                onPressedChange={() => editor.chain().focus().toggleBold().run()}
                pressed={editor.isActive('bold')}
                size="sm"
                title="Bold (Ctrl+B)"
            >
                <Bold />
            </Toggle>
            <Toggle
                aria-label="Italic"
                onPressedChange={() => editor.chain().focus().toggleItalic().run()}
                pressed={editor.isActive('italic')}
                size="sm"
                title="Italic (Ctrl+I)"
            >
                <Italic />
            </Toggle>
            <Toggle
                aria-label="Strikethrough"
                onPressedChange={() => editor.chain().focus().toggleStrike().run()}
                pressed={editor.isActive('strike')}
                size="sm"
                title="Strikethrough"
            >
                <Strikethrough />
            </Toggle>
            <Toggle
                aria-label="Inline code"
                onPressedChange={() => editor.chain().focus().toggleCode().run()}
                pressed={editor.isActive('code')}
                size="sm"
                title="Inline code"
            >
                <Code />
            </Toggle>

            <Separator
                className="mx-1 h-5"
                orientation="vertical"
            />

            <Toggle
                aria-label="Heading 1"
                onPressedChange={() => editor.chain().focus().toggleHeading({ level: 1 }).run()}
                pressed={editor.isActive('heading', { level: 1 })}
                size="sm"
                title="Heading 1"
            >
                <Heading1 />
            </Toggle>
            <Toggle
                aria-label="Heading 2"
                onPressedChange={() => editor.chain().focus().toggleHeading({ level: 2 }).run()}
                pressed={editor.isActive('heading', { level: 2 })}
                size="sm"
                title="Heading 2"
            >
                <Heading2 />
            </Toggle>
            <Toggle
                aria-label="Heading 3"
                onPressedChange={() => editor.chain().focus().toggleHeading({ level: 3 }).run()}
                pressed={editor.isActive('heading', { level: 3 })}
                size="sm"
                title="Heading 3"
            >
                <Heading3 />
            </Toggle>

            <Separator
                className="mx-1 h-5"
                orientation="vertical"
            />

            <Toggle
                aria-label="Bullet list"
                onPressedChange={() => editor.chain().focus().toggleBulletList().run()}
                pressed={editor.isActive('bulletList')}
                size="sm"
                title="Bullet list"
            >
                <List />
            </Toggle>
            <Toggle
                aria-label="Ordered list"
                onPressedChange={() => editor.chain().focus().toggleOrderedList().run()}
                pressed={editor.isActive('orderedList')}
                size="sm"
                title="Ordered list"
            >
                <ListOrdered />
            </Toggle>

            <Separator
                className="mx-1 h-5"
                orientation="vertical"
            />

            <Toggle
                aria-label="Blockquote"
                onPressedChange={() => editor.chain().focus().toggleBlockquote().run()}
                pressed={editor.isActive('blockquote')}
                size="sm"
                title="Blockquote"
            >
                <Quote />
            </Toggle>
            <Toggle
                aria-label="Code block"
                onPressedChange={() => editor.chain().focus().toggleCodeBlock().run()}
                pressed={editor.isActive('codeBlock')}
                size="sm"
                title="Code block"
            >
                <Code2 />
            </Toggle>
            <Toggle
                aria-label="Link"
                onPressedChange={handleSetLink}
                pressed={editor.isActive('link')}
                size="sm"
                title="Insert link"
            >
                <LinkIcon />
            </Toggle>
            <Toggle
                aria-label="Horizontal rule"
                onPressedChange={() => editor.chain().focus().setHorizontalRule().run()}
                pressed={false}
                size="sm"
                title="Horizontal rule"
            >
                <Minus />
            </Toggle>

            <div className="ml-auto flex items-center gap-0.5">
                <Toggle
                    aria-label="Undo"
                    disabled={!editor.can().undo()}
                    onPressedChange={() => editor.chain().focus().undo().run()}
                    pressed={false}
                    size="sm"
                    title="Undo (Ctrl+Z)"
                >
                    <Undo />
                </Toggle>
                <Toggle
                    aria-label="Redo"
                    disabled={!editor.can().redo()}
                    onPressedChange={() => editor.chain().focus().redo().run()}
                    pressed={false}
                    size="sm"
                    title="Redo (Ctrl+Shift+Z)"
                >
                    <Redo />
                </Toggle>
            </div>
        </div>
    );
}

export { MarkdownEditor };
