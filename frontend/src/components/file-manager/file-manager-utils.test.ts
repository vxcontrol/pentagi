import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

import type { FileManagerInternalNode, FileManagerRootGroup, FileNode } from './file-manager-types';

import {
    buildFileManagerGridTemplate,
    buildFileManagerTree,
    clamp,
    collectAllFilePaths,
    collectAllNodePaths,
    collectDirectoryPaths,
    collectVisibleFlat,
    dedupeOverlappingPaths,
    filterFileManagerTree,
    findNodeByPath,
    formatFileSize,
    formatModified,
    getCheckboxState,
    normalizeRootGroups,
    pluralizeItemsEnglish,
    resolveSelectionModifier,
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

describe('pluralizeItemsEnglish', () => {
    it('uses singular for 1', () => {
        expect(pluralizeItemsEnglish(1)).toBe('1 item');
    });

    it('uses plural for 0 and >1', () => {
        expect(pluralizeItemsEnglish(0)).toBe('0 items');
        expect(pluralizeItemsEnglish(5)).toBe('5 items');
    });
});
