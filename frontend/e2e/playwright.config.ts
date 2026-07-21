import { defineConfig, devices } from '@playwright/test';
import { fileURLToPath } from 'node:url';

import type { BackendOptions, BackendTier } from './fixtures/backend.ts';

const rawTier = process.env.E2E_TIER ?? 'mock';
const isCI = Boolean(process.env.CI);
// Visual snapshots only ever run inside the pinned Playwright Linux container
// (e2e/tools/run-visual.sh) — a darwin run would generate parallel baselines
// that never match CI pixels.
const isVisual = process.env.E2E_VISUAL === '1';

// `vite preview` listens on VITE_PORT + 100 and reuses the dev proxy config.
const PREVIEW_PORT = 8100;

// Reporter paths resolve against the process cwd, not the config file — pin them
// so CI artifact uploads have one deterministic location.
const here = (relative: string) => fileURLToPath(new URL(relative, import.meta.url));

export const AUTH_STATE_PATH = here('./.auth/user.json');

const TIERS: Record<BackendTier, { baseURL: string; installMocks: boolean }> = {
    local: { baseURL: process.env.E2E_BASE_URL ?? 'https://localhost:8443', installMocks: false },
    mock: { baseURL: `http://localhost:${PREVIEW_PORT}`, installMocks: true },
    stand: { baseURL: process.env.E2E_BASE_URL ?? '', installMocks: false },
};

if (!(rawTier in TIERS)) {
    throw new Error(`E2E_TIER must be one of ${Object.keys(TIERS).join('|')}, got "${rawTier}"`);
}

const tier = rawTier as BackendTier;

if (tier === 'stand' && !TIERS.stand.baseURL) {
    throw new Error('E2E_TIER=stand requires E2E_BASE_URL');
}

export default defineConfig<BackendOptions>({
    // Keep Playwright's default per-pixel tolerance. Strict comparison was tried and
    // reverted: the pinned container fixes the renderer but not the host, and text
    // rasterisation still differs by 300-400 pixels between a local run and CI —
    // consistently, not as jitter. Absorbing that needs a budget far larger than the
    // ~70 pixels a palette change moves, so pixels cannot police colour here.
    // cross/contrast.spec.ts measures colour numerically instead.
    forbidOnly: isCI,
    fullyParallel: true,
    globalTimeout: isCI ? 10 * 60_000 : undefined,
    outputDir: './test-results',
    // Cassette specs run only on the mock tier; specs/real/** run only against a
    // live backend, which authenticates once in the setup project and reuses
    // storageState.
    projects:
        tier === 'mock'
            ? [
                  isVisual
                      ? {
                            name: 'visual',
                            testMatch: '**/specs/visual/**',
                            use: { ...devices['Desktop Chrome'] },
                        }
                      : {
                            name: 'mock-chromium',
                            testIgnore: ['**/specs/real/**', '**/specs/visual/**'],
                            use: { ...devices['Desktop Chrome'] },
                        },
              ]
            : [
                  {
                      name: 'setup',
                      testMatch: '**/auth.setup.ts',
                  },
                  {
                      dependencies: ['setup'],
                      // The stand tier runs only @stand-tagged specs: flow-run
                      // drives a real (paid) agent run and must never target a
                      // shared stand.
                      ...(tier === 'stand' ? { grep: /@stand/ } : {}),
                      name: `${tier}-chromium`,
                      testMatch: '**/specs/real/**',
                      use: { ...devices['Desktop Chrome'], storageState: AUTH_STATE_PATH },
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
        backend: { installMocks: TIERS[tier].installMocks },
        baseURL: TIERS[tier].baseURL,
        ignoreHTTPSErrors: tier === 'local',
        locale: 'en-US',
        screenshot: 'only-on-failure',
        timezoneId: 'UTC',
        // Stand traces/videos embed the live session cookie and the setup
        // project's password fill; CI uploads test-results as a public-repo
        // artifact, so they must never be recorded on that tier.
        trace: tier === 'stand' ? 'off' : 'on-first-retry',
        video: tier === 'stand' ? 'off' : 'retain-on-failure',
    },
    // The gate must test the shipped artifact: keep this on the production build,
    // never a dev server.
    webServer:
        tier === 'mock'
            ? {
                  // The visual container cannot load host-built native vite
                  // binaries — it serves a pre-built dist with plain Node.
                  command: isVisual ? 'node e2e/tools/serve-dist.mjs' : 'pnpm run build && pnpm exec vite preview',
                  cwd: fileURLToPath(new URL('..', import.meta.url)),
                  env: { VITE_PORT: '8000', VITE_USE_HTTPS: 'false' },
                  // Never reuse a listener on the port: the build runs inside
                  // this command, so a reused (possibly orphaned) preview
                  // silently serves a previous commit's dist as the gate.
                  reuseExistingServer: false,
                  timeout: 240_000,
                  url: `http://localhost:${PREVIEW_PORT}`,
              }
            : undefined,
    workers: isCI ? 1 : undefined,
});
