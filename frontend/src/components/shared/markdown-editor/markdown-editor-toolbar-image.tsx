import type { Editor } from '@tiptap/react';

import { ImagePlus } from 'lucide-react';
import { useState } from 'react';

import { Button } from '@/components/ui/button';
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/popover';
import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/ui/tooltip';

import { ImageEditForm } from './markdown-editor-image-edit-form';

interface ImagePopoverProps {
    disabled?: boolean;
    editor: Editor;
}

export function ImagePopover({ disabled, editor }: ImagePopoverProps) {
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
                            aria-label="Insert image"
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
                <TooltipContent>Insert image</TooltipContent>
            </Tooltip>
            <PopoverContent
                align="start"
                className="w-80"
                onCloseAutoFocus={(event) => {
                    event.preventDefault();
                    editor.commands.focus();
                }}
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
