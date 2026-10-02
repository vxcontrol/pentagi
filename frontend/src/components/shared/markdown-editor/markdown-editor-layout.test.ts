import type { JSONContent } from '@tiptap/core';

import { Editor } from '@tiptap/core';
import { beforeAll, describe, expect, it } from 'vitest';

import { GROWTH_IF_QUADRATIC, slowdownWhenInputQuadruples } from '@/test-utils/cost-growth';

import { createMarkdownExtensions } from './markdown-editor-extensions';
import { markdownCodec, roundTrip, setupEditorJsdom } from './markdown-editor-test-setup';

beforeAll(setupEditorJsdom);

const load = (content: JSONContent | string) =>
    new Editor({
        content,
        ...(typeof content === 'string' ? { contentType: 'markdown' as const } : {}),
        extensions: createMarkdownExtensions(),
    });

const kindsOf = (markdown: string): string[] => {
    const editor = load(markdown);
    const kinds = (editor.getJSON().content ?? []).map((block) => block.type ?? '');

    editor.destroy();

    return kinds;
};

const bullets = (text: string): JSONContent => ({
    content: [{ content: [{ content: [{ text, type: 'text' }], type: 'paragraph' }], type: 'listItem' }],
    type: 'bulletList',
});

// What the document is after `edit`, and what it is saved as.
const afterEdit = (content: JSONContent | string, edit: (editor: Editor) => void): string => {
    const editor = load(content);

    edit(editor);

    const saved = editor.getMarkdown();

    editor.destroy();

    return saved;
};

describe('a document is saved the way it was written', () => {
    it.each([
        ['text right under a heading', '# Title\ntext\n\nmore'],
        ['a list right under a heading', '## Findings\n- one\n- two'],
        ['a numbered list right under a heading', '## Steps\n3. three\n4. four'],
        ['a fence right under a heading', '### Run\n```sh\nid\n```'],
        ['a fence right under a label', '**Request:**\n```http\nGET / HTTP/1.1\n```'],
        ['text right under a fence', '```sh\nid\n```\nIt prints the uid.'],
        ['a heading right under text', 'intro\n## Next'],
        ['a list right under text', 'Checked:\n- one\n- two'],
        ['a numbered list right under text', 'Steps:\n1. one\n2. two'],
        ['a table right under a heading', '## Ports\n| p | s |\n|---|---|\n| 22 | ssh |'],
        ['a table right under text', 'Ports:\n| p | s |\n|---|---|\n| 22 | ssh |'],
        ['a quote right under text', 'He said:\n> no'],
        ['a starred rule right under text', 'text\n***\nmore'],
        ['a rule right under an image', '![a](http://x.test/i.png)\n***'],
        ['a list right under a list of another bullet', '- a\n* b\n+ c'],
        ['a numbered list right under a bullet list', '- a\n2. b'],
        ['a quote right under a list', '- a\n> q'],
        ['a dashed rule right under a quote', '> q\n---\nafter'],
        ['a dashed rule right under a table', '| a |\n|---|\n| 1 |\n---\nafter'],
        ['a task list right under a heading', '# H\n- [ ] a\n- [x] b'],
        ['a comment right under a list', '- one\n<!-- note -->\ntext'],
        ['a list with an empty item under a starred rule under text', 'text\n***\n- a\n-'],
        ['a blank line written with spaces in it', 'a\n   \nb'],
        ['two blank lines between two blocks', '# Title\n\n\ntext\n\n\n- a\n- b'],
        ['four and six blank lines between two blocks', 'a\n\n\n\n\nb\n\n\n\n\n\n\nc'],
        ['two blank lines between the blocks of a quote', '> a\n>\n>\n> b'],
        ['two blank lines between the blocks of a list item', '- a\n\n\n  b\n- c'],
        ['a blank line written with a tab in it', '# T\n \t\n- x'],
        ['a tab on the line under text, which is a line of it', 'para\n\t\n# H'],
        ['a blank line written with a tab in it, under a list', '- a\n\t\ntext'],
        ['a numbered list of checkboxes written two spaces in', '  1. [ ] a\n  2. [ ] b'],
        ['a list written two spaces in', 'Checked:\n\n  - one\n  - two\n\nafter'],
        ['a blank line in a list item written with more spaces than the item is indented by', '- b\n   \n  c'],
        [
            "a blank line between the blocks of an item, written with the item's indentation",
            '1. Run:\n   ```\n   id\n   ```\n   \n   It prints the uid.\n2. next',
        ],
        ['a blank line between two items, written with spaces', '1. one\n   - nested\n   \n2. two'],
        ['a blank line between two nested items, written with spaces', '- a\n  - b\n    \n  - c\n- d'],
        [
            'a list written two spaces in, with spaces past them on the blank line between its items',
            'Steps:\n\n  - a\n    \n  - b',
        ],
        [
            "an empty line of code in a list item, written with the item's indentation",
            '1. Run:\n   ```sh\n   id\n   \n   whoami\n   ```\n2. next',
        ],
        ['the same in a nested item', '- a\n  - b\n    ```sh\n    x\n    \n    y\n    ```\n- c'],
        ['the same with fewer spaces than the item is indented by', '1. step\n   ```\n   a\n  \n   b\n   ```'],
        [
            'the same in a list written two spaces in',
            'Steps:\n\n  1. Run:\n     ```\n     a\n     \n     b\n     ```\n     \n     Done.',
        ],
        ["the same under a fence on the marker's line", '1. ```\n   a\n   \n   b\n   ```'],
        ['a fence written three spaces in, with its code', 'Run:\n\n   ```sh\n   id\n   ```\n\nDone.'],
        ['a closing fence longer than the opening one, with spaces after it', '```\nid\n`````  \n\nDone.'],
        ['a fence written three spaces in, with an empty last line of code', '   ```\n   id\n\n   ```'],
        ['a fence with spaces before its language and after it', '```  sh  \nid\n```'],
        ['a tilde fence one space in, right under text', 'Run:\n ~~~\n id\n ~~~'],
        ['a fence in a quote, two spaces in', '> Run:\n>\n>   ```\n>   id\n>   ```'],
        ['code written by indentation', 'Run:\n\n    nmap -sV host\n     \n    curl host  \n\nDone.'],
        ['code indented with a tab', 'Run:\n\n\tnmap -sV host\n\tcurl host\n\nDone.'],
        ['indented code in a quote', '> Run:\n>\n>     nmap -sV host\n\nDone.'],
        ['indented code right under a heading and right above text', '# Run\n    nmap -sV host\nDone.'],
        ['indented code under a quote', '> Run:\n\n    nmap -sV host'],
        ['two indented blocks that a comment keeps apart', '    one\n\n<!-- c -->\n\n    two'],
        [
            'bullets written two spaces in under numbered steps, which is not deep enough to nest them',
            '1. Identify the source:\n  - which sensor\n  - which rule\n2. Check the host:\n  - look it up',
        ],
        ['a numbered list written three spaces in, with a wrapped item', '   1. one\n      more\n   2. two'],
        ['a nested list written one space past its item', '- a\n   - b\n   - c\n- d'],
        ['a one-line heading over a row of equals signs', 'Title\n=====\n\ntext'],
        ['a list with spaces after its last line', '- one\n- **Payload**: \n```\nid\n```\n\n1. a\n   - b  \n\ntext'],
        ['a heading with spaces after it', '### Pattern C: `os.popen()` \n\ntext\n\n## Next  \n- a'],
        ['a link around an image beside a bare address', '[![a](https://img.test/s.png)](http://h.test) http://x.test'],
        ['a link whose title holds a quote, around an image', '[![a](s.png)](http://h.test "say \\"hi\\"")'],
        ['a link whose target holds a space, around an image', '[![a](s.png)](<http://h.test/a b>)'],
        ['a link whose target holds a parenthesis, around an image', '[![a](s.png)](<http://h.test/a)b>)'],
        ["text two spaces in under a numbered list, which is not deep enough to be its item's", '1. step\n\n  Note.'],
        ['text under text two spaces in under a numbered list', '1. step\n\n  Note.\n\n   More.'],
        ['a tag two spaces in under a numbered list', '1. step\n\n  <b>Note</b> under the list.'],
        ['a link definition two spaces in under a numbered list', '1. step\n\n  [ref]: http://x.test'],
        ['two blank lines between a list of stars and a task list', '* item\n\n\n- [ ] task'],
        ['five blank lines under a task list', '- [ ] task\n\n\n\n\n\npara'],
        ['seven blank lines under a task list in a quote', '> - [ ] task\n\n\n\n\n\n\n\npara'],
        ['three blank lines between two task lists', '- [ ] t\n\n\n\n- [ ] u'],
        ['text two spaces in under an item that opens with code', '-    foo\n\n  bar'],
        ['text two spaces in under an empty item', '-\n\n  foo'],
        ['text three spaces in under an item numbered with two digits', '10. step\n\n   Note.'],
        ['text one space in under a bullet list', '- one\n\n two'],
        ['text as far in as the markers of the list above it', '  1. a\n  2. b\n\n  Return ONLY the JSON.'],
        ['four blank lines between a task list and a bullet list', '- [ ] a\n\n\n\n\n- b'],
        ['two blank lines between a bullet list and a numbered one', '- a\n\n\n1. b'],
        ['two blank lines between lists of two bullets', '* item\n\n\n- item'],
        ['a table whose header was written without a pipe', 'Name\n|---|\nvalue\n| `x|y` |'],
        ['a last paragraph of one ideographic space', 'x\n\n\u3000'],
        ['a link around an image', '[![build](https://img.test/b.svg)](https://ci.test/run) and text'],
        [
            'a link with a title around an image with one',
            'see [![a](https://img.test/a.png "shot")](https://h.test/ "open")',
        ],
        ['a one-line heading over a row of dashes', 'text\n\nSub\n---'],
        ['front matter of one key', '---\ntitle: Nightly job\n---\n\nBody'],
        ['a closing tag right under a table', '| a | b |\n|---|---|\n| 1 | 2 |\n</details>\ntext'],
        [
            'a line of text over a row of dashes and pipes, right under text',
            '{{ range .R }}\n| {{ .A }} | {{ .B }} |\n{{ else }}\n| - | - |\n{{ end }}',
        ],
        ['a tag line right under a tag line', '<id>{{.ID}}</id>\n<title>{{.Title}}</title>'],
        ['a closing tag line right under text', 'text\n</details>\nmore'],
        ['an opening tag right under a list', '- one\n- two\n<steps>\ntext'],
        ['blank lines kept where they were written', '# Title\n\ntext\n\n- one\n\n```\nx\n```'],
        ['a loose list', '1. one\n\n2. two\n\n3. three'],
        ['a loose task list', '- [ ] one\n\n- [ ] two'],
        ['a list loose in one place', '- a\n- b\n\n- c'],
        ['a blank line inside an item', '1. Run:\n\n   ```sh\n   id\n   ```\n2. next'],
        ['two blank lines between items', '- a\n\n\n- b'],
        ['three blank lines between blocks', 'a\n\n\n\nb'],
    ])('%s', (_name, md) => {
        expect(roundTrip(md)).toBe(md);
    });

    it.each([
        ['a row of three dashes', '| a | b |\n|---|---|\n| 1 | 2 |'],
        ['dashes as wide as the header', '| Tool | Purpose |\n|------|---------|\n| nmap | scan |'],
        ['dashes of any width', '| a | b |\n|--|-------|\n| 1 | 2 |'],
        ['alignment colons', '| L | C | R |\n|:--|:-:|--:|\n| a | b | c |'],
        ['an empty cell', '| a | b |\n|---|---|\n| 1 | |\n| | 2 |'],
        [
            'a table in a list item, its rows padded each in its own way',
            '- Ports:\n\n  | p | s |\n  |---|:-:|\n  |22|ssh|\n  | 80   | http |',
        ],
        [
            'rows padded each in its own way',
            '| Name | Value |\n|------|:-----:|\n| a    |   1   |\n| longer name | 2 |\n|c|3|',
        ],
        ['a row without its outer pipes and one with spaces after it', '| a | b |\n| --- | --- |\n1 | 2\n| 3 | 4 |  '],
        ['a table written three spaces in', 'Ports:\n\n   | p | s |\n   |---|---|\n   | 22 | ssh |'],
        ['a line of text among the rows', '| a | b |\n|---|---|\n| 1 | 2 |\nValues above are reserved.\n| 3 | 4 |'],
        ['two rows that hold the same cells', '| a | b |\n|---|---|\n|1|2|\n| 1 | 2 |\n|  1  |  2  |'],
        ['two spaces in an empty cell', '| a | b |\n|---|---|\n| 1 |  |\n|  | 2 |'],
        ['a cell that ends with a backslash', '| a | b |\n|---|---|\n| C:\\dir\\ | x |'],
        ['a pipe escaped in a cell', '| a | b |\n|---|---|\n| `x \\| y` | 2 |'],
        ['a pipe written bare in a code span', '| a | b |\n|---|---|\n| `x | y` | 2 |'],
        ['a Go action with a pipe', '| a | b |\n|---|---|\n| {{.X|upper}} | 2 |'],
        ['a padded table', '| a   | b   |\n| --- | --- |\n| 1   | 2   |'],
    ])('a table with %s', (_name, md) => {
        expect(roundTrip(md)).toBe(md);
    });

    it.each([
        ['starred bullets', '* a\n* b'],
        ['plus bullets', '+ a\n+ b'],
        ['a paren delimiter', '3) a\n4) b'],
        ['every item numbered one', '1. a\n1. b\n1. c'],
        ['numbers that are data', '22. ssh\n80. http\n443. https'],
        ['a count that jumps', '1. a\n2. b\n5. c\n6. d'],
        ['four spaces after a bullet', '-    four spaces\n-    b'],
        ['three spaces after a bullet', '*   a\n    ```sh\n    id\n    ```\n*   b'],
        ['numbers padded to one column', '9.  nine\n10. ten'],
        ['a starred rule in a dashed item', '- ***\n- b'],
        ['a dashed rule in a starred item', '* ---\n* b'],
    ])('a list with %s', (_name, md) => {
        expect(roundTrip(md)).toBe(md);
    });

    it.each([
        ['aligned to its parent', '1. step\n   - detail\n   This confirms it.\n2. next'],
        ['at the margin under a nested item', '1. step\n   - detail\n</steps>'],
        ['aligned to the middle of three levels', '- a\n  - b\n    - c\n  aligned to a'],
        ['four levels deep', '- a\n  - b\n    - c\n      - d\n      aligned to c'],
        ['at the margin under three levels', '- a\n  - b\n    - c\ntail'],
        ['after an indented one', '- first line\n  second line\nthird line'],
        ['followed by an indented one', '- **Repo**: x\ndef f(order_id):\n    user_id = 1\n        # deeper'],
    ])('a line left of its item, %s', (_name, md) => {
        expect(roundTrip(md)).toBe(md);
    });

    it.each([
        ['a tilde fence', '~~~python\nprint(1)\n~~~'],
        ['a four-backtick fence around fences', '````md\n```\nx\n```\n````'],
        ['a four-backtick fence around text', '````\ncode\n````'],
        ['a fence opener inside a fence', '```md\n```bash\nid\n```'],
        ['a rule of underscores', 'a\n\n___\n\nb'],
        ['a spaced rule', 'a\n\n- - -\n\nb'],
    ])('%s', (_name, md) => {
        expect(roundTrip(md)).toBe(md);
    });

    it.each([
        ['one line break', '# Title\n\n- a\n- b\n'],
        ['two line breaks', '# Big\n\n'],
        ['a closed fence and one line break', 'text\n\n```bash\necho hi\n```\n'],
        ['a fence nothing closes', 'text\n\n```bash\necho hi\n'],
        ['a fence nothing closes, without a line break', 'text\n\n```bash\necho hi'],
        ['an empty fence nothing closes', '```\n'],
        ['a fence nothing closes, with blank lines in it', '```\ncode\n\n\n'],
        ['a code span that holds one space', 'text\n\n` `'],
    ])('a document that ends with %s', (_name, md) => {
        expect(roundTrip(md)).toBe(md);
    });
});

describe('what was written gives way to an edit', () => {
    it('keeps the empty paragraph put above a block that stood right under text', () => {
        const saved = afterEdit('Label:\n```\ncode\n```\n\nnext', (editor) =>
            editor.commands.insertContentAt(7, { type: 'paragraph' }),
        );

        expect(saved).toBe('Label:\n\n\n\n```\ncode\n```\n\nnext');
        expect(kindsOf(saved)).toEqual(['paragraph', 'paragraph', 'codeBlock', 'paragraph']);
    });

    it('keeps the paragraph that stood right under a heading once it is empty', () => {
        const saved = afterEdit('# H\ntext\n\nnext', (editor) => editor.commands.deleteRange({ from: 4, to: 8 }));

        expect(saved).toBe('# H\n\n\n\nnext');
        expect(kindsOf(saved)).toEqual(['heading', 'paragraph', 'paragraph']);
    });

    it('puts a blank line above a dashed rule once the heading above it is text', () => {
        const saved = afterEdit('# H\n---\nafter', (editor) => editor.chain().setTextSelection(1).setParagraph().run());

        expect(saved).toBe('H\n\n---\nafter');
        expect(kindsOf(saved)).toEqual(['paragraph', 'horizontalRule', 'paragraph']);
    });

    it.each([
        [
            'a list whose first item is emptied under a starred rule that stands right under text',
            'Intro line\n***\n- one\n- two',
            (editor: Editor) => editor.commands.deleteRange({ from: 16, to: 19 }),
            'Intro line\n***\n\n-\n- two',
            ['paragraph', 'horizontalRule', 'bulletList'],
        ],
        [
            'a two-line heading made under a starred rule that stands right under text',
            'Intro line\n***\nTitle line\nsecond line\n\nBody',
            (editor: Editor) => editor.chain().setTextSelection(14).setHeading({ level: 1 }).run(),
            'Intro line\n***\n\nTitle line\nsecond line\n===\n\nBody',
            ['paragraph', 'horizontalRule', 'heading', 'paragraph'],
        ],
        [
            'a block under a definition that has lost its target',
            '[ref]: http://x\n```\ncode\n```',
            (editor: Editor) => editor.commands.deleteRange({ from: 7, to: 16 }),
            '[ref]:\n\n```\ncode\n```',
            ['paragraph', 'codeBlock'],
        ],
        [
            'the fence of a block nothing closes once its code is deleted',
            'Label:\n```sh\nls -la',
            (editor: Editor) => editor.commands.deleteRange({ from: 9, to: 15 }),
            'Label:\n```sh\n\n```',
            ['paragraph', 'codeBlock'],
        ],
    ])('keeps %s a block of its own', (_name, md, edit, saved, kinds) => {
        expect(afterEdit(md, edit)).toBe(saved);
        expect(kindsOf(saved)).toEqual(kinds);
    });

    // marked reads text, a `***` rule right under it and a row of dashes as one heading.
    it.each([
        ['text', { content: [{ text: 'x', type: 'text' }], type: 'paragraph' }, 'x\n***\n\n---', '- x\n  ***\n\n  ---'],
        [
            'an image',
            { attrs: { alt: 'a', src: 'http://x.test/i.png' }, type: 'image' },
            '![a](http://x.test/i.png)\n***\n\n---',
            '- ![a](http://x.test/i.png)\n  ***\n\n  ---',
        ],
    ])('puts a blank line above a dashed rule under a starred one that stands right under %s', (...row) => {
        const [, first, saved, savedInItem] = row;
        const rule = (attrs: Record<string, unknown>) => ({ attrs, type: 'horizontalRule' });
        const blocks = [first, rule({ gap: 0, rule: '***' }), rule({ gap: 0 })];
        const item = { content: [{ content: blocks, type: 'listItem' }], type: 'bulletList' };

        expect(afterEdit({ content: blocks, type: 'doc' }, () => {})).toBe(saved);
        expect(kindsOf(saved).slice(1)).toEqual(['horizontalRule', 'horizontalRule']);
        expect(afterEdit({ content: [item], type: 'doc' }, () => {})).toBe(savedInItem);
    });

    it('puts a blank line above a list once its first item is empty', () => {
        const saved = afterEdit('Checked:\n- one\n- two', (editor) =>
            editor.commands.deleteRange({ from: 13, to: 16 }),
        );

        expect(saved).toBe('Checked:\n\n-\n- two');
        expect(kindsOf(saved)).toEqual(['paragraph', 'bulletList']);
    });

    it('puts a blank line above a numbered list that no longer starts at one', () => {
        const saved = afterEdit('Steps:\n1. a\n5. b', (editor) => editor.commands.deleteRange({ from: 9, to: 14 }));

        expect(saved).toBe('Steps:\n\n5. b');
        expect(kindsOf(saved)).toEqual(['paragraph', 'orderedList']);
    });

    it('puts a blank line above a list whose first line is a delimiter row for the text above', () => {
        const saved = afterEdit('a\n- | -\n- x', (editor) =>
            editor.chain().setTextSelection(2).insertContent(' | b').run(),
        );

        expect(saved).toBe('a | b\n\n- | -\n- x');
        expect(kindsOf(saved)).toEqual(['paragraph', 'bulletList']);
    });

    it.each([
        [
            'a heading of two lines under text',
            [
                { content: [{ text: 'text', type: 'text' }], type: 'paragraph' },
                {
                    attrs: { gap: 0, level: 1 },
                    content: [{ text: 'a', type: 'text' }, { type: 'hardBreak' }, { text: 'b', type: 'text' }],
                    type: 'heading',
                },
            ],
            'text\n\na  \nb\n===',
        ],
        [
            'a quote under a quote',
            [
                { content: [{ content: [{ text: 'a', type: 'text' }], type: 'paragraph' }], type: 'blockquote' },
                {
                    attrs: { gap: 0 },
                    content: [{ content: [{ text: 'b', type: 'text' }], type: 'paragraph' }],
                    type: 'blockquote',
                },
            ],
            '> a\n\n> b',
        ],
        [
            'text under text',
            [
                { content: [{ text: 'one', type: 'text' }], type: 'paragraph' },
                { attrs: { gap: 0 }, content: [{ text: 'two', type: 'text' }], type: 'paragraph' },
            ],
            'one\n\ntwo',
        ],
    ])('puts a blank line above %s that claims to stand right under it', (_name, content, saved) => {
        const editor = load({ content, type: 'doc' });
        const md = editor.getMarkdown();

        editor.destroy();

        expect(md).toBe(saved);
    });

    it.each([
        ['with no blank line written between them', [bullets('a'), bullets('b')], '- a\n- b'],
        ['with one', [bullets('a'), { ...bullets('b'), attrs: { gap: 1 } }], '- a\n- b'],
        ['of two bullets', [bullets('a'), { ...bullets('b'), attrs: { marker: '*' } }], '- a\n\n* b'],
    ])('writes two lists an edit left side by side, %s, as the lists they read as', (_name, content, saved) => {
        const editor = load({ content, type: 'doc' });
        const md = editor.getMarkdown();

        editor.destroy();

        expect(md).toBe(saved);
    });

    it.each([
        ['bullets', '- a\n- b\n- c', '- ab\n- c'],
        ['numbers', '1. a\n2. b\n3. c', '1. ab\n2. c'],
    ])('writes %s as one list once Backspace has joined an item into the one above', (_name, md, saved) => {
        const out = afterEdit(md, (editor) => {
            const press = () =>
                editor.view.someProp('handleKeyDown', (handle) =>
                    handle(editor.view, new KeyboardEvent('keydown', { key: 'Backspace' })),
                );

            editor.commands.setTextSelection(8);
            press();
            press();
        });

        expect(out).toBe(saved);
        expect(roundTrip(saved)).toBe(saved);
    });

    it('writes two lists as one once the text between them is deleted', () => {
        const saved = afterEdit('A\n\n- x\n\nmid\n\n- y', (editor) =>
            editor.commands.deleteRange({ from: 10, to: 15 }),
        );

        expect(saved).toBe('A\n\n- x\n- y');
        expect(roundTrip(saved)).toBe(saved);
    });

    it('writes a line of text right under a table as a row of it once the line is edited', () => {
        const md = '| a | b |\n|---|---|\nback\\';
        const saved = afterEdit(md, (editor) => {
            let at = 0;

            editor.state.doc.descendants((node, position) => {
                at = node.text === 'back\\' ? position : at;
            });
            editor.chain().setTextSelection(at).insertContent('way ').run();
        });

        expect(roundTrip(md)).toBe(md);
        expect(saved).toBe('| a | b |\n|---|---|\n| way back\\ | |');
        expect(roundTrip(saved)).toBe(saved);
    });

    it('writes a list at the margin when one of its lines stood left of its markers', () => {
        const saved = roundTrip('  1.  A paragraph\nwith two lines.\n\n      > A quote.');

        expect(saved).toBe('1.  A paragraph\nwith two lines.\n\n    > A quote.');
        expect(roundTrip(saved)).toBe(saved);
    });

    it.each([
        ['two spaces under a bullet list', 'bulletList', 2, '- a\n\n* b'],
        ['three spaces under a numbered list', 'orderedList', 3, '1. a\n\n* b'],
        ['one space under a bullet list, which nests nothing', 'bulletList', 1, '- a\n\n * b'],
        ['two spaces under a numbered list, which nest nothing', 'orderedList', 2, '1. a\n\n  * b'],
    ])(
        'carries the spaces a list was written after only where they would not nest it: %s',
        (_name, type, indent, saved) => {
            const indented = { ...bullets('b'), attrs: { indent, marker: '*' } };

            expect(afterEdit({ content: [{ ...bullets('a'), type }, indented], type: 'doc' }, () => {})).toBe(saved);
            expect(roundTrip(saved)).toBe(saved);
        },
    );

    it('carries the spaces a list was written after when the list under it is another', () => {
        const indented = { ...bullets('b'), attrs: { indent: 2, marker: '*' } };

        expect(afterEdit({ content: [indented, bullets('a')], type: 'doc' }, () => {})).toBe('  * b\n\n- a');
    });

    it.each([
        ['two blank lines of which one holds spaces', 'a\n\n  \nb', 'a\n\n\nb'],
        ['a blank line that holds a non-breaking space', 'a\n\u00a0\nb', 'a\n\u00a0\nb'],
        ['a list whose markers stand at different columns', ' - a\n  - b', '- a\n- b'],
        [
            'a link around an image and its caption, as a link around each',
            '[![a](https://img.test/a.png) caption](https://h.test/)',
            '[![a](https://img.test/a.png)](https://h.test/) [caption](https://h.test/)',
        ],
        [
            'a list written two spaces in, with only those on the blank line between its items',
            'Steps:\n\n  - a\n  \n  - b',
            'Steps:\n\n  - a\n\n  - b',
        ],
        ['a tab on the blank line under a heading in a list item', '- # H\n\t\n  b', '- # H\n\n  b'],
        [
            'a tab on the line under a table, which is an empty row of it',
            '| a | b |\n|---|---|\n| 1 | 2 |\n\t\n# H',
            '| a | b |\n|---|---|\n| 1 | 2 |\n| | |\n# H',
        ],
        [
            'empty lines of code in a list item, of which only one holds spaces',
            '1. step\n   ```\n   a\n   \n\n   b\n   ```',
            '1. step\n   ```\n   a\n\n\n   b\n   ```',
        ],
        [
            'text as far in as the content of the list above it, which goes to the margin',
            ' - a\n  - b\n\n  text',
            '- a\n- b\n\ntext',
        ],
        [
            'an empty cell of a row that is written anew, with the two spaces the table writes one with',
            '| a | b | c |\n|---|---|---|\n| x |  | &lt;y&gt; |',
            '| a | b | c |\n|---|---|---|\n| x |  | <y> |',
        ],
        [
            'a list with a task item written two spaces in',
            '  - plain item\n  - [ ] task item',
            '- plain item\n\n- [ ] task item',
        ],
        [
            'an underlined heading in a task item, which is written with a number sign',
            '- plain\n- [ ] task\n\n  note\n\n  Heading\n  ---',
            '- plain\n\n- [ ] task\n\n  note\n  ## Heading',
        ],
        [
            'a table with a bare pipe in a code span, in a list item',
            '10. Run the scan\nand read the result:\n| flag | meaning |\n|---|---|\n| `a|b` | either |',
            '10. Run the scan\nand read the result:\n    | flag | meaning |\n    |---|---|\n    | `a\\|b` | either |',
        ],
    ])('writes %s the way they read back', (_name, md, saved) => {
        expect(roundTrip(md)).toBe(saved);
        expect(roundTrip(saved)).toBe(saved);
    });

    it.each([
        ['a blank line that claims to hold text', { blank: 'x', gap: 1 }, 'one\n\ntwo'],
        ['text that claims to stand right under a table', { gap: 0 }, '| a   |\n| --- |\n| 1   |\n\ntwo'],
        [
            'a tab on the blank line under a quote, which would keep the text in it',
            { blank: '\t', gap: 1 },
            '> one\n\ntwo',
        ],
        ['a tab on the blank line under text, which would be a line of it', { blank: '\t', gap: 1 }, 'one\n\ntwo'],
        [
            'a tab on the blank line under a table, which would be a row of it',
            { blank: '\t', gap: 1 },
            '| a   |\n| --- |\n| 1   |\n\ntwo',
        ],
    ])('does not write %s', (_name, attrs, saved) => {
        const paragraph = (text: string, facts?: Record<string, unknown>) => ({
            attrs: facts,
            content: [{ text, type: 'text' }],
            type: 'paragraph',
        });
        const cell = (type: string, text: string) => ({ content: [paragraph(text)], type });
        const table = {
            content: [
                { content: [cell('tableHeader', 'a')], type: 'tableRow' },
                { content: [cell('tableCell', '1')], type: 'tableRow' },
            ],
            type: 'table',
        };
        const quote = { content: [paragraph('one')], type: 'blockquote' };
        const first = saved.startsWith('|') ? table : saved.startsWith('>') ? quote : paragraph('one');

        expect(afterEdit({ content: [first, paragraph('two', attrs)], type: 'doc' }, () => {})).toBe(saved);
    });

    it.each<[string, (editor: Editor) => void, string]>([
        [
            'the empty paragraph above it is removed',
            (editor) => editor.commands.deleteRange({ from: 3, to: 5 }),
            'a\n\nb',
        ],
        [
            'another empty paragraph is put above it',
            (editor) => editor.commands.insertContentAt(3, { type: 'paragraph' }),
            'a\n\n\n\n\n\nb',
        ],
    ])(
        'writes a block that stood under four blank lines under the ones its paragraphs make once %s',
        (_name, edit, saved) => {
            const written = afterEdit('a\n\n\n\n\nb', edit);

            expect(written).toBe(saved);
            expect(roundTrip(written)).toBe(written);
        },
    );

    it.each<[string, unknown, JSONContent[], string]>([
        ['the spaces after the last line of a list', '  ', [], '- a  \n\ntext'],
        ['nothing after the last line of a list where what stood there is not spaces', 'x', [], '- a\n\ntext'],
        ['nothing after it where what stood there holds a tab', ' \t', [], '- a\n\ntext'],
    ])('writes %s', (_name, trail, _under, saved) => {
        const paragraph = { content: [{ text: 'text', type: 'text' }], type: 'paragraph' };

        expect(afterEdit({ content: [{ ...bullets('a'), attrs: { trail } }, paragraph], type: 'doc' }, () => {})).toBe(
            saved,
        );
        expect(roundTrip(saved)).toBe(saved);
    });

    it.each<[string, JSONContent, JSONContent[], string]>([
        ['at the end of the document, where nothing reads them back', bullets('a'), [], '- a'],
        [
            'above a list, where they would be text of the item',
            bullets('a'),
            [{ ...bullets('b'), attrs: { marker: '*' } }],
            '- a\n\n* b',
        ],
        [
            'above nothing but an empty paragraph and a list',
            bullets('a'),
            [{ type: 'paragraph' }, { ...bullets('b'), attrs: { marker: '*' } }],
            '- a\n\n\n\n* b',
        ],
        [
            'after a quote that ends the list',
            {
                content: [
                    {
                        content: [
                            { content: [{ text: 'a', type: 'text' }], type: 'paragraph' },
                            {
                                content: [{ content: [{ text: 'q', type: 'text' }], type: 'paragraph' }],
                                type: 'blockquote',
                            },
                        ],
                        type: 'listItem',
                    },
                ],
                type: 'bulletList',
            },
            [{ content: [{ text: 'text', type: 'text' }], type: 'paragraph' }],
            '- a\n  > q\n\ntext',
        ],
    ])('does not write the spaces a list ended with %s', (_name, list, under, saved) => {
        const written = afterEdit({ content: [{ ...list, attrs: { trail: '  ' } }, ...under], type: 'doc' }, () => {});

        expect(written).toBe(saved);
    });

    it('keeps the spaces of the blank line above a block once the block above it is another', () => {
        const saved = afterEdit('x\n\na\n   \nb\n\nc', (editor) => editor.commands.deleteRange({ from: 3, to: 6 }));

        expect(saved).toBe('x\n   \nb\n\nc');
    });

    it.each<[string, string, (editor: Editor) => void, string]>([
        [
            'stands beside the block that had spaces above it',
            'a\n \nb',
            (editor) => editor.chain().setTextSelection(2).splitBlock().run(),
            'a\n\n\n\nb',
        ],
        [
            'is what is left of the block that had spaces above it',
            'a\n \nb\n\nc',
            (editor) => editor.commands.deleteRange({ from: 4, to: 5 }),
            'a\n\n\n\nc',
        ],
    ])('writes a plain blank line where an empty paragraph %s', (_name, md, edit, expected) => {
        const saved = afterEdit(md, edit);

        expect(saved).toBe(expected);
        expect(roundTrip(saved)).toBe(saved);
    });

    it('does not carry the spaces of a list under a list that only an empty paragraph keeps apart', () => {
        const saved = afterEdit('- prev\n\nSteps:\n\n  1. x\n  2. y', (editor) =>
            editor.view.dispatch(editor.state.tr.delete(11, 17)),
        );

        expect(saved).toBe('- prev\n\n\n\n1. x\n2. y');
        expect(kindsOf(saved)).toEqual(['bulletList', 'paragraph', 'orderedList']);
    });

    it('does not carry the spaces of a list that holds a line left of its items', () => {
        const item = (content: JSONContent[], attrs?: Record<string, unknown>) => ({
            attrs,
            content: [{ content, type: 'paragraph' }],
            type: 'listItem',
        });
        const lazy = item(
            [
                { text: 'c', type: 'text' },
                { attrs: { marker: '' }, type: 'hardBreak' },
                { text: 'lazy', type: 'text' },
            ],
            { lazyFrom: 1 },
        );
        const list = { attrs: { indent: 2 }, content: [item([{ text: 'a', type: 'text' }]), lazy], type: 'bulletList' };

        expect(afterEdit({ content: [list], type: 'doc' }, () => {})).toBe('- a\n- c\nlazy');
    });

    it.each([
        ['one blank line above it', { blank: '  ', gap: 1 }, '- a\n  \n- b'],
        ['no blank line above it', { blank: '  ', gap: 0 }, '- a\n- b'],
        ['two blank lines above it', { blank: '  ', gap: 2 }, '- a\n\n- b'],
        ['a tab on the blank line above it', { blank: '\t', gap: 1 }, '- a\n\n- b'],
        ['text on the blank line above it', { blank: 'x', gap: 1 }, '- a\n\n- b'],
    ])(
        'writes the spaces of the blank line above an item only as spaces on that one line: %s',
        (_name, attrs, saved) => {
            const [first] = bullets('a').content ?? [];
            const [second] = bullets('b').content ?? [];

            expect(
                afterEdit(
                    { content: [{ content: [first!, { ...second, attrs }], type: 'bulletList' }], type: 'doc' },
                    () => {},
                ),
            ).toBe(saved);
            expect(roundTrip(saved)).toBe(saved);
        },
    );

    it('puts text that reads as too many cells right under the table it was cut from', () => {
        const paragraph = (text: string, attrs?: Record<string, unknown>) => ({
            attrs,
            content: [{ text, type: 'text' }],
            type: 'paragraph',
        });
        const cell = (type: string, text: string) => ({ content: [paragraph(text)], type });
        const table = {
            attrs: { delimiterRow: '|---|---|' },
            content: [{ content: [cell('tableHeader', 'a'), cell('tableHeader', 'b')], type: 'tableRow' }],
            type: 'table',
        };

        expect(afterEdit({ content: [table, paragraph('| 1 | 2 | 3 |', { gap: 0 })], type: 'doc' }, () => {})).toBe(
            '| a | b |\n|---|---|\n| 1 | 2 | 3 |',
        );
        expect(afterEdit({ content: [table, paragraph('| 1 | 2 |', { gap: 0 })], type: 'doc' }, () => {})).toBe(
            '| a | b |\n|---|---|\n\n| 1 | 2 |',
        );
    });

    it('puts a blank line between two paragraphs of an item, where a tag line does not end the text above it', () => {
        const paragraph = (text: string, attrs?: Record<string, unknown>) => ({
            attrs,
            content: [{ text, type: 'text' }],
            type: 'paragraph',
        });
        const editor = load({
            content: [
                {
                    content: [{ content: [paragraph('a'), paragraph('<details>', { gap: 0 })], type: 'listItem' }],
                    type: 'bulletList',
                },
            ],
            type: 'doc',
        });
        const md = editor.getMarkdown();

        editor.destroy();

        expect(md).toBe('- a\n\n  <details>');
    });

    it('numbers a new item after the one above it', () => {
        const saved = afterEdit('22. ssh\n80. http\n443. https', (editor) =>
            editor.chain().setTextSelection(14).splitListItem('listItem').insertContent('alt').run(),
        );

        expect(saved).toBe('22. ssh\n80. http\n81. alt\n443. https');
    });

    it('leaves a number behind when its item is in a list that had none', () => {
        const item = (text: string, attrs?: Record<string, unknown>) => ({
            attrs,
            content: [{ content: [{ text, type: 'text' }], type: 'paragraph' }],
            type: 'listItem',
        });
        const editor = load({
            content: [{ content: [item('a'), item('https', { number: 443 }), item('b')], type: 'orderedList' }],
            type: 'doc',
        });
        const saved = editor.getMarkdown();

        editor.destroy();

        expect(saved).toBe('1. a\n2. https\n3. b');
    });

    it('writes a plain delimiter row over one that is not a row of the table', () => {
        const cell = (type: string, text: string) => ({
            content: [{ content: [{ text, type: 'text' }], type: 'paragraph' }],
            type,
        });
        const editor = load({
            content: [
                {
                    attrs: { delimiterRow: '---' },
                    content: [
                        { content: [cell('tableHeader', 'a')], type: 'tableRow' },
                        { content: [cell('tableCell', '1')], type: 'tableRow' },
                    ],
                    type: 'table',
                },
            ],
            type: 'doc',
        });
        const saved = editor.getMarkdown();

        editor.destroy();

        expect(saved).toBe('| a |\n| --- |\n| 1 |');
    });

    it('writes a plain delimiter row once the table has another column', () => {
        const saved = afterEdit('| a | b |\n|:--|-----|\n| 1 | 2 |', (editor) =>
            editor.chain().setTextSelection(4).addColumnAfter().run(),
        );

        expect(saved).toBe('| a | | b |\n| :--- | --- | --- |\n| 1 | | 2 |');
        expect(roundTrip(saved)).toBe(saved);
    });

    it('closes a fence that nothing closed once a block follows it', () => {
        const saved = afterEdit('```sh\nid\n', (editor) =>
            editor.commands.insertContentAt(editor.state.doc.content.size, {
                content: [{ text: 'after', type: 'text' }],
                type: 'paragraph',
            }),
        );

        expect(saved).toBe('```sh\nid\n```\n\nafter\n');
    });

    it.each([
        ['a typed document', 'hello', 'hello'],
        ['a document that ends with a line break', 'hello\n', 'hello\n'],
    ])('ends %s the way it ended once an empty heading is its last block', (_name, md, saved) => {
        expect(
            afterEdit(md, (editor) => editor.chain().setTextSelection(6).splitBlock().setHeading({ level: 1 }).run()),
        ).toBe(saved);
    });

    it.each<[string, JSONContent | string]>([
        ['nothing but spaces', '   '],
        ['nothing but a line break', { type: 'hardBreak' }],
    ])('ends a document the way it ended once a paragraph of %s is its last block', (_name, content) => {
        expect(
            afterEdit('hello\n', (editor) =>
                editor.chain().setTextSelection(6).splitBlock().insertContent(content).run(),
            ),
        ).toBe('hello\n');
    });

    it('ends a document the way it ended once a heading of one ideographic space is its last block', () => {
        expect(
            afterEdit('hello\n', (editor) =>
                editor.chain().setTextSelection(6).splitBlock().setHeading({ level: 2 }).insertContent('\u3000').run(),
            ),
        ).toBe('hello\n');
    });

    it('drops the empty paragraph TrailingNode adds under a last block', () => {
        expect(afterEdit('- a\n- b', (editor) => editor.chain().setTextSelection(4).insertContent('!').run())).toBe(
            '- a!\n- b',
        );
    });
});

describe('code written by indentation stays that way while it reads as that code', () => {
    const code = (text: string, attrs: Record<string, unknown> = { written: `    ${text}` }): JSONContent => ({
        attrs,
        content: text ? [{ text, type: 'text' }] : [],
        type: 'codeBlock',
    });
    const paragraph = (text: string): JSONContent => ({ content: [{ text, type: 'text' }], type: 'paragraph' });
    const saved = (...content: JSONContent[]) => afterEdit({ content, type: 'doc' }, () => {});

    it('writes every line four spaces in once the code is edited', () => {
        const edited = afterEdit('Run:\n\n\tnmap -sV host\n \n\tcurl host\n\nDone.', (editor) =>
            editor.chain().setTextSelection(7).insertContent('sudo ').run(),
        );

        expect(edited).toBe('Run:\n\n    sudo nmap -sV host\n\n    curl host\n\nDone.');
        expect(roundTrip(edited)).toBe(edited);
    });

    it.each<[string, JSONContent[], string]>([
        ['a language', [code('id', { language: 'sh', written: '    id' })], '```sh\nid\n```'],
        ['a blank first line', [code('\nid')], '```\n\nid\n```'],
        ['a line break at its end', [code('id\n')], '```\nid\n\n```'],
        ['no code', [paragraph('a'), code('')], 'a\n\n```\n\n```'],
        ['a list above it', [bullets('a'), code('id')], '- a\n\n```\nid\n```'],
        [
            'a list above it and an empty paragraph between',
            [bullets('a'), { type: 'paragraph' }, code('id')],
            '- a\n\n\n\n```\nid\n```',
        ],
        ['an indented block above it', [code('a'), code('b')], '    a\n\n```\nb\n```'],
    ])('is fenced with %s', (_name, blocks, written) => {
        expect(saved(...blocks)).toBe(written);
        expect(roundTrip(written)).toBe(written);
    });

    it("is fenced in a list item, where indentation is the item's own", () => {
        const item = { content: [paragraph('a'), code('id')], type: 'listItem' };

        expect(saved({ content: [item], type: 'bulletList' })).toBe('- a\n  ```\n  id\n  ```');
    });

    it('stands a blank line under text, which a line right under it would continue', () => {
        const written = saved(paragraph('a'), code('id', { gap: 0, written: '    id' }));

        expect(written).toBe('a\n\n    id');
        expect(roundTrip(written)).toBe(written);
    });

    it('stands right under a block that ends by itself', () => {
        const heading = { attrs: { level: 2 }, content: [{ text: 'Run', type: 'text' }], type: 'heading' };

        expect(saved(heading, code('id', { gap: 0, written: '    id' }))).toBe('## Run\n    id');
    });

    it.each([
        ['other code', '    whoami', '    id'],
        ['a heading', '# Title', '```\nid\n```'],
        ['other code between fences', '~~~\nwhoami\n~~~', '```\nid\n```'],
        ['a fence that nothing closes', '```\nid', '```\nid\n```'],
        ['a language the block does not have', '```sh\nid\n```', '```\nid\n```'],
        ['text under the code', '```\nid\n```\ntext', '```\nid\n```'],
    ])('does not write lines that hold %s', (_name, written, expected) => {
        expect(saved(code('id', { written }))).toBe(expected);
    });

    it.each<[string, JSONContent[], string]>([
        [
            'at the margin under a list',
            [bullets('a'), code('id', { written: '````\nid\n````' })],
            '- a\n\n````\nid\n````',
        ],
        [
            "anew under a list when its fence stood past the margin, where it would be the item's",
            [bullets('a'), code('id', { written: '  ```\n  id\n  ```' })],
            '- a\n\n```\nid\n```',
        ],
    ])('writes a fenced block %s', (_name, blocks, expected) => {
        expect(saved(...blocks)).toBe(expected);
        expect(roundTrip(expected)).toBe(expected);
    });

    it.each([
        ['with the fence it had', 'Run:\n\n   ````sh\n   id\n   ````  ', 'Run:\n\n````sh\nid!\n````'],
        ['between fences though it has no language', 'Run:\n\n   ```\n   id\n   ```', 'Run:\n\n```\nid!\n```'],
    ])('writes a fenced block anew once its code is edited, %s', (_name, md, saved) => {
        expect(afterEdit(md, (editor) => editor.chain().setTextSelection(9).insertContent('!').run())).toBe(saved);
    });
});

describe('a row of a table is saved as the line it was written as while it holds the same cells', () => {
    const TABLE = '| Name | Value |\n|------|:-----:|\n| a    |   1   |\n| longer name | 2 |\n|c|3|';

    const positionOf = (editor: Editor, text: string): number => {
        let at = 0;

        editor.state.doc.descendants((node, position) => {
            at = node.text === text ? position : at;
        });

        return at;
    };

    it('writes only the row that was edited anew', () => {
        const saved = afterEdit(TABLE, (editor) =>
            editor.chain().setTextSelection(positionOf(editor, 'longer name')).insertContent('a ').run(),
        );

        expect(saved).toBe('| Name | Value |\n|------|:-----:|\n| a    |   1   |\n| a longer name | 2     |\n|c|3|');
        expect(roundTrip(saved)).toBe(saved);
    });

    it('keeps the lines of the other rows when a row is removed, and when one is added', () => {
        const removed = afterEdit(TABLE, (editor) =>
            editor.chain().setTextSelection(positionOf(editor, 'a')).deleteRow().run(),
        );
        const added = afterEdit(TABLE, (editor) =>
            editor.chain().setTextSelection(positionOf(editor, 'a')).addRowAfter().run(),
        );

        expect(removed).toBe('| Name | Value |\n|------|:-----:|\n| longer name | 2 |\n|c|3|');
        expect(added).toBe(
            '| Name | Value |\n|------|:-----:|\n| a    |   1   |\n|             |       |\n| longer name | 2 |\n|c|3|',
        );
        expect(roundTrip(added)).toBe(added);
    });

    it('writes every row anew once the table has another number of columns', () => {
        const saved = afterEdit(TABLE, (editor) =>
            editor.chain().setTextSelection(positionOf(editor, 'a')).addColumnAfter().run(),
        );

        expect(saved).toBe(
            [
                '| Name        |     | Value |',
                '| ----------- | --- | :-----: |',
                '| a           |     | 1     |',
                '| longer name |     | 2     |',
                '| c           |     | 3     |',
            ].join('\n'),
        );
        expect(roundTrip(saved)).toBe(saved);
    });

    it('writes the delimiter row anew once a column is aligned another way', () => {
        const saved = afterEdit(TABLE, (editor) =>
            editor
                .chain()
                .setTextSelection(positionOf(editor, 'Value'))
                .updateAttributes('tableHeader', { align: 'right' })
                .run(),
        );

        expect(saved.split('\n')[1]).toBe('| ----------- | -----: |');
    });

    it.each<[string, unknown, string]>([
        ['is not a list', 'x', '| a |\n| --- |\n| 1 |'],
        ['holds a line break in a row', [['| a |'], ['|---|'], ['| 1 |\n| 2 |', '1']], '| a |\n|---|\n| 1 |'],
        ['holds an empty line for a row', [['| a |'], ['|---|'], ['', '1']], '| a |\n|---|\n| 1 |'],
        ['holds other cells for a row', [['| a |'], ['|---|'], ['| 2 |', '2']], '| a |\n|---|\n| 1 |'],
        ['holds other cells for the header', [['| b |', 'b'], ['|---|'], ['|1|', '1']], '| a |\n|---|\n|1|'],
        ['holds two lines for the header', [['|a|\n|z|', 'a'], ['|---|'], ['|1|', '1']], '| a |\n|---|\n|1|'],
        ['holds a delimiter row of two columns', [['|a|', 'a'], ['|---|---|'], ['|1|', '1']], '|a|\n| --- |\n|1|'],
        ['holds a delimiter row with a line break', [['|a|', 'a'], ['|---|\n'], ['|1|', '1']], '|a|\n| --- |\n|1|'],
        ['holds a row that is not a line', [[7, 'a'], ['|---|'], [null, '1']], '| a |\n|---|\n| 1 |'],
    ])('writes a line of the table from its cells where what it was written as %s', (_name, written, saved) => {
        const cell = (type: string, text: string) => ({
            content: [{ content: [{ text, type: 'text' }], type: 'paragraph' }],
            type,
        });
        const table = {
            attrs: { delimiterRow: '| --- |', written },
            content: [
                { content: [cell('tableHeader', 'a')], type: 'tableRow' },
                { content: [cell('tableCell', '1')], type: 'tableRow' },
            ],
            type: 'table',
        };

        expect(afterEdit({ content: [table], type: 'doc' }, () => {})).toBe(saved);
    });

    it('keeps the line of a row whose cell gained only a space at its end', () => {
        const saved = afterEdit(TABLE, (editor) =>
            editor
                .chain()
                .setTextSelection(positionOf(editor, 'longer name') + 'longer name'.length)
                .insertContent(' ')
                .run(),
        );

        expect(saved).toBe(TABLE);
    });
});

describe('a stored line is not written where it would be read as something else', () => {
    const shapeOf = (markdown: string): unknown => {
        const editor = load(markdown);
        const strip = (node: JSONContent): unknown => ({
            ...(node.content ? { content: node.content.map(strip) } : {}),
            ...(node.text ? { text: node.text } : {}),
            type: node.type,
        });
        const shape = strip(editor.getJSON());

        editor.destroy();

        return shape;
    };

    const tasks = (text: string): JSONContent => ({
        content: [
            {
                attrs: { checked: false },
                content: [{ content: [{ text, type: 'text' }], type: 'paragraph' }],
                type: 'taskItem',
            },
        ],
        type: 'taskList',
    });

    it.each([
        [
            'a fence line that is a row only as the last line of its quote',
            '> | a | b |\n> |---|---|\n> | 1 | 2 |\n> ```\nlazy text\n',
            '> | a | b |\n> |---|---|\n> | 1 | 2 |\n> | ``` | |\n>\n> lazy text\n',
        ],
        [
            'a row of nothing but a tab in a quote',
            '> | a | b |\n> |---|---|\n> | 1 | 2 |\n> \t\n> | 3 | 4 |\n',
            '> | a | b |\n> |---|---|\n> | 1 | 2 |\n> | | |\n> | 3 | 4 |\n',
        ],
        [
            'a table two spaces in under a list that is written at the margin',
            '-\titem\n\n  | a | b |\n  |---|---|\n  | 1 | 2 |\n',
            '- item\n\n| a | b |\n|---|---|\n| 1 | 2 |\n',
        ],
        [
            'a table two spaces in under a task list',
            'Intro\n\n  - [ ] check ports\n\n  | port | state |\n  |---|---|\n  | 22 | open |\n\nEnd\n',
            'Intro\n\n- [ ] check ports\n\n| port | state |\n|---|---|\n| 22 | open |\n\nEnd\n',
        ],
        [
            'a table whose header alone is indented, under a list',
            ' - a\n  - b\n\n  | a | b |\n|---|---|\n| `x|y` | 2 |\n',
            '- a\n- b\n\n| a | b |\n|---|---|\n| `x|y` | 2 |\n',
        ],
        [
            'a table in a quote written with a tab after its marker',
            '>\t| a | b |\n>\t|---|---|\n>\t| `x|y` | 2 |\n',
            '> | a | b |\n> |---|---|\n> | `x|y` | 2 |\n',
        ],
        [
            'an indented block that holds a fence line, above a list with a tab in its code',
            'Open the block with:\n\n    ```\n\n- then a recipe:\n\n      all:\tdeps\n\n```\nend\n```\n',
            'Open the block with:\n\n````\n```\n````\n\n- then a recipe:\n\n  ```\n  all:\tdeps\n  ```\n\n```\nend\n```\n',
        ],
        [
            'two blank lines above a list of plain and task items',
            'Intro\n\n\n- note\n- [ ] todo\n\nEnd\n',
            'Intro\n\n- note\n\n- [ ] todo\n\nEnd\n',
        ],
        [
            'a blank line with spaces above a list of plain and task items',
            'Intro\n  \n- note\n- [ ] todo\n',
            'Intro\n\n- note\n\n- [ ] todo\n',
        ],
        [
            'a table in a list item whose row holds a pipeline',
            '- item\n\n  |a|b|\n  |-|-|\n  |{{ .X | upper }}|2|',
            '- item\n\n  |a|b|\n  |-|-|\n  | {{ .X | upper }} | 2   |',
        ],
        [
            'a list one space in under a task list',
            'Intro\n\n - [ ] todo\n - note\n\nEnd\n',
            'Intro\n\n- [ ] todo\n\n- note\n\nEnd\n',
        ],
        [
            'spaces after a list of plain and task items',
            'Intro\n\n- a\n- [ ] t\n- b  \n\nEnd\n',
            'Intro\n\n- a\n\n- [ ] t\n\n- b\n\nEnd\n',
        ],
        [
            'a line of white space under a quoted table in a list item',
            '- item\n  > | {{ .A | f }} | c |\n  > |---|---|\n  >  \t\n',
            '- item\n  > | {{ .A | f }} | c |\n  > |---|---|\n',
        ],
        [
            'a fence closed by a run of two kinds of character, above a list with a tab in its code',
            '```\ncode\n```~~\n\n- item\n\n      a\tb\n\n```\nx\n```\n',
            '```\ncode\n```\n\n- item\n\n  ```\n  a\tb\n  ```\n\n```\nx\n```\n',
        ],
    ])('%s', (_name, md, saved) => {
        expect(roundTrip(md)).toBe(saved);
        expect(roundTrip(saved)).toBe(saved);
        expect(shapeOf(saved)).toEqual(shapeOf(md));
    });

    it.each([
        [
            'text two spaces in and a fence under it',
            '  - [ ] build\n  - [x] test\n\n  Then install:\n\n  ```sh\n  sudo make install\n  ```\n',
            '- [ ] build\n- [x] test\n\nThen install:\n\n  ```sh\n  sudo make install\n  ```\n',
        ],
        [
            'text two spaces in and indented code under it',
            'Steps:\n\n  - [ ] build\n  - [x] test\n\n  Note: run as root.\n\n      sudo make install\n',
            'Steps:\n\n- [ ] build\n- [x] test\n\nNote: run as root.\n\n      sudo make install\n',
        ],
    ])('keeps what stands under a task list out of its last item: %s', (_name, md, saved) => {
        expect(roundTrip(md)).toBe(saved);
        expect(roundTrip(saved)).toBe(saved);
        expect(kindsOf(saved)).toEqual(kindsOf(md));
    });

    it.each<[string, JSONContent[], string]>([
        [
            'indented code under a list and a paragraph of one space',
            [
                bullets('item'),
                { content: [{ text: ' ', type: 'text' }], type: 'paragraph' },
                { attrs: { written: '    code' }, content: [{ text: 'code', type: 'text' }], type: 'codeBlock' },
            ],
            '- item\n\n \n\n```\ncode\n```',
        ],
        [
            'indented code under indented code and a paragraph of one space',
            [
                { attrs: { written: '    one' }, content: [{ text: 'one', type: 'text' }], type: 'codeBlock' },
                { content: [{ text: ' ', type: 'text' }], type: 'paragraph' },
                { attrs: { written: '    two' }, content: [{ text: 'two', type: 'text' }], type: 'codeBlock' },
            ],
            '    one\n\n \n\n```\ntwo\n```',
        ],
        [
            'a task list with two blank lines above it, under a bullet list',
            [
                bullets('item'),
                {
                    attrs: { gap: 2 },
                    content: [
                        {
                            attrs: { checked: false },
                            content: [{ content: [{ text: 'task', type: 'text' }], type: 'paragraph' }],
                            type: 'taskItem',
                        },
                    ],
                    type: 'taskList',
                },
            ],
            '- item\n\n- [ ] task',
        ],
        [
            'an empty paragraph between a task list and text',
            [tasks('t'), { type: 'paragraph' }, { content: [{ text: 'x', type: 'text' }], type: 'paragraph' }],
            '- [ ] t\n\n\n\n\nx',
        ],
        [
            'text that stood five blank lines under a task list, without the empty paragraphs it was read with',
            [tasks('t'), { attrs: { gap: 5 }, content: [{ text: 'x', type: 'text' }], type: 'paragraph' }],
            '- [ ] t\n\nx',
        ],
        [
            'a list of `-` bullets that stood three blank lines under a task list under such a list',
            [bullets('a'), tasks('b'), { ...bullets('c'), attrs: { gap: 3 } }],
            '- a\n\n- [ ] b\n\n- c',
        ],
        [
            'text one space in under a task list',
            [
                {
                    content: [
                        {
                            attrs: { checked: false },
                            content: [{ content: [{ text: 'task', type: 'text' }], type: 'paragraph' }],
                            type: 'taskItem',
                        },
                    ],
                    type: 'taskList',
                },
                { content: [{ text: ' note', type: 'text' }], type: 'paragraph' },
            ],
            '- [ ] task\n\nnote',
        ],
        [
            'a table written two spaces in as the first block of a list item',
            [
                {
                    content: [
                        {
                            content: [
                                {
                                    attrs: {
                                        delimiterRow: '|---|',
                                        written: [['  | a |', 'a'], ['  |---|'], ['  | 1 |', '1']],
                                    },
                                    content: [
                                        {
                                            content: [
                                                {
                                                    content: [
                                                        { content: [{ text: 'a', type: 'text' }], type: 'paragraph' },
                                                    ],
                                                    type: 'tableHeader',
                                                },
                                            ],
                                            type: 'tableRow',
                                        },
                                        {
                                            content: [
                                                {
                                                    content: [
                                                        { content: [{ text: '1', type: 'text' }], type: 'paragraph' },
                                                    ],
                                                    type: 'tableCell',
                                                },
                                            ],
                                            type: 'tableRow',
                                        },
                                    ],
                                    type: 'table',
                                },
                            ],
                            type: 'listItem',
                        },
                    ],
                    type: 'bulletList',
                },
            ],
            '- | a |\n  |---|\n  | 1 |',
        ],
        [
            'code with a tab in a quote in a list item',
            [
                {
                    content: [
                        {
                            content: [
                                { content: [{ text: 'a', type: 'text' }], type: 'paragraph' },
                                {
                                    content: [
                                        {
                                            attrs: { written: '\tif x:\n\t\treturn 1' },
                                            content: [{ text: 'if x:\n\treturn 1', type: 'text' }],
                                            type: 'codeBlock',
                                        },
                                    ],
                                    type: 'blockquote',
                                },
                            ],
                            type: 'listItem',
                        },
                    ],
                    type: 'bulletList',
                },
            ],
            '- a\n  > ```\n  > if x:\n  > \treturn 1\n  > ```',
        ],
    ])('after an edit: %s', (_name, blocks, saved) => {
        const written = afterEdit({ content: blocks, type: 'doc' }, () => {});

        expect(written).toBe(saved);
        expect(shapeOf(roundTrip(written))).toEqual(shapeOf(written));
    });

    it('writes no blank line of its own between a bullet list and a task list once the block between them is gone', () => {
        const written = afterEdit('- item\n\nx\n\n\n\n\n- [ ] task\n', (editor) => {
            let from = 0;
            let to = 0;

            editor.state.doc.forEach((node, at) => {
                [from, to] = node.textContent === 'x' ? [at, at + node.nodeSize] : [from, to];
            });
            editor.view.dispatch(editor.state.tr.delete(from, to));
        });

        expect(roundTrip(written)).toBe(written);
        expect(shapeOf(roundTrip(written))).toEqual(shapeOf(written));
    });

    it('writes no blank line of its own above a list of `-` bullets that stands under a task list under such a list', () => {
        const written = afterEdit('- a\n\n- [x] b\n\ntext\n\n\n- c\n', (editor) => {
            let from = 0;
            let to = 0;

            editor.state.doc.forEach((node, at) => {
                [from, to] = node.textContent === 'text' ? [at, at + node.nodeSize] : [from, to];
            });
            editor.view.dispatch(editor.state.tr.delete(from, to));
        });

        expect(written).toBe('- a\n\n- [x] b\n\n- c\n');
        expect(roundTrip(written)).toBe(written);
    });

    it('writes a row with a fence in its first cell once a block stands under the table', () => {
        const written = afterEdit('| a | b |\n|---|---|\n| 1 | 2 |\n```', (editor) =>
            editor.commands.insertContentAt(editor.state.doc.content.size, {
                content: [{ text: 'typed', type: 'text' }],
                type: 'paragraph',
            }),
        );

        expect(written).toBe('| a | b |\n|---|---|\n| 1 | 2 |\n| ``` | |\n\ntyped');
        expect(kindsOf(written)).toEqual(['table', 'paragraph']);
    });

    it('keeps an action typed into a row under a header that was written without a pipe', () => {
        const written = afterEdit('Name\n|---|\nvalue\n', (editor) => {
            let at = 0;

            editor.state.doc.descendants((node, position) => {
                at = node.text === 'value' ? position : at;
            });
            editor.chain().setTextSelection(at).insertContent('{{ .Z | upper }} ').run();
        });

        expect(written).toBe('Name\n|---|\n| {{ .Z | upper }} value |\n');
        expect(JSON.stringify(shapeOf(written))).toContain('{{ .Z | upper }} value');
    });

    it('keeps the empty last line of the code under a fence nothing closes, in a document that ends with no line break', () => {
        const codeOf = (markdown: string) => {
            const editor = load(markdown);
            const code = editor.state.doc.firstChild?.textContent;

            editor.destroy();

            return code;
        };

        const saved = roundTrip('  ```\ncode\n  ');

        expect(saved).toBe('```\ncode\n\n');
        expect(roundTrip(saved)).toBe(saved);
        expect([codeOf('  ```\ncode\n  '), codeOf(saved)]).toEqual(['code\n', 'code\n']);
    });

    it('writes one more line break under a fence nothing closes once its code ends with an empty line', () => {
        const written = afterEdit('intro\n\n```\nfirst\nlast', (editor) =>
            editor
                .chain()
                .setTextSelection(editor.state.doc.content.size - 1)
                .insertContent('\n')
                .run(),
        );

        expect(written).toBe('intro\n\n```\nfirst\nlast\n\n');
        expect(roundTrip(written)).toBe(written);
    });
});

describe('the spaces an empty line of code was written with in a list item', () => {
    const list = (type: string, blankLine: unknown): JSONContent => ({
        content: [
            {
                content: [
                    { content: [{ text: 'step', type: 'text' }], type: 'paragraph' },
                    { attrs: { blankLine }, content: [{ text: 'a\n\nb', type: 'text' }], type: 'codeBlock' },
                ],
                type: 'listItem',
            },
        ],
        type,
    });

    it.each([
        [
            'are written while they fit under the marker',
            'orderedList',
            '   ',
            '1. step\n   ```\n   a\n   \n   b\n   ```',
        ],
        ['are dropped under a marker they reach past', 'bulletList', '   ', '- step\n  ```\n  a\n\n  b\n  ```'],
        ['are not written when they are not spaces', 'orderedList', '\t', '1. step\n   ```\n   a\n\n   b\n   ```'],
        ['are not written when they are text', 'orderedList', 'x', '1. step\n   ```\n   a\n\n   b\n   ```'],
    ])('%s', (_name, type, blankLine, written) => {
        expect(afterEdit({ content: [list(type, blankLine)], type: 'doc' }, () => {})).toBe(written);
        expect(roundTrip(written)).toBe(written);
    });

    it('are written on a line added to the code', () => {
        const written = afterEdit('1. step\n   ```\n   a\n   \n   b\n   ```', (editor) =>
            editor.chain().setTextSelection(13).insertContent('\n\nc').run(),
        );

        expect(written).toBe('1. step\n   ```\n   a\n   \n   b\n   \n   c\n   ```');
        expect(roundTrip(written)).toBe(written);
    });

    it('are not written once the code stands outside a list', () => {
        const [item] = list('orderedList', '   ').content ?? [];

        expect(afterEdit({ content: [item!.content![1]!], type: 'doc' }, () => {})).toBe('```\na\n\nb\n```');
    });
});

describe('a line left of its item goes back under it when marked would not read it there', () => {
    const lazyItem = (lines: string[], attrs: Record<string, unknown> = { lazyFrom: 1 }): JSONContent => ({
        content: [
            {
                content: [
                    {
                        attrs,
                        content: [
                            {
                                content: lines.flatMap((text, at) => [
                                    ...(at ? [{ attrs: { marker: '' }, type: 'hardBreak' }] : []),
                                    ...(text ? [{ text, type: 'text' }] : []),
                                ]),
                                type: 'paragraph',
                            },
                        ],
                        type: 'listItem',
                    },
                ],
                type: 'bulletList',
            },
        ],
        type: 'doc',
    });

    it.each([
        ['an opening tag', ['item', '<host> is up', 'last'], '- item\n  <host> is up\n  last'],
        ['a hash', ['item', '#hashtag line', 'last'], '- item\n  #hashtag line\n  last'],
        ['a bullet', ['item', '- not an item'], '- item\n  - not an item'],
        ['a fence', ['item', '```sh'], '- item\n  ```sh'],
        [
            'leading spaces, and the line under it',
            ['item', '  two spaces in', 'last'],
            '- item\n    two spaces in\n  last',
        ],
        ['a line whose rest opens a heading', ['item', 'x # comment', 'last'], '- item\nx # comment\n  last'],
        ['a closing tag', ['item', '</steps>'], '- item\n</steps>'],
    ])('%s', (_name, lines, saved) => {
        const editor = load(lazyItem(lines));
        const md = editor.getMarkdown();

        editor.destroy();

        expect(md).toBe(saved);
    });

    it('stays at the margin whatever depth it claims', () => {
        const editor = load(lazyItem(['item', 'lazy'], { lazyDepth: -1, lazyFrom: 1 }));
        const md = editor.getMarkdown();

        editor.destroy();

        expect(md).toBe('- item\nlazy');
    });

    it('does not follow a line indented like code', () => {
        const editor = load(lazyItem(['item', '    deep', 'last'], { lazyFrom: 2 }));
        const md = editor.getMarkdown();

        editor.destroy();

        expect(md).toBe('- item\n      deep\n  last');
    });

    it('stays at the margin under a line that is only partly indented', () => {
        const saved = roundTrip('- item\n partially\nflush');

        expect(saved).toBe('- item\n   partially\nflush');
        expect(roundTrip(saved)).toBe(saved);
    });

    it('does not follow an empty line', () => {
        const editor = load(lazyItem(['item\n\nlazy']));
        const md = editor.getMarkdown();

        editor.destroy();

        expect(md).toBe('- item\n\n  lazy');
    });

    it('is not copied to the item Enter makes', () => {
        const saved = afterEdit('- first\n- item\nlazy one\nlazy two\n- third', (editor) =>
            editor.chain().setTextSelection(16).splitListItem('listItem').run(),
        );

        expect(saved).toBe('- first\n- item\n- lazy one\n  lazy two\n- third');
    });

    it('reads back the marker it escapes on an indented line', () => {
        const editor = load({
            content: [
                {
                    content: [
                        { text: 'a', type: 'text' },
                        { type: 'hardBreak' },
                        { text: '  # not a heading', type: 'text' },
                    ],
                    type: 'paragraph',
                },
            ],
            type: 'doc',
        });
        const saved = editor.getMarkdown();

        editor.destroy();

        expect(saved).toBe('a  \n  \\# not a heading');
        const reread = load(saved);
        const texts = JSON.stringify(reread.getJSON());

        reread.destroy();

        expect(texts).toContain('"text":"  # not a heading"');
        expect(roundTrip(saved)).toBe(saved);
    });

    it('leaves the other private-use character, the one that shields tabs in fences, alone', () => {
        expect(roundTrip('glyph \uE000 here')).toBe('glyph \uE000 here');
    });

    it.each([
        [
            'in the text of a line',
            '- item\ntail \uE001 here\n\n\uE001 opens a paragraph',
            '- item\n  tail \uE001 here\n\n\uE001 opens a paragraph',
        ],
        ['first on a line of an item', '- item\n\uE001 second', '- item\n  \uE001 second'],
        ['first in a paragraph of an item', '- item\n\n  \uE001 para', '- item\n\n  \uE001 para'],
        ['in the text of an image', '![a\uE001b](http://x.test/i.png)', '![a\uE001b](http://x.test/i.png)'],
        [
            'in a link target under a list',
            '- item\nlazy [x](http://x.test/\uE001b)',
            '- item\n  lazy [x](http://x.test/\uE001b)',
        ],
    ])('keeps a private-use character the author wrote %s, and indents every line', (_name, md, saved) => {
        expect(roundTrip(md)).toBe(saved);
        expect(roundTrip(saved)).toBe(saved);
    });
});

describe('the line break a document ends with outlives the block that carried it', () => {
    it.each([
        [
            'joined into the block above',
            'A\n\nB\n',
            (editor: Editor) => {
                editor.commands.setTextSelection(4);
                editor.commands.keyboardShortcut('Backspace');
            },
            'AB\n',
        ],
        ['deleted', 'A\n\nB\n', (editor: Editor) => editor.commands.deleteRange({ from: 3, to: 6 }), 'A\n'],
        ['emptied', 'A\n\nB\n', (editor: Editor) => editor.view.dispatch(editor.state.tr.delete(4, 5)), 'A\n'],
        [
            'wrapped in a list',
            'A\n\nB\n',
            (editor: Editor) => editor.chain().setTextSelection(4).toggleBulletList().run(),
            'A\n\n- B\n',
        ],
        [
            'made a list of another kind',
            'x\n\n* a\n* b\n\n',
            (editor: Editor) => editor.chain().setTextSelection(6).toggleOrderedList().run(),
            'x\n\n1. a\n2. b\n\n',
        ],
        [
            'followed by a new block',
            'A\n',
            (editor: Editor) => editor.chain().setTextSelection(2).splitBlock().insertContent('B').run(),
            'A\n\nB\n',
        ],
    ])('%s', (_name, md, edit, saved) => {
        expect(afterEdit(md, edit)).toBe(saved);
    });

    it.each([
        ['replaced', (editor: Editor) => editor.commands.setContent('new', { contentType: 'markdown' }), 'new'],
        ['left with no text', (editor: Editor) => editor.view.dispatch(editor.state.tr.delete(1, 2)), ''],
    ])('but not a document that is %s', (_name, edit, saved) => {
        expect(afterEdit('A\n', edit)).toBe(saved);
    });
});

describe('a fact read off pasted HTML is not one', () => {
    it('ignores layout attributes on the elements', () => {
        const editor = new Editor({
            content:
                '<ul marker="*" gap="0"><li number="7" lazyfrom="1"><p>a</p></li></ul><ol delimiter=")"><li><p>b</p></li></ol>',
            extensions: createMarkdownExtensions(),
        });
        const saved = editor.getMarkdown();

        editor.destroy();

        expect(saved).toBe('- a\n\n1. b');
    });
});

describe('the cost of reading the layout grows with the document, not with the square of it', () => {
    it.each<[string, 'read' | 'saved', (size: number) => string, number]>([
        ['blank lines between two paragraphs', 'read', (size) => `a\n${'\n'.repeat(size)}b\n`, 16_000],
        [
            'a line of backticks in a fence nothing closes',
            'read',
            (size) => `\`\`\`\nstart\n${'`'.repeat(size)}x`,
            16_000,
        ],
    ])('%s, %s', (_name, side, build, size) => {
        const { destroy, parse, serialize } = markdownCodec();
        const growth =
            side === 'read'
                ? slowdownWhenInputQuadruples(build, parse, size)
                : slowdownWhenInputQuadruples((length) => parse(build(length)), serialize, size);

        destroy();

        expect(growth).toBeLessThan(GROWTH_IF_QUADRATIC / 2);
    });
});
