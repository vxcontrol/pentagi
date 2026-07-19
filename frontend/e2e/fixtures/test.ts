import { test as base, mergeTests } from '@playwright/test';

import type { PageErrorLog } from '../helpers/errors.ts';

import { watchPageErrors } from '../helpers/errors.ts';
import { test as backendTest } from './backend.ts';

const errorsTest = base.extend<{ pageErrorLog: PageErrorLog }>({
    pageErrorLog: async ({ page }, use) => {
        await use(watchPageErrors(page));
    },
});

export const test = mergeTests(backendTest, errorsTest);

export { expect } from '@playwright/test';
