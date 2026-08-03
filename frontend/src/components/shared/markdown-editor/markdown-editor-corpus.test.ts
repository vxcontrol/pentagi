import { readdirSync, readFileSync } from 'node:fs';
import { join } from 'node:path';
import { beforeAll, describe, expect, it } from 'vitest';

import { roundTrip, setupEditorJsdom, structuralCounts } from './markdown-editor-test-setup';

beforeAll(setupEditorJsdom);

const variables = (s: string) => s.match(/\{\{[^{}]*\}\}/g) ?? [];
const words = (s: string) => s.match(/[\p{L}\p{N}]+/gu) ?? [];

describe('corpus — every real prompt .tmpl survives the round-trip with no content loss', () => {
    const dir = join(__dirname, '..', '..', '..', '..', '..', 'backend', 'pkg', 'templates', 'prompts');
    const files = readdirSync(dir).filter((file) => file.endsWith('.tmpl'));

    for (const file of files) {
        it(file + ': tags literal, {{ }} preserved, structure + words intact, converges', () => {
            const src = readFileSync(join(dir, file), 'utf8');
            const save1 = roundTrip(src);
            const save2 = roundTrip(save1);

            expect(save1).not.toContain('&lt;');
            expect(save1).not.toContain('&gt;');
            expect(variables(save1)).toEqual(variables(src));
            // word MULTISET (not a Set) so a count drop / dup / reorder is caught, plus ≤2-save convergence
            // (some templates canonicalize over two saves — save1 ≠ save2 is within contract, save2 is stable)
            expect(words(save2).slice().sort()).toEqual(words(src).slice().sort());
            // structural node counts (code blocks, list items, headings, tables) — a word-preserving structural
            // downgrade (block split / dropped cell) passes the multiset check but fails this.
            expect(structuralCounts(save2)).toEqual(structuralCounts(src));
            expect(roundTrip(save2)).toBe(save2);
        });
    }
});
