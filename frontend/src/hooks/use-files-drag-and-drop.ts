import { useCallback, useRef, useState } from 'react';

interface DragHandlers {
    onDragEnter: (event: React.DragEvent<HTMLDivElement>) => void;
    onDragLeave: (event: React.DragEvent<HTMLDivElement>) => void;
    onDragOver: (event: React.DragEvent<HTMLDivElement>) => void;
    onDrop: (event: React.DragEvent<HTMLDivElement>) => void;
    /**
     * Capture-phase drop listener: resets the internal counter and the
     * `isDragging` flag *before* any nested handler can call
     * `event.stopPropagation()`. Necessary so a child component (e.g. the
     * `FileManager`'s row-level external-file drop) can claim the drop and
     * stop the bubble — the bubble-phase `onDrop` below would otherwise
     * never fire and `isDragging` would stay stuck on `true`.
     */
    onDropCapture: (event: React.DragEvent<HTMLDivElement>) => void;
}

interface UseFilesDragAndDropParams {
    canAcceptDrop: boolean;
    onDrop: (droppedFiles: File[]) => void;
}

interface UseFilesDragAndDropResult {
    dragHandlers: DragHandlers;
    isDragging: boolean;
}

const isFileDragEvent = (event: React.DragEvent<HTMLDivElement>): boolean =>
    event.dataTransfer.types?.includes('Files') ?? false;

/**
 * Encapsulates the drag-counter pattern (drag-enter / drag-leave fire for every nested
 * element, so we increment/decrement a counter to know when the user actually leaves
 * the drop zone). The mutable counter lives entirely inside this hook; the consumer
 * sees only an immutable `isDragging` flag and four event handlers.
 *
 * The hook does not auto-reset on external identity changes: the consumer is expected
 * to remount the subtree (via `key={...}` on the surrounding component) which discards
 * both the counter and the `isDragging` state.
 */
export function useFilesDragAndDrop({ canAcceptDrop, onDrop }: UseFilesDragAndDropParams): UseFilesDragAndDropResult {
    const dragCounterRef = useRef(0);
    const [isDragging, setIsDragging] = useState(false);

    const handleDragEnter = useCallback(
        (event: React.DragEvent<HTMLDivElement>) => {
            if (!canAcceptDrop || !isFileDragEvent(event)) {
                return;
            }

            event.preventDefault();
            event.stopPropagation();
            dragCounterRef.current += 1;
            setIsDragging(true);
        },
        [canAcceptDrop],
    );

    const handleDragOver = useCallback(
        (event: React.DragEvent<HTMLDivElement>) => {
            if (!canAcceptDrop || !isFileDragEvent(event)) {
                return;
            }

            event.preventDefault();
            event.stopPropagation();
            event.dataTransfer.dropEffect = 'copy';
        },
        [canAcceptDrop],
    );

    const handleDragLeave = useCallback((event: React.DragEvent<HTMLDivElement>) => {
        event.preventDefault();
        event.stopPropagation();
        dragCounterRef.current = Math.max(dragCounterRef.current - 1, 0);

        if (dragCounterRef.current === 0) {
            setIsDragging(false);
        }
    }, []);

    // Capture-phase: ALWAYS reset the local state, even if a descendant claims
    // the drop and stops bubble propagation. Without this, a child handler
    // (e.g. row-level upload-into-folder) leaves the page-level overlay
    // visually stuck on "Drop files to upload" because no `dragleave` fires
    // after a drop and the bubble-phase reset never runs.
    const handleDropCapture = useCallback(() => {
        dragCounterRef.current = 0;
        setIsDragging(false);
    }, []);

    const handleDrop = useCallback(
        (event: React.DragEvent<HTMLDivElement>) => {
            event.preventDefault();
            event.stopPropagation();
            dragCounterRef.current = 0;
            setIsDragging(false);

            if (!canAcceptDrop) {
                return;
            }

            const droppedFiles = Array.from(event.dataTransfer.files ?? []);

            if (droppedFiles.length === 0) {
                return;
            }

            onDrop(droppedFiles);
        },
        [canAcceptDrop, onDrop],
    );

    return {
        dragHandlers: {
            onDragEnter: handleDragEnter,
            onDragLeave: handleDragLeave,
            onDragOver: handleDragOver,
            onDrop: handleDrop,
            onDropCapture: handleDropCapture,
        },
        isDragging,
    };
}
