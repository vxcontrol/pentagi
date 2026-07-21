import { ArrowDownToLine, ArrowUp, FolderOpen, Loader2, RefreshCw, TriangleAlert } from 'lucide-react';
import { useCallback, useMemo, useState } from 'react';

import {
    dedupeOverlappingPaths,
    FileManager,
    type FileManagerBulkAction,
    type FileNode,
} from '@/components/shared/file-manager';
import { OverwriteButtons, OverwriteDialog, useOverwrite } from '@/components/shared/overwrite';
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert';
import {
    Autocomplete,
    AutocompleteContent,
    AutocompleteEmpty,
    AutocompleteGroup,
    AutocompleteInput,
    AutocompleteItem,
} from '@/components/ui/autocomplete';
import { Button } from '@/components/ui/button';
import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogFooter,
    DialogHeader,
    DialogTitle,
} from '@/components/ui/dialog';
import { Empty, EmptyDescription, EmptyHeader, EmptyMedia, EmptyTitle } from '@/components/ui/empty';
import { Label } from '@/components/ui/label';
import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/ui/tooltip';

import { findPullConflicts } from './flow-files-conflicts';
import { CONTAINER_DEFAULT_PATH, CONTAINER_PATH_PREFIX } from './flow-files-constants';
import { pluralizeItems } from './flow-files-utils';
import { useFlowContainerFiles } from './use-flow-container-files';
import { useFlowFilesPull } from './use-flow-files-pull';

interface FlowFilesPullDialogFormProps {
    cachedFiles: readonly FileNode[];
    flowId: null | string;
    onClose: () => void;
    /**
     * Optional UI hook fired after a successful pull. The flow-files Apollo
     * cache itself is updated via the `flowFileAdded` subscription, so callers
     * should NOT use this to drive an imperative refetch.
     */
    onSuccess?: () => void;
}

interface FlowFilesPullDialogProps {
    cachedFiles: readonly FileNode[];
    flowId: null | string;
    isOpen: boolean;
    onClose: () => void;
    /** See {@link FlowFilesPullDialogFormProps.onSuccess}. */
    onSuccess?: () => void;
}

/**
 * Normalise a user-entered container path:
 *   - trim whitespace,
 *   - convert empty / "" to root "/",
 *   - guarantee a leading slash so the backend treats it as absolute,
 *   - strip a trailing slash for everything except root "/" itself.
 */
const normalizeContainerPath = (raw: string): string => {
    const trimmed = raw.trim();

    if (trimmed === '' || trimmed === '/') {
        return '/';
    }

    const withSlash = trimmed.startsWith('/') ? trimmed : `/${trimmed}`;

    return withSlash.length > 1 ? withSlash.replace(/\/+$/, '') : withSlash;
};

/** Parent of an absolute container path; root collapses onto itself. */
const getParentContainerPath = (path: string): string => {
    if (path === '/' || path === '') {
        return '/';
    }

    const idx = path.lastIndexOf('/');

    if (idx <= 0) {
        return '/';
    }

    return path.slice(0, idx);
};

export function FlowFilesPullDialog({ cachedFiles, flowId, isOpen, onClose, onSuccess }: FlowFilesPullDialogProps) {
    const handleDialogOpenChange = (nextOpen: boolean) => {
        if (!nextOpen) {
            onClose();
        }
    };

    return (
        <Dialog
            onOpenChange={handleDialogOpenChange}
            open={isOpen}
        >
            {isOpen && (
                <FlowFilesPullDialogForm
                    cachedFiles={cachedFiles}
                    flowId={flowId}
                    onClose={onClose}
                    onSuccess={onSuccess}
                />
            )}
        </Dialog>
    );
}

/**
 * Inner component holding the live browser state. Mounted only while the dialog
 * is open so closing it discards every transient field without an imperative reset.
 *
 * The actual overwrite orchestration (preflight → execute → ConflictDialog
 * fallback) is delegated to {@link useOverwrite}; this component only
 * owns the listing browser UI and the per-action plan derivation.
 */
function FlowFilesPullDialogForm({ cachedFiles, flowId, onClose, onSuccess }: FlowFilesPullDialogFormProps) {
    const [currentPath, setCurrentPath] = useState<string>(CONTAINER_DEFAULT_PATH);
    const [pathInputValue, setPathInputValue] = useState<string>(CONTAINER_DEFAULT_PATH);
    const [selectedPaths, setSelectedPaths] = useState<ReadonlySet<string>>(() => new Set<string>());

    // Stable single-element array so the listing hook's effect doesn't re-fire
    // on every parent re-render.
    const listingPaths = useMemo(() => [currentPath], [currentPath]);

    const {
        error: listingError,
        failures: listingFailures,
        files,
        isLoading: isListingLoading,
        refetch: refetchListing,
        truncated: isListingTruncated,
    } = useFlowContainerFiles({ flowId, paths: listingPaths });

    /**
     * Feed the FileManager a flat single-level view of the current directory.
     *
     * The container endpoint returns absolute paths (`/work/foo.txt`) and
     * passing those straight in would have `buildFileManagerTree` synthesise
     * placeholder parent folders for every leading segment (e.g. a collapsed
     * `work/` wrapper around the actual entries). That wrapper makes the
     * navigation-style chevron / double-click drill-in feel broken — the
     * chevron of the synthetic root just toggles a wrapper that has no
     * meaningful navigation target ("we are already there").
     *
     * Workaround: expose `name` as `path` (so every entry is a top-level
     * sibling) and stash the absolute container path inside `id`. The dialog
     * uses `id` for navigation / pull, the FileManager uses `path` for
     * selection / row keys / focus management — the two stay in sync because
     * directory listings always have unique entry names.
     */
    const flatFiles = useMemo<FileNode[]>(
        () => files.map((file) => ({ ...file, id: file.path, path: file.name })),
        [files],
    );

    /**
     * Reverse lookup `name → absolute container path`, used to map the
     * FileManager's name-keyed selection back to the absolute paths the
     * backend's pull endpoint expects.
     */
    const nameToAbsolutePath = useMemo(() => {
        const map = new Map<string, string>();

        for (const file of files) {
            map.set(file.name, file.path);
        }

        return map;
    }, [files]);

    const { isPulling, pull } = useFlowFilesPull({
        flowId,
        // Refresh the listing after a successful pull so newly available entries
        // (or, with `force=true`, replaced ones) reflect their fresh state.
        onSuccess: () => {
            void refetchListing();
            onSuccess?.();
        },
    });

    const overwriteAction = useOverwrite<readonly string[]>({
        execute: (paths, force) => pull(paths, force),
        findConflicts: (paths) => findPullConflicts(paths, cachedFiles),
        onSuccess: onClose,
        synthesizeFallbackConflicts: (paths) =>
            paths.map((path) => ({
                destination: path,
                destinationName: path.split('/').pop() ?? path,
            })),
    });

    const navigateTo = useCallback((nextPath: string) => {
        const normalized = normalizeContainerPath(nextPath);

        setCurrentPath(normalized);
        setPathInputValue(normalized);
        setSelectedPaths(new Set<string>());
    }, []);

    const handleOpenDirectory = useCallback(
        (dir: FileNode) => {
            // `dir.id` is the absolute container path (we flattened the
            // listing into `flatFiles` for the FileManager). `dir.path` is
            // just the entry's name in this dialog and would normalise to
            // a wrong absolute path (`/${name}`) if we used it directly.
            navigateTo(dir.id);
        },
        [navigateTo],
    );

    const handleNavigateUp = useCallback(() => {
        navigateTo(getParentContainerPath(currentPath));
    }, [currentPath, navigateTo]);

    const handleRefresh = useCallback(() => {
        void refetchListing();
    }, [refetchListing]);

    /**
     * Suggestions surfaced under the path input. We collect every container
     * directory the user has already touched: ancestors of cached files (those
     * sit under the synthetic `container/` cache prefix and need to be
     * unwrapped back into absolute container paths) plus directories from the
     * directory currently being browsed. Files are mapped to their parent
     * directory — pulling a file path through the address bar would only
     * produce a listing error.
     *
     * Seeded with `/` and `CONTAINER_DEFAULT_PATH` so the dropdown is useful
     * the first time the dialog is opened, before anything is cached.
     */
    const pathSuggestions = useMemo<readonly string[]>(() => {
        const paths = new Set<string>();
        const containerCachePrefix = `${CONTAINER_PATH_PREFIX}/`;

        const addWithAncestors = (dirPath: string) => {
            let current = dirPath;

            while (current && current !== '/') {
                paths.add(current);
                current = getParentContainerPath(current);
            }

            paths.add('/');
        };

        paths.add('/');
        paths.add(CONTAINER_DEFAULT_PATH);
        addWithAncestors(currentPath);

        for (const file of cachedFiles) {
            if (!file.path.startsWith(containerCachePrefix)) {
                continue;
            }

            const stripped = file.path.slice(CONTAINER_PATH_PREFIX.length);
            const containerPath = stripped.startsWith('/') ? stripped : `/${stripped}`;
            const dir = file.isDir ? containerPath : getParentContainerPath(containerPath);

            addWithAncestors(dir);
        }

        for (const file of files) {
            const dir = file.isDir ? file.path : getParentContainerPath(file.path);

            addWithAncestors(dir);
        }

        return [...paths].sort((a, b) => a.localeCompare(b));
    }, [cachedFiles, currentPath, files]);

    // Final list of paths to pull. Empty selection → fall back to the directory
    // the user is currently browsing. Non-empty selection wins and is mapped
    // back from the FileManager's name-keyed selection to absolute container
    // paths, then deduped so a folder + one of its descendants don't
    // double-process.
    const pullTargets = useMemo<readonly string[]>(() => {
        if (selectedPaths.size === 0) {
            return [currentPath];
        }

        const absolutePaths: string[] = [];

        for (const name of selectedPaths) {
            const absolute = nameToAbsolutePath.get(name);

            if (absolute) {
                absolutePaths.push(absolute);
            }
        }

        return dedupeOverlappingPaths(absolutePaths);
    }, [currentPath, nameToAbsolutePath, selectedPaths]);

    const isUpDisabled = currentPath === '/' || isListingLoading || isPulling;
    const isPullDisabled = isListingLoading || pullTargets.length === 0 || !flowId;

    const primaryLabel = useMemo(() => {
        if (selectedPaths.size === 0) {
            return `Pull ${currentPath}`;
        }

        return `Pull ${selectedPaths.size} ${pluralizeItems(selectedPaths.size)}`;
    }, [currentPath, selectedPaths.size]);

    const overwriteLabel = useMemo(() => {
        if (selectedPaths.size === 0) {
            return 'Pull with overwrite';
        }

        return `Pull ${selectedPaths.size} with overwrite`;
    }, [selectedPaths.size]);

    // The FileManager doesn't ship a "selection only" mode — passing an empty
    // bulk-actions array is the cheapest way to surface the checkboxes.
    const bulkActions = useMemo<FileManagerBulkAction[]>(() => [], []);

    const emptyState = listingError ? (
        <Empty>
            <EmptyHeader>
                <EmptyMedia variant="icon">
                    <FolderOpen />
                </EmptyMedia>
                <EmptyTitle>Failed to list container</EmptyTitle>
                <EmptyDescription>{listingError.message}</EmptyDescription>
            </EmptyHeader>
        </Empty>
    ) : listingFailures.length > 0 ? (
        <Empty>
            <EmptyHeader>
                <EmptyMedia variant="icon">
                    <TriangleAlert />
                </EmptyMedia>
                <EmptyTitle>Nothing readable here</EmptyTitle>
                <EmptyDescription>
                    None of the {listingFailures.length} {listingFailures.length === 1 ? 'entry' : 'entries'} in{' '}
                    <code>{currentPath}</code> could be read.
                </EmptyDescription>
            </EmptyHeader>
            <ul className="text-muted-foreground max-w-full space-y-1 px-4 text-left text-xs">
                {listingFailures.slice(0, 5).map((failure) => (
                    <li
                        className="truncate"
                        key={failure.path}
                    >
                        <span className="text-foreground font-medium">{failure.name}</span> — {failure.message}
                    </li>
                ))}
                {listingFailures.length > 5 && <li>…and {listingFailures.length - 5} more</li>}
            </ul>
        </Empty>
    ) : (
        <Empty>
            <EmptyHeader>
                <EmptyMedia variant="icon">
                    <FolderOpen />
                </EmptyMedia>
                <EmptyTitle>Directory is empty</EmptyTitle>
                <EmptyDescription>
                    Nothing to pull from <code>{currentPath}</code>.
                </EmptyDescription>
            </EmptyHeader>
        </Empty>
    );

    return (
        <>
            <DialogContent className="sm:max-w-2xl">
                <DialogHeader>
                    <DialogTitle className="flex items-center gap-2">
                        <ArrowDownToLine className="size-4" />
                        Pull from container
                    </DialogTitle>
                    <DialogDescription>
                        Browse the running container and select files or directories to sync into the local cache under{' '}
                        <code>container/</code>. Click the arrow on a folder row or double-click the row to drill in.
                    </DialogDescription>
                </DialogHeader>

                <div className="flex flex-col gap-3">
                    <div className="flex items-end gap-2">
                        <div className="flex-1">
                            <Label className="mb-1.5 block text-sm font-normal">Container path</Label>
                            <Autocomplete
                                onCommit={navigateTo}
                                onValueChange={setPathInputValue}
                                value={pathInputValue}
                            >
                                <AutocompleteInput
                                    autoFocus
                                    disabled={isPulling}
                                    placeholder="/work"
                                />
                                <AutocompleteContent>
                                    <AutocompleteEmpty>No matching paths</AutocompleteEmpty>
                                    <AutocompleteGroup>
                                        {pathSuggestions.map((suggestion) => (
                                            <AutocompleteItem
                                                key={suggestion}
                                                value={suggestion}
                                            >
                                                {suggestion}
                                            </AutocompleteItem>
                                        ))}
                                    </AutocompleteGroup>
                                </AutocompleteContent>
                            </Autocomplete>
                        </div>

                        <Tooltip>
                            <TooltipTrigger asChild>
                                <span>
                                    <Button
                                        disabled={isUpDisabled}
                                        onClick={handleNavigateUp}
                                        size="icon-sm"
                                        type="button"
                                        variant="outline"
                                    >
                                        <ArrowUp />
                                    </Button>
                                </span>
                            </TooltipTrigger>
                            <TooltipContent>Parent directory</TooltipContent>
                        </Tooltip>

                        <Tooltip>
                            <TooltipTrigger asChild>
                                <span>
                                    <Button
                                        disabled={isListingLoading || isPulling}
                                        onClick={handleRefresh}
                                        size="icon-sm"
                                        type="button"
                                        variant="outline"
                                    >
                                        {isListingLoading ? <Loader2 className="animate-spin" /> : <RefreshCw />}
                                    </Button>
                                </span>
                            </TooltipTrigger>
                            <TooltipContent>Refresh listing</TooltipContent>
                        </Tooltip>
                    </div>

                    {listingFailures.length > 0 && flatFiles.length > 0 && (
                        <Alert>
                            <TriangleAlert />
                            <AlertTitle>
                                {listingFailures.length} {listingFailures.length === 1 ? 'entry' : 'entries'} could not
                                be read
                            </AlertTitle>
                            <AlertDescription>
                                The readable entries are shown below. Skipped:{' '}
                                {listingFailures
                                    .slice(0, 5)
                                    .map((failure) => failure.name)
                                    .join(', ')}
                                {listingFailures.length > 5 ? `, and ${listingFailures.length - 5} more` : ''}.
                            </AlertDescription>
                        </Alert>
                    )}

                    {isListingTruncated && (
                        <Alert>
                            <TriangleAlert />
                            <AlertTitle>Directory truncated</AlertTitle>
                            <AlertDescription>
                                This directory has too many entries to list in full; only the first {files.length} are
                                shown. Open a subfolder to see the rest.
                            </AlertDescription>
                        </Alert>
                    )}

                    <FileManager
                        bulkActions={bulkActions}
                        className="h-[360px]"
                        emptyState={emptyState}
                        enableSelection
                        files={flatFiles}
                        isLoading={isListingLoading && flatFiles.length === 0}
                        onOpenDirectory={handleOpenDirectory}
                        onSelectionChange={setSelectedPaths}
                    />
                </div>

                <DialogFooter>
                    <Button
                        disabled={isPulling}
                        onClick={onClose}
                        type="button"
                        variant="outline"
                    >
                        Cancel
                    </Button>
                    <OverwriteButtons
                        isDisabled={isPullDisabled}
                        isProcessing={isPulling}
                        onOverwrite={() => {
                            void overwriteAction.forceExecute(pullTargets);
                        }}
                        onPrimary={() => {
                            void overwriteAction.primaryExecute(pullTargets);
                        }}
                        overwriteLabel={overwriteLabel}
                        primaryIcon={ArrowDownToLine}
                        primaryLabel={primaryLabel}
                    />
                </DialogFooter>
            </DialogContent>

            <OverwriteDialog
                conflicts={overwriteAction.conflicts}
                onCancel={overwriteAction.resetConflicts}
                onReplaceAll={overwriteAction.handleReplaceAll}
            />
        </>
    );
}
