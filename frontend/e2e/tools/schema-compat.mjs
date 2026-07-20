// Deploy-skew pre-flight for the stand tier: validates the frontend's persisted
// GraphQL operations against the TARGET backend's live schema (fetched via
// introspection), so a renamed/removed field — e.g. a new frontend field the
// deployed backend doesn't yet expose — fails in one readable step here
// instead of as dozens of red specs.
//
// Usage: node schema-compat.mjs
//   E2E_BASE_URL  (required)  e.g. https://localhost:8444
//   E2E_USER / E2E_PASSWORD   (default admin@pentagi.com / admin)
import { buildClientSchema, getIntrospectionQuery, parse, separateOperations, validate } from 'graphql';
import { readFileSync } from 'node:fs';
import { fileURLToPath } from 'node:url';

const BASE_URL = process.env.E2E_BASE_URL;
const USER = process.env.E2E_USER ?? 'admin@pentagi.com';
const PASSWORD = process.env.E2E_PASSWORD ?? 'admin';

if (!BASE_URL) {
    console.error('schema-compat: E2E_BASE_URL is required');
    process.exit(2);
}

// A real stand has a valid cert. Only the local self-signed Tier-2 stack needs
// TLS verification relaxed — the operator opts in for that with
// `NODE_TLS_REJECT_UNAUTHORIZED=0 node …` in their own shell, never in code.

// Despite the name, this file holds the frontend's operations (named queries /
// mutations / subscriptions + fragments), not a schema.
const OPERATIONS_DOC = fileURLToPath(new URL('../../graphql-schema.graphql', import.meta.url));

const login = async () => {
    const response = await fetch(`${BASE_URL}/api/v1/auth/login`, {
        body: JSON.stringify({ mail: USER, password: PASSWORD }),
        headers: { 'content-type': 'application/json' },
        method: 'POST',
    });

    if (!response.ok) {
        throw new Error(`login failed: ${response.status}`);
    }

    const cookie = response.headers.get('set-cookie');

    if (!cookie) {
        throw new Error('login returned no session cookie');
    }

    return cookie.split(';')[0];
};

const introspect = async (cookie) => {
    const response = await fetch(`${BASE_URL}/api/v1/graphql`, {
        body: JSON.stringify({ query: getIntrospectionQuery() }),
        headers: { 'content-type': 'application/json', cookie },
        method: 'POST',
    });

    if (!response.ok) {
        throw new Error(`introspection failed: HTTP ${response.status}`);
    }

    const { data, errors } = await response.json();

    if (errors) {
        throw new Error(`introspection failed: ${JSON.stringify(errors)}`);
    }

    return buildClientSchema(data);
};

const main = async () => {
    const schema = await introspect(await login());
    const document = parse(readFileSync(OPERATIONS_DOC, 'utf8'));
    const operations = separateOperations(document);

    // "All 0 operations satisfied" is a broken gate, not a pass: a codegen
    // change that turns the doc into SDL/fragments must fail loudly.
    if (!Object.keys(operations).length) {
        throw new Error(`no operations found in ${OPERATIONS_DOC} — the document no longer contains operations`);
    }

    const failures = [];

    for (const [name, operation] of Object.entries(operations)) {
        const errors = validate(schema, operation);

        if (errors.length) {
            failures.push({ errors: errors.map((error) => error.message), name });
        }
    }

    if (failures.length) {
        console.error(`schema-compat: ${failures.length} operation(s) incompatible with ${BASE_URL}\n`);

        for (const { errors, name } of failures) {
            console.error(`  ✗ ${name}`);
            errors.forEach((message) => console.error(`      ${message}`));
        }

        console.error('\nThe deployed backend does not satisfy the frontend contract (deploy skew).');
        process.exit(1);
    }

    console.log(`schema-compat: all ${Object.keys(operations).length} operations satisfied by ${BASE_URL}`);
};

main().catch((error) => {
    console.error(`schema-compat: ${error.message}`);
    process.exit(2);
});
