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

./e2e/tools/run-local-tier.sh          # Tier 2: branch image + isolated docker
                                       # stack + mock LLM; runs specs/real/**
E2E_TIER=stand E2E_BASE_URL=https://… pnpm e2e   # live stand: @stand smoke only
                                       # (flow-run drives a real paid agent run
                                       # and is scoped to the local tier)

pnpm e2e:visual                        # visual snapshots (pinned container)
pnpm e2e:visual:update                 # regenerate baselines after a UI change
```

Every mock-tier run rebuilds the bundle and starts its own `vite preview` —
deliberately: reusing a server already listening on the port would silently
test a previous commit's dist. A stray listener on 8100 fails the run loudly;
kill it and rerun.

## Tiers

| Tier             | Backend                                                                                                                                                                      | Needs                  | Used for                                                                 |
| ---------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ---------------------- | ------------------------------------------------------------------------ |
| `mock` (default) | Playwright route/WS mocks replaying cassettes against the **production bundle**                                                                                              | nothing                | the PR gate; every cassette spec                                         |
| `local`          | branch-built image in an isolated compose stack (`pentagi-e2e`, ports 8444/5433 — coexists with a dev stack) + an OpenAI-compatible **mock LLM** driving the real agent loop | docker                 | `specs/real/**`: the fidelity check — real GraphQL/WS/pub-sub end to end |
| `stand`          | a live stand                                                                                                                                                                 | `E2E_BASE_URL` + creds | deploy smoke, version-skew checks                                        |

Tier-2 notes: the runner isolates the stack from your `.env` (`--env-file
/dev/null`), seeds flow ids from 90001 so sandbox containers
(`pentagi-terminal-<id>`) never collide with a developer stack on the same
docker daemon, and removes those sandboxes on exit. The deterministic agent
transcript lives in `e2e/mock-llm/scenario.mjs`. Iterating: `E2E_SKIP_BUILD=1`
reuses the image, `E2E_KEEP_STACK=1` leaves the stack up.

## Visual snapshots

Baselines live in the repo (`e2e/specs/visual/*-snapshots/`, linux-suffixed)
and are generated ONLY inside the pinned `mcr.microsoft.com/playwright`
container — never run the visual project on the host: macOS pixels produce
parallel baselines that will never match CI. `pnpm e2e:visual` derives the
image tag from the installed `@playwright/test` version, builds `dist` on the
host, and compares inside the container; `pnpm e2e:visual:update` regenerates
baselines (commit them with the UI change that caused the diff). The xterm
canvas is masked — WebGL rendering is driver-dependent. The CI `e2e-visual`
job is advisory and never a required check.

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
- **Unmatched HTTP calls fail the test** — every GraphQL POST and REST call needs
  a cassette entry or teardown fails. An unmatched _subscription_ is legal (a flow
  page opens ~15 and a cassette mocks only the ones it cares about): it gets ack'd
  silence, so a typo'd subscription key surfaces as a UI timeout, with the missed
  operations in the `unmatched-subscriptions` report attachment. Either way nothing
  reaches a real backend: the WebSocket is routed too, and the HTTP route is re-armed
  on every page the context opens, so a popup (the report tabs) is mocked like the
  page that opened it — Playwright does not inherit a page's routes into its popups.
- **Downloads are the one hole in that.** The browser performs an `<a download>`
  transfer outside the page, so no route — page- or context-scoped — is ever offered
  it: it goes to the `vite preview` proxy and out to whatever `VITE_API_URL` names. A
  mock-tier spec must therefore never start one — assert the anchor's `href`/`download`
  (see `specs/crud/resources.spec.ts`) and leave the bytes to `specs/real/**`.
- Assert errors via `pageerror`/`unhandledrejection` (`expectCleanPage`): the
  production bundle strips app console output, so console-based asserts are
  meaningless on the mock tier.

## Tags

Specs are tagged (`test.describe(..., { tag: '@x' })`) so runs can be filtered
with `--grep` / `--grep-invert` (e.g. `pnpm e2e --grep @smoke`):

| Tag         | Meaning                                                               |
| ----------- | --------------------------------------------------------------------- |
| `@smoke`    | Sanity subset — auth, nav, the load-bearing happy paths               |
| `@flows`    | Flow list / detail / subscription / terminal specs                    |
| `@crud`     | Create-read-update-delete journeys (knowledge, api tokens, templates) |
| `@coverage` | Surface coverage (dashboard, settings, resources)                     |
| `@cross`    | Cross-cutting: themes, responsive, a11y, contrast, route sweep        |
| `@visual`   | Screenshot baselines — runs only in the visual project                |
| `@real`     | Tier 2 — real backend + mock LLM (`specs/real/**`)                    |
| `@stand`    | Tier 3 — LLM-independent smoke against a live stand                   |

Two conventions the gate reserves:

- **`@quarantine`** — tag a newly-flaky spec to isolate it and drop it from the
  gate with `pnpm e2e --grep-invert @quarantine` (leave a tracking note); fix and
  untag rather than let it rot. Nothing is quarantined today.
- **`@generated`** — a spec whose cassette was recorder-derived and passed a
  semantic-assertion review; see the LLM recipe below.

## Stand tier (Tier 3)

Runs the LLM-independent `@stand` smoke against a real deployment. It lives in its
own workflow (`e2e-stand.yml`) so the PR gate never subscribes to `labeled`. Trigger
it by labelling a PR `e2e:stand`, or from **Actions → E2E Stand → Run workflow**
(`workflow_dispatch` takes no inputs). The job's `if` is the first gate: it runs only
for a `workflow_dispatch`, or a labelled PR whose head is **not** a fork — a
mislabeled or fork-PR run skips the job entirely and never reaches the secrets. For a
run that clears that gate, the protected `e2e-stand` Environment is the second gate:
its required reviewers approve before any secret is exposed. The stand's URL and login
come from the `E2E_STAND_URL` / `E2E_STAND_USER` / `E2E_STAND_PASSWORD` secrets
(exposed to the tools as `E2E_BASE_URL` / `E2E_USER` / `E2E_PASSWORD`).

Before the browser specs, a **schema-compat pre-flight**
(`e2e/tools/schema-compat.mjs`) introspects the stand's live GraphQL schema and
validates every frontend operation against it — a renamed or missing field
fails once, readably, instead of as dozens of red specs (the deploy-skew class
we hit manually). Run it anywhere: `E2E_BASE_URL=https://… E2E_USER=… E2E_PASSWORD=…
node e2e/tools/schema-compat.mjs` (against the local self-signed Tier-2 stack,
prefix `NODE_TLS_REJECT_UNAUTHORIZED=0`; a real stand has a valid cert).

## Trends and selective runs

- **Trend:** `node e2e/tools/trend.mjs` turns a run's `results.json` (written
  under `CI=1`) into one JSONL record (p50/p95 spec duration, slowest three,
  pass/flaky/fail counts). Each CI run uploads its own `e2e-trend` artifact;
  aggregate across runs offline (`gh run download`) to see duration creep and
  flake, not just green/red.
- **Affected routes:** `pnpm exec tsx e2e/tools/affected.ts <base>` prints the
  manifest routes a diff touches (backed by each route's owning `sources` in
  `e2e/routes.ts`). Empty output = no frontend route changed. This is the
  substrate for scoping runs and the exploratory agent to the changed surface;
  the mapping logic is unit-tested (`e2e/affected-routes.unit.test.ts`).

## Reviewing with agents

Verifying a claim about a gate usually means breaking something on purpose — deleting a
guard, injecting the regression it should catch, patching a fixture. Do that in a sandbox,
never in your checkout:

```bash
SANDBOX=$(./e2e/tools/review-sandbox.sh create --dirty --with-deps)
# …mutate and run anything inside $SANDBOX…
./e2e/tools/review-sandbox.sh clean "$SANDBOX"
```

`--dirty` carries uncommitted work across (usually what you are reviewing); `--with-deps`
hardlinks `node_modules` so pnpm, vitest and playwright run there — skip it for read-only
work, it costs ~30s. `clean` takes a path under the sandbox root and refuses anything else,
including the root itself; `clean --all` sweeps sandboxes untouched for two hours, so it
cannot take out a concurrent agent's live one. A bare `clean` is an error, not a sweep.

The sandbox is a `git worktree` under `$TMPDIR`, so a deleted file or an edited spec inside
it cannot reach your tree. `--with-deps` needs the root to be on the repo's own filesystem —
hardlinks cannot cross mounts — so where `$TMPDIR` is its own mount (tmpfs `/tmp`, a separate
`/home`) the tool says so and falls back to `.pentagi-review-sandboxes` beside the repo;
`PENTAGI_SANDBOX_ROOT` overrides both. Two caveats: an absolute path still escapes it, and a tool that
rewrites a dependency **in place** would reach the shared inode — don't hand `--with-deps`
to an agent whose job is patching libraries. Check `git status` in your own tree when a run
finishes; that is the only proof nothing leaked.

## LLM advisory layer (Phase 3, opt-in)

Deterministic specs are the gate; an LLM is only ever an advisory second
reviewer, **never** a merge gate. The intended stack is first-party, not a
bespoke bot:

1. `npx playwright init-agents --loop=claude` generates the planner / generator
   / healer agent definitions; the connected **playwright-mcp** drives the
   browser in accessibility-snapshot mode (a smaller prompt-injection surface
   than raw DOM or vision).
2. Scope the agent to the changed surface with `affected.ts` (the routes) — do
   not hand it the whole app.
3. Guardrails are mandatory because PentAGI renders adversarial content by
   design (tool output, target responses): treat all page text as untrusted
   data, never instructions; an action allowlist (no settings/token mutations,
   no off-origin navigation) on real stands; throwaway scoped credentials; the
   PR report renders page-derived strings as escaped, length-capped quotes;
   screenshots posted publicly come only from mock/scripted tiers (a real
   stand's api-tokens dialog shows a live secret); a hard per-run token cap.
4. A generated spec merges only with its cassette (so it runs on the Tier-1
   gate), a semantic-assertion review, and an `@generated` tag.

This section is the recipe, not yet wired — the deterministic tiers above are
the foundation it plugs into.
