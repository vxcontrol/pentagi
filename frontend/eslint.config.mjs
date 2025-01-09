// @ts-check
import { defineConfig } from 'eslint-config-hyoban';

export default defineConfig(
    {
        formatting: {
            quotes: 'single',
            arrowParens: true,
            braceStyle: '1tbs',
            lineBreak: 'after',
            semi: true,
            indent: 4,
        },
        lessOpinionated: true,
        preferESM: false,
        ignores: ['public/mockServiceWorker.js', 'src/components/ui', 'pnpm-lock.yaml', 'src/graphql/types.ts'],
    },
    {
        settings: {
            tailwindcss: {
                whitelist: ['center'],
            },
        },
        rules: {
            '@stylistic/indent': ['error', 4],
            'unicorn/template-indent': ['error', { indent: 4 }],
            '@stylistic/quote-props': ['error', 'as-needed'],
            '@stylistic/operator-linebreak': [
                'error',
                'after',
                {
                    overrides: {
                        '?': 'before',
                        ':': 'before',
                    },
                },
            ],
        },
    },
);
