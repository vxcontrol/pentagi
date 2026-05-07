import type { ComponentType, ReactNode } from 'react';

export interface FileManagerAction {
    /** When true, action is shown for both files and directories. Defaults to false (files only). */
    appliesToDirs?: boolean;
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
    /** When true, separator is rendered before this action. */
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
    isModifiedVisible?: boolean;
    isSizeVisible?: boolean;
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
    columnModified?: string;
    columnName?: string;
    columnSize?: string;
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
    isLoading?: boolean;
    /** Localizable user-facing strings. */
    labels?: FileManagerLabels;
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
     * Fired when the user "opens" a *directory* row via double-click or `Enter`.
     * When provided, **replaces** the default expand/collapse gesture — useful
     * for navigation-style file browsers (e.g. drilling into a remote directory
     * by replacing the listing instead of expanding inline). When omitted,
     * directories keep the default expand/collapse behaviour.
     *
     * The chevron icon on the row's left edge always toggles expand/collapse
     * regardless of this prop, so the user still has access to inline
     * exploration when it makes sense.
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
    /** Synthetic top-level groups (e.g. Uploads / Container). When omitted, root is flat. */
    rootGroups?: FileManagerRootGroup[];
    /** Search query and matching empty state. Provide `query` to enable filtering. */
    search?: FileManagerSearchConfig;
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

export interface FileNode {
    id: string;
    isDir: boolean;
    modifiedAt?: Date | string;
    name: string;
    path: string;
    size?: number;
}
