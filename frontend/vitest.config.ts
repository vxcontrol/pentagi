import path from 'node:path';
import { defineConfig } from 'vitest/config';

// Vitest config lives in its own file so the production `vite.config.ts`
// stays free of test-only concerns (jsdom, setup files, etc).
//
// `jsdom` is the environment because most tests under `src/` reach into
// React Testing Library at some point — pure-function suites pay a small
// fixed cost (a few ms per file) but the alternative of per-file
// `@vitest-environment` comments fragments quickly.
export default defineConfig({
    resolve: {
        alias: {
            '@': path.resolve(__dirname, './src'),
        },
    },
    test: {
        environment: 'jsdom',
        globals: false,
        include: ['src/**/*.{test,spec}.{ts,tsx}'],
        setupFiles: ['./vitest.setup.ts'],
    },
});
