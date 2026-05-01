import { FolderInput, Loader2, Search, X } from 'lucide-react';
import { useCallback, useMemo, useState } from 'react';

import { FileManager, type FileNode } from '@/components/file-manager';
import { Button } from '@/components/ui/button';
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from '@/components/ui/dialog';
import { Empty, EmptyDescription, EmptyHeader, EmptyMedia, EmptyTitle } from '@/components/ui/empty';
import { InputGroup, InputGroupAddon, InputGroupButton, InputGroupInput } from '@/components/ui/input-group';
import { Label } from '@/components/ui/label';
import { Switch } from '@/components/ui/switch';
import { toFileNode } from '@/features/resources/resources-utils';
import { useResources } from '@/providers/resources-provider';

import { useFlowFilesAttachResources } from './use-flow-files-attach-resources';

interface FlowFilesAttachResourcesDialogProps {
    flowId: null | string;
    isOpen: boolean;
    onClose: () => void;
    onSuccess: () => void;
}

const EMPTY_SELECTION: ReadonlySet<string> = new Set();

const FlowFilesAttachResourcesDialogBody = ({
    flowId,
    onClose,
    onSuccess,
}: Omit<FlowFilesAttachResourcesDialogProps, 'isOpen'>) => {
    const { error: resourcesError, isInitialLoading: isResourcesLoading, resources } = useResources();
    const { attach, isAttaching } = useFlowFilesAttachResources({
        flowId,
        onSuccess: () => {
            onSuccess();
            onClose();
        },
    });

    const [selectedPaths, setSelectedPaths] = useState<ReadonlySet<string>>(EMPTY_SELECTION);
    const [shouldOverwrite, setShouldOverwrite] = useState(false);
    const [searchQuery, setSearchQuery] = useState('');

    const files = useMemo<FileNode[]>(() => resources.map(toFileNode), [resources]);

    // Map selected paths back to resource ids on submit. Built lazily — the lookup
    // is only walked once per attach, so a per-render `Map` allocation is wasteful.
    // `resource.id` is canonically numeric in the cache (see `resources-rest.ts`)
    // even though codegen types it as `string`; coerce here so downstream callers
    // that rely on the `string` contract (`useFlowFilesAttachResources`) stay safe.
    const pathToIdRef = useMemo(() => {
        const map = new Map<string, string>();

        for (const resource of resources) {
            map.set(resource.path, String(resource.id));
        }

        return map;
    }, [resources]);

    const handleSelectionChange = useCallback((next: Set<string>) => {
        setSelectedPaths(next);
    }, []);

    const handleAttach = useCallback(async () => {
        if (selectedPaths.size === 0) {
            return;
        }

        // Forward every explicitly selected path's ID to the backend. Earlier
        // versions deduped descendants of any picked directory on the assumption
        // that the backend would copy directory trees recursively — it does not
        // (`flowfiles.CopyResourcesToFlow` only `MkdirAll`s a directory ref and
        // file refs are copied individually). Until that becomes recursive on
        // the backend, the user must multi-select a folder together with its
        // children to attach the contents.
        const ids: string[] = [];

        for (const path of selectedPaths) {
            const id = pathToIdRef.get(path);

            if (id) {
                ids.push(id);
            }
        }

        if (ids.length === 0) {
            return;
        }

        await attach({ ids, shouldOverwrite });
    }, [attach, pathToIdRef, selectedPaths, shouldOverwrite]);

    const selectedCount = selectedPaths.size;
    const hasResources = resources.length > 0;
    const isAttachDisabled = isAttaching || selectedCount === 0;

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
        <DialogContent className="flex max-h-[85vh] max-w-3xl flex-col gap-4">
            <DialogHeader>
                <DialogTitle className="flex items-center gap-2">
                    <FolderInput className="size-4" />
                    Attach resources from library
                </DialogTitle>
                <DialogDescription>
                    Pick files and/or folders from your global library — they will be copied into{' '}
                    <code>resources/</code> of this flow and made available at <code>/work/resources</code> inside the
                    container.
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
                        className="min-h-[280px] flex-1"
                        emptyState={emptyState}
                        enableSelection
                        files={files}
                        isLoading={isResourcesLoading}
                        onSelectionChange={handleSelectionChange}
                        search={{ emptyState: noMatchesState, query: searchQuery }}
                    />
                )}

                <div className="flex items-center gap-2">
                    <Switch
                        checked={shouldOverwrite}
                        disabled={isAttaching}
                        id="attach-resources-force"
                        onCheckedChange={setShouldOverwrite}
                    />
                    <Label
                        className="cursor-pointer font-normal"
                        htmlFor="attach-resources-force"
                    >
                        Overwrite existing files in flow
                    </Label>
                </div>
            </div>

            <div className="flex items-center justify-between gap-2">
                <span className="text-muted-foreground text-xs">
                    {selectedCount > 0 ? `${selectedCount} selected` : hasResources ? 'Select one or more items' : ''}
                </span>
                <div className="flex gap-2">
                    <Button
                        disabled={isAttaching}
                        onClick={onClose}
                        type="button"
                        variant="outline"
                    >
                        Cancel
                    </Button>
                    <Button
                        disabled={isAttachDisabled}
                        onClick={handleAttach}
                        type="button"
                    >
                        {isAttaching ? <Loader2 className="animate-spin" /> : <FolderInput />}
                        {selectedCount > 0 ? `Attach ${selectedCount}` : 'Attach'}
                    </Button>
                </div>
            </div>
        </DialogContent>
    );
};

export const FlowFilesAttachResourcesDialog = ({
    flowId,
    isOpen,
    onClose,
    onSuccess,
}: FlowFilesAttachResourcesDialogProps) => {
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
                    flowId={flowId}
                    onClose={onClose}
                    onSuccess={onSuccess}
                />
            )}
        </Dialog>
    );
};

export type { FlowFilesAttachResourcesDialogProps };
