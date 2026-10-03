# Live testing (chrome-devtools MCP)

The Playwright tiers in [`e2e.md`](./e2e.md) are the gate: deterministic, offline,
run on every PR. This document covers the other half — driving a **real stack**
(real Postgres, real GraphQL, real agent loop) through a browser and through raw
HTTP, which is what proves a fix on real data and what covers surfaces no
cassette reaches.

The two are not interchangeable:

|                      | deterministic tiers                   | live pass                                               |
| -------------------- | ------------------------------------- | ------------------------------------------------------- |
| answers              | "does the code still do what it did?" | "does this actually work against a real stack?"         |
| runs                 | every PR, minutes                     | by hand, tens of minutes                                |
| evidence it produces | pass/fail                             | before/after on a named input                           |
| catches              | regressions                           | wrong assumptions, environment truths, missing coverage |

Some paths exist only here, but fewer than it looks. The mock cassettes carry
`answer`, `input`, `report` and `terminal` messages and a `thinking` field on
them; a local-tier run adds `thoughts` and `done` through the mock LLM, and the
subscription plumbing that delivers all of them is covered by both tiers — so
`pnpm e2e` and `run-local-tier.sh` already answer for those render paths. What
neither tier produces is `advice`, `ask`, `browser`, `file` and `search`: those
are exercised by a live agent run and nothing else.

A live green proves nothing about your branch unless the stack **contains** your
branch. That check comes first, every time.

## Which stack

|                                                               | what it costs | what it proves                      |
| ------------------------------------------------------------- | ------------- | ----------------------------------- |
| `cd frontend && pnpm dev` (:8000) against an existing backend | seconds       | frontend changes only               |
| full local compose stack, image built from your branch        | ~10 min       | anything, including backend changes |

The dev server is the trap: it proxies `/api` to whatever backend you point it
at, so a backend fix you just wrote is **not** in the picture — a green result
there says nothing about it. Use it for UI work; the moment the change touches Go,
build an image.

You point it with `VITE_API_URL` in `frontend/.env`, which is gitignored and has
no example file — a fresh checkout has none, the proxy target resolves to
`https://undefined`, and every request fails before it reaches any backend. Write
it before the first `pnpm dev`:

```bash
printf 'VITE_API_URL=localhost:8443\nVITE_USE_HTTPS=true\n' > frontend/.env
```

The value carries no scheme: `VITE_USE_HTTPS` picks it, for the proxy, for the
subscription socket's `wss://`, and for TLS on the dev server itself — which is
why the UI is then at `https://localhost:8000`. Against a stack brought up with
`SERVER_USE_SSL=false` it must be `false` too.

## Bringing up the stack

```bash
docker build -t local/pentagi:live .                       # from the repo root
PENTAGI_IMAGE=local/pentagi:live \
  docker compose up -d --wait --wait-timeout 300 pentagi
```

The app answers on `https://localhost:${PENTAGI_LISTEN_PORT:-8443}` with a
self-signed certificate — every `curl` below therefore uses `-k`.

**Prove the container runs the image you just built.** `--wait` returns when the
container is healthy, not when it is the right one; compose reuses a running
container whose config did not change:

```bash
docker inspect -f '{{.Config.Image}} {{.Image}}' pentagi
docker images --no-trunc --format '{{.Repository}}:{{.Tag}} {{.ID}}' | grep live
```

The two ids must match. A stale container from an earlier session is the single
most common way a live result lies — it is why "I tested it and it works" is
worth nothing without this line.

The image is built from what is committed plus whatever is in your working tree
at build time, so check `git status` before you build: an uncommitted experiment
gets baked in, and a colleague's half-finished edit in the same checkout gets
baked in with it.

Readiness, before touching the UI:

```bash
curl -sk -o /dev/null -w '%{http_code}\n' https://localhost:8443/api/v1/info
```

Backend startup replays flows and can take a minute; poll rather than sleep.

### Environment truths that decide what you can test

- **`COOKIE_SIGNING_SALT`** — while it is empty or the literal `salt`, the server
  refuses both to mint and to validate API tokens ("token creation is disabled
  with default salt"). Anything touching tokens or the bearer tier needs a real
  value: `COOKIE_SIGNING_SALT=<throwaway> docker compose up -d …`.
- **Provider credentials** — a provider type with no server-side credentials
  cannot be created at all, so a stack with no LLM keys cannot host a provider,
  and without a provider you cannot create a flow. The repository's own `.env` is
  deliberately keyless; start the test stand from a credentialed env file of
  your own (`docker compose --env-file .env.local-test up -d …`). That file is
  yours to create — every `.env.*` is gitignored, so no checkout carries one:
  copy `.env.example`, put one provider's real key in it, and give it a salt
  (see above). Then check what it actually enables — `{ settingsProviders
  { enabled { … } } }` answers in one request. A type with a placeholder key
  still shows as enabled and fails at the first call, so run the flow against
  the one type whose credentials are real.
- **`/settings/providers` is not that list.** The page renders
  `settingsProviders.userDefined` — providers a user created — while the
  env-configured types live in `enabled`, which only fills that page's **Add
  Provider** menu, and in the `providers` query behind the composer's picker. A
  credentialed stack with no user-created rows therefore reads "No providers
  configured" while its flows run on `anthropic`: the empty page is correct, and
  it is not a reason to skip the flow walk.
- **OAuth is not testable locally** — the Google and GitHub flows need a real
  callback origin. Cover the local-account paths live and stub the rest.
- **`PENTAGI_LISTEN_PORT`** — an `.env` carried over from another checkout can
  publish the stack on `443`, and everything you then run against `:8443` quietly
  hits nothing. Keep a dedicated env file for the test stand
  (`docker compose --env-file .env.local-test …`) rather than editing the one you
  develop with.
- **No flow ⇒ no container** — paths behind a running sandbox (container file
  listing, terminal input) answer `FlowFiles.ContainerNotRunning` and are simply
  unreachable on a keyless stack. Cover those at the handler level instead of
  pretending the live 400 is the case under test.
- **The compose file names an image, it does not build one.** The `pentagi`
  service is declared `image: ${PENTAGI_IMAGE:-vxcontrol/pentagi:latest}`, so a
  plain `docker compose up -d` starts a published build and every green you then
  collect belongs to somebody else's code. Build first and name it:
  `docker build -t local/pentagi:branch . && PENTAGI_IMAGE=local/pentagi:branch docker compose up -d`.
  Prove the running stack is yours before trusting it — the served bundle carries
  the operations it was built from
  (`docker exec pentagi grep -o 'value:"<yourNewField>"' /opt/pentagi/fe/assets/index-*.js`),
  and the schema answers for the backend half
  (`{ __type(name: "FlowFile") { fields { name } } }`).
- **GraphQL is at `/api/v1/graphql`.** `/graphql` is not registered at all, so it
  falls to the router's `NoRoute`, which redirects anything that is neither a
  frontend route nor an asset with `301` to `/`. That redirect looks exactly like
  an auth failure and will send you hunting for a session you already have — the
  real path answers `403` when the session is missing, and `/api/v1/info` names
  your user when it is not.
- **Sandbox ports and names** — a flow's sandbox is `pentagi-terminal-<flowID>`,
  and its host ports come out of a 2000-wide window from `DOCKER_PORTS_BASE`
  (see [`backend/docs/docker.md`](../../backend/docs/docker.md)), so flow ids
  1000 apart land on the same ports. A Tier-2 run seeds ids from 90001 and owns
  `[30000, 32000)`: its overlay sets `DOCKER_PORTS_BASE=30000` so that its flows
  cannot ask for the 28002/28003 a developer stack is already publishing. A stack
  you bring up by hand inherits the default window and will collide — give it a
  base of its own. Containers left behind by a previous database are orphans —
  their flow ids no longer exist, and they still hold ports.

  The name collides before the port does, and it fails the whole mutation:
  `createFlow` answers `Conflict. The container name "/pentagi-terminal-<id>" is
  already in use`, so a stack whose database you recreated cannot start a flow
  until the old sandboxes are removed. `run-local-tier.sh` handles both halves —
  it removes `pentagi-terminal-9xxxx` and their `-data` volumes, then runs
  `ALTER SEQUENCE flows_id_seq RESTART WITH 90001`. Bringing the same compose
  project up by hand does neither, and ids restart at 1 straight into a developer
  stack's names.

## Driving the UI

Use the **chrome-devtools MCP** tools. Not computer-use, not screenshots read by
eye: the accessibility tree is the contract the app promises, and it is what the
Playwright specs assert against too.

The loop is: `navigate_page` → `take_snapshot` → act on a `uid` → assert with
`evaluate_script` → `list_console_messages`.

```
navigate_page  {type: "url", url: "https://localhost:8443/flows"}
take_snapshot                       # a11y tree; every control carries a uid
click / fill   {uid: "…"}           # act through the tree, not coordinates
evaluate_script                     # read state back: text, computed styles, DOM
list_console_messages {types:["error"], includePreservedMessages:true}
```

Prefer a snapshot over a screenshot — it is smaller, diffable, and names things
the way a user's screen reader does. Take screenshots only when the question is
genuinely visual.

**Zero console errors is part of the pass**, on every route you visit.

### Traps that make a live check lie

- **The router ignores `history.pushState`.** Pushing a path and dispatching
  `popstate` changes the URL without re-rendering, and the assertions then read
  the previous page. Navigate for real, one page per `navigate_page`.
- **A first login lands on "Update Password".** The migrated admin carries
  `password_change_required`, so the SPA holds you there; click **Skip for now**
  or every later step runs against the login route. (The Playwright tiers clear
  the flag in setup instead — see [`e2e.md`](./e2e.md).)
- **Logging out, or changing a password, ends that user's other sessions.** The
  server keeps a generation counter per user; a cookie carries the value it was
  issued under and the middleware refuses one that lags, answering `403` with
  `session revoked` in the log. So the parallel `curl` jar goes dead the moment
  you log out or change the password in the browser — mint it again rather than
  reading the refusal as a broken probe. A self-service password change is the
  one case that keeps its own session: it re-stamps the caller's cookie. And the
  first run of a build that introduced the counter refuses every cookie issued
  before it, which is a one-time re-login, not a fault.
- **Controlled inputs ignore `el.value = …`.** React does not see it. Use the
  `fill` tool, or the native setter plus an `input` event:
  `Object.getOwnPropertyDescriptor(HTMLInputElement.prototype,'value').set.call(el, v)`.
  That covers filling, not clearing: `fill` with an empty string empties the DOM
  value while React keeps the old one, so the list stays filtered and the `?q=`
  stays in the address — which reads exactly like "clearing the filter does not
  restore the list". Clear with the keyboard instead: focus, select all, type or
  press Backspace.
- **Radix does not answer a scripted click.** Menus, tabs, selects and popovers
  open on pointer events, so `el.click()` from `evaluate_script` silently does
  nothing — the menu stays shut, the tab stays unselected, and the assertion that
  follows reads the previous state. Act on those through the `click` tool with a
  uid from a fresh snapshot. Plain buttons and links do respond to a scripted
  click, which is what makes the failure look arbitrary.
- **An inactive tab panel is unmounted.** Measuring a panel you have not switched
  to returns zero characters, which reads exactly like an empty state. Switch the
  tab, then measure the panel whose `data-state` is `active`.
- **Validation is reported in the field, not in a toast.** A submit that fails
  client-side leaves the page silent: no toast, no navigation, and the reason
  sits in `[data-slot="form-message"]` with `aria-invalid="true"` on the control.
  Read those before concluding the button did nothing.
- **Filling a field is not typing in it.** The flow composer takes 140
  characters through the fill tool and keeps **Submit** disabled: the value is
  in the textarea, the form state never saw it. One real keystroke into the same
  field enables the button. Assert on the control you are about to press, not
  on the value you just wrote, and type at least one character before concluding
  that a submit path is broken.
- **Forms guard navigation.** Leaving a dirty create/edit form opens **Unsaved
  changes** (Cancel / Discard / Save) and the click that "did not navigate" is
  waiting on it.
- **The self-signed interstitial has to be cleared once per tab.** Click through
  it (**Advanced → Proceed**) from the snapshot; you cannot type `thisisunsafe`,
  because that needs real keystrokes on a `chrome-error://` page where
  `evaluate_script` does not run. If the tab refuses to keep the exception, bring
  the stack up with `SERVER_USE_SSL=false` and drive it over `http://`.
- **Names in the tree are not unique.** A breadcrumb renders `role="link"` with
  the same accessible name as the sidebar item that led there; match on the uid
  from the snapshot, not on a name you assume is unambiguous.
- **The terminal is a WebGL canvas.** Pixels tell you nothing; read the xterm
  buffer through its handle.
- **A response shape is not the UI's shape.** The resources list returns
  `data.items`; asserting on `data.resources` produces a confident empty list and
  a green check that means nothing. Print the payload once before trusting a key.
- **Long calendar walks time out.** Paging a date picker three years forward is
  ~36 re-renders; drive it once and assert both boundary days from the same grid.
- **DevTools "Offline" does not cut an open WebSocket to localhost.** It blocks
  fetches — they fail immediately — while an established socket keeps delivering,
  so a tab that looks dead for a minute is still subscribed. Read that as "the
  emulation did not reproduce a drop", never as "the client failed to notice one":
  the cheap disproof is a mutation over `curl` while the tab sits offline, and a
  sidebar that updates says the channel is alive. Cut the connection from the
  server side instead — `docker restart pentagi` produces a real `close 1006`, a
  reconnect within a couple of seconds, and the `ws:reconnected` event the app
  hangs its reconcile on.
- **A GraphQL error's wording is not the resolver's — unless nothing classified
  it.** On any built image the presenter (`services/graphql_errors.go`) collapses
  a refusal, a missing row, a driver error and an unreachable dependency into
  seven fixed strings behind six codes; the original is attached only in develop
  mode, which every `docker build` defeats. What it does not recognise — a
  resolver's own validation, for one — reaches the client verbatim and carries no
  `extensions.code` on any build, so waiting for a code on `token creation is
  disabled with default salt` waits forever. Match the code where there is one,
  never a redacted sentence.
- **Everything is gzipped, except a ranged request.** The compression middleware
  sits on the root engine, so GraphQL, every REST route and the static bundle all
  answer `content-encoding: gzip` to a client that advertises it. Byte-exact
  ranged downloads are kept by a predicate instead of by scope: a request carrying
  `Range` is served uncompressed, and so is anything whose extension is already
  compressed. A client advertising `Accept-Encoding: gzip` and not decoding now
  reads every route as broken, not just the endpoint.
- **Forcing a logout through GraphQL needs a code.** The client dispatches its
  session re-check on `extensions.code` being `UNAUTHENTICATED` or `FORBIDDEN`
  and no longer reads the message, so a refusal from one of the resolvers that
  still writes bare `unauthorized: …` prose leaves the session exactly as it was.
- **An open Radix menu hides the rest of the page.** While a dropdown, popover or
  sheet is open, everything behind it carries `aria-hidden`, so every role query
  outside it finds nothing and reads as "the control disappeared". Two things
  follow: close the menu before looking elsewhere, and do not count right after
  pressing Escape — the close is animated, and a count taken in the same tick
  still returns zero. Wait for `[role=menu]` to detach.
- **Settings is a second shell.** `/settings/*` replaces the main sidebar with its
  own (Account, Providers, Prompts, API Tokens), so a walk that clicks "Flows"
  from a settings route waits forever. Leave through **Back to App** in the
  settings sidebar footer, and remember that a probe which ends inside settings
  starts the next one there.
- **The theme switcher is a tab group, not menu items.** It lives inside the user
  dropdown as Radix `Tabs` labelled `System theme` / `Light theme` / `Dark theme`,
  so `getByRole('menuitem', …)` waits forever. Open the menu, then act on
  `role=tab`.
- **Table sorting is a button inside the header cell.** Clicking the `th` itself
  does nothing and leaves `aria-sort="none"`, which reads exactly like sorting
  being broken. Click the button; the cycle is `none → ascending → descending`,
  and it is not written to the query string — it goes to the page's
  `table_4_<path>` slot instead, so a reload keeps it and a case that changes the
  sort leaks into the next one.
- **The empty state is a table row.** A search with no matches renders one `tr`
  carrying "No matches", so counting `tbody tr` reports one result where there
  are none. Assert on the text, not the row count.
- **`git checkout -- <file>` discards uncommitted work, not the last edit.** It is
  a tempting way to undo a mutation during a mutation test and it reverts the fix
  you are testing along with it — the tell is a suite that goes redder than the
  single mutation should explain. Copy the file aside and restore from the copy.

## Logging in

The first login is not a single step:

1. `navigate_page` to `https://localhost:8443/` — an unauthenticated session
   lands on `/login`.
2. `take_snapshot`; fill **Login** and **Password** through the `fill` tool (a
   raw `el.value = …` is invisible to React), then click **Sign in**.
3. The seeded admin carries `password_change_required`, so the next screen is
   **Update Password**. Click **Skip for now**. Skipping is the point: changing
   the password here invalidates the credentials every later session and every
   API probe uses.
4. You land on the `returnUrl` you came from, or `/flows/new` when there is
   none. From here on, **never type another URL** — see below.

Keep a parallel `curl` session (`-c jar.txt`) logged in as the same user. The
browser proves what a user sees; the cookie jar proves what the server stored,
and you will need both on every write.

## Navigate the way a user does

Typing a route into `navigate_page` is a different code path from clicking a
link: it remounts the app, refetches everything, and skips the transition that
actually breaks — the one where a provider keeps stale state, a subscription is
not torn down, or a list does not refresh after a mutation. Half the bugs worth
finding live in that transition.

So: reach every page by clicking the sidebar, a breadcrumb, a row, or a tab.
Direct navigation is for exactly two things — the first login, and re-entering
after a deliberate full reload (which is itself a check: the page must survive
F5 with its state).

Two corollaries:

- **After every mutation, look at the list you came from** without reloading.
  A row that only appears after F5 is a broken subscription or a missing cache
  update, and no unit test will tell you.
- **`history.pushState` is not navigation.** It changes the URL without
  re-rendering; assertions afterwards read the previous page.

## Sweeping the controls of a page

The rule is _every_ control, not the happy path: buttons, tabs, dropdown items,
right-click context menus, dialogs and their confirm buttons, toggles, selects.
Enumerate them mechanically rather than from memory — the snapshot is the
inventory:

```js
// list everything actionable on the current page, with its accessible name
[...document.querySelectorAll('button,[role="tab"],[role="menuitem"],a[href],[role="switch"]')]
    .map((el) => ({
        role: el.getAttribute('role') || el.tagName.toLowerCase(),
        name: (el.getAttribute('aria-label') || el.textContent || '').trim().slice(0, 40),
        disabled: el.disabled || el.getAttribute('aria-disabled') === 'true',
    }))
    .filter((c) => c.name);
```

Work that list top to bottom. Three kinds of control hide from a naive pass and
have to be sought out:

- **Hover-only** — row action cells are `opacity-0` until the row is hovered;
  they are in the tree the whole time, so click them by name.
- **Menu-only** — a control that exists only after its trigger is opened
  (**Open menu**, **Flow actions**, **Templates and resources**). Open the
  trigger, snapshot again, then act: uids from the previous snapshot are stale.
- **State-gated** — controls that render only for a given status. **Finish**
  disappears once a flow is Finished or Failed; the composer shows **Stop**
  instead of **Submit** while a flow is running; every assistant row's delete
  button hides once the flow **or the selected assistant** is terminal — a
  Finished assistant hides the button on all the rows, not only its own. If you
  never saw the alternative state, you did not test the control — seed one.

  On the Tier-2 stack a whole run settles in about 300 ms, which is shorter than
  the state you are trying to look at. Bring the mock up with a per-completion
  delay and **Stop**, the disabled composer and the working placeholder stay on
  screen long enough to click and read:

  ```bash
  MOCK_LLM_DELAY_MS=4000 docker compose -p pentagi-e2e --env-file /dev/null \
    -f docker-compose.yml -f docker-compose.e2e.yml up -d --force-recreate mock-llm
  ```

- **Roleless** — a `div` with `onClick` and nothing else. The message bubble's
  **Show thinking** / **Hide thinking** and **Show details** / **Hide details**
  are exactly that, so the snippet above never lists them and a role-based sweep
  reports the bubble as having no controls at all. Reach them by text:

  ```js
  [...document.querySelectorAll('div.cursor-pointer')].map((el) => el.textContent.trim());
  ```

A control with no accessible name is itself a finding: it is unreachable for a
screen reader and for every spec that matches by role and name. The four roleless
toggles above are that finding, already known — do not re-file them, but do click
them, because the content they reveal is tested nowhere else.

Two more sweeps come back empty for a reason that is not a defect. A **roving
tab strip** — the editor toolbar, the file tree — gives `tabIndex 0` to exactly
one child and `-1` to the rest, so tabbing finds one control and the arrows find
the others. And a **horizontally scrolled strip** with a hidden scrollbar keeps
its overflow in the DOM: at a narrow viewport the controls are present, clickable
by name, and invisible on a screenshot.

## The sidebar header carries the version, and it is a control

The header of both sidebars — the app's and the settings one — is a button, not a
label. It opens a panel that states which build is running and what the update
service thinks of it. Three things are worth a pass of their own.

**What it must say.** The trigger shows the shell's title — PentAGI in the app,
Settings in the settings — and the version under it;
the panel repeats the version, adds a `build` line with the revision, a channel
badge (`preview`, `stable` or `nightly` — the server cannot answer `disabled`), the
verdict with its own tone, the facts as pairs (`Installed`, `Available` when an
update is offered, `Checked`), and the release links. The `build` line is
unconditional: when the pipeline stamped no git revision, the server answers with
the sha256 prefix of the running binary instead, so an empty build line is a defect,
not a state.

**The version string depends on how the image was built, not on the code.** A stand
built without build args reports a bare `develop`: the builder stage defaults
`PACKAGE_VER` to that word and stamps it through ldflags, and the runtime reports it
unchanged. Build it the way the pipeline does — `source ./scripts/version.sh`, then
pass `--build-arg PACKAGE_VER --build-arg PACKAGE_REV` — and the binary reports
`2.1.0-ce.h<rev>`, which the panel shows as `v2.1.0-ce` with a `build <rev>`
line. Check which you are looking at before filing "the version is wrong":
`docker inspect <container> --format '{{index .Config.Labels "com.pentagi.version"}}'`.
The label is composed from the same two args and reads exactly what the binary
reports — `2.1.0-ce.h<rev>`, without the `v` the panel adds — so a disagreement
between the label and the binary is itself the defect.

**Version skew removes the header silently.** The panel asks for a field the older
backend does not have, the query fails validation, and the provider hands the panel
nothing — so the sidebar header loses its version line and the trigger disappears
from the tree, with no error on screen. The only sign is one console line,
`Failed to load resource: 422`. A frontend dev server pointed at a stale stand is
the usual way to see it, and it is worth reproducing deliberately after any change
to `versionInfoFragment`.

The indicator follows the same tones as `FlowStatusIcon` — green for up to date,
yellow for an update, muted for anything unknown — and appears on the trigger only
when an update is offered. Every other state says its piece inside the panel and
leaves the trigger quiet. The collapsed rail keeps only the mark and, for an
update, a dot in its corner.

A development stand is not a build the update service published, so it never reads
up to date and never offers an update. What it can show live: **Checking for
updates** until the first check completes — it starts at a random moment within two
minutes of the backend starting and may take up to 30 seconds — then **Update status
unknown**, whether the service answered or could not be reached; and **Update checks
are disabled** when `UPDATE_CHECK_INTERVAL` is `0` or the service could not be set
up. The rest are pinned by
`src/components/shared/version-panel.test.tsx`, which renders the panel against
every answer the service can give — read the verdicts there rather than trying to
provoke them on a stand.

## CRUD, entity by entity

Do the full cycle on each entity, and close every one at the trust boundary —
the toast is the UI's opinion, the API read-back is the fact.

Two rows depend on the stack, not on the code: a **flow** cannot be created
without a configured provider, and a **knowledge document** cannot be created
without a reachable embedder — the mutation answers "embedding provider is not
available" and nothing is written. On a stack without those, walk the read,
update and delete halves and say plainly which creates you could not exercise.

| entity             | create                                                               | read                         | update                                                               | delete                                  |
| ------------------ | -------------------------------------------------------------------- | ---------------------------- | -------------------------------------------------------------------- | --------------------------------------- |
| flow               | composer on `/flows/new`, both **Automation** and **Assistant** tabs | row in the list, detail page | inline rename (row menu and breadcrumb), favorite toggle, **Finish** | row menu → dialog, header menu → dialog |
| assistant          | first message with no assistant selected                             | picker popover               | follow-up message, **Stop** while running                            | trash in the picker row → dialog        |
| flow template      | `/templates/new`                                                     | list, detail                 | **Edit** opens the detail; **Rename** edits the title in place, in the row itself | row menu → dialog                       |
| knowledge document | `/knowledges/new`                                                    | list, detail                 | edit, rename                                                         | row menu → dialog                       |
| resource           | upload, create folder                                                | file manager list            | rename, move, copy                                                   | delete, bulk delete                     |
| provider           | create by type, clone                                                | settings list, detail        | edit fields, per-agent **Test**                                      | delete → dialog                         |
| prompt             | —                                                                    | list, detail                 | edit, **Reset System**, **Reset All**                                | —                                       |
| API token          | inline create row                                                    | list                         | rename, revoke/activate                                              | row menu → dialog                       |

Three creates are gated by something other than the obvious field, and each looks
like a dead button until you find it: a **knowledge document** needs an **answer
type** picked in the select (the message "Answer type is required" appears in the
field, not in a toast), an **API token** keeps **Submit** disabled until an expiry
date is picked in the row's date popover, and an **assistant** keeps its own
**Submit** disabled until a provider is chosen in the composer's "Select
Provider" dropdown. A knowledge create also goes through the embedder, so it
answers in seconds, not milliseconds — read the list back after the mutation
resolves, not after a fixed wait.

Two rows in the table read differently from the rest. A **prompt** row does not
open on click, only through its row menu → **Edit**; its reset entries are named
**Reset System**, **Reset Human** and **Reset All**, each shown only when that
half of the prompt actually differs from the default, so a prompt with a custom
system text offers two of the three and never the third. A **resource** row
carries the same menu on the button and on right-click, and a move onto an
occupied path opens a second dialog, **Replace existing item?** — cancelling it
must leave the target file untouched.

For each: create it, find it in the list **without a reload**, open it, change
it, watch the change land in the list, delete it, confirm the row is gone and a
direct read returns not-found. Then check the same object through the API — the
stored value is what matters, not the rendered one.

Cancel paths count too: open every confirmation dialog and dismiss it, and check
nothing was written.

### The furniture every table carries

Three controls sit around the filter box and none of them appear in the CRUD
table above. **Search in** picks which columns the filter looks at — narrow it
and a row that matched a second ago disappears, which reads as lost data.
**Columns** hides columns; a hidden column takes its sort header with it. The
filter input itself stops at 200 characters, so a paste longer than that is
truncated silently at the DOM and the query you think you ran is not the query
that ran.

All three, plus sorting and page size, persist per page under a localStorage key
shaped `table_4_<path>` — the `_4_` is the separator from `src/lib/storage-keys.ts`
and reads as "table *for* /flows", not as a user id, so two accounts sharing a
browser share the state. That is the mechanism behind "it behaved differently
yesterday": clear the key before claiming a table is broken, and clear it between
cases that change columns.

### Knowledge has two searches, and they are not the same search

The box in the page header is **semantic** and writes `?qs=`; the box below it is
a plain text filter over the rows already on screen and writes `?q=`. They
compose — `?q=` narrows what `?qs=` returned — and only the first one asks the
server. Two things follow. The semantic query is capped at 100 results, and the
detail page's Prev/Next walks that same capped array, so navigation inside a
search is navigation inside those 100. And a semantic search that matches nothing
renders the empty state of an empty library — same heading, same **New
Knowledge** button, no table at all — so "search returned nothing" and "there is
no data" look identical on screen; check the address bar before believing either.

`?page=` is inert here. The table's own pager keeps its position in component
state and the page passes it no page props, so a page number carried over from
another screen changes nothing — and `?page=1` is stripped from the address bar
everywhere by the shared table hook, since the first page is the canonical URL.

Creation gates on more than the answer type. **Document type** — `answer`,
`guide`, `code` — decides which second field appears, and switching it wipes the
other two subtype values while marking the form dirty, so a tester who flips the
type to look around leaves a form that warns about unsaved changes.

### The API token you can only read once

The token is created from an inline row on the settings page, and the secret
appears exactly once, in a dialog titled **API Token Created** with a **Copy
Token** button. Dismiss it and the value is gone for good — there is no second
view, and the row in the table offers only **Copy Token ID**, which is not the
secret. Any case that needs a working token must copy it in that dialog or start
over.

**The expiry picker offers tomorrow onwards** — today and every past day are
disabled — and the day you pick is the moment the token dies, not its last
working day: picking 16 Sep issues a token whose `exp` is 15 Sep 23:59:59, and
the row prints `00:00, 16 Sep`. A case that needs the token to work on a given
day has to pick the day after it.

The same page's header carries a **Developer tools** menu that opens **GraphQL
Playground** and **Swagger UI**. Both are live surfaces served by this stand and
neither is named anywhere else in this document.

## The markdown editor, which is the body of three entities

**What leaves the editor is markdown, not what was typed.** An address or a URL is
autolinked as it is entered, so a body typed as `Contact admin@example.com` reaches
the server as a link. Anything comparing the two — the anonymizer decides "nothing
was found" by comparing its answer with the current body — is comparing against the
marked-up form. A case built on the typed text instead of the stored one tests a
branch that cannot be reached.

**What was not touched is saved as it was written.** A body that was loaded keeps its
own layout through a save: the blank lines between its blocks and how many there were,
the spaces on a blank line, the bullet character, the numbers and the leading spaces of
its lists, every row of a table as the line it was (its padding, its pipes — bare or
escaped inside a code span — and a line of text written right under the table),
a code block as it stood (its fence, its indentation, a block written by indentation and
not by a fence), the fence of a code span, a bare URL, the link around an image,
the underline of a setext heading, the spaces a heading or a list ended with, a line break
that is not a hard break. A case that types one character and compares the stored body
with the original should find that character and nothing else outside the block
it was typed into; inside that block the editor writes its own form — a table row
with one space around each cell where the table was written that way and padded
to its column otherwise, `\|` for a pipe in a code span of that row, a code block between
fences or four spaces in. What is kept gives way wherever it would be read as something else:
under a list, a table or a fenced block written a few spaces in goes to the margin,
text goes there when the list would take it for its last item's, and
inside a list item a code block is written between fences at the item's column.
The body is stored as it was sent, the line break it ends with included:
neither the forms nor the backend trim it, and only the single-line fields
(title, question, description, code language) are trimmed.

A Go template action `{{ … }}` is literal text to the editor — nothing inside
it is read as markdown and a pipe in it is never escaped — and a line that
opens with a control action (`{{ range }}`, `{{ end }}`) right under a table is
the template's line, not a row. What still differs is not a defect of such a case:
a link typed into the editor, as opposed to one that was loaded,
is stored as `[url](url)`; a named HTML entity outside code and
outside an action is decoded (`&amp;` becomes `&`);
a list item that opens with indented code is written with a fence;
and a document with CRLF line endings is saved with LF.

A template's body, a knowledge document's content and both halves of a prompt are
the same component. The document tells you to "edit" those entities; this is what
you are editing, and none of it is reachable from the entity's own controls.

**Getting to raw and back.** The mode switch is not in the editor. It lives in
each page's header menu — **Template actions**, **Knowledge actions**,
**Prompt actions** — on a row labelled **View**, which holds two icon triggers:
**Rich editor** and **Raw source**. The row keeps its menu open while you switch.
The mode is component state: it resets on reload, on remount, and when you move
to another prompt. In raw mode the toolbar, the table grips and the popovers do
not exist — `role="toolbar"` is absent from the page — so a sweep run in raw
reports the whole group missing.

**The toolbar** is `role="toolbar"`, accessible name **Formatting**,
`data-slot="markdown-editor-toolbar"`. Left to right: a **Text style** dropdown
(Heading 1…6 and Text), then **Bold**, **Italic**, **Strikethrough**,
**Inline code**, **Link**, then a **List** dropdown (bullet, ordered, task), a
**Table** dropdown, then **Blockquote**, **Code block**, **Insert image**,
**Horizontal rule**, **Clear formatting**, and pinned to the right **Undo** and
**Redo**. Every control is icon-only: its visible text appears in a tooltip after
a hover delay, so search by accessible name, never by text.

Three things about it will make a check lie.

- The two dropdown triggers name their current value — **Text style: Heading 2**,
  **List: Bullet list**, **List: None**. A locator captured by name goes stale as
  soon as the caret moves.
- Below the `md` breakpoint the strip renders *under* the text area, and it
  scrolls horizontally with the scrollbar hidden: controls exist and answer to a
  click by name while being invisible on a screenshot.
- Inside a code block the five inline marks are disabled; inside a table cell
  every heading except **Text** is disabled. A disabled control is not a missing
  control.

**Insertion is one level down.** There is no toolbar button that inserts a table:
open the **Table** dropdown and the single item **Insert table** is there while
the caret is outside a table. Move the caret inside and the same trigger opens a
ten-item menu — **Header row**, insert row/column in four directions, an **Align
column** submenu, and the three deletions.

**The popovers.** **Link** opens a popover with a **Link URL** field and the icon
buttons **Apply link**, **Open link in new tab** and — only over an existing link
— **Remove link**. The same popover appears on its own when the caret lands
inside a link; clicking a link seats the caret instead of navigating. Its URL
validation is authoring-only: pasted markdown never passes through it, so a
security check written against this field proves nothing about what the viewer
renders. **Insert image** works the same way and has no file picker anywhere —
images are URLs, and "upload an image" has no surface to test.

**Table grips do not answer to a driver.** Hovering a cell with a real mouse
summons two buttons, **Column actions** and **Row actions** (`data-table-grip`),
portalled into `document.body`. They are driven by a document-level `mousemove`,
so a synthetic click or a programmatic focus never brings them up; their menus
are the only way to align a column or delete a row without the toolbar.

**What the editor never shows you.** An invalid body renders its message
`sr-only` — the only visible signal is a red border from `aria-invalid` on the
contenteditable. A check that looks for error text finds none and reads the
field as valid.

## The file manager, which the document used to cover in six words

Two routes mount it: `/resources`, and a flow's **Files** tab. It is
`role="tree"`, accessible name **File tree**; rows are `role="treeitem"` carrying
`data-path`. On a flow the rows sit under three synthetic groups — **Uploads**,
**Resources**, **Container** — which are not themselves selectable.

**It has a keyboard model, and the document never said so.** One Tab lands on the
active row; Arrow Up/Down move between visible rows without wrapping, Home/End
jump to the ends, Arrow Right expands a directory and Arrow Left collapses it,
Enter opens a file or toggles a directory. The handler is bound to the wrapper
rather than to the tree, so those keys also fire while focus sits on the
select-all checkbox or a sort header.

**Selection is not Finder's.** A plain click replaces the selection with the row
— and a directory brings its whole subtree, collapsed or not. Cmd/Ctrl+click
toggles a row, a directory all-or-nothing. Shift+click is *additive* over the
visible range, deliberately unlike Finder: a check asserting that a range click
replaces the selection is asserting the wrong contract.

**The header row** carries a three-state **Select all** checkbox
(`data-state="indeterminate"` in the middle state) that is nonetheless a two-way
toggle: clicking it while indeterminate selects everything, and it only clears
when everything was already selected. Its universe is the *visible* tree, so with
a search active it selects what the filter left. Beside it sits **Expand all** /
**Collapse all**, whose accessible name flips to name the next action — and which
operates on the full tree, expanding directories the filter is hiding.

**Sorting is three-position** — ascending, descending, cleared — on **Name**,
**Size** and **Modified**. Two traps: the accessible name states the *next*
action (**Sort by name (ascending)** → **… (descending)** → **Clear sorting on
name**), so a locator captured by name is stale after one click; and the arrow
glyph is inverted from intuition, ascending drawing a down arrow. Judge the order
by the rows, never by the arrow. The order is not persisted anywhere: a reload
returns to insertion order.

**Row actions live in two places with identical items.** Hover a row and an
ellipsis button named **Row actions** appears — it is `opacity-0` until hover,
so it is invisible to a screenshot and immediately available to a query by name.
Right-clicking the row opens the same items as a context menu. On `/resources` a
directory row has seven items (**Download**, **Copy path**, **New folder**,
**Upload files**, **Rename or move**, **Copy to…**, **Delete**) and a file row
five; right-clicking the empty padding under the last row opens a two-item menu
targeting the library root.

**The bulk bar** appears once anything is ticked: `<N> selected · <size>` on the
left, then **Download**, **Move to…**, **Delete** with **Copy to…** and **Copy
paths** under **More actions** — on a flow's Files tab, **Save as resources**
instead. Two things to know. The N counts *affected* rows including descendants,
so ticking one folder of three files reads "4 selected". And below the `sm`
breakpoint the buttons keep only their icons, so a text query for **Download**
fails while the button is right there.

**Column settings** on `/resources` toggles the Size and Modified columns and the
folders-first order. It persists per user and per path under a localStorage key
shaped `viewOptions_4_<path>` — keyed on the path, not on the user — which means a
case that changes columns leaks into the next one until you clear that key, and
into the next account on that browser too.

**Move, copy and pull all end in the same second stage.** The dialogs are
**Rename or move resource**, **Copy resource** and **Pull from container**;
dropping a row onto a folder takes the same path. When the target path is
occupied, a second dialog, **Replace existing item?**, is what actually
overwrites — cancelling it must leave the target untouched, and that is the case
worth running, because the first dialog looks like it already did the work.

## The settings surfaces that no case in this document opens

### Presets write over what you typed

A template detail page and `/templates/new` both carry a **Preset templates**
panel — a card in the left split panel at 1280px and up, a button that opens a
sheet below that. It holds eleven presets, and each row is two buttons: the title
applies the preset, and a second, nameless one expands a preview. A sweep by role
alone counts twenty-two controls where a sweep by role and name counts eleven.

Applying a preset onto a form that already has a title or a body opens a dialog
titled **Replace content?** — so on an existing template, where the form is
pre-filled from the server, the dialog always appears, and on `/templates/new` it
only appears the second time. Either way the apply marks the form dirty, which
arms the unsaved-changes blocker: the next navigation, including the pagination
arrows described below, stops on a confirmation the case did not plan for.

The same **Replace content?** dialog also guards the flow form, so a case that
learns the string on one page will find it on the other.

### Validate, and the counter that is not the counter next to it

The prompt editor's header carries a **Validate** button that checks the text
currently in the editor, not what is saved. While the mutation is in flight both
the label and the accessible name become **Validating...**, so a lookup by the name
**Validate** finds nothing for the duration — wait for the name to come back
rather than treating the miss as a missing button. Below 768px only the icon is
rendered; the name still resolves, the visible text does not.

Beside it sits **Available variables**: a card at 1280px and up, a popover behind
a button below that, and nothing at all when the active tab declares no variables
— the System and Human tabs have different variable sets, so the panel appearing
and disappearing as you switch tabs is correct. Two numbers there are easy to
swap. The badge in the heading counts variables *declared* for the tab; the `×N`
on a chip counts *uses* in the text, per `{{ … }}` block rather than per name, so
`{{.Foo}}{{.Foo}}` counts two and `{{ .Foo .Foo }}` counts one. A chip's
accessible name changes with that count: unused, it offers to insert at the
cursor; used, it offers to go to the next occurrence.

### The pagination arrows are not pagination

Detail pages for templates, knowledge and flows share one three-button cluster:
**Previous**, **Next**, and between them a button whose accessible name carries
the position — `Open templates list (3/11)` — opening a sheet of siblings. Four
things about it will make a case lie.

They walk the filtered sibling list, not pages: there is no `?page=` on these
screens, and a `?q=` left in the address bar from an earlier case shrinks the set
until both arrows go disabled, which reads as "navigation is broken". They
navigate with `replace: true`, so Back does not undo a step — it leaves the
detail page entirely. Below 768px the cluster is not in the header at all but
inside the **…** actions menu. And on a dirty form the unsaved-changes blocker
catches the step, so the arrow appears to do nothing.

### The indicator that stays silent while the socket is down

A pill at the bottom of the screen, `role="status"` with
`data-slot="connection-status"`, reads **Not receiving updates — reconnecting**
(that is an em dash). It is mounted above the router, so it can appear over any
page, and it is absent from the DOM the rest of the time.

It is not an online/offline badge. Apollo raises the event behind it only when at
least one GraphQL subscription is live at the moment the socket closes, so
killing the network on a page that subscribes to nothing shows nothing — and its
absence is not evidence that the connection is up. In the other direction,
leaving a subscribing page while still disconnected clears the pill, because
tearing the subscription down reports the socket as recovered. Assert on it only
on a page you know holds a live subscription, and treat the flows list, which
does, as the reliable place to see it.

## The flow page in depth

This is where most of the surface lives, and where a shallow pass is worst.

**Header** — favorite toggle (`aria-pressed` flips and the list agrees), the
**Flow actions** menu (rename, finish, delete), the **Report** menu, and the
previous/next navigation between flows.

**Both tab strips** — `Automation`, `Assistant`, `Dashboard` on the left at wide
viewports; `Terminal`, `Tasks`, `Agents`, `Searches`, `Vector Store`, `Files`,
`Screenshots` on the right. Below 1280 px they collapse into one strip of ten;
check both layouts. The two strips are two panes side by side, both mounted at
once, so `[role=tabpanel]:visible` resolves to the left one whatever you clicked
on the right: take the tab's `aria-controls` and read that element by id. An
empty state is a pass only when the database agrees the collection is empty, so
confirm the zero with the API.

Two tabs answer to an address that does not match their label: **Searches** is
`tools` and **Vector Store** is `vectorStores`. The rest follow their
label in lower case — `automation`, `assistant`, `dashboard`, `terminal`, `tasks`,
`agents`, `files`, `screenshots`.

**Which parameter carries them depends on the width.** At 1280 px and above the
right-hand strip has its own: `?side=terminal`, `?side=tasks`, `?side=agents`,
`?side=tools`, `?side=vectorStores`, `?side=files`, `?side=screenshots`, while
`?tab=` keeps the three central ones. Below 1280 the strips merge and everything
goes back to `?tab=`. A deep link copied from a wide window therefore opens a
different tab on a narrow one.

Open a tab by address when you need it open
before the page settles: clicking the trigger through a tool that synthesises
events does not move a Radix strip.

**Composer states** — with the flow running, the composer must show **Stop**
(which stops generation) and the textarea carries `disabled`, not `readOnly`, with
the placeholder "PentAGI is working... Click Stop to interrupt"; with the flow
finished or failed the form stays disabled and the placeholder becomes "This flow
has ended. Create a new one to continue." Both strings end in three dots or a full
sentence exactly as quoted — a check written against a typographic ellipsis matches
nothing. Assert `el.disabled` — `el.readOnly` is false in every state, and a check
written against it reports a working composer as broken. Send a message into a
running flow and watch it appear in the log without a reload.

The assistant composer is a separate form with its own provider picker, and its
**Submit** stays disabled until a provider is chosen there — the automation
composer arrives with one preselected, so a tester who only knows that side reads
the disabled button as a bug.

It also keys off a different status. The automation composer follows the flow; the
assistant form disables on the flow **or** the selected assistant being Finished or
Failed, and shows its working state on the assistant being Created or Running. Its
placeholder names which one it is reading ("Assistant is running... Click Stop to
interrupt", "This assistant session has ended. Create a new one to continue.").
Finishing the flow therefore exercises one branch only — seed a terminal assistant
under a live flow for the other.

**Terminal** — a WebGL canvas. Screenshots and DOM text tell you nothing; read
the xterm buffer through its handle and assert on the text you find there.

**Files** — the flow's own file manager, driven over REST rather than GraphQL:
upload, pull from the container, promote to resources, download, delete, plus
the bulk selection. This is the surface where path handling bugs live; pair the
UI walk with raw requests that send an absolute path or `..`.

The panel lists the **server's** data directory, not the sandbox: the resolver
calls `flowfiles.List(DataDir, flowID)`, which reads
`<DATA_DIR>/flow-<id>-data/{uploads,container,resources}` inside the backend
container. Writing into the sandbox (`docker exec pentagi-terminal-<id> …`) or
into a `flow-<id>` directory leaves the panel empty and looks like a broken
query. To seed one file for a read path, without going through an upload:

```bash
docker exec pentagi sh -c 'mkdir -p /opt/pentagi/data/flow-<id>-data/uploads &&
  echo body > /opt/pentagi/data/flow-<id>-data/uploads/probe.txt'
```

Each listed file carries the flow it belongs to, and it has to: the id is an md5
of the path relative to the flow, so `{ flowFiles { id flowId } }` returning
`flowId: 0` is the cache-collision regression, not a cosmetic gap.

### Which runs to start

A pass needs both modes and the seams between them, and one run of each is not
enough — the interesting failures live in the mixes. Six starts cover it, and
they can overlap:

1. **Automation only** — the ordinary path: composer on `/flows/new`, Automation
   tab, a provider already selected.
2. **A second automation, started while the first runs** — this is what makes
   cross-flow leakage visible at all.
3. **Assistant only** — note that this start does not call `createFlow`; the
   composer calls `createAssistant` and the flow appears as a side effect.
4. **An assistant inside a running automation flow** — switch to the Assistant tab
   on the flow you just started and send a message there.
5. **An automation task inside an assistant flow** — the mirror of 4, and the one
   that exercises `putUserInput` on a flow that has no task yet.
6. **A second assistant session in a flow that already has one** — through
   **Select assistant** → the new-session entry. Then read `assistantLogs` for
   both ids: each session must hold only its own turns.

Keep every prompt inside the sandbox container — this is a pentest product, and a
prompt that names an outside host starts a real scan of it. "Print the kernel
version and the current user", "list the top-level directories with their sizes",
"write a small script to /tmp and run it" are enough to produce a full run.

Watch which message types the run actually emits. A sandbox-only automation
produces `input`, `thoughts`, `terminal`, `file`, `report` and `done` — and never
`search`, `browser` or `advice`. Those three appear only when the prompt sends
the agent to the web, so a pass built from local prompts leaves their rendering
untested, and the cassettes do not cover them either.

## Subscriptions and concurrency

Live data is pushed, and a subscription that silently died looks exactly like a
quiet system. Test it deliberately:

1. **Mutate outside the tab.** With the flow page open, make the change through
   `curl` — rename the flow, add a resource, finish it. The open page must
   update on its own. Nothing changing is the bug.
2. **Two tabs on one flow.** Act in one, watch the other. Both must converge
   without a reload.
3. **Run several flows at once.** Start at least two automation flows and one
   assistant, each with a different prompt, and leave them running side by side.
   Then check that:
    - each flow's messages, tasks and terminal output land only in that flow's
      panels — cross-flow leakage is the failure this catches, and it is invisible
      with a single flow;
    - the flows list shows every one of them in a live status, updating as they
      move Created → Running → Finished;
    - the message types your prompts actually produce all render. A sandbox-only
      run gives `input`, `thoughts`, `terminal`, `file`, `report` and `done`; a
      prompt that sends the agent to the web adds `search` and `browser`. The
      cassettes carry `answer`, `input`, `report` and `terminal` and a local-tier
      run adds `thoughts` and `done` — so a sandbox pass exercises `file` here,
      and `search` and `browser` only if you asked for the web.
4. **Reload mid-run.** The page must rebuild from the query and keep receiving
   updates; a flow that stops updating after F5 has a resubscription bug.
5. **Count the frames, not just the pixels.** Wrap `window.WebSocket` in an
   `initScript` and record every `payload.data` — the key, the row id and its
   `flowId`. Isolation then reads as a number instead of an impression: while
   another flow ran an assistant, the open flow held 305 messages before and
   after, its panel stayed at 42600 characters, and zero frames carried the other
   flow's id, while the rename of the open flow arrived as exactly one
   `flowUpdated`. A panel that merely "looks unchanged" cannot tell you a delta
   was dropped rather than correctly filtered.
6. **Kill the backend mid-generation.** `docker restart pentagi` while an
   assistant streams is the cheapest test of two separate things: the socket must
   close `1006`, reconnect and fire `ws:reconnected`; and the reply the user was
   already reading must survive in `assistantlogs`. A row left at `length 0` means
   the running message never reached storage.
7. **Verify the payload, not the pixels.** For any panel that looks wrong, ask
   the API for the same data (`flow`, `messages`, `tasks`, `usageStatsBy…`) and
   compare. That single query separates a rendering bug from a data bug.

## What the page opens, and what hides on it

Two things a tester cannot infer from the screen: which subscriptions a route
holds open, and which controls are in the tree but invisible until something
else happens. Both are counted from the source; re-count them when the UI
changes.

**Which subscriptions a page holds** — every one of them must be shown to
deliver, and the cheap proof is the same for all: cause the event from `curl`,
watch the open page react without a reload. They are opened by nested providers,
not by the route, so a page holds its own plus everything above it — which is why
the same event has to keep arriving on pages that look unrelated to it.

| held by                                                 | subscriptions                                                                                                                                                                                                                                                                                |
| ------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| the app shell — every authenticated route **except** `/flows/:flowId/report`, which is protected but mounts its own layout | `settingsUserUpdated`, `flowTemplateCreated/Updated/Deleted`, `resourceAdded/Updated/Deleted`, `providerCreated/Updated/Deleted`                                                                                                                                                             |
| the flows layout — `/flows`, `/flows/new`, `/flows/:id` | `flowCreated`, `flowDeleted`, `flowUpdated`                                                                                                                                                                                                                                                  |
| `/flows/:id` itself                                     | `flowUpdated`, `taskCreated`, `taskUpdated`, `messageLogAdded`, `messageLogUpdated`, `agentLogAdded`, `terminalLogAdded`, `searchLogAdded`, `vectorStoreLogAdded`, `screenshotAdded`, `assistantCreated`, `assistantUpdated`, `assistantDeleted`, `assistantLogAdded`, `assistantLogUpdated` |
| the flow's **Files** panel, while it is mounted         | `flowFileAdded`, `flowFileUpdated`, `flowFileDeleted`                                                                                                                                                                                                                                        |
| `/knowledges*`                                          | `knowledgeDocumentCreated/Updated/Deleted`                                                                                                                                                                                                                                                   |
| `/settings/api-tokens`                                  | `apiTokenCreated/Updated/Deleted`                                                                                                                                                                                                                                                            |

So the flow page holds twenty-eight, thirty-one with the Files panel open, and
`flowUpdated` is genuinely open twice — the layout's copy refreshes the list behind
you, the page's copy refreshes the header. Ten of them ride on every other route
too: a resource added by `curl` must land on the settings page and on a flow page
alike, and a pass that only watches messages arrive has tested one subscription of
twenty-eight. The three `flowFile*` are the trap in the other direction — leave the
Files panel closed and they were never open to fail.

**Controls that hide.** Roughly three quarters of the app's controls are
conditional; these are the categories, with the ones most often missed:

- **Hover-only** — the sidebar's per-group actions (**New flow**, **New
  template**, **New knowledge**, **Upload file**) and every row action cell
  (**Toggle favorite**, **Open menu**) are `opacity-0` until their row or group
  is hovered. They are in the accessibility tree the whole time: click them by
  name, do not chase the hover.
- **Menu-only** — **Rename**, **Finish**, **Delete** on a flow live inside
  **Open menu** (list) and **Flow actions** (detail); the report's four exports
  live inside **Report**; the composer's attachments live inside **Templates and
  resources**. Open the trigger, snapshot again, then act — uids from the
  previous snapshot are stale.
- **Right-click twins** — every list repeats its whole row menu as a context menu
  (flows, templates, knowledges, providers, prompts, API tokens), and the file
  manager adds one on its rows and one on its empty area. Same writes, different
  path: test both.
- **State-gated** — **Finish** exists only while the flow is neither Failed nor
  Finished; the composer shows **Stop** instead of **Submit** exactly while it
  is Created or Running; the assistant row's delete hides once the flow or the
  selected assistant is terminal; **Report** appears only after the flow has at
  least one task. Seed
  the state or you have not tested the control.
- **Empty-state-only** — the **New Flow** / **New Knowledge** buttons inside an
  empty list, and **Try again**, which renders only when a query errored _and_
  the list is empty. A failed refetch with data on screen shows nothing.
- **Breakpoint-gated** — the flow header's favorite toggle and the
  previous/next navigation exist only at ≥768 px; on mobile the same writes move
  into the **Flow actions** menu. Below 1280 px the two tab strips merge into
  one strip of ten.
- **Account-type-gated** — only the password card on the account page is local-
  account-only; the email card renders for an OAuth account too, saying which
  provider it is linked from and dropping just its **Change** button. Asserting
  the card's absence therefore fails; assert the missing control.

**Table state lives in three places** — the URL (`?q=`, `?page=`), localStorage
(sorting, column visibility, page size, per-column search under a `table_*` key)
and the query itself. Check that a reload restores all three, and that clearing
the search does not quietly drop a filter you set elsewhere.

## Three widths, not two, and the routes that catch a miss

The app has three layouts, and the middle one is the reason "tested on mobile and
desktop" misses things. `useBreakpoint` calls under 768 px **mobile**, under
1280 px **tablet**, and the rest **desktop**. Between 768 and 1279 the sheet
navigation is gone but the split-pane layouts are still off — a case written for
"mobile" never visits it.

**Under 768** the sidebar is a drawer: a **Toggle Sidebar** button opens a dialog
titled **Sidebar** with the links **Dashboard**, **Flows**, **Templates**,
**Resources**, **Knowledges**, **Settings**. Its own close X is hidden by CSS, so
a role sweep finds no close control — Escape and the overlay are the way out.
Tapping any link except **Settings** leaves the drawer open on top of the new
page, and nothing resets it when the window widens: widen past 768, come back,
and the drawer is open without a click.

Also under 768, header actions keep their icons and drop their text. A check by
visible text reports **Save** as missing while a check by accessible name finds
it — the label is `display: none`, not absent. Detail-page prev/next leaves the
header and reappears inside the **…** menu as a row that opens a list sheet.

**At 768 and above** the sidebar is an icon rail. Its state is written to the
cookie `sidebar:state` for seven days and read only at mount, so a collapse in
one case silently changes the layout of every later one — and of the next
session on that browser. Ctrl/Cmd+B toggles it from anywhere with no visible
control; the shortcut is skipped while focus is in an input, a textarea or the
editor, which is why it does not steal Bold. On the main routes **two** controls
answer to the name **Toggle Sidebar** — the header trigger and an invisible rail
strip on the sidebar's edge — so a strict by-name query throws there and passes
on `/settings/*`, where the rail does not exist.

**The flow tabs change shape with the window.** Below 1280 the three central
tabs — **Automation**, **Assistant**, **Dashboard** — join the single strip and
the active one is written to `?tab=`. At 1280 and above they are a separate
pane, and a deep link like `?tab=terminal` lands on a different tab than it does
narrow. Test a deep link at the width you claim to have tested.

**A wrong address never shows a 404.** The root catch-all redirects anything
unknown to `/dashboard`, and `/settings/*` has its own; a typo therefore looks
like a working page, and a test that asserts "the app rejects a bad URL" cannot
pass. What does exist is a route error boundary with a **Reload** card, and
per-page not-found branches that differ from each other: a template shows a
**Template not found** card, a prompt shows a card that is a dead end, a
knowledge document shows a toast and bounces back to the list, and a provider
redirects with nothing said at all. Seed a bad id per entity — the four
behaviours are four different tests.

## The rest of a full pass

- **Search and filters** — case-insensitive across mixed case, the empty state,
  clearing restores the list, and the URL carries the query.
- **Pagination and sorting** — page size, "Page 1 of N", first/previous disabled
  on page 1, `aria-sort` correct before and after clicking a header.
- **Theme** — toggles and survives a reload.
- **Narrow viewport** — the tab strips collapse, nothing overflows horizontally
  on any route.
- **Console** — zero errors, everywhere, all pass.
- **The API matrix** for whatever validation the branch touched.

Budget: about an hour for the walk, plus whatever the parallel flow runs take to
finish on their own.

## The API matrix

Validation and authorization are **never** proven through the UI. A disabled
button is UX; the boundary is the endpoint, and an attacker sends the request
directly. Every rule gets the same five cells:

| cell                 | what it proves                                           |
| -------------------- | -------------------------------------------------------- |
| valid                | the value is accepted **and stored**                     |
| invalid              | blank / malformed / over-limit / unauthorized is refused |
| exactly at the limit | the boundary is where it claims to be                    |
| limit + 1            | it is not off by one                                     |
| no side effect       | after a reject, nothing was written                      |

A session and a bearer token, side by side, is the whole harness:

```bash
curl -sk -c jar.txt -X POST https://localhost:8443/api/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"mail":"…","password":"…"}'

curl -sk -b jar.txt -X POST https://localhost:8443/api/v1/graphql \
  -H 'Content-Type: application/json' -d @query.json     # session

curl -sk -H "Authorization: Bearer $TOKEN" \
  "https://localhost:8443/api/v1/flows/?page=1&pageSize=5&type=init"   # token tier
```

`/users`, `/roles` and `/tokens` are session-only and answer `403` to any bearer
token by design — one that reached them could mint an administrator and escape its
own scope. The split is per-route, not per-prefix: `GET /user/` is on the bearer
tier, while `PUT /user/password`, `/user/email` and `/user/name` are session-only. Mint the token through the session, then probe the token tier on a route
it owns (`/flows`, `/settings`, `/graphql`, the log tables): a `403` from `/users`
is the boundary working, not a bad token.

`page` and `type` are both required on every list endpoint that binds the shared
table query, and the binder answers before the handler ever runs —
`…/knowledge/?page=1&pageSize=100` is `400 Knowledge.InvalidRequest`, the same URL
with `&type=init` is the list. `pageSize` is the one that is optional: leave it
out and the page comes back at its documented default of five, and `-1` asks for
the whole table. Match on the `code` in the body rather than the status, or a
binder 400 scores the "invalid is refused" cell for a request that never reached
the rule under test.

The same requests isolate a broken panel: ask the API for the data the panel
renders. If the payload is there and the card is empty, the bug is in the UI; if
the payload is empty, stop looking at React. That one query is usually shorter
than the argument about where the bug lives.

Two habits keep the results honest:

- **Read the artifact, not the tail of a pipe.** `cmd | tail && echo $?` reports
  `tail`'s exit code. Redirect to a file and check `$?`, or verify the result
  itself (the row exists, the image id changed).
- **A refusal is only half the check.** Follow every rejected request with a read
  that proves nothing was written — the list does not contain the row, the file
  is still a file, the blob was not orphaned.

## Proving a fix

The standard is a measured difference, not a passing check:

1. Run the exact input against the **previous** image and record the status and
   body. Keep the old tag around for this — `local/pentagi:<prev>`.
2. Rebuild, restart, re-verify the container id, run the same input.
3. Report both. "409 with the file intact, where the previous build answered 200
   and turned it into a directory" is evidence; "works now" is not.

The same discipline applies to a test: revert the fix, watch the new test fail
with the symptom that started it, restore. A test that passes either way is not
coverage.

Two ways that mutation lies, both seen here:

- **A deadline derived from the constant under test stretches instead of
  failing.** A test that waits `interval * 4` for a write to land, mutated by
  setting the interval to an hour, waits four hours — the run dies on the Go test
  timeout with no assertion message, and it reads like a hang rather than the
  failure you were looking for. Bound the wait in wall-clock seconds.
- **A mutation that breaks the build is not a mutation.** Deleting the call site
  leaves the helper unused, Go refuses to compile, and a red package tells you
  nothing about the assertion. Change the value, not the shape: an interval, a
  flag, one entry in a table.

Where the stack cannot produce the failure at all — an LLM key you do not have,
a process kill no unit test can stage — the fake of the interface under test is
the honest instrument. `flowMsgLogWorker` takes a `Querier` and a `FlowPublisher`,
so two structs embedding those interfaces answer "does any path publish an empty
result over a set one" in seconds, with no database and no credentials.

**A failed write is one of those.** Stopping the database mid-generation does not
stage it: the mutation that would start the generation fails first, answering
`a required service is unavailable`, so nothing is streaming when the writes
begin to fail. Timing an outage into an already-running generation is worse than
unreliable — the whole reply lands inside ten to twenty seconds, and the write
that must fail is one call inside that window. Three attempts here produced no
frame carrying the defect. The fake fails the exact call by ordinal
(`failContentOn: 1`) and reproduces it every run, which is why the assistant
worker's failure paths are pinned in `aslog_persist_test.go` and not on a stand.

What the stand *is* good for on that path is the healthy half: watching the
periodic write land. Poll the stored row while a long answer streams and it grows
in steps a few seconds apart — 1659, 2634, 3274, 4134 characters on one run —
which is the ticker doing its job, visible nowhere else.

## Leaving the stand clean

Every probe is a row someone else will trip over. Delete the tokens, providers,
templates, knowledge documents, resources and uploaded files you created, and
restart the stack without the environment overrides you added
(`COOKIE_SIGNING_SALT`, `OLLAMA_SERVER_URL`) so the next session inherits the
stand's own configuration. Then list what remains and check it against the seeds.

A shared stand is stricter still: it has no isolation, several people use the
same accounts, and anything you change is visible to everyone. Roll back what
you touch, and never put its URL or credentials in the repository — they belong
with the stand-tier secrets described in [`e2e.md`](./e2e.md#stand-tier-tier-3).
