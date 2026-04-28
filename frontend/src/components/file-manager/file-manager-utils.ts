import type { MouseEvent as ReactMouseEvent } from 'react';

import type { FileManagerInternalNode, FileManagerRootGroup, FileNode } from './file-manager-types';

export const formatFileSize = (size?: number): string => {
    if (size == null) {
        return '';
    }

    if (size < 1024) {
        return `${size} B`;
    }

    const units = ['KB', 'MB', 'GB', 'TB'];
    let unitIndex = -1;
    let value = size;

    while (value >= 1024 && unitIndex < units.length - 1) {
        value /= 1024;
        unitIndex += 1;
    }

    return `${value.toFixed(value >= 10 ? 0 : 1)} ${units[unitIndex]}`;
};

export const formatModified = (modifiedAt?: Date | string): string => {
    if (!modifiedAt) {
        return '';
    }

    const date = typeof modifiedAt === 'string' ? new Date(modifiedAt) : modifiedAt;

    if (Number.isNaN(date.getTime())) {
        return '';
    }

    const now = Date.now();
    const diffMs = now - date.getTime();
    const diffMin = Math.round(diffMs / 60_000);

    if (diffMin < 1) {
        return 'just now';
    }

    if (diffMin < 60) {
        return `${diffMin}m ago`;
    }

    const diffHours = Math.round(diffMin / 60);

    if (diffHours < 24) {
        return `${diffHours}h ago`;
    }

    const diffDays = Math.round(diffHours / 24);

    if (diffDays < 7) {
        return `${diffDays}d ago`;
    }

    return date.toLocaleDateString();
};

/** Strip trailing slashes from a `pathPrefix`. Empty / `'/'` collapse to `''`. */
const stripTrailingSlash = (input: string): string => input.replace(/\/+$/, '');

/**
 * Normalize root groups once: trim trailing slashes from `pathPrefix`.
 * Returning the original reference when nothing changed lets `useMemo` keep identity.
 */
export const normalizeRootGroups = (
    rootGroups: FileManagerRootGroup[] | undefined,
): FileManagerRootGroup[] | undefined => {
    if (!rootGroups?.length) {
        return rootGroups;
    }

    let mutated = false;
    const normalized = rootGroups.map((group) => {
        const trimmed = stripTrailingSlash(group.pathPrefix);

        if (trimmed === group.pathPrefix) {
            return group;
        }

        mutated = true;

        return { ...group, pathPrefix: trimmed };
    });

    return mutated ? normalized : rootGroups;
};

const buildSyntheticGroupRoot = (group: FileManagerRootGroup): FileManagerInternalNode => ({
    children: [],
    depth: 0,
    groupIcon: group.icon,
    id: `__group__${group.id}`,
    isDir: true,
    isGroupRoot: true,
    name: group.label,
    path: group.pathPrefix,
});

const buildSyntheticFolder = (folderPath: string, folderName: string, depth: number): FileManagerInternalNode => ({
    children: [],
    depth,
    id: `__dir__${folderPath}`,
    isDir: true,
    name: folderName,
    path: folderPath,
});

const findGroupIndexForFile = (path: string, groups: FileManagerRootGroup[]): number =>
    groups.findIndex((group) => path === group.pathPrefix || path.startsWith(`${group.pathPrefix}/`));

export const buildFileManagerTree = (
    files: FileNode[],
    rootGroups?: FileManagerRootGroup[],
): FileManagerInternalNode[] => {
    const sorted = [...files].sort((a, b) => a.path.localeCompare(b.path));
    const groups = normalizeRootGroups(rootGroups);
    const roots: FileManagerInternalNode[] = groups?.length ? groups.map(buildSyntheticGroupRoot) : [];

    // O(1) folder lookup by absolute path; replaces O(n) `siblings.find(...)`.
    const folderByPath = new Map<string, FileManagerInternalNode>();

    for (const file of sorted) {
        const segments = file.path.split('/').filter(Boolean);

        if (segments.length === 0) {
            continue;
        }

        let siblings: FileManagerInternalNode[];
        let startIndex: number;
        let baseDepth: number;

        if (groups?.length) {
            const groupIndex = findGroupIndexForFile(file.path, groups);

            if (groupIndex < 0) {
                continue;
            }

            const groupRoot = roots[groupIndex];

            if (!groupRoot) {
                continue;
            }

            siblings = groupRoot.children;
            startIndex = 1;
            baseDepth = 1;
        } else {
            siblings = roots;
            startIndex = 0;
            baseDepth = 0;
        }

        for (let i = startIndex; i < segments.length - 1; i += 1) {
            const folderName = segments[i] ?? '';
            const folderPath = segments.slice(0, i + 1).join('/');

            let folder = folderByPath.get(folderPath);

            if (!folder) {
                folder = buildSyntheticFolder(folderPath, folderName, baseDepth + (i - startIndex));
                folderByPath.set(folderPath, folder);
                siblings.push(folder);
            }

            siblings = folder.children;
        }

        const lastDepth = baseDepth + (segments.length - 1 - startIndex);
        const existingFolder = folderByPath.get(file.path);

        if (existingFolder) {
            // Promote a synthetic folder placeholder into the real `FileNode` while
            // keeping any children that were attached to it earlier in the loop.
            Object.assign(existingFolder, file, { children: existingFolder.children, depth: lastDepth });
        } else {
            const node: FileManagerInternalNode = { ...file, children: [], depth: lastDepth };

            siblings.push(node);

            if (file.isDir) {
                folderByPath.set(file.path, node);
            }
        }
    }

    return roots;
};

/**
 * Collect paths of *file* nodes only (no synthetic group roots, no directories).
 * Used as the universe for "select all".
 */
export const collectAllFilePaths = (nodes: FileManagerInternalNode[], result: string[] = []): string[] => {
    for (const node of nodes) {
        if (!node.isGroupRoot && !node.isDir) {
            result.push(node.path);
        }

        if (node.children.length > 0) {
            collectAllFilePaths(node.children, result);
        }
    }

    return result;
};

/** Collect paths of every selectable node (files + real directories), used for bulk-delete validation. */
export const collectAllNodePaths = (nodes: FileManagerInternalNode[], result: string[] = []): string[] => {
    for (const node of nodes) {
        if (!node.isGroupRoot) {
            result.push(node.path);
        }

        if (node.children.length > 0) {
            collectAllNodePaths(node.children, result);
        }
    }

    return result;
};

export const collectVisibleFlat = (
    nodes: FileManagerInternalNode[],
    expandedPaths: Set<string>,
    result: string[] = [],
): string[] => {
    for (const node of nodes) {
        result.push(node.path);

        if (node.isDir && expandedPaths.has(node.path) && node.children.length > 0) {
            collectVisibleFlat(node.children, expandedPaths, result);
        }
    }

    return result;
};

/**
 * Filter the tree so it only contains:
 *   - nodes whose `name` or `path` matches the query,
 *   - all descendants of a matching node (so the user sees what's inside a matching folder),
 *   - all ancestors of any match (so the path is preserved).
 * Synthetic group roots are never matched by name/path; they're kept only if any descendant matches.
 */
export const filterFileManagerTree = (
    nodes: FileManagerInternalNode[],
    query: string,
): FileManagerInternalNode[] => {
    const trimmed = query.trim();

    if (!trimmed) {
        return nodes;
    }

    const lower = trimmed.toLowerCase();

    const filterNode = (
        node: FileManagerInternalNode,
        ancestorMatched: boolean,
    ): FileManagerInternalNode | null => {
        if (node.isGroupRoot) {
            const keptChildren = node.children
                .map((child) => filterNode(child, false))
                .filter((child): child is FileManagerInternalNode => child !== null);

            return keptChildren.length > 0 ? { ...node, children: keptChildren } : null;
        }

        const selfMatched =
            ancestorMatched ||
            node.name.toLowerCase().includes(lower) ||
            node.path.toLowerCase().includes(lower);

        if (!node.isDir) {
            return selfMatched ? node : null;
        }

        const keptChildren = node.children
            .map((child) => filterNode(child, selfMatched))
            .filter((child): child is FileManagerInternalNode => child !== null);

        if (selfMatched) {
            return { ...node, children: keptChildren };
        }

        return keptChildren.length > 0 ? { ...node, children: keptChildren } : null;
    };

    return nodes
        .map((node) => filterNode(node, false))
        .filter((node): node is FileManagerInternalNode => node !== null);
};

/** Collects paths of all directory nodes (including synthetic group roots) for auto-expansion. */
export const collectDirectoryPaths = (nodes: FileManagerInternalNode[], result: string[] = []): string[] => {
    for (const node of nodes) {
        if (node.isDir) {
            result.push(node.path);

            if (node.children.length > 0) {
                collectDirectoryPaths(node.children, result);
            }
        }
    }

    return result;
};

export const resolveSelectionModifier = (
    event: KeyboardEvent | MouseEvent | ReactMouseEvent,
): 'range' | 'single' | 'toggle' => {
    if (event.shiftKey) {
        return 'range';
    }

    if (event.metaKey || event.ctrlKey) {
        return 'toggle';
    }

    return 'single';
};

export const buildFileManagerGridTemplate = (showSize: boolean, showModified: boolean, hasActions: boolean): string => {
    const cols = ['auto', 'minmax(0,1fr)'];

    if (showSize) {
        cols.push('auto');
    }

    if (showModified) {
        cols.push('auto');
    }

    if (hasActions) {
        cols.push('auto');
    }

    return cols.join(' ');
};

/** Recursively locate a node by its absolute path. Returns `undefined` when missing. */
export const findNodeByPath = (
    nodes: FileManagerInternalNode[],
    path: string,
): FileManagerInternalNode | undefined => {
    for (const node of nodes) {
        if (node.path === path) {
            return node;
        }

        if (node.children.length > 0) {
            const found = findNodeByPath(node.children, path);

            if (found) {
                return found;
            }
        }
    }

    return undefined;
};

/**
 * Drop paths whose ancestor is also in the set, so bulk operations don't double-process
 * a directory and its descendants (caller deletes the parent only).
 */
export const dedupeOverlappingPaths = (paths: Iterable<string>): string[] => {
    const sorted = [...paths].sort();
    const result: string[] = [];

    for (const path of sorted) {
        const last = result[result.length - 1];

        if (last && (path === last || path.startsWith(`${last}/`))) {
            continue;
        }

        result.push(path);
    }

    return result;
};
