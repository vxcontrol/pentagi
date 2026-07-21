import { Editor } from '@tiptap/core';
import { Slice } from '@tiptap/pm/model';
import { beforeAll, describe, expect, it } from 'vitest';

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
        const evil = '[x]'.repeat(50_000);
        const started = performance.now();

        shouldParseMarkdownOnPaste(evil, '<span>x</span>', false);

        expect(performance.now() - started).toBeLessThan(150);
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
