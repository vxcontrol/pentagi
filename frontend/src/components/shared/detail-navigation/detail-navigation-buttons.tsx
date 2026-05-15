import { ChevronLeft, ChevronRight } from 'lucide-react';

import { Button } from '@/components/ui/button';
import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/ui/tooltip';

interface DetailNavigationButtonsProps {
    /** Disable the position-button when the filtered subset is empty. */
    hasEntries: boolean;
    /** No next sibling — disable the right chevron. */
    nextDisabled: boolean;
    onNext: () => void;
    onOpen: () => void;
    onPrev: () => void;
    /** Pre-formatted `"3/10"` (or `"–/0"` when no current). */
    positionLabel: string;
    /** No previous sibling — disable the left chevron. */
    prevDisabled: boolean;
    /** Lowercased plural used in the aria-label / tooltip ("flows", "templates"). */
    sheetTitle: string;
}

/**
 * Prev / Position / Next button cluster for a detail page. Stateless and
 * presentation-only — `DetailNavigationToolbar` owns the navigation logic
 * and feeds the resolved indices, labels, and click handlers down.
 *
 * Kept separate from `DetailNavigationSheet` so the same buttons could in
 * principle be reused without the sheet (e.g. a future variant that ships
 * keyboard-only navigation without an overlay).
 */
export const DetailNavigationButtons = ({
    hasEntries,
    nextDisabled,
    onNext,
    onOpen,
    onPrev,
    positionLabel,
    prevDisabled,
    sheetTitle,
}: DetailNavigationButtonsProps) => {
    const lowerTitle = sheetTitle.toLowerCase();

    return (
        <div className="flex items-center">
            <Tooltip>
                <TooltipTrigger asChild>
                    <Button
                        aria-label="Previous"
                        className="size-8 rounded-r-none border-r-0 p-0"
                        disabled={prevDisabled}
                        onClick={onPrev}
                        size="icon"
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
                        aria-label={`Open ${lowerTitle} list (${positionLabel})`}
                        className="h-8 min-w-12 rounded-none border-x px-2 font-mono text-xs tabular-nums"
                        disabled={!hasEntries}
                        onClick={onOpen}
                        variant="outline"
                    >
                        {positionLabel}
                    </Button>
                </TooltipTrigger>
                <TooltipContent>Show all matching {lowerTitle}</TooltipContent>
            </Tooltip>
            <Tooltip>
                <TooltipTrigger asChild>
                    <Button
                        aria-label="Next"
                        className="size-8 rounded-l-none border-l-0 p-0"
                        disabled={nextDisabled}
                        onClick={onNext}
                        size="icon"
                        variant="outline"
                    >
                        <ChevronRight />
                    </Button>
                </TooltipTrigger>
                <TooltipContent>Next</TooltipContent>
            </Tooltip>
        </div>
    );
};
