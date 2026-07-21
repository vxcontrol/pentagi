import { Editor } from '@tiptap/core';
import { beforeAll, describe, expect, it } from 'vitest';

import { createMarkdownExtensions } from './markdown-editor-extensions';
import { roundTrip, setupEditorJsdom } from './markdown-editor-test-setup';

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

    it('leaves an ordinary code block on a 3-backtick fence', () => {
        expect(roundTrip('```js\nconst x = 1;\n```')).toContain('```js');
    });
});

// KNOWN UPSTREAM LIMITATION — not fixable in this layer. @tiptap/markdown derives a MARK's opening/closing
// delimiter by rendering it around a fixed placeholder (getMarkOpening/getMarkClosing), so the `code` mark's
// real content is never seen at delimiter time and it always emits a SINGLE backtick. A code span whose
// content contains a backtick therefore collapses on save and never converges. A NODE (codeBlock, above) DOES
// receive its real content and is content-aware — hence the fence widening is fixable but this is not. Pinned so
// the gap stays visible: this test turns RED (the canary to revisit it) if a future @tiptap/markdown makes mark
// delimiters content-aware and the span starts round-tripping.
describe('inline code containing a backtick — known lossy on save (upstream mark-delimiter limitation)', () => {
    it('collapses a backtick-containing code span (does not round-trip)', () => {
        const out = roundTrip('x ``a `b` c`` y');

        expect(out).not.toContain('``a `b` c``');
        expect(roundTrip(out)).not.toBe(out);
    });
});
