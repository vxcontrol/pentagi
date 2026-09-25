import { ESLint } from 'eslint';
import { join } from 'node:path';
import { beforeAll, describe, expect, it } from 'vitest';

const FRONTEND = join(__dirname, '..');

const budgetProblems = async (eslint: ESLint, code: string): Promise<number> => {
    const [result] = await eslint.lintText(code, { filePath: join(FRONTEND, 'src', 'waitfor-budget-probe.test.tsx') });

    return (result?.messages ?? []).filter((m) => m.ruleId === 'local/waitfor-budget').length;
};

const body = (waitForCall: string) => `async () => { await ${waitForCall}; }`;
const OVER = 'waitFor(() => {}, { timeout: 10_000 })';
const UNDER = 'waitFor(() => {}, { timeout: 3_000 })';
const OVER_VI = 'vi.waitFor(() => {}, { timeout: 10_000 })';

describe('the waitFor-budget lint rule sees every form this repo writes tests in', () => {
    let eslint: ESLint;

    beforeAll(async () => {
        eslint = new ESLint({ cwd: FRONTEND, overrideConfigFile: join(FRONTEND, 'eslint.config.mjs') });

        // Resolving the flat config and loading the TS parser happens on the first lint, not in
        // the constructor. Without this warm-up the cost lands on whichever row runs first, which
        // reaches vitest's 5s default on a CI runner while staying well under it here.
        await budgetProblems(eslint, 'const warmUp = 1;');
    }, 30_000);

    it.each([
        ['a plain it() with no options', `it('x', ${body(OVER)});`],
        ['it.each, the repo default for table tests', `it.each([1])('x %s', ${body(OVER)});`],
        ['test.each', `test.each([1])('x %s', ${body(OVER)});`],
        ['it.only', `it.only('x', ${body(OVER)});`],
        ['it.skip', `it.skip('x', ${body(OVER)});`],
        ['it.concurrent', `it.concurrent('x', ${body(OVER)});`],
        ['vi.waitFor', `it('x', ${body(OVER_VI)});`],
        ['an options object carrying no budget at all', `it('x', { retry: 2 }, ${body(OVER)});`],
        ['a declared budget under the vitest default', `it('x', { timeout: 4_000 }, ${body(OVER)});`],
        ['a declared budget over the default but under the waitFor', `it('x', { timeout: 6_000 }, ${body(OVER)});`],
        ['it.each as a tagged template', `it.each\`a\`('x', ${body(OVER)});`],
    ])('flags %s', async (_name, code) => {
        expect(await budgetProblems(eslint, code)).toBe(1);
    });

    it.each([
        ['it() that declares its own budget', `it('x', { timeout: 15_000 }, ${body(OVER)});`],
        ['it.each that declares its own budget', `it.each([1])('x %s', { timeout: 15_000 }, ${body(OVER)});`],
        ['a waitFor budget inside the vitest default', `it('x', ${body(UNDER)});`],
        ['a budget declared in the positional form', `it('x', ${body(OVER)}, 15_000);`],
        ['a Playwright suite hook, which this rule must not claim', `test.describe('x', ${body(OVER)});`],
        [
            "Playwright's own locator.waitFor, whose default is 30s and whose fix is not it()",
            `test('x', async ({ page }) => { await page.locator('#a').waitFor({ timeout: 10_000 }); });`,
        ],
    ])('stays silent on %s', async (_name, code) => {
        expect(await budgetProblems(eslint, code)).toBe(0);
    });
});
