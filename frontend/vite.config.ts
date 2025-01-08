import { readFileSync } from 'node:fs';
import path from 'node:path';

import react from '@vitejs/plugin-react-swc';
import { defineConfig, loadEnv } from 'vite';
import { createHtmlPlugin } from 'vite-plugin-html';
import tsconfigPaths from 'vite-tsconfig-paths';

import { getGitHash } from './scripts/lib.ts';

const pkg = JSON.parse(readFileSync('package.json', 'utf8'));
const readme = readFileSync('README.md', 'utf8');

export default defineConfig(({ mode }) => {
    const viteEnv = loadEnv(mode, process.cwd(), '');

    return {
        plugins: [
            tsconfigPaths(),
            react(),
            createHtmlPlugin({
                template: 'index.html',
                inject: {
                    data: {
                        title: viteEnv.VITE_APP_NAME,
                    },
                },
            }),
        ],
        resolve: {
            alias: {
                '@': path.resolve(__dirname, './src'),
            },
        },
        define: {
            APP_VERSION: JSON.stringify(pkg.version),
            APP_NAME: JSON.stringify(pkg.name),
            APP_DEV_CWD: JSON.stringify(process.cwd()),
            GIT_COMMIT_SHA: JSON.stringify(getGitHash()),
            dependencies: JSON.stringify(pkg.dependencies),
            devDependencies: JSON.stringify(pkg.devDependencies),
            README: JSON.stringify(readme),
            pkg: JSON.stringify(pkg),
        },
        server: {
            proxy: {
                '/api/v1': {
                    target: `https://${viteEnv.VITE_API_URL}`,
                    changeOrigin: true,
                    rewrite: (path) => path.replace(/^\/api\/v1/, ''),
                    secure: false,
                },
                '/api/v1/graphql': {
                    target: `wss://${viteEnv.VITE_API_URL}`,
                    changeOrigin: true,
                    rewrite: (path) => path.replace(/^\/api\/v1\/graphql/, '/graphql'),
                    wss: true,
                    secure: false,
                },
            },
            port: 5173,
            host: '0.0.0.0',
            https: {
                key: readFileSync('ssl/service.key'),
                cert: readFileSync('ssl/service.crt'),
            },
        },
    };
});
