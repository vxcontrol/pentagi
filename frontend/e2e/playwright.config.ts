import { defineConfig, devices } from '@playwright/test';
import { fileURLToPath } from 'node:url';

import type { BackendOptions } from './fixtures/backend.ts';

export type BackendTier = 'local' | 'mock' | 'stand';

const tier = (process.env.E2E_TIER ?? 'mock') as BackendTier;
const isCI = Boolean(process.env.CI);

// `vite preview` listens on VITE_PORT + 100 and reuses the dev proxy config.
const PREVIEW_PORT = 8100;

// Reporter paths resolve against the process cwd, not the config file — pin them
// so CI artifact uploads have one deterministic location.
const here = (relative: string) => fileURLToPath(new URL(relative, import.meta.url));

const TIERS: Record<BackendTier, { baseURL: string; installMocks: boolean }> = {
    local: { baseURL: process.env.E2E_BASE_URL ?? 'https://localhost:8443', installMocks: false },
    mock: { baseURL: `http://localhost:${PREVIEW_PORT}`, installMocks: true },
    stand: { baseURL: process.env.E2E_BASE_URL ?? '', installMocks: false },
};

if (tier === 'stand' && !TIERS.stand.baseURL) {
    throw new Error('E2E_TIER=stand requires E2E_BASE_URL');
}

export default defineConfig<BackendOptions>({
    forbidOnly: isCI,
    fullyParallel: true,
    globalTimeout: isCI ? 10 * 60_000 : undefined,
    outputDir: './test-results',
    projects: [
        {
            name: `${tier}-chromium`,
            use: { ...devices['Desktop Chrome'] },
        },
    ],
    reporter: isCI
        ? [
              ['blob', { outputDir: here('./blob-report') }],
              ['json', { outputFile: here('./test-results/results.json') }],
              ['github'],
              ['list'],
          ]
        : [['html', { open: 'never', outputFolder: here('./playwright-report') }], ['list']],
    // Retries and trace policy are intentionally identical local vs CI so a red CI
    // run reproduces byte-identically with `CI=1 pnpm e2e`.
    retries: 1,
    testDir: './specs',
    use: {
        backend: { installMocks: TIERS[tier].installMocks, tier },
        baseURL: TIERS[tier].baseURL,
        ignoreHTTPSErrors: tier === 'local',
        locale: 'en-US',
        screenshot: 'only-on-failure',
        timezoneId: 'UTC',
        trace: 'on-first-retry',
        video: 'retain-on-failure',
    },
    // The gate must test the shipped artifact: keep this on the production build,
    // never a dev server.
    webServer:
        tier === 'mock'
            ? {
                  command: 'pnpm run build && pnpm exec vite preview',
                  cwd: fileURLToPath(new URL('..', import.meta.url)),
                  env: { VITE_PORT: '8000', VITE_USE_HTTPS: 'false' },
                  reuseExistingServer: !isCI,
                  timeout: 240_000,
                  url: `http://localhost:${PREVIEW_PORT}`,
              }
            : undefined,
    workers: isCI ? 1 : undefined,
});
