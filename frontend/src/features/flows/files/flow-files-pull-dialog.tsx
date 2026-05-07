import { ArrowDownToLine, ArrowUp, FolderOpen, Loader2, RefreshCw } from 'lucide-react';
import { useCallback, useMemo, useState } from 'react';

import {
    dedupeOverlappingPaths,
    FileManager,
    type FileManagerBulkAction,
    type FileNode,
} from '@/components/shared/file-manager';
import { OverwriteConfirmDialog } from '@/components/shared/overwrite-confirm-dialog';
import { OverwriteCtaButtons } from '@/components/shared/overwrite-cta-buttons';
import { useOverwriteAction } from '@/components/shared/use-overwrite-action';
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
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/ui/tooltip';

import { findPullConflicts } from './flow-files-conflicts';
import { CONTAINER_DEFAULT_PATH } from './flow-files-constants';
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

/**
 * Inner component holding the live browser state. Mounted only while the dialog
 * is open so closing it discards every transient field without an imperative reset.
 *
 * The actual overwrite orchestration (preflight → execute → ConflictDialog
 * fallback) is delegated to {@link useOverwriteAction}; this component only
 * owns the listing browser UI and the per-action plan derivation.
 */
const FlowFilesPullDialogForm = ({ cachedFiles, flowId, onClose, onSuccess }: FlowFilesPullDialogFormProps) => {
    const [currentPath, setCurrentPath] = useState<string>(CONTAINER_DEFAULT_PATH);
    const [pathInputValue, setPathInputValue] = useState<string>(CONTAINER_DEFAULT_PATH);
    const [selectedPaths, setSelectedPaths] = useState<ReadonlySet<string>>(() => new Set<string>());

    // Stable single-element array so the listing hook's effect doesn't re-fire
    // on every parent re-render.
    const listingPaths = useMemo(() => [currentPath], [currentPath]);

    const {
        error: listingError,
        files,
        isLoading: isListingLoading,
        refetch: refetchListing,
    } = useFlowContainerFiles({ flowId, paths: listingPaths });

    const { isPulling, pull } = useFlowFilesPull({
        flowId,
        // Refresh the listing after a successful pull so newly available entries
        // (or, with `force=true`, replaced ones) reflect their fresh state.
        onSuccess: () => {
            void refetchListing();
            onSuccess?.();
        },
    });

    /**
     * Drive the canonical "Pull / Pull with overwrite / Replace all" workflow
     * from the shared hook. The hook owns conflict-state, race-fallback and
     * close-on-success — this dialog just provides the plan (paths) and the
     * three pure helpers (find / execute / synthesize).
     */
    const overwriteAction = useOverwriteAction<readonly string[]>({
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
            navigateTo(dir.path);
        },
        [navigateTo],
    );

    const handleNavigateUp = useCallback(() => {
        navigateTo(getParentContainerPath(currentPath));
    }, [currentPath, navigateTo]);

    const handlePathInputKeyDown = useCallback(
        (event: React.KeyboardEvent<HTMLInputElement>) => {
            if (event.key === 'Enter') {
                event.preventDefault();
                navigateTo(pathInputValue);
            }
        },
        [navigateTo, pathInputValue],
    );

    const handleRefresh = useCallback(() => {
        void refetchListing();
    }, [refetchListing]);

    // Final list of paths to pull. Empty selection → fall back to the directory
    // the user is currently browsing. Non-empty selection wins and is deduped
    // so a folder + one of its descendants don't double-process.
    const pullTargets = useMemo<readonly string[]>(() => {
        if (selectedPaths.size > 0) {
            return dedupeOverlappingPaths(selectedPaths);
        }

        return [currentPath];
    }, [currentPath, selectedPaths]);

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
                        <code>container/</code>. Double-click a folder to drill in.
                    </DialogDescription>
                </DialogHeader>

                <div className="flex flex-col gap-3">
                    <div className="flex items-end gap-2">
                        <div className="flex-1">
                            <Label
                                className="mb-1.5 block text-sm font-normal"
                                htmlFor="flow-files-pull-path"
                            >
                                Container path
                            </Label>
                            <Input
                                autoComplete="off"
                                autoFocus
                                disabled={isPulling}
                                id="flow-files-pull-path"
                                onChange={(event) => setPathInputValue(event.target.value)}
                                onKeyDown={handlePathInputKeyDown}
                                placeholder="/work"
                                value={pathInputValue}
                            />
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

                    <FileManager
                        bulkActions={bulkActions}
                        className="h-[360px]"
                        emptyState={emptyState}
                        enableSelection
                        files={files}
                        isLoading={isListingLoading && files.length === 0}
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
                    <OverwriteCtaButtons
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

            <OverwriteConfirmDialog
                conflicts={overwriteAction.conflicts}
                onCancel={overwriteAction.resetConflicts}
                onReplaceAll={overwriteAction.handleReplaceAll}
            />
        </>
    );
};

export const FlowFilesPullDialog = ({
    cachedFiles,
    flowId,
    isOpen,
    onClose,
    onSuccess,
}: FlowFilesPullDialogProps) => {
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
};
