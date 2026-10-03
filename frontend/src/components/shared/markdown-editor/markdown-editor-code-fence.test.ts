import { Editor } from '@tiptap/core';
import { beforeAll, describe, expect, it } from 'vitest';

import { GROWTH_IF_QUADRATIC, slowdownWhenInputQuadruples } from '@/test-utils/cost-growth';

import { createMarkdownExtensions } from './markdown-editor-extensions';
import { markdownCodec, roundTrip, setupEditorJsdom } from './markdown-editor-test-setup';

beforeAll(setupEditorJsdom);

const nodeCount = (md: string, type: string): number => {
    const editor = new Editor({ content: md, contentType: 'markdown', extensions: createMarkdownExtensions() });
    const count = (editor.getJSON().content ?? []).filter((node) => node.type === type).length;
    editor.destroy();

    return count;
};

describe('code-block fence lengthening — a block documenting a ``` fence stays one block', () => {
    it('widens the outer fence past an inner ``` run so the block does not split on reload', () => {
        const src = '````\ninner\n```\nnested fence\n```\nafter\n````';
        const out = roundTrip(src);

        expect(nodeCount(out, 'codeBlock')).toBe(1);
        expect(roundTrip(out)).toBe(out);
        expect(out).toContain('nested fence');
    });

    it('sizes the fence in time that grows with a line of backticks, not with the square of it', () => {
        const { destroy, parse, serialize } = markdownCodec();
        const growth = slowdownWhenInputQuadruples(
            (size: number) => parse(`\`\`\`\nstart\n${'`'.repeat(size)}x\nend\n\`\`\``),
            serialize,
            16_000,
        );

        destroy();

        expect(growth).toBeLessThan(GROWTH_IF_QUADRATIC / 2);
    });

    it('leaves an ordinary code block on a 3-backtick fence', () => {
        expect(roundTrip('```js\nconst x = 1;\n```')).toBe('```js\nconst x = 1;\n```');
    });
});

// A tilde fence's info string may hold a backtick; a backtick fence's may not — re-emitting such a block on a
// backtick fence produces an opening line that is no longer a fence at all, and the code block is destroyed.
describe('code-block fence character — a language containing a backtick keeps its tilde fence', () => {
    it('keeps a language that opens with a tilde apart from the fence', () => {
        const src = '~~~ ~x\ncode\n~~~';

        expect(roundTrip(src)).toBe(src);
        expect(nodeCount(src, 'codeBlock')).toBe(1);
    });

    it('re-emits a tilde fence so the block survives reload', () => {
        const src = '~~~`js`\nconst x = 1;\n~~~';
        const out = roundTrip(src);

        expect(out).toBe(src);
        expect(nodeCount(out, 'codeBlock')).toBe(1);
        expect(roundTrip(out)).toBe(out);
    });

    it('widens the tilde fence past an inner ~~~ run', () => {
        const src = '~~~~`js`\ninner\n~~~\nafter\n~~~~';
        const out = roundTrip(src);

        expect(out).toBe(src);
        expect(nodeCount(out, 'codeBlock')).toBe(1);
        expect(roundTrip(out)).toBe(out);
    });
});

// The fence language must reach node.attrs.language (what CodeBlockLowlight highlights and the
// data-language caption names) and survive save unchanged — highlighting is a view-only decoration.
describe('code-block language attribute', () => {
    const language = (md: string): null | string => {
        const editor = new Editor({ content: md, contentType: 'markdown', extensions: createMarkdownExtensions() });
        const node = (editor.getJSON().content ?? []).find((candidate) => candidate.type === 'codeBlock');
        editor.destroy();

        return (node?.attrs?.language as null | string) ?? null;
    };

    it('captures the fence info string into attrs.language', () => {
        expect(language('```ts\nconst x = 1;\n```')).toBe('ts');
        expect(language('```\nplain\n```')).toBeNull();
    });

    it('round-trips the language on the fence for several languages', () => {
        for (const lang of ['ts', 'python', 'bash', 'json', 'go']) {
            expect(roundTrip(`\`\`\`${lang}\ncode\n\`\`\``)).toBe(`\`\`\`${lang}\ncode\n\`\`\``);
        }
    });
});

describe('code-block highlighting', () => {
    const highlightClasses = (md: string): string[] => {
        const element = document.createElement('div');
        const editor = new Editor({
            content: md,
            contentType: 'markdown',
            element,
            extensions: createMarkdownExtensions(),
        });
        const classes = [...editor.view.dom.querySelectorAll('pre code span')].flatMap((span) => [...span.classList]);
        editor.destroy();

        return [...new Set(classes)];
    };

    const SNIPPET = 'const total = items.map((item) => item.price).reduce((a, b) => a + b, 0);';

    it('highlights a fence that declares its language', () => {
        expect(highlightClasses(`\`\`\`javascript\n${SNIPPET}\n\`\`\``)).toContain('hljs-keyword');
    });

    // An undeclared fence must resolve to the plaintext grammar, not to highlightAuto: auto-detection re-runs
    // over every code block in the document on each keystroke inside one, and it is quadratic in document size.
    it('leaves a fence with no language unhighlighted rather than guessing', () => {
        expect(highlightClasses(`\`\`\`\n${SNIPPET}\n\`\`\``)).toEqual([]);
    });
});

// Pins serializeCodeSpan (markdown-editor-marked.ts) and the code-mark renderMarkdown override that zeroes the fence.
describe('inline code containing a backtick round-trips exactly', () => {
    it('preserves and converges a backtick-containing code span', () => {
        const out = roundTrip('x ``a `b` c`` y');

        expect(out).toContain('``a `b` c``');
        expect(roundTrip(out)).toBe(out);
    });

    it.each([
        ['one backtick around a doubled one', 'where `name = ``admin`` and 1` holds'],
        ['more backticks than the code needs', 'x ```` a`b ```` y'],
        ['one backtick, over two lines', 'chars: `< > + \\``\nAllowed: `img`, `svg`'],
        ['five backticks', 'x `````ab````` y'],
        ['the fence the writer would choose', 'x ``a `b` c`` y'],
    ])('keeps the fence a span was written with: %s', (_name, md) => {
        expect(roundTrip(md)).toBe(md);
    });

    it.each<[string, { fence?: number; text: string }[], string]>([
        [
            'code without a fence of its own inside it',
            [{ fence: 3, text: 'nmap' }, { text: 'X' }, { fence: 3, text: ' -sV' }],
            '`nmapX -sV`',
        ],
        [
            'code with another fence beside it',
            [
                { fence: 3, text: 'nmap' },
                { fence: 2, text: 'id' },
            ],
            '`nmapid`',
        ],
        [
            'every piece of it with the same fence',
            [
                { fence: 3, text: 'nmap' },
                { fence: 3, text: ' -sV' },
            ],
            '```nmap -sV```',
        ],
    ])('fences a span once: %s', (_name, pieces, saved) => {
        const editor = new Editor({
            content: {
                content: [
                    {
                        content: [
                            { text: 'Run ', type: 'text' },
                            ...pieces.map(({ fence, text }) => ({
                                marks: [{ attrs: { fence: fence ?? null }, type: 'code' }],
                                text,
                                type: 'text',
                            })),
                            { text: ' first', type: 'text' },
                        ],
                        type: 'paragraph',
                    },
                ],
                type: 'doc',
            },
            extensions: createMarkdownExtensions(),
        });
        const written = editor.getMarkdown();

        editor.destroy();

        expect(written).toBe(`Run ${saved} first`);
        expect(roundTrip(written)).toBe(written);
    });

    it('writes a span over two lines with a fence that holds its lines where the one it had does not', () => {
        const written = roundTrip('x ```a\nb``` y');

        expect(written).toBe('x `a\nb` y');
        expect(roundTrip(written)).toBe(written);
    });

    it.each([
        ['a run as long as the fence is typed into the span', 'x ```` a`b ```` y', '````', 'x `````  a`b ```` ````` y'],
        ['the same with a fence of one', 'where `a``b` holds', '`', 'where ``` a``b` ``` holds'],
    ])('writes the fence the code needs once %s', (_name, md, typed, saved) => {
        const editor = new Editor({ content: md, contentType: 'markdown', extensions: createMarkdownExtensions() });
        let end = 0;

        editor.state.doc.descendants((node, at) => {
            if (node.marks.some((mark) => mark.type.name === 'code')) {
                end = at + node.nodeSize;
            }
        });
        editor.chain().setTextSelection(end).insertContent(typed).run();

        const written = editor.getMarkdown();

        editor.destroy();

        expect(written).toBe(saved);
        expect(roundTrip(written)).toBe(written);
    });

    it.each([0, -1, 1.5, 65, '1'])('writes the fence the code needs where the one it carries is %j', (fence) => {
        const editor = new Editor({
            content: {
                content: [
                    {
                        content: [{ marks: [{ attrs: { fence }, type: 'code' }], text: 'a``b', type: 'text' }],
                        type: 'paragraph',
                    },
                ],
                type: 'doc',
            },
            extensions: createMarkdownExtensions(),
        });
        const written = editor.getMarkdown();

        editor.destroy();

        expect(written).toBe('```a``b```');
    });

    it.each([
        'a ``` `XSS` ``` b',
        'x ``` `key` = value ``` y',
        'run ``` echo `whoami` ``` now',
        'p ```` a```b ```` q',
        'a `simple` b',
        'use `{{.Host}}` here',
    ])('reaches a stable fixed point for %j', (src) => {
        const once = roundTrip(src);

        expect(roundTrip(once)).toBe(once);
    });
});

// A block created in the editor carries no author-chosen language, so it must serialize as a bare fence.
// `defaultLanguage` exists for the highlight plugin (`attrs.language || defaultLanguage`); leaking it into
// the attribute default writes ```plaintext into every prompt, template and knowledge document saved.
describe('a code block created in the editor', () => {
    const created = (): { language: null | string; markdown: string } => {
        const editor = new Editor({
            content: 'hello',
            contentType: 'markdown',
            extensions: createMarkdownExtensions(),
        });

        editor.commands.selectAll();
        editor.commands.toggleCodeBlock();

        const node = (editor.getJSON().content ?? []).find((candidate) => candidate.type === 'codeBlock');
        const result = {
            language: (node?.attrs?.language as null | string) ?? null,
            markdown: editor.getMarkdown(),
        };

        editor.destroy();

        return result;
    };

    it('serializes as a bare fence, not as ```plaintext', () => {
        const { language, markdown } = created();

        expect(language).toBeNull();
        expect(markdown).toContain('```\nhello\n```');
        expect(markdown).not.toContain('plaintext');
    });
});
