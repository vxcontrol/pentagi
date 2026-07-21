import type { Editor } from '@tiptap/react';
import type { LucideIcon } from 'lucide-react';

import { Check, ChevronDown, List, ListOrdered, ListTodo } from 'lucide-react';

import { Button } from '@/components/ui/button';
import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuItem,
    DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/ui/tooltip';
import { cn } from '@/lib/utils';

export type ListType = 'bullet' | 'ordered' | 'task';

interface ListOption {
    icon: LucideIcon;
    label: string;
    value: ListType;
}

const OPTIONS: ListOption[] = [
    { icon: List, label: 'Bullet list', value: 'bullet' },
    { icon: ListOrdered, label: 'Ordered list', value: 'ordered' },
    { icon: ListTodo, label: 'Task list', value: 'task' },
];

interface ListMenuProps {
    activeType: ListType | null;
    disabled?: boolean;
    editor: Editor;
}

export function ListMenu({ activeType, disabled, editor }: ListMenuProps) {
    const active = OPTIONS.find((option) => option.value === activeType);
    // The bullet-list icon doubles as the resting affordance, so the active background — not the glyph — is what
    // distinguishes "in a bullet list" from "no list".
    const TriggerIcon = active?.icon ?? List;

    const applyOption = (value: ListType) => {
        const chain = editor.chain().focus();

        if (value === 'bullet') {
            chain.toggleBulletList().run();
        } else if (value === 'ordered') {
            chain.toggleOrderedList().run();
        } else {
            chain.toggleTaskList().run();
        }
    };

    return (
        <DropdownMenu>
            <Tooltip>
                <TooltipTrigger asChild>
                    <DropdownMenuTrigger asChild>
                        <Button
                            aria-label={`List: ${active?.label ?? 'None'}`}
                            className={cn('gap-0.5 px-1.5', active && 'bg-accent text-accent-foreground')}
                            data-toolbar-item=""
                            disabled={disabled}
                            size="sm"
                            type="button"
                            variant="ghost"
                        >
                            <TriggerIcon />
                            <ChevronDown className="size-4 opacity-50" />
                        </Button>
                    </DropdownMenuTrigger>
                </TooltipTrigger>
                <TooltipContent>Lists</TooltipContent>
            </Tooltip>
            <DropdownMenuContent
                align="start"
                className="min-w-[160px]"
                onCloseAutoFocus={(event) => {
                    event.preventDefault();
                    editor.commands.focus();
                }}
            >
                {OPTIONS.map((option) => (
                    <DropdownMenuItem
                        aria-checked={activeType === option.value}
                        key={option.value}
                        onSelect={() => applyOption(option.value)}
                        role="menuitemradio"
                    >
                        <option.icon className="text-muted-foreground size-4 shrink-0" />
                        <span>{option.label}</span>
                        {activeType === option.value ? <Check className="ml-auto size-4 shrink-0" /> : null}
                    </DropdownMenuItem>
                ))}
            </DropdownMenuContent>
        </DropdownMenu>
    );
}
