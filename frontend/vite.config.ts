import { existsSync, readFileSync } from 'node:fs';
import path from 'node:path';

import react from '@vitejs/plugin-react-swc';
import { defineConfig, loadEnv } from 'vite';
import { createHtmlPlugin } from 'vite-plugin-html';
import tsconfigPaths from 'vite-tsconfig-paths';

import { generateCertificates } from './scripts/generate-ssl.ts';
import { getGitHash } from './scripts/lib.ts';

const pkg = JSON.parse(readFileSync('package.json', 'utf8'));
const readme = readFileSync('README.md', 'utf8');

export default defineConfig(({ mode }) => {
    const viteEnv = loadEnv(mode, process.cwd(), '');
    const vitePort = viteEnv.VITE_PORT ? Number.parseInt(viteEnv.VITE_PORT, 10) : 8000;
    const viteHost = viteEnv.VITE_HOST ? viteEnv.VITE_HOST : '0.0.0.0';
    const useHttps = viteEnv.VITE_USE_HTTPS === 'true';

    const sslKeyPath = 'ssl/server.key';
    const sslCertPath = 'ssl/server.crt';

    if (useHttps && (!existsSync(sslKeyPath) || !existsSync(sslCertPath))) {
        console.log('SSL certificates not found. Attempting to generate them...');
        try {
            generateCertificates();
        } catch {
            console.warn('Failed to generate SSL certificates. Falling back to HTTP.');
            process.env.VITE_USE_HTTPS = 'false';
        }
    }

    const serverConfig = {
        proxy: {
            '/api/v1': {
                target: `${useHttps ? 'https' : 'http'}://${viteEnv.VITE_API_URL}`,
                changeOrigin: true,
                secure: false,
            },
            '/api/v1/graphql': {
                target: `${useHttps ? 'wss' : 'ws'}://${viteEnv.VITE_API_URL}`,
                changeOrigin: true,
                wss: `${useHttps}`,
                secure: false,
            },
        },
        port: vitePort,
        host: viteHost,
        ...(useHttps && {
            https: {
                key: readFileSync(sslKeyPath),
                cert: readFileSync(sslCertPath),
            },
        }),
    };

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
        server: serverConfig,
    };
});
