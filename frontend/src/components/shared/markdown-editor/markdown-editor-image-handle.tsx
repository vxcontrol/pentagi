import type { Editor } from '@tiptap/react';

import { posToDOMRect } from '@tiptap/core';
import { NodeSelection } from '@tiptap/pm/state';
import { useEffect, useState } from 'react';

import { Popover, PopoverAnchor, PopoverContent } from '@/components/ui/popover';

import { ImageEditForm } from './markdown-editor-image-edit-form';
import { getEditorScrollParent } from './markdown-editor-styles';

interface ImageTarget {
    alt: string;
    key: number;
    rect: { height: number; left: number; top: number; width: number };
    src: string;
}

export function ImageHandle({ editor }: { editor: Editor }) {
    const { close, target } = useImageHandle(editor);

    if (!target) {
        return null;
    }

    const { alt, key, rect, src } = target;

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
                className="w-80"
                data-image-handle=""
                onCloseAutoFocus={(event) => {
                    event.preventDefault();
                    editor.commands.focus();
                }}
                onOpenAutoFocus={(event) => {
                    // The popover appears when an image is selected — keep focus in the doc so it doesn't steal it;
                    // the user clicks into the URL field only to actually edit.
                    event.preventDefault();
                }}
                side="bottom"
                sideOffset={6}
            >
                <ImageEditForm
                    autoFocus={false}
                    editor={editor}
                    initialAlt={alt}
                    initialSrc={src}
                    isEditing
                    key={key}
                    onDone={close}
                />
            </PopoverContent>
        </Popover>
    );
}

function useImageHandle(editor: Editor) {
    const [target, setTarget] = useState<ImageTarget | null>(null);

    useEffect(() => {
        const update = () => {
            const { selection } = editor.state;

            if (!(selection instanceof NodeSelection) || selection.node.type.name !== 'image') {
                setTarget(null);

                return;
            }

            const rect = posToDOMRect(editor.view, selection.from, selection.to);
            const { alt, src } = selection.node.attrs;

            setTarget({ alt: (alt as string) ?? '', key: selection.from, rect, src: (src as string) ?? '' });
        };

        const clear = () => setTarget(null);

        editor.on('selectionUpdate', update);

        // Fixed-positioned anchor goes stale on scroll/resize — drop it (it reappears on the next selection).
        const scrollParent = getEditorScrollParent(editor.view.dom);

        scrollParent.addEventListener('scroll', clear, { passive: true });
        window.addEventListener('resize', clear);

        return () => {
            editor.off('selectionUpdate', update);
            scrollParent.removeEventListener('scroll', clear);
            window.removeEventListener('resize', clear);
        };
    }, [editor]);

    return { close: () => setTarget(null), target };
}
