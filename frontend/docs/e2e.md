# Frontend E2E tests (Playwright)

End-to-end tests for the PentAGI UI. The default tier runs fully offline against
network mocks — **no backend, no secrets, no VPN, no LLM keys** — so anyone
(including fork-PR authors) can run it, and CI runs it on every pull request.

Driving a real stack by hand — through the browser and through raw requests, to
prove a fix or to reach a surface no cassette covers — is a different job with
its own traps: [`live-testing.md`](./live-testing.md).

## One-time bootstrap

```bash
cd frontend
nvm use                  # Node from .nvmrc
corepack enable          # once per Node install, activates the pinned pnpm
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

Retries differ by tier, and that changes what a green means. The mock tier runs
`retries: 0` — a pinned clock, locale and network make a retry-recovered failure a
real race rather than infrastructure noise. The `local` and `stand` tiers run
`retries: 1`, so a spec that fails once and passes on retry is reported **flaky**
and the run still exits `0`. That is an unfinished question, not a pass:
`specs/real/flow-run.spec.ts` allows 30s for the redirect off `/flows/new` after
**Submit**, which a host busy with a docker build misses while the retry passes in
seconds. Re-run the tier on an idle host before treating a flaky result as a
defect — and if it survives that, it is a race worth fixing, not worth tagging.

**A wall of timeouts is the host, not the diff.** The mock tier runs `retries: 0`
and every worker carries its own jsdom/browser context, so a machine that is busy
with something else turns the 5s `expect` deadline into mass failure. Measured on
one box while a game held ~5 cores: the unit suite reported 130 failures across 39
files, `--maxWorkers=3` on the same tree reported 15, and each of those files
passed alone. The signature is unmistakable — every failure reads
`Test timed out` / `Hook timed out`, most of them in a setup step (waiting for a
header button to enable) rather than in the assertion the spec is about, and the
failing files have nothing to do with what changed. Check `uptime` and re-run one
failing file by itself before reading a single trace; the same suite went
`1472 passed` on the idle box minutes later.

**A repeat on the retry is the diff; a timeout is the host.** The paragraph above
sends you to the host first, and that is right for a wall of `Test timed out`.
It is wrong for a single spec that fails twice with the same assertion message:
measured on 10.09.2026 at load average 26.8, `specs/real/flow-run.spec.ts` — the
one this file names as the usual suspect — passed in 16s while
`specs/real/account-password.spec.ts` failed identically on the run and on its
retry, and that was a real regression in the branch. Read the failure text
before reading `uptime`.

Read the outcome from `e2e/test-results/.last-run.json`, not from the exit status
of a pipeline: `pnpm e2e | tail` reports `tail`'s status, so a run whose summary
you piped away says `0` while `.last-run.json` says `"status": "failed"`. Both
tiers write that file, and it is the one artifact a pipe cannot rewrite.

A finished Tier-2 run leaves the backend and mock-LLM logs in
`e2e/test-results/compose.log`, captured before `down -v` destroys the containers;
CI publishes that directory as `e2e-local-report` (`e2e-report` is the mock job's).
`E2E_KEEP_STACK=1` writes no such file — the stack is still up, so read the logs
from it.

**`specs/real/flow-run.spec.ts` passes, so a red run is a signal, not noise to wave
through.** Its failures show as a 30s wait for the redirect off `/flows/new` after
**Submit**, or a 90s wait for `Linux` in the terminal buffer. Before blaming the
branch under test, build its base commit in a worktree and run the spec there: the
comparison says whether the branch broke it. Everything else in the tier passes.

**The mock tier cannot test the order a collection arrives in.** The cassette's
array *is* the order — `createdAt` on a fixture changes nothing, because no
server sorts it. A spec written against timestamps there passes whatever you do
to them. What the tier can hold is the client half: that nothing re-sorts what it
was handed (see `flows/tabs.spec.ts`). Server ordering is held by
`e2e/list-order-gate.unit.test.ts` for every collection the cache mirrors, and by
`backend/pkg/server/services/resources_order_gate_test.go` for the one REST
collection that has its own gate. It
reads `../../backend/sqlc/models` directly (so it throws in a frontend-only
checkout), compares each collection's direction with what `src/lib/list-order.ts`
declares, and separately requires every `created_at` order inside
`screenshots.sql` and `subtasks.sql` to be ASC. Three collections are declared
`unmirrored` — `flowFiles`, `knowledgeDocuments`, `resources` — because the cache
does not follow the server's order for them. Two of the three are checked by
nothing; `resources` is the exception, held by the Go gate named above.

**A page that refetches instead of writing the cache needs two answers.**
`/settings/api-tokens` answers each of its three subscription frames by refetching
`apiTokens` rather than writing the collection, so a cassette that lists one answer
serves the same list back and the case proves nothing. List the state before the
event and the state after it, in that order — entries are consumed in turn.

Note also that two mechanisms deliver such an event, and either one alone is enough:
the central `subscriptionToCacheFieldMap` in `src/lib/apollo.ts` writes the
collection, and the page refetches on top of it. A mutation test that silences one
of them will not redden the spec; silence both.

Three more structural gates live in `src/lib/apollo.test.ts`: every collection a
subscription writes to declares an order, every subscription in the operations
document either routes to a cache field or is named as writing no list, and every
membership subscription selects at least what its collection's query reads.

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

A red visual run on a busy machine is usually the machine. The specs wait 5s for
their landmark and then shoot the page, so a starved container yields both kinds of
noise: "element(s) not found" after 30s and a 6%-of-pixels diff on a page that had
not finished painting — and it picks different specs each run, which is the tell.
Measured on a host carrying dozens of other containers: 4 then 6 failures at the
default parallelism, 20 of 20 with `--workers=1`. Pass it through:
`E2E_SKIP_BUILD=1 ./e2e/tools/run-visual.sh --workers=1`.

## Debugging a red CI run

1. Open the failed run (link in the PR comment) and download the `e2e-report`
   artifact.
2. `pnpm exec playwright show-trace <path-to>/trace.zip` — full timeline,
   network, console, and DOM snapshots for each failed test.
3. Reproduce locally with `CI=1 pnpm e2e`.

A red run with `0 failed` and a large skipped count is not a test failure: the CI
mock job is capped at ten minutes of wall clock (`globalTimeout`, one worker), and
Playwright reports the specs it never reached as skipped, so `results.json` keeps
`unexpected` at zero and only the job's own conclusion carries the abort. That is
duration creep — compare the `e2e-trend` artifact against earlier runs instead of
opening traces or re-running the job.

A spec that fails once in ten runs, with a failure snapshot showing the page it
came *from*, is usually asserting a state the product only passes through. The
clone spec did exactly that: the providers fixture ships `custom` disabled, the
create form bounces a clone of a disabled type straight back to the list, and the
assertion won whenever it caught the form in the ~50 ms before the redirect. The
trace tells this apart from a slow render — its frame snapshots carry the address,
and there the address had already gone back. The fix is to assert something that
survives: the value a field was seeded with, the toast, the address the app
settles on. `--repeat-each=10 --workers=4` reproduces this class of flake in under
a minute.

The same trap wears a second face: a role query — and `toBeHidden()` with it — answers
"gone" for the whole page while a modal dialog stands, because Radix marks the rest of the
tree inert. A spec that reads a row straight after confirming a delete therefore passes
whether or not the product removed anything. Wait for the dialog to leave
(`await expect(page.getByRole('dialog')).toHaveCount(0)`) before asserting on what is behind
it — proven here by unwiring `flowFileDeleted` from `subscriptionToCacheFieldMap`, which the
spec only noticed once that wait was in place.

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
  mock-llm/              # the deterministic agent transcript Tier 2 answers with
  tools/                 # run-local-tier.sh, run-visual.sh, review-sandbox.sh, trend.mjs,
                         # schema-compat.mjs, serve-dist.mjs, affected.ts and its
                         # default-base-ref.ts
  routes.ts              # the route manifest `affected.ts` reads
  specs/                 # *.spec.ts, tagged (@smoke, …)
  *.unit.test.ts         # structural gates — vitest, NOT playwright. They live at the
                         # e2e root AND under helpers/, mocks/ and tools/, so a glob
                         # anchored at the root alone misses most of them
```

The `e2e/**/*.unit.test.ts` files run under `pnpm test` (`vitest.config.ts`
includes them), not under `pnpm e2e`: they read source and SQL rather than
driving a browser.

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
- **The clipboard is the second hole.** `navigator.clipboard.writeText` needs a
  permission this context does not hold, and the app turns the rejection into a
  "failed to copy" toast, so a spec asserting the happy path would fail for a reason
  that has nothing to do with the product. `helpers/clipboard.ts` records the writes
  from an init script instead; install it before the first navigation and assert on
  `clipboardWrites(page)`.
- **Drag and drop needs no synthetic events.** The file manager moves rows through
  HTML5 DnD with its own MIME type, and `locator.dragTo()` drives it as the browser's
  own drag — the drop lands, `dataTransfer` carries the paths, and the move request
  goes out.
- Assert errors via `pageerror`/`unhandledrejection` (`expectCleanPage`): the
  production bundle strips app console output, so console-based asserts are
  meaningless on the mock tier.
- **Scope a panel query by the panel's name.** The flow page keeps **two**
  tabpanels marked `data-state="active"` at once, so
  `locator('[role="tabpanel"][data-state="active"]')` reads out of both — measured,
  with the second holding no links, which is the only reason such a query ever
  matched what its spec expected. Radix labels each panel with its trigger, so
  `getByRole('tabpanel', { name: 'Screenshots' })` names exactly one.
- **The screenshots panel is bottom-anchored and mounts its images on
  intersection.** It opens scrolled to the newest shot, and a card that has not
  been in the viewport renders no `<img>` at all — a fixture with enough rows to
  overflow leaves the topmost ones imageless, so `getByRole('img')` never resolves
  there. Assert order off each card's source link, which renders regardless, and
  point an image-decoding assertion at the shot the panel actually opens on.

## Tags

Specs are tagged (`test.describe(..., { tag: '@x' })`) so runs can be filtered
with `--grep` / `--grep-invert` (e.g. `pnpm e2e --grep @smoke`):

| Tag         | Meaning                                                               |
| ----------- | --------------------------------------------------------------------- |
| `@smoke`    | Sanity subset — auth, nav, the load-bearing happy paths               |
| `@flows`    | Flow list / detail / subscription / terminal specs                    |
| `@crud`     | Create-read-update-delete journeys (knowledge, api tokens, templates) |
| `@coverage` | Surface coverage (dashboard, settings, resources)                     |
| `@settings` | The account page: identity, password validation, submit and rejection |
| `@cross`    | Cross-cutting: themes, responsive, a11y, contrast, route sweep        |
| `@visual`   | Screenshot baselines — runs only in the visual project                |
| `@real`     | Tier 2 — real backend + mock LLM (`specs/real/**`)                    |
| `@stand`    | Tier 3 — LLM-independent smoke against a live stand                   |

Two conventions the gate reserves:

- **`@quarantine`** — tag a newly-flaky spec to drop it from the gate (leave a
  tracking note); fix and untag rather than let it rot. The exclusion is wired
  into the mock project's `grepInvert`, not a CLI flag, and the real-tier projects
  do not read it: tagging a `specs/real/**` spec isolates it from nothing.
  Nothing is quarantined today.
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
  `e2e/routes.ts`). Empty output = no frontend route changed. Without a base
  argument it diffs against the branch the remote calls default, and with no
  remote branch to compare with it exits 2 rather than print nothing. This is the
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
