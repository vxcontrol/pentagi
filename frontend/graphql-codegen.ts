import type { CodegenConfig } from '@graphql-codegen/cli';

const config: CodegenConfig = {
    documents: './graphql-schema.graphql',
    generates: {
        './src/graphql/types.ts': {
            config: {
                dedupeFragments: true,
                exportFragmentSpreadSubTypes: true,
                apolloReactCommonImportFrom: '@apollo/client/react',
                apolloReactHooksImportFrom: '@apollo/client/react',
                enumType: 'native',
                inlineFragmentTypes: 'combine',
                skipTypename: true,
                useTypeImports: true,
                withHooks: true,
            },
            plugins: ['typescript-operations', 'typescript-react-apollo'],
        },
    },
    hooks: {
        afterOneFileWrite: ['npx prettier --write'],
    },
    schema: '../backend/pkg/graph/schema.graphqls',
};

export default config;
