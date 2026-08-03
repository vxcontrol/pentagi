import type { ReactNode } from 'react';

import { Button } from '@/components/ui/button';
import { Toggle } from '@/components/ui/toggle';
import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/ui/tooltip';

interface ToolbarButtonProps {
    children: ReactNode;
    disabled?: boolean;
    label: string;
    onClick: () => void;
    shortcut?: string;
}

interface ToolbarToggleProps {
    children: ReactNode;
    disabled?: boolean;
    label: string;
    onPressedChange: () => void;
    pressed: boolean;
    shortcut?: string;
}

interface ToolbarTooltipProps {
    children: ReactNode;
    label: string;
    shortcut?: string;
}

// One-shot actions (undo, insert…). A plain button — no aria-pressed, unlike a Toggle.
export function ToolbarButton({ children, disabled, label, onClick, shortcut }: ToolbarButtonProps) {
    return (
        <ToolbarTooltip
            label={label}
            shortcut={shortcut}
        >
            <Button
                aria-label={label}
                data-toolbar-item=""
                disabled={disabled}
                onClick={onClick}
                size="icon-sm"
                tabIndex={-1}
                type="button"
                variant="ghost"
            >
                {children}
            </Button>
        </ToolbarTooltip>
    );
}

// Marks/blocks with a genuine on/off state (bold, italic, lists…). Renders aria-pressed via Radix Toggle.
export function ToolbarToggle({ children, disabled, label, onPressedChange, pressed, shortcut }: ToolbarToggleProps) {
    return (
        <ToolbarTooltip
            label={label}
            shortcut={shortcut}
        >
            <Toggle
                aria-label={label}
                data-toolbar-item=""
                disabled={disabled}
                onPressedChange={onPressedChange}
                pressed={pressed}
                size="sm"
            >
                {children}
            </Toggle>
        </ToolbarTooltip>
    );
}

function ToolbarTooltip({ children, label, shortcut }: ToolbarTooltipProps) {
    return (
        <Tooltip>
            <TooltipTrigger asChild>{children}</TooltipTrigger>
            <TooltipContent className="flex items-center gap-2">
                <span>{label}</span>
                {shortcut ? (
                    <kbd className="bg-muted text-muted-foreground rounded px-1 font-mono text-[10px]">{shortcut}</kbd>
                ) : null}
            </TooltipContent>
        </Tooltip>
    );
}
