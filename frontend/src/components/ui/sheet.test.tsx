import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { useState } from 'react';
import { describe, expect, it } from 'vitest';

import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuItem,
    DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';

import { Sheet, SheetContent, SheetDescription, SheetTitle } from './sheet';

const SheetFromButton = () => {
    const [isOpen, setIsOpen] = useState(false);

    return (
        <>
            <button
                onClick={() => setIsOpen(true)}
                type="button"
            >
                Open details
            </button>
            <Sheet
                onOpenChange={setIsOpen}
                open={isOpen}
            >
                <SheetContent>
                    <SheetTitle>Details</SheetTitle>
                    <SheetDescription>Everything about this row.</SheetDescription>
                </SheetContent>
            </Sheet>
        </>
    );
};

const SheetFromDropdown = () => {
    const [isOpen, setIsOpen] = useState(false);

    return (
        <>
            <DropdownMenu>
                <DropdownMenuTrigger>Row actions</DropdownMenuTrigger>
                <DropdownMenuContent>
                    <DropdownMenuItem onSelect={() => setIsOpen(true)}>Details</DropdownMenuItem>
                </DropdownMenuContent>
            </DropdownMenu>
            <Sheet
                onOpenChange={setIsOpen}
                open={isOpen}
            >
                <SheetContent>
                    <SheetTitle>Details</SheetTitle>
                    <SheetDescription>Everything about this row.</SheetDescription>
                </SheetContent>
            </Sheet>
        </>
    );
};

describe('Sheet focus return', () => {
    it('returns focus to the button that opened the sheet', async () => {
        const user = userEvent.setup();
        render(<SheetFromButton />);

        const opener = screen.getByRole('button', { name: 'Open details' });
        await user.click(opener);
        await screen.findByRole('dialog');

        await user.keyboard('{Escape}');
        await waitFor(() => {
            expect(screen.queryByRole('dialog')).not.toBeInTheDocument();
        });

        expect(document.activeElement).toBe(opener);
    });

    it('returns focus to the menu trigger when the sheet was opened from a menu item', async () => {
        const user = userEvent.setup();
        render(<SheetFromDropdown />);

        const trigger = screen.getByRole('button', { name: 'Row actions' });
        await user.click(trigger);
        await user.click(await screen.findByRole('menuitem', { name: 'Details' }));
        await screen.findByRole('dialog');

        await user.keyboard('{Escape}');
        await waitFor(() => {
            expect(screen.queryByRole('dialog')).not.toBeInTheDocument();
        });

        expect(document.activeElement).toBe(trigger);
    });
});
