import type { Node as PMNode } from '@tiptap/pm/model';
import type { EditorState, Transaction } from '@tiptap/pm/state';

import { Extension } from '@tiptap/core';
import { Plugin, PluginKey } from '@tiptap/pm/state';

const HEADING_MARKER = /^(#{1,6}) /;

const containsHardBreak = (node: PMNode, hardBreak: null | PMNode['type']): boolean => {
    if (!hardBreak) {
        return false;
    }

    let found = false;

    node.forEach((child) => {
        found ||= child.type === hardBreak;
    });

    return found;
};

// The heading input-rule fires only on the keystroke that types `# ` at a block start; a block that comes to
// start with `# ` any other way — deleting the text before an existing `#`, Enter in front of it, paste — stays
// a paragraph, unlike CommonMark (a leading ATX marker IS a heading) and unlike what that same text becomes on
// the next load. This plugin promotes such a paragraph to the matching heading on every doc-changing transaction.
// Two guards keep the promotion sound:
//   • it keys off the FIRST child, not textContent: a `# ` AFTER a hardBreak (Shift+Enter) must stay body text,
//     and there the first child is the pre-break text, so the block does not read as starting with `#`;
//   • it skips a paragraph that CONTAINS a hardBreak: a heading is single-line, so promoting a multi-line block
//     (e.g. `# a`⏎`# b`) would emit `# a  \n# b`, which re-parses as TWO headings on the next load. Such a
//     block stays a paragraph; escapeLineLeadingBlockMarkers then escapes its leading `# ` so it round-trips.
export const HeadingAutoformat = Extension.create({
    addProseMirrorPlugins() {
        return [
            new Plugin({
                appendTransaction(transactions: readonly Transaction[], _oldState: EditorState, newState: EditorState) {
                    if (!transactions.some((transaction) => transaction.docChanged)) {
                        return null;
                    }

                    const heading = newState.schema.nodes.heading;
                    const paragraph = newState.schema.nodes.paragraph;
                    const hardBreak = newState.schema.nodes.hardBreak ?? null;

                    if (!heading || !paragraph) {
                        return null;
                    }

                    const targets: { level: number; markerEnd: number; start: number }[] = [];

                    newState.doc.descendants((node: PMNode, pos: number) => {
                        if (node.type !== paragraph) {
                            return true;
                        }

                        const first = node.firstChild;
                        const match = first?.isText && first.text ? HEADING_MARKER.exec(first.text) : null;

                        if (match && !containsHardBreak(node, hardBreak)) {
                            const start = pos + 1;
                            const $start = newState.doc.resolve(start);

                            if ($start.node(-1).canReplaceWith($start.index(-1), $start.indexAfter(-1), heading)) {
                                targets.push({ level: match[1]!.length, markerEnd: start + match[0].length, start });
                            }
                        }

                        return false;
                    });

                    if (targets.length === 0) {
                        return null;
                    }

                    const tr = newState.tr;

                    // High position → low so each edit leaves the earlier, not-yet-applied positions valid.
                    for (const { level, markerEnd, start } of targets.reverse()) {
                        tr.delete(start, markerEnd).setBlockType(start, start, heading, { level });
                    }

                    return tr;
                },
                key: new PluginKey('headingAutoformat'),
            }),
        ];
    },
    name: 'headingAutoformat',
});
