import { type DragEvent as ReactDragEvent, useCallback, useEffect, useRef, useState } from 'react';

import type { FileManagerInternalNode, FileNode } from './file-manager-types';

import { dedupeOverlappingPaths } from './file-manager-utils';

const FM_DND_MIME = 'application/x-fm-paths';

/** Sentinel destination — represents the root area outside any directory. */
export const FM_ROOT_DROP_SENTINEL = '__fm_root__';

export interface FileManagerContainerDndHandlers {
    isRootDropTarget: boolean;
    onDragEnter: (event: ReactDragEvent<HTMLDivElement>) => void;
    onDragLeave: (event: ReactDragEvent<HTMLDivElement>) => void;
    onDragOver: (event: ReactDragEvent<HTMLDivElement>) => void;
    onDrop: (event: ReactDragEvent<HTMLDivElement>) => void;
}

export interface FileManagerNodeDndHandlers {
    /**
     * `true` when this row is part of the in-flight drag operation. Drives the
     * "ghosted" appearance for every selected row when the user drags one of them,
     * so it's obvious that the whole batch is being moved (not just the row whose
     * native drag image the browser is showing).
     */
    isBeingDragged: boolean;
    isDropTarget: boolean;
    onDragEnd: () => void;
    onDragEnter: (event: ReactDragEvent<HTMLDivElement>) => void;
    onDragLeave: (event: ReactDragEvent<HTMLDivElement>) => void;
    onDragOver: (event: ReactDragEvent<HTMLDivElement>) => void;
    onDragStart: (event: ReactDragEvent<HTMLDivElement>) => void;
    onDrop: (event: ReactDragEvent<HTMLDivElement>) => void;
}

interface UseFileManagerDndParams {
    /** Stable "real" `FileNode` reference for a path — used to build the source list. */
    findNode: (path: string) => FileManagerInternalNode | undefined;
    /**
     * Invoked when the user starts dragging a row that's NOT part of the current
     * selection (Finder-style: "you're acting on something else, the selection is
     * dropped") and after a successful drop (selection paths are now stale because
     * the items live elsewhere). Optional — when omitted, selection isn't touched.
     */
    onClearSelection?: () => void;
    /** When undefined, DnD is fully disabled. */
    onMoveItems?: (sources: FileNode[], destinationDir: string) => Promise<void> | void;
    /**
     * Currently-selected paths. When the user grabs a row that's part of this set,
     * every selected item is dragged together; otherwise only the grabbed row.
     * Synthetic group roots and overlapping descendants are filtered automatically.
     */
    selectedPaths?: Set<string>;
}

interface UseFileManagerDndResult {
    /** Returns drag/drop handlers bound to a specific tree node, or `null` when DnD is off. */
    bindNodeDnd: (node: FileManagerInternalNode) => FileManagerNodeDndHandlers | null;
    /**
     * Drag/drop handlers for the outer tree container. Drops here are interpreted as
     * "move to library root" and only fire when the cursor is over the empty area
     * outside any row (rows always stop event propagation).
     */
    container: FileManagerContainerDndHandlers;
    /** When true, rows have to set `draggable={true}` themselves. */
    isEnabled: boolean;
}

/** Returns the parent directory of a virtual path, or `''` for root. */
const getParentDir = (path: string): string => {
    const idx = path.lastIndexOf('/');

    return idx === -1 ? '' : path.slice(0, idx);
};

/**
 * Validates that every source can be moved into `destDir`:
 *   - never into itself,
 *   - never into its current parent (no-op),
 *   - never into one of its own descendants.
 */
const isValidMove = (sources: FileManagerInternalNode[], destDir: string): boolean => {
    if (sources.length === 0) {
        return false;
    }

    for (const src of sources) {
        if (src.path === destDir) {
            return false;
        }

        if (getParentDir(src.path) === destDir) {
            return false;
        }

        if (src.isDir && (destDir === src.path || destDir.startsWith(`${src.path}/`))) {
            return false;
        }
    }

    return true;
};

const isFmDragEvent = (event: ReactDragEvent<HTMLDivElement>): boolean =>
    event.dataTransfer.types?.includes(FM_DND_MIME) ?? false;

/**
 * Top-level files (no parent directory) act as a pass-through to the container's
 * root-drop logic: instead of treating the row as its own drop target, we let the
 * drag event bubble so the container accepts it as a "move to library root".
 *
 * The user mental model: anything outside any folder is "the root" — that includes
 * the empty padding *and* loose files sitting at the top level.
 */
const isRootPassthrough = (node: FileManagerInternalNode): boolean =>
    !node.isDir && !node.isGroupRoot && getParentDir(node.path) === '';

/**
 * Encapsulates the drag-counter pattern + path-set tracking used by `FileManager` for
 * intra-tree move-via-drag. Scoped to a single `FileManager` instance — all intra-instance
 * drags share one set of refs, but two separate instances do not interfere.
 *
 * Row handlers stop event propagation so the container only sees drags over the empty
 * (root) area. Counters live on `enter`/`leave` (per the W3C drag-counter pattern); drop
 * targets are resolved on `over` only via `preventDefault`.
 */
export const useFileManagerDnd = ({
    findNode,
    onClearSelection,
    onMoveItems,
    selectedPaths,
}: UseFileManagerDndParams): UseFileManagerDndResult => {
    const isEnabled = !!onMoveItems;

    // Stash via ref so the dragstart handler doesn't re-create on every selection
    // change (which would invalidate `bindNodeDnd` and re-render every row through
    // the tree). Synced via effect to keep ESLint happy about render-time mutations.
    const selectionRef = useRef(selectedPaths);

    useEffect(() => {
        selectionRef.current = selectedPaths;
    }, [selectedPaths]);

    const dragSourcesRef = useRef<FileManagerInternalNode[]>([]);
    const containerCounterRef = useRef(0);
    const nodeCounterRef = useRef(new Map<string, number>());

    // Drop-target highlighting. `null` = no active drag, otherwise the path of the directory
    // currently hovered, or `FM_ROOT_DROP_SENTINEL` when the user is over the empty root area.
    const [dropTargetPath, setDropTargetPath] = useState<null | string>(null);
    // Paths of every row currently being dragged. Used to ghost rows in the UI so the
    // user sees that the whole selection is on the move, not just the grabbed row
    // (whose drag image the browser already shows).
    const [draggingPaths, setDraggingPaths] = useState<ReadonlySet<string>>(() => new Set());

    const resetDragState = useCallback(() => {
        dragSourcesRef.current = [];
        containerCounterRef.current = 0;
        nodeCounterRef.current.clear();
        setDropTargetPath(null);
        setDraggingPaths(new Set());
    }, []);

    // ── per-node handlers ───────────────────────────────────────────────────────

    const handleNodeDragStart = useCallback(
        (node: FileManagerInternalNode, event: ReactDragEvent<HTMLDivElement>): void => {
            if (!isEnabled || node.isGroupRoot) {
                return;
            }

            const selection = selectionRef.current;
            const isPartOfSelection = !!selection && selection.has(node.path);

            // Build the source path list:
            //   - If the dragged row IS part of the selection → drag everything selected
            //     (deduped so a parent dir doesn't ship together with its descendants).
            //   - Otherwise → drag just this row, and clear the previous selection so
            //     the user isn't left with a stale highlight pointing at unrelated items.
            const sourcePaths = isPartOfSelection && selection ? dedupeOverlappingPaths(selection) : [node.path];

            if (!isPartOfSelection && selection && selection.size > 0) {
                onClearSelection?.();
            }

            // Re-resolve through `findNode` so we drag the freshest data — the row's
            // `node` prop may be stale if the tree was re-rendered mid-drag.
            const sources: FileManagerInternalNode[] = [];

            for (const path of sourcePaths) {
                const fresh = findNode(path);

                // Group roots are synthetic headers, never valid sources.
                if (fresh && !fresh.isGroupRoot) {
                    sources.push(fresh);
                }
            }

            if (sources.length === 0) {
                return;
            }

            dragSourcesRef.current = sources;
            event.dataTransfer.effectAllowed = 'move';
            // Browsers require *some* payload for drag to start; the actual list lives
            // in the ref. Newline-separate so external listeners can still parse it.
            event.dataTransfer.setData(FM_DND_MIME, sources.map((source) => source.path).join('\n'));

            // For multi-source drags swap the browser's default drag image (= the
            // grabbed row only) with a compact "N items" badge — without it the
            // user can't tell at a glance that the whole selection is on the move.
            if (sources.length > 1) {
                const badge = document.createElement('div');

                badge.textContent = `${sources.length} items`;
                badge.style.cssText = [
                    'position: fixed',
                    // Render off-screen so we never flash it to the user — `setDragImage`
                    // captures the visual immediately, the element itself is throwaway.
                    'top: -1000px',
                    'left: -1000px',
                    'padding: 6px 12px',
                    'background: hsl(var(--primary))',
                    'color: hsl(var(--primary-foreground))',
                    'border-radius: 6px',
                    'font: 500 13px system-ui, -apple-system, "Segoe UI", sans-serif',
                    'white-space: nowrap',
                    'pointer-events: none',
                    'box-shadow: 0 4px 12px rgba(0, 0, 0, 0.18)',
                ].join(';');

                document.body.appendChild(badge);
                event.dataTransfer.setDragImage(badge, 12, 12);
                // The browser snapshots the element synchronously into a bitmap, so
                // it's safe to remove on the next frame.
                requestAnimationFrame(() => badge.remove());
            }

            // Mark every dragged source so the rows can ghost themselves. setState
            // here is safe — React commits the update *after* this handler returns,
            // so the browser captures the drag image with the original (un-ghosted)
            // styles before the dim style is applied.
            setDraggingPaths(new Set(sources.map((source) => source.path)));
        },
        [findNode, isEnabled, onClearSelection],
    );

    const handleNodeDragEnter = useCallback(
        (node: FileManagerInternalNode, event: ReactDragEvent<HTMLDivElement>): void => {
            if (!isEnabled || !isFmDragEvent(event)) {
                return;
            }

            // Top-level files behave as part of the root drop area — let the event
            // bubble so the container handler shows the root highlight.
            if (isRootPassthrough(node)) {
                return;
            }

            // Otherwise stop propagation so a drag over a row never bubbles into the
            // container handler — otherwise the container would treat the row position
            // as a drop into the empty (root) area and call the move API on release.
            event.stopPropagation();

            const isValidDirTarget =
                node.isDir && !node.isGroupRoot && isValidMove(dragSourcesRef.current, node.path);

            if (!isValidDirTarget) {
                // Cursor is now over a non-droppable row — make sure the previously
                // shown root highlight (if any) gets cleared. Container `dragleave`
                // gates clearing on `relatedTarget` to avoid flicker, so the row
                // handler is responsible for this case.
                setDropTargetPath((current) => (current === FM_ROOT_DROP_SENTINEL ? null : current));

                return;
            }

            const counters = nodeCounterRef.current;
            const next = (counters.get(node.path) ?? 0) + 1;
            counters.set(node.path, next);

            if (next === 1) {
                setDropTargetPath(node.path);
            }
        },
        [isEnabled],
    );

    const handleNodeDragLeave = useCallback(
        (node: FileManagerInternalNode, event: ReactDragEvent<HTMLDivElement>): void => {
            if (!isEnabled || !isFmDragEvent(event)) {
                return;
            }

            // Mirrors `handleNodeDragEnter`: pass-through rows must let `dragleave`
            // bubble too, so the container's enter/leave counter stays balanced.
            if (isRootPassthrough(node)) {
                return;
            }

            event.stopPropagation();

            const counters = nodeCounterRef.current;
            const current = counters.get(node.path) ?? 0;

            if (current <= 1) {
                counters.delete(node.path);
                setDropTargetPath((value) => (value === node.path ? null : value));
            } else {
                counters.set(node.path, current - 1);
            }
        },
        [isEnabled],
    );

    const handleNodeDragOver = useCallback(
        (node: FileManagerInternalNode, event: ReactDragEvent<HTMLDivElement>): void => {
            if (!isEnabled || !isFmDragEvent(event)) {
                return;
            }

            // Pass-through rows defer drop-acceptance to the container (= root drop).
            if (isRootPassthrough(node)) {
                return;
            }

            // Otherwise stop propagation; only `preventDefault` (= "drop allowed") for
            // valid directory targets. Without the stop, the container would
            // `preventDefault` on top of us and accept any row position as a root drop.
            event.stopPropagation();

            if (!node.isDir || node.isGroupRoot) {
                return;
            }

            if (!isValidMove(dragSourcesRef.current, node.path)) {
                return;
            }

            event.preventDefault();
            event.dataTransfer.dropEffect = 'move';
        },
        [isEnabled],
    );

    const handleNodeDrop = useCallback(
        (node: FileManagerInternalNode, event: ReactDragEvent<HTMLDivElement>): void => {
            if (!isEnabled || !isFmDragEvent(event)) {
                return;
            }

            // Pass-through rows let the container handle the actual drop (= root move).
            if (isRootPassthrough(node)) {
                return;
            }

            // Otherwise stop propagation so a drop on a row never bubbles to the
            // container, which would otherwise treat the drop as a root move (the source
            // bug behind "drag a file, return it back" → unintended API call).
            event.stopPropagation();

            if (!node.isDir || node.isGroupRoot) {
                resetDragState();

                return;
            }

            const sources = dragSourcesRef.current;

            if (!isValidMove(sources, node.path)) {
                resetDragState();

                return;
            }

            event.preventDefault();
            resetDragState();
            // Selection paths reference the OLD locations, which the move call is
            // about to invalidate. Clear them so the bulk-actions bar / "select-all"
            // checkbox don't show stale state.
            onClearSelection?.();
            void onMoveItems?.(sources, node.path);
        },
        [isEnabled, onClearSelection, onMoveItems, resetDragState],
    );

    const bindNodeDnd = useCallback(
        (node: FileManagerInternalNode): FileManagerNodeDndHandlers | null => {
            if (!isEnabled) {
                return null;
            }

            return {
                isBeingDragged: draggingPaths.has(node.path),
                isDropTarget: dropTargetPath === node.path,
                onDragEnd: resetDragState,
                onDragEnter: (event) => handleNodeDragEnter(node, event),
                onDragLeave: (event) => handleNodeDragLeave(node, event),
                onDragOver: (event) => handleNodeDragOver(node, event),
                onDragStart: (event) => handleNodeDragStart(node, event),
                onDrop: (event) => handleNodeDrop(node, event),
            };
        },
        [
            draggingPaths,
            dropTargetPath,
            handleNodeDragEnter,
            handleNodeDragLeave,
            handleNodeDragOver,
            handleNodeDragStart,
            handleNodeDrop,
            isEnabled,
            resetDragState,
        ],
    );

    // ── container-level handlers (root drop area) ───────────────────────────────

    const handleContainerDragEnter = useCallback(
        (event: ReactDragEvent<HTMLDivElement>): void => {
            if (!isEnabled || !isFmDragEvent(event)) {
                return;
            }

            containerCounterRef.current += 1;

            if (!isValidMove(dragSourcesRef.current, '')) {
                return;
            }

            // Highlight only when the user is *outside* every directory row — otherwise the
            // node-level handler already owns the highlight and propagation was stopped there.
            if (containerCounterRef.current === 1) {
                setDropTargetPath((current) => current ?? FM_ROOT_DROP_SENTINEL);
            }
        },
        [isEnabled],
    );

    const handleContainerDragLeave = useCallback(
        (event: ReactDragEvent<HTMLDivElement>): void => {
            if (!isEnabled || !isFmDragEvent(event)) {
                return;
            }

            containerCounterRef.current = Math.max(containerCounterRef.current - 1, 0);

            // Only clear the root highlight when the cursor truly left the container.
            // `dragleave` also fires when the cursor enters a child element — clearing
            // there would cause a flicker for pass-through rows (which immediately
            // re-set the highlight via bubbled `dragenter`/`dragover`).
            const relatedTarget = event.relatedTarget;
            const cursorLeftContainer =
                !(relatedTarget instanceof Node) || !event.currentTarget.contains(relatedTarget);

            if (cursorLeftContainer) {
                setDropTargetPath((current) => (current === FM_ROOT_DROP_SENTINEL ? null : current));
            }
        },
        [isEnabled],
    );

    const handleContainerDragOver = useCallback(
        (event: ReactDragEvent<HTMLDivElement>): void => {
            if (!isEnabled || !isFmDragEvent(event)) {
                return;
            }

            if (!isValidMove(dragSourcesRef.current, '')) {
                return;
            }

            event.preventDefault();
            event.dataTransfer.dropEffect = 'move';
            // `dragover` fires continuously while the cursor is over the container's
            // direct area (or a pass-through child). It is the moment-of-truth for the
            // root highlight: relying on the counter pattern alone misses the case
            // where the cursor is implicitly inside the container at `dragstart`
            // (no `dragenter` ever fires on the container, so the counter stays at 0
            // and the highlight never appears until the user leaves and re-enters).
            setDropTargetPath((current) => current ?? FM_ROOT_DROP_SENTINEL);
        },
        [isEnabled],
    );

    const handleContainerDrop = useCallback(
        (event: ReactDragEvent<HTMLDivElement>): void => {
            if (!isEnabled || !isFmDragEvent(event)) {
                return;
            }

            const sources = dragSourcesRef.current;

            if (!isValidMove(sources, '')) {
                resetDragState();

                return;
            }

            event.preventDefault();
            resetDragState();
            // Selection paths are about to be invalidated by the move — clear them so
            // the bulk bar doesn't display stale "N selected".
            onClearSelection?.();
            void onMoveItems?.(sources, '');
        },
        [isEnabled, onClearSelection, onMoveItems, resetDragState],
    );

    return {
        bindNodeDnd,
        container: {
            isRootDropTarget: dropTargetPath === FM_ROOT_DROP_SENTINEL,
            onDragEnter: handleContainerDragEnter,
            onDragLeave: handleContainerDragLeave,
            onDragOver: handleContainerDragOver,
            onDrop: handleContainerDrop,
        },
        isEnabled,
    };
};
