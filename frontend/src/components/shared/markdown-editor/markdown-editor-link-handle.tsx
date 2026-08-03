import type { Editor } from '@tiptap/react';

import { getMarkRange, posToDOMRect } from '@tiptap/core';
import { useEffect, useRef, useState } from 'react';

import { Popover, PopoverAnchor, PopoverContent } from '@/components/ui/popover';

import { returnFocusToEditor } from './markdown-editor-focus';
import { LinkEditForm } from './markdown-editor-link-edit-form';
import { getEditorScrollParent } from './markdown-editor-styles';

interface LinkTarget {
    href: string;
    key: string;
    rect: { height: number; left: number; top: number; width: number };
}

// Shows the shared link editor anchored to the link under the caret. openOnClick:false (markdown-editor-extensions)
// seats the caret in a clicked link instead of navigating, so `selection.empty && isActive('link')` detects "the
// caret is on a link" for both mouse and keyboard. Overlay-only — it never mutates the doc until a form action
// runs, so it stays out of the byte round-trip (like TableHandles).
export function LinkHandle({ editor }: { editor: Editor }) {
    const { close, target } = useLinkHandle(editor);

    if (!target) {
        return null;
    }

    const { href, key, rect } = target;

    return (
        <Popover
            onOpenChange={(isOpen) => {
                if (!isOpen) {
                    close();
                }
            }}
            open
        >
            <PopoverAnchor asChild>
                <span
                    aria-hidden
                    className="pointer-events-none fixed"
                    style={{ height: rect.height, left: rect.left, top: rect.top, width: rect.width }}
                />
            </PopoverAnchor>
            <PopoverContent
                align="start"
                className="w-80 p-2"
                data-link-handle=""
                onCloseAutoFocus={returnFocusToEditor(editor)}
                onOpenAutoFocus={(event) => {
                    // The popover appears whenever the caret enters a link, so keep focus in the doc — stealing it
                    // would interrupt typing/navigation. The user clicks into the URL field only to actually edit.
                    event.preventDefault();
                }}
                side="bottom"
                sideOffset={6}
            >
                <LinkEditForm
                    autoFocus={false}
                    editor={editor}
                    initialUrl={href}
                    isActive
                    key={key}
                    onDone={close}
                />
            </PopoverContent>
        </Popover>
    );
}

function useLinkHandle(editor: Editor) {
    const [target, setTarget] = useState<LinkTarget | null>(null);
    // Escape only cleared `target`, and the next selectionUpdate re-targeted unconditionally while the caret
    // was still in the link — so a dismiss lasted exactly one keystroke. Remember which link was dismissed and
    // stay closed for it until the caret leaves, an edit shifts its start, or the toolbar opens it on purpose.
    const dismissedKeyRef = useRef<null | string>(null);

    useEffect(() => {
        const linkType = editor.schema.marks.link;

        const update = () => {
            const { selection } = editor.state;

            if (!linkType || !selection.empty || !editor.isActive('link')) {
                dismissedKeyRef.current = null;
                setTarget(null);

                return;
            }

            const range = getMarkRange(selection.$from, linkType);

            if (!range) {
                dismissedKeyRef.current = null;
                setTarget(null);

                return;
            }

            if (dismissedKeyRef.current === `${range.from}`) {
                return;
            }

            const rect = posToDOMRect(editor.view, range.from, range.to);
            const href = (editor.getAttributes('link').href as string | undefined) ?? '';

            // Key on the link's START only. Typing inside the link grows range.to every keystroke, and
            // <LinkEditForm key={key}> would then unmount/remount and re-seed its URL field from initialUrl,
            // silently discarding an in-progress edit. from is stable while the caret stays in the same link.
            setTarget({ href, key: `${range.from}`, rect });
        };

        const clear = () => setTarget(null);

        // An edit before the link moves its start, so the remembered key no longer identifies it.
        const forget = () => {
            dismissedKeyRef.current = null;
        };

        editor.on('selectionUpdate', update);
        editor.on('update', forget);

        // Fixed-positioned anchor goes stale on scroll/resize — drop it (it reappears on the next caret entry).
        const scrollParent = getEditorScrollParent(editor.view.dom);

        scrollParent.addEventListener('scroll', clear, { passive: true });
        window.addEventListener('resize', clear);

        return () => {
            editor.off('selectionUpdate', update);
            editor.off('update', forget);
            scrollParent.removeEventListener('scroll', clear);
            window.removeEventListener('resize', clear);
        };
    }, [editor]);

    return {
        close: () => {
            dismissedKeyRef.current = target?.key ?? null;
            setTarget(null);
        },
        target,
    };
}
