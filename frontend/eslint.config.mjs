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
            '@typescript-eslint/no-explicit-any': 'warn',
            '@typescript-eslint/no-unused-vars': [
                'error',
                {
                    argsIgnorePattern: '^_',
                    varsIgnorePattern: '^_',
                },
            ],
            curly: ['error', 'all'],
            'no-fallthrough': 'off',
            'no-restricted-syntax': [
                'error',
                {
                    message:
                        'Do not set mode/reValidateMode on useForm — use useAppForm, which owns the silent-until-submit form timing.',
                    selector:
                        'CallExpression[callee.name="useForm"] > ObjectExpression > Property[key.name="mode"], CallExpression[callee.name="useForm"] > ObjectExpression > Property[key.name="reValidateMode"]',
                },
            ],
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
                    blankLine: 'any',
                    next: 'block-like',
                    prev: 'case',
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
            // Off until React Compiler lands: its compatibility lint flags every RHF
            // watch() / useReactTable as unactionable noise (see OPEN-ISSUES §3).
            'react-hooks/incompatible-library': 'off',
            'react/no-unescaped-entities': 'off', // Allow quotes in JSX
            'react/prop-types': 'off', // TypeScript provides type checking
        },
    },
    {
        // useAppForm is the single owner of the form timing — the one place
        // allowed to set mode/reValidateMode on useForm.
        files: ['src/hooks/use-app-form.ts'],
        rules: { 'no-restricted-syntax': 'off' },
    },
    perfectionist.configs['recommended-natural'],
    {
        ignores: ['node_modules/**', 'dist/**', 'build/**', 'public/mockServiceWorker.js', 'src/graphql/types.ts'],
    },
];

export default eslintConfig;
