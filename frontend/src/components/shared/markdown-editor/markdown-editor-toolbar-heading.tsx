import type { Editor } from '@tiptap/react';
import type { LucideIcon } from 'lucide-react';

import { Check, ChevronDown, Heading1, Heading2, Heading3, Heading4, Heading5, Heading6, Type } from 'lucide-react';

import { Button } from '@/components/ui/button';
import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuItem,
    DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/ui/tooltip';
import { cn } from '@/lib/utils';

export type HeadingLevel = 1 | 2 | 3 | 4 | 5 | 6;

interface HeadingOption {
    icon: LucideIcon;
    // Optical trim for the dropdown list ONLY (icons sit next to each other there): lucide's Type glyph is drawn
    // taller (16u) than the Heading glyphs (12u), so at the same 16px box it reads bigger — scale it down to match
    // without shrinking the box (which would misalign labels). The trigger shows one icon alone, so it stays full size.
    iconClassName?: string;
    label: string;
    value: 'paragraph' | HeadingLevel;
}

const OPTIONS: HeadingOption[] = [
    { icon: Heading1, label: 'Heading 1', value: 1 },
    { icon: Heading2, label: 'Heading 2', value: 2 },
    { icon: Heading3, label: 'Heading 3', value: 3 },
    { icon: Heading4, label: 'Heading 4', value: 4 },
    { icon: Heading5, label: 'Heading 5', value: 5 },
    { icon: Heading6, label: 'Heading 6', value: 6 },
    { icon: Type, iconClassName: 'scale-[0.75]', label: 'Text', value: 'paragraph' },
];

interface HeadingMenuProps {
    // 0 = paragraph / any non-heading block; 1-6 = the active heading level.
    activeLevel: 0 | HeadingLevel;
    disabled?: boolean;
    editor: Editor;
}

export function HeadingMenu({ activeLevel, disabled, editor }: HeadingMenuProps) {
    const isSelected = (value: HeadingOption['value']) =>
        value === 'paragraph' ? activeLevel === 0 : value === activeLevel;
    const active = OPTIONS.find((option) => isSelected(option.value)) ?? OPTIONS[0];
    const ActiveIcon = active?.icon ?? Type;

    const applyOption = (value: HeadingOption['value']) => {
        if (value === 'paragraph') {
            editor.chain().focus().setParagraph().run();

            return;
        }

        editor.chain().focus().toggleHeading({ level: value }).run();
    };

    return (
        <DropdownMenu>
            <Tooltip>
                <TooltipTrigger asChild>
                    <DropdownMenuTrigger asChild>
                        <Button
                            aria-label={`Text style: ${active?.label ?? 'Text'}`}
                            className="gap-0.5 px-1.5"
                            data-toolbar-item=""
                            disabled={disabled}
                            size="sm"
                            type="button"
                            variant="ghost"
                        >
                            <ActiveIcon />
                            <ChevronDown className="size-4 opacity-50" />
                        </Button>
                    </DropdownMenuTrigger>
                </TooltipTrigger>
                <TooltipContent>Text style</TooltipContent>
            </Tooltip>
            <DropdownMenuContent
                align="start"
                className="min-w-[140px]"
                onCloseAutoFocus={(event) => {
                    // Return focus to the editor caret (not the trigger button) so the user keeps typing.
                    event.preventDefault();
                    editor.commands.focus();
                }}
            >
                {OPTIONS.map((option) => (
                    <DropdownMenuItem
                        aria-checked={isSelected(option.value)}
                        key={option.value}
                        onSelect={() => applyOption(option.value)}
                        role="menuitemradio"
                    >
                        <option.icon className={cn('text-muted-foreground size-4 shrink-0', option.iconClassName)} />
                        <span>{option.label}</span>
                        {isSelected(option.value) ? <Check className="ml-auto size-4 shrink-0" /> : null}
                    </DropdownMenuItem>
                ))}
            </DropdownMenuContent>
        </DropdownMenu>
    );
}
