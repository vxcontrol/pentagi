import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

import type { FileManagerInternalNode, FileManagerRootGroup, FileNode } from './file-manager-types';

import {
    addAllToSet,
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
    formatModified,
    getCheckboxState,
    isEverySelected,
    normalizeRootGroups,
    pluralizeItemsEnglish,
    removeAllFromSet,
    resolveSelectionModifier,
    toggleSubtreeOnSet,
    walkTree,
} from './file-manager-utils';

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

describe('formatModified', () => {
    const NOW = new Date('2026-04-28T12:00:00Z').getTime();

    beforeEach(() => {
        vi.useFakeTimers();
        vi.setSystemTime(NOW);
    });

    afterEach(() => {
        vi.useRealTimers();
    });

    it('returns empty for missing/invalid input', () => {
        expect(formatModified(undefined)).toBe('');
        expect(formatModified('not-a-date')).toBe('');
    });

    it('handles "just now"', () => {
        expect(formatModified(new Date(NOW - 10_000))).toBe('just now');
    });

    it('formats minutes / hours / days within a week', () => {
        expect(formatModified(new Date(NOW - 5 * 60_000))).toBe('5m ago');
        expect(formatModified(new Date(NOW - 3 * 60 * 60_000))).toBe('3h ago');
        expect(formatModified(new Date(NOW - 4 * 24 * 60 * 60_000))).toBe('4d ago');
    });

    it('falls back to locale date for older entries', () => {
        const old = new Date(NOW - 14 * 24 * 60 * 60_000);

        expect(formatModified(old)).toBe(old.toLocaleDateString());
    });

    it('accepts ISO strings', () => {
        const iso = new Date(NOW - 10 * 60_000).toISOString();

        expect(formatModified(iso)).toBe('10m ago');
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

describe('addAllToSet', () => {
    it('adds every path from the iterable into the set', () => {
        const set = new Set(['a']);

        addAllToSet(set, ['b', 'c']);

        expect([...set].sort()).toEqual(['a', 'b', 'c']);
    });

    it('is a no-op for an empty iterable', () => {
        const set = new Set(['a']);

        addAllToSet(set, []);

        expect([...set]).toEqual(['a']);
    });

    it('accepts another Set as the source', () => {
        const set = new Set<string>();

        addAllToSet(set, new Set(['x', 'y']));

        expect([...set].sort()).toEqual(['x', 'y']);
    });
});

describe('removeAllFromSet', () => {
    it('removes every present path and ignores missing ones', () => {
        const set = new Set(['a', 'b', 'c']);

        removeAllFromSet(set, ['b', 'missing']);

        expect([...set].sort()).toEqual(['a', 'c']);
    });

    it('is a no-op for an empty iterable', () => {
        const set = new Set(['a']);

        removeAllFromSet(set, []);

        expect([...set]).toEqual(['a']);
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
});

describe('computeRowClickSelection — range modifier', () => {
    const flatVisible = ['a', 'b', 'c', 'd', 'e'];

    it('selects from anchor to target (forward)', () => {
        const result = computeRowClickSelection({
            anchor: 'b',
            flatVisible,
            modifier: 'range',
            path: 'd',
            prev: new Set(['a', 'b']),
        });

        expect([...result.next]).toEqual(['b', 'c', 'd']);
        expect(result.nextAnchor).toBe('b');
    });

    it('selects from anchor to target (backward)', () => {
        const result = computeRowClickSelection({
            anchor: 'd',
            flatVisible,
            modifier: 'range',
            path: 'b',
            prev: new Set(['d']),
        });

        expect([...result.next]).toEqual(['b', 'c', 'd']);
        expect(result.nextAnchor).toBe('d');
    });

    it('replaces previous selection (does not extend it)', () => {
        const result = computeRowClickSelection({
            anchor: 'c',
            flatVisible,
            modifier: 'range',
            path: 'e',
            prev: new Set(['a']),
        });

        expect([...result.next]).toEqual(['c', 'd', 'e']);
        expect(result.nextAnchor).toBe('c');
    });

    it('reduces to a single-row selection when anchor === target', () => {
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

    it('falls back to "single" when no anchor is set yet (and moves the anchor)', () => {
        const result = computeRowClickSelection({
            anchor: null,
            flatVisible,
            modifier: 'range',
            path: 'c',
            prev: new Set(['a']),
        });

        expect([...result.next]).toEqual(['c']);
        expect(result.nextAnchor).toBe('c');
    });

    it('falls back to "single" when the anchor is no longer visible (and PRESERVES the anchor)', () => {
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

        expect([...result.next]).toEqual(['other']);
        expect(result.nextAnchor).toBe('dir/a');
    });

    it('falls back to "single" when the target is not in flatVisible (defensive)', () => {
        const result = computeRowClickSelection({
            anchor: 'a',
            flatVisible,
            modifier: 'range',
            path: 'ghost',
            prev: new Set(['a']),
        });

        expect([...result.next]).toEqual(['ghost']);
        expect(result.nextAnchor).toBe('a');
    });

    it('range click on a directory uses flatVisible (not subtreePaths) when anchor is valid', () => {
        // The directory's own subtree is irrelevant for range clicks — only the
        // visible-order span matters, mirroring Finder/Explorer.
        const result = computeRowClickSelection({
            anchor: 'a',
            flatVisible: ['a', 'b', 'dir', 'e'],
            modifier: 'range',
            path: 'dir',
            prev: new Set(['a']),
            subtreePaths: ['dir', 'dir/x', 'dir/y'],
        });

        expect([...result.next]).toEqual(['a', 'b', 'dir']);
        expect(result.nextAnchor).toBe('a');
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

    it('toggle click moves the anchor (so subsequent shift-click extends from the new origin)', () => {
        // single 'a' → anchor 'a'; toggle 'c' → anchor 'c'; shift 'd' → c..d
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

        expect([...s3.next]).toEqual(['c', 'd']);
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
        expect(
            [
                ...computeToggleSelection({
                    path: 'dir',
                    prev: new Set(['dir', 'dir/a', 'dir/b']),
                    subtreePaths: ['dir', 'dir/a', 'dir/b'],
                }),
            ],
        ).toEqual([]);
    });

    it('preserves unrelated paths on subtree toggle', () => {
        expect(
            [
                ...computeToggleSelection({
                    path: 'dir',
                    prev: new Set(['dir', 'dir/a', 'dir/b', 'x']),
                    subtreePaths: ['dir', 'dir/a', 'dir/b'],
                }),
            ],
        ).toEqual(['x']);
    });

    it('treats empty subtreePaths as "single path" (defensive)', () => {
        expect([...computeToggleSelection({ path: 'a', prev: new Set(), subtreePaths: [] })]).toEqual(['a']);
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
        expect([
            ...computeToggleSelectAll({ allSelectablePaths: ['a', 'b'], prev: new Set(['a', 'b']) }),
        ]).toEqual([]);
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
