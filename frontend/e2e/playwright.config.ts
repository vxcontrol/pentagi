import { defineConfig, devices } from '@playwright/test';
import { fileURLToPath } from 'node:url';

import type { BackendOptions, BackendTier } from './fixtures/backend.ts';

const rawTier = process.env.E2E_TIER ?? 'mock';
const isCI = Boolean(process.env.CI);
const isVisual = process.env.E2E_VISUAL === '1';

// Baselines are pinned-container pixels. Gate on a marker the container run sets
// (run-visual.sh and the CI job), not on the host OS: a Linux workstation passes a
// bare platform check and an --update there would overwrite the baselines with
// host-font pixels the container then fails against.
if (isVisual && process.env.E2E_VISUAL_CONTAINER !== '1') {
    throw new Error(
        'visual snapshots must run in the pinned container — use e2e/tools/run-visual.sh ' +
            `(E2E_VISUAL_CONTAINER unset, platform "${process.platform}")`,
    );
}

// `vite preview` listens on VITE_PORT + 100 and reuses the dev proxy config.
const PREVIEW_PORT = 8100;

const here = (relative: string) => fileURLToPath(new URL(relative, import.meta.url));

export const AUTH_STATE_PATH = here('./.auth/user.json');

const TIERS: Record<BackendTier, { baseURL: string; installMocks: boolean }> = {
    local: { baseURL: process.env.E2E_BASE_URL ?? '', installMocks: false },
    mock: { baseURL: `http://localhost:${PREVIEW_PORT}`, installMocks: true },
    stand: { baseURL: process.env.E2E_BASE_URL ?? '', installMocks: false },
};

if (!(rawTier in TIERS)) {
    throw new Error(`E2E_TIER must be one of ${Object.keys(TIERS).join('|')}, got "${rawTier}"`);
}

const tier = rawTier as BackendTier;

// Both stand and local hit a real backend and bake the real (paid) flow-run specs, so neither may
// fall back to a default: a bare `E2E_TIER=local` must not silently target the dev stack. run-local-tier.sh
// supplies E2E_BASE_URL for the legitimate path.
if ((tier === 'stand' || tier === 'local') && !TIERS[tier].baseURL) {
    throw new Error(`E2E_TIER=${tier} requires E2E_BASE_URL`);
}

export default defineConfig<BackendOptions>({
    // Absolute, not a ratio: 1% of a 1280x720 baseline is 9,216 pixels, more than any one foreground
    // token covers, so a token could change hue with every baseline still matching. Baselines and
    // run both render in the pinned container (the host path is blocked above), so 200px absorbs the
    // glyph anti-aliasing without host-vs-CI noise.
    expect: {
        toHaveScreenshot: { maxDiffPixels: 200, threshold: 0.02 },
    },
    forbidOnly: isCI,
    fullyParallel: true,
    // Only the mock tier: the real tiers' serial worst case (auth.setup + two flow-run
    // attempts + the stand smokes) exceeds 10 min and would abort mid-retry, reporting
    // "timed out" instead of the real failure while the job still has budget.
    globalTimeout: isCI && tier === 'mock' ? 10 * 60_000 : undefined,
    outputDir: './test-results',
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
                            // @quarantine specs are documented as excluded from the gate; wire it here
                            // (there is no CLI grep) so they cannot fail the retry:0 mock run.
                            grepInvert: /@quarantine/,
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
        ? [['json', { outputFile: here('./test-results/results.json') }], ['github'], ['list']]
        : [['html', { open: 'never', outputFolder: here('./playwright-report') }], ['list']],
    // No retry on the hermetic mock tier: a pinned clock/locale/network with no real
    // backend makes a retry-recovered failure a real race, not infrastructure noise, and
    // "flaky" counts as pass. The real tiers keep one retry for genuine network flake.
    retries: tier === 'mock' ? 0 : 1,
    testDir: './specs',
    use: {
        backend: { installMocks: TIERS[tier].installMocks },
        baseURL: TIERS[tier].baseURL,
        ignoreHTTPSErrors: tier === 'local',
        locale: 'en-US',
        screenshot: tier === 'stand' ? 'off' : 'only-on-failure',
        timezoneId: 'UTC',
        // Stand traces/videos embed the live session cookie and the setup
        // project's password fill; CI uploads test-results as a public-repo
        // artifact, so they must never be recorded on that tier.
        // retain-on-failure, not on-first-retry: the mock PR gate runs retries:0,
        // so on-first-retry would never write a trace — yet the docs and the PR
        // comment tell you to open trace.zip from that tier's report.
        trace: tier === 'stand' ? 'off' : 'retain-on-failure',
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
                  // Pin PORT so an ambient PORT can't move serve-dist off the port Playwright waits on.
                  env: { PORT: String(PREVIEW_PORT), VITE_PORT: '8000', VITE_USE_HTTPS: 'false' },
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
