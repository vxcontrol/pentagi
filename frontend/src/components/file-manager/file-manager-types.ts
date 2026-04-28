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

export interface FileManagerInternalNode extends FileNode {
    children: FileManagerInternalNode[];
    depth: number;
    groupIcon?: ComponentType<{ className?: string }>;
    isGroupRoot?: boolean;
}

export interface FileManagerProps {
    /** Single, ordered list of available row actions (built-in helpers live in `file-manager-actions`). */
    actions?: FileManagerAction[];
    className?: string;
    /** Empty state node (rendered when files.length === 0 and not loading). */
    emptyState?: ReactNode;
    files: FileNode[];
    isLoading?: boolean;
    /**
     * Callback invoked when bulk delete is confirmed. The list is already deduped:
     * if both a directory and one of its descendants were selected, only the directory is included.
     */
    onBulkDelete?: (files: FileNode[]) => Promise<void> | void;
    /** Synthetic top-level groups (e.g. Uploads / Container). When omitted, root is flat. */
    rootGroups?: FileManagerRootGroup[];
    /** Empty state node rendered when `searchQuery` is set but yields no results. */
    searchEmptyState?: ReactNode;
    /** When set, the tree is filtered to nodes matching the query (and their ancestors), and matches are highlighted. */
    searchQuery?: string;
    showModifiedColumn?: boolean;
    showSizeColumn?: boolean;
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

export interface FileNode {
    id: string;
    isDir: boolean;
    modifiedAt?: Date | string;
    name: string;
    path: string;
    size?: number;
}
