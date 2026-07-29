# List & Detail Page Building Blocks

Reusable surface for **list-and-detail** pages: a filterable/sortable table, a
Prev/Next/Sheet toolbar that walks the _same_ filtered subset on detail pages,
inline-rename inputs, and the URL-vs-storage state model that keeps them in
lockstep. Every list page in the app (`/flows`, `/knowledges`, `/templates`,
`/settings/prompts`, `/settings/providers`, `/settings/api-tokens`) renders a
`<DataTable>`. The detail half applies to `/flows/:id`, `/knowledges/:id` and
`/templates/:id` only. `/settings/api-tokens` has no detail route at all (it edits
rows in place, through its own state rather than the inline-edit surface below),
and `/settings/prompts/:name` and `/settings/providers/:id` are single-entity
forms with no sibling navigation, so neither has a feature nav hook.

> **Scope / how to trust this doc.** It describes code under `frontend/src`.
> Signatures below are copied verbatim from source. If a signature and the code
> disagree, the code wins — re-grep the named file and update this doc. Symbols
> flagged **removed / does not exist** are called out explicitly (see
> [Removed / renamed](#removed--renamed--do-not-use)) so nothing resurrects them.

## Table of Contents

- [Overview](#overview)
- [Mental model](#mental-model)
- [Where the code lives](#where-the-code-lives)
- [Core concepts](#core-concepts)
- [Hooks](#hooks)
- [Components](#components)
- [Library helpers (`@/lib`)](#library-helpers-lib)
- [Recipe: add a new list + detail pair](#recipe-add-a-new-list--detail-pair)
- [Design rationale](#design-rationale)
- [Gotchas](#gotchas)
- [Testing](#testing)
- [Removed / renamed — do not use](#removed--renamed--do-not-use)

## Overview

The surface splits cleanly into two responsibilities:

- **List page** — renders a `<DataTable>` driven by `useTableState()`. The
  free-text filter (`?q=`) and page number (`?page=`) live in the **URL**; view
  preferences (sorting, column visibility, page size, search columns) persist in
  **localStorage** under a single `table_4_<path>` slot.
- **Detail page** — renders a `<DetailNavigationToolbar>` driven by a
  feature-scoped hook built on `useDetailNavigation()`. It **reads** the list's
  `?q=` filter (never writes it) and offers Prev / Next / a searchable listbox
  over exactly the subset the user filtered on the list.

Two rules make the whole thing predictable:

1. **URL is authoritative** for filter + page, so a shared link reproduces the
   exact view.
2. **Prev/Next walks the same subset** — the detail toolbar filters on the same
   `?q=` the table does, so siblings stay in step with the list. The two
   matchers are close but not identical; see the caveat below.

## Mental model

```
                ┌──────────────────────────────────────────────┐
                │  URL   ?q=foo   ?page=3   ?qs=bar             │
                │  (source of truth — bookmarkable, shareable)  │
                └────────┬──────────────────────────┬──────────┘
                         │ read + write             │ read-only
                         ▼                          ▼
                   useTableState            useTableQueryFilterReader
                   (list pages)                     │
                         │                          ▼
                         ▼                   useDetailNavigation
                    <DataTable>                     │
                         │                          ▼
                         │              <DetailNavigationToolbar>
                         ▼                   (detail pages)
                 table_4_<path>
                 in localStorage
              (sorting / columns / pageSize;
               NOT filter, NOT page index)
```

- **URL** carries the ad-hoc query (`?q=`), the page index (`?page=`, 1-based),
  and — on knowledge/semantic screens — a server-side search (`?qs=`).
- **Storage** carries only durable view preferences. A detail page **never**
  writes storage and never replays storage into the URL, so opening a shared
  `/flows/:id?q=foo` link shows exactly what the link says.

## Where the code lives

```
frontend/src/
├── hooks/
│   ├── use-table-state.ts            # list-page URL state (filter + page)  ← writer
│   ├── use-table-query-filter.ts     # detail-page read-only filter subscription
│   └── use-page-storage-keys.ts      # resolve per-route localStorage keys
├── lib/
│   ├── table-state.ts                # table_4_<path> slot: load / update / migrate
│   ├── view-options-storage.ts       # viewOptions_4_<path> slot (FileManager screens)
│   ├── storage-keys.ts               # key builders + getTopLevelPath
│   └── url-params.ts                 # URL_PARAMS + mergeHrefWithSearchParams
├── components/
│   ├── ui/
│   │   └── data-table.tsx            # <DataTable>, <DataTableColumnHeader>, cycleColumnSort
│   └── shared/
│       ├── detail-navigation/        # prev/next toolbar + headless controller
│       │   ├── index.ts              #   public barrel
│       │   ├── use-detail-navigation.ts     #   useDetailNavigation + DetailNavigationController
│       │   ├── use-navigation.ts             #   pure core (internal, unit-tested headless)
│       │   ├── text-filter.ts                #   text matcher (internal)
│       │   ├── detail-navigation-toolbar.tsx #   <DetailNavigationToolbar>
│       │   ├── detail-navigation-buttons.tsx #   <DetailNavigationButtons> (leaf)
│       │   └── detail-navigation-sheet.tsx   #   <DetailNavigationSheet> (leaf)
│       ├── inline-edit/              # inline rename input + state machine
│       │   ├── index.ts
│       │   ├── inline-edit-input.tsx #   <InlineEditInput>
│       │   └── use-inline-edit.ts     #   useInlineEdit
│       ├── error-state.tsx           # <ErrorState> (region-level error + Retry)
│       └── loading-state.tsx         # <LoadingState>
└── features/<entity>/
    └── use-<entity>-detail-navigation.ts  # feature-scoped nav hook (one per entity)
```

| Area            | File(s)                                             | Public surface                                                                                                                          |
| --------------- | --------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------- |
| List URL state  | `hooks/use-table-state.ts`                          | `useTableState`                                                                                                                         |
| Detail URL read | `hooks/use-table-query-filter.ts`                   | `useTableQueryFilterReader`                                                                                                             |
| Storage keys    | `hooks/use-page-storage-keys.ts`                    | `usePageStorageKeys`                                                                                                                    |
| Table           | `components/ui/data-table.tsx`                      | `DataTable`, `DataTableColumnHeader`, `cycleColumnSort`                                                                                 |
| Detail nav      | `components/shared/detail-navigation/`              | `DetailNavigationToolbar`, `DetailNavigationButtons`, `DetailNavigationSheet`, `useDetailNavigation`, type `DetailNavigationController` |
| Inline edit     | `components/shared/inline-edit/`                    | `InlineEditInput`, `useInlineEdit`                                                                                                      |
| Storage slots   | `lib/table-state.ts`, `lib/view-options-storage.ts` | `loadTableState`/`updateTableState`/`migrateLegacyTableState`, `loadViewOptions`/`saveViewOptions`/`migrateLegacyViewOptions`           |
| Key + URL utils | `lib/storage-keys.ts`, `lib/url-params.ts`          | `getTableStorageKey`/`getTopLevelPath`, `URL_PARAMS`/`mergeHrefWithSearchParams`                                                        |

## Core concepts

### URL is authoritative

`useTableState` reads and writes three-ish query params through
`react-router`'s `useSearchParams`:

| Param    | Meaning                                                         | Owner                                                                    |
| -------- | --------------------------------------------------------------- | ------------------------------------------------------------------------ |
| `?q=`    | client-side free-text filter                                    | `useTableState` (list) / `useTableQueryFilterReader` (detail, read-only) |
| `?page=` | **1-based** page number; absent ≡ page 1                        | `useTableState`                                                          |
| `?qs=`   | server-side semantic search (vector store); orthogonal to `?q=` | page-local `useSearchParams` (e.g. knowledges)                           |

`pageIndex` is exposed **0-based** in JS; the URL param is **1-based**.
`useTableState` converts on read (`parsed - 1`) and write (`index + 1`) and
canonicalizes `?page=1` away so `/flows` and `/flows?page=1` don't split the
history stack.

### Storage is a warm-restart bag

Each list page persists **at most one** localStorage key — `table_4_<path>` — a
single JSON object bundling `filter`, `sorting`, `columnVisibility`, `pageSize`,
and `searchColumns`. `DataTable` owns these writes via `usePageStorageKeys().table`.
The current page index and the ad-hoc `?q=` filter stay URL-only.

### Prev/Next walks the same subset

`useDetailNavigation` subscribes to the URL filter via
`useTableQueryFilterReader`, applies a pure matcher over the same `?q=`, and
resolves the Prev / Next siblings + a `"3/10"` position label over that filtered
subset. It also threads the current `?q=` through every prev/next/select
destination via `mergeHrefWithSearchParams`, so navigating siblings never drops
the filter.

The two matchers are not the same function, and they disagree on diacritics:
the toolbar folds them (`createTextMatcher` normalizes NFKD and strips combining
marks), while `DataTable`'s `globalFilterFn` only lowercases. Searching `cafe`
therefore hides a `café` row from the table while Prev/Next still steps onto it.
Align the two before relying on the subsets being identical.

## Hooks

| Hook                        | File                                   | Source of truth      | Writes URL?         | Used by                                            |
| --------------------------- | -------------------------------------- | -------------------- | ------------------- | -------------------------------------------------- |
| `useTableState`             | `hooks/use-table-state.ts`             | URL `?q=` + `?page=` | **yes**             | List pages                                         |
| `useTableQueryFilterReader` | `hooks/use-table-query-filter.ts`      | URL `?q=`            | no                  | Detail pages (indirect, via `useDetailNavigation`) |
| `usePageStorageKeys`        | `hooks/use-page-storage-keys.ts`       | router               | no                  | `DataTable`, dashboards                            |
| `useDetailNavigation`       | `components/shared/detail-navigation/` | URL + props          | no (navigates only) | Feature-scoped nav hooks                           |
| `useInlineEdit`             | `components/shared/inline-edit/`       | local state          | no                  | List cells, detail breadcrumbs                     |

### `useTableState` — list-page URL state

```ts
function useTableState(options?: {
    clearPageOnFilterChange?: boolean; // default true — setFilter also drops ?page=
    debounceMs?: number; // default 200
    filterParamName?: string; // default 'q'  (URL_PARAMS.QUERY)
    pageParamName?: string; // default 'page' (URL_PARAMS.PAGE)
}): {
    debouncedFilter: string;
    filter: string;
    pageIndex: number; // 0-based
    resetFilter: () => void;
    setFilter: (value: string) => void;
    setPage: (pageIndex: number, options?: { replace?: boolean }) => void;
    update: (patch: { filter?: null | string; pageIndex?: number; replace?: boolean }) => void;
};
```

- **`setFilter`** always navigates with `replace: true` (keystrokes don't clutter
  history) and, when `clearPageOnFilterChange` is `true`, bundles `pageIndex: 0`
  into the **same atomic update** so narrowing the results also resets the page
  (no "page 5 of nothing").
- **`setPage`** pushes a history entry by default; pass `{ replace: true }` for
  out-of-range clamping so a bad URL isn't one back-press away.
- **`update`** is the atomic multi-field primitive. Every call in the same
  microtask is coalesced into **one** `setSearchParams`, which is what makes a
  simultaneous filter+page change race-free by construction. Reach for it when
  changing both fields from one event.

### `useTableQueryFilterReader` — detail-page read-only filter

```ts
function useTableQueryFilterReader(options?: {
    debounceMs?: number; // default 200
    paramName?: string; // default 'q'
}): { debouncedFilter: string; filter: string };
```

Observes `?q=` but **never** writes the URL. This is the piece that lets a fresh
detail-page mount avoid injecting a previous tab's `?q=` into the URL. It is the
only export of `use-table-query-filter.ts` — there is no writer counterpart (see
[Removed / renamed](#removed--renamed--do-not-use)). Most code doesn't call it
directly; `useDetailNavigation` uses it internally.

### `usePageStorageKeys` — per-route storage keys

```ts
function usePageStorageKeys(options?: {
    pathname?: string; // override the path used to build keys
    useTopLevel?: boolean; // default false; /flows/abc-123 → /flows when true
}): { period: string; table: string; viewOptions: string };
```

Resolves the three per-route localStorage keys reactively via `useLocation()`
(not a non-reactive global `location.pathname` read). A detail page that wants
to share its parent list's storage bucket passes `{ useTopLevel: true }`.

### `useDetailNavigation` — headless detail-nav controller

```ts
function useDetailNavigation<T extends { id: string }>(options: {
    currentId: null | string | undefined; // pass null while creating/new
    getHref: (item: T) => string;
    getLabel: (item: T) => string;
    items: readonly T[];
    getId?: (item: T) => string; // default item.id
    getSearchableText?: (item: T) => null | string | undefined; // default getLabel
    sortFn?: (a: T, b: T) => number;
    // controlled sheet (optional):
    open?: boolean;
    defaultOpen?: boolean;
    onOpenChange?: (open: boolean) => void;
    // controlled in-sheet local search (optional):
    searchQuery?: string;
    defaultSearchQuery?: string;
    onSearchQueryChange?: (query: string) => void;
    searchDebounceMs?: number; // default 150
}): DetailNavigationController<T>;
```

Returns a large, identity-stable `DetailNavigationController<T>` (exported type):
`currentItem`, `currentIndex`, `prevId`, `nextId`, `positionLabel` (`"3/10"` or
`"–/0"`), `goToPrev` / `goToNext` / `handleItemSelect`, `filteredItems`, `total`,
sheet state (`isSheetOpen` / `openSheet` / `closeSheet` / `setSheetOpen`), local
search state (`searchQuery` / `setSearchQuery` / `clearSearchQuery` /
`debouncedSearchQuery` / `isSearchActive`), and the guards `hasEntries` /
`itemsEmpty`. The filtered subset is narrowed by the URL `?q=` **AND** the
in-sheet local search, combined with AND.

> Hoist `getLabel`/`getHref`/`getId`/`getSearchableText` to **module scope** (or
> wrap in `useCallback`). Inline arrows still work but forfeit the controller's
> internal memoization on `getSearchableText`.

### `useInlineEdit` — inline rename state machine

```ts
function useInlineEdit<TElement extends HTMLElement = HTMLInputElement>(options?: {
    resetKey?: null | string | undefined; // changing it closes any open editor (pass the entity id)
}): {
    handleDropdownCloseAutoFocus: (event: Event) => void; // spread onto Radix <DropdownMenuContent>
    inputRef: React.RefObject<null | TElement>;
    isEditing: boolean;
    startEdit: () => void;
    stopEdit: () => void;
};
```

Owns the edit-mode toggle, a ref for the input, and **deferred focus** on the
next animation frame so focus lands _after_ Radix's dropdown close-focus-restore
(otherwise Radix wins the race and the input never focuses). Pair with
`<InlineEditInput>` and pass `resetKey={entity.id}` so navigating items drops a
stale draft.

## Components

### `<DataTable>` — the table primitive

Generic TanStack-backed table (filter, sort, pagination, column visibility,
optional window virtualization). Controlled or uncontrolled per axis.

```ts
// exported: DataTable, DataTableColumnHeader, cycleColumnSort
function DataTable<TData, TValue = unknown>(props: {
    columns: ColumnDef<TData, TValue>[]; // required
    data: TData[]; // required
    empty?: { entityName?: string }; // plural lowercase, e.g. "flows"
    filterColumn?: string | string[]; // omit → zero-config: columns w/ meta.searchable === true
    filterPlaceholder?: string; // default "Filter..."
    filterValue?: string; // controlled filter (pair with onFilterChange)
    onFilterChange?: (value: string) => void;
    pageIndex?: number; // controlled page (0-based)
    onPageChange?: (pageIndex: number, options?: { replace?: boolean }) => void;
    initialPageSize?: number; // default 10
    initialSorting?: SortingState; // default []
    columnVisibility?: VisibilityState;
    onColumnVisibilityChange?: (visibility: VisibilityState) => void;
    onRowClick?: (row: TData) => void;
    renderRowContextMenu?: (row: TData) => ReactNode;
    renderSubComponent?: (props: { row: Row<TData> }) => ReactElement;
    isVirtualized?: boolean; // default false; ignored when renderSubComponent set; activates > 50 rows
    storageKey?: string; // default `table_4_<pathname>` (usePageStorageKeys().table)
}): JSX.Element;

function DataTableColumnHeader<TData, TValue = unknown>(props: {
    column: Column<TData, TValue>; // TanStack column — sorting + cycle action target it
    title: ReactNode;
}): JSX.Element;
```

Key rules:

- **Controlled needs the pair.** The filter is controlled only when **both**
  `filterValue` and `onFilterChange` are set; passing `filterValue` alone
  silently reverts to internal uncontrolled state and your value is ignored. A
  controlled page needs `pageIndex` **and** `onPageChange` (the latter is what
  lets the table push `pageIndex → 0` into the URL when the filter narrows).
- **Search columns are declared per column** via `meta: { searchable: true }` on
  the `ColumnDef` (zero-config multi-column mode). If no column is `searchable`
  and no `filterColumn` is passed, the search input isn't rendered at all.
- **Empty state** text comes from `empty={{ entityName: 'flows' }}`.
- **Storage aliasing:** two `<DataTable>`s on the same route share the default
  `table_4_<path>` key and overwrite each other — give each a distinct
  `storageKey` (e.g. `` `${base}:agents` ``).

### `<DetailNavigationToolbar>` and leaves

```ts
function DetailNavigationToolbar<T extends { id: string }>(props: {
    controller: DetailNavigationController<T>; // required — from the feature nav hook
    sheetTitle: string; // required
    sheetIcon?: ReactNode;
    renderItem?: (item: T, isCurrent: boolean) => ReactNode;
    hasSearch?: boolean; // default true — in-sheet search input
    searchPlaceholder?: string;
}): JSX.Element;
```

The toolbar composes `<DetailNavigationButtons>` (Prev / position / Next) and
`<DetailNavigationSheet>` (the searchable listbox) against one controller. Most
desktop call sites use the toolbar directly; pages with non-standard chrome
(e.g. mobile prev/next inside a `<DropdownMenuItem>`) compose the two leaves
themselves and read from the same controller. All three accept `controller` +
their own presentation props — **there is no `toolbarProps` object to spread.**

### `<InlineEditInput>`

```ts
function InlineEditInput(props: {
    onSave: () => void; // required — read latest text from the bound inputRef
    onCancel: () => void; // required
    autoFocus?: boolean; // default false
    busy?: boolean; // default false — disables the Save + Cancel buttons while saving (the input stays editable)
    className?: string;
    defaultValue?: string; // uncontrolled input
    inputRef?: Ref<HTMLInputElement>; // pair with useInlineEdit().inputRef
    maxLength?: number; // default 200 (native maxLength guard, not a validation gate)
    placeholder?: string;
}): JSX.Element;
```

`Enter` saves, `Escape` cancels. The input is **uncontrolled** — callers read
the value from `inputRef.current` at save time, avoiding a re-render per
keystroke.

## Library helpers (`@/lib`)

### `lib/table-state.ts` — the unified table slot

```ts
type TableState = {
    columnVisibility?: Record<string, boolean>;
    filter?: string;
    pageSize?: number;
    searchColumns?: string[];
    sorting?: { desc: boolean; id: string }[];
};

const loadTableState: (key: string) => TableState; // {} when missing/invalid
const updateTableState: (key: string, patch: Partial<TableState>) => TableState; // undefined/empty clears a field; empty object removes the key
const migrateLegacyTableState: (path: string, unifiedKey: string) => TableState;
```

`migrateLegacyTableState` is a one-shot, idempotent reader of the pre-unification
layout: it folds the four legacy keys (`column_4_`, `sorting_4_`, `filter_4_`,
`page_4_`) into the unified `table_4_<path>` slot and deletes them. From the
`page_4_` slot it carries over **only** `pageSize` (the page index now lives in
`?page=`). Safe to call on every mount.

### `lib/view-options-storage.ts` — FileManager view options

```ts
type ViewOptionsRecord = Record<string, boolean>;
const loadViewOptions: (key: string) => ViewOptionsRecord;
const saveViewOptions: (key: string, value: ViewOptionsRecord) => void; // removes key on empty
const migrateLegacyViewOptions: (path: string, unifiedKey: string) => ViewOptionsRecord;
```

Separate `viewOptions_4_<path>` slot for FileManager-style screens (folders-first
toggle, expanded dirs) that aren't backed by `DataTable`.

### `lib/storage-keys.ts` — key conventions

```ts
const STORAGE_KEY_SEPARATOR = '_4_'; // reads as "for": table_4_/flows = "table for /flows"
type LocalStorageKeyType = 'period' | 'table' | 'viewOptions';
function getStorageKey(type: LocalStorageKeyType, urlPath: string): string; // `${type}_4_${urlPath}`
function getTableStorageKey(urlPath: string): string; // table_4_<path>
function getPeriodStorageKey(urlPath: string): string; // period_4_<path> (dashboard time window)
function getViewOptionsStorageKey(urlPath: string): string; // viewOptions_4_<path>
function getTopLevelPath(pathname: string): string; // /flows/abc → /flows ; / → ''
```

The key embeds the full `urlPath` **including** its leading slash — the literal
is `table_4_/flows`, not `table_4_flows`. `getTopLevelPath` only walks one level
deep, so a nested list like `/admin/flows/:id` would resolve to `/admin`; such a
detail page must hardcode its parent path instead.

### `lib/url-params.ts` — URL constants + href merge

```ts
const URL_PARAMS = { PAGE: 'page', QUERY: 'q', SEARCH: 'qs' } as const;
function mergeHrefWithSearchParams(base: string, incoming: Iterable<[string, string]> | URLSearchParams): string;
```

`mergeHrefWithSearchParams` adds every incoming key to `base` unless `base`
already specifies it (the href wins), and preserves any `#hash`. This is how the
detail toolbar forwards the current `?q=` into prev/next/select destinations.

## Recipe: add a new list + detail pair

For a new entity at `/<entities>` + `/<entities>/:id`:

### 1. List page — `pages/<entities>/<entities>.tsx`

```tsx
export function EntitiesPage() {
    const { entities, isLoading, error, refetch } = useEntities();

    // Paginated variant. For a page without ?page=, destructure only { filter, setFilter }.
    const { filter, pageIndex: currentPage, setFilter, setPage: handlePageChange } = useTableState();

    const columns: ColumnDef<Entity>[] = [
        {
            accessorKey: 'title',
            header: ({ column }) => (
                <DataTableColumnHeader
                    column={column}
                    title="Title"
                />
            ),
            meta: { searchable: true }, // opt this column into the global filter
        },
        // ...more columns
    ];

    const pageHeader = <AppHeader title="Entities" /* ...actions */ />;

    // Canonical 4-branch render gate — pageHeader renders in ALL branches.
    // Both loading and error gate on "nothing to show yet": the query is
    // cache-and-network, so a background revalidation reports loading (and, on
    // failure, error) with cached data still present. Without `&& !data` that
    // refetch blanks a working list — or an edit form with unsaved changes —
    // with the spinner/error for the round-trip.
    if (isLoading && entities.length === 0) {
        return (
            <>
                {pageHeader}
                <LoadingState title="Loading entities…" />
            </>
        );
    }
    if (error && entities.length === 0) {
        return (
            <>
                {pageHeader}
                <ErrorState
                    title="Couldn't load entities"
                    message={error.message}
                    onRetry={refetch}
                />
            </>
        );
    }
    if (entities.length === 0) {
        return (
            <>
                {pageHeader}
                <Empty>{/* EmptyHeader / EmptyTitle / EmptyContent */}</Empty>
            </>
        );
    }

    return (
        <>
            {pageHeader}
            <DataTable<Entity>
                columns={columns}
                data={entities}
                empty={{ entityName: 'entities' }}
                filterPlaceholder="Filter entities..."
                filterValue={filter}
                onFilterChange={setFilter}
                onPageChange={handlePageChange} // omit this + pageIndex on non-paginated pages
                pageIndex={currentPage}
                onRowClick={(entity) => navigate(routes.entity(entity.id))}
            />
        </>
    );
}
```

### 2. Feature nav hook — `features/<entity>/use-<entity>-detail-navigation.ts`

```ts
import { useDetailNavigation } from '@/components/shared/detail-navigation';
import { routes } from '@/lib/routes';
import { type Entity, useEntities } from '@/providers/entities-provider';

// Module scope → stable identity → controller memoization holds.
const getLabel = (item: Entity) => item.title || `Entity #${item.id}`;
const getSearchableText = (item: Entity) => item.title;
const getId = (item: Entity) => String(item.id);
const getHref = (item: Entity) => routes.entity(item.id);

export function useEntityDetailNavigation(currentId: null | string | undefined) {
    const { entities } = useEntities();

    return useDetailNavigation<Entity>({
        currentId,
        getHref,
        getId,
        getLabel,
        getSearchableText,
        items: entities,
    });
}
```

### 3. Detail page — `pages/<entities>/<entity>.tsx`

```tsx
const entityNav = useEntityDetailNavigation(isNew ? null : entityId); // null while creating

// Desktop: inside the header actions.
{
    !isMobile && (
        <DetailNavigationToolbar<Entity>
            controller={entityNav}
            renderItem={(item, isCurrent) => (
                <span className={cn('min-w-0 flex-1 truncate', isCurrent && 'font-medium')}>{item.title}</span>
            )}
            sheetIcon={<EntityIcon className="size-4" />}
            sheetTitle="Entities"
        />
    );
}

// Mobile: leaves composed by hand — <DetailNavigationButtons> inside a
// <DropdownMenuItem>, and a standalone <DetailNavigationSheet> outside the header,
// both reading `controller={entityNav}`.
```

#### Ordering the header actions

`<AppHeaderContent>` takes the remaining width (`flex-1`), so `<AppHeaderActions>` ends up
against the right edge and grows leftward. Two rules follow:

1. **Controls that appear on a data condition go first** in the children. Everything after
   them keeps its position when they arrive late — the flow header's Report button loads with
   the task list, and the pager beside it must not move under a cursor that is clicking Next.
2. **Controls that are always meaningful for the route render always**, taking `disabled` from
   an explicit loading flag rather than being unmounted while the entity is null. Unmounting
   collapses the cluster for the length of every fetch, which costs a step per pager click.

The flow header is the reference: `Report · favourite · pager · actions menu`. A page whose
entity can be genuinely absent (a not-found card) is the exception — it renders no actions at
all, because none of them are meaningful there.

## Design rationale

### Why URL > storage

A user opens `/flows?q=alpha` and clicks "Next" in the detail toolbar to reach
flow B, then shares `/flows/b?q=alpha` with a teammate. The teammate opens it
cold: the detail page reads `q=alpha` from the URL and renders Prev/Next over the
filtered subset; hitting "Next" lands on `/flows/c?q=alpha` — still inside the
filter they never typed. `useTableQueryFilterReader` is the key: it observes the
URL but never writes, so a fresh detail mount can't inject a previous tab's
`?q=` into the URL.

### Why one storage key per page

Before unification every list page wrote four keys (`column_4_`, `sorting_4_`,
`filter_4_`, `page_4_`) across two write paths (sync + debounced); refreshing
mid-typing could land inconsistent state. The unified `table_4_<path>` slot is
one JSON object all preferences live in; `migrateLegacyTableState` folds the four
legacy keys in on first mount and deletes them.

## Gotchas

- **`pageIndex` is 0-based; `?page=` is 1-based.** Never hand `<DataTable>` or
  `useTableState` a 1-based index — the hook does the ±1 conversion.
- **Controlled axes need the prop pair.** `filterValue` without `onFilterChange`
  (or `pageIndex` without `onPageChange`) silently reverts that axis to the
  table's internal uncontrolled state, and the URL stops updating.
- **No `toolbarProps`.** Detail pages pass `controller={nav}` plus discrete
  `renderItem` / `sheetIcon` / `sheetTitle` props. There is no bundle to spread.
- **`storageKey` aliasing.** Multiple `<DataTable>`s on one route share the
  default `table_4_<path>` key; give each a distinct `storageKey`.
- **Two search mechanisms coexist on knowledge screens.** The `DataTable` client
  filter (`?q=`, via `useTableState`) and a server-side semantic search (`?qs=`,
  via page-local `useSearchParams`) are independent — don't conflate them. A
  copied list recipe only needs the `?q=` one.
- **`DetailNavigationToolbar` always renders.** It has no empty-list branch, so
  it mounts before the provider's list arrives and shows the buttons' own
  disabled/"–/0" state until items land. `controller.itemsEmpty` exists for a
  caller that wants to skip rendering it, but nothing reads it today.

## Testing

`pnpm run test` (Vitest, `jsdom`, `globals: false` — every test imports
`describe`/`it`/`expect`/`vi` explicitly). Coverage across the surface:

| File                                                     | Focus                                                                                                                                           |
| -------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------- |
| `lib/table-state.test.ts`                                | `load`/`update`/`migrate` — partial patch, empty-collapse, key removal, legacy fold, path variants                                              |
| `lib/view-options-storage.test.ts`                       | load/save/migrate for the view-options slot                                                                                                     |
| `lib/url-params.test.ts`                                 | `mergeHrefWithSearchParams` — append, href-wins, hash preserved, encoding                                                                       |
| `lib/storage-keys.test.ts`                               | `getTopLevelPath` edge cases + `_4_` key join                                                                                                   |
| `hooks/use-table-state.test.tsx`                         | URL read/write, **atomic `update` race regression** (two updaters in one tick keep both params), `?page=1` canonicalization, filter-is-URL-only |
| `hooks/use-table-query-filter.test.tsx`                  | read-only filter subscription                                                                                                                   |
| `hooks/use-page-storage-keys.test.tsx`                   | reactive key resolution (incl. `table_4_` for empty path)                                                                                       |
| `components/ui/data-table.test.tsx`                      | filter/sort/pagination + sorting persistence into `table_4_<path>`                                                                              |
| `components/shared/detail-navigation/*.test.*`           | pure `use-navigation` core (headless), toolbar + sheet behavior                                                                                 |
| `components/shared/inline-edit/use-inline-edit.test.tsx` | edit toggle + deferred focus                                                                                                                    |

`frontend/vitest.setup.ts` polyfills jsdom gaps used by these components and
wires RTL cleanup:

- `Element.prototype.scrollIntoView` (no-op) — roving-focus / scroll-into-view effects.
- `globalThis.ResizeObserver` (no-op class) — Radix ScrollArea in the nav sheet.
- `Element.prototype.{has,set,release}PointerCapture` (no-ops) — Radix Select gating.
- `afterEach(() => cleanup())` — RTL DOM cleanup after every test.

## Removed / renamed — do not use

These names appear in older docs/comments but **do not exist** in the code. Do
not reintroduce them:

| Old / phantom name                                     | Reality                                                                                                                                 |
| ------------------------------------------------------ | --------------------------------------------------------------------------------------------------------------------------------------- |
| `usePagination`                                        | Never existed as a hook. Page state lives in `useTableState` (`pageIndex` in `?page=`) and `<DataTable>` (`pageIndex`/`onPageChange`).  |
| `useTableQueryFilter` (writer)                         | Removed. `use-table-query-filter.ts` exports **only** `useTableQueryFilterReader`. List pages write the filter through `useTableState`. |
| `toolbarProps` spread into `<DetailNavigationToolbar>` | No such object. Pass `controller={nav}` + discrete `renderItem`/`sheetIcon`/`sheetTitle`.                                               |
| `hooks/use-inline-edit`                                | Wrong path. It lives at `components/shared/inline-edit/use-inline-edit.ts`.                                                             |
| `components/shared/data-table`                         | Wrong path. It lives at `components/ui/data-table.tsx`.                                                                                 |
| `features/providers/use-provider-detail-navigation.ts` | Never existed. `/settings/providers/:id` is a form with no sibling nav — only `flows`, `knowledges` and `templates` have nav hooks.     |
| `features/prompts/use-prompt-detail-navigation.ts`     | Never existed. Same for `/settings/prompts/:name`.                                                                                      |
