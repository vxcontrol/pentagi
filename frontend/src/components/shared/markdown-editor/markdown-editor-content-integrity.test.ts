import { beforeAll, describe, expect, it } from 'vitest';

import { roundTrip, setupEditorJsdom, structuralCounts } from './markdown-editor-test-setup';

beforeAll(setupEditorJsdom);

const mulberry32 = (seed: number) => () => {
    seed = (seed + 0x6d2b79f5) | 0;
    let t = Math.imul(seed ^ (seed >>> 15), 1 | seed);
    t = (t + Math.imul(t ^ (t >>> 7), 61 | t)) ^ t;

    return ((t ^ (t >>> 14)) >>> 0) / 4294967296;
};

// Content atoms that MUST survive load→serialize verbatim: identifiers/dunders, Go-template variables,
// xml-like tags, regex/path backslashes, numeric HTML entities + a bare `&`, C++. Whitespace/formatting may
// reflow; these bytes may not disappear or mutate. (Named entities like `&lt;` decode to `<` by design —
// see markdown-editor-marked.ts — so they are NOT survive-verbatim atoms.)
const ATOMS = [
    '__init__',
    '__call__',
    'os.execute()',
    'ngx_http_lua_module',
    'snake_case_name',
    '{{.TargetURL}}',
    '{{.Lang}}',
    '{{- if .Enabled}}',
    '<container_environment>',
    '<input>',
    '<br>',
    'match \\d+ then \\w*',
    'C:\\Users\\bob\\a.txt',
    'regex \\.php files',
    'glob \\* and \\?',
    'escaped \\[ \\| \\+ here',
    '&#40;paren&#41;',
    'AT&T',
    '2>&1 redirect',
    'C++ then C++',
    '~10% left',
];

const WORDS = ['firewall', 'payload', 'exploit', 'nmap', 'recon', 'shell', 'token', 'vector'];

const wrap = (context: number, text: string): string => {
    switch (context) {
        case 1:
            return `# ${text}`;
        case 2:
            return `- ${text}`;
        case 3:
            return `1. ${text}`;
        case 4:
            return `> ${text}`;
        case 5:
            return `\`${text}\``;
        default:
            return text;
    }
};

describe('generative content-integrity — atoms survive load↔serialize across random combinations', () => {
    it('300 seeded random docs converge and drop no atom', { timeout: 30000 }, () => {
        const rng = mulberry32(0xc0ffee);
        const pick = <T>(arr: T[]): T => arr[Math.floor(rng() * arr.length)] as T;

        for (let iteration = 0; iteration < 300; iteration++) {
            const blockCount = 1 + Math.floor(rng() * 5);
            const placed: string[] = [];
            const blocks: string[] = [];

            for (let block = 0; block < blockCount; block++) {
                const atom = pick(ATOMS);

                placed.push(atom);
                const inline = rng() < 0.5 ? `${pick(WORDS)} ${atom} ${pick(WORDS)}` : `${atom} ${pick(WORDS)}`;
                blocks.push(wrap(Math.floor(rng() * 6), inline));
            }

            const doc = blocks.join('\n\n');
            const out = roundTrip(doc);
            const out2 = roundTrip(out);

            expect(out2, `did not converge (seed 0xc0ffee, iteration ${iteration}):\n${doc}`).toBe(out);

            for (const atom of placed) {
                expect(out.includes(atom), `atom "${atom}" lost (iteration ${iteration}):\n${doc}\n--> ${out}`).toBe(
                    true,
                );
            }
        }
    });

    // Source-fidelity is strictly stronger than the survives/converges oracles above: a backslash-escaped
    // punctuation char is dropped on the FIRST load and then converges, so convergence passes while the byte
    // is already gone. Assert byte-exact round-trip for the escape class directly.
    it.each([
        'regex \\.php and \\d+',
        'glob \\* \\? and range \\[a-z\\]',
        'table \\| pipe and plus \\+',
        'unc \\\\server\\\\share double \\\\',
        '\\* line-leading star, not a bullet',
    ])('round-trips %s byte-identical', (source) => {
        expect(roundTrip(source)).toBe(source);
    });

    // The survives/converges oracles above only place atoms in SINGLE contexts. Nesting primitives inside
    // one another (ordered>bullet, blockquote>list, list>code, fence-in-fence) is exactly the class where a
    // whole block can drop silently — assert structure (via structuralCounts, catching a dropped block/item)
    // AND convergence AND atom survival across depth-≥2 compositions.
    it('nested primitive compositions preserve structure, content, and converge', { timeout: 30000 }, () => {
        const rng = mulberry32(0x5eeded);
        const nestAtoms = ['{{.TargetURL}}', '<container_environment>', '__init__', 'C++ then C++', 'os.execute()'];
        const nesters: ((a: string) => string)[] = [
            (a) => `1. first ${a}\n   - nested ${a}\n2. second ${a}`,
            (a) => `> quote ${a}\n> - list item ${a}`,
            (a) => `- top ${a}\n  - mid ${a}\n    - deep ${a}`,
            (a) => `1. step ${a}\n\n   \`\`\`\n   payload ${a}\n   \`\`\``,
            (a) => `\`\`\`\`\nouter ${a}\n\`\`\`\ninner ${a}\n\`\`\`\n\`\`\`\``,
        ];

        for (let i = 0; i < 120; i++) {
            const nester = nesters[Math.floor(rng() * nesters.length)] as (a: string) => string;
            const atom = nestAtoms[Math.floor(rng() * nestAtoms.length)] as string;
            const doc = nester(atom);
            const out = roundTrip(doc);

            expect(structuralCounts(out), `structure changed (i=${i}):\n${doc}\n-->\n${out}`).toEqual(
                structuralCounts(doc),
            );
            expect(roundTrip(out), `did not converge (i=${i}):\n${doc}`).toBe(out);
            expect(out.includes(atom), `atom "${atom}" lost (i=${i}):\n${doc}\n-->\n${out}`).toBe(true);
        }
    });

    // A raw `|` inside a table cell is the one context the single-atom oracle above cannot generate: marked
    // splits the row on it before inline-tokenizing, so an unprotected pipe drops the trailing cells. Build
    // tables whose cells carry pipe-bearing atoms inside code spans / Go actions, plus a sentinel trailing
    // cell; assert the sentinel and the atom's words survive and the table converges.
    it('pipe-bearing atoms inside table cells keep their row intact and converge', { timeout: 30000 }, () => {
        const rng = mulberry32(0x7ab1e);
        const pipeAtoms = ['x | y', '{{.Host | lower}}', 'a || b', 'grep foo | wc -l', 'no-pipe-here'];

        for (let i = 0; i < 120; i++) {
            const rowCount = 1 + Math.floor(rng() * 3);
            // GFM makes the outer pipe optional; generate BOTH styles so the pipe-less body-row path is exercised.
            const outerPipe = rng() < 0.5;
            const wrap = (cells: string) => (outerPipe ? `| ${cells} |` : cells);
            const rows: string[] = [];
            const sentinels: string[] = [];
            const atoms: string[] = [];

            for (let row = 0; row < rowCount; row++) {
                const atom = pipeAtoms[Math.floor(rng() * pipeAtoms.length)] as string;
                const sentinel = `${WORDS[Math.floor(rng() * WORDS.length)]}${row}`;

                sentinels.push(sentinel);
                atoms.push(atom);
                rows.push(wrap(`\`${atom}\` | ${sentinel}`));
            }

            const doc = `${wrap('code | note')}\n${wrap('--- | ---')}\n${rows.join('\n')}`;
            const out = roundTrip(doc);

            expect(roundTrip(out), `did not converge (i=${i}):\n${doc}\n-->\n${out}`).toBe(out);

            for (const sentinel of sentinels) {
                expect(out.includes(sentinel), `cell "${sentinel}" dropped (i=${i}):\n${doc}\n-->\n${out}`).toBe(true);
            }

            // The atom itself must survive as a code span with its structural pipes escaped for the table — the
            // sentinel guards the row's cell count, this guards the pipe-bearing content from silent corruption.
            for (const atom of atoms) {
                const span = `\`${atom.replace(/\|/g, '\\|')}\``;

                expect(out.includes(span), `atom ${span} corrupted (i=${i}):\n${doc}\n-->\n${out}`).toBe(true);
            }
        }
    });
});
