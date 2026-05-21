import type { ReactNode } from 'react';

import type { DetailNavigationController } from './use-detail-navigation';

import { DetailNavigationButtons } from './detail-navigation-buttons';
import { DetailNavigationSheet } from './detail-navigation-sheet';

export interface DetailNavigationToolbarProps<T extends { id: string }> {
    controller: DetailNavigationController<T>;
    /**
     * Forwarded to `<DetailNavigationSheet>` — controls the in-sheet search
     * input. Defaults to `true`; pass `false` to opt out.
     */
    hasSearch?: boolean;
    renderItem?: (item: T, isCurrent: boolean) => ReactNode;
    /** Forwarded placeholder for the in-sheet search input. */
    searchPlaceholder?: string;
    sheetIcon?: ReactNode;
    sheetTitle: string;
}

/**
 * Convenience wrapper that composes `<DetailNavigationButtons>` and
 * `<DetailNavigationSheet>` against a single `DetailNavigationController`.
 * Most desktop call sites use this directly; pages with non-standard chrome
 * (e.g. mobile prev/position/next inside a `<DropdownMenuItem>`) can compose
 * the leaves themselves and read from the same controller.
 *
 * Renders `null` when the controller reports `itemsEmpty` — saves the user
 * from a momentary "–/0" flash while the parent provider's data is in flight.
 */
export function DetailNavigationToolbar<T extends { id: string }>({
    controller,
    hasSearch,
    renderItem,
    searchPlaceholder,
    sheetIcon,
    sheetTitle,
}: DetailNavigationToolbarProps<T>) {
    if (controller.itemsEmpty) {
        return null;
    }

    return (
        <>
            <DetailNavigationButtons
                controller={controller}
                sheetTitle={sheetTitle}
            />
            <DetailNavigationSheet
                controller={controller}
                hasSearch={hasSearch}
                renderItem={renderItem}
                searchPlaceholder={searchPlaceholder}
                sheetIcon={sheetIcon}
                sheetTitle={sheetTitle}
            />
        </>
    );
}
