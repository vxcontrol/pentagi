import { ProviderType } from '@/graphql/types';

import { expect, test } from '../../fixtures/test.ts';
import { expectCleanPage } from '../../helpers/errors.ts';
import { entity } from '../../mocks/cassette.ts';
import { flowsCassette, makeFlow, PROVIDER } from '../../mocks/cassettes/flows.ts';

const TITLES = [
    'Zulu recon',
    'Alpha recon',
    'Mike recon',
    'Bravo recon',
    'Yankee recon',
    'Charlie recon',
    'Xray recon',
    'Delta recon',
    'Whiskey recon',
    'Echo recon',
    'Victor recon',
    'Foxtrot recon',
    'Uniform recon',
    'Golf recon',
];

const manyFlows = TITLES.map((title, index) => makeFlow(String(100 + index), title));

const titleColumn = () => /flow|title|name/i;

test.describe('flows table', { tag: '@flows' }, () => {
    test.use({ cassette: flowsCassette({ queries: { flows: [{ data: { flows: manyFlows } }] } }) });

    test('sorts by the column the reader clicks, and says so out loud', async ({ page, pageErrorLog }) => {
        await page.goto('/flows');

        const header = page.getByRole('columnheader', { name: titleColumn() }).first();
        const firstCell = () => page.locator('tbody tr').first().locator('td').nth(1).innerText();

        await expect(header).toHaveAttribute('aria-sort', 'none');

        await header.getByRole('button').click();
        await expect(header).toHaveAttribute('aria-sort', 'ascending');

        const ascending = await firstCell();

        await header.getByRole('button').click();
        await expect(header).toHaveAttribute('aria-sort', 'descending');

        const descending = await firstCell();

        const titles = manyFlows.map(({ title }) => title).sort((left, right) => left.localeCompare(right));

        expect(ascending, 'ascending starts at the first title of the whole set').toBe(titles.at(0));
        expect(descending, 'descending starts at the last one').toBe(titles.at(-1));
        expectCleanPage(pageErrorLog);
    });

    test('remembers the order the reader chose across a reload', async ({ page, pageErrorLog }) => {
        await page.goto('/flows');

        const header = page.getByRole('columnheader', { name: titleColumn() }).first();
        const firstCell = page.locator('tbody tr').first().locator('td').nth(1);

        await header.getByRole('button').click();
        await expect(header).toHaveAttribute('aria-sort', 'ascending');

        const chosen = await firstCell.innerText();

        await page.reload();

        await expect(page.getByRole('columnheader', { name: titleColumn() }).first()).toHaveAttribute(
            'aria-sort',
            'ascending',
        );
        await expect(page.locator('tbody tr').first().locator('td').nth(1)).toHaveText(chosen);
        expectCleanPage(pageErrorLog);
    });

    test('pages through the rows and stops at both ends', async ({ page, pageErrorLog }) => {
        await page.goto('/flows');

        const rows = page.locator('tbody tr');
        const back = page.getByRole('button', { name: /previous page/i });
        const forward = page.getByRole('button', { name: /next page/i });

        await expect(rows).toHaveCount(10);
        await expect(page.getByText(/Page 1 of 2/)).toBeVisible();
        await expect(back).toBeDisabled();

        const firstPage = await rows.first().locator('td').nth(1).innerText();

        await forward.click();

        await expect(page.getByText(/Page 2 of 2/)).toBeVisible();
        await expect(rows).toHaveCount(4);
        await expect(forward).toBeDisabled();
        await expect(rows.first().locator('td').nth(1)).not.toHaveText(firstPage);
        expectCleanPage(pageErrorLog);
    });
});

test.describe('a flow from last year on a provider since removed', { tag: '@flows' }, () => {
    const lastYear = {
        ...makeFlow('300', 'Last year recon'),
        createdAt: '2025-06-15T12:00:00Z',
        provider: entity('Provider', { name: 'retired', type: ProviderType.Custom }),
        updatedAt: '2025-06-16T12:00:00Z',
    };

    test.use({
        cassette: flowsCassette({
            queries: { flows: [{ data: { flows: [lastYear] } }], providers: [{ data: { providers: [PROVIDER] } }] },
        }),
    });

    test('dates the row without the time and keeps the provider name readable', async ({ page, pageErrorLog }) => {
        await page.goto('/flows');

        const row = page.getByRole('row', { name: /Last year recon/ });
        const name = row.locator('[data-slot="flow-provider-label"] span', { hasText: 'retired' });
        const mark = row.getByRole('img', { name: 'Unavailable' });

        await expect(row.getByRole('cell', { exact: true, name: '15 Jun 2025' })).toBeVisible();
        await expect(row.getByRole('cell', { exact: true, name: '16 Jun 2025' })).toBeVisible();
        await expect(mark).toBeVisible();
        expect(
            await name.evaluate((element) => element.scrollWidth <= element.clientWidth),
            'the mark leaves the name room to show in full',
        ).toBe(true);

        await mark.hover();

        await expect(page.getByRole('tooltip')).toHaveText('Unavailable');
        expectCleanPage(pageErrorLog);
    });
});
