import { expect, test } from '../../fixtures/test.ts';
import { expectCleanPage } from '../../helpers/errors.ts';
import { flowTabsCassette } from '../../mocks/cassettes/flows.ts';

// The query-backed detail tabs were uncovered — only Terminal had a spec. Each
// marker is unique to its tab's seeded data, so a visible marker proves that tab
// rendered its content.
const TABS = [
    { marker: 'E2E Task Alpha', name: 'Tasks' },
    { marker: 'E2E agent reconnaissance', name: 'Agents' },
    { marker: 'E2E search for the CVE', name: 'Searches' },
    { marker: 'E2E recall prior findings', name: 'Vector Store' },
] as const;

test.describe('flow detail tabs', { tag: '@flows' }, () => {
    test.use({ cassette: flowTabsCassette() });

    test('each query-backed tab renders its populated content', async ({ page, pageErrorLog }) => {
        await page.goto('/flows/5');
        await expect(page.getByRole('button', { name: 'Flow actions' })).toBeVisible();

        for (const { marker, name } of TABS) {
            await page.getByRole('tab', { name }).click();
            await expect(page.getByText(marker)).toBeVisible();
        }

        expectCleanPage(pageErrorLog);
    });
});
