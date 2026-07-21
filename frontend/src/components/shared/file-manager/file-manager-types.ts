import type { ComponentType, ReactNode } from 'react';

export interface FileManagerAction {
    /**
     * When true, the action is shown for directory rows. Defaults to false.
     *
     * Combine with {@link FileManagerAction.appliesToFiles} to scope the action:
     *   - `appliesToDirs: false, appliesToFiles: true` (default) → files only
     *   - `appliesToDirs: true,  appliesToFiles: true` (default) → files + directories
     *   - `appliesToDirs: true,  appliesToFiles: false`           → directories only
     *   - `appliesToDirs: false, appliesToFiles: false`           → never (filtered out)
     */
    appliesToDirs?: boolean;
    /**
     * When true (default), the action is shown for file rows. Set to `false` for
     * actions that only make sense on directory rows (e.g. "Upload files",
     * "New folder") combined with `appliesToDirs: true`.
     */
    appliesToFiles?: boolean;
    /**
     * If provided, the action is rendered as an `<a href>` link instead of a button.
     * Useful for native browser downloads.
     */
    getHref?: (file: FileNode) => string;
    /**
     * `download` attribute for the `<a>` (only meaningful with `getHref`). Browser
     * uses it as a hint for the suggested filename — the actual content/extension
     * is determined by the server.
     */
    getHrefDownloadAttr?: (file: FileNode) => string;
    icon?: ComponentType<{ className?: string }>;
    /** Stable identifier — used as React `key`. */
    id: string;
    label: string;
    /** Triggered on item activation. Ignored when `getHref` is set. */
    onSelect: (file: FileNode) => void;
    separatorBefore?: boolean;
    variant?: 'default' | 'destructive';
}

/**
 * Single entry in the bulk-actions bar. By analogy with `FileManagerAction` for
 * row dropdowns: the host owns the array, the bar just renders it.
 *
 * The `files` argument always arrives **deduped** — a directory and its descendants
 * never both appear in the list, the parent wins (see `dedupeOverlappingPaths`).
 */
export interface FileManagerBulkAction {
    /**
     * Optional confirm dialog. When set, the bar gates `onSelect` behind it
     * (used for destructive actions like delete).
     */
    confirm?: FileManagerBulkActionConfirm;
    icon?: ComponentType<{ className?: string }>;
    /** Stable identifier — used as React `key`. */
    id: string;
    /**
     * Greys-out the button when `true`. Receives the deduped selection so the
     * action can disable itself contextually (e.g. when only directories are
     * selected and it doesn't apply to them).
     */
    isDisabled?: (files: FileNode[]) => boolean;
    /**
     * Removes the entry entirely when `true`. Same args as `isDisabled` —
     * useful for actions that simply don't make sense for the current selection.
     */
    isHidden?: (files: FileNode[]) => boolean;
    label: string;
    /** Invoked with the deduped selection. */
    onSelect: (files: FileNode[]) => Promise<void> | void;
    /**
     * When `true`, the action is collapsed into the trailing overflow `…` menu
     * instead of being rendered as a standalone button. Use for less-frequent
     * actions to keep the bar uncluttered.
     */
    overflow?: boolean;
    /** Visual treatment of the inline button (no effect when `overflow: true`). */
    variant?: 'default' | 'destructive';
}

/** Optional confirmation dialog config for a bulk action. */
export interface FileManagerBulkActionConfirm {
    /** Submit-button label (default: action's `label`). */
    confirmText?: string;
    /** Body text. Receives the already-pluralized count label (e.g. "3 items"). */
    description?: (countLabel: string) => string;
    /** Title text. Receives the already-pluralized count label (e.g. "3 items"). */
    title: (countLabel: string) => string;
}

/** Optional column toggles. Both columns default to visible. */
export interface FileManagerColumnsConfig {
    /** Disable sorting on the "Modified" column header. Defaults to `true`. */
    isModifiedSortable?: boolean;
    isModifiedVisible?: boolean;
    /** Disable sorting on the "Name" column header. Defaults to `true`. */
    isNameSortable?: boolean;
    /** Disable sorting on the "Size" column header. Defaults to `true`. */
    isSizeSortable?: boolean;
    isSizeVisible?: boolean;
}

/**
 * Single entry in the right-click context menu surfaced over the empty area
 * of the tree (i.e. anywhere outside a row). Distinct from {@link FileManagerAction}
 * because it cannot reference a `FileNode` — the user clicked between rows,
 * not on one. Typical use: "Upload files", "New folder" at the tree's root.
 *
 * The menu only renders when the host supplies a non-empty
 * `emptyAreaActions` array. When a row has its own context items, the
 * row-level menu wins (right-clicks on rows do not propagate to the empty-area
 * menu — see `file-manager-row.tsx`).
 */
export interface FileManagerEmptyAreaAction {
    icon?: ComponentType<{ className?: string }>;
    /** Stable identifier — used as React `key`. */
    id: string;
    label: string;
    onSelect: () => void;
    separatorBefore?: boolean;
    variant?: 'default' | 'destructive';
}

export interface FileManagerInternalNode extends FileNode {
    children: FileManagerInternalNode[];
    depth: number;
    groupIcon?: ComponentType<{ className?: string }>;
    isGroupRoot?: boolean;
}

/**
 * All user-facing strings. Pass to `FileManager` via the `labels` prop to localize.
 * Every field is optional; defaults are English.
 */
export interface FileManagerLabels {
    /** Cancel button in the bulk-actions bar. */
    bulkCancel?: string;
    /** Trigger label / aria-label for the trailing overflow `…` menu in the bulk bar. */
    bulkMoreActions?: string;
    /** aria-label for the header "expand/collapse all" toggle when the tree is fully expanded. */
    collapseAllAriaLabel?: string;
    columnModified?: string;
    columnName?: string;
    columnSize?: string;
    /** aria-label for the header "expand/collapse all" toggle when at least one directory is collapsed. */
    expandAllAriaLabel?: string;
    /**
     * Custom formatter for the "Modified" column. Receives the raw `modifiedAt`
     * and must return a display string (or empty string for no value). When omitted,
     * the default English relative formatter is used.
     */
    formatModified?: (modifiedAt: Date | string | undefined) => string;
    /**
     * Custom formatter for the cumulative size summary shown in the bulk bar.
     * Receives the byte total of every selected file (directories contribute
     * the recursive sum of their descendants); when omitted, the default
     * `formatFileSize` formatter is used. Return an empty string to suppress
     * the size suffix entirely.
     */
    formatSelectionSize?: (totalBytes: number) => string;
    /** Pluralized "N item" / "N items" used for confirmation dialogs. */
    pluralizeItems?: (count: number) => string;
    /** aria-label for the header "select all" checkbox. */
    selectAllAriaLabel?: string;
    /** Label rendered in the bulk bar, e.g. "3 selected". */
    selectedLabel?: (count: number) => string;
    /**
     * aria-label for a sortable column header button. Receives the column id and the
     * current sort direction (`null` when the column is not currently sorted), and
     * should describe the action the next click will perform. Defaults to a plain
     * English description like `"Sort by name (ascending)"`.
     */
    sortHeaderAriaLabel?: (column: FileManagerSortColumn, direction: FileManagerSortDirection | null) => string;
}

export interface FileManagerProps {
    /** Single, ordered list of available row actions (built-in helpers live in `file-manager-actions`). */
    actions?: readonly FileManagerAction[];
    /**
     * Ordered list of bulk actions rendered in the footer bar when at least one row
     * is selected. When the list is empty / undefined, the bar is not shown at all
     * unless `enableSelection` forces checkboxes (e.g. picker dialogs).
     *
     * Built-in helpers live in `file-manager-actions` (`bulkDeleteAction`,
     * `bulkCopyPathsAction`, `bulkMoveAction`, `bulkCopyAction`).
     *
     * Each callback receives the **deduped** `FileNode[]`: if both a directory and
     * one of its descendants were selected, only the directory is forwarded.
     */
    bulkActions?: readonly FileManagerBulkAction[];
    className?: string;
    /** Per-column visibility flags. Defaults: `{ isSizeVisible: true, isModifiedVisible: true }`. */
    columns?: FileManagerColumnsConfig;
    /**
     * Items rendered in the right-click context menu over the tree's empty
     * area (anywhere outside a row). When omitted / empty, no empty-area
     * menu is registered and the browser's native context menu is shown.
     * See {@link FileManagerEmptyAreaAction} for the item shape.
     */
    emptyAreaActions?: readonly FileManagerEmptyAreaAction[];
    /** Empty state node (rendered when files.length === 0 and not loading). */
    emptyState?: ReactNode;
    /**
     * Forces row checkboxes to be visible / hidden, independently of `bulkActions`.
     * - `true` → checkboxes shown even without bulk actions (e.g. picker dialogs).
     * - `false` → checkboxes hidden even when bulk actions are provided.
     * - `undefined` (default) → checkboxes shown whenever `bulkActions` is non-empty.
     */
    enableSelection?: boolean;
    files: FileNode[];
    /**
     * Initial sort applied on first render when no value is loaded from
     * `sortStorageKey`. Defaults to `null` (no sort, insertion order preserved).
     */
    initialSorting?: FileManagerSortState;
    /**
     * Group directories before files at every tree level when a sort is active.
     * Default: `true` — matches Finder/Explorer behaviour. Set to `false` to mix
     * directories and files together by the chosen sort criterion.
     *
     * Has no effect while sorting is `null` (insertion order is preserved as-is).
     */
    isFoldersFirst?: boolean;
    isLoading?: boolean;
    /** Localizable user-facing strings. */
    labels?: FileManagerLabels;
    /**
     * Fires whenever the focused row changes (roving tabindex). The path is `null`
     * until the user actually focuses a row via click or keyboard navigation —
     * it does NOT auto-fall back to the first visible row, so callers can
     * distinguish "user picked something" from "tree just rendered".
     *
     * Use it to implement context-aware actions (e.g. "Upload here" defaulting
     * to the focused directory, or its parent for files). The supplied callback
     * is read through a ref, so it does not need to be memoized.
     *
     * Emitted values may reference paths that no longer exist in `files` (e.g.
     * the focused row was deleted by an external mutation); consumers should
     * validate against their own data before using the path.
     */
    onActiveRowChange?: (path: null | string) => void;
    /**
     * Optional handler for files dragged in from outside the page (e.g. the
     * desktop / OS file explorer). When provided, dropping files onto a
     * directory row — or onto any file row whose parent is a real directory —
     * fires this callback with the dropped `File[]` and the resolved
     * destination directory path. Drops on the empty area outside any row,
     * on synthetic group-root headers and on top-level files fall through
     * to whatever drag handler the host attaches around `FileManager`
     * (e.g. a page-level DnD upload zone).
     *
     * The callback is independent of {@link onMoveItems}: external-file drop
     * support can be enabled without intra-tree move support, and vice versa.
     */
    onExternalFileDrop?: (files: File[], destinationDir: string) => Promise<void> | void;
    /**
     * Enables internal drag-and-drop: rows become draggable and directories accept drops.
     * Invoked with the dragged item(s) and the destination directory path (`''` for root).
     *
     * `FileManager` does **not** mutate its own list — the caller is expected to refresh
     * the data (or rely on subscriptions) so the new positions become visible.
     */
    onMoveItems?: (sources: FileNode[], destinationDir: string) => Promise<void> | void;
    /**
     * Fired when the user "opens" a *file* row via double-click or `Enter`.
     * Directories go through `onOpenDirectory` instead; when that is omitted,
     * they fall back to the default Finder/Explorer-style expand/collapse.
     *
     * Use it to wire downloads, in-app previews, or open-in-tab behavior.
     */
    onOpen?: (file: FileNode) => void;
    /**
     * Fired when the user "opens" a *directory* row via double-click, `Enter`,
     * or a click on the row's chevron. When provided, **replaces** the default
     * expand/collapse gesture — useful for navigation-style file browsers
     * (e.g. drilling into a remote directory by replacing the listing instead
     * of expanding inline). When omitted, directories keep the default
     * expand/collapse behaviour for all three gestures.
     */
    onOpenDirectory?: (dir: FileNode) => void;
    /**
     * Fires whenever the multi-selection changes. Use it from selection-only
     * flows (e.g. resource pickers) where the parent owns its own confirm button
     * and needs to know which items are currently checked.
     *
     * The supplied callback is read through a ref, so it does not need to be
     * memoized — only meaningful selection changes will trigger it.
     *
     * The Set is owned by the manager and must be treated as read-only —
     * mutating it will desync the manager's internal state.
     */
    onSelectionChange?: (selectedPaths: ReadonlySet<string>) => void;
    /**
     * Fires whenever the active sort changes (header click). The supplied
     * callback is read through a ref, so it does not need to be memoized.
     */
    onSortingChange?: (sorting: FileManagerSortState) => void;
    /** Synthetic top-level groups (e.g. Uploads / Container). When omitted, root is flat. */
    rootGroups?: FileManagerRootGroup[];
    /** Search query and matching empty state. Provide `query` to enable filtering. */
    search?: FileManagerSearchConfig;
    /**
     * Controlled sort state. When set, the manager renders this value and
     * delegates updates to `onSortingChange` (it does NOT update its own
     * state and ignores `sortStorageKey`). Pass `null` to render with no sort.
     */
    sorting?: FileManagerSortState;
    /**
     * When set, the active sort is persisted to `localStorage` under this key
     * across page reloads. Ignored in controlled mode (when `sorting` is set).
     * Pass a route-scoped key (e.g. `getTableStorageKey('/flows/files')`)
     * to avoid collisions across pages.
     */
    sortStorageKey?: string;
}

export interface FileManagerRootGroup {
    defaultOpen?: boolean;
    icon?: ComponentType<{ className?: string }>;
    /** Stable identifier (also used for synthetic node id). */
    id: string;
    label: string;
    /** Path prefix without trailing slash, e.g. `'uploads'` or `'container'`. */
    pathPrefix: string;
}

/** Search-related props, grouped because `query` and `emptyState` always travel together. */
export interface FileManagerSearchConfig {
    /** Empty state node rendered when `query` is set but yields no results. */
    emptyState?: ReactNode;
    /** When set, the tree is filtered to nodes matching the query (and their ancestors), and matches are highlighted. */
    query?: string;
}

/** Sortable column identifier. Mirrors the visible column headers. */
export type FileManagerSortColumn = 'modified' | 'name' | 'size';

/** Direction of an active sort. */
export type FileManagerSortDirection = 'asc' | 'desc';

/** Active sort descriptor; `null` means "no sort, preserve insertion order". */
export type FileManagerSortState = null | {
    column: FileManagerSortColumn;
    direction: FileManagerSortDirection;
};

export interface FileNode {
    id: string;
    isDir: boolean;
    modifiedAt?: Date | string;
    name: string;
    path: string;
    size?: number;
}
