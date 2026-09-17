import { render } from '@testing-library/react';
import { describe, expect, it } from 'vitest';

import Markdown, { preprocessMarkdownFences } from './markdown';

const tableMarkdown = [
    '| Name | Email | Token |',
    '| --- | --- | --- |',
    '| sample | sample@example.com | TOKEN{0000000000000000000000000000000000000000000000000000000000000000} |',
].join('\n');

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

    it('keeps nested info-string fences inside a single code block (issue #361)', () => {
        const markdown = [
            '```',
            'Example:',
            '```json',
            '{"flowId": 85}',
            '```',
            'After nested fence',
            '```',
        ].join('\n');

        const { container } = render(<Markdown>{markdown}</Markdown>);
        const pres = container.querySelectorAll('pre');

        expect(pres).toHaveLength(1);
        expect(pres[0]?.textContent).toContain('```json');
        expect(pres[0]?.textContent).toContain('{"flowId": 85}');
        expect(pres[0]?.textContent).toContain('After nested fence');
    });
});

describe('preprocessMarkdownFences', () => {
    it('widens the outer fence past nested ```json … ``` examples', () => {
        const input = [
            '```',
            'Example:',
            '```json',
            '{"a":1}',
            '```',
            'After',
            '```',
        ].join('\n');

        const out = preprocessMarkdownFences(input);

        expect(out.startsWith('````\n')).toBe(true);
        expect(out.endsWith('\n````')).toBe(true);
        expect(out).toContain('```json\n{"a":1}\n```\nAfter');
    });

    it('leaves ordinary sequential code blocks on 3-backtick fences', () => {
        const input = ['```', 'block1', '```', '', '```', 'block2', '```'].join('\n');

        expect(preprocessMarkdownFences(input)).toBe(input);
    });

    it('leaves a simple code block unchanged', () => {
        const input = ['```js', 'const x = 1;', '```'].join('\n');

        expect(preprocessMarkdownFences(input)).toBe(input);
    });

    it('preserves fence info strings when widening', () => {
        const input = ['```markdown', 'Use:', '```js', 'x', '```', 'done', '```'].join('\n');

        const out = preprocessMarkdownFences(input);

        expect(out.startsWith('````markdown\n')).toBe(true);
        expect(out).toContain('```js\nx\n```\ndone');
        expect(out.endsWith('\n````')).toBe(true);
    });
});
