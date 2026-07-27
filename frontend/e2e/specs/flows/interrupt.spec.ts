import type { ResultOf } from '@graphql-typed-document-node/core';
import type { Page } from '@playwright/test';

import type { PutUserInputDocument, StopFlowDocument } from '@/graphql/types';

import { ResultType, StatusType } from '@/graphql/types';

import { expect, test } from '../../fixtures/test.ts';
import { expectCleanPage } from '../../helpers/errors.ts';
import {
    FLOW_A_INITIAL_IDS,
    FLOW_B,
    flowQueryData,
    flowsCassette,
    makeFlow,
    makeMessage,
    PROVIDER,
} from '../../mocks/cassettes/flows.ts';

const WAITING_FLOW = makeFlow('5', 'E2E Alpha', StatusType.Waiting);

const openFlowA = async (page: Page) => {
    await page.goto('/flows/5?tab=automation');
    await expect(page.locator('header').getByText('E2E Alpha', { exact: true })).toBeVisible();
};

const operationRequest = (page: Page, operationName: string) =>
    page.waitForRequest(
        (request) => request.method() === 'POST' && request.postDataJSON()?.operationName === operationName,
    );

test.describe('flow interrupt', { tag: '@flows' }, () => {
    test.describe('stop', () => {
        const stopped: ResultOf<typeof StopFlowDocument> = { stopFlow: ResultType.Success };

        test.use({
            cassette: flowsCassette({
                mutations: { stopFlow: [{ data: stopped, setFlag: 'flow-stopped', variables: { flowId: '5' } }] },
                subscriptions: {
                    flowUpdated: [
                        {
                            frames: [{ payload: { data: { flowUpdated: WAITING_FLOW } }, whenFlag: 'flow-stopped' }],
                        },
                    ],
                },
            }),
        });

        test('stops a running flow from the composer and hands the input back', async ({ page, pageErrorLog }) => {
            await openFlowA(page);

            const stop = page.getByRole('button', { name: 'Cancel' });

            await expect(stop, 'a running flow offers Stop, not Submit').toBeVisible();
            await expect(page.getByRole('button', { name: 'Submit' })).toBeHidden();

            const request = operationRequest(page, 'stopFlow');

            await stop.click();
            expect((await request).postDataJSON().variables).toEqual({ flowId: '5' });

            await expect(page.getByRole('button', { name: 'Submit' })).toBeVisible();
            await expect(page.getByPlaceholder('Provide additional context or instructions...')).toBeVisible();
            expectCleanPage(pageErrorLog);
        });
    });

    test.describe('user input', () => {
        const accepted: ResultOf<typeof PutUserInputDocument> = { putUserInput: ResultType.Success };

        test.use({
            cassette: flowsCassette({
                mutations: {
                    putUserInput: [
                        {
                            data: accepted,
                            variables: { flowId: '5', input: 'follow up on the open port' },
                        },
                    ],
                },
                queries: {
                    flow: [
                        {
                            data: flowQueryData(
                                WAITING_FLOW,
                                FLOW_A_INITIAL_IDS.map((id) => makeMessage(id, '5')),
                            ),
                            variables: { id: '5' },
                        },
                    ],
                    // The list query normalizes into the same cache entry, so a Running row here
                    // would overwrite the status the detail query just delivered.
                    flows: [{ data: { flows: [WAITING_FLOW, FLOW_B] } }],
                },
            }),
        });

        test('sends the typed message to the waiting flow', async ({ page, pageErrorLog }) => {
            await openFlowA(page);

            const composer = page.getByPlaceholder('Provide additional context or instructions...');

            await composer.fill('follow up on the open port');

            const request = operationRequest(page, 'putUserInput');

            await page.getByRole('button', { name: 'Submit' }).click();

            expect((await request).postDataJSON().variables).toEqual({
                flowId: '5',
                input: 'follow up on the open port',
                modelProvider: PROVIDER.name,
            });
            await expect(composer).toHaveValue('');
            expectCleanPage(pageErrorLog);
        });
    });
});
