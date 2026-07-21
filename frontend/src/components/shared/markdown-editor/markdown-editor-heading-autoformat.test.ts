import type { JSONContent } from '@tiptap/core';

import { Editor } from '@tiptap/core';
import { TextSelection } from '@tiptap/pm/state';
import { beforeAll, describe, expect, it } from 'vitest';

import { createMarkdownExtensions } from './markdown-editor-extensions';
import { roundTrip, setupEditorJsdom, structuralCounts } from './markdown-editor-test-setup';

beforeAll(setupEditorJsdom);

const doc = (...content: JSONContent[]): JSONContent => ({ content, type: 'doc' });
const para = (...content: JSONContent[]): JSONContent => ({ content, type: 'paragraph' });
const t = (text: string, marks?: JSONContent['marks']): JSONContent =>
    marks ? { marks, text, type: 'text' } : { text, type: 'text' };
const hb: JSONContent = { type: 'hardBreak' };

const newEditor = (content: JSONContent | string) =>
    new Editor({ content: content as string, extensions: createMarkdownExtensions() });

const liveHeadings = (editor: Editor): number => {
    let count = 0;

    editor.state.doc.descendants((node) => {
        if (node.type.name === 'heading') {
            count += 1;
        }
    });

    return count;
};

const posOf = (editor: Editor, needle: string): number => {
    let at = -1;

    editor.state.doc.descendants((node, pos) => {
        if (at < 0 && node.isText && node.text?.includes(needle)) {
            at = pos + node.text.indexOf(needle);
        }
    });

    return at;
};

type EditResult = { heading: number; json: JSONContent; md: string; reloadHeading: number };

// appendTransaction runs only on transactions, never on construction — so fire ONE real transaction.
const run = (content: JSONContent | string, edit: (editor: Editor) => void): EditResult => {
    const editor = newEditor(content);

    edit(editor);

    const md = editor.getMarkdown();
    const result: EditResult = {
        heading: liveHeadings(editor),
        json: editor.getJSON(),
        md,
        reloadHeading: structuralCounts(md).heading ?? 0,
    };

    editor.destroy();

    return result;
};

// Most cases start with a throwaway char before the marker (so construction leaves a plain paragraph), then
// delete it so the block comes to start with `# ` — the exact shape of "delete the text before a #".
const promote = (content: JSONContent, junk: string, junkLen = 1): EditResult =>
    run(content, (editor) => {
        const at = posOf(editor, junk);

        editor.commands.command(({ dispatch, tr }) => {
            dispatch?.(tr.delete(at, at + junkLen));

            return true;
        });
    });

const firstBlock = (json: JSONContent): JSONContent => json.content![0]!;

describe('levels & marker strip', () => {
    it.each([
        ['#', 1],
        ['##', 2],
        ['###', 3],
        ['####', 4],
        ['#####', 5],
        ['######', 6],
    ])('%s promotes to the matching heading level and strips the marker', (marker, level) => {
        const { heading, json, md } = promote(doc(para(t(`x${marker} Big`))), `x${marker} `);

        expect(heading).toBe(1);
        expect(firstBlock(json).attrs?.level).toBe(level);
        expect(md.trimEnd()).toBe(`${marker} Big`);
    });

    it('only the leading marker is stripped — an inner # survives in the heading body', () => {
        const { json, md } = promote(doc(para(t('x# Big # small'))), 'x# ');

        expect(firstBlock(json).content?.[0]?.text).toBe('Big # small');
        expect(md.trimEnd()).toBe('# Big # small');
    });

    it('exactly one space is consumed — a second space is kept as a leading space in the body', () => {
        const { json } = promote(doc(para(t('x#  Big'))), 'x#  ');

        expect(firstBlock(json).content?.[0]?.text).toBe(' Big');
    });

    it('a marker-only block ("## ") becomes an empty heading of that level', () => {
        const { heading, json } = promote(doc(para(t('x## '))), 'x## ');

        expect(heading).toBe(1);
        expect(firstBlock(json).type).toBe('heading');
        expect(firstBlock(json).attrs?.level).toBe(2);
        expect(firstBlock(json).content ?? []).toHaveLength(0);
    });
});

describe('negatives — a docChanged edit that does not leave a block-leading "# " must not convert', () => {
    it.each([
        ['no space after the hash', 'x#foo', 'x#foo'],
        ['hashes with no space', 'x#Big', 'x#Big'],
        ['seven hashes (level cap is 6)', 'x####### Big', 'x####### '],
        ['a leading space/indent', 'x # foo', 'x # '],
    ])('%s stays a paragraph', (_label, text, junk) => {
        const { heading, json } = promote(doc(para(t(text))), junk);

        expect(heading).toBe(0);
        expect(firstBlock(json).type).toBe('paragraph');
    });

    it('a # left mid-line by an interior delete stays a paragraph', () => {
        const { heading, json } = run(doc(para(t('ab# Big'))), (editor) => {
            editor.commands.command(({ dispatch, tr }) => {
                dispatch?.(tr.delete(2, 3));

                return true;
            });
        });

        expect(heading).toBe(0);
        expect(firstBlock(json).type).toBe('paragraph');
    });

    it('the marker space is byte-exact (0x20) — a tab, an NBSP, or a bare # do not convert in the editor', () => {
        for (const [text, junk] of [
            ['x#\tfoo', 'x#\t'],
            ['x# foo', 'x# '],
            ['x#', 'x#'],
        ]) {
            const { heading, json } = promote(doc(para(t(text!))), junk!);

            expect(heading).toBe(0);
            expect(firstBlock(json).type).toBe('paragraph');
        }
    });
});

describe('multi-line blocks (hardBreak / Shift+Enter) stay body text', () => {
    it('a # AFTER a hardBreak is not a block start (firstChild is the pre-break text)', () => {
        const { heading } = run(doc(para(t('foo'), hb, t('# bar'))), (editor) => {
            editor.commands.command(({ dispatch, tr }) => {
                dispatch?.(tr.insertText('!', 1));

                return true;
            });
        });

        expect(heading).toBe(0);
    });

    it('a leading hardBreak (empty first line) then "# x" is not text-first, so it stays a paragraph', () => {
        const { heading } = run(doc(para(hb, t('# x'))), (editor) => {
            editor.commands.command(({ dispatch, tr }) => {
                dispatch?.(tr.insertText('!', 5));

                return true;
            });
        });

        expect(heading).toBe(0);
    });

    // A heading is single-line; promoting `# a`⏎`# b` would emit `# a  \n# b`, which re-parses as TWO headings on
    // reload. A block that contains a hardBreak stays a paragraph and round-trips as escaped body text instead.
    it('a multi-line block whose FIRST line starts with "# " is NOT promoted (would be a lossy heading)', () => {
        const { heading, md, reloadHeading } = promote(doc(para(t('z# A'), hb, t('# B'))), 'z# A');

        expect(heading).toBe(0);
        expect(reloadHeading).toBe(0);
        expect(roundTrip(md)).toBe(md);
    });

    it('a multi-line block "# a"⏎"body" stays a paragraph and round-trips', () => {
        const { heading, md, reloadHeading } = promote(doc(para(t('z# A'), hb, t('body'))), 'z# A');

        expect(heading).toBe(0);
        expect(reloadHeading).toBe(0);
        expect(roundTrip(md)).toBe(md);
    });

    it('a marker-only first line before a hardBreak is not promoted and converges on reload', () => {
        const { heading, md, reloadHeading } = promote(doc(para(t('x## '), hb, t('rest'))), 'x## ');

        expect(heading).toBe(0);
        expect(reloadHeading).toBe(0);
        expect(roundTrip(roundTrip(md))).toBe(roundTrip(md));
    });

    it('splitting AT the break boundary promotes the post-break "# " tail (it becomes its own single-line block)', () => {
        const { heading } = run(doc(para(t('foo'), hb, t('# bar'))), (editor) => {
            editor.commands.command(({ dispatch, tr }) => {
                dispatch?.(tr.split(5));

                return true;
            });
        });

        expect(heading).toBe(1);
    });
});

describe('trigger paths', () => {
    it('a split in front of a mid-line # promotes the new tail block', () => {
        const { heading, json } = run(doc(para(t('123 # Big'))), (editor) => {
            editor.commands.command(({ dispatch, tr }) => {
                dispatch?.(tr.split(5));

                return true;
            });
        });

        expect(heading).toBe(1);
        expect(firstBlock(json).type).toBe('paragraph');
    });

    it('a split before a no-space #Big leaves the tail a paragraph', () => {
        const { heading } = run(doc(para(t('123 #Big'))), (editor) => {
            editor.commands.command(({ dispatch, tr }) => {
                dispatch?.(tr.split(5));

                return true;
            });
        });

        expect(heading).toBe(0);
    });

    it('backspace-merging a "# bar" block into an EMPTY preceding paragraph promotes the merged block', () => {
        const { heading, md } = run(doc(para(), para(t('# bar'))), (editor) => {
            editor.commands.command(({ dispatch, tr }) => {
                dispatch?.(tr.join(2));

                return true;
            });
        });

        expect(heading).toBe(1);
        expect(md.trimEnd()).toBe('# bar');
    });

    it('backspace-merging into a NON-empty paragraph puts the # mid-line, so it does not convert', () => {
        const { heading } = run(doc(para(t('foo')), para(t('# bar'))), (editor) => {
            editor.commands.command(({ dispatch, tr }) => {
                dispatch?.(tr.join(5));

                return true;
            });
        });

        expect(heading).toBe(0);
    });

    it('a raw insertText of "# " at a block start converts (the input-rule never fired)', () => {
        const { heading, md } = run(doc(para(t('body'))), (editor) => {
            editor.commands.command(({ dispatch, tr }) => {
                dispatch?.(tr.insertText('# ', 1));

                return true;
            });
        });

        expect(heading).toBe(1);
        expect(md.trimEnd()).toBe('# body');
    });

    it('a doc change ELSEWHERE promotes a pre-existing "# " paragraph the edit never touched', () => {
        const { heading, md } = run(doc(para(t('# Title')), para(t('body'))), (editor) => {
            editor.commands.command(({ dispatch, tr }) => {
                dispatch?.(tr.insertText('X', 10));

                return true;
            });
        });

        expect(heading).toBe(1);
        expect(md.trimEnd()).toBe('# Title\n\nXbody');
    });

    it('a no-op transaction (pure selection change, docChanged false) does NOT convert', () => {
        const { heading } = run(doc(para(t('# Title'))), (editor) => {
            editor.commands.command(({ dispatch, state }) => {
                dispatch?.(state.tr.setSelection(TextSelection.create(state.doc, 3)));

                return true;
            });
        });

        expect(heading).toBe(0);
    });

    it('setContent of a JSON doc with three marker paragraphs converts all three', () => {
        const editor = newEditor('');

        editor.commands.setContent(doc(para(t('# A')), para(t('## B')), para(t('### C'))));

        const md = editor.getMarkdown();
        const heading = liveHeadings(editor);

        editor.destroy();

        expect(heading).toBe(3);
        expect(md.trimEnd()).toBe('# A\n\n## B\n\n### C');
    });

    it('insertContent of two marker paragraphs converts both without merging the anchor block', () => {
        const editor = newEditor(doc(para(t('head'))));

        editor.commands.insertContentAt(editor.state.doc.content.size, [para(t('# One')), para(t('## Two'))]);

        const md = editor.getMarkdown();
        const heading = liveHeadings(editor);

        editor.destroy();

        expect(heading).toBe(2);
        expect(md.trimEnd()).toBe('head\n\n# One\n\n## Two');
    });
});

describe('container context (canReplaceWith gate)', () => {
    const cell = (text: string): JSONContent => ({ content: [para(t(text))], type: 'tableCell' });

    it('the first paragraph of a LIST ITEM does not convert (listItem content requires a leading paragraph)', () => {
        const { heading } = promote(
            doc({ content: [{ content: [para(t('X# Item'))], type: 'listItem' }], type: 'bulletList' }),
            'X# Item',
        );

        expect(heading).toBe(0);
    });

    it('a LATER paragraph of a list item DOES convert (the block* tail allows a heading)', () => {
        const { heading } = promote(
            doc({
                content: [{ content: [para(t('first')), para(t('X# Second'))], type: 'listItem' }],
                type: 'bulletList',
            }),
            'X# Second',
        );

        expect(heading).toBe(1);
    });

    it('the first paragraph of a TASK ITEM does not convert', () => {
        const { heading } = promote(
            doc({
                content: [{ attrs: { checked: false }, content: [para(t('X# Task'))], type: 'taskItem' }],
                type: 'taskList',
            }),
            'X# Task',
        );

        expect(heading).toBe(0);
    });

    it('a paragraph inside a BLOCKQUOTE converts and round-trips', () => {
        const { heading, md, reloadHeading } = promote(
            doc({ content: [para(t('X# Quote'))], type: 'blockquote' }),
            'X# Quote',
        );

        expect(heading).toBe(1);
        expect(reloadHeading).toBe(1);
        expect(md.trimEnd()).toBe('> # Quote');
    });

    it('a paragraph nested two blockquotes deep converts', () => {
        const { heading, md } = promote(
            doc({ content: [{ content: [para(t('X# Deep'))], type: 'blockquote' }], type: 'blockquote' }),
            'X# Deep',
        );

        expect(heading).toBe(1);
        expect(md.trimEnd()).toBe('> > # Deep');
    });

    // A heading IS schema-valid in a table cell (block+), so the plugin promotes it live and does not throw —
    // but a GFM cell is single-line, so the saved `| # Big | z |` reloads with the marker as literal cell text.
    // Accepted divergence (the markdown is stable across further edits); pin the live conversion + no-throw.
    it('a paragraph inside a TABLE CELL converts live (heading is not representable in saved GFM — accepted)', () => {
        const { heading, reloadHeading } = promote(
            doc({ content: [{ content: [cell('X# Big'), cell('z')], type: 'tableRow' }], type: 'table' }),
            'X# Big',
        );

        expect(heading).toBe(1);
        expect(reloadHeading).toBe(0);
    });
});

describe('multi-target ordering (high → low)', () => {
    const deleteBoth = (content: JSONContent, first: string, second: string): EditResult =>
        run(content, (editor) => {
            const a = posOf(editor, first);
            const b = posOf(editor, second);
            const [hi, lo] = a > b ? [a, b] : [b, a];

            editor.commands.command(({ dispatch, tr }) => {
                dispatch?.(tr.delete(hi, hi + 1).delete(lo, lo + 1));

                return true;
            });
        });

    it('two paragraphs convert to their own levels in one transaction', () => {
        const { heading, md } = deleteBoth(doc(para(t('x# Foo')), para(t('y## Bar'))), 'x# Foo', 'y## Bar');

        expect(heading).toBe(2);
        expect(md.trimEnd()).toBe('# Foo\n\n## Bar');
    });

    it('inverted marker widths attach each level to its own block', () => {
        const { md } = deleteBoth(doc(para(t('z### High')), para(t('w# Low'))), 'z### High', 'w# Low');

        expect(md.trimEnd()).toBe('### High\n\n# Low');
    });

    it('a non-matching paragraph between two targets is left alone', () => {
        const { heading, md } = deleteBoth(
            doc(para(t('x# Alpha')), para(t('plain body')), para(t('y### Gamma'))),
            'x# Alpha',
            'y### Gamma',
        );

        expect(heading).toBe(2);
        expect(md.trimEnd()).toBe('# Alpha\n\nplain body\n\n### Gamma');
    });

    it('a whole-doc rescan converts dormant marker paragraphs the edit did not touch', () => {
        const { heading, md } = run(doc(para(t('# Foo')), para(t('## Bar'))), (editor) => {
            editor.commands.command(({ dispatch, tr }) => {
                dispatch?.(tr.insertText('s', 14));

                return true;
            });
        });

        expect(heading).toBe(2);
        expect(md.trimEnd()).toBe('# Foo\n\n## Bars');
    });
});

describe('inline content survives promotion', () => {
    it.each([
        ['bold', 'bold', '# **Big**'],
        ['italic', 'italic', '# *Big*'],
        ['code', 'code', '# `Big`'],
        ['strike', 'strike', '# ~~Big~~'],
    ])('a %s-marked body keeps its mark and serializes correctly', (_label, mark, expected) => {
        const { heading, md } = promote(doc(para(t('x# '), t('Big', [{ type: mark }]))), 'x# ');

        expect(heading).toBe(1);
        expect(md.trimEnd()).toBe(expected);
    });

    it('a link body keeps its href and round-trips', () => {
        const { md } = promote(
            doc(para(t('x# '), t('link', [{ attrs: { href: 'https://example.com' }, type: 'link' }]))),
            'x# ',
        );

        expect(md.trimEnd()).toBe('# [link](https://example.com)');
        expect(roundTrip('# [link](https://example.com)')).toBe('# [link](https://example.com)');
    });

    it('a {{.Variable}} in the body survives verbatim', () => {
        const { heading, md } = promote(doc(para(t('x# Deploy {{.Target}} now'))), 'x# Deploy');

        expect(heading).toBe(1);
        expect(md.trimEnd()).toBe('# Deploy {{.Target}} now');
    });

    it('an <xml_tag> in the body survives verbatim (no entity-encode, no swallow)', () => {
        const { md } = promote(doc(para(t('x# Use <container_environment> here'))), 'x# Use');

        expect(md.trimEnd()).toBe('# Use <container_environment> here');
        expect(md).not.toContain('&lt;');
    });

    it('a marker split across a mark boundary ("#" bold, " Big" plain) is deliberately NOT detected', () => {
        const { heading } = run(doc(para(t('#', [{ type: 'bold' }]), t(' Big'))), (editor) => {
            editor.commands.command(({ dispatch, tr }) => {
                dispatch?.(tr.insertText('!', 6));

                return true;
            });
        });

        expect(heading).toBe(0);
    });
});

describe('byte-fidelity interaction (init / reload paths)', () => {
    it('a real ATX heading loaded from markdown is byte-identical and untouched (no transaction on init)', () => {
        expect(roundTrip('# foo')).toBe('# foo');
        expect(roundTrip('###### foo')).toBe('###### foo');
    });

    it('an after-hardBreak escaped "\\# bar" round-trips byte-exact as body text', () => {
        const source = 'foo  \n\\# bar';

        expect(roundTrip(source)).toBe(source);
        expect(structuralCounts(source).heading ?? 0).toBe(0);
    });

    it('escaped "\\# foo" stays a paragraph on INIT but converts through a setContent transaction', () => {
        expect(roundTrip('\\# foo')).toBe('\\# foo');
        expect(structuralCounts('\\# foo').heading ?? 0).toBe(0);

        const editor = newEditor('');

        editor.commands.setContent('\\# foo', { contentType: 'markdown' });

        const heading = liveHeadings(editor);
        const md = editor.getMarkdown();

        editor.destroy();

        expect(heading).toBe(1);
        expect(md.trimEnd()).toBe('# foo');
    });

    it('a converted heading serialization "# Big\\n\\n" reloads byte-stable', () => {
        expect(roundTrip('# Big\n\n')).toBe('# Big\n\n');
    });

    it('regex/glob/pipe literals and a mid-line # are byte-identical and produce no heading', () => {
        for (const s of ['match \\d+ then \\* glob', '\\| pipe here', 'note # not-a-heading', 'trailing hash #']) {
            expect(roundTrip(s)).toBe(s);
            expect(structuralCounts(s).heading ?? 0).toBe(0);
        }
    });
});

describe('re-entrancy & inert nodes', () => {
    it('a nested "# # Big" strips only the leading marker (no loop, inner # kept)', () => {
        const { heading, json, md } = promote(doc(para(t('x# # Big'))), 'x# ');

        expect(heading).toBe(1);
        expect(firstBlock(json).content?.[0]?.text).toBe('# Big');
        expect(md.trimEnd()).toBe('# # Big');
    });

    it('an existing heading is left inert when an unrelated block is edited', () => {
        const { heading, md } = run(
            doc({ attrs: { level: 1 }, content: [t('Title')], type: 'heading' }, para(t('body'))),
            (editor) => {
                editor.commands.command(({ dispatch, tr }) => {
                    dispatch?.(tr.insertText('!', 8));

                    return true;
                });
            },
        );

        expect(heading).toBe(1);
        expect(md.trimEnd()).toBe('# Title\n\n!body');
    });

    it('a "# " line inside a code block is never promoted', () => {
        const { heading, json } = run(doc({ content: [t('# foo')], type: 'codeBlock' }, para(t('body'))), (editor) => {
            editor.commands.command(({ dispatch, tr }) => {
                dispatch?.(tr.insertText('!', 8));

                return true;
            });
        });

        expect(heading).toBe(0);
        expect(firstBlock(json).type).toBe('codeBlock');
    });

    it('an empty paragraph is skipped without throwing during the scan', () => {
        expect(() =>
            run(doc(para(), para(t('body'))), (editor) => {
                editor.commands.command(({ dispatch, tr }) => {
                    dispatch?.(tr.insertText('!', 3));

                    return true;
                });
            }),
        ).not.toThrow();
    });
});
