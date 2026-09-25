import { render, screen } from '@testing-library/react';
import { afterEach, describe, expect, it, vi } from 'vitest';

import { Table, TableBody, TableCell, TableRow } from './table';

const measureAs = (scrollWidth: number, clientWidth: number) => {
    vi.spyOn(Element.prototype, 'scrollWidth', 'get').mockReturnValue(scrollWidth);
    vi.spyOn(Element.prototype, 'clientWidth', 'get').mockReturnValue(clientWidth);
};

const renderTable = () =>
    render(
        <Table aria-label="Valid Access">
            <TableBody>
                <TableRow>
                    <TableCell>cell</TableCell>
                </TableRow>
            </TableBody>
        </Table>,
    );

const container = () => screen.getByTestId('table-container');

afterEach(() => {
    vi.restoreAllMocks();
});

describe('Table container', () => {
    it('takes a tab stop when the table is wider than the space it has', () => {
        measureAs(743, 606);
        renderTable();

        expect(container()).toHaveAttribute('tabindex', '0');
    });

    it('stays out of the tab order when the whole table fits', () => {
        measureAs(606, 606);
        renderTable();

        expect(container()).not.toHaveAttribute('tabindex');
    });

    it('names the scrollable region after the table it holds', () => {
        measureAs(743, 606);
        renderTable();

        expect(container()).toHaveAttribute('role', 'region');
        expect(container()).toHaveAccessibleName('Valid Access');
    });

    it('adds neither role nor name to a table that fits', () => {
        measureAs(606, 606);
        renderTable();

        expect(container()).not.toHaveAttribute('role');
        expect(container()).not.toHaveAttribute('aria-label');
    });
});
