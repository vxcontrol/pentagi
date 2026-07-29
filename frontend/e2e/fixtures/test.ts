import { test as base, expect as baseExpect, mergeTests } from '@playwright/test';

import type { PageErrorLog } from '../helpers/errors.ts';

import { watchPageErrors } from '../helpers/errors.ts';
import { test as backendTest } from './backend.ts';

const errorsTest = base.extend<{ pageErrorLog: PageErrorLog }>({
    pageErrorLog: [
        async ({ page }, use) => {
            const log = watchPageErrors(page);

            await use(log);

            // Auto (not opt-in): a spec that never destructured this fixture would assert nothing.
            // Console errors stay with `expectCleanPage` — several specs drive 4xx paths that log one.
            baseExpect(log.pageErrors, 'no uncaught page errors').toEqual([]);
        },
        { auto: true },
    ],
});

export const test = mergeTests(backendTest, errorsTest);

export { expect } from '@playwright/test';
