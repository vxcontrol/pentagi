import { subMonths, subYears } from 'date-fns';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

import type { FileManagerInternalNode, FileManagerRootGroup, FileNode } from './file-manager-types';

import {
    addAll,
    buildFileManagerGridTemplate,
    buildFileManagerTree,
    clamp,
    collectAllFilePaths,
    collectAllNodePaths,
    collectDirectoryPaths,
    collectSubtreePaths,
    collectVisibleFlat,
    computeDirSelectionState,
    computeRowClickSelection,
    computeToggleSelectAll,
    computeToggleSelection,
    dedupeOverlappingPaths,
    filterFileManagerTree,
    findNodeByPath,
    formatFileSize,
    formatModifiedAbsolute,
    formatModifiedRelative,
    getCheckboxState,
    isEverySelected,
    normalizeRootGroups,
    pluralizeItemsEnglish,
    removeAll,
    resolveSelectionModifier,
    sortFileManagerTree,
    toggleSubtreeOnSet,
    walkTree,
} from './file-manager-utils';
import { computeNextSort } from './use-file-manager-sorting';

const file = (path: string, overrides: Partial<FileNode> = {}): FileNode => ({
    id: path,
    isDir: false,
    name: path.split('/').pop() ?? path,
    path,
    ...overrides,
});

const dir = (path: string, overrides: Partial<FileNode> = {}): FileNode => file(path, { isDir: true, ...overrides });

describe('formatFileSize', () => {
    it('returns empty string when size is missing', () => {
        expect(formatFileSize(undefined)).toBe('');
    });

    it('renders bytes for sizes < 1KB', () => {
        expect(formatFileSize(0)).toBe('0 B');
        expect(formatFileSize(1023)).toBe('1023 B');
    });

    it('renders 1.0 KB on the 1KB boundary', () => {
        expect(formatFileSize(1024)).toBe('1.0 KB');
    });

    it('uses one decimal under 10 of a unit, none above', () => {
        expect(formatFileSize(1024 * 5)).toBe('5.0 KB');
        expect(formatFileSize(1024 * 99)).toBe('99 KB');
    });

    it('scales up to TB and stops there', () => {
        expect(formatFileSize(1024 ** 4)).toBe('1.0 TB');
        expect(formatFileSize(1024 ** 4 * 5000)).toMatch(/TB$/);
    });
});

describe('formatModifiedRelative', () => {
    const NOW = new Date('2026-04-28T12:00:00Z').getTime();

    beforeEach(() => {
        vi.useFakeTimers();
        vi.setSystemTime(NOW);
    });

    afterEach(() => {
        vi.useRealTimers();
    });

    it('returns empty for missing/invalid input', () => {
        expect(formatModifiedRelative(undefined)).toBe('');
        expect(formatModifiedRelative('not-a-date')).toBe('');
    });

    it('handles "just now"', () => {
        expect(formatModifiedRelative(new Date(NOW - 10_000))).toBe('just now');
    });

    it('formats minutes / hours / days within a week', () => {
        expect(formatModifiedRelative(new Date(NOW - 5 * 60_000))).toBe('5m ago');
        expect(formatModifiedRelative(new Date(NOW - 3 * 60 * 60_000))).toBe('3h ago');
        expect(formatModifiedRelative(new Date(NOW - 4 * 24 * 60 * 60_000))).toBe('4d ago');
    });

    it('formats weeks for entries within 4 weeks', () => {
        expect(formatModifiedRelative(new Date(NOW - 14 * 24 * 60 * 60_000))).toBe('2w ago');
        expect(formatModifiedRelative(new Date(NOW - 21 * 24 * 60 * 60_000))).toBe('3w ago');
    });

    it('formats months for entries within a year', () => {
        // 90 days ≈ 2 calendar months from the fake `NOW`. Using subMonths
        // instead of arithmetic keeps the test robust against month-length
        // variation (date-fns' differenceInMonths is calendar-based, not 30d).
        expect(formatModifiedRelative(subMonths(new Date(NOW), 2))).toBe('2mo ago');
        expect(formatModifiedRelative(subMonths(new Date(NOW), 6))).toBe('6mo ago');
    });

    it('formats years for entries older than a year', () => {
        expect(formatModifiedRelative(subYears(new Date(NOW), 1))).toBe('1y ago');
        expect(formatModifiedRelative(subYears(new Date(NOW), 5))).toBe('5y ago');
    });

    it('accepts ISO strings', () => {
        const iso = new Date(NOW - 10 * 60_000).toISOString();

        expect(formatModifiedRelative(iso)).toBe('10m ago');
    });
});

describe('formatModifiedAbsolute', () => {
    // Use local-time constructors throughout: `format()` and `isToday` /
    // `isThisYear` resolve in the host TZ, so anchoring the fake clock to
    // a UTC literal would let the rendered hours / "today" boundary drift
    // depending on where the test runs (CI's TZ is not guaranteed).
    const NOW = new Date(2026, 3, 28, 14, 30);

    beforeEach(() => {
        vi.useFakeTimers();
        vi.setSystemTime(NOW);
    });

    afterEach(() => {
        vi.useRealTimers();
    });

    it('returns empty for missing/invalid input', () => {
        expect(formatModifiedAbsolute(undefined)).toBe('');
        expect(formatModifiedAbsolute('not-a-date')).toBe('');
    });

    it('renders only the time for today', () => {
        expect(formatModifiedAbsolute(new Date(2026, 3, 28, 9, 5))).toBe('09:05');
        expect(formatModifiedAbsolute(new Date(2026, 3, 28, 23, 59))).toBe('23:59');
    });

    it('renders day, month and time within the current calendar year', () => {
        expect(formatModifiedAbsolute(new Date(2026, 0, 3, 8, 7))).toBe('3 Jan, 08:07');
        expect(formatModifiedAbsolute(new Date(2026, 11, 15, 18, 0))).toBe('15 Dec, 18:00');
    });

    it('includes the year for entries from previous calendar years', () => {
        expect(formatModifiedAbsolute(new Date(2024, 5, 1, 6, 30))).toBe('1 Jun 2024, 06:30');
        expect(formatModifiedAbsolute(new Date(2020, 11, 31, 23, 45))).toBe('31 Dec 2020, 23:45');
    });

    it('accepts ISO strings', () => {
        // Build the ISO from the same local-time `Date` so the render is stable
        // across host time zones — toISOString() shifts to UTC, but formatting
        // back through the local clock returns the original wall-clock time.
        const local = new Date(2026, 2, 14, 10, 5);

        expect(formatModifiedAbsolute(local.toISOString())).toBe('14 Mar, 10:05');
    });
});

describe('normalizeRootGroups', () => {
    it('returns input reference when nothing needs trimming', () => {
        const groups: FileManagerRootGroup[] = [{ id: 'a', label: 'A', pathPrefix: 'uploads' }];

        expect(normalizeRootGroups(groups)).toBe(groups);
    });

    it('trims trailing slashes from pathPrefix', () => {
        const groups: FileManagerRootGroup[] = [
            { id: 'a', label: 'A', pathPrefix: 'uploads/' },
            { id: 'b', label: 'B', pathPrefix: 'container///' },
        ];
        const normalized = normalizeRootGroups(groups);

        expect(normalized).not.toBe(groups);
        expect(normalized?.map((group) => group.pathPrefix)).toEqual(['uploads', 'container']);
    });

    it('passes through empty/undefined', () => {
        expect(normalizeRootGroups(undefined)).toBeUndefined();
        expect(normalizeRootGroups([])).toEqual([]);
    });
});

describe('buildFileManagerTree', () => {
    it('builds a flat tree without groups', () => {
        const tree = buildFileManagerTree([file('a.txt'), file('b.txt')]);

        expect(tree.map((node) => node.path)).toEqual(['a.txt', 'b.txt']);
        expect(tree.every((node) => !node.isGroupRoot)).toBe(true);
    });

    it('creates intermediate folders for nested files', () => {
        const tree = buildFileManagerTree([file('foo/bar/baz.txt')]);

        expect(tree).toHaveLength(1);
        expect(tree[0]!.path).toBe('foo');
        expect(tree[0]!.isDir).toBe(true);
        expect(tree[0]!.children[0]!.path).toBe('foo/bar');
        expect(tree[0]!.children[0]!.children[0]!.path).toBe('foo/bar/baz.txt');
    });

    it('places files into matching root groups', () => {
        const groups: FileManagerRootGroup[] = [
            { defaultOpen: true, id: 'uploads', label: 'Uploads', pathPrefix: 'uploads' },
            { defaultOpen: true, id: 'container', label: 'Container', pathPrefix: 'container' },
        ];
        const tree = buildFileManagerTree(
            [file('uploads/a.txt'), file('container/sub/b.txt'), file('orphan.txt')],
            groups,
        );

        expect(tree).toHaveLength(2);
        expect(tree[0]!.isGroupRoot).toBe(true);
        expect(tree[0]!.children.map((child) => child.path)).toEqual(['uploads/a.txt']);
        expect(tree[1]!.children[0]!.path).toBe('container/sub');
        expect(tree[1]!.children[0]!.children[0]!.path).toBe('container/sub/b.txt');
    });

    it('promotes a real directory entry over a synthetic placeholder while keeping descendants', () => {
        const tree = buildFileManagerTree([
            file('foo/bar.txt'),
            dir('foo', { id: 'real-foo', modifiedAt: '2026-04-01T00:00:00Z' }),
        ]);

        expect(tree).toHaveLength(1);
        expect(tree[0]!.id).toBe('real-foo');
        expect(tree[0]!.modifiedAt).toBe('2026-04-01T00:00:00Z');
        expect(tree[0]!.children).toHaveLength(1);
        expect(tree[0]!.children[0]!.path).toBe('foo/bar.txt');
    });

    it('tolerates trailing slashes in group pathPrefix', () => {
        const tree = buildFileManagerTree(
            [file('uploads/a.txt')],
            [{ id: 'u', label: 'Uploads', pathPrefix: 'uploads/' }],
        );

        expect(tree[0]!.children[0]!.path).toBe('uploads/a.txt');
    });

    it('skips files outside any defined group', () => {
        const tree = buildFileManagerTree([file('orphan.txt')], [{ id: 'u', label: 'Uploads', pathPrefix: 'uploads' }]);

        expect(tree[0]!.children).toEqual([]);
    });
});

describe('collectAllFilePaths', () => {
    it('returns only file paths (no directories, no group roots)', () => {
        const tree = buildFileManagerTree(
            [file('a/b.txt'), file('a/c.txt'), file('d/e.txt')],
            [
                { id: 'a', label: 'A', pathPrefix: 'a' },
                { id: 'd', label: 'D', pathPrefix: 'd' },
            ],
        );

        expect(collectAllFilePaths(tree).sort()).toEqual(['a/b.txt', 'a/c.txt', 'd/e.txt']);
    });

    it('omits synthetic intermediate directories from the result', () => {
        const tree = buildFileManagerTree([file('foo/bar/baz.txt')]);

        expect(collectAllFilePaths(tree)).toEqual(['foo/bar/baz.txt']);
    });
});

describe('collectVisibleFlat', () => {
    it('walks only expanded directories', () => {
        const tree = buildFileManagerTree([file('a/b.txt'), file('c.txt')]);
        const expanded = new Set<string>();

        expect(collectVisibleFlat(tree, expanded)).toEqual(['a', 'c.txt']);
        expanded.add('a');
        expect(collectVisibleFlat(tree, expanded)).toEqual(['a', 'a/b.txt', 'c.txt']);
    });
});

describe('filterFileManagerTree', () => {
    const tree: FileManagerInternalNode[] = buildFileManagerTree([
        file('foo/bar.txt'),
        file('foo/baz.md'),
        file('readme.md'),
    ]);

    it('returns input unchanged when query is empty', () => {
        expect(filterFileManagerTree(tree, '')).toBe(tree);
    });

    it('keeps matched files and their ancestors', () => {
        const filtered = filterFileManagerTree(tree, 'bar');

        expect(filtered).toHaveLength(1);
        expect(filtered[0]!.path).toBe('foo');
        expect(filtered[0]!.children.map((child) => child.path)).toEqual(['foo/bar.txt']);
    });

    it('keeps all descendants of a matched directory', () => {
        const filtered = filterFileManagerTree(tree, 'foo');

        expect(filtered[0]!.children.map((child) => child.path).sort()).toEqual(['foo/bar.txt', 'foo/baz.md']);
    });

    it('matches case-insensitively against name and path', () => {
        expect(filterFileManagerTree(tree, 'README')).toHaveLength(1);
        expect(filterFileManagerTree(tree, 'README')[0]!.path).toBe('readme.md');
    });
});

describe('collectDirectoryPaths', () => {
    it('returns directory paths in DFS order', () => {
        const tree = buildFileManagerTree([file('a/b/c.txt'), file('d.txt')]);

        expect(collectDirectoryPaths(tree)).toEqual(['a', 'a/b']);
    });
});

describe('resolveSelectionModifier', () => {
    const makeEvent = (init: { ctrlKey?: boolean; metaKey?: boolean; shiftKey?: boolean }) =>
        ({
            ctrlKey: false,
            metaKey: false,
            shiftKey: false,
            ...init,
        }) as unknown as MouseEvent;

    it('detects shift as range', () => {
        expect(resolveSelectionModifier(makeEvent({ shiftKey: true }))).toBe('range');
    });

    it('detects meta/ctrl as toggle', () => {
        expect(resolveSelectionModifier(makeEvent({ metaKey: true }))).toBe('toggle');
        expect(resolveSelectionModifier(makeEvent({ ctrlKey: true }))).toBe('toggle');
    });

    it('falls back to single', () => {
        expect(resolveSelectionModifier(makeEvent({}))).toBe('single');
    });
});

describe('buildFileManagerGridTemplate', () => {
    it('always includes checkbox column and name column', () => {
        expect(buildFileManagerGridTemplate(false, false, false)).toBe('auto minmax(0,1fr)');
    });

    it('appends optional columns in order', () => {
        expect(buildFileManagerGridTemplate(true, true, true)).toBe('auto minmax(0,1fr) auto auto auto');
    });
});

describe('findNodeByPath', () => {
    const tree = buildFileManagerTree([file('foo/bar.txt')]);

    it('finds nested nodes', () => {
        expect(findNodeByPath(tree, 'foo/bar.txt')?.path).toBe('foo/bar.txt');
    });

    it('returns undefined for missing paths', () => {
        expect(findNodeByPath(tree, 'nope')).toBeUndefined();
    });
});

describe('dedupeOverlappingPaths', () => {
    it('drops descendants when an ancestor is selected', () => {
        const result = dedupeOverlappingPaths(['foo/bar', 'foo', 'foo/baz/qux', 'other']);

        expect(result.sort()).toEqual(['foo', 'other']);
    });

    it('preserves non-overlapping siblings', () => {
        expect(dedupeOverlappingPaths(['a/b', 'a/c']).sort()).toEqual(['a/b', 'a/c']);
    });

    it('handles duplicates', () => {
        expect(dedupeOverlappingPaths(['foo', 'foo'])).toEqual(['foo']);
    });

    it('does not match by prefix-only (foobar is not a child of foo)', () => {
        expect(dedupeOverlappingPaths(['foo', 'foobar']).sort()).toEqual(['foo', 'foobar']);
    });
});

describe('walkTree', () => {
    const tree = buildFileManagerTree([file('a/b/c.txt'), file('a/d.txt'), file('e.txt')]);

    it('walks every node by default', () => {
        expect(walkTree(tree).sort()).toEqual(['a', 'a/b', 'a/b/c.txt', 'a/d.txt', 'e.txt'].sort());
    });

    it('stops descending when `descend` returns false', () => {
        expect(walkTree(tree, { descend: (node) => !(node.path === 'a/b') }).sort()).toEqual(
            ['a', 'a/b', 'a/d.txt', 'e.txt'].sort(),
        );
    });

    it('skips nodes when `include` returns false', () => {
        expect(walkTree(tree, { include: (node) => !node.isDir }).sort()).toEqual(
            ['a/b/c.txt', 'a/d.txt', 'e.txt'].sort(),
        );
    });

    it('powers all collect* helpers consistently', () => {
        expect(collectAllFilePaths(tree).sort()).toEqual(['a/b/c.txt', 'a/d.txt', 'e.txt']);
        expect(collectAllNodePaths(tree).sort()).toEqual(['a', 'a/b', 'a/b/c.txt', 'a/d.txt', 'e.txt'].sort());
        expect(collectDirectoryPaths(tree).sort()).toEqual(['a', 'a/b']);
    });
});

describe('clamp', () => {
    it('returns value when in range', () => {
        expect(clamp(0, 5, 10)).toBe(5);
    });

    it('clamps to min', () => {
        expect(clamp(0, -3, 10)).toBe(0);
    });

    it('clamps to max', () => {
        expect(clamp(0, 99, 10)).toBe(10);
    });

    it('returns min when min > max (degenerate input)', () => {
        // `Math.min(value, max)` runs first, then `Math.max(min, ...)` wins.
        expect(clamp(10, 5, 0)).toBe(10);
    });
});

describe('getCheckboxState', () => {
    it('returns true when all selected', () => {
        expect(getCheckboxState(true, false)).toBe(true);
    });

    it('returns indeterminate when some but not all selected', () => {
        expect(getCheckboxState(false, true)).toBe('indeterminate');
    });

    it('returns false when nothing selected', () => {
        expect(getCheckboxState(false, false)).toBe(false);
    });

    it('prefers all-selected over some-selected if both somehow set', () => {
        expect(getCheckboxState(true, true)).toBe(true);
    });
});

describe('collectSubtreePaths', () => {
    it('returns the directory itself plus every descendant', () => {
        const tree = buildFileManagerTree([file('a/b/c.txt'), file('a/d.txt')]);
        const root = tree[0]!; // 'a'

        expect(collectSubtreePaths(root).sort()).toEqual(['a', 'a/b', 'a/b/c.txt', 'a/d.txt'].sort());
    });

    it('returns just the leaf path for a single file node', () => {
        const tree = buildFileManagerTree([file('lonely.txt')]);

        expect(collectSubtreePaths(tree[0]!)).toEqual(['lonely.txt']);
    });

    it('omits synthetic group roots but keeps their descendants', () => {
        const tree = buildFileManagerTree(
            [file('uploads/a.txt'), file('uploads/sub/b.txt')],
            [{ id: 'u', label: 'Uploads', pathPrefix: 'uploads' }],
        );
        const groupRoot = tree[0]!;

        expect(groupRoot.isGroupRoot).toBe(true);
        expect(collectSubtreePaths(groupRoot).sort()).toEqual(['uploads/a.txt', 'uploads/sub', 'uploads/sub/b.txt']);
    });
});

describe('pluralizeItemsEnglish', () => {
    it('uses singular for 1', () => {
        expect(pluralizeItemsEnglish(1)).toBe('1 item');
    });

    it('uses plural for 0 and >1', () => {
        expect(pluralizeItemsEnglish(0)).toBe('0 items');
        expect(pluralizeItemsEnglish(5)).toBe('5 items');
    });
});

describe('addAll', () => {
    it('returns a new set containing every path from prev plus the iterable', () => {
        const prev = new Set(['a']);
        const next = addAll(prev, ['b', 'c']);

        expect([...next].sort()).toEqual(['a', 'b', 'c']);
    });

    it('does not mutate the input set', () => {
        const prev = new Set(['a']);
        const next = addAll(prev, ['b', 'c']);

        expect(next).not.toBe(prev);
        expect([...prev]).toEqual(['a']);
    });

    it('returns a freshly cloned set even for an empty iterable', () => {
        const prev = new Set(['a']);
        const next = addAll(prev, []);

        expect(next).not.toBe(prev);
        expect([...next]).toEqual(['a']);
    });

    it('accepts another Set as the source', () => {
        const next = addAll(new Set<string>(), new Set(['x', 'y']));

        expect([...next].sort()).toEqual(['x', 'y']);
    });
});

describe('removeAll', () => {
    it('returns a new set without the requested paths and ignores missing ones', () => {
        const prev = new Set(['a', 'b', 'c']);
        const next = removeAll(prev, ['b', 'missing']);

        expect([...next].sort()).toEqual(['a', 'c']);
    });

    it('does not mutate the input set', () => {
        const prev = new Set(['a', 'b']);
        const next = removeAll(prev, ['b']);

        expect(next).not.toBe(prev);
        expect([...prev].sort()).toEqual(['a', 'b']);
    });

    it('returns a freshly cloned set even for an empty iterable', () => {
        const prev = new Set(['a']);
        const next = removeAll(prev, []);

        expect(next).not.toBe(prev);
        expect([...next]).toEqual(['a']);
    });
});

describe('isEverySelected', () => {
    it('returns true when every path is in the selected set', () => {
        expect(isEverySelected(['a', 'b'], new Set(['a', 'b', 'c']))).toBe(true);
    });

    it('returns false when at least one path is missing', () => {
        expect(isEverySelected(['a', 'b'], new Set(['a']))).toBe(false);
    });

    it('returns true (vacuously) for an empty iterable', () => {
        expect(isEverySelected([], new Set())).toBe(true);
        expect(isEverySelected([], new Set(['x']))).toBe(true);
    });
});

describe('toggleSubtreeOnSet', () => {
    it('adds every path when the subtree is fully unselected', () => {
        expect([...toggleSubtreeOnSet(new Set(), ['a', 'a/b', 'a/c'])].sort()).toEqual(['a', 'a/b', 'a/c']);
    });

    it('adds the missing paths when the subtree is partially selected', () => {
        expect([...toggleSubtreeOnSet(new Set(['a/b']), ['a', 'a/b', 'a/c'])].sort()).toEqual(['a', 'a/b', 'a/c']);
    });

    it('removes every path when the subtree is fully selected', () => {
        expect([...toggleSubtreeOnSet(new Set(['a', 'a/b', 'a/c']), ['a', 'a/b', 'a/c'])]).toEqual([]);
    });

    it('preserves unrelated entries on both add and remove', () => {
        expect([...toggleSubtreeOnSet(new Set(['x']), ['a', 'a/b'])].sort()).toEqual(['a', 'a/b', 'x']);
        expect([...toggleSubtreeOnSet(new Set(['a', 'a/b', 'x']), ['a', 'a/b'])]).toEqual(['x']);
    });

    it('returns a freshly cloned Set (does not mutate input)', () => {
        const prev = new Set(['x']);
        const next = toggleSubtreeOnSet(prev, ['a']);

        expect(next).not.toBe(prev);
        expect([...prev]).toEqual(['x']);
    });

    it('REGRESSION: with `rootPath`, clears the branch when every descendant is selected even though the dir itself is missing from prev', () => {
        // User scenario: inside an expanded folder, select every file via the
        // row checkboxes one by one (or via the header "select all" while the
        // folder is the only visible group). The folder's own path is NEVER
        // added to `selectedPaths` — the tri-state still renders "checked"
        // because `computeDirSelectionState` only counts descendants. A single
        // click on the folder's checkbox must clear the branch; without
        // `rootPath` the strict `isEverySelected` check would treat the dir's
        // missing path as "not all selected" and ADD it instead, requiring a
        // second click to actually deselect.
        expect([...toggleSubtreeOnSet(new Set(['dir/a', 'dir/b']), ['dir', 'dir/a', 'dir/b'], 'dir')]).toEqual([]);
    });

    it('with `rootPath`, still adds the missing pieces when only some descendants are selected', () => {
        expect([...toggleSubtreeOnSet(new Set(['dir/a']), ['dir', 'dir/a', 'dir/b'], 'dir')].sort()).toEqual([
            'dir',
            'dir/a',
            'dir/b',
        ]);
    });

    it('with `rootPath`, treats a fully selected branch (dir + descendants) the same as descendants-only', () => {
        // Both shapes ("dir + every descendant" and "every descendant, no dir")
        // render the checkbox as fully checked, so a single click must clear
        // the entire branch in either case.
        expect([...toggleSubtreeOnSet(new Set(['dir', 'dir/a', 'dir/b']), ['dir', 'dir/a', 'dir/b'], 'dir')]).toEqual(
            [],
        );
    });

    it('with `rootPath` on an empty folder (paths === [folder]), falls back to a simple binary toggle of the folder itself', () => {
        // Empty folder: `paths.length === 1`, the descendant-only short-circuit
        // would treat "no descendants" as vacuously all-selected, which would
        // make every click on an unselected empty folder remove its own path
        // (a no-op). Falling through to the strict path keeps the binary
        // semantics that match `computeDirSelectionState` for empty folders.
        expect([...toggleSubtreeOnSet(new Set(), ['empty'], 'empty')]).toEqual(['empty']);
        expect([...toggleSubtreeOnSet(new Set(['empty']), ['empty'], 'empty')]).toEqual([]);
    });

    it('with `rootPath`, preserves unrelated entries on both add and remove', () => {
        expect([...toggleSubtreeOnSet(new Set(['dir/a', 'x']), ['dir', 'dir/a', 'dir/b'], 'dir')].sort()).toEqual([
            'dir',
            'dir/a',
            'dir/b',
            'x',
        ]);
        expect([...toggleSubtreeOnSet(new Set(['dir/a', 'dir/b', 'x']), ['dir', 'dir/a', 'dir/b'], 'dir')]).toEqual([
            'x',
        ]);
    });
});

describe('computeRowClickSelection — single modifier', () => {
    it('replaces selection with [path] for a file row', () => {
        const result = computeRowClickSelection({
            anchor: null,
            flatVisible: ['a', 'b', 'c'],
            modifier: 'single',
            path: 'b',
            prev: new Set(['a', 'c']),
        });

        expect([...result.next]).toEqual(['b']);
        expect(result.nextAnchor).toBe('b');
    });

    it('replaces selection with the whole subtree for a directory row', () => {
        const result = computeRowClickSelection({
            anchor: null,
            flatVisible: ['dir', 'dir/a', 'dir/b'],
            modifier: 'single',
            path: 'dir',
            prev: new Set(['unrelated']),
            subtreePaths: ['dir', 'dir/a', 'dir/b'],
        });

        expect([...result.next].sort()).toEqual(['dir', 'dir/a', 'dir/b']);
        expect(result.nextAnchor).toBe('dir');
    });

    it('treats an empty subtreePaths array as "no subtree" (defensive — never happens for real folders)', () => {
        const result = computeRowClickSelection({
            anchor: null,
            flatVisible: ['p'],
            modifier: 'single',
            path: 'p',
            prev: new Set(),
            subtreePaths: [],
        });

        expect([...result.next]).toEqual(['p']);
        expect(result.nextAnchor).toBe('p');
    });
});

describe('computeRowClickSelection — toggle modifier', () => {
    it('flips a file path on when not selected', () => {
        const result = computeRowClickSelection({
            anchor: 'a',
            flatVisible: ['a', 'b', 'c'],
            modifier: 'toggle',
            path: 'b',
            prev: new Set(['a']),
        });

        expect([...result.next].sort()).toEqual(['a', 'b']);
        expect(result.nextAnchor).toBe('b');
    });

    it('flips a file path off when already selected', () => {
        const result = computeRowClickSelection({
            anchor: 'a',
            flatVisible: ['a', 'b', 'c'],
            modifier: 'toggle',
            path: 'b',
            prev: new Set(['a', 'b']),
        });

        expect([...result.next]).toEqual(['a']);
        expect(result.nextAnchor).toBe('b');
    });

    it('all-or-nothing adds the whole branch for a partially selected directory', () => {
        const result = computeRowClickSelection({
            anchor: null,
            flatVisible: ['dir', 'dir/a', 'dir/b'],
            modifier: 'toggle',
            path: 'dir',
            prev: new Set(['dir/a', 'unrelated']),
            subtreePaths: ['dir', 'dir/a', 'dir/b'],
        });

        expect([...result.next].sort()).toEqual(['dir', 'dir/a', 'dir/b', 'unrelated']);
        expect(result.nextAnchor).toBe('dir');
    });

    it('all-or-nothing strips the whole branch for a fully selected directory', () => {
        const result = computeRowClickSelection({
            anchor: null,
            flatVisible: ['dir', 'dir/a', 'dir/b'],
            modifier: 'toggle',
            path: 'dir',
            prev: new Set(['dir', 'dir/a', 'dir/b', 'unrelated']),
            subtreePaths: ['dir', 'dir/a', 'dir/b'],
        });

        expect([...result.next]).toEqual(['unrelated']);
        expect(result.nextAnchor).toBe('dir');
    });

    it('REGRESSION: cmd-click on a folder whose descendants are all selected (dir itself missing) clears the branch in one gesture', () => {
        // Mirror of the checkbox-toggle regression — `Cmd`/`Ctrl`+click on a
        // folder row goes through the same `toggleSubtreeOnSet` path. If the
        // user single-clicked each child to fill the branch (so `dir`'s own
        // path was never added), a follow-up cmd-click on the folder row must
        // strip everything in one gesture, not silently add `dir` and require
        // a second click.
        const result = computeRowClickSelection({
            anchor: null,
            flatVisible: ['dir', 'dir/a', 'dir/b'],
            modifier: 'toggle',
            path: 'dir',
            prev: new Set(['dir/a', 'dir/b', 'unrelated']),
            subtreePaths: ['dir', 'dir/a', 'dir/b'],
        });

        expect([...result.next]).toEqual(['unrelated']);
        expect(result.nextAnchor).toBe('dir');
    });
});

describe('computeRowClickSelection — range modifier', () => {
    const flatVisible = ['a', 'b', 'c', 'd', 'e'];

    it('selects from anchor to target (forward) and unions onto prev', () => {
        const result = computeRowClickSelection({
            anchor: 'b',
            flatVisible,
            modifier: 'range',
            path: 'd',
            prev: new Set(['a', 'b']),
        });

        expect([...result.next].sort()).toEqual(['a', 'b', 'c', 'd']);
        expect(result.nextAnchor).toBe('b');
    });

    it('selects from anchor to target (backward) and unions onto prev', () => {
        const result = computeRowClickSelection({
            anchor: 'd',
            flatVisible,
            modifier: 'range',
            path: 'b',
            prev: new Set(['d']),
        });

        expect([...result.next].sort()).toEqual(['b', 'c', 'd']);
        expect(result.nextAnchor).toBe('d');
    });

    it('extends — never erases — earlier picks that fall outside the range', () => {
        // Regression: a Cmd-click added `a` to the selection at anchor `c`,
        // then Shift+click on `e` must keep `a` AROUND. Replacing `prev` with
        // just the visible slice would silently drop the user's first pick.
        const result = computeRowClickSelection({
            anchor: 'c',
            flatVisible,
            modifier: 'range',
            path: 'e',
            prev: new Set(['a', 'c']),
        });

        expect([...result.next].sort()).toEqual(['a', 'c', 'd', 'e']);
        expect(result.nextAnchor).toBe('c');
    });

    it('reduces to "just add the anchor row" when anchor === target', () => {
        const result = computeRowClickSelection({
            anchor: 'c',
            flatVisible,
            modifier: 'range',
            path: 'c',
            prev: new Set(),
        });

        expect([...result.next]).toEqual(['c']);
        expect(result.nextAnchor).toBe('c');
    });

    it('does not shrink the selection when the new range is a subset of prev', () => {
        // After a wide Shift+click `a..e` (sel={a,b,c,d,e}), a user who Shift-
        // clicks `c` is asking to "still cover c"; additive semantics keep the
        // wider selection so the gesture never feels destructive.
        const result = computeRowClickSelection({
            anchor: 'a',
            flatVisible,
            modifier: 'range',
            path: 'c',
            prev: new Set(['a', 'b', 'c', 'd', 'e']),
        });

        expect([...result.next].sort()).toEqual(['a', 'b', 'c', 'd', 'e']);
        expect(result.nextAnchor).toBe('a');
    });

    it('falls back to a single-row ADD when no anchor is set yet (and moves the anchor)', () => {
        const result = computeRowClickSelection({
            anchor: null,
            flatVisible,
            modifier: 'range',
            path: 'c',
            prev: new Set(['a']),
        });

        expect([...result.next].sort()).toEqual(['a', 'c']);
        expect(result.nextAnchor).toBe('c');
    });

    it('falls back to a single-row ADD when the anchor is no longer visible (and PRESERVES the anchor)', () => {
        // Scenario: user cmd-clicked `dir/a`, then collapsed `dir`. The anchor
        // is now off-screen but should not be silently overwritten — a follow-up
        // shift-click after re-expanding the dir should still extend from there.
        const result = computeRowClickSelection({
            anchor: 'dir/a',
            flatVisible: ['dir', 'other'],
            modifier: 'range',
            path: 'other',
            prev: new Set(['dir/a']),
        });

        expect([...result.next].sort()).toEqual(['dir/a', 'other']);
        expect(result.nextAnchor).toBe('dir/a');
    });

    it('falls back to a single-row ADD when the target is not in flatVisible (defensive)', () => {
        const result = computeRowClickSelection({
            anchor: 'a',
            flatVisible,
            modifier: 'range',
            path: 'ghost',
            prev: new Set(['a']),
        });

        expect([...result.next].sort()).toEqual(['a', 'ghost']);
        expect(result.nextAnchor).toBe('a');
    });

    it('range click that lands on a directory expands it via dirSubtreePaths', () => {
        // The slice walks `flatVisible`; when one of the items is a directory,
        // `dirSubtreePaths` is consulted so the folder's full branch joins the
        // selection (matching the contract of a plain folder click). Without
        // this, the folder's tri-state checkbox would render unchecked because
        // its descendants stayed out of the selection.
        const result = computeRowClickSelection({
            anchor: 'a',
            dirSubtreePaths: new Map<string, readonly string[]>([['dir', ['dir', 'dir/x', 'dir/y']]]),
            flatVisible: ['a', 'b', 'dir', 'e'],
            modifier: 'range',
            path: 'dir',
            prev: new Set(['a']),
            subtreePaths: ['dir', 'dir/x', 'dir/y'],
        });

        expect([...result.next].sort()).toEqual(['a', 'b', 'dir', 'dir/x', 'dir/y']);
        expect(result.nextAnchor).toBe('a');
    });

    it('range click after a folder click keeps the folder subtree alive', () => {
        // Regression: clicking a collapsed folder selects {dir, dir/x, dir/y},
        // a follow-up Shift+click on a sibling file must keep the subtree (the
        // pre-additive behaviour wiped it because the descendants weren't in
        // `flatVisible`).
        const result = computeRowClickSelection({
            anchor: 'dir',
            flatVisible: ['dir', 'sib'],
            modifier: 'range',
            path: 'sib',
            prev: new Set(['dir', 'dir/x', 'dir/y']),
        });

        expect([...result.next].sort()).toEqual(['dir', 'dir/x', 'dir/y', 'sib']);
        expect(result.nextAnchor).toBe('dir');
    });

    it('range click across three sibling folders selects every folder + its subtree', () => {
        // User scenario: three collapsed folders side-by-side. Click `f1`,
        // Shift+click `f3` — every folder must end up "fully checked", which
        // requires every descendant to be in the selection. Without
        // `dirSubtreePaths` the slice would only contain the folder paths
        // themselves and the tri-state checkbox would render `false` for f2
        // and f3 because their descendants are missing from `selectedPaths`.
        const dirSubtreePaths = new Map<string, readonly string[]>([
            ['f1', ['f1', 'f1/a', 'f1/b']],
            ['f2', ['f2', 'f2/a', 'f2/b']],
            ['f3', ['f3', 'f3/a', 'f3/b']],
        ]);
        const result = computeRowClickSelection({
            anchor: 'f1',
            dirSubtreePaths,
            flatVisible: ['f1', 'f2', 'f3'],
            modifier: 'range',
            path: 'f3',
            prev: new Set(['f1', 'f1/a', 'f1/b']),
            subtreePaths: ['f3', 'f3/a', 'f3/b'],
        });

        expect([...result.next].sort()).toEqual(['f1', 'f1/a', 'f1/b', 'f2', 'f2/a', 'f2/b', 'f3', 'f3/a', 'f3/b']);
        expect(result.nextAnchor).toBe('f1');
    });
});

describe('computeRowClickSelection — anchor stability across chained clicks', () => {
    const flatVisible = ['a', 'b', 'c', 'd', 'e'];

    it('chained shift-click extends from the same anchor (anchor is preserved)', () => {
        // Step 1: single click on 'b' → anchor = 'b'
        const step1 = computeRowClickSelection({
            anchor: null,
            flatVisible,
            modifier: 'single',
            path: 'b',
            prev: new Set(),
        });

        expect(step1.nextAnchor).toBe('b');

        // Step 2: shift-click on 'd' → range b..d, anchor still 'b'
        const step2 = computeRowClickSelection({
            anchor: step1.nextAnchor,
            flatVisible,
            modifier: 'range',
            path: 'd',
            prev: step1.next,
        });

        expect([...step2.next]).toEqual(['b', 'c', 'd']);
        expect(step2.nextAnchor).toBe('b');

        // Step 3: shift-click on 'e' → range b..e, anchor still 'b'
        const step3 = computeRowClickSelection({
            anchor: step2.nextAnchor,
            flatVisible,
            modifier: 'range',
            path: 'e',
            prev: step2.next,
        });

        expect([...step3.next]).toEqual(['b', 'c', 'd', 'e']);
        expect(step3.nextAnchor).toBe('b');
    });

    it('toggle click moves the anchor and the next shift-click EXTENDS prev from the new origin', () => {
        // single 'a' → anchor 'a' / sel={a}; toggle 'c' → anchor 'c' / sel={a,c};
        // shift 'd' → range c..d ADDED to {a,c} = {a,c,d}, anchor still 'c'.
        // The cmd-pick `a` survives the shift gesture — that's the whole point
        // of the additive contract for users layering selections.
        const s1 = computeRowClickSelection({
            anchor: null,
            flatVisible,
            modifier: 'single',
            path: 'a',
            prev: new Set(),
        });
        const s2 = computeRowClickSelection({
            anchor: s1.nextAnchor,
            flatVisible,
            modifier: 'toggle',
            path: 'c',
            prev: s1.next,
        });
        const s3 = computeRowClickSelection({
            anchor: s2.nextAnchor,
            flatVisible,
            modifier: 'range',
            path: 'd',
            prev: s2.next,
        });

        expect([...s3.next].sort()).toEqual(['a', 'c', 'd']);
        expect(s3.nextAnchor).toBe('c');
    });

    it('plain (single) click clears any existing multi-selection and resets the anchor', () => {
        const cmdA = computeRowClickSelection({
            anchor: null,
            flatVisible,
            modifier: 'toggle',
            path: 'a',
            prev: new Set(),
        });
        const cmdC = computeRowClickSelection({
            anchor: cmdA.nextAnchor,
            flatVisible,
            modifier: 'toggle',
            path: 'c',
            prev: cmdA.next,
        });

        expect([...cmdC.next].sort()).toEqual(['a', 'c']);

        const plainE = computeRowClickSelection({
            anchor: cmdC.nextAnchor,
            flatVisible,
            modifier: 'single',
            path: 'e',
            prev: cmdC.next,
        });

        expect([...plainE.next]).toEqual(['e']);
        expect(plainE.nextAnchor).toBe('e');
    });
});

describe('computeRowClickSelection — purity', () => {
    it('never mutates the input prev Set', () => {
        const prev = new Set(['a']);
        const snapshot = [...prev];

        computeRowClickSelection({
            anchor: 'a',
            flatVisible: ['a', 'b'],
            modifier: 'toggle',
            path: 'b',
            prev,
        });

        expect([...prev]).toEqual(snapshot);
    });

    it('returns a freshly allocated Set on every call', () => {
        const prev = new Set(['a']);
        const result = computeRowClickSelection({
            anchor: null,
            flatVisible: ['a', 'b'],
            modifier: 'single',
            path: 'b',
            prev,
        });

        expect(result.next).not.toBe(prev);
    });

    it('never mutates the input prev Set on a range click that expands directories', () => {
        // Defensive: the additive range branch funnels every expansion through
        // `addAll`, which clones internally. If that contract ever regressed,
        // every directory expansion would silently corrupt React state.
        // Shaped after the user's three-folder scenario so a regression here
        // always lights up a real-world failure mode.
        const prev = new Set(['f1', 'f1/a', 'f1/b']);
        const snapshot = [...prev];

        computeRowClickSelection({
            anchor: 'f1',
            dirSubtreePaths: new Map<string, readonly string[]>([
                ['f1', ['f1', 'f1/a', 'f1/b']],
                ['f3', ['f3', 'f3/a', 'f3/b']],
            ]),
            flatVisible: ['f1', 'f2', 'f3'],
            modifier: 'range',
            path: 'f3',
            prev,
        });

        expect([...prev]).toEqual(snapshot);
    });
});

describe('computeRowClickSelection — range edge cases (additive + folder expansion)', () => {
    /** Visible order with all three folders collapsed at the same level. */
    const collapsedFlat = ['f1', 'f2', 'f3'];
    /** Same tree, but every folder expanded (so descendants are visible too). */
    const expandedFlat = ['f1', 'f1/a', 'f1/b', 'f2', 'f2/a', 'f2/b', 'f3', 'f3/a', 'f3/b'];
    /** Subtree map kept stable across the cases — three sibling folders, two children each. */
    const dirSubtreePaths = new Map<string, readonly string[]>([
        ['f1', ['f1', 'f1/a', 'f1/b']],
        ['f2', ['f2', 'f2/a', 'f2/b']],
        ['f3', ['f3', 'f3/a', 'f3/b']],
    ]);

    it('range click backward across folders (anchor = last, target = first) still selects every folder + subtree', () => {
        // Symmetric to the forward case the user reported. The bug looked
        // worse forward (because forward clicks land on the visually-collapsed
        // tail), but the same `dirSubtreePaths` expansion has to fire when the
        // range is computed in reverse — otherwise `f1` and `f2` would render
        // unchecked even though their paths are in the selection.
        const result = computeRowClickSelection({
            anchor: 'f3',
            dirSubtreePaths,
            flatVisible: collapsedFlat,
            modifier: 'range',
            path: 'f1',
            prev: new Set(['f3', 'f3/a', 'f3/b']),
            subtreePaths: ['f1', 'f1/a', 'f1/b'],
        });

        expect([...result.next].sort()).toEqual(['f1', 'f1/a', 'f1/b', 'f2', 'f2/a', 'f2/b', 'f3', 'f3/a', 'f3/b']);
        expect(result.nextAnchor).toBe('f3');
    });

    it('range across already-EXPANDED folders does not double-count children (Set dedup)', () => {
        // When folders are open, the slice already contains every descendant
        // AND `expandDir` also returns them via `dirSubtreePaths`. The two
        // sources overlap fully — `Set` semantics must collapse them so the
        // resulting size matches the unique-paths count exactly.
        const result = computeRowClickSelection({
            anchor: 'f1',
            dirSubtreePaths,
            flatVisible: expandedFlat,
            modifier: 'range',
            path: 'f3',
            prev: new Set(),
            subtreePaths: ['f3', 'f3/a', 'f3/b'],
        });

        expect(result.next.size).toBe(expandedFlat.length);
        expect([...result.next].sort()).toEqual([...expandedFlat].sort());
        expect(result.nextAnchor).toBe('f1');
    });

    it('range with anchor = file and target = collapsed folder expands ONLY the target folder', () => {
        // Files in the slice contribute just themselves; folders contribute
        // their whole subtree. Mixed slice — `a` (file) → `f1` (folder) —
        // must include `a` flat plus `f1`'s entire branch.
        const result = computeRowClickSelection({
            anchor: 'a',
            dirSubtreePaths,
            flatVisible: ['a', 'f1', 'f3'],
            modifier: 'range',
            path: 'f1',
            prev: new Set(),
            subtreePaths: ['f1', 'f1/a', 'f1/b'],
        });

        expect([...result.next].sort()).toEqual(['a', 'f1', 'f1/a', 'f1/b']);
        expect(result.nextAnchor).toBe('a');
    });

    it('range with anchor = collapsed folder and target = file expands ONLY the anchor folder', () => {
        // Mirror of the previous case from the other side — the anchor folder
        // expands, the trailing file stays single.
        const result = computeRowClickSelection({
            anchor: 'f1',
            dirSubtreePaths,
            flatVisible: ['f1', 'a', 'b'],
            modifier: 'range',
            path: 'b',
            prev: new Set(['f1', 'f1/a', 'f1/b']),
        });

        expect([...result.next].sort()).toEqual(['a', 'b', 'f1', 'f1/a', 'f1/b']);
        expect(result.nextAnchor).toBe('f1');
    });

    it('range expands an empty folder to just its own path (no descendants to fold in)', () => {
        // Defensive: an empty folder's `dirSubtreePaths` entry is `[folder]`.
        // The expansion has to gracefully no-op — adding the same path twice
        // through a Set is fine, but the result must not silently grow.
        const result = computeRowClickSelection({
            anchor: 'a',
            dirSubtreePaths: new Map<string, readonly string[]>([['empty', ['empty']]]),
            flatVisible: ['a', 'empty', 'b'],
            modifier: 'range',
            path: 'b',
            prev: new Set(),
        });

        expect([...result.next].sort()).toEqual(['a', 'b', 'empty']);
        expect(result.nextAnchor).toBe('a');
    });

    it('range expands NESTED directories — outer folder pulls inner folder + grandchildren', () => {
        // `outer` contains `outer/inner` which contains `outer/inner/leaf`.
        // A range click that picks up `outer` must include every level via
        // the precomputed subtree (we do NOT recurse `expandDir` ourselves —
        // `dirSubtreePaths['outer']` already lists every descendant).
        const dirSubtreeMap = new Map<string, readonly string[]>([
            ['outer', ['outer', 'outer/inner', 'outer/inner/leaf']],
            ['outer/inner', ['outer/inner', 'outer/inner/leaf']],
        ]);
        const result = computeRowClickSelection({
            anchor: 'a',
            dirSubtreePaths: dirSubtreeMap,
            flatVisible: ['a', 'outer', 'b'],
            modifier: 'range',
            path: 'outer',
            prev: new Set(),
            subtreePaths: ['outer', 'outer/inner', 'outer/inner/leaf'],
        });

        expect([...result.next].sort()).toEqual(['a', 'outer', 'outer/inner', 'outer/inner/leaf']);
        expect(result.nextAnchor).toBe('a');
    });

    it('range works with dirSubtreePaths omitted — every slice path stays flat (tree-less callers)', () => {
        // Backwards-compatibility contract: callers that don't have a tree
        // map should still get a sensible additive range. Folders inside the
        // slice degrade to single-path inserts, which is exactly the
        // pre-expansion behaviour and is fine for flat consumers.
        const result = computeRowClickSelection({
            anchor: 'f1',
            flatVisible: collapsedFlat,
            modifier: 'range',
            path: 'f3',
            prev: new Set(['x']),
        });

        expect([...result.next].sort()).toEqual(['f1', 'f2', 'f3', 'x']);
        expect(result.nextAnchor).toBe('f1');
    });

    it('anchor-null fallback adds the target folder + subtree (not just the bare path)', () => {
        // Edge case the caller hits when the user's very first interaction is
        // a Shift+click. The fallback path must still respect folder
        // semantics so the click "feels like a click" — the folder lights up
        // fully, not as an indeterminate row.
        const result = computeRowClickSelection({
            anchor: null,
            dirSubtreePaths,
            flatVisible: collapsedFlat,
            modifier: 'range',
            path: 'f2',
            prev: new Set(),
            subtreePaths: ['f2', 'f2/a', 'f2/b'],
        });

        expect([...result.next].sort()).toEqual(['f2', 'f2/a', 'f2/b']);
        expect(result.nextAnchor).toBe('f2');
    });

    it('off-screen anchor fallback adds the target folder + subtree and PRESERVES the original anchor', () => {
        // Scenario: user cmd-clicked `dir/a` (anchor lives inside a folder),
        // collapsed the folder, then Shift+clicked another folder. The anchor
        // is no longer in `flatVisible` so the reducer falls back to a
        // single-shaped add — and that add still has to expand the target.
        const result = computeRowClickSelection({
            anchor: 'dir/a',
            dirSubtreePaths,
            flatVisible: ['f1', 'f2'],
            modifier: 'range',
            path: 'f2',
            prev: new Set(['dir/a']),
            subtreePaths: ['f2', 'f2/a', 'f2/b'],
        });

        expect([...result.next].sort()).toEqual(['dir/a', 'f2', 'f2/a', 'f2/b']);
        expect(result.nextAnchor).toBe('dir/a');
    });

    it('chained shift-clicks accumulate folder subtrees from the same anchor', () => {
        // Step 1: single-click `f1` → sel = whole f1 branch, anchor = f1.
        const step1 = computeRowClickSelection({
            anchor: null,
            dirSubtreePaths,
            flatVisible: collapsedFlat,
            modifier: 'single',
            path: 'f1',
            prev: new Set(),
            subtreePaths: ['f1', 'f1/a', 'f1/b'],
        });

        expect([...step1.next].sort()).toEqual(['f1', 'f1/a', 'f1/b']);

        // Step 2: shift-click `f2` → range f1..f2 expands BOTH folders.
        const step2 = computeRowClickSelection({
            anchor: step1.nextAnchor,
            dirSubtreePaths,
            flatVisible: collapsedFlat,
            modifier: 'range',
            path: 'f2',
            prev: step1.next,
            subtreePaths: ['f2', 'f2/a', 'f2/b'],
        });

        expect([...step2.next].sort()).toEqual(['f1', 'f1/a', 'f1/b', 'f2', 'f2/a', 'f2/b']);
        expect(step2.nextAnchor).toBe('f1');

        // Step 3: shift-click `f3` → anchor still f1, range now covers all 3.
        const step3 = computeRowClickSelection({
            anchor: step2.nextAnchor,
            dirSubtreePaths,
            flatVisible: collapsedFlat,
            modifier: 'range',
            path: 'f3',
            prev: step2.next,
            subtreePaths: ['f3', 'f3/a', 'f3/b'],
        });

        expect([...step3.next].sort()).toEqual(['f1', 'f1/a', 'f1/b', 'f2', 'f2/a', 'f2/b', 'f3', 'f3/a', 'f3/b']);
        expect(step3.nextAnchor).toBe('f1');
    });

    it('toggle a folder, then shift-click another folder → cmd-picked subtree survives + target subtree joins', () => {
        // Real workflow: pick one folder, Cmd-pick a different one, then
        // Shift-click a third to grab everything in between. The toggle moves
        // the anchor to its target, the shift-range expands every folder it
        // touches, and the *original* cmd-pick that lives outside the range
        // stays intact thanks to the additive contract.
        const cmdPick = computeRowClickSelection({
            anchor: null,
            dirSubtreePaths,
            flatVisible: collapsedFlat,
            modifier: 'toggle',
            path: 'f1',
            prev: new Set(),
            subtreePaths: ['f1', 'f1/a', 'f1/b'],
        });
        const shiftPick = computeRowClickSelection({
            anchor: cmdPick.nextAnchor,
            dirSubtreePaths,
            flatVisible: collapsedFlat,
            modifier: 'range',
            path: 'f3',
            prev: cmdPick.next,
            subtreePaths: ['f3', 'f3/a', 'f3/b'],
        });

        expect([...shiftPick.next].sort()).toEqual(['f1', 'f1/a', 'f1/b', 'f2', 'f2/a', 'f2/b', 'f3', 'f3/a', 'f3/b']);
        // Anchor moved to `f1` on toggle, so the range was f1..f3 — and a
        // follow-up shift-click would still extend from `f1`, not `f3`.
        expect(shiftPick.nextAnchor).toBe('f1');
    });

    it('two cmd-picks on separate folders, then shift-click a third → all three subtrees are present', () => {
        // Anchor moves with each toggle, so the final range is rooted at the
        // most recent cmd-pick. Earlier folders survive via the additive
        // union — the test makes sure the union expands their subtrees too,
        // not just the bare folder paths.
        const cmd1 = computeRowClickSelection({
            anchor: null,
            dirSubtreePaths,
            flatVisible: collapsedFlat,
            modifier: 'toggle',
            path: 'f1',
            prev: new Set(),
            subtreePaths: ['f1', 'f1/a', 'f1/b'],
        });
        const cmd2 = computeRowClickSelection({
            anchor: cmd1.nextAnchor,
            dirSubtreePaths,
            flatVisible: collapsedFlat,
            modifier: 'toggle',
            path: 'f2',
            prev: cmd1.next,
            subtreePaths: ['f2', 'f2/a', 'f2/b'],
        });
        const shift = computeRowClickSelection({
            anchor: cmd2.nextAnchor,
            dirSubtreePaths,
            flatVisible: collapsedFlat,
            modifier: 'range',
            path: 'f3',
            prev: cmd2.next,
            subtreePaths: ['f3', 'f3/a', 'f3/b'],
        });

        expect([...shift.next].sort()).toEqual(['f1', 'f1/a', 'f1/b', 'f2', 'f2/a', 'f2/b', 'f3', 'f3/a', 'f3/b']);
        expect(shift.nextAnchor).toBe('f2');
    });

    it('shift-click with anchor === target (a folder) adds the folder subtree exactly once', () => {
        // `slice` reduces to `[folder]`; the expansion has to fire just like
        // a normal range so the user sees the row "fully checked" instead of
        // indeterminate. Idempotent — re-clicking shouldn't toggle anything.
        const result = computeRowClickSelection({
            anchor: 'f1',
            dirSubtreePaths,
            flatVisible: collapsedFlat,
            modifier: 'range',
            path: 'f1',
            prev: new Set(['f1', 'f1/a', 'f1/b']),
            subtreePaths: ['f1', 'f1/a', 'f1/b'],
        });

        expect([...result.next].sort()).toEqual(['f1', 'f1/a', 'f1/b']);
        expect(result.nextAnchor).toBe('f1');
    });

    it('shift-click does not shrink a wider selection even when range is fully inside it', () => {
        // Additive contract: a Shift-click that picks a *narrower* range than
        // the previous one keeps the wider selection intact. This is the one
        // place the behaviour openly diverges from Finder/Explorer (those
        // would replace), and the test pins it down so the divergence stays
        // intentional.
        const result = computeRowClickSelection({
            anchor: 'f1',
            dirSubtreePaths,
            flatVisible: collapsedFlat,
            modifier: 'range',
            path: 'f2',
            prev: new Set(['f1', 'f1/a', 'f1/b', 'f2', 'f2/a', 'f2/b', 'f3', 'f3/a', 'f3/b']),
            subtreePaths: ['f2', 'f2/a', 'f2/b'],
        });

        expect([...result.next].sort()).toEqual(['f1', 'f1/a', 'f1/b', 'f2', 'f2/a', 'f2/b', 'f3', 'f3/a', 'f3/b']);
        expect(result.nextAnchor).toBe('f1');
    });

    it('shift-click adds a target folder subtree when the anchor sits inside an unrelated branch', () => {
        // Anchor is the deepest leaf of `dir1`; user shift-clicks `dir2` (a
        // sibling folder). The range slice covers both folder rows and the
        // file rows in between, but the test explicitly verifies that the
        // target folder's full subtree joins the selection — not just `dir2`.
        const dirSubtreeMap = new Map<string, readonly string[]>([
            ['dir1', ['dir1', 'dir1/a']],
            ['dir2', ['dir2', 'dir2/x', 'dir2/y']],
        ]);
        const result = computeRowClickSelection({
            anchor: 'dir1/a',
            dirSubtreePaths: dirSubtreeMap,
            flatVisible: ['dir1', 'dir1/a', 'between', 'dir2'],
            modifier: 'range',
            path: 'dir2',
            prev: new Set(['dir1', 'dir1/a']),
            subtreePaths: ['dir2', 'dir2/x', 'dir2/y'],
        });

        expect([...result.next].sort()).toEqual(['between', 'dir1', 'dir1/a', 'dir2', 'dir2/x', 'dir2/y']);
        expect(result.nextAnchor).toBe('dir1/a');
    });
});

describe('computeToggleSelection (Space / row checkbox)', () => {
    it('flips a file path on/off', () => {
        expect([...computeToggleSelection({ path: 'a', prev: new Set() })]).toEqual(['a']);
        expect([...computeToggleSelection({ path: 'a', prev: new Set(['a']) })]).toEqual([]);
    });

    it('all-or-nothing for a directory subtree', () => {
        // partial → fully selected
        expect(
            [
                ...computeToggleSelection({
                    path: 'dir',
                    prev: new Set(['dir/a']),
                    subtreePaths: ['dir', 'dir/a', 'dir/b'],
                }),
            ].sort(),
        ).toEqual(['dir', 'dir/a', 'dir/b']);

        // fully selected → cleared
        expect([
            ...computeToggleSelection({
                path: 'dir',
                prev: new Set(['dir', 'dir/a', 'dir/b']),
                subtreePaths: ['dir', 'dir/a', 'dir/b'],
            }),
        ]).toEqual([]);
    });

    it('preserves unrelated paths on subtree toggle', () => {
        expect([
            ...computeToggleSelection({
                path: 'dir',
                prev: new Set(['dir', 'dir/a', 'dir/b', 'x']),
                subtreePaths: ['dir', 'dir/a', 'dir/b'],
            }),
        ]).toEqual(['x']);
    });

    it('treats empty subtreePaths as "single path" (defensive)', () => {
        expect([...computeToggleSelection({ path: 'a', prev: new Set(), subtreePaths: [] })]).toEqual(['a']);
    });

    it('REGRESSION: clears the branch in one click when every descendant is selected but the dir itself was never added', () => {
        // Real-world: user opens `dir`, ticks each child checkbox (or the
        // header's "select all" with `dir` being the only visible group). The
        // folder's tri-state shows "checked" because every descendant is in
        // the selection — but `dir` itself isn't. A single click on the
        // folder checkbox must DESELECT the whole branch; the previous code
        // required two clicks (the first one silently added `dir`'s own path,
        // the second one finally removed everything because by then the
        // strict "every path including dir" check passed).
        expect([
            ...computeToggleSelection({
                path: 'dir',
                prev: new Set(['dir/a', 'dir/b']),
                subtreePaths: ['dir', 'dir/a', 'dir/b'],
            }),
        ]).toEqual([]);
    });
});

describe('computeToggleSelectAll', () => {
    it('selects every path when nothing is selected', () => {
        expect([...computeToggleSelectAll({ allSelectablePaths: ['a', 'b', 'c'], prev: new Set() })].sort()).toEqual([
            'a',
            'b',
            'c',
        ]);
    });

    it('selects every path when the selection is partial (some-selected state)', () => {
        expect(
            [...computeToggleSelectAll({ allSelectablePaths: ['a', 'b', 'c'], prev: new Set(['a']) })].sort(),
        ).toEqual(['a', 'b', 'c']);
    });

    it('clears the selection when everything is already selected', () => {
        expect([...computeToggleSelectAll({ allSelectablePaths: ['a', 'b'], prev: new Set(['a', 'b']) })]).toEqual([]);
    });

    it('returns an empty Set when the universe is empty (no selectable paths)', () => {
        expect([...computeToggleSelectAll({ allSelectablePaths: [], prev: new Set() })]).toEqual([]);
        expect([...computeToggleSelectAll({ allSelectablePaths: [], prev: new Set(['stale']) })]).toEqual([]);
    });

    it('returns a freshly allocated Set (never mutates prev)', () => {
        const prev = new Set(['a']);
        const result = computeToggleSelectAll({ allSelectablePaths: ['a', 'b'], prev });

        expect(result).not.toBe(prev);
        expect([...prev]).toEqual(['a']);
    });
});

describe('computeDirSelectionState — non-empty directory', () => {
    // Subtree paths convention: the directory's own path always comes first,
    // descendants follow. `dirSubtreePaths` in `FileManager` builds it via
    // `collectSubtreePaths(node)` — same shape.
    const paths = ['dir', 'dir/a', 'dir/b'];

    it('returns false when nothing inside is selected', () => {
        expect(computeDirSelectionState({ path: 'dir', paths, selectedPaths: new Set() })).toBe(false);
    });

    it('returns true when every descendant is selected (and the dir itself happens to be too)', () => {
        expect(
            computeDirSelectionState({
                path: 'dir',
                paths,
                selectedPaths: new Set(['dir', 'dir/a', 'dir/b']),
            }),
        ).toBe(true);
    });

    it('returns true when every descendant is selected even WITHOUT the dir in selectedPaths', () => {
        // Confirms the count is purely descendant-based.
        expect(
            computeDirSelectionState({
                path: 'dir',
                paths,
                selectedPaths: new Set(['dir/a', 'dir/b']),
            }),
        ).toBe(true);
    });

    it('returns indeterminate when only some descendants are selected', () => {
        expect(
            computeDirSelectionState({
                path: 'dir',
                paths,
                selectedPaths: new Set(['dir/a']),
            }),
        ).toBe('indeterminate');
    });

    it('returns indeterminate when the dir itself + some descendants are selected', () => {
        expect(
            computeDirSelectionState({
                path: 'dir',
                paths,
                selectedPaths: new Set(['dir', 'dir/a']),
            }),
        ).toBe('indeterminate');
    });

    it('REGRESSION: returns false when ONLY the directory itself is selected (no descendants)', () => {
        // Reproduces the bug where the user single-clicked a folder
        // (selecting `{dir, dir/a, dir/b}`), then cmd-clicked away every child
        // one by one — leaving just `{dir}` in the selection. Previously the
        // checkbox stuck at "indeterminate" even though the folder's content
        // was empty in the selection.
        expect(
            computeDirSelectionState({
                path: 'dir',
                paths,
                selectedPaths: new Set(['dir']),
            }),
        ).toBe(false);
    });
});

describe('computeDirSelectionState — empty directory (no real descendants)', () => {
    // For an empty folder the "subtree" is just the folder's own path.
    const paths = ['empty'];

    it('returns false when the empty folder is not selected', () => {
        expect(computeDirSelectionState({ path: 'empty', paths, selectedPaths: new Set() })).toBe(false);
    });

    it('returns true when the empty folder is itself selected', () => {
        expect(
            computeDirSelectionState({
                path: 'empty',
                paths,
                selectedPaths: new Set(['empty']),
            }),
        ).toBe(true);
    });

    it('never returns indeterminate for empty folders', () => {
        // Edge case: `selectedPaths` contains an unrelated entry — empty folder
        // must still resolve to a binary state.
        expect(
            computeDirSelectionState({
                path: 'empty',
                paths,
                selectedPaths: new Set(['unrelated']),
            }),
        ).toBe(false);
    });
});

describe('computeDirSelectionState — nested hierarchy', () => {
    // Real-world shape: outer/sub/file.txt
    const outerPaths = ['outer', 'outer/sub', 'outer/sub/file.txt'];
    const subPaths = ['outer/sub', 'outer/sub/file.txt'];

    it('inner folder reflects only its own descendants (file)', () => {
        expect(
            computeDirSelectionState({
                path: 'outer/sub',
                paths: subPaths,
                selectedPaths: new Set(['outer/sub/file.txt']),
            }),
        ).toBe(true);

        expect(
            computeDirSelectionState({
                path: 'outer/sub',
                paths: subPaths,
                selectedPaths: new Set(),
            }),
        ).toBe(false);

        // Inner folder itself selected, but file inside isn't → reflects content only.
        expect(
            computeDirSelectionState({
                path: 'outer/sub',
                paths: subPaths,
                selectedPaths: new Set(['outer/sub']),
            }),
        ).toBe(false);
    });

    it('outer folder treats the inner folder + file as separate descendants', () => {
        // Only the inner folder is selected (no file inside) → outer is indeterminate
        // because something within outer (= the sub-dir) IS in the selection.
        expect(
            computeDirSelectionState({
                path: 'outer',
                paths: outerPaths,
                selectedPaths: new Set(['outer/sub']),
            }),
        ).toBe('indeterminate');

        // Both inner and file selected → outer becomes fully selected.
        expect(
            computeDirSelectionState({
                path: 'outer',
                paths: outerPaths,
                selectedPaths: new Set(['outer/sub', 'outer/sub/file.txt']),
            }),
        ).toBe(true);
    });

    it('outer folder ignores its OWN path when counting (regression)', () => {
        expect(
            computeDirSelectionState({
                path: 'outer',
                paths: outerPaths,
                selectedPaths: new Set(['outer']),
            }),
        ).toBe(false);
    });
});

describe('computeDirSelectionState — synthetic group root (path === "")', () => {
    // Top-level group root has an empty path; children carry full paths.
    // The reducer must treat the empty `path` correctly as "the group root".
    const paths = ['', 'uploads/a.txt', 'uploads/b.txt'];

    it('returns false when nothing inside the group is selected', () => {
        expect(computeDirSelectionState({ path: '', paths, selectedPaths: new Set() })).toBe(false);
    });

    it('returns true when every child is selected', () => {
        expect(
            computeDirSelectionState({
                path: '',
                paths,
                selectedPaths: new Set(['uploads/a.txt', 'uploads/b.txt']),
            }),
        ).toBe(true);
    });

    it('returns indeterminate when only some children are selected', () => {
        expect(
            computeDirSelectionState({
                path: '',
                paths,
                selectedPaths: new Set(['uploads/a.txt']),
            }),
        ).toBe('indeterminate');
    });
});

describe('sortFileManagerTree', () => {
    const collectNames = (nodes: FileManagerInternalNode[]): string[] =>
        nodes.flatMap((n) => [n.name, ...collectNames(n.children)]);

    const collectTopNames = (nodes: FileManagerInternalNode[]): string[] => nodes.map((n) => n.name);

    it('returns input unchanged (by reference) when sorting is null AND isFoldersFirst is false', () => {
        const tree = buildFileManagerTree([file('b.txt'), file('a.txt')]);
        const result = sortFileManagerTree(tree, null, false);

        expect(result).toBe(tree);
    });

    it('partitions folders before files when sorting=null and isFoldersFirst=true', () => {
        // Insertion order is `b.txt`, `alpha/...`, `a.txt`, `beta/...`. With no
        // active sort and isFoldersFirst=true, the partitioner pulls folders to
        // the top while *preserving* the original order WITHIN each partition.
        const tree = buildFileManagerTree([
            file('b.txt'),
            file('alpha/inner.txt'),
            file('a.txt'),
            file('beta/inner.txt'),
        ]);
        const sorted = sortFileManagerTree(tree, null, true);

        // `buildFileManagerTree` itself sorts the input by path before walking,
        // so folders appear in path-alphabetical order (`alpha`, `beta`) and
        // files likewise (`a.txt`, `b.txt`). The folders-first partition keeps
        // both groups intact, only flipping their relative position.
        expect(collectTopNames(sorted)).toEqual(['alpha', 'beta', 'a.txt', 'b.txt']);
    });

    it('preserves insertion order within each partition when sorting=null', () => {
        // Same structure as above but verifying stability explicitly. Since
        // `buildFileManagerTree` pre-sorts files by path, the file partition
        // here reads in path order (z, then m), and the folder partition
        // contains a single entry. The point is no comparator runs, only the
        // folders-first partition.
        const tree = buildFileManagerTree([file('m.txt'), file('z.txt'), file('dir/inner.txt')]);
        const sorted = sortFileManagerTree(tree, null, true);

        expect(collectTopNames(sorted)).toEqual(['dir', 'm.txt', 'z.txt']);
    });

    it('sorts files alphabetically asc with isFoldersFirst=true', () => {
        const tree = buildFileManagerTree([file('b.txt'), file('a.txt'), dir('zzz'), file('c.txt')]);
        const sorted = sortFileManagerTree(tree, { column: 'name', direction: 'asc' }, true);

        expect(collectTopNames(sorted)).toEqual(['zzz', 'a.txt', 'b.txt', 'c.txt']);
    });

    it('sorts files alphabetically desc', () => {
        const tree = buildFileManagerTree([file('a.txt'), file('b.txt'), file('c.txt')]);
        const sorted = sortFileManagerTree(tree, { column: 'name', direction: 'desc' }, false);

        expect(collectTopNames(sorted)).toEqual(['c.txt', 'b.txt', 'a.txt']);
    });

    it('mixes folders with files when isFoldersFirst=false', () => {
        const tree = buildFileManagerTree([file('alpha/file.txt'), file('zzz.txt'), file('beta/file.txt')]);
        const sorted = sortFileManagerTree(tree, { column: 'name', direction: 'asc' }, false);

        expect(collectTopNames(sorted)).toEqual(['alpha', 'beta', 'zzz.txt']);
    });

    it('keeps folders above files when isFoldersFirst=true', () => {
        const tree = buildFileManagerTree([
            file('alpha/file.txt'),
            file('zzz.txt'),
            file('beta/file.txt'),
            file('aaa.txt'),
        ]);
        const sorted = sortFileManagerTree(tree, { column: 'name', direction: 'asc' }, true);

        expect(collectTopNames(sorted)).toEqual(['alpha', 'beta', 'aaa.txt', 'zzz.txt']);
    });

    it('sorts by size with `0` fallback for files lacking the field', () => {
        const tree = buildFileManagerTree([
            file('a.txt', { size: 100 }),
            file('b.txt', { size: 50 }),
            file('c.txt', { size: 200 }),
        ]);
        const sorted = sortFileManagerTree(tree, { column: 'size', direction: 'asc' }, false);

        expect(collectTopNames(sorted)).toEqual(['b.txt', 'a.txt', 'c.txt']);
    });

    it('sorts by modified ascending then descending', () => {
        const tree = buildFileManagerTree([
            file('a.txt', { modifiedAt: '2024-01-01' }),
            file('b.txt', { modifiedAt: '2026-01-01' }),
            file('c.txt', { modifiedAt: '2025-01-01' }),
        ]);
        const ascending = sortFileManagerTree(tree, { column: 'modified', direction: 'asc' }, false);
        const descending = sortFileManagerTree(tree, { column: 'modified', direction: 'desc' }, false);

        expect(collectTopNames(ascending)).toEqual(['a.txt', 'c.txt', 'b.txt']);
        expect(collectTopNames(descending)).toEqual(['b.txt', 'c.txt', 'a.txt']);
    });

    it('preserves the order of synthetic group roots but sorts their children', () => {
        const groups: FileManagerRootGroup[] = [
            { id: 'uploads', label: 'Uploads', pathPrefix: 'uploads' },
            { id: 'container', label: 'Container', pathPrefix: 'container' },
        ];
        const tree = buildFileManagerTree(
            [file('uploads/b.txt'), file('uploads/a.txt'), file('container/x.txt'), file('container/m.txt')],
            groups,
        );

        const sorted = sortFileManagerTree(tree, { column: 'name', direction: 'asc' }, false);

        expect(collectTopNames(sorted)).toEqual(['Uploads', 'Container']);
        expect(collectNames(sorted[0]?.children ?? [])).toEqual(['a.txt', 'b.txt']);
        expect(collectNames(sorted[1]?.children ?? [])).toEqual(['m.txt', 'x.txt']);
    });

    it('recursively sorts nested folder children', () => {
        const tree = buildFileManagerTree([file('outer/sub/c.txt'), file('outer/sub/a.txt'), file('outer/sub/b.txt')]);
        const sorted = sortFileManagerTree(tree, { column: 'name', direction: 'asc' }, true);

        const outer = sorted[0];
        const sub = outer?.children[0];
        expect(sub?.name).toBe('sub');
        expect(collectTopNames(sub?.children ?? [])).toEqual(['a.txt', 'b.txt', 'c.txt']);
    });
});

// `compareSizes` collapses directories (and synthetic group roots) to size `0`.
// While `isFoldersFirst: true` is active the folders-first partition runs
// before the comparator, so that detail is invisible. Once the user flips the
// "Folders first" toggle off (Resources page), directories sort as 0-byte
// rows alongside files. The behaviour is intentional but easy to regress;
// these tests pin both sides of the contract explicitly.
describe('sortFileManagerTree by size with directories', () => {
    const collectTopNames = (nodes: FileManagerInternalNode[]): string[] => nodes.map((n) => n.name);

    it('places directories with files at 0 bytes when folders-first is disabled', () => {
        const tree = buildFileManagerTree([
            file('a.txt', { size: 100 }),
            dir('big-empty-folder'),
            file('b.txt', { size: 50 }),
        ]);

        const sorted = sortFileManagerTree(tree, { column: 'size', direction: 'asc' }, false);

        expect(collectTopNames(sorted)).toEqual(['big-empty-folder', 'b.txt', 'a.txt']);
    });

    it('keeps directories above files when folders-first is enabled, regardless of size', () => {
        const tree = buildFileManagerTree([
            file('big.txt', { size: 1000 }),
            dir('zzz-folder'),
            file('small.txt', { size: 1 }),
        ]);

        const sorted = sortFileManagerTree(tree, { column: 'size', direction: 'asc' }, true);

        expect(collectTopNames(sorted)).toEqual(['zzz-folder', 'small.txt', 'big.txt']);
    });

    it('keeps directories above files in desc direction too when folders-first is enabled', () => {
        const tree = buildFileManagerTree([
            file('big.txt', { size: 1000 }),
            dir('zzz-folder'),
            file('small.txt', { size: 1 }),
        ]);

        const sorted = sortFileManagerTree(tree, { column: 'size', direction: 'desc' }, true);

        expect(collectTopNames(sorted)).toEqual(['zzz-folder', 'big.txt', 'small.txt']);
    });
});

describe('computeNextSort', () => {
    it('cycles none → asc', () => {
        expect(computeNextSort(null, 'name')).toEqual({ column: 'name', direction: 'asc' });
    });

    it('cycles asc → desc on the same column', () => {
        expect(computeNextSort({ column: 'name', direction: 'asc' }, 'name')).toEqual({
            column: 'name',
            direction: 'desc',
        });
    });

    it('cycles desc → none on the same column', () => {
        expect(computeNextSort({ column: 'name', direction: 'desc' }, 'name')).toBeNull();
    });

    it('switches to a different column starting at asc', () => {
        expect(computeNextSort({ column: 'name', direction: 'desc' }, 'size')).toEqual({
            column: 'size',
            direction: 'asc',
        });
    });
});
