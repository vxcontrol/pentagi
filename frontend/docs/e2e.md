# Frontend E2E tests (Playwright)

End-to-end tests for the PentAGI UI. The default tier runs fully offline against
network mocks — **no backend, no secrets, no VPN, no LLM keys** — so anyone
(including fork-PR authors) can run it, and CI runs it on every pull request.

## One-time bootstrap

```bash
cd frontend
corepack enable          # once per machine, activates the pinned pnpm
pnpm install
pnpm e2e:setup           # downloads the Chromium build (~300MB)
```

On Linux, browser system dependencies may be needed once:
`pnpm exec playwright install --with-deps chromium` (requires sudo).
On Windows, use WSL — native Windows is untested.

## Running

```bash
pnpm e2e                 # mock tier (default): builds the production bundle,
                         # serves it via `vite preview`, mocks the whole API
pnpm e2e:ui              # same, in Playwright UI mode (watch/debug)
CI=1 pnpm e2e            # byte-identical reproduction of a CI run
                         # (same retries and trace policy; CI-style reporters)

E2E_TIER=local pnpm e2e  # against the local docker stack (https://localhost:8443)
E2E_TIER=stand E2E_BASE_URL=https://… pnpm e2e   # against a live stand
```

Tips for the mock tier: `vite preview` is reused between runs if you keep it
running (`reuseExistingServer`), so only the first run pays the build.

## Tiers

| Tier | Backend | Needs | Used for |
|---|---|---|---|
| `mock` (default) | Playwright route/WS mocks replaying cassettes against the **production bundle** | nothing | the PR gate; every functional spec |
| `local` | `docker compose` stack | docker | fidelity checks, real backend paths |
| `stand` | a live stand | `E2E_BASE_URL` + creds | deploy smoke, version-skew checks |

## Debugging a red CI run

1. Open the failed run (link in the PR comment) and download the `e2e-report`
   artifact.
2. `pnpm exec playwright show-trace <path-to>/trace.zip` — full timeline,
   network, console, and DOM snapshots for each failed test.
3. Reproduce locally with `CI=1 pnpm e2e`.

A test failing with `every API call the app made must have a cassette entry`
means the app issued a request the cassette does not cover — the failure lists
the exact method/operation. Add the missing entry to the spec's cassette.

## Layout

```
e2e/
  playwright.config.ts   # tiers via E2E_TIER; mock tier builds+serves the prod bundle
  fixtures/              # merged `test`: backend/cassette options, auth seeding, error log
  mocks/                 # cassette-driven GraphQL + REST + graphql-ws mock engine
    cassettes/           # typed cassettes (checked by `pnpm typescript`)
  helpers/               # shared asserts (page errors, …)
  specs/                 # *.spec.ts, tagged (@smoke, …)
```

Key conventions:

- **Selectors:** `getByRole`/`getByLabel` first; `data-testid` only where no
  accessible name exists (add it to the component in the same PR).
- **Cassettes are TypeScript modules** typed against the generated GraphQL
  types — schema or operation drift fails `pnpm typescript`, not the runtime.
- **The clock is pinned** (UTC, fixed epoch) on the mock tier: keep cassette
  timestamps on the `CASSETTE_EPOCH` day or date renders change under you.
- **Unmatched calls fail the test** — the mock tier is hermetic by design;
  nothing ever leaks to a real backend.
- Assert errors via `pageerror`/`unhandledrejection` (`expectCleanPage`): the
  production bundle strips app console output, so console-based asserts are
  meaningless on the mock tier.

The full design (tiers, CI topology, mocking contract, roadmap) lives in the
team's E2E testing plan; scenario coverage maps 1:1 to the scenario catalog IDs.
