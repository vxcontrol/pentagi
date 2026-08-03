import type { ResultOf } from '@graphql-typed-document-node/core';

import type { CreateFlowDocument } from '@/graphql/types';

import { StatusType } from '@/graphql/types';

import { expect, test } from '../../fixtures/test.ts';
import { expectCleanPage } from '../../helpers/errors.ts';
import { flowQueryData, flowsCassette, makeFlow, PROVIDER } from '../../mocks/cassettes/flows.ts';

const CREATED_FLOW = makeFlow('7', 'Say Hello Flow', StatusType.Created);
const created: ResultOf<typeof CreateFlowDocument> = { createFlow: CREATED_FLOW };

test.describe('flow create', { tag: ['@flows', '@smoke'] }, () => {
    test.use({
        cassette: flowsCassette({
            mutations: {
                createFlow: [
                    {
                        data: created,
                        setFlag: 'flow-created',
                        variables: { input: 'Say hello', modelProvider: PROVIDER.name },
                    },
                ],
            },
            queries: {
                flow: [{ data: flowQueryData(CREATED_FLOW, []), variables: { id: '7' } }],
                providers: [{ data: { providers: [PROVIDER] } }],
            },
            subscriptions: {
                flowCreated: [
                    { frames: [{ payload: { data: { flowCreated: CREATED_FLOW } }, whenFlag: 'flow-created' }] },
                ],
            },
        }),
    });

    test('creates a flow from the form and lands on its detail page', async ({ page, pageErrorLog }) => {
        await page.goto('/flows/new');

        await expect(page.getByText(PROVIDER.name)).toBeVisible();
        await page.getByPlaceholder(/Describe what you would like PentAGI to test/).fill('Say hello');
        await page.getByRole('button', { name: 'Submit' }).click();

        await expect(page).toHaveURL(/\/flows\/7/);
        await expect(page.locator('header').getByRole('button', { name: 'Toggle favorite' })).toBeEnabled();
        await expect(page.locator('header').getByText('Say Hello Flow')).toBeVisible();
        expectCleanPage(pageErrorLog);
    });
});
