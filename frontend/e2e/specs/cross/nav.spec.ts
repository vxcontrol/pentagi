import { expect, test } from '../../fixtures/test.ts';
import { expectCleanPage } from '../../helpers/errors.ts';
import { ROUTE_MANIFEST } from '../../routes.ts';

test.describe('route sweep', { tag: '@cross' }, () => {
    for (const entry of ROUTE_MANIFEST) {
        test.describe(entry.path, () => {
            test.use({ cassette: entry.cassette() });

            test('renders with a clean page', async ({ page, pageErrorLog }) => {
                await page.goto(entry.path);

                await expect(entry.ready(page)).toBeVisible();
                expectCleanPage(pageErrorLog);
            });
        });
    }
});
