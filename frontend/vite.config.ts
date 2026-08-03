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
        build: {
            chunkSizeWarningLimit: 1000,
            minify: 'terser',
            rollupOptions: {
                output: {
                    // A group captures its modules' dependencies too (`includeDependenciesRecursively`
                    // defaults to true), so a shared module must outrank the heavy libraries depending
                    // on it. At equal priority react/jsx-runtime lands inside `markdown` and clsx inside
                    // `charts`, and index.html then preloads both on every route.
                    codeSplitting: {
                        groups: [
                            {
                                name: 'react-vendor',
                                priority: 100,
                                test: /node_modules[\\/](react|react-dom|react-router-dom)[\\/]/,
                            },
                            {
                                name: 'utils',
                                priority: 90,
                                test: /node_modules[\\/](clsx|tailwind-merge)[\\/]/,
                            },
                            {
                                name: 'apollo-client',
                                priority: 50,
                                test: /node_modules[\\/](@apollo[\\/]client|graphql|graphql-ws)[\\/]/,
                            },
                            {
                                name: 'markdown',
                                priority: 50,
                                test: /node_modules[\\/](react-markdown|rehype-highlight|rehype-slug|remark-gfm)[\\/]/,
                            },
                            { name: 'radix-ui', priority: 50, test: /node_modules[\\/]@radix-ui[\\/]/ },
                            { name: 'terminal', priority: 50, test: /node_modules[\\/]@xterm[\\/]/ },
                            { name: 'charts', priority: 50, test: /node_modules[\\/]recharts[\\/]/ },
                            { name: 'pdf', priority: 50, test: /node_modules[\\/]@react-pdf[\\/]/ },
                            {
                                name: 'tiptap',
                                priority: 50,
                                test: /node_modules[\\/](@tiptap|prosemirror-[a-z-]+)[\\/]/,
                            },
                            // Without an own chunk, marked gets folded into the tiptap chunk, and report-pdf's
                            // static `import { marked }` would download all of it for a ~40KB library.
                            { name: 'marked', priority: 60, test: /node_modules[\\/]marked[\\/]/ },
                        ],
                    },
                },
            },
            sourcemap: false,
            terserOptions: {
                compress: {
                    drop_console: mode === 'production',
                    drop_debugger: mode === 'production',
                },
            },
        },
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
        // `vite preview` serves the production build, which is the only place chunking exists —
        // but it has no proxy of its own, so without this the built app cannot reach the API.
        preview: { ...serverConfig, port: vitePort + 100 },
        server: serverConfig,
    };
});
