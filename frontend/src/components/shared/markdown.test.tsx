import { render } from '@testing-library/react';
import { afterEach, describe, expect, it, vi } from 'vitest';

import Markdown from './markdown';

const tableMarkdown = [
    '| Name | Email | Token |',
    '| --- | --- | --- |',
    '| sample | sample@example.com | TOKEN{0000000000000000000000000000000000000000000000000000000000000000} |',
].join('\n');

afterEach(() => {
    vi.restoreAllMocks();
});

describe('Markdown', () => {
    it('wraps a table in a horizontally scrollable container so wide content cannot widen the message list', () => {
        const { container } = render(<Markdown>{tableMarkdown}</Markdown>);

        const table = container.querySelector('table');
        expect(table).not.toBeNull();
        expect(table?.parentElement?.className).toContain('overflow-x-auto');
    });

    it('keeps the table wrapped when search highlighting is active', () => {
        const { container } = render(<Markdown searchValue="sample">{tableMarkdown}</Markdown>);

        const table = container.querySelector('table');
        expect(table?.parentElement?.className).toContain('overflow-x-auto');
    });

    it('takes a named tab stop while the table is wider than the space it has', () => {
        vi.spyOn(Element.prototype, 'scrollWidth', 'get').mockReturnValue(743);
        vi.spyOn(Element.prototype, 'clientWidth', 'get').mockReturnValue(606);

        const { container } = render(<Markdown>{tableMarkdown}</Markdown>);

        const region = container.querySelector('table')?.parentElement;
        expect(region).toHaveAttribute('tabindex', '0');
        expect(region).toHaveAttribute('role', 'region');
        expect(region).toHaveAccessibleName('Table');
    });

    it('stays out of the tab order when the whole table fits', () => {
        vi.spyOn(Element.prototype, 'scrollWidth', 'get').mockReturnValue(606);
        vi.spyOn(Element.prototype, 'clientWidth', 'get').mockReturnValue(606);

        const { container } = render(<Markdown>{tableMarkdown}</Markdown>);

        expect(container.querySelector('table')?.parentElement).not.toHaveAttribute('tabindex');
    });
});
