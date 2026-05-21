import { FolderInput, Search, X } from 'lucide-react';
import { useCallback, useMemo, useState } from 'react';

import { FileManager, type FileNode } from '@/components/shared/file-manager';
import { OverwriteButtons, OverwriteDialog, useOverwrite } from '@/components/shared/overwrite';
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
import { InputGroup, InputGroupAddon, InputGroupButton, InputGroupInput } from '@/components/ui/input-group';
import { toFileNode } from '@/features/resources/resources-utils';
import { useResources } from '@/providers/resources-provider';

import { findAttachConflicts } from './flow-files-conflicts';
import { useFlowFilesAttachResources } from './use-flow-files-attach-resources';

interface AttachPlan {
    /** GraphQL `UserResource.id` values. Sent to the backend (after numeric coercion). */
    ids: readonly string[];
    /** Library paths used for both client-side preflight and the race-fallback synthesizer. */
    resourcePaths: readonly string[];
}

interface FlowFilesAttachResourcesDialogProps {
    /**
     * Snapshot of the flow's cache used for client-side conflict preflight.
     * Provided by the parent so we don't refetch a list the page already owns.
     */
    cachedFiles: readonly FileNode[];
    flowId: null | string;
    isOpen: boolean;
    onClose: () => void;
    /**
     * Optional UI hook fired after a successful attach. The flow-files Apollo
     * cache itself is updated via the `flowFileAdded` subscription, so callers
     * should NOT use this to drive an imperative refetch.
     */
    onSuccess?: () => void;
}

const EMPTY_SELECTION: ReadonlySet<string> = new Set();

export function FlowFilesAttachResourcesDialog({
    cachedFiles,
    flowId,
    isOpen,
    onClose,
    onSuccess,
}: FlowFilesAttachResourcesDialogProps) {
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
                <FlowFilesAttachResourcesDialogBody
                    cachedFiles={cachedFiles}
                    flowId={flowId}
                    onClose={onClose}
                    onSuccess={onSuccess}
                />
            )}
        </Dialog>
    );
}

function FlowFilesAttachResourcesDialogBody({
    cachedFiles,
    flowId,
    onClose,
    onSuccess,
}: Omit<FlowFilesAttachResourcesDialogProps, 'isOpen'>) {
    const { error: resourcesError, isInitialLoading: isResourcesLoading, resources } = useResources();
    const { attach, isAttaching } = useFlowFilesAttachResources({ flowId });

    const [selectedPaths, setSelectedPaths] = useState<ReadonlySet<string>>(EMPTY_SELECTION);
    const [searchQuery, setSearchQuery] = useState('');

    const files = useMemo<FileNode[]>(() => resources.map(toFileNode), [resources]);

    // Map selected paths back to resource ids on submit. `resource.id` is
    // canonically numeric in the cache (see `resources-rest.ts`) even though
    // codegen types it as `string`; coerce here so downstream callers that
    // rely on the `string` contract stay safe.
    const pathToIdRef = useMemo(() => {
        const map = new Map<string, string>();

        for (const resource of resources) {
            map.set(resource.path, String(resource.id));
        }

        return map;
    }, [resources]);

    const handleSelectionChange = useCallback((next: ReadonlySet<string>) => {
        setSelectedPaths(next);
    }, []);

    /**
     * Drive the canonical "Attach / Attach with overwrite / Replace all"
     * workflow from the shared hook. The plan carries both the IDs (sent to
     * the backend) and the resource paths (used by preflight against the
     * flow's existing cache mirror).
     */
    const overwriteAction = useOverwrite<AttachPlan>({
        execute: async ({ ids }, force) => attach({ ids: [...ids], shouldOverwrite: force }),
        findConflicts: ({ resourcePaths }) => findAttachConflicts(resourcePaths, cachedFiles),
        onSuccess: () => {
            onSuccess?.();
            onClose();
        },
        synthesizeFallbackConflicts: ({ resourcePaths }) =>
            resourcePaths.map((path) => ({
                destination: `resources/${path}`,
                destinationName: path.split('/').pop() ?? path,
            })),
    });

    /**
     * Build the plan for the current selection. Earlier versions deduped
     * descendants of any picked directory on the assumption that the backend
     * would copy directory trees recursively — it does not. Until that
     * becomes recursive on the backend, the user must multi-select a folder
     * together with its children to attach the contents.
     */
    const buildPlan = useCallback((): AttachPlan | null => {
        if (selectedPaths.size === 0) {
            return null;
        }

        const ids: string[] = [];
        const resourcePaths: string[] = [];

        for (const path of selectedPaths) {
            const id = pathToIdRef.get(path);

            if (id) {
                ids.push(id);
                resourcePaths.push(path);
            }
        }

        if (ids.length === 0) {
            return null;
        }

        return { ids, resourcePaths };
    }, [pathToIdRef, selectedPaths]);

    const handlePrimary = useCallback(() => {
        const plan = buildPlan();

        if (plan) {
            void overwriteAction.primaryExecute(plan);
        }
    }, [buildPlan, overwriteAction]);

    const handleOverwrite = useCallback(() => {
        const plan = buildPlan();

        if (plan) {
            void overwriteAction.forceExecute(plan);
        }
    }, [buildPlan, overwriteAction]);

    const selectedCount = selectedPaths.size;
    const hasResources = resources.length > 0;
    const isAttachDisabled = selectedCount === 0;
    const primaryLabel = selectedCount > 0 ? `Attach ${selectedCount}` : 'Attach';
    const overwriteLabel = selectedCount > 0 ? `Attach ${selectedCount} with overwrite` : 'Attach with overwrite';

    const emptyState = (
        <Empty className="border-0">
            <EmptyHeader>
                <EmptyMedia variant="icon">
                    <FolderInput />
                </EmptyMedia>
                <EmptyTitle>Resource library is empty</EmptyTitle>
                <EmptyDescription>Upload resources first to attach them to a flow.</EmptyDescription>
            </EmptyHeader>
        </Empty>
    );

    const noMatchesState = (
        <Empty className="border-0">
            <EmptyHeader>
                <EmptyMedia variant="icon">
                    <Search />
                </EmptyMedia>
                <EmptyTitle>No matches</EmptyTitle>
                <EmptyDescription>
                    No resources match <code>{searchQuery.trim()}</code>.
                </EmptyDescription>
            </EmptyHeader>
        </Empty>
    );

    return (
        <>
            <DialogContent className="flex max-h-[85vh] min-h-[min(85vh,580px)] flex-col gap-4 sm:max-w-3xl">
                <DialogHeader className="text-left">
                    <DialogTitle className="flex items-center gap-2">
                        <FolderInput className="size-4" />
                        Attach resources
                    </DialogTitle>
                    <DialogDescription>
                        Pick files and/or folders from your global library — they will be copied into{' '}
                        <code>resources/</code> of this flow and made available at <code>/work/resources</code> inside
                        the container.
                    </DialogDescription>
                </DialogHeader>

                <div className="flex min-h-0 flex-1 flex-col gap-3">
                    <InputGroup>
                        <InputGroupAddon>
                            <Search />
                        </InputGroupAddon>
                        <InputGroupInput
                            autoComplete="off"
                            disabled={isAttaching || isResourcesLoading}
                            onChange={(event) => setSearchQuery(event.target.value)}
                            placeholder="Search resources..."
                            type="text"
                            value={searchQuery}
                        />
                        {searchQuery && (
                            <InputGroupAddon align="inline-end">
                                <InputGroupButton
                                    onClick={() => setSearchQuery('')}
                                    type="button"
                                >
                                    <X />
                                </InputGroupButton>
                            </InputGroupAddon>
                        )}
                    </InputGroup>

                    {!isResourcesLoading && resourcesError ? (
                        <div className="text-destructive flex flex-1 items-center justify-center rounded-md border p-6 text-center text-sm">
                            {resourcesError.message}
                        </div>
                    ) : (
                        <FileManager
                            className="min-h-0 flex-1"
                            emptyState={emptyState}
                            enableSelection
                            files={files}
                            isLoading={isResourcesLoading}
                            onSelectionChange={handleSelectionChange}
                            search={{ emptyState: noMatchesState, query: searchQuery }}
                        />
                    )}
                </div>

                <DialogFooter className="flex-wrap gap-4 sm:items-center">
                    <span className="text-muted-foreground order-last mr-auto text-xs sm:order-first">
                        {selectedCount > 0
                            ? `${selectedCount} selected`
                            : hasResources
                              ? 'Select one or more items'
                              : ''}
                    </span>
                    <div className="flex flex-col-reverse gap-2 sm:ml-auto sm:flex-row sm:justify-end">
                        <Button
                            disabled={isAttaching}
                            onClick={onClose}
                            type="button"
                            variant="outline"
                        >
                            Cancel
                        </Button>
                        <OverwriteButtons
                            isDisabled={isAttachDisabled}
                            isProcessing={isAttaching}
                            onOverwrite={handleOverwrite}
                            onPrimary={handlePrimary}
                            overwriteLabel={overwriteLabel}
                            primaryIcon={FolderInput}
                            primaryLabel={primaryLabel}
                        />
                    </div>
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

export type { FlowFilesAttachResourcesDialogProps };
