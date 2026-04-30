import { Copy, Download, Trash2 } from 'lucide-react';

import type { FileManagerAction, FileNode } from './file-manager-types';

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
    icon: Copy,
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
