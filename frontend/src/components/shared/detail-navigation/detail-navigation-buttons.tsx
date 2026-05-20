import { ChevronLeft, ChevronRight } from 'lucide-react';

import { Button } from '@/components/ui/button';
import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/ui/tooltip';
import { cn } from '@/lib/utils';

import type { DetailNavigationController } from './use-detail-navigation';

interface DetailNavigationButtonsProps<T extends { id: string }> {
    controller: DetailNavigationController<T>;
    /** Lowercased plural used in the aria-label / tooltip ("flows", "templates"). */
    sheetTitle: string;
    /**
     * Size variant. `'default'` is the desktop toolbar's `size-8` cluster;
     * `'sm'` shrinks the cluster to `size-7` for embedding inside a
     * `<DropdownMenuItem>` on mobile, where the host row is already padded.
     */
    size?: 'default' | 'sm';
}

/**
 * Prev / Position / Next button cluster bound to a `DetailNavigationController`.
 * Stateless: the controller owns navigation, `isSheetOpen`, and the
 * pre-formatted `positionLabel`.
 *
 * Reused in both the desktop toolbar (`size="default"`) and the mobile
 * dropdown row (`size="sm"`) — same a11y contract, same tooltips, same
 * keyboard semantics in both places.
 */
export function DetailNavigationButtons<T extends { id: string }>({
    controller,
    sheetTitle,
    size = 'default',
}: DetailNavigationButtonsProps<T>) {
    const lowerTitle = sheetTitle.toLowerCase();
    const isSm = size === 'sm';
    const sideButtonSize = isSm ? 'size-7' : 'size-8';
    const middleHeight = isSm ? 'h-7' : 'h-8';

    return (
        <div className="flex items-center">
            <Tooltip>
                <TooltipTrigger asChild>
                    <Button
                        aria-label="Previous"
                        className={cn(sideButtonSize, 'rounded-r-none border-r-0 p-0')}
                        disabled={!controller.prevId}
                        onClick={controller.goToPrev}
                        size="icon"
                        type="button"
                        variant="outline"
                    >
                        <ChevronLeft />
                    </Button>
                </TooltipTrigger>
                <TooltipContent>Previous</TooltipContent>
            </Tooltip>
            <Tooltip>
                <TooltipTrigger asChild>
                    <Button
                        aria-label={`Open ${lowerTitle} list (${controller.positionLabel})`}
                        className={cn(
                            middleHeight,
                            'min-w-12 rounded-none border-x px-2 font-mono text-xs tabular-nums',
                        )}
                        disabled={!controller.hasEntries}
                        onClick={controller.openSheet}
                        type="button"
                        variant="outline"
                    >
                        {controller.positionLabel}
                    </Button>
                </TooltipTrigger>
                <TooltipContent>Show all matching {lowerTitle}</TooltipContent>
            </Tooltip>
            <Tooltip>
                <TooltipTrigger asChild>
                    <Button
                        aria-label="Next"
                        className={cn(sideButtonSize, 'rounded-l-none border-l-0 p-0')}
                        disabled={!controller.nextId}
                        onClick={controller.goToNext}
                        size="icon"
                        type="button"
                        variant="outline"
                    >
                        <ChevronRight />
                    </Button>
                </TooltipTrigger>
                <TooltipContent>Next</TooltipContent>
            </Tooltip>
        </div>
    );
}
