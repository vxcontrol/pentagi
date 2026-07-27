import type { ResultOf } from '@graphql-typed-document-node/core';
import type { Page } from '@playwright/test';

import type { CallAssistantDocument, DeleteAssistantDocument } from '@/graphql/types';

import { ResultType, StatusType } from '@/graphql/types';

import { expect, test } from '../../fixtures/test.ts';
import { expectCleanPage } from '../../helpers/errors.ts';
import {
    assistantsCassette,
    CREATED_ASSISTANT,
    makeAssistant,
    RUNNING_ASSISTANT,
    WAITING_ASSISTANT,
} from '../../mocks/cassettes/assistants.ts';
import { PROVIDER } from '../../mocks/cassettes/flows.ts';

const openAssistantTab = async (page: Page) => {
    await page.goto('/flows/5');
    await page.getByRole('tab', { name: 'Assistant' }).click();
    await expect(page.getByPlaceholder('Search messages...')).toBeVisible();
};

const openPicker = async (page: Page) => {
    await page.getByRole('button', { name: 'Select assistant' }).click();
    await expect(page.getByPlaceholder('Search assistants...')).toBeVisible();
};

const operationRequest = (page: Page, operationName: string) =>
    page.waitForRequest(
        (request) => request.method() === 'POST' && request.postDataJSON()?.operationName === operationName,
    );

test.describe('flow assistants', { tag: '@flows' }, () => {
    test.describe('create', () => {
        test.use({
            cassette: assistantsCassette({
                mutations: {
                    createAssistant: [
                        {
                            data: {
                                createAssistant: {
                                    assistant: CREATED_ASSISTANT,
                                    flow: { __typename: 'Flow', id: '5' },
                                },
                            } as never,
                            setFlag: 'assistant-created',
                            variables: { flowId: '5', input: 'plan the recon', modelProvider: PROVIDER.name },
                        },
                    ],
                },
                subscriptions: {
                    assistantCreated: [
                        {
                            frames: [
                                {
                                    payload: { data: { assistantCreated: CREATED_ASSISTANT } },
                                    whenFlag: 'assistant-created',
                                },
                            ],
                        },
                    ],
                },
            }),
        });

        test('creates an assistant from the empty composer and selects it', async ({ page, pageErrorLog }) => {
            await openAssistantTab(page);
            await openPicker(page);
            await page.getByRole('option', { name: 'Create new assistant' }).click();

            const composer = page.getByPlaceholder('Type a message to create a new assistant...');
            const picker = page.getByRole('button', { name: 'Select assistant' });

            await expect(composer).toBeVisible();
            await expect(picker).toHaveText('New');
            await composer.fill('plan the recon');

            // A brand-new assistant inherits no provider. The only one on offer is also the one the
            // previously selected assistant used — the case a defaults-sync that ignores the picked
            // value silently reverts.
            await page.getByRole('button', { name: 'Select Provider' }).click();
            await page.getByRole('menuitem', { name: PROVIDER.name }).click();

            const request = operationRequest(page, 'createAssistant');

            await page.getByRole('button', { name: 'Submit' }).click();

            expect((await request).postDataJSON().variables).toEqual({
                flowId: '5',
                input: 'plan the recon',
                modelProvider: PROVIDER.name,
                useAgents: false,
            });

            // `Create new assistant` leaves the picker on an explicit "none" that the
            // fall-back-to-first selection never overrides — only the create response does.
            await expect(picker).toHaveText(/1$/);

            await openPicker(page);
            await expect(page.getByRole('option', { name: new RegExp(CREATED_ASSISTANT.title) })).toBeVisible();
            expectCleanPage(pageErrorLog);
        });
    });

    test.describe('call', () => {
        const called: ResultOf<typeof CallAssistantDocument> = { callAssistant: ResultType.Success };

        test.use({
            cassette: assistantsCassette({
                mutations: {
                    callAssistant: [
                        {
                            data: called,
                            variables: { assistantId: '11', flowId: '5', input: 'and the other ports?' },
                        },
                    ],
                },
            }),
        });

        test('sends a follow-up to the assistant the picker already selected', async ({ page, pageErrorLog }) => {
            await openAssistantTab(page);

            const composer = page.getByPlaceholder('Continue the conversation...');

            await composer.fill('and the other ports?');

            const request = operationRequest(page, 'callAssistant');

            await page.getByRole('button', { name: 'Submit' }).click();

            expect((await request).postDataJSON().variables).toEqual({
                assistantId: WAITING_ASSISTANT.id,
                flowId: '5',
                input: 'and the other ports?',
                useAgents: false,
            });
            await expect(composer).toHaveValue('');
            expectCleanPage(pageErrorLog);
        });
    });

    test.describe('stop', () => {
        test.use({
            cassette: assistantsCassette({
                mutations: {
                    stopAssistant: [
                        {
                            data: {
                                stopAssistant: makeAssistant(
                                    RUNNING_ASSISTANT.id,
                                    RUNNING_ASSISTANT.title,
                                    StatusType.Waiting,
                                ),
                            },
                            setFlag: 'assistant-stopped',
                            variables: { assistantId: '12', flowId: '5' },
                        },
                    ],
                },
                subscriptions: {
                    assistantUpdated: [
                        {
                            frames: [
                                {
                                    payload: {
                                        data: {
                                            assistantUpdated: makeAssistant(
                                                RUNNING_ASSISTANT.id,
                                                RUNNING_ASSISTANT.title,
                                                StatusType.Waiting,
                                            ),
                                        },
                                    },
                                    whenFlag: 'assistant-stopped',
                                },
                            ],
                        },
                    ],
                },
            }),
        });

        test('stops the running assistant and hands the composer back', async ({ page, pageErrorLog }) => {
            await openAssistantTab(page);
            await openPicker(page);
            await page.getByRole('option', { name: new RegExp(RUNNING_ASSISTANT.title) }).click();

            const stop = page.getByRole('button', { name: 'Cancel' });

            await expect(stop).toBeVisible();

            const request = operationRequest(page, 'stopAssistant');

            await stop.click();

            expect((await request).postDataJSON().variables).toEqual({
                assistantId: RUNNING_ASSISTANT.id,
                flowId: '5',
            });
            await expect(page.getByPlaceholder('Continue the conversation...')).toBeVisible();
            expectCleanPage(pageErrorLog);
        });
    });

    test.describe('delete', () => {
        const deleted: ResultOf<typeof DeleteAssistantDocument> = { deleteAssistant: ResultType.Success };

        test.use({
            cassette: assistantsCassette({
                mutations: {
                    deleteAssistant: [
                        {
                            data: deleted,
                            setFlag: 'assistant-deleted',
                            variables: { assistantId: '12', flowId: '5' },
                        },
                    ],
                },
                subscriptions: {
                    assistantDeleted: [
                        {
                            frames: [
                                {
                                    payload: { data: { assistantDeleted: RUNNING_ASSISTANT } },
                                    whenFlag: 'assistant-deleted',
                                },
                            ],
                        },
                    ],
                },
            }),
        });

        test('deletes an assistant through the confirm dialog', async ({ page, pageErrorLog }) => {
            await openAssistantTab(page);
            await openPicker(page);

            const doomed = page.getByRole('option', { name: new RegExp(RUNNING_ASSISTANT.title) });

            await doomed.hover();
            await doomed.getByRole('button').click();

            const dialog = page.getByRole('dialog');

            await expect(dialog.getByText('Delete Assistant')).toBeVisible();

            const request = operationRequest(page, 'deleteAssistant');

            await dialog.getByRole('button', { name: 'Delete' }).click();

            expect((await request).postDataJSON().variables).toEqual({
                assistantId: RUNNING_ASSISTANT.id,
                flowId: '5',
            });

            await openPicker(page);
            await expect(page.getByRole('option', { name: new RegExp(WAITING_ASSISTANT.title) })).toBeVisible();
            await expect(page.getByRole('option', { name: new RegExp(RUNNING_ASSISTANT.title) })).toBeHidden();
            expectCleanPage(pageErrorLog);
        });
    });
});
