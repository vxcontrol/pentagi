import { useCallback, useEffect, useRef, useState } from 'react';

export interface IdentifiableItem {
    id: string;
}

interface UseChatScrollOptions {
    autoScrollBehavior?: ScrollBehavior;
}

export const useChatScroll = <T extends IdentifiableItem>(
    items: T[] | undefined,
    resetKey: null | string | undefined,
    options?: UseChatScrollOptions,
) => {
    const containerRef = useRef<HTMLDivElement>(null);
    const endRef = useRef<HTMLDivElement>(null);

    const [isScrolledToBottom, setIsScrolledToBottom] = useState(true);
    const [hasNewMessages, setHasNewMessages] = useState(false);

    const isScrolledToBottomRef = useRef(true);
    const lastMessageIdRef = useRef<null | string>(null);
    const previousScrollTopRef = useRef<number>(0);

    const behavior = options?.autoScrollBehavior ?? 'smooth';

    useEffect(() => {
        isScrolledToBottomRef.current = isScrolledToBottom;

        if (isScrolledToBottom) {
            setHasNewMessages(false);
        }
    }, [isScrolledToBottom]);

    const scrollToEnd = useCallback(() => {
        if (endRef.current) {
            endRef.current.scrollIntoView({ behavior });
            setIsScrolledToBottom(true);
            setHasNewMessages(false);
        }
    }, [behavior]);

    // Initialize previous scrollTop
    useEffect(() => {
        if (containerRef.current) {
            previousScrollTopRef.current = containerRef.current.scrollTop;
        }
    }, []);

    const handleScroll = useCallback(() => {
        const containerElement = containerRef.current;

        if (!containerElement) {
            return;
        }

        const { clientHeight, scrollHeight, scrollTop } = containerElement;
        const effectivelyAtBottom = scrollHeight - scrollTop <= clientHeight + 2;

        if (effectivelyAtBottom) {
            if (!isScrolledToBottomRef.current) {
                setIsScrolledToBottom(true);
            }
        } else {
            const scrollUpThreshold = 10;

            if (isScrolledToBottomRef.current && scrollTop < previousScrollTopRef.current - scrollUpThreshold) {
                setIsScrolledToBottom(false);
            }
        }

        previousScrollTopRef.current = scrollTop;
    }, []);

    // Attach scroll listener
    useEffect(() => {
        const containerElement = containerRef.current;

        if (!containerElement) {
            return;
        }

        containerElement.addEventListener('scroll', handleScroll);

        return () => {
            containerElement.removeEventListener('scroll', handleScroll);
        };
    }, [handleScroll]);

    const hasVerticalScroll = useCallback(() => {
        const containerElement = containerRef.current;

        if (!containerElement) {
            return false;
        }

        return containerElement.scrollHeight - containerElement.clientHeight > 2;
    }, []);

    // Detect new messages when user is not at the bottom
    useEffect(() => {
        if (!items?.length) {
            lastMessageIdRef.current = null;

            return;
        }

        const lastItem = items.at(-1);
        const lastId = lastItem?.id;

        if (!lastId) {
            lastMessageIdRef.current = null;

            return;
        }

        if (
            lastMessageIdRef.current &&
            lastMessageIdRef.current !== lastId &&
            !isScrolledToBottomRef.current &&
            hasVerticalScroll()
        ) {
            setHasNewMessages(true);
        }

        lastMessageIdRef.current = lastId;
    }, [items]);

    // Treat absence of vertical scroll as being at the bottom
    useEffect(() => {
        if (!hasVerticalScroll()) {
            if (!isScrolledToBottom) {
                setIsScrolledToBottom(true);
            }

            if (hasNewMessages) {
                setHasNewMessages(false);
            }
        }
    }, [items, hasVerticalScroll, isScrolledToBottom, hasNewMessages]);

    // Auto-scroll when items update and user is at bottom
    useEffect(() => {
        if (isScrolledToBottomRef.current) {
            scrollToEnd();
        }
    }, [items, scrollToEnd]);

    // Reset indicators on context switch (e.g., another flow or assistant)
    useEffect(() => {
        setHasNewMessages(false);
        const lastId = items?.at(-1)?.id ?? null;
        lastMessageIdRef.current = lastId ?? null;
    }, [resetKey]);

    return {
        containerRef,
        endRef,
        hasNewMessages,
        isScrolledToBottom,
        scrollToEnd,
        setHasNewMessages, // exposed for explicit resets if needed
        setIsScrolledToBottom, // exposed in case component needs manual override
    } as const;
};
