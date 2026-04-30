// ── REST API paths ──────────────────────────────────────────────────────────

export const RESOURCES_API_PATH = '/resources/';
export const RESOURCES_MKDIR_API_PATH = '/resources/mkdir';
export const RESOURCES_MOVE_API_PATH = '/resources/move';
export const RESOURCES_COPY_API_PATH = '/resources/copy';
export const RESOURCES_DOWNLOAD_API_PATH = '/resources/download';

// ── UI constants ────────────────────────────────────────────────────────────

export const SEARCH_DEBOUNCE_MS = 300;

// ── Upload limits (mirror backend's `pkg/resources/resources.go`) ───────────
//
// The backend does not whitelist file extensions — any file is accepted as
// long as the file name passes `validatePathComponent` (no `/ \ : * ? " < > |`,
// no control characters, ≤ 255 bytes). We re-state the size limits here so the
// client can fail fast instead of waiting for a 4xx response on huge uploads.

/** Mirrors `resources.MaxUploadFileSize` (300 MB). */
export const MAX_FILE_SIZE_MB = 300;

/** Mirrors `resources.MaxUploadTotalSize` (2 GB) — combined size of one request. */
export const MAX_UPLOAD_TOTAL_SIZE_MB = 2 * 1024;

/** Mirrors `resources.MaxUploadFiles`. */
export const MAX_UPLOAD_FILES_PER_REQUEST = 1000;
