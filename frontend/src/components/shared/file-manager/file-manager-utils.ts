import type { MouseEvent as ReactMouseEvent } from 'react';

import {
    differenceInDays,
    differenceInHours,
    differenceInMinutes,
    differenceInMonths,
    differenceInSeconds,
    differenceInWeeks,
    differenceInYears,
    format,
    isThisYear,
    isToday,
    isValid,
} from 'date-fns';
import { enUS } from 'date-fns/locale';

import type {
    FileManagerInternalNode,
    FileManagerRootGroup,
    FileManagerSortState,
    FileNode,
} from './file-manager-types';

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

export const formatModifiedRelative = (modifiedAt?: Date | string): string => {
    if (!modifiedAt) {
        return '';
    }

    const date = typeof modifiedAt === 'string' ? new Date(modifiedAt) : modifiedAt;

    if (!isValid(date)) {
        return '';
    }

    const now = Date.now();
    // All `differenceIn*` helpers are computed up-front so each branch's
    // fallback (`!minutes`, `!hours`, …) can guard against the upper unit
    // rounding down to zero — without it, an edge case like 59 minutes 30s
    // could fall through to the next branch and render as `0h ago`.
    const seconds = differenceInSeconds(now, date);
    const minutes = differenceInMinutes(now, date);
    const hours = differenceInHours(now, date);
    const days = differenceInDays(now, date);
    const weeks = differenceInWeeks(now, date);
    const months = differenceInMonths(now, date);
    const years = differenceInYears(now, date);

    if (seconds < 60 || !minutes) {
        return 'just now';
    }

    if (minutes < 60 || !hours) {
        return `${minutes}m ago`;
    }

    if (hours < 24 || !days) {
        return `${hours}h ago`;
    }

    if (days < 7 || !weeks) {
        return `${days}d ago`;
    }

    if (weeks < 4 || !months) {
        return `${weeks}w ago`;
    }

    if (months < 12 || !years) {
        return `${months}mo ago`;
    }

    return `${years}y ago`;
};

/**
 * Alternative "Modified" formatter that prints an absolute, minute-precision
 * timestamp instead of a relative ("5m ago") label. Used by hosts that let
 * the user opt out of relative dates via a column toggle — see `resources.tsx`.
 *
 * The output is contextual, mirroring `lib/utils/format.ts#formatDate`:
 *   - same day            → `HH:mm` (e.g. `14:32`)
 *   - same calendar year  → `d MMM, HH:mm` (e.g. `15 Apr, 14:32`)
 *   - any other year      → `d MMM yyyy, HH:mm` (e.g. `15 Apr 2024, 14:32`)
 *
 * Locale is forced to `en-US` so day/month tokens stay stable regardless of
 * the user's browser locale (matches the rest of the FileManager UI strings,
 * which are English-only by default).
 */
export const formatModifiedAbsolute = (modifiedAt?: Date | string): string => {
    if (!modifiedAt) {
        return '';
    }

    const date = typeof modifiedAt === 'string' ? new Date(modifiedAt) : modifiedAt;

    if (!isValid(date)) {
        return '';
    }

    if (isToday(date)) {
        return format(date, 'HH:mm');
    }

    if (isThisYear(date)) {
        return format(date, 'd MMM, HH:mm', { locale: enUS });
    }

    return format(date, 'd MMM yyyy, HH:mm', { locale: enUS });
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

/** Locale-aware case-insensitive name comparator. */
const compareNames = (a: FileManagerInternalNode, b: FileManagerInternalNode): number =>
    a.name.localeCompare(b.name, undefined, { numeric: true, sensitivity: 'base' });

/**
 * Compare by `size`. Directories and synthetic group roots have no own size —
 * they are treated as `0` so the comparator stays total. Combined with the
 * default `isFoldersFirst: true`, directories never actually mix with files
 * inside a single comparator call.
 */
const compareSizes = (a: FileManagerInternalNode, b: FileManagerInternalNode): number => {
    const aSize = a.isDir || a.isGroupRoot ? 0 : (a.size ?? 0);
    const bSize = b.isDir || b.isGroupRoot ? 0 : (b.size ?? 0);

    return aSize - bSize;
};

const toTimestamp = (value: Date | string | undefined): number => {
    if (!value) {
        return 0;
    }

    const date = typeof value === 'string' ? new Date(value) : value;
    const time = date.getTime();

    return Number.isNaN(time) ? 0 : time;
};

const compareModified = (a: FileManagerInternalNode, b: FileManagerInternalNode): number =>
    toTimestamp(a.modifiedAt) - toTimestamp(b.modifiedAt);

const SORT_COMPARATORS = {
    modified: compareModified,
    name: compareNames,
    size: compareSizes,
} as const;

/**
 * Recursively sort the tree by the given criterion. Pure: returns freshly
 * cloned arrays so React reference-equality bail-outs still work for
 * memoized consumers; nodes themselves are reused by reference (only the
 * `children` array is rebuilt for nodes whose children actually changed).
 *
 * Sort rules:
 *   - The synthetic group-root level (top level when groups are configured) is
 *     never re-ordered — groups must follow the order specified in `rootGroups`.
 *     Their children are sorted as usual.
 *   - When `isFoldersFirst` is true (default), directories are grouped above
 *     files at every level. This applies even when `sorting` is `null` —
 *     within each group (folders / files) the original insertion order is
 *     preserved (`Array.sort` is stable per the TC39 spec).
 *   - When a `sorting` is supplied, the comparator runs *inside* each group
 *     after the folders-first partition. A stable secondary sort by `name`
 *     breaks ties so the result is deterministic across renders.
 *   - When both `sorting` is `null` AND `isFoldersFirst` is `false`, the input
 *     is returned by reference — zero-cost pass-through for the no-op case.
 */
export const sortFileManagerTree = (
    nodes: readonly FileManagerInternalNode[],
    sorting: FileManagerSortState,
    isFoldersFirst: boolean,
): FileManagerInternalNode[] => {
    if (!sorting && !isFoldersFirst) {
        return nodes as FileManagerInternalNode[];
    }

    const comparator = sorting ? SORT_COMPARATORS[sorting.column] : null;
    const directionMultiplier = sorting?.direction === 'desc' ? -1 : 1;

    const sortLevel = (level: readonly FileManagerInternalNode[], skipReorder: boolean): FileManagerInternalNode[] => {
        const recursed = level.map((node) => {
            if (node.children.length === 0) {
                return node;
            }

            const sortedChildren = sortLevel(node.children, false);

            return sortedChildren === node.children ? node : { ...node, children: sortedChildren };
        });

        if (skipReorder) {
            return recursed;
        }

        const reordered = [...recursed].sort((a, b) => {
            // `isFoldersFirst` is treated as a primary partition before any
            // other criterion so directories always stay above files
            // regardless of the active comparator (or its absence).
            if (isFoldersFirst) {
                const aIsDir = a.isDir || a.isGroupRoot;
                const bIsDir = b.isDir || b.isGroupRoot;

                if (aIsDir !== bIsDir) {
                    return aIsDir ? -1 : 1;
                }
            }

            // No comparator → preserve insertion order within each partition
            // (Array.sort is stable per the TC39 spec since 2019).
            if (!comparator) {
                return 0;
            }

            const primary = comparator(a, b) * directionMultiplier;

            if (primary !== 0) {
                return primary;
            }

            // Stable secondary sort by name so equal-by-criterion rows have a
            // deterministic order across renders. Direction is preserved so
            // `desc` still reads top-down alphabetically reversed.
            return compareNames(a, b) * directionMultiplier;
        });

        return reordered;
    };

    // Top-level group roots keep their declared order. We detect that case by
    // checking whether every root is a group root — if any root is a regular
    // node, the level is sorted normally. Children of group roots are always
    // sorted (the recursion call inside `sortLevel` re-enables sorting via
    // `skipReorder = false`).
    const skipTopReorder = nodes.length > 0 && nodes.every((node) => node.isGroupRoot);

    return sortLevel(nodes, skipTopReorder);
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
    nodes: readonly FileManagerInternalNode[],
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

/**
 * Collect the absolute paths of every selectable node in the subtree rooted at `node`,
 * including `node` itself. Synthetic group roots are skipped (they aren't real files
 * and have no usable path of their own — only their descendants count).
 *
 * Used by directory checkboxes to flip every nested item in one gesture and to
 * derive the parent's tri-state value from the descendants' selection.
 */
export const collectSubtreePaths = (node: FileManagerInternalNode): string[] => {
    const result: string[] = [];

    const visit = (current: FileManagerInternalNode): void => {
        if (!current.isGroupRoot) {
            result.push(current.path);
        }

        for (const child of current.children) {
            visit(child);
        }
    };

    visit(node);

    return result;
};

/** Default English pluralization for "N item" / "N items". Override via `labels.pluralizeItems`. */
export const pluralizeItemsEnglish = (count: number): string => `${count} ${count === 1 ? 'item' : 'items'}`;

/**
 * Recursively sum every file's `size` inside a node's subtree. Synthetic group
 * roots and directories themselves contribute nothing — only their leaf files
 * carry bytes. Files without a `size` are treated as 0 so a partial backend
 * payload doesn't poison the total.
 */
const sumSubtreeBytes = (node: FileManagerInternalNode): number => {
    if (node.children.length > 0) {
        let total = 0;

        for (const child of node.children) {
            total += sumSubtreeBytes(child);
        }

        return total;
    }

    if (node.isDir || node.isGroupRoot) {
        return 0;
    }

    return node.size ?? 0;
};

/**
 * Cumulative byte total for the bulk-bar size summary. Walks `paths` (which the
 * caller is expected to have already deduped via `dedupeOverlappingPaths`) and
 * adds each entry's contribution: a file pulls its own `size`, a directory pulls
 * the recursive sum of its descendants.
 *
 * Missing nodes are skipped silently — selection paths can transiently lag the
 * tree during async refreshes, and counting "0 bytes" for a vanished entry is
 * less surprising than throwing.
 */
export const computeSelectionTotalBytes = (
    tree: readonly FileManagerInternalNode[],
    paths: Iterable<string>,
): number => {
    let total = 0;

    for (const path of paths) {
        const node = findNodeByPath(tree, path);

        if (!node) {
            continue;
        }

        total += sumSubtreeBytes(node);
    }

    return total;
};

/**
 * Returns a freshly cloned Set with every path from `paths` added. Pure —
 * safe to call on a state-owned Set inside a `setState` updater because the
 * input is never mutated. Empty `paths` still produces a fresh clone.
 */
export const addAll = (prev: ReadonlySet<string>, paths: Iterable<string>): Set<string> => {
    const next = new Set(prev);

    for (const path of paths) {
        next.add(path);
    }

    return next;
};

/**
 * Returns a freshly cloned Set with every path from `paths` removed. Same
 * purity contract as `addAll`: input is never mutated, missing paths are
 * silently ignored.
 */
export const removeAll = (prev: ReadonlySet<string>, paths: Iterable<string>): Set<string> => {
    const next = new Set(prev);

    for (const path of paths) {
        next.delete(path);
    }

    return next;
};

/**
 * Returns `true` when every path in `paths` is present in `selected`. Empty
 * `paths` short-circuits to `true` (vacuous truth) — callers that want
 * "non-empty AND all selected" must guard explicitly.
 */
export const isEverySelected = (paths: Iterable<string>, selected: ReadonlySet<string>): boolean => {
    for (const path of paths) {
        if (!selected.has(path)) {
            return false;
        }
    }

    return true;
};

/**
 * Flip the membership of every path in `paths` against `prev` in lock-step:
 * if every path is already selected, the whole batch is removed; otherwise
 * the missing ones are added. Mirrors the directory-checkbox semantics.
 *
 * When `rootPath` is provided (the directory's own path inside `paths`), the
 * "is every path selected?" check is computed over the *descendants only* —
 * i.e. it ignores whether `rootPath` itself is in `prev`. This matches what
 * `computeDirSelectionState` shows on the visible tri-state checkbox, so a
 * single click reliably toggles a folder whose descendants are all selected
 * even though the folder's own path was never added to the selection (which
 * happens whenever the user fills the branch one file at a time, e.g. via
 * the row checkboxes or the header "select all" inside an expanded folder).
 * Without `rootPath`, the legacy strict contract is used (every path,
 * including the dir's own, must be present to qualify as "all selected").
 *
 * Always returns a freshly cloned Set so React's reference-equality bail-out
 * still works for downstream memoized consumers.
 */
export const toggleSubtreeOnSet = (
    prev: ReadonlySet<string>,
    paths: readonly string[],
    rootPath?: string,
): Set<string> => {
    let allSelected: boolean;

    if (rootPath !== undefined && paths.length > 1) {
        allSelected = true;

        for (const p of paths) {
            if (p === rootPath) {
                continue;
            }

            if (!prev.has(p)) {
                allSelected = false;
                break;
            }
        }
    } else {
        allSelected = isEverySelected(paths, prev);
    }

    return allSelected ? removeAll(prev, paths) : addAll(prev, paths);
};

interface ComputeRowClickSelectionArgs {
    /** Last-clicked path used as the range anchor. `null` when nothing has been clicked yet. */
    anchor: null | string;
    /**
     * Subtree map for every directory in the *visible* tree, keyed by the
     * directory's own path. Used to expand directories that appear inside a
     * `range`-modifier slice into their full subtree (folder rows are atomic
     * for the Shift+click gesture too — clicking one selects its branch, so a
     * Shift-range that crosses a folder must do the same; otherwise the
     * folder's checkbox would render unchecked even though its own path is in
     * the selection, because `computeDirSelectionState` derives the state
     * from descendants). Optional — when omitted, range slices stay flat
     * (covers tree-less callers and unit tests that don't care about
     * directory expansion).
     */
    dirSubtreePaths?: ReadonlyMap<string, readonly string[]>;
    /** Visible nodes in DFS order; range modifier resolves anchor/target indices against this list. */
    flatVisible: readonly string[];
    modifier: 'range' | 'single' | 'toggle';
    /** Path of the row that was clicked. */
    path: string;
    /** Current selection. Pure reducer — never mutated. */
    prev: ReadonlySet<string>;
    /**
     * For directory rows: the directory's own path plus every descendant. Single
     * and toggle modifiers operate on the whole branch; range modifier uses it
     * for the fallback payload when the range cannot be resolved AND for the
     * target row inside the slice (via `dirSubtreePaths`, which already
     * contains the same data keyed by every directory's path).
     */
    subtreePaths?: readonly string[];
}

interface ComputeRowClickSelectionResult {
    /** New selection Set (always freshly allocated). */
    next: Set<string>;
    /**
     * Anchor to store for the next interaction. Returns the same `anchor` on
     * a successful range click (so chained Shift+clicks keep extending from the
     * same origin) and on a fallback range click (so the user's intent isn't
     * accidentally overwritten when the original anchor scrolled off-screen).
     */
    nextAnchor: null | string;
}

/**
 * Pure reducer that computes the new selection + anchor for a row-click
 * gesture. Encapsulates the full single / toggle / range matrix so the React
 * hook can stay a thin wrapper that just stores state. Tested directly without
 * a React renderer.
 *
 * Behaviour summary:
 *   - **single**: replaces selection with `subtreePaths` (folders) or `[path]`
 *     (files). Anchor moves to `path`.
 *   - **toggle**: all-or-nothing flip of `subtreePaths` (folders) or single
 *     `path` (files). Anchor moves to `path`.
 *   - **range**: ADDITIVELY unions the slice of `flatVisible` between `anchor`
 *     and `path` (inclusive, direction-agnostic) onto the previous selection
 *     so a Shift-click extends — never erases — what's already selected.
 *     Concretely: a Cmd+click‑based selection set, an explicitly clicked
 *     folder's full subtree, and any earlier range slice are all preserved
 *     when the user reaches for Shift+click to grab "everything from the last
 *     click down to here". Anchor is preserved across chained Shift+clicks so
 *     each one re-extends from the same origin (matches Finder/Explorer for
 *     the anchor; the additive merge is the deliberate divergence the rest of
 *     this codebase relies on — folder rows would otherwise drop their
 *     subtree on every Shift+click). If `anchor` is `null` → behaves like a
 *     `single`-shaped add (no anchor, nothing to extend FROM) and moves the
 *     anchor. If either anchor or target is missing from `flatVisible` →
 *     falls back to the same single-shaped add but preserves the anchor (the
 *     original reference point may still be valid once it scrolls back in).
 */
export const computeRowClickSelection = ({
    anchor,
    dirSubtreePaths,
    flatVisible,
    modifier,
    path,
    prev,
    subtreePaths,
}: ComputeRowClickSelectionArgs): ComputeRowClickSelectionResult => {
    const hasSubtree = !!subtreePaths && subtreePaths.length > 0;
    const buildReplacement = (): Set<string> => (hasSubtree && subtreePaths ? new Set(subtreePaths) : new Set([path]));

    if (modifier === 'single') {
        return { next: buildReplacement(), nextAnchor: path };
    }

    if (modifier === 'toggle') {
        if (hasSubtree && subtreePaths) {
            return { next: toggleSubtreeOnSet(prev, subtreePaths, path), nextAnchor: path };
        }

        const next = new Set(prev);

        if (next.has(path)) {
            next.delete(path);
        } else {
            next.add(path);
        }

        return { next, nextAnchor: path };
    }

    // Range modifier — additive. The slice computed below is unioned onto
    // `prev`, never replaced, so the user can build up a selection across
    // multiple gestures (single → cmd-click → shift-click → shift-click…)
    // without watching earlier picks vanish. The "fallback to a single-shaped
    // add" branches below honour the same contract: each one extends `prev`
    // by exactly one item (the clicked path or its subtree), never wipes it.
    //
    // Directories inside the slice expand to their full subtree via
    // `dirSubtreePaths` so a Shift-range over collapsed folders behaves like
    // a series of plain folder clicks: every folder picks up its children
    // even when those children aren't in `flatVisible` (folder collapsed).
    // This keeps the tri-state checkbox visibly "fully checked" — without
    // the expansion, `computeDirSelectionState` would render a folder
    // unchecked because its descendants stayed out of the selection.
    const expandDir = (p: string): readonly string[] => dirSubtreePaths?.get(p) ?? [p];

    const addRangeToPrev = (paths: Iterable<string>): Set<string> =>
        addAll(
            prev,
            Array.from(paths).flatMap((p) => [...expandDir(p)]),
        );

    const buildAdditiveFallback = (): Set<string> => addRangeToPrev(hasSubtree && subtreePaths ? subtreePaths : [path]);

    if (!anchor) {
        return { next: buildAdditiveFallback(), nextAnchor: path };
    }

    const fromIndex = flatVisible.indexOf(anchor);
    const toIndex = flatVisible.indexOf(path);

    if (fromIndex < 0 || toIndex < 0) {
        return { next: buildAdditiveFallback(), nextAnchor: anchor };
    }

    const [startIndex, endIndex] = fromIndex <= toIndex ? [fromIndex, toIndex] : [toIndex, fromIndex];

    return { next: addRangeToPrev(flatVisible.slice(startIndex, endIndex + 1)), nextAnchor: anchor };
};

interface ComputeToggleSelectionArgs {
    path: string;
    prev: ReadonlySet<string>;
    subtreePaths?: readonly string[];
}

/**
 * Pure reducer behind `onToggleSelection` (row checkbox + Space key):
 *   - When `subtreePaths` is provided and non-empty → all-or-nothing flip of
 *     the whole branch (matches the directory checkbox semantics).
 *   - Otherwise → flip just `path`.
 *
 * The hook layer is responsible for updating the anchor — it always moves to
 * `path` after a toggle, regardless of whether anything was added or removed.
 */
export const computeToggleSelection = ({ path, prev, subtreePaths }: ComputeToggleSelectionArgs): Set<string> => {
    if (subtreePaths && subtreePaths.length > 0) {
        return toggleSubtreeOnSet(prev, subtreePaths, path);
    }

    const next = new Set(prev);

    if (next.has(path)) {
        next.delete(path);
    } else {
        next.add(path);
    }

    return next;
};

interface ComputeToggleSelectAllArgs {
    /** Universe of every selectable path in the current visible tree. */
    allSelectablePaths: readonly string[];
    prev: ReadonlySet<string>;
}

/**
 * Pure reducer behind the header "select all" checkbox and Ctrl/Cmd+A:
 * clears the selection when everything is currently selected, otherwise
 * replaces it with the full universe. Empty universe yields an empty Set.
 */
export const computeToggleSelectAll = ({ allSelectablePaths, prev }: ComputeToggleSelectAllArgs): Set<string> => {
    const isCurrentlyAllSelected = allSelectablePaths.length > 0 && prev.size === allSelectablePaths.length;

    return isCurrentlyAllSelected ? new Set() : new Set(allSelectablePaths);
};

interface ComputeDirSelectionStateArgs {
    /**
     * Path of the directory itself. Excluded from the tri-state count — the
     * checkbox should reflect *content* selection, not whether the directory
     * itself happens to be in `selectedPaths` (which can legitimately be the
     * case after the user toggled descendants off one by one).
     */
    path: string;
    /**
     * Pre-computed list of every selectable path in the directory's subtree —
     * the directory's own path PLUS all descendants. Same array that powers
     * the "toggle whole subtree" gesture; reusing it avoids re-walking the tree.
     */
    paths: readonly string[];
    selectedPaths: ReadonlySet<string>;
}

/**
 * Pure reducer for the directory checkbox tri-state. Treats the *content* as
 * the source of truth:
 *
 *   - **false** → no descendant is selected (regardless of whether the
 *     directory's own path is in `selectedPaths` — the row's hover/selected
 *     background still reflects that via `isSelected` on the row).
 *   - **'indeterminate'** → some, but not all, descendants are selected.
 *   - **true** → every descendant is selected.
 *
 * For an empty directory (no real descendants — e.g. an empty folder or a
 * synthetic group root with no children), the tri-state mirrors the folder's
 * own selection as a binary value, so an empty folder still shows checked
 * when explicitly picked.
 */
export const computeDirSelectionState = ({
    path,
    paths,
    selectedPaths,
}: ComputeDirSelectionStateArgs): 'indeterminate' | boolean => {
    let descendantsTotal = 0;
    let descendantsSelected = 0;

    for (const p of paths) {
        if (p === path) {
            continue;
        }

        descendantsTotal += 1;

        if (selectedPaths.has(p)) {
            descendantsSelected += 1;
        }
    }

    if (descendantsTotal === 0) {
        return selectedPaths.has(path);
    }

    if (descendantsSelected === 0) {
        return false;
    }

    if (descendantsSelected === descendantsTotal) {
        return true;
    }

    return 'indeterminate';
};
