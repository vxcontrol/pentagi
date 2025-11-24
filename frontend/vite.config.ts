import react from '@vitejs/plugin-react-swc';
import { existsSync, readFileSync } from 'node:fs';
import path from 'node:path';
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
    const viteHost = viteEnv.VITE_HOST || '0.0.0.0';
    const useHttps = viteEnv.VITE_USE_HTTPS === 'true';

    const sslKeyPath = viteEnv.VITE_SSL_KEY_PATH || 'ssl/server.key';
    const sslCertPath = viteEnv.VITE_SSL_CERT_PATH || 'ssl/server.crt';

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
        host: viteHost,
        port: vitePort,
        proxy: {
            '/api/v1': {
                changeOrigin: true,
                secure: false,
                target: `${useHttps ? 'https' : 'http'}://${viteEnv.VITE_API_URL}`,
            },
            '/api/v1/graphql': {
                changeOrigin: true,
                secure: false,
                target: `${useHttps ? 'wss' : 'ws'}://${viteEnv.VITE_API_URL}`,
                wss: `${useHttps}`,
            },
        },
        ...(useHttps && {
            https: {
                cert: readFileSync(sslCertPath),
                key: readFileSync(sslKeyPath),
            },
        }),
    };

    return {
        define: {
            APP_DEV_CWD: JSON.stringify(process.cwd()),
            APP_NAME: JSON.stringify(pkg.name),
            APP_VERSION: JSON.stringify(pkg.version),
            dependencies: JSON.stringify(pkg.dependencies),
            devDependencies: JSON.stringify(pkg.devDependencies),
            GIT_COMMIT_SHA: JSON.stringify(getGitHash()),
            pkg: JSON.stringify(pkg),
            README: JSON.stringify(readme),
        },
        plugins: [
            tsconfigPaths(),
            react(),
            createHtmlPlugin({
                inject: {
                    data: {
                        title: viteEnv.VITE_APP_NAME,
                    },
                },
                template: 'index.html',
            }),
        ],
        resolve: {
            alias: {
                '@': path.resolve(__dirname, './src'),
            },
        },
        server: serverConfig,
    };
});
