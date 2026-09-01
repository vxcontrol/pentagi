import { CircleArrowDown, CircleCheck, CircleQuestionMark, ExternalLink, Info } from 'lucide-react';
import { useState } from 'react';

import type { VersionInfoFragmentFragment } from '@/graphql/types';

import { Badge } from '@/components/ui/badge';
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/popover';
import { Separator } from '@/components/ui/separator';
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from '@/components/ui/tooltip';
import { cn } from '@/lib/utils';
import { useVersionInfo } from '@/providers/version-info-provider';

import {
    describeTiming,
    describeVersion,
    PENTAGI_RELEASES_URL,
    releaseNotesUrl,
    type VersionIndicator,
    type VersionPresentation,
} from './version-status';

interface VersionBadgeProps {
    className?: string;
    /** Prefix the version with the product name, for places without the logo beside it. */
    showProductName?: boolean;
}

const INDICATOR_ICON: Record<Exclude<VersionIndicator, 'none'>, typeof CircleArrowDown> = {
    question: CircleQuestionMark,
    update: CircleArrowDown,
};

const INDICATOR_CLASS = 'text-yellow-500 dark:text-yellow-400';

/**
 * A dot for the places where the badge itself cannot fit — the logo of a sidebar
 * collapsed to icons. Rendered only when the badge would carry an indicator.
 */
export function VersionAttentionDot({ className }: { className?: string }) {
    const { versionInfo } = useVersionInfo();

    if (!versionInfo || describeVersion(versionInfo).indicator === 'none') {
        return null;
    }

    return (
        <span
            aria-hidden
            className={cn('size-2 rounded-full bg-yellow-500 dark:bg-yellow-400', className)}
            data-testid="version-attention-dot"
        />
    );
}

/**
 * The running version, with a yellow "?" when nothing is known about updates and a
 * yellow "↓" when a newer version is offered. Hover explains in a line; click opens
 * the details with links to the release notes.
 *
 * Inside a collapsible sidebar the version text hides in the icon state and only
 * the indicator stays, which is the part that matters when space is short.
 */
export function VersionBadge({ className, showProductName = false }: VersionBadgeProps) {
    const { versionInfo } = useVersionInfo();
    const [popoverOpen, setPopoverOpen] = useState(false);
    const [tooltipOpen, setTooltipOpen] = useState(false);

    if (!versionInfo) {
        return null;
    }

    const presentation = describeVersion(versionInfo);
    const Indicator = presentation.indicator === 'none' ? null : INDICATOR_ICON[presentation.indicator];
    const label = showProductName ? `PentAGI ${presentation.label}` : presentation.label;

    return (
        <Popover
            onOpenChange={setPopoverOpen}
            open={popoverOpen}
        >
            <TooltipProvider delayDuration={300}>
                <Tooltip
                    onOpenChange={setTooltipOpen}
                    open={tooltipOpen && !popoverOpen}
                >
                    <TooltipTrigger asChild>
                        <PopoverTrigger asChild>
                            <button
                                aria-label={`PentAGI ${presentation.label}. ${presentation.tooltip}`}
                                className={cn(
                                    'text-sidebar-foreground/70 hover:text-sidebar-foreground focus-visible:ring-sidebar-ring data-[state=open]:text-sidebar-foreground inline-flex h-6 max-w-full min-w-0 items-center gap-1 overflow-hidden rounded-md px-1 text-xs outline-hidden transition-colors focus-visible:ring-2',
                                    className,
                                )}
                                type="button"
                            >
                                <span className="truncate group-data-[collapsible=icon]:hidden">{label}</span>
                                {Indicator ? (
                                    <Indicator
                                        aria-hidden
                                        className={cn('size-3.5 shrink-0', INDICATOR_CLASS)}
                                        data-testid={`version-indicator-${presentation.indicator}`}
                                    />
                                ) : (
                                    <Info
                                        aria-hidden
                                        className="hidden size-3.5 shrink-0 group-data-[collapsible=icon]:block"
                                    />
                                )}
                            </button>
                        </PopoverTrigger>
                    </TooltipTrigger>
                    <TooltipContent
                        className="max-w-64"
                        side="right"
                    >
                        {presentation.tooltip}
                    </TooltipContent>
                </Tooltip>
            </TooltipProvider>
            <PopoverContent
                align="start"
                className="w-80 p-0"
                side="right"
            >
                <VersionDetails
                    presentation={presentation}
                    versionInfo={versionInfo}
                />
            </PopoverContent>
        </Popover>
    );
}

function ExternalLinkRow({ children, href }: { children: React.ReactNode; href: string }) {
    return (
        <a
            className="text-foreground hover:bg-accent hover:text-accent-foreground -mx-2 flex items-center justify-between gap-2 rounded-md px-2 py-1"
            href={href}
            rel="noreferrer"
            target="_blank"
        >
            <span className="truncate">{children}</span>
            <ExternalLink
                aria-hidden
                className="text-muted-foreground size-3.5 shrink-0"
            />
        </a>
    );
}

function VersionDetails({
    presentation,
    versionInfo,
}: {
    presentation: VersionPresentation;
    versionInfo: VersionInfoFragmentFragment;
}) {
    const timing = describeTiming(versionInfo);
    const StatusIcon = presentation.indicator === 'none' ? CircleCheck : INDICATOR_ICON[presentation.indicator];
    const statusClass = presentation.indicator === 'none' ? 'text-green-600 dark:text-green-400' : INDICATOR_CLASS;
    const releaseNotes =
        presentation.indicator === 'update' && versionInfo.latest ? releaseNotesUrl(versionInfo.latest) : null;

    return (
        <div className="flex flex-col gap-3 p-4 text-sm">
            <div className="flex items-start justify-between gap-2">
                <div className="min-w-0">
                    <div className="truncate font-semibold">PentAGI {presentation.label}</div>
                    {presentation.revision && (
                        <div className="text-muted-foreground truncate text-xs">build {presentation.revision}</div>
                    )}
                </div>
                <Badge
                    className="shrink-0 capitalize"
                    variant="secondary"
                >
                    {versionInfo.strategy}
                </Badge>
            </div>

            <Separator />

            <div className="flex gap-2">
                <StatusIcon
                    aria-hidden
                    className={cn('mt-0.5 size-4 shrink-0', statusClass)}
                />
                <div className="flex min-w-0 flex-col gap-1">
                    <div className="font-medium">{presentation.headline}</div>
                    <p className="text-muted-foreground">{presentation.detail}</p>
                    {presentation.action && <p className="text-muted-foreground">{presentation.action}</p>}
                    {timing.length > 0 && <p className="text-muted-foreground text-xs">{timing.join(' · ')}</p>}
                </div>
            </div>

            <Separator />

            <div className="flex flex-col gap-1">
                {releaseNotes && (
                    <ExternalLinkRow href={releaseNotes}>Release notes for {presentation.latest}</ExternalLinkRow>
                )}
                <ExternalLinkRow href={PENTAGI_RELEASES_URL}>All releases</ExternalLinkRow>
            </div>
        </div>
    );
}
