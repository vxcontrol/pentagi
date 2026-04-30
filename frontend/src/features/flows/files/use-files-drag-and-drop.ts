import { useCallback, useRef, useState } from 'react';

interface DragHandlers {
    onDragEnter: (event: React.DragEvent<HTMLDivElement>) => void;
    onDragLeave: (event: React.DragEvent<HTMLDivElement>) => void;
    onDragOver: (event: React.DragEvent<HTMLDivElement>) => void;
    onDrop: (event: React.DragEvent<HTMLDivElement>) => void;
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
 * to remount the subtree (via `key={flowId}` or similar) which discards both the
 * counter and the `isDragging` state.
 */
export const useFilesDragAndDrop = ({
    canAcceptDrop,
    onDrop,
}: UseFilesDragAndDropParams): UseFilesDragAndDropResult => {
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
        },
        isDragging,
    };
};
