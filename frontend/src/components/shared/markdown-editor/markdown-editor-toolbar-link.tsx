import type { Editor } from '@tiptap/react';

import { Link as LinkIcon } from 'lucide-react';
import { useState } from 'react';

import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/popover';
import { Toggle } from '@/components/ui/toggle';
import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/ui/tooltip';
import { useI18n } from '@/hooks/use-i18n';

import { returnFocusToEditor } from './markdown-editor-focus';
import { LinkEditForm } from './markdown-editor-link-edit-form';

interface LinkPopoverProps {
    disabled?: boolean;
    editor: Editor;
    isActive: boolean;
}

export function LinkPopover({ disabled, editor, isActive }: LinkPopoverProps) {
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
                        <Toggle
                            aria-label={t('markdownEditor.link')}
                            data-toolbar-item=""
                            disabled={disabled}
                            pressed={isActive}
                            size="sm"
                        >
                            <LinkIcon />
                        </Toggle>
                    </PopoverTrigger>
                </TooltipTrigger>
                <TooltipContent>{t('markdownEditor.link')}</TooltipContent>
            </Tooltip>
            <PopoverContent
                align="start"
                className="w-80 p-2"
                onCloseAutoFocus={returnFocusToEditor(editor)}
            >
                <LinkEditForm
                    editor={editor}
                    initialUrl={(editor.getAttributes('link').href as string | undefined) ?? ''}
                    isActive={isActive}
                    onDone={() => setOpen(false)}
                />
            </PopoverContent>
        </Popover>
    );
}
