import type { Editor } from '@tiptap/react';

import { Check, Trash2 } from 'lucide-react';
import { useId, useState } from 'react';

import { Input } from '@/components/ui/input';
import { InputGroup, InputGroupAddon, InputGroupButton, InputGroupInput } from '@/components/ui/input-group';
import { Label } from '@/components/ui/label';
import { useI18n } from '@/hooks/use-i18n';

import { normalizeImageSrc } from './markdown-editor-toolbar-url';

interface ImageEditFormProps {
    // Focus the URL input on mount. True when the user explicitly opened the form (toolbar button); false for the
    // on-image popover, which appears when an image is selected and must not steal focus.
    autoFocus?: boolean;
    editor: Editor;
    initialAlt: string;
    initialSrc: string;
    isEditing: boolean;
    onDone: () => void;
}

// Shared body of the image editor — the same src field + validation + alt used by the toolbar Insert-image popover
// AND the on-image popover (markdown-editor-image-handle.tsx). Seeds its own state from the initial props on mount,
// so consumers give it a fresh `key` per editing session.
export function ImageEditForm({
    autoFocus = true,
    editor,
    initialAlt,
    initialSrc,
    isEditing,
    onDone,
}: ImageEditFormProps) {
    const { t } = useI18n();
    const [src, setSrc] = useState(initialSrc);
    const [alt, setAlt] = useState(initialAlt);
    const srcId = useId();
    const altId = useId();
    const errorId = useId();

    const normalizedSrc = normalizeImageSrc(src);
    const isInvalid = src !== '' && normalizedSrc === null;

    const apply = () => {
        if (!normalizedSrc) {
            return;
        }

        if (isEditing) {
            editor
                .chain()
                .focus()
                .updateAttributes('image', { alt: alt || null, src: normalizedSrc })
                .run();
        } else {
            editor
                .chain()
                .focus()
                .setImage({ alt: alt || undefined, src: normalizedSrc })
                .run();
        }

        onDone();
    };

    const remove = () => {
        editor.chain().focus().deleteSelection().run();
        onDone();
    };

    const applyOnEnter = (event: React.KeyboardEvent) => {
        if (event.key === 'Enter') {
            event.preventDefault();
            apply();
        }
    };

    return (
        <div className="flex flex-col gap-3">
            <div className="flex flex-col gap-1.5">
                <Label htmlFor={srcId}>{t('markdownEditor.imageUrl')}</Label>
                <InputGroup>
                    <InputGroupInput
                        aria-describedby={isInvalid ? errorId : undefined}
                        aria-invalid={isInvalid}
                        autoFocus={autoFocus}
                        id={srcId}
                        onChange={(event) => setSrc(event.target.value)}
                        onKeyDown={applyOnEnter}
                        placeholder={t('markdownEditor.imageUrlPlaceholder')}
                        type="url"
                        value={src}
                    />
                    <InputGroupAddon
                        align="inline-end"
                        className="gap-0"
                    >
                        <InputGroupButton
                            aria-label={isEditing ? t('markdownEditor.applyImage') : t('markdownEditor.insertImage')}
                            disabled={src === '' || isInvalid}
                            onClick={apply}
                            size="icon-xs"
                        >
                            <Check />
                        </InputGroupButton>
                        {isEditing ? (
                            <InputGroupButton
                                aria-label={t('markdownEditor.removeImage')}
                                onClick={remove}
                                size="icon-xs"
                            >
                                <Trash2 />
                            </InputGroupButton>
                        ) : null}
                    </InputGroupAddon>
                </InputGroup>
            </div>
            <div className="flex flex-col gap-1.5">
                <Label htmlFor={altId}>{t('markdownEditor.imageAlt')}</Label>
                <Input
                    id={altId}
                    onChange={(event) => setAlt(event.target.value)}
                    onKeyDown={applyOnEnter}
                    placeholder={t('markdownEditor.imageAltPlaceholder')}
                    value={alt}
                />
            </div>
            {isInvalid ? (
                <p
                    className="text-destructive text-xs"
                    id={errorId}
                    role="alert"
                >
                    {t('markdownEditor.imageUrlInvalid')}
                </p>
            ) : null}
        </div>
    );
}
