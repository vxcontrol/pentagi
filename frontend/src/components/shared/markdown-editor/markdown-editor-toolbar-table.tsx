import type { Editor } from '@tiptap/react';

import {
    AlignLeft,
    ArrowDown,
    ArrowLeft,
    ArrowRight,
    ArrowUp,
    Check,
    ChevronDown,
    Delete,
    PanelTop,
    Table as TableIcon,
    Trash2,
} from 'lucide-react';

import { Button } from '@/components/ui/button';
import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuItem,
    DropdownMenuSeparator,
    DropdownMenuSub,
    DropdownMenuSubContent,
    DropdownMenuSubTrigger,
    DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { Switch } from '@/components/ui/switch';
import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/ui/tooltip';
import { cn } from '@/lib/utils';

import { ALIGN_OPTIONS, type ColumnAlign, setColumnAlign } from './markdown-editor-table-commands';

export type { ColumnAlign };

interface TableMenuProps {
    columnAlign: ColumnAlign | null;
    disabled?: boolean;
    editor: Editor;
    isActive: boolean;
    isHeaderRow: boolean;
}

export function TableMenu({ columnAlign, disabled, editor, isActive, isHeaderRow }: TableMenuProps) {
    const run = (fn: (chain: ReturnType<Editor['chain']>) => ReturnType<Editor['chain']>) =>
        fn(editor.chain().focus()).run();

    return (
        <DropdownMenu>
            <Tooltip>
                <TooltipTrigger asChild>
                    <DropdownMenuTrigger asChild>
                        <Button
                            aria-label="Table"
                            className={cn('gap-0.5 px-1.5', isActive && 'bg-accent text-accent-foreground')}
                            data-toolbar-item=""
                            disabled={disabled}
                            size="sm"
                            type="button"
                            variant="ghost"
                        >
                            <TableIcon />
                            <ChevronDown className="size-4 opacity-50" />
                        </Button>
                    </DropdownMenuTrigger>
                </TooltipTrigger>
                <TooltipContent>Table</TooltipContent>
            </Tooltip>
            <DropdownMenuContent
                align="start"
                className="min-w-[180px]"
                onCloseAutoFocus={(event) => {
                    event.preventDefault();
                    editor.commands.focus();
                }}
            >
                {isActive ? (
                    <>
                        <DropdownMenuItem
                            aria-checked={isHeaderRow}
                            onSelect={(event) => {
                                // Keep the menu open so the switch animates in place, like Notion's block menu.
                                event.preventDefault();
                                editor.chain().toggleHeaderRow().run();
                            }}
                            role="menuitemcheckbox"
                        >
                            <PanelTop className="text-muted-foreground size-4 shrink-0" />
                            <span>Header row</span>
                            <Switch
                                checked={isHeaderRow}
                                className="pointer-events-none ml-auto"
                                tabIndex={-1}
                            />
                        </DropdownMenuItem>
                        <DropdownMenuSeparator />
                        <DropdownMenuItem onSelect={() => run((chain) => chain.addRowBefore())}>
                            <ArrowUp className="text-muted-foreground size-4 shrink-0" />
                            Insert row above
                        </DropdownMenuItem>
                        <DropdownMenuItem onSelect={() => run((chain) => chain.addRowAfter())}>
                            <ArrowDown className="text-muted-foreground size-4 shrink-0" />
                            Insert row below
                        </DropdownMenuItem>
                        <DropdownMenuItem onSelect={() => run((chain) => chain.addColumnBefore())}>
                            <ArrowLeft className="text-muted-foreground size-4 shrink-0" />
                            Insert column left
                        </DropdownMenuItem>
                        <DropdownMenuItem onSelect={() => run((chain) => chain.addColumnAfter())}>
                            <ArrowRight className="text-muted-foreground size-4 shrink-0" />
                            Insert column right
                        </DropdownMenuItem>
                        <DropdownMenuSub>
                            <DropdownMenuSubTrigger>
                                <AlignLeft className="text-muted-foreground size-4 shrink-0" />
                                Align column
                            </DropdownMenuSubTrigger>
                            <DropdownMenuSubContent>
                                {ALIGN_OPTIONS.map((option) => (
                                    <DropdownMenuItem
                                        aria-checked={(columnAlign ?? 'left') === option.value}
                                        key={option.value}
                                        onSelect={() => setColumnAlign(editor, option.value)}
                                        role="menuitemradio"
                                    >
                                        <option.icon className="text-muted-foreground size-4 shrink-0" />
                                        <span>{option.label}</span>
                                        {(columnAlign ?? 'left') === option.value ? (
                                            <Check className="ml-auto size-4 shrink-0" />
                                        ) : null}
                                    </DropdownMenuItem>
                                ))}
                            </DropdownMenuSubContent>
                        </DropdownMenuSub>
                        <DropdownMenuSeparator />
                        <DropdownMenuItem onSelect={() => run((chain) => chain.deleteRow())}>
                            <Delete className="text-muted-foreground size-4 shrink-0 -rotate-90" />
                            Delete row
                        </DropdownMenuItem>
                        <DropdownMenuItem onSelect={() => run((chain) => chain.deleteColumn())}>
                            <Delete className="text-muted-foreground size-4 shrink-0" />
                            Delete column
                        </DropdownMenuItem>
                        <DropdownMenuSeparator />
                        <DropdownMenuItem onSelect={() => run((chain) => chain.deleteTable())}>
                            <Trash2 className="text-muted-foreground size-4 shrink-0" />
                            Delete table
                        </DropdownMenuItem>
                    </>
                ) : (
                    <DropdownMenuItem
                        onSelect={() => run((chain) => chain.insertTable({ cols: 3, rows: 3, withHeaderRow: true }))}
                    >
                        <TableIcon className="text-muted-foreground size-4 shrink-0" />
                        Insert table
                    </DropdownMenuItem>
                )}
            </DropdownMenuContent>
        </DropdownMenu>
    );
}
