import type { CodegenConfig } from '@graphql-codegen/cli';

const config: CodegenConfig = {
    documents: './graphql-schema.graphql',
    generates: {
        './src/graphql/types.ts': {
            config: {
                dedupeFragments: true,
                exportFragmentSpreadSubTypes: true,
                inlineFragmentTypes: 'combine',
                preResolveTypes: true,
                skipTypename: true,
                useTypeImports: true,
                withHooks: true,
            },
            plugins: ['typescript', 'typescript-operations', 'typescript-react-apollo'],
        },
    },
    schema: '../backend/pkg/graph/schema.graphqls',
};

export default config;
