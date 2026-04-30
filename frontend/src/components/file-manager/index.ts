export { FileManager } from './file-manager';
export { copyPathAction, deleteAction, downloadAction } from './file-manager-actions';
export { getFileTypeIcon } from './file-manager-icons';
export type { FileTypeIcon } from './file-manager-icons';
export type {
    FileManagerAction,
    FileManagerLabels,
    FileManagerProps,
    FileManagerRootGroup,
    FileNode,
} from './file-manager-types';
export { dedupeOverlappingPaths } from './file-manager-utils';
