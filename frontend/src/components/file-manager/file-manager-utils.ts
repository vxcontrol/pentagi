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

        for (let index = startIndex; index < segments.length - 1; index += 1) {
            const folderName = segments[index] ?? '';
            const folderPath = segments.slice(0, index + 1).join('/');

            let folder = folderByPath.get(folderPath);

            if (!folder) {
                folder = buildSyntheticFolder(folderPath, folderName, baseDepth + (index - startIndex));
                folderByPath.set(folderPath, folder);
                siblings.push(folder);
            }

            siblings = folder.children;
        }

        const lastDepth = baseDepth + (segments.length - 1 - startIndex);
        const existingFolder = folderByPath.get(file.path);

        if (existingFolder) {
            // Promote a synthetic folder placeholder into a real `FileNode`, keeping any
            // children attached earlier in the loop. `file` is `FileNode` (no `children`
            // field per the type), so the spread can't clobber the descendants.
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

interface WalkTreeOptions {
    /** Predicate deciding whether to descend into the node's children. Default: always. */
    descend?: (node: FileManagerInternalNode) => boolean;
    /** Predicate deciding whether to include the node in the result. Default: always. */
    include?: (node: FileManagerInternalNode) => boolean;
}

const ALWAYS = (): true => true;

/**
 * Single DFS walker that powers all `collect*` helpers. Pure: returns a freshly
 * allocated array. Internally uses a private accumulator to avoid `concat`
 * allocations on deep trees, but the buffer is never exposed to callers.
 */
export const walkTree = (
    nodes: FileManagerInternalNode[],
    { descend = ALWAYS, include = ALWAYS }: WalkTreeOptions = {},
): string[] => {
    const result: string[] = [];

    const visit = (current: FileManagerInternalNode[]): void => {
        for (const node of current) {
            if (include(node)) {
                result.push(node.path);
            }

            if (node.children.length > 0 && descend(node)) {
                visit(node.children);
            }
        }
    };

    visit(nodes);

    return result;
};

/** Collect paths of *file* nodes only (no synthetic group roots, no directories). */
export const collectAllFilePaths = (nodes: FileManagerInternalNode[]): string[] =>
    walkTree(nodes, { include: (node) => !node.isGroupRoot && !node.isDir });

/**
 * Collect paths of every selectable node (files + real directories), excluding
 * synthetic group roots. Used as the universe for "select all" / bulk operations.
 */
export const collectAllNodePaths = (nodes: FileManagerInternalNode[]): string[] =>
    walkTree(nodes, { include: (node) => !node.isGroupRoot });

export const collectVisibleFlat = (nodes: FileManagerInternalNode[], expandedPaths: Set<string>): string[] =>
    walkTree(nodes, { descend: (node) => node.isDir && expandedPaths.has(node.path) });

/**
 * Filter the tree so it only contains:
 *   - nodes whose `name` or `path` matches the query,
 *   - all descendants of a matching node (so the user sees what's inside a matching folder),
 *   - all ancestors of any match (so the path is preserved).
 * Synthetic group roots are never matched by name/path; they're kept only if any descendant matches.
 */
export const filterFileManagerTree = (nodes: FileManagerInternalNode[], query: string): FileManagerInternalNode[] => {
    const trimmed = query.trim();

    if (!trimmed) {
        return nodes;
    }

    const lower = trimmed.toLowerCase();

    const filterNode = (node: FileManagerInternalNode, ancestorMatched: boolean): FileManagerInternalNode | null => {
        if (node.isGroupRoot) {
            const keptChildren = node.children
                .map((child) => filterNode(child, false))
                .filter((child): child is FileManagerInternalNode => child !== null);

            return keptChildren.length > 0 ? { ...node, children: keptChildren } : null;
        }

        const selfMatched =
            ancestorMatched || node.name.toLowerCase().includes(lower) || node.path.toLowerCase().includes(lower);

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
export const collectDirectoryPaths = (nodes: FileManagerInternalNode[]): string[] =>
    walkTree(nodes, {
        descend: (node) => node.isDir,
        include: (node) => node.isDir,
    });

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
export const findNodeByPath = (nodes: FileManagerInternalNode[], path: string): FileManagerInternalNode | undefined => {
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

/** Clamp `value` into the inclusive `[min, max]` range. */
export const clamp = (min: number, value: number, max: number): number => Math.max(min, Math.min(value, max));

/** Translate `(isAllSelected, isSomeSelected)` into the tri-state value the Checkbox understands. */
export const getCheckboxState = (isAllSelected: boolean, isSomeSelected: boolean): 'indeterminate' | boolean => {
    if (isAllSelected) {
        return true;
    }

    if (isSomeSelected) {
        return 'indeterminate';
    }

    return false;
};

/** Default English pluralization for "N item" / "N items". Override via `labels.pluralizeItems`. */
export const pluralizeItemsEnglish = (count: number): string => `${count} ${count === 1 ? 'item' : 'items'}`;
