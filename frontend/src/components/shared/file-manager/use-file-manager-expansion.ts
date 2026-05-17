import { useCallback, useMemo, useState } from 'react';

import type { FileManagerInternalNode, FileManagerRootGroup } from './file-manager-types';

import { collectDirectoryPaths } from './file-manager-utils';

interface UseFileManagerExpansion {
    /** Effective set of expanded directory paths (auto-expansion + user overrides). */
    expandedPaths: Set<string>;
    /**
     * Bulk override for "expand all" / "collapse all" gestures. Each path in
     * `paths` gets its override set to `isExpanded`, replacing any previous
     * per-path value. Passing an empty iterable is a no-op.
     */
    setExpansion: (paths: Iterable<string>, isExpanded: boolean) => void;
    /**
     * Toggle expansion of a directory based on the state the caller is currently
     * displaying. Passing `wasExpanded` is intentional — it captures the user's
     * intent ("invert what I see") and lets the callback stay stable across renders,
     * which keeps memoized rows from invalidating on every expansion change.
     */
    toggleExpand: (path: string, wasExpanded: boolean) => void;
}

interface UseFileManagerExpansionArgs {
    isFiltering: boolean;
    /** Already-normalized root groups (pathPrefix without trailing slash). */
    normalizedRootGroups: FileManagerRootGroup[] | undefined;
    visibleTree: FileManagerInternalNode[];
}

/**
 * Expansion state without `useEffect`: we never store the full set in state.
 * Instead we keep a small map of *user-driven* overrides (path -> isExpanded) on top
 * of two automatic sources:
 *   1) `rootGroups[].defaultOpen` — initial group expansion,
 *   2) on-search auto-expansion of every directory inside the filtered tree.
 */
export function useFileManagerExpansion({
    isFiltering,
    normalizedRootGroups,
    visibleTree,
}: UseFileManagerExpansionArgs): UseFileManagerExpansion {
    const [overrides, setOverrides] = useState<Map<string, boolean>>(() => new Map());

    const expandedPaths = useMemo(() => {
        const result = new Set<string>();

        if (normalizedRootGroups?.length) {
            for (const group of normalizedRootGroups) {
                if (group.defaultOpen ?? true) {
                    result.add(group.pathPrefix);
                }
            }
        }

        if (isFiltering) {
            for (const dirPath of collectDirectoryPaths(visibleTree)) {
                result.add(dirPath);
            }
        }

        for (const [path, isExpanded] of overrides) {
            if (isExpanded) {
                result.add(path);
            } else {
                result.delete(path);
            }
        }

        return result;
    }, [isFiltering, normalizedRootGroups, overrides, visibleTree]);

    const toggleExpand = useCallback((path: string, wasExpanded: boolean) => {
        setOverrides((prev) => {
            const next = new Map(prev);

            next.set(path, !wasExpanded);

            return next;
        });
    }, []);

    const setExpansion = useCallback((paths: Iterable<string>, isExpanded: boolean) => {
        setOverrides((prev) => {
            const next = new Map(prev);
            let didChange = false;

            for (const path of paths) {
                if (next.get(path) !== isExpanded) {
                    next.set(path, isExpanded);
                    didChange = true;
                }
            }

            return didChange ? next : prev;
        });
    }, []);

    return { expandedPaths, setExpansion, toggleExpand };
}
