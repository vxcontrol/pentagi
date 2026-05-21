# Shared list/detail building blocks

This directory hosts the reusable surface for list-and-detail pages: a
filterable table, a Prev/Next/Sheet toolbar that walks the _same_ filtered
subset on detail pages, and the inline-rename + sortable-header primitives
that every list reuses.

## Mental model

```
                ┌──────────────────────────────────────────┐
                │  URL  ?q=foo  ?page=3                    │
                │  (source of truth — bookmarkable)        │
                └────────┬─────────────────────┬───────────┘
                         │ read/write          │ read-only
                         ▼                     ▼
              useTableQueryFilter     useTableQueryFilterReader
              usePagination                    │
                         │                     │
                         ▼                     ▼
                    <DataTable>      useNavigation
                    (list page)             │
                         │                  ▼
                         ▼          <DetailNavigationToolbar>
                  table_4_<path>       (detail page)
                  in localStorage
                  (cold-start fallback)
```

- **URL is authoritative.** Filter (`?q=`) and page (`?page=`) live in the
  URL so links/bookmarks always reproduce the user's view.
- **Storage is a warm-restart bag.** The list page persists the URL filter
  into `localStorage` under `table_4_<path>`. The detail page never writes
  storage and never replays storage into the URL — opening a shared link
  shows exactly what the link says.
- **Prev/Next walks the same subset.** `DetailNavigationToolbar` runs the
  same matcher (`createTextMatcher`) the list filter uses, so siblings stay
  in lockstep with what the user sees in the table.

## Components

| File                                                       | Role                                                                                                    |
| ---------------------------------------------------------- | ------------------------------------------------------------------------------------------------------- |
| [`detail-navigation/`](detail-navigation/)                 | Prev / Position / Next toolbar + listbox sheet for detail pages, and the navigation hooks that feed it. |
| [`inline-edit/`](inline-edit/)                             | Generic inline-edit input (Save/Cancel addons, Enter/Escape) plus the paired `useInlineEdit` state machine. |

## Hooks

| Hook                        | Where                           | Source of truth | Writes? | Notes                                                              |
| --------------------------- | ------------------------------- | --------------- | ------- | ------------------------------------------------------------------ |
| `useTableQueryFilter`       | `@/hooks/`                      | URL `?q=`       | yes     | List pages. Restores from `localStorage` on cold start.            |
| `useTableQueryFilterReader` | `@/hooks/`                      | URL `?q=`       | no      | Detail pages. Storage-blind — shared links never gain stale `?q=`. |
| `usePagination`             | `@/hooks/`                      | URL `?page=`    | yes     | Canonicalizes `?page=1` away so the URL has one form per view.     |
| `useNavigation`             | `detail-navigation/` (internal) | props           | no      | Pure computation of Prev/Next around a `currentId`.                |
| `useDetailNavigation`       | `detail-navigation/`            | URL + props     | no      | Bundles the three above into a single hook for detail pages.       |
| `useInlineEdit`             | `inline-edit/`                  | local state     | no      | Edit-mode toggle + deferred focus (Radix dropdown race fix).       |
| `usePageStorageKeys`        | `@/hooks/`                      | router          | no      | Resolves the three per-page storage keys reactively.               |

## Library helpers (in `@/lib/`)

| Module                    | Purpose                                                                              |
| ------------------------- | ------------------------------------------------------------------------------------ |
| `table-state.ts`          | Unified `table_4_<path>` JSON slot. Carries filter + sorting + columnVis + pageSize. |
| `view-options-storage.ts` | `viewOptions_4_<path>` for FileManager-style screens (folders-first, etc.).          |
| `storage-keys.ts`         | Single source of truth for storage-key conventions and `getTopLevelPath`.            |
| `url-params.ts`           | `URL_PARAMS` constants + `mergeHrefWithSearchParams` (preserves hash on merge).      |

## How to add a new list + detail pair

1. **List page** (`/<entities>/`):

    ```tsx
    const { filter, setFilter } = useTableQueryFilter();
    const { pageIndex, setPage } = usePagination();

    return (
        <DataTable
            columns={columns /* use <DataTableColumnHeader column={column} title="..." /> */}
            data={entities}
            filterColumn="title"
            filterValue={filter}
            onFilterChange={setFilter}
            onPageChange={setPage}
            pageIndex={pageIndex}
        />
    );
    ```

2. **Feature-scoped navigation hook** (`@/features/<entity>/use-<entity>-detail-navigation.ts`):

    ```ts
    const getLabel = (item: Entity) => item.title;
    const getHref = (item: Entity) => `/<entities>/${item.id}`;

    export const useEntityDetailNavigation = (currentId: null | string | undefined) => {
        const { entities } = useEntities();

        return useDetailNavigation<Entity>({ currentId, getHref, getLabel, items: entities });
    };
    ```

3. **Detail page** (`/<entities>/:id`):

    ```tsx
    const { toolbarProps } = useEntityDetailNavigation(entityId);

    return (
        <header>
            <DetailNavigationToolbar<Entity>
                {...toolbarProps}
                sheetIcon={<Icon className="size-4" />}
                sheetTitle="Entities"
                renderItem={(item, isCurrent) => <span>{item.title}</span>}
            />
        </header>
    );
    ```

## Why URL > storage

A user opens `/flows?q=alpha` in tab A. They navigate to flow B by clicking
"Next" in the toolbar. They share `/flows/b?q=alpha` with a teammate.

- The teammate opens the link cold. Their detail page reads `q=alpha` from
  the URL and renders Prev/Next over the filtered subset.
- The teammate hits "Next". They land on `/flows/c?q=alpha` — still inside
  the filter, even though they never typed it.

`useTableQueryFilterReader` is the key piece: it observes the URL but never
writes anything, so a fresh detail-page mount can't accidentally inject the
**previous tab's** `?q=` into the URL.

## Why one storage key per page

Before the unification, every list page wrote four storage keys
(`column_4_/flows`, `sorting_4_/flows`, `filter_4_/flows`, `page_4_/flows`)
in two different write paths (sync + debounced). Refreshing during a typing
session could land you in an inconsistent state. The unified
`table_4_<path>` slot is a single JSON object that all preferences live in;
`migrateLegacyTableState` folds the four legacy keys into it on first mount
and deletes them.

## Testing notes

- `vitest run` covers the pure utilities (`table-state`,
  `view-options-storage`, `url-params`), the hook behaviours
  (`use-pagination`, `use-table-query-filter`, `use-inline-edit`,
  `use-page-storage-keys`, `use-detail-navigation`), and the components
  (`detail-navigation/`, `data-table`).
- jsdom doesn't ship `Element.prototype.scrollIntoView` or `ResizeObserver`
  — both are polyfilled in `vitest.setup.ts`.
- React Testing Library auto-cleans the DOM after every test (see the same
  setup file). Tests can freely call `render` without leaking nodes.
