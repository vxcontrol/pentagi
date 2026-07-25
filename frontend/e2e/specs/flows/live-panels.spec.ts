import { expect, test } from '../../fixtures/test.ts';
import { expectCleanPage } from '../../helpers/errors.ts';
import { MESSAGE_ID_SELECTOR } from '../../helpers/subscriptions.ts';
import { readTerminalBuffer } from '../../helpers/terminal.ts';
import {
    livePanelsCassette,
    makeMessage,
    STREAMED,
    VARIED_MESSAGES,
    variedMessagesCassette,
} from '../../mocks/cassettes/flows.ts';

const REVISED_ANSWER = 'Revised plan after the retry';

test.describe('live panels', { tag: '@flows' }, () => {
    test.use({ cassette: livePanelsCassette() });

    // Each panel below is fed by a subscription and nothing else: the seeded query answers hold
    // none of these strings, so a frame wired to the wrong cache field leaves its panel stale.
    test('every streamed panel picks up its own frame', async ({ page, pageErrorLog }) => {
        await page.goto('/flows/5');
        await expect(page.locator('header').getByRole('button', { name: 'Toggle favorite' })).toBeEnabled();

        await page.getByRole('tab', { name: 'Tasks' }).click();
        await expect(page.getByText(STREAMED.task)).toBeVisible();
        await expect(page.getByText(STREAMED.taskRetitled)).toBeVisible();

        await page.getByRole('tab', { name: 'Agents' }).click();
        await expect(page.getByText(STREAMED.agent)).toBeVisible();

        await page.getByRole('tab', { name: 'Searches' }).click();
        await expect(page.getByText(STREAMED.search)).toBeVisible();

        await page.getByRole('tab', { name: 'Vector Store' }).click();
        await expect(page.getByText(STREAMED.vector)).toBeVisible();

        await page.getByRole('tab', { name: 'Terminal' }).click();
        await expect(async () => {
            expect(await readTerminalBuffer(page)).toContain(STREAMED.terminal);
        }).toPass();

        expectCleanPage(pageErrorLog);
    });

    test.describe('a revised message', () => {
        test.use({
            cassette: variedMessagesCassette({
                subscriptions: {
                    messageLogUpdated: [
                        {
                            frames: [
                                {
                                    delayMs: 60,
                                    payload: {
                                        data: {
                                            messageLogUpdated: makeMessage('501', '5', { message: REVISED_ANSWER }),
                                        },
                                    },
                                },
                            ],
                            variables: { flowId: '5' },
                        },
                    ],
                },
            }),
        });

        test('replaces the message in place instead of appending a second one', async ({ page, pageErrorLog }) => {
            await page.goto('/flows/5?tab=automation');

            await expect(page.getByText(REVISED_ANSWER)).toBeVisible();
            await expect(page.getByText(VARIED_MESSAGES[0]!.message)).toBeHidden();
            await expect(page.locator(MESSAGE_ID_SELECTOR)).toHaveCount(VARIED_MESSAGES.length);

            expectCleanPage(pageErrorLog);
        });
    });
});
