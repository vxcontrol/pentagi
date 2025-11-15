// @ts-check
import { FlatCompat } from '@eslint/eslintrc';
import js from '@eslint/js';
import perfectionist from 'eslint-plugin-perfectionist';

const compat = new FlatCompat({
    baseDirectory: import.meta.dirname,
    recommendedConfig: js.configs.recommended,
});

const eslintConfig = [
    ...compat.config({
        extends: [
            'eslint:recommended',
            'plugin:@typescript-eslint/recommended',
            'plugin:react/recommended',
            'plugin:react/jsx-runtime',
            'plugin:react-hooks/recommended',
            'prettier',
        ],
        settings: {
            react: {
                version: 'detect',
            },
        },
    }),
    {
        rules: {
            '@typescript-eslint/no-unused-vars': [
                'error',
                {
                    argsIgnorePattern: '^_',
                    varsIgnorePattern: '^_',
                },
            ],
            '@typescript-eslint/no-explicit-any': 'warn',
            curly: ['error', 'all'],
            'padding-line-between-statements': [
                'error',
                {
                    blankLine: 'always',
                    next: 'return',
                    prev: '*',
                },
                {
                    blankLine: 'always',
                    next: 'block-like',
                    prev: '*',
                },
                {
                    blankLine: 'always',
                    next: '*',
                    prev: 'block-like',
                },
                {
                    blankLine: 'always',
                    next: 'block-like',
                    prev: 'block-like',
                },
                {
                    blankLine: 'any',
                    next: 'while',
                    prev: 'do',
                },
            ],
        },
    },
    perfectionist.configs['recommended-natural'],
    {
        ignores: ['node_modules/**', 'dist/**', 'build/**', 'public/mockServiceWorker.js', 'src/graphql/types.ts'],
    },
];

export default eslintConfig;
