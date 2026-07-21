import { useWindowVirtualizer, type VirtualItem } from '@tanstack/react-virtual';
import { type RefObject, useLayoutEffect, useRef, useState } from 'react';

/**
 * Headless window-scrolled list virtualization.
 *
 * Wraps `@tanstack/react-virtual`'s `useWindowVirtualizer` and supplies the
 * three pieces its raw API leaves to the caller:
 *
 * - keeps `scrollMargin` in sync with the anchor element's document-Y
 *   position, including when content above the anchor shifts size (page
 *   header growing, banner appearing, sidebar variant flipping);
 * - exposes the stable `measureElement` reference so dynamic row heights
 *   work without recreating the ref callback on every render;
 * - returns ready-to-render spacer pixels (`paddingStart` / `paddingEnd`)
 *   so consumers don't reimplement the arithmetic.
 *
 * Render structure the consumer is responsible for:
 *
 * ```
 *   <Anchor ref={anchorRef}>
 *     <Spacer height={paddingStart} />
 *     {virtualItems.map((vi) => (
 *       <Item key={...} ref={measureItem} data-index={vi.index} />
 *     ))}
 *     <Spacer height={paddingEnd} />
 *   </Anchor>
 * ```
 *
 * Anchor placement — its top edge defines item 0's document position. For
 * tables that's `<tbody>` (the header sits above it, so attaching the ref to
 * the table wrapper would offset all items by the header height); for
 * div-based lists it's the list container itself.
 *
 * Anchor must be present when the hook mounts. The `scrollMargin` sync runs
 * once in `useLayoutEffect` and does not re-attach if the anchor renders
 * later. If the anchor is conditional in the consumer, mount/unmount the
 * hook owner together with it (the recommended pattern below already does
 * this).
 *
 * Mount the hook owner conditionally — render `<VirtualizedX/>` only when
 * `count > threshold` — so non-virtualized consumers never pay the cost of the
 * window scroll/resize listeners the virtualizer attaches on mount.
 *
 * Scroll container — assumes `window`. Layouts with an inner `overflow-auto`
 * container need {@link useElementVirtualList} instead; this hook will
 * misposition items there.
 *
 * `overscan` defaults to 5 (not react-virtual's default of 1) so the user
 * sees rows materialize ahead of the viewport edge during fast scrolling.
 */
type UseWindowVirtualListOptions = {
    count: number;
    /**
     * Estimated item size in pixels. Pass a memoized reference (`useCallback`
     * or module-scope constant) so the virtualizer doesn't treat each render
     * as a measurement-affecting change.
     */
    estimateSize: () => number;
    /**
     * Stable identifier per index. TanStack uses this as the key for the
     * internal measurement cache; without it the cache is keyed by index,
     * which mis-attributes measurements when rows reorder (sort, filter) and
     * heights vary.
     */
    getItemKey?: (index: number) => number | string;
    /** Items rendered above + below the viewport. */
    overscan?: number;
};

type UseWindowVirtualListResult<T extends Element> = {
    /**
     * Attach to the element whose top edge marks item 0's document-Y
     * position. See the structural example in the module JSDoc.
     */
    anchorRef: RefObject<null | T>;
    /**
     * Stable ref to forward to each rendered item. Pair with
     * `data-index={virtualItem.index}` on the same node — TanStack reads the
     * attribute to know which item the measurement belongs to.
     */
    measureItem: (node: Element | null) => void;
    /** Pixels of spacer to render after the last visible item. */
    paddingEnd: number;
    /** Pixels of spacer to render before the first visible item. */
    paddingStart: number;
    /** Sum of all item sizes — the list's full scrollable height. */
    totalSize: number;
    /** Items currently in the render window (visible + overscan). */
    virtualItems: VirtualItem[];
};

export function useWindowVirtualList<T extends Element = HTMLDivElement>({
    count,
    estimateSize,
    getItemKey,
    overscan = 5,
}: UseWindowVirtualListOptions): UseWindowVirtualListResult<T> {
    const anchorRef = useRef<null | T>(null);
    const [scrollMargin, setScrollMargin] = useState(0);

    useLayoutEffect(() => {
        if (!anchorRef.current) {
            return;
        }

        const syncScrollMargin = () => {
            const node = anchorRef.current;

            if (!node) {
                return;
            }

            // Anchor's document-Y (scroll-invariant): rect.top + scrollY.
            setScrollMargin(node.getBoundingClientRect().top + window.scrollY);
        };

        syncScrollMargin();

        // Re-sync when content *above* the anchor changes height (header
        // growth, banner mount, sidebar variant flip) and shifts its
        // document-Y. The anchor's own resize can't move its own top edge, so
        // observing `document.body` is both necessary and sufficient.
        const observer = new ResizeObserver(syncScrollMargin);

        observer.observe(document.body);

        window.addEventListener('resize', syncScrollMargin);

        return () => {
            observer.disconnect();
            window.removeEventListener('resize', syncScrollMargin);
        };
    }, []);

    const virtualizer = useWindowVirtualizer({
        count,
        estimateSize,
        getItemKey,
        overscan,
        scrollMargin,
    });

    const virtualItems = virtualizer.getVirtualItems();
    const totalSize = virtualizer.getTotalSize();

    const firstItem = virtualItems[0];
    const lastItem = virtualItems.at(-1);
    const paddingStart = firstItem ? firstItem.start - scrollMargin : 0;
    const paddingEnd = lastItem ? totalSize - (lastItem.end - scrollMargin) : 0;

    return {
        anchorRef,
        measureItem: virtualizer.measureElement,
        paddingEnd,
        paddingStart,
        totalSize,
        virtualItems,
    };
}
