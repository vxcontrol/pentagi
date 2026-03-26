import { useCallback, useEffect, useRef, useState } from 'react';

export interface IdentifiableItem {
    id: string;
}

interface ScrollTracker {
    lastItemId: null | string;
    resetKey: null | string | undefined;
}

export const useAutoScroll = <T extends IdentifiableItem>(
    items: T[] | undefined,
    resetKey: null | string | undefined,
) => {
    const containerElementRef = useRef<HTMLDivElement | null>(null);
    const endRef = useRef<HTMLDivElement>(null);

    const [isScrolledToBottom, setIsScrolledToBottom] = useState(true);
    const [hasNewMessages, setHasNewMessages] = useState(false);

    const isScrolledToBottomRef = useRef(true);
    const isScrollAnimatingRef = useRef(false);

    const currentLastItemId = items?.at(-1)?.id ?? null;
    const [tracker, setTracker] = useState<ScrollTracker>({
        lastItemId: null,
        resetKey,
    });

    if (resetKey !== tracker.resetKey || currentLastItemId !== tracker.lastItemId) {
        const isReset = resetKey !== tracker.resetKey;
        setTracker({ lastItemId: currentLastItemId, resetKey });

        if (isReset) {
            setIsScrolledToBottom(true);
            setHasNewMessages(false);
        } else if (tracker.lastItemId && currentLastItemId && !isScrolledToBottom) {
            setHasNewMessages(true);
        }
    }

    const scrollToEnd = useCallback((behavior: ScrollBehavior = 'smooth') => {
        if (endRef.current) {
            isScrollAnimatingRef.current = true;
            endRef.current.scrollIntoView({ behavior });
            isScrolledToBottomRef.current = true;
            setIsScrolledToBottom(true);
            setHasNewMessages(false);
        }
    }, []);

    const handleScroll = useCallback(() => {
        const containerElement = containerElementRef.current;

        if (!containerElement) {
            return;
        }

        const { clientHeight, scrollHeight, scrollTop } = containerElement;
        const distanceFromBottom = scrollHeight - scrollTop - clientHeight;

        if (distanceFromBottom <= 2) {
            isScrollAnimatingRef.current = false;

            if (!isScrolledToBottomRef.current) {
                isScrolledToBottomRef.current = true;
                setIsScrolledToBottom(true);
                setHasNewMessages(false);
            }
        } else if (isScrolledToBottomRef.current && !isScrollAnimatingRef.current) {
            isScrolledToBottomRef.current = false;
            setIsScrolledToBottom(false);
        }
    }, []);

    const containerRef = useCallback(
        (node: HTMLDivElement | null) => {
            if (containerElementRef.current) {
                containerElementRef.current.removeEventListener('scroll', handleScroll);
            }

            containerElementRef.current = node;

            if (node) {
                node.addEventListener('scroll', handleScroll);
            }
        },
        [handleScroll],
    );

    useEffect(() => {
        isScrolledToBottomRef.current = true;
    }, [resetKey]);

    useEffect(() => {
        const container = containerElementRef.current;

        if (!container) {
            return;
        }

        const hasVerticalScroll = container.scrollHeight - container.clientHeight > 2;

        if (!hasVerticalScroll && !isScrolledToBottomRef.current) {
            isScrolledToBottomRef.current = true;
            queueMicrotask(() => {
                setIsScrolledToBottom(true);
                setHasNewMessages(false);
            });

            return;
        }

        if (isScrolledToBottomRef.current) {
            endRef.current?.scrollIntoView({ behavior: 'instant' });
        }
    }, [items, resetKey]);

    return {
        containerRef,
        endRef,
        hasNewMessages,
        isScrolledToBottom,
        scrollToEnd,
    } as const;
};
