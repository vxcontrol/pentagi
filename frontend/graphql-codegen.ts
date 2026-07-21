import type { CodegenConfig } from '@graphql-codegen/cli';

const config: CodegenConfig = {
    documents: './graphql-schema.graphql',
    generates: {
        './src/graphql/types.ts': {
            config: {
                dedupeFragments: true,
                enumType: 'native',
                exportFragmentSpreadSubTypes: true,
                inlineFragmentTypes: 'combine',
                scalars: { Time: 'string' },
                skipTypename: true,
                useTypeImports: true,
            },
            plugins: ['typescript-operations', 'typed-document-node'],
        },
    },
    hooks: {
        afterOneFileWrite: ['npx prettier --write'],
    },
    schema: '../backend/pkg/graph/schema.graphqls',
};

export default config;
