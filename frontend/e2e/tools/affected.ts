// Prints the routes touched by a diff — one path per line — for selective runs
// and for scoping the exploratory agent. Empty output = no frontend route
// changed (caller should run the full suite).
//
// Usage: pnpm exec tsx e2e/tools/affected.ts <base-ref>
import { execFileSync } from 'node:child_process';
import { fileURLToPath } from 'node:url';

import { affectedRoutes } from '../affected-routes.ts';
import { ROUTE_MANIFEST } from '../routes.ts';
import { defaultBaseRef } from './default-base-ref.ts';

const repoRoot = fileURLToPath(new URL('../../..', import.meta.url));
const base = process.argv[2] ?? defaultBaseRef(repoRoot);

if (!base) {
    console.error('affected: no origin/HEAD, origin/master or origin/main to diff against; pass a base ref');
    process.exit(2);
}

const changed = execFileSync('git', ['diff', '--name-only', `${base}...HEAD`], { cwd: repoRoot })
    .toString()
    .split('\n')
    .filter(Boolean);

for (const route of affectedRoutes(changed, ROUTE_MANIFEST)) {
    console.log(route.path);
}
