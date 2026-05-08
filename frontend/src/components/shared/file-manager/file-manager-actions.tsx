import { ClipboardCopy, Copy, Download, FileSymlink, FolderOutput, Trash2 } from 'lucide-react';

import type { FileManagerAction, FileManagerBulkAction, FileNode } from './file-manager-types';

/**
 * Built-in download action.
 *
 * The `getDownloadHref` callback receives the file node and must return a fully-formed
 * URL. For directories the browser is hinted to save the response as `${name}.zip` —
 * the backend is responsible for actually returning a zip stream.
 */
export const downloadAction = (
    getDownloadHref: (file: FileNode) => string,
    options: { directoryArchiveExtension?: string } = {},
): FileManagerAction => {
    const archiveExtension = options.directoryArchiveExtension ?? 'zip';

    return {
        appliesToDirs: true,
        getHref: getDownloadHref,
        getHrefDownloadAttr: (file) => (file.isDir ? `${file.name}.${archiveExtension}` : file.name),
        icon: Download,
        id: '__builtin_download',
        label: 'Download',
        onSelect: () => {},
    };
};

/** Built-in copy-path action. */
export const copyPathAction = (onCopyPath: (file: FileNode) => void): FileManagerAction => ({
    appliesToDirs: true,
    icon: ClipboardCopy,
    id: '__builtin_copy_path',
    label: 'Copy path',
    onSelect: onCopyPath,
});

/**
 * Built-in delete action. Always rendered with a leading separator and destructive variant.
 * Caller is responsible for showing a confirmation dialog inside `onSelect`.
 */
export const deleteAction = (onDelete: (file: FileNode) => void): FileManagerAction => ({
    appliesToDirs: true,
    icon: Trash2,
    id: '__builtin_delete',
    label: 'Delete',
    onSelect: onDelete,
    separatorBefore: true,
});

// ── Bulk-action helpers ─────────────────────────────────────────────────────
//
// Each helper produces a `FileManagerBulkAction` with sensible defaults; the
// caller passes any callback / config it needs and lets the bar handle the
// rendering, confirmation and dedup. Mirrors the row-action helpers above so
// consumers compose `bulkActions={[bulkXAction(...), bulkYAction(...)]}` the
// same way they compose `actions={[xAction(...), yAction(...)]}`.

interface BulkDeleteOptions {
    /** Confirm-dialog body formatter. Default: "This will delete N items. This action cannot be undone." */
    confirmDescription?: (countLabel: string) => string;
    /** Confirm-button label. Default: action `label`. */
    confirmText?: string;
    /** Confirm-dialog title formatter. Default: "Delete N items". */
    confirmTitle?: (countLabel: string) => string;
    /** Trigger label in the bar. Default: "Delete". */
    label?: string;
}

/**
 * Built-in bulk delete. Wraps the destructive variant + confirm dialog so callers
 * just supply the API call. Always shown as a standalone (non-overflow) button so
 * the destructive intent stays visible.
 */
export const bulkDeleteAction = (
    onDelete: (files: FileNode[]) => Promise<void> | void,
    options: BulkDeleteOptions = {},
): FileManagerBulkAction => {
    const label = options.label ?? 'Delete';

    return {
        confirm: {
            confirmText: options.confirmText ?? label,
            description:
                options.confirmDescription ??
                ((countLabel) => `This will delete ${countLabel}. This action cannot be undone.`),
            title: options.confirmTitle ?? ((countLabel) => `Delete ${countLabel}`),
        },
        icon: Trash2,
        id: '__builtin_bulk_delete',
        label,
        onSelect: onDelete,
        variant: 'destructive',
    };
};

/**
 * Built-in "copy paths" bulk action. Joins every selected file's `path` with `\n`
 * and hands the resulting string to `onCopy`. Lives in the overflow menu by default
 * so it doesn't crowd the bar — pass `overflow: false` to promote it to a button.
 */
export const bulkCopyPathsAction = (
    onCopy: (paths: string[]) => Promise<void> | void,
    options: { label?: string; overflow?: boolean } = {},
): FileManagerBulkAction => ({
    icon: ClipboardCopy,
    id: '__builtin_bulk_copy_paths',
    label: options.label ?? 'Copy paths',
    onSelect: (files) => onCopy(files.map((file) => file.path)),
    overflow: options.overflow ?? true,
});

/**
 * Built-in "move to…" bulk action. The host opens its own destination picker
 * inside `onMove` (we don't ship a path-picker UI inside the file-manager).
 */
export const bulkMoveAction = (
    onMove: (files: FileNode[]) => void,
    options: { label?: string; overflow?: boolean } = {},
): FileManagerBulkAction => ({
    icon: FileSymlink,
    id: '__builtin_bulk_move',
    label: options.label ?? 'Move to…',
    onSelect: onMove,
    overflow: options.overflow,
});

/**
 * Built-in "copy to…" bulk action. The host opens its own destination picker
 * inside `onCopy`.
 */
export const bulkCopyAction = (
    onCopy: (files: FileNode[]) => void,
    options: { label?: string; overflow?: boolean } = {},
): FileManagerBulkAction => ({
    icon: Copy,
    id: '__builtin_bulk_copy',
    label: options.label ?? 'Copy to…',
    onSelect: onCopy,
    overflow: options.overflow,
});

/**
 * Built-in "save as resource" bulk action used by Flow Files. Identical shape to
 * `bulkMoveAction` — the host opens its own promote dialog inside `onPromote`.
 */
export const bulkPromoteAction = (
    onPromote: (files: FileNode[]) => void,
    options: { label?: string; overflow?: boolean } = {},
): FileManagerBulkAction => ({
    icon: FolderOutput,
    id: '__builtin_bulk_promote',
    label: options.label ?? 'Save as resources',
    onSelect: onPromote,
    overflow: options.overflow,
});

interface BulkDownloadOptions {
    /** Override the suggested filename hint passed to the browser. */
    getDownloadName?: (files: FileNode[]) => string;
    label?: string;
    overflow?: boolean;
}

/**
 * Default browser-side filename hint for bulk download. Mirrors what the
 * backend will set via `Content-Disposition` so the browser shows a sensible
 * name even when the response header is missing or stripped:
 *
 *   - 1 file       → original filename
 *   - 1 directory  → `${name}.zip`
 *   - many entries → `download.zip`
 */
const defaultBulkDownloadName = (files: FileNode[]): string => {
    const [single, ...rest] = files;

    if (single && rest.length === 0) {
        return single.isDir ? `${single.name}.zip` : single.name;
    }

    return 'download.zip';
};

/**
 * Built-in "download" bulk action. The host supplies a function that builds the
 * download URL from the selection (typically with `?paths[]=…&paths[]=…`); the
 * helper takes care of triggering the actual browser download via a transient
 * `<a download>` element. Backend decides the response shape: a single file is
 * served as-is, anything else is packaged into a ZIP archive.
 */
export const bulkDownloadAction = (
    getDownloadHref: (files: FileNode[]) => string,
    options: BulkDownloadOptions = {},
): FileManagerBulkAction => ({
    icon: Download,
    id: '__builtin_bulk_download',
    label: options.label ?? 'Download',
    onSelect: (files) => {
        if (files.length === 0) {
            return;
        }

        const href = getDownloadHref(files);

        // Trigger the download via a transient anchor so the browser respects
        // the `download` attribute and the backend's `Content-Disposition`.
        // `window.open` would do, but it can be blocked as a popup and doesn't
        // honour the filename hint the same way.
        const anchor = document.createElement('a');

        anchor.href = href;
        anchor.download = (options.getDownloadName ?? defaultBulkDownloadName)(files);
        anchor.rel = 'noopener';
        document.body.appendChild(anchor);
        anchor.click();
        anchor.remove();
    },
    overflow: options.overflow,
});
