import { Folder, FolderUp, HardDrive } from 'lucide-react';

import type { FileManagerRootGroup } from '@/components/shared/file-manager';

export const SEARCH_DEBOUNCE_MS = 300;

export const UPLOADS_PATH_PREFIX = 'uploads';
export const RESOURCES_PATH_PREFIX = 'resources';
export const CONTAINER_PATH_PREFIX = 'container';

export const ROOT_GROUPS: FileManagerRootGroup[] = [
    { defaultOpen: true, icon: FolderUp, id: 'uploads', label: 'Uploads', pathPrefix: UPLOADS_PATH_PREFIX },
    { defaultOpen: true, icon: Folder, id: 'resources', label: 'Resources', pathPrefix: RESOURCES_PATH_PREFIX },
    { defaultOpen: true, icon: HardDrive, id: 'container', label: 'Container', pathPrefix: CONTAINER_PATH_PREFIX },
];

export const FLOW_FILES_API_PATH = (flowId: string) => `/flows/${flowId}/files/`;
export const FLOW_FILES_PULL_API_PATH = (flowId: string) => `/flows/${flowId}/files/pull`;
export const FLOW_FILES_CONTAINER_API_PATH = (flowId: string) => `/flows/${flowId}/files/container`;
export const FLOW_FILES_ATTACH_RESOURCES_API_PATH = (flowId: string) => `/flows/${flowId}/files/resources`;
export const FLOW_FILES_PROMOTE_API_PATH = (flowId: string) => `/flows/${flowId}/files/to-resources`;
export const RESOURCES_LIST_API_PATH = '/resources/';

export const UPLOADS_TARGET_DIRECTORY = '/work/uploads';
export const CONTAINER_TARGET_DIRECTORY = 'container/';
export const RESOURCES_TARGET_DIRECTORY = '/work/resources';

/** Default container path browsed when the Pull dialog opens. */
export const CONTAINER_DEFAULT_PATH = '/work';

// ── Upload limits (mirror backend's `pkg/flowfiles/files.go`) ──────────────
//
// Symmetric with the resources library limits (`features/resources/resources-constants.ts`)
// today, but kept as a separate set so they can diverge from resources without
// touching the resources call sites.

/** Mirrors `flowfiles.MaxUploadFileSize` (300 MB). */
export const FLOW_FILES_MAX_FILE_SIZE_MB = 300;

/** Mirrors `flowfiles.MaxUploadTotalSize` (2 GB) — combined size of one request. */
export const FLOW_FILES_MAX_UPLOAD_TOTAL_SIZE_MB = 2 * 1024;

/** Mirrors `flowfiles.MaxUploadFiles`. */
export const FLOW_FILES_MAX_UPLOAD_FILES_PER_REQUEST = 1000;
