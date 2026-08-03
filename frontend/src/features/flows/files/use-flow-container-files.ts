import { useCallback, useEffect, useMemo, useRef, useState } from 'react';

import type { FileNode } from '@/components/shared/file-manager';

import { buildPathsQuery } from '@/features/resources/resources-utils';
import { api, getApiErrorMessage, unwrapApiResponse } from '@/lib/axios';

import { FLOW_FILES_CONTAINER_API_PATH } from './flow-files-constants';
import { type ContainerFileFailure, type ContainerFilesResponse, containerFileToFileNode } from './flow-files-utils';

interface UseFlowContainerFilesParams {
    flowId: null | string;
    /**
     * Absolute container paths to list (e.g. `['/work']`, `['/etc', '/var/log']`).
     * Backend accepts multiple via `?paths[]=…&paths[]=…` and returns a deduped,
     * sort-stable union. Empty array disables the request entirely — handy for
     * "skip until ready" patterns when the dialog is closed.
     */
    paths: readonly string[];
}

interface UseFlowContainerFilesResult {
    error: Error | null;
    /**
     * Entries the backend could not stat (dangling symlink, a file removed
     * mid-listing, a transient /proc pid). The listing still succeeds with the
     * readable entries in `files`; these are surfaced so the UI can warn that
     * the directory is shown partially.
     */
    failures: ContainerFileFailure[];
    files: FileNode[];
    isLoading: boolean;
    /**
     * Re-fetches the listing using the current `paths`. Use after a Pull
     * succeeds so the user sees the updated container state without closing
     * the dialog.
     */
    refetch: () => Promise<void>;
    /** True when the directory had more entries than the cap and only the first page was returned. */
    truncated: boolean;
}

/**
 * Live listing of files inside the running container, returned as `FileNode[]`
 * so the FileManager component can render them directly.
 *
 * The Pull dialog uses this to let the user browse `/work` (or any other
 * absolute path inside the container) and pick which entries to copy into the
 * local cache. The listing is shallow — backend returns a single directory's
 * contents — and the dialog handles drill-in by replacing `paths` with the
 * folder the user just opened.
 *
 * When `flowId` is `null` or `paths` is empty, the hook idles: no request
 * is issued, `files` stays an empty array, `isLoading` stays `false`.
 */
export function useFlowContainerFiles({ flowId, paths }: UseFlowContainerFilesParams): UseFlowContainerFilesResult {
    const [files, setFiles] = useState<FileNode[]>([]);
    const [failures, setFailures] = useState<ContainerFileFailure[]>([]);
    const [truncated, setTruncated] = useState(false);
    const [isLoading, setIsLoading] = useState(false);
    const [error, setError] = useState<Error | null>(null);

    // Stringify the paths array so that effect / refetch deps stay stable across
    // renders (a fresh array reference would re-fire the effect on every parent
    // re-render even when the contents are identical). The `\u0000` separator
    // can't appear in any container path so it's a safe delimiter.
    const pathsKey = useMemo(() => paths.join('\u0000'), [paths]);

    // Each in-flight fetch increments this counter; the response-handling code
    // bails when its captured token no longer matches `currentTokenRef.current`,
    // which guarantees the latest request always wins (no race-induced flicker
    // when the user drills in / out faster than the server can respond).
    const currentTokenRef = useRef(0);

    // The path key of the last fetch. A fetch clears the visible listing only when
    // the path actually changes, so an in-place refetch (Refresh / post-Pull) keeps
    // the current rows under the loading guard instead of flashing a skeleton.
    const lastPathsKeyRef = useRef<null | string>(null);

    const fetchListing = useCallback(async () => {
        if (!flowId || paths.length === 0) {
            currentTokenRef.current += 1;
            lastPathsKeyRef.current = null;
            setFiles([]);
            setFailures([]);
            setTruncated(false);
            setIsLoading(false);
            setError(null);

            return;
        }

        const token = (currentTokenRef.current += 1);

        setIsLoading(true);
        setError(null);

        // Only reset the listing when navigating to a different path, so a pending
        // navigation never shows the old listing under the new breadcrumb — a
        // same-path refetch keeps the current rows visible.
        if (pathsKey !== lastPathsKeyRef.current) {
            setFiles([]);
            setFailures([]);
            setTruncated(false);
        }

        lastPathsKeyRef.current = pathsKey;

        try {
            const url = `${FLOW_FILES_CONTAINER_API_PATH(flowId)}?${buildPathsQuery(paths)}`;
            // Listing the container can take several seconds for huge directory
            // trees; opt out of the default 30s axios timeout so the dialog
            // doesn't fail spuriously while the kernel walks /proc, /sys, etc.
            const response = await api.get<ContainerFilesResponse>(url, { timeout: 0 });
            const data = unwrapApiResponse(response);

            if (token !== currentTokenRef.current) {
                return;
            }

            setFiles(data.files.map(containerFileToFileNode));
            setFailures(data.failures ?? []);
            setTruncated(data.truncated ?? false);
        } catch (caught) {
            if (token !== currentTokenRef.current) {
                return;
            }

            setError(new Error(getApiErrorMessage(caught, 'Failed to load container files')));
            setFiles([]);
            setFailures([]);
            setTruncated(false);
        } finally {
            if (token === currentTokenRef.current) {
                setIsLoading(false);
            }
        }
    }, [flowId, paths, pathsKey]);

    useEffect(() => {
        // fetchListing is an async callback that handles its own loading state via setState
        // after await — not synchronously inside the effect body. We also depend on the
        // stable `pathsKey` instead of the array reference so identical contents don't
        // re-fire the request on every render. `flowId` triggers a fresh fetch when the
        // user switches flows.
        // eslint-disable-next-line react-hooks/set-state-in-effect
        void fetchListing();
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [flowId, pathsKey]);

    return {
        error,
        failures,
        files,
        isLoading,
        refetch: fetchListing,
        truncated,
    };
}
