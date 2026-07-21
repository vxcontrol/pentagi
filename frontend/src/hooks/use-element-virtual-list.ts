import { useVirtualizer, type VirtualItem, type Virtualizer } from '@tanstack/react-virtual';

interface UseElementVirtualListOptions {
    count: number;
    enabled?: boolean;
    estimateSize: () => number;
    gap?: number;
    /** Stable key per index so measurements survive reordering (sort/filter). */
    getItemKey?: (index: number) => number | string;
    getScrollElement: () => HTMLElement | null;
    overscan?: number;
    padding?: number;
}

interface UseElementVirtualListResult {
    scrollToIndex: Virtualizer<HTMLElement, Element>['scrollToIndex'];
    totalSize: number;
    virtualItems: VirtualItem[];
}

/**
 * Element-scrolled list virtualization — the inner-`overflow-auto`-container
 * analogue of {@link useWindowVirtualList}.
 */
export function useElementVirtualList({
    count,
    enabled = true,
    estimateSize,
    gap,
    getItemKey,
    getScrollElement,
    overscan = 10,
    padding = 0,
}: UseElementVirtualListOptions): UseElementVirtualListResult {
    const virtualizer = useVirtualizer({
        count,
        enabled,
        estimateSize,
        gap,
        getItemKey,
        getScrollElement,
        overscan,
        paddingEnd: padding,
        paddingStart: padding,
    });

    return {
        scrollToIndex: virtualizer.scrollToIndex,
        totalSize: virtualizer.getTotalSize(),
        virtualItems: virtualizer.getVirtualItems(),
    };
}
