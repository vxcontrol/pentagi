import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { useState } from 'react';
import { describe, expect, it } from 'vitest';

import { ContextMenu, ContextMenuContent, ContextMenuItem, ContextMenuTrigger } from '@/components/ui/context-menu';
import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuItem,
    DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';

import { Dialog, DialogContent, DialogDescription, DialogTitle } from './dialog';

const DialogFromButton = () => {
    const [isOpen, setIsOpen] = useState(false);

    return (
        <>
            <button
                onClick={() => setIsOpen(true)}
                type="button"
            >
                Delete resource
            </button>
            <Dialog
                onOpenChange={setIsOpen}
                open={isOpen}
            >
                <DialogContent>
                    <DialogTitle>Delete resource</DialogTitle>
                    <DialogDescription>This cannot be undone.</DialogDescription>
                </DialogContent>
            </Dialog>
        </>
    );
};

const DialogFromDropdown = () => {
    const [isOpen, setIsOpen] = useState(false);

    return (
        <>
            <DropdownMenu>
                <DropdownMenuTrigger>Row actions</DropdownMenuTrigger>
                <DropdownMenuContent>
                    <DropdownMenuItem onSelect={() => setIsOpen(true)}>Delete</DropdownMenuItem>
                </DropdownMenuContent>
            </DropdownMenu>
            <Dialog
                onOpenChange={setIsOpen}
                open={isOpen}
            >
                <DialogContent>
                    <DialogTitle>Delete resource</DialogTitle>
                    <DialogDescription>This cannot be undone.</DialogDescription>
                </DialogContent>
            </Dialog>
        </>
    );
};

describe('Dialog focus return', () => {
    it('returns focus to the button that opened the dialog', async () => {
        const user = userEvent.setup();
        render(<DialogFromButton />);

        const opener = screen.getByRole('button', { name: 'Delete resource' });
        await user.click(opener);
        await screen.findByRole('dialog');

        await user.keyboard('{Escape}');
        await waitFor(() => {
            expect(screen.queryByRole('dialog')).not.toBeInTheDocument();
        });

        expect(document.activeElement).toBe(opener);
    });

    it('returns focus to the menu trigger when the dialog was opened from a menu item', async () => {
        const user = userEvent.setup();
        render(<DialogFromDropdown />);

        const trigger = screen.getByRole('button', { name: 'Row actions' });
        await user.click(trigger);
        await user.click(await screen.findByRole('menuitem', { name: 'Delete' }));
        await screen.findByRole('dialog');

        await user.keyboard('{Escape}');
        await waitFor(() => {
            expect(screen.queryByRole('dialog')).not.toBeInTheDocument();
        });

        expect(document.activeElement).toBe(trigger);
    });
});

const DialogFromContextMenu = () => {
    const [isOpen, setIsOpen] = useState(false);

    return (
        <>
            <ContextMenu>
                <ContextMenuTrigger>Row</ContextMenuTrigger>
                <ContextMenuContent>
                    <ContextMenuItem onSelect={() => setIsOpen(true)}>Delete</ContextMenuItem>
                </ContextMenuContent>
            </ContextMenu>
            <Dialog
                onOpenChange={setIsOpen}
                open={isOpen}
            >
                <DialogContent>
                    <DialogTitle>Delete resource</DialogTitle>
                    <DialogDescription>This cannot be undone.</DialogDescription>
                </DialogContent>
            </Dialog>
        </>
    );
};

describe('Dialog opened from a context menu', () => {
    it('returns focus to the context-menu trigger', async () => {
        const user = userEvent.setup();

        render(<DialogFromContextMenu />);

        const trigger = document.querySelector<HTMLElement>('[data-slot="context-menu-trigger"]')!;

        await user.pointer({ keys: '[MouseRight]', target: trigger });
        await user.click(await screen.findByRole('menuitem', { name: 'Delete' }));
        await screen.findByRole('dialog');

        await user.keyboard('{Escape}');
        await waitFor(() => {
            expect(screen.queryByRole('dialog')).not.toBeInTheDocument();
        });

        expect(document.activeElement).toBe(trigger);
    });
});
