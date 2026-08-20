import {
    type ColumnDef,
    type ExpandedState,
    flexRender,
    getCoreRowModel,
    getExpandedRowModel,
    type Row,
    useReactTable,
} from '@tanstack/react-table';
import { Check, ChevronDown, ChevronRight, ListFilter, Search, Shield, ShieldAlert, ShieldX, X } from 'lucide-react';
import { Fragment, type ReactElement, useEffect, useMemo, useState } from 'react';
import { useDebouncedCallback } from 'use-debounce';

import type { DeafGuardEventFragmentFragment } from '@/graphql/types';

import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import {
    Command,
    CommandEmpty,
    CommandGroup,
    CommandItem,
    CommandList,
    CommandSeparator,
} from '@/components/ui/command';
import { Empty, EmptyContent, EmptyDescription, EmptyHeader, EmptyMedia, EmptyTitle } from '@/components/ui/empty';
import { InputGroup, InputGroupAddon, InputGroupButton, InputGroupInput } from '@/components/ui/input-group';
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/popover';
import { ScrollArea } from '@/components/ui/scroll-area';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { cn } from '@/lib/utils';
import { useFlow } from '@/providers/flow-provider';

import {
    countByAction,
    countByTier,
    DEAF_GUARD_ACTIONS,
    DEAF_GUARD_TIERS,
    type DeafGuardAction,
    filterEvents,
    getTierBand,
    hasActiveFilters,
    sortNewestFirst,
} from './lib';

// Formatted HH:MM:SS from a Unix-seconds timestamp.
const formatTimestamp = (ts: number): string => {
    const date = new Date(ts * 1000);

    return date.toLocaleTimeString('en-US', { hour: '2-digit', hour12: false, minute: '2-digit', second: '2-digit' });
};

// Maps a severity band to Tailwind row-border + background classes. The left
// border is the main cue, with a faint tinted background for easier scanning.
const ROW_BAND_CLASSES: Record<ReturnType<typeof getTierBand>, string> = {
    destructive: 'border-l-4 border-l-destructive bg-destructive/5',
    neutral: 'border-l-4 border-l-muted-foreground/30',
    warn: 'border-l-4 border-l-amber-500 bg-amber-500/5',
};

// Badge variant for the action cell — visually reinforces severity at a glance.
const actionBadgeVariant = (action: string): 'default' | 'destructive' | 'outline' | 'secondary' => {
    if (action === 'block') {
        return 'destructive';
    }

    if (action === 'warn') {
        return 'secondary';
    }

    return 'outline';
};

// Badge styling for the tier cell, aligned with row banding.
const tierBadgeClasses = (tier: number): string => {
    const band = getTierBand(tier);

    if (band === 'destructive') {
        return 'bg-destructive/10 text-destructive border-destructive/30';
    }

    if (band === 'warn') {
        return 'bg-amber-500/10 text-amber-700 dark:text-amber-400 border-amber-500/40';
    }

    return 'bg-muted text-muted-foreground';
};

const ACTION_ICONS: Record<DeafGuardAction, ReactElement> = {
    block: <ShieldX className="size-3" />,
    log: <Shield className="size-3" />,
    warn: <ShieldAlert className="size-3" />,
};

const SEARCH_DEBOUNCE_MS = 500;

// ---------- Multi-select popover (tier or action) ----------

interface MultiSelectFilterProps<T extends number | string> {
    label: string;
    onClear: () => void;
    onToggle: (value: T) => void;
    options: readonly { label: string; value: T }[];
    selected: ReadonlySet<T>;
}

function MultiSelectFilter<T extends number | string>({
    label,
    onClear,
    onToggle,
    options,
    selected,
}: MultiSelectFilterProps<T>) {
    const [open, setOpen] = useState(false);
    const selectedCount = selected.size;

    return (
        <Popover
            onOpenChange={setOpen}
            open={open}
        >
            <PopoverTrigger asChild>
                <Button
                    className="gap-2"
                    size="sm"
                    variant="outline"
                >
                    <ListFilter className="size-4" />
                    {label}
                    {selectedCount > 0 && (
                        <Badge
                            className="ml-1 px-1.5 py-0 text-[10px]"
                            variant="secondary"
                        >
                            {selectedCount}
                        </Badge>
                    )}
                </Button>
            </PopoverTrigger>
            <PopoverContent
                align="start"
                className="w-48 p-0"
            >
                <Command>
                    <CommandList>
                        <CommandEmpty>No options</CommandEmpty>
                        <CommandGroup>
                            {options.map((option) => {
                                const isSelected = selected.has(option.value);

                                return (
                                    <CommandItem
                                        key={String(option.value)}
                                        onSelect={() => onToggle(option.value)}
                                    >
                                        <div
                                            className={cn(
                                                'border-primary mr-2 flex size-4 items-center justify-center rounded-sm border',
                                                isSelected
                                                    ? 'bg-primary text-primary-foreground'
                                                    : 'opacity-50 [&_svg]:invisible',
                                            )}
                                        >
                                            <Check className="size-3" />
                                        </div>
                                        <span>{option.label}</span>
                                    </CommandItem>
                                );
                            })}
                        </CommandGroup>
                        {selectedCount > 0 && (
                            <>
                                <CommandSeparator />
                                <CommandGroup>
                                    <CommandItem
                                        className="justify-center text-xs"
                                        onSelect={onClear}
                                    >
                                        Clear {label.toLowerCase()}
                                    </CommandItem>
                                </CommandGroup>
                            </>
                        )}
                    </CommandList>
                </Command>
            </PopoverContent>
        </Popover>
    );
}

// ---------- Row expansion sub-component (full reason) ----------

function ReasonDetails({ event }: { event: DeafGuardEventFragmentFragment }) {
    return (
        <div className="bg-muted/40 grid gap-2 p-4 text-sm">
            <div>
                <span className="text-muted-foreground mr-2 text-xs font-medium tracking-wide uppercase">Reason</span>
                <span>{event.reason}</span>
            </div>
            <div className="grid grid-cols-2 gap-2 text-xs md:grid-cols-4">
                <div>
                    <span className="text-muted-foreground mr-1">Category:</span>
                    <span className="font-mono">{event.category}</span>
                </div>
                <div>
                    <span className="text-muted-foreground mr-1">Risk:</span>
                    <span className="font-mono">{event.risk}</span>
                </div>
                <div>
                    <span className="text-muted-foreground mr-1">Mode:</span>
                    <span className="font-mono">{event.mode}</span>
                </div>
                <div>
                    <span className="text-muted-foreground mr-1">Allowed:</span>
                    <span className="font-mono">{event.allowed ? 'true' : 'false'}</span>
                </div>
            </div>
            <div className="text-xs">
                <span className="text-muted-foreground mr-1">Command:</span>
                <code className="bg-background rounded border px-1 py-0.5 font-mono break-all">{event.command}</code>
            </div>
        </div>
    );
}

// ---------- Column definitions ----------

const columns: ColumnDef<DeafGuardEventFragmentFragment>[] = [
    {
        cell: ({ row }) => (
            <div className="flex size-6 items-center justify-center">
                {row.getIsExpanded() ? <ChevronDown className="size-4" /> : <ChevronRight className="size-4" />}
            </div>
        ),
        header: '',
        id: 'expander',
        size: 32,
    },
    {
        accessorFn: (row) => row.timestamp,
        cell: ({ row }) => (
            <span className="text-muted-foreground font-mono text-xs">{formatTimestamp(row.original.timestamp)}</span>
        ),
        header: 'Time',
        id: 'time',
        size: 92,
    },
    {
        accessorFn: (row) => row.tier,
        cell: ({ row }) => (
            <Badge
                className={cn('font-mono', tierBadgeClasses(row.original.tier))}
                variant="outline"
            >
                T{row.original.tier}
            </Badge>
        ),
        header: 'Tier',
        id: 'tier',
        size: 64,
    },
    {
        accessorFn: (row) => row.action,
        cell: ({ row }) => {
            const action = row.original.action as DeafGuardAction;

            return (
                <Badge
                    className="gap-1 tracking-wide uppercase"
                    variant={actionBadgeVariant(action)}
                >
                    {ACTION_ICONS[action]}
                    {row.original.action}
                </Badge>
            );
        },
        header: 'Action',
        id: 'action',
        size: 96,
    },
    {
        accessorFn: (row) => row.allowed,
        cell: ({ row }) => (
            <span
                className={cn(
                    'text-xs font-semibold',
                    row.original.allowed ? 'text-muted-foreground' : 'text-destructive',
                )}
            >
                {row.original.allowed ? 'allowed' : 'blocked'}
            </span>
        ),
        header: 'Outcome',
        id: 'allowed',
        size: 80,
    },
    {
        accessorFn: (row) => row.command,
        cell: ({ row }) => (
            <code
                className="text-muted-foreground block max-w-full truncate font-mono text-xs"
                title={row.original.command}
            >
                {row.original.command}
            </code>
        ),
        header: 'Command',
        id: 'command',
    },
];

// ---------- Main component ----------

function FlowDeafGuard() {
    const { flowData, flowId } = useFlow();

    const events = useMemo(() => flowData?.deafGuardEvents ?? [], [flowData?.deafGuardEvents]);

    const sortedEvents = useMemo(() => sortNewestFirst(events), [events]);

    // Filter state — reset whenever the user navigates between flows.
    const [selectedTiers, setSelectedTiers] = useState<Set<number>>(new Set());
    const [selectedActions, setSelectedActions] = useState<Set<string>>(new Set());
    const [searchInput, setSearchInput] = useState('');
    const [debouncedSearch, setDebouncedSearch] = useState('');
    const [expanded, setExpanded] = useState<ExpandedState>({});

    const debouncedUpdateSearch = useDebouncedCallback((value: string) => {
        setDebouncedSearch(value);
    }, SEARCH_DEBOUNCE_MS);

    useEffect(() => {
        debouncedUpdateSearch(searchInput);

        return () => {
            debouncedUpdateSearch.cancel();
        };
    }, [searchInput, debouncedUpdateSearch]);

    useEffect(() => {
        return () => {
            debouncedUpdateSearch.cancel();
        };
    }, [debouncedUpdateSearch]);

    // Reset filters + expansion state when switching flows — state is purely
    // per-flow, and a stale filter from flow A shouldn't hide events in flow B.
    useEffect(() => {
        setSelectedTiers(new Set());
        setSelectedActions(new Set());
        setSearchInput('');
        setDebouncedSearch('');
        setExpanded({});
        debouncedUpdateSearch.cancel();
    }, [flowId, debouncedUpdateSearch]);

    const filters = useMemo(
        () => ({ actions: selectedActions, search: debouncedSearch, tiers: selectedTiers }),
        [selectedActions, debouncedSearch, selectedTiers],
    );

    const filteredEvents = useMemo(() => filterEvents(sortedEvents, filters), [sortedEvents, filters]);

    // Counters are computed against the full event set so the pills reflect
    // the true distribution regardless of the currently applied filters.
    const actionCounts = useMemo(() => countByAction(events), [events]);
    const tierCounts = useMemo(() => countByTier(events), [events]);

    const table = useReactTable({
        columns,
        data: filteredEvents,
        getCoreRowModel: getCoreRowModel(),
        getExpandedRowModel: getExpandedRowModel(),
        getRowCanExpand: () => true,
        getRowId: (row) => String(row.id),
        onExpandedChange: setExpanded,
        state: { expanded },
    });

    const filtersActive = hasActiveFilters(filters);
    const hasEvents = events.length > 0;
    const hasVisibleEvents = filteredEvents.length > 0;

    const toggleTier = (tier: number) =>
        setSelectedTiers((prev) => {
            const next = new Set(prev);

            if (next.has(tier)) {
                next.delete(tier);
            } else {
                next.add(tier);
            }

            return next;
        });

    const toggleAction = (action: string) =>
        setSelectedActions((prev) => {
            const next = new Set(prev);

            if (next.has(action)) {
                next.delete(action);
            } else {
                next.add(action);
            }

            return next;
        });

    const clearAllFilters = () => {
        setSelectedTiers(new Set());
        setSelectedActions(new Set());
        setSearchInput('');
        setDebouncedSearch('');
        debouncedUpdateSearch.cancel();
    };

    const handleRowClick = (row: Row<DeafGuardEventFragmentFragment>) => row.toggleExpanded();

    return (
        <div className="flex size-full flex-col gap-3">
            {/* Filter bar — sticky above the table */}
            <div className="bg-background sticky top-0 z-10 flex flex-wrap items-center gap-2 pr-4">
                <InputGroup className="min-w-[240px] flex-1">
                    <InputGroupAddon>
                        <Search />
                    </InputGroupAddon>
                    <InputGroupInput
                        autoComplete="off"
                        onChange={(event) => setSearchInput(event.target.value)}
                        placeholder="Search command or reason..."
                        type="text"
                        value={searchInput}
                    />
                    {searchInput && (
                        <InputGroupAddon align="inline-end">
                            <InputGroupButton
                                onClick={() => {
                                    setSearchInput('');
                                    setDebouncedSearch('');
                                    debouncedUpdateSearch.cancel();
                                }}
                                size="icon-xs"
                                title="Clear search"
                                type="button"
                            >
                                <X className="size-4" />
                            </InputGroupButton>
                        </InputGroupAddon>
                    )}
                </InputGroup>
                <MultiSelectFilter
                    label="Tier"
                    onClear={() => setSelectedTiers(new Set())}
                    onToggle={toggleTier}
                    options={DEAF_GUARD_TIERS.map((tier) => ({ label: `Tier ${tier}`, value: tier }))}
                    selected={selectedTiers}
                />
                <MultiSelectFilter
                    label="Action"
                    onClear={() => setSelectedActions(new Set())}
                    onToggle={toggleAction}
                    options={DEAF_GUARD_ACTIONS.map((action) => ({ label: action, value: action }))}
                    selected={selectedActions}
                />
                {filtersActive && (
                    <Button
                        className="gap-1"
                        onClick={clearAllFilters}
                        size="sm"
                        variant="ghost"
                    >
                        <X className="size-4" />
                        Reset
                    </Button>
                )}
            </div>

            {/* Counter pill rows — action counts (required) + tier counts (stretch) */}
            <div
                className="flex flex-col gap-2 pr-4"
                data-testid="deaf-guard-counters"
            >
                <div className="flex flex-wrap items-center gap-2">
                    <span className="text-muted-foreground text-xs font-medium tracking-wide uppercase">By action</span>
                    {DEAF_GUARD_ACTIONS.map((action) => (
                        <Badge
                            className="gap-1 tracking-wide uppercase"
                            data-action={action}
                            data-testid={`count-action-${action}`}
                            key={action}
                            variant={actionBadgeVariant(action)}
                        >
                            {ACTION_ICONS[action]}
                            {action}: {actionCounts[action]}
                        </Badge>
                    ))}
                </div>
                <div className="flex flex-wrap items-center gap-2">
                    <span className="text-muted-foreground text-xs font-medium tracking-wide uppercase">By tier</span>
                    {DEAF_GUARD_TIERS.map((tier) => (
                        <Badge
                            className={cn('font-mono', tierBadgeClasses(tier))}
                            data-testid={`count-tier-${tier}`}
                            key={tier}
                            variant="outline"
                        >
                            T{tier}: {tierCounts[tier]}
                        </Badge>
                    ))}
                </div>
            </div>

            {/* Table / empty states */}
            <ScrollArea className="flex-1 pr-4">
                {!hasEvents ? (
                    <Empty>
                        <EmptyHeader>
                            <EmptyMedia variant="icon">
                                <Shield />
                            </EmptyMedia>
                            <EmptyTitle>No classifications yet</EmptyTitle>
                            <EmptyDescription>
                                Deaf Guard classifies every terminal command the flow executes. Events will stream in
                                here as the flow runs.
                            </EmptyDescription>
                        </EmptyHeader>
                    </Empty>
                ) : !hasVisibleEvents ? (
                    <Empty>
                        <EmptyHeader>
                            <EmptyMedia variant="icon">
                                <ListFilter />
                            </EmptyMedia>
                            <EmptyTitle>No events match your filters</EmptyTitle>
                            <EmptyDescription>Try adjusting the tier, action, or search filters.</EmptyDescription>
                        </EmptyHeader>
                        <EmptyContent>
                            <Button
                                onClick={clearAllFilters}
                                variant="outline"
                            >
                                <X />
                                Reset filters
                            </Button>
                        </EmptyContent>
                    </Empty>
                ) : (
                    <Table>
                        <TableHeader className="bg-background sticky top-0">
                            {table.getHeaderGroups().map((headerGroup) => (
                                <TableRow key={headerGroup.id}>
                                    {headerGroup.headers.map((header) => (
                                        <TableHead
                                            className="text-xs"
                                            key={header.id}
                                            style={
                                                header.column.columnDef.size
                                                    ? { width: header.column.columnDef.size }
                                                    : undefined
                                            }
                                        >
                                            {header.isPlaceholder
                                                ? null
                                                : flexRender(header.column.columnDef.header, header.getContext())}
                                        </TableHead>
                                    ))}
                                </TableRow>
                            ))}
                        </TableHeader>
                        <TableBody>
                            {table.getRowModel().rows.map((row) => (
                                <Fragment key={row.id}>
                                    <TableRow
                                        className={cn(
                                            'hover:bg-muted/50 cursor-pointer',
                                            ROW_BAND_CLASSES[getTierBand(row.original.tier)],
                                        )}
                                        data-testid={`deaf-guard-row-${row.original.id}`}
                                        onClick={() => handleRowClick(row)}
                                    >
                                        {row.getVisibleCells().map((cell) => (
                                            <TableCell
                                                className="py-2"
                                                key={cell.id}
                                            >
                                                {flexRender(cell.column.columnDef.cell, cell.getContext())}
                                            </TableCell>
                                        ))}
                                    </TableRow>
                                    {row.getIsExpanded() && (
                                        <TableRow className="bg-muted/10 hover:bg-muted/10 border-0">
                                            <TableCell
                                                className="p-0"
                                                colSpan={row.getVisibleCells().length}
                                            >
                                                <ReasonDetails event={row.original} />
                                            </TableCell>
                                        </TableRow>
                                    )}
                                </Fragment>
                            ))}
                        </TableBody>
                    </Table>
                )}
            </ScrollArea>
        </div>
    );
}

export default FlowDeafGuard;
