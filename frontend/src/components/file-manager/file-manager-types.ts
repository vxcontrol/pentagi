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
    /** Confirm button in the bulk-delete dialog. */
    bulkConfirm?: string;
    /** Description of the bulk-delete dialog. Receives the already-pluralized count label. */
    bulkDescription?: (countLabel: string) => string;
    /** Title of the bulk-delete dialog. Receives the already-pluralized count label. */
    bulkTitle?: (countLabel: string) => string;
    columnModified?: string;
    columnName?: string;
    columnSize?: string;
    /**
     * Custom formatter for the "Modified" column. Receives the raw `modifiedAt`
     * and must return a display string (or empty string for no value). When omitted,
     * the default English relative formatter is used.
     */
    formatModified?: (modifiedAt: Date | string | undefined) => string;
    /** Pluralized "N item" / "N items" used for the dialog title and description. */
    pluralizeItems?: (count: number) => string;
    /** aria-label for the header "select all" checkbox. */
    selectAllAriaLabel?: string;
    /** Label rendered in the bulk bar, e.g. "3 selected". */
    selectedLabel?: (count: number) => string;
}

export interface FileManagerProps {
    /** Single, ordered list of available row actions (built-in helpers live in `file-manager-actions`). */
    actions?: readonly FileManagerAction[];
    className?: string;
    /** Per-column visibility flags. Defaults: `{ isSizeVisible: true, isModifiedVisible: true }`. */
    columns?: FileManagerColumnsConfig;
    /** Empty state node (rendered when files.length === 0 and not loading). */
    emptyState?: ReactNode;
    files: FileNode[];
    isLoading?: boolean;
    /** Localizable user-facing strings. */
    labels?: FileManagerLabels;
    /**
     * Callback invoked when bulk delete is confirmed. The list is already deduped:
     * if both a directory and one of its descendants were selected, only the directory is included.
     */
    onBulkDelete?: (files: FileNode[]) => Promise<void> | void;
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
