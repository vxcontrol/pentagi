import type { Editor } from '@tiptap/react';

import { ImagePlus } from 'lucide-react';
import { useState } from 'react';

import { Button } from '@/components/ui/button';
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/popover';
import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/ui/tooltip';
import { useI18n } from '@/hooks/use-i18n';

import { returnFocusToEditor } from './markdown-editor-focus';
import { ImageEditForm } from './markdown-editor-image-edit-form';

interface ImagePopoverProps {
    disabled?: boolean;
    editor: Editor;
}

export function ImagePopover({ disabled, editor }: ImagePopoverProps) {
    const { t } = useI18n();
    const [open, setOpen] = useState(false);

    return (
        <Popover
            onOpenChange={setOpen}
            open={open}
        >
            <Tooltip>
                <TooltipTrigger asChild>
                    <PopoverTrigger asChild>
                        <Button
                            aria-label={t('markdownEditor.insertImage')}
                            data-toolbar-item=""
                            disabled={disabled}
                            size="icon-sm"
                            type="button"
                            variant="ghost"
                        >
                            <ImagePlus />
                        </Button>
                    </PopoverTrigger>
                </TooltipTrigger>
                <TooltipContent>{t('markdownEditor.insertImage')}</TooltipContent>
            </Tooltip>
            <PopoverContent
                align="start"
                className="w-80"
                onCloseAutoFocus={returnFocusToEditor(editor)}
            >
                <ImageEditForm
                    editor={editor}
                    initialAlt=""
                    initialSrc=""
                    isEditing={false}
                    onDone={() => setOpen(false)}
                />
            </PopoverContent>
        </Popover>
    );
}
