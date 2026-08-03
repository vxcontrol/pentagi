import type { Node as PMNode } from '@tiptap/pm/model';

export interface InlineMatch {
    from: number;
    text: string;
    to: number;
}

// Find `regex` matches in each textblock's inline text, mapped to document positions. Unlike a per-text-node
// scan, this reunites a token split across text nodes by a mark (e.g. a user styles one brace of `{{.Var}}`,
// so ProseMirror splits it) — the brace and the rest sit in one block string again. ProseMirror positions are
// one per UTF-16 unit and mark-independent; `from`/`to` are read from the per-character position map (NOT
// `from + length`) so a token that also contains a non-text inline node — a hard break from Shift+Enter —
// still spans the right range. `regex` MUST be global (`/g`). Scanning per textblock — NOT over
// `doc.textContent` — keeps positions aligned across blocks.
export const collectInlineMatches = (doc: PMNode, regex: RegExp): InlineMatch[] => {
    const matches: InlineMatch[] = [];

    doc.descendants((node, pos) => {
        if (!node.isTextblock) {
            return;
        }

        let text = '';
        const positions: number[] = [];
        let inlinePos = pos + 1;

        node.forEach((child) => {
            if (child.isText && child.text) {
                for (let i = 0; i < child.text.length; i += 1) {
                    text += child.text[i];
                    positions.push(inlinePos + i);
                }
            }

            inlinePos += child.nodeSize;
        });

        for (const match of text.matchAll(regex)) {
            const start = match.index;
            const from = positions[start];
            const last = positions[start + match[0].length - 1];

            if (from !== undefined && last !== undefined) {
                matches.push({ from, text: match[0], to: last + 1 });
            }
        }

        return false;
    });

    return matches;
};
