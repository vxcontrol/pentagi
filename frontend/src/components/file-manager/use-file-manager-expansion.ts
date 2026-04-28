import { useCallback, useMemo, useState } from 'react';

import type { FileManagerInternalNode, FileManagerRootGroup } from './file-manager-types';

import { collectDirectoryPaths } from './file-manager-utils';

type ExpansionOverride = 'force-collapsed' | 'force-expanded';

interface UseFileManagerExpansion {
    /** Effective set of expanded directory paths (auto-expansion + user overrides). */
    expandedPaths: Set<string>;
    /** Toggle expansion of a directory; persists as a user override. */
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
 * Instead we keep a small map of *user-driven* overrides on top of two automatic sources:
 *   1) `rootGroups[].defaultOpen` — initial group expansion,
 *   2) on-search auto-expansion of every directory inside the filtered tree.
 */
export const useFileManagerExpansion = ({
    isFiltering,
    normalizedRootGroups,
    visibleTree,
}: UseFileManagerExpansionArgs): UseFileManagerExpansion => {
    const [overrides, setOverrides] = useState<Map<string, ExpansionOverride>>(() => new Map());

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

        for (const [path, override] of overrides) {
            if (override === 'force-expanded') {
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

            next.set(path, wasExpanded ? 'force-collapsed' : 'force-expanded');

            return next;
        });
    }, []);

    return { expandedPaths, toggleExpand };
};
