import type { CodegenConfig } from '@graphql-codegen/cli';

const config: CodegenConfig = {
    schema: './graphql-schema.graphql',
    documents: './graphql-schema.graphql',
    generates: {
        './src/graphql/types.ts': {
            plugins: ['typescript', 'typescript-operations', 'typescript-react-apollo'],
            config: {
                skipTypename: true,
                dedupeFragments: true,
                preResolveTypes: true,
                inlineFragmentTypes: 'combine',
                exportFragmentSpreadSubTypes: true,
                useTypeImports: true,
                withHooks: true,
            },
        },
    },
};

export default config;
