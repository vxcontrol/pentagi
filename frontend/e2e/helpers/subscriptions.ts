import type { Locator, TestInfo } from '@playwright/test';

import { expect } from '@playwright/test';

/** The message-id badge carries no accessible name, so it is addressed by its slot like every other
 *  shadcn primitive in this codebase — `data-testid` is not used in product code. */
export const MESSAGE_ID_SELECTOR = '[data-slot="flow-message-id"]';

export const attachIdSet = async (testInfo: TestInfo, name: string, ids: string[]): Promise<void> => {
    await testInfo.attach(name, {
        body: JSON.stringify(ids, null, 2),
        contentType: 'application/json',
    });
};

export const extractMessageIds = async (panel: Locator): Promise<string[]> => {
    const raw = await panel.locator(MESSAGE_ID_SELECTOR).allTextContents();

    return raw.map((text) => text.replace(/\D/g, '')).filter(Boolean);
};

export const assertNoDuplicates = (ids: string[]): void => {
    const seen = new Set<string>();
    const duplicates = ids.filter((id) => (seen.has(id) ? true : (seen.add(id), false)));

    expect(duplicates, 'no message ID may render twice within a flow').toEqual([]);
};

/**
 * Exact-set intersection, not range overlap: concurrent flows interleave global
 * IDs, so only an empty intersection proves no cross-flow leak.
 */
export const assertDisjoint = (idsA: string[], idsB: string[]): void => {
    const setB = new Set(idsB);
    const shared = [...new Set(idsA)].filter((id) => setB.has(id));

    expect(shared, 'flows must not share a single message ID').toEqual([]);
};
