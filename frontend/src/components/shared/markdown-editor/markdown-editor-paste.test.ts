import { Editor } from '@tiptap/core';
import { Slice } from '@tiptap/pm/model';
import { beforeAll, describe, expect, it, vi } from 'vitest';

import { GROWTH_IF_QUADRATIC, slowdownWhenInputQuadruples } from '@/test-utils/cost-growth';

import { createMarkdownExtensions } from './markdown-editor-extensions';
import { shouldParseMarkdownOnPaste } from './markdown-editor-paste';
import { setupEditorJsdom } from './markdown-editor-test-setup';

beforeAll(setupEditorJsdom);

describe('shouldParseMarkdownOnPaste — markdown-parse plain text, defer rich sources', () => {
    it('parses plain-text markdown (no HTML on the clipboard)', () => {
        expect(shouldParseMarkdownOnPaste('# Heading', '', false)).toBe(true);
    });

    it('ignores an empty / whitespace-only paste', () => {
        expect(shouldParseMarkdownOnPaste('   \n\t ', '', false)).toBe(false);
    });

    it('defers an in-editor copy (data-pm-slice) so ProseMirror keeps its fidelity', () => {
        expect(shouldParseMarkdownOnPaste('# x', '<p data-pm-slice="1 1 []">x</p>', false)).toBe(false);
    });

    it.each([
        '<ul><li>x</li></ul>',
        '<table><tbody><tr><td>x</td></tr></tbody></table>',
        '<h2>x</h2>',
        '<blockquote>x</blockquote>',
        '<pre>x</pre>',
    ])('defers rich HTML carrying block tags: %s', (html) => {
        expect(shouldParseMarkdownOnPaste('# x', html, false)).toBe(false);
    });

    it('still parses when the HTML is only styled-inline but the text IS markdown (VS Code markdown copy)', () => {
        expect(shouldParseMarkdownOnPaste('**x**', '<span style="color:red">**x**</span>', false)).toBe(true);
    });

    it.each([
        ['# heading', '<span># heading</span>'],
        ['- item one\n- item two', '<span>- item one<br>- item two</span>'],
        ['| a | b |', '<span>| a | b |</span>'],
        ['```\ncode\n```', '<span>```</span>'],
        ['> quoted', '<span>&gt; quoted</span>'],
        ['see [docs](https://x.dev)', '<span>see [docs](https://x.dev)</span>'],
        ['1. first', '<span>1. first</span>'],
    ])('parses markdown-looking text %s despite an inline-only HTML wrapper', (text, html) => {
        expect(shouldParseMarkdownOnPaste(text, html, false)).toBe(true);
    });

    it.each([
        ['hello world', '<b style="font-weight:700">hello world</b>'], // Google Docs bold paragraph
        ['see the docs', '<a href="https://example.com">see the docs</a>'], // inline link
        ['plain sentence', '<span style="font-style:italic">plain sentence</span>'], // Word italic
    ])('defers inline-formatted HTML whose text %s has no markdown (native parse keeps the marks)', (text, html) => {
        expect(shouldParseMarkdownOnPaste(text, html, false)).toBe(false);
    });

    it('keeps a paste inside a code context literal', () => {
        expect(shouldParseMarkdownOnPaste('# x', '', true)).toBe(false);
    });

    it('scans the markdown cues in linear time on a large bracket-heavy paste (ReDoS guard)', () => {
        const brackets = (count: number) => '[x]'.repeat(count);
        const decide = (text: string) => shouldParseMarkdownOnPaste(text, '<span>x</span>', false);

        expect(decide(brackets(50_000))).toBe(false);
        expect(slowdownWhenInputQuadruples(brackets, decide, 50_000)).toBeLessThan(GROWTH_IF_QUADRATIC / 2);
    });
});

describe('MarkdownPaste — the parsed payload matches load (same tuned markdown layer)', () => {
    const pasteEvent = (text: string, html = ''): ClipboardEvent =>
        ({
            clipboardData: {
                getData: (type: string) => (type === 'text/html' ? html : type === 'text/plain' ? text : ''),
            },
        }) as unknown as ClipboardEvent;

    const pasteHtml = (text: string, html = ''): string => {
        const editor = new Editor({ content: '', contentType: 'markdown', extensions: createMarkdownExtensions() });
        editor.view.someProp('handlePaste', (handler) => handler(editor.view, pasteEvent(text, html), Slice.empty));
        const out = editor.getHTML();
        editor.destroy();

        return out;
    };

    it('leaves numbered lines pasted into a code block to the code block', () => {
        const editor = new Editor({
            content: '```\nnotes\n```',
            contentType: 'markdown',
            extensions: createMarkdownExtensions(),
        });

        editor.commands.setTextSelection(2);

        const isHandled = editor.view.someProp('handlePaste', (handler) =>
            handler(editor.view, pasteEvent('1. nmap -sV 10.0.0.5\n2. curl http://example.com/'), Slice.empty),
        );
        const html = editor.getHTML();

        editor.destroy();

        expect(isHandled).toBeFalsy();
        expect(html).not.toContain('<ol');
        expect(html).toContain('notes');
    });

    it.each([
        ['a document that ends without one', 'first\n\nlast', '- x\n- y\n', 'first\n\n- x\n- y\n\nlast'],
        ['a document that ends with one', 'first\n\nlast\n', '# T\n\ntext\n\n\n', 'first\n\n# T\n\ntext\n\nlast\n'],
    ])('leaves the line breaks the pasted text ends with out of %s', (_name, md, pasted, saved) => {
        const editor = new Editor({ content: md, contentType: 'markdown', extensions: createMarkdownExtensions() });

        editor.chain().setTextSelection(6).splitBlock().run();
        editor.view.someProp('handlePaste', (handler) => handler(editor.view, pasteEvent(pasted), Slice.empty));

        const out = editor.getMarkdown();

        editor.destroy();

        expect(out).toBe(saved);
    });

    // jsdom has neither ClipboardEvent nor DataTransfer, which the paste rules build for their handlers.
    it.each([
        ['a URL that runs into an action', 'https://e.x/{{.P}}'],
        ['an address that begins inside one', '{{.User}}@example.com'],
    ])('pastes %s from a rich clipboard as the text it is', (_name, text) => {
        vi.stubGlobal('ClipboardEvent', class extends Event {});
        vi.stubGlobal(
            'DataTransfer',
            class {
                getData = () => '';
                setData = () => {};
            },
        );

        const editor = new Editor({ content: 'x', contentType: 'markdown', extensions: createMarkdownExtensions() });

        editor.chain().focus('end').insertContent(` ${text}`, { applyPasteRules: true }).run();

        const out = editor.getMarkdown();

        editor.destroy();

        vi.unstubAllGlobals();

        expect(out).toBe(`x ${text}`);
    });

    it('parses pasted block markdown into rich blocks (heading + list + table)', () => {
        const html = pasteHtml('# Heading\n\n- one\n- two\n\n| A | B |\n| --- | --- |\n| 1 | 2 |');

        expect(html).toContain('<h1>');
        expect(html).toContain('<ul>');
        expect(html).toContain('<table');
    });

    it('keeps _ and __ literal on paste, matching load (no <strong>/<em>)', () => {
        const html = pasteHtml('a __dunder__ and _under_ here');

        expect(html).not.toContain('<strong>');
        expect(html).not.toContain('<em>');
        expect(html).toContain('__dunder__');
    });

    it('defers to ProseMirror (no markdown parse) when the clipboard carries rich block HTML', () => {
        const html = pasteHtml('# fake', '<ul><li>web item</li></ul>');

        expect(html).not.toContain('<h1>');
    });

    it('defers inline-formatted HTML with markdown-free text (plugin inserts nothing)', () => {
        const html = pasteHtml('hello bold world', '<b>hello bold world</b>');

        expect(html).not.toContain('hello bold world');
    });
});
