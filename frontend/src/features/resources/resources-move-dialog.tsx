import { zodResolver } from '@hookform/resolvers/zod';
import { FolderInput } from 'lucide-react';
import { useEffect, useMemo } from 'react';
import { useForm } from 'react-hook-form';

import type { FileNode } from '@/components/shared/file-manager';

import { OverwriteCtaButtons } from '@/components/shared/overwrite-cta-buttons';
import { Button } from '@/components/ui/button';
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from '@/components/ui/dialog';
import { Form, FormControl, FormDescription, FormField, FormItem, FormLabel, FormMessage } from '@/components/ui/form';
import { Input } from '@/components/ui/input';

import { ResourcesConflictDialog } from './resources-conflict-dialog';
import { resourcesMoveFormSchema, type ResourcesMoveFormValues, useResourcesMove } from './use-resources-move';

/** Guaranteed non-empty by `ResourcesMoveDialog` (which gates rendering on `files.length > 0`). */
interface ResourcesMoveDialogFormProps {
    files: readonly [FileNode, ...FileNode[]];
    onClose: () => void;
}

interface ResourcesMoveDialogProps {
    /**
     * One or more files to move. Single-element arrays render the "rename or move"
     * UI (full path edit), multi-element arrays render the "move N items into…"
     * UI (destination directory only — every file keeps its current name).
     *
     * Use `null` or an empty array to close the dialog.
     */
    files: FileNode[] | null;
    onClose: () => void;
}

/** Parent directory of a virtual path; `''` for root. Mirrors `getParentDir` from the DnD hook. */
const getParentDir = (path: string): string => {
    const idx = path.lastIndexOf('/');

    return idx === -1 ? '' : path.slice(0, idx);
};

/**
 * Default destination directory for a multi-file move. Use the common parent
 * directory when every selected file lives under the same one (so the user
 * sees "where things came from"); otherwise default to the library root so
 * they don't have to first clear an unrelated path.
 */
const computeCommonParent = (files: readonly [FileNode, ...FileNode[]]): string => {
    const first = getParentDir(files[0].path);

    return files.every((file) => getParentDir(file.path) === first) ? first : '';
};

/**
 * Resolve the per-file destination from the form value:
 *   - single rename / move → use the typed path verbatim,
 *   - multi-file batch     → derive `<targetDir>/<file.name>`, root when empty.
 */
const resolveDestination = (
    file: FileNode,
    values: ResourcesMoveFormValues,
    isMulti: boolean,
): string => {
    if (!isMulti) {
        return values.destination;
    }

    const targetDir = values.destination.trim().replace(/\/+$/, '');

    return targetDir ? `${targetDir}/${file.name}` : file.name;
};

const ResourcesMoveDialogForm = ({ files, onClose }: ResourcesMoveDialogFormProps) => {
    const { cancelConflicts, isMoving, move, pendingConflicts, resolveConflicts } = useResourcesMove();
    const isMulti = files.length > 1;

    // Default destination differs by mode: single-file rename keeps the existing
    // path so the user can edit only the name part; multi-file move pre-fills
    // the common parent directory so a no-op submission is impossible.
    const defaultDestination = useMemo(() => {
        if (isMulti) {
            return computeCommonParent(files);
        }

        return files[0].path;
    }, [files, isMulti]);

    const form = useForm<ResourcesMoveFormValues>({
        defaultValues: { destination: defaultDestination },
        mode: 'onChange',
        resolver: zodResolver(resourcesMoveFormSchema),
    });

    useEffect(() => {
        form.reset({ destination: defaultDestination });
    }, [defaultDestination, form]);

    /**
     * Run every move in parallel with the given `force` flag. Returns `true`
     * when every operation either succeeded or surfaced as a 409 conflict
     * (the latter feeds into `pendingConflicts` for the aggregated dialog).
     * `false` means at least one entry hit a non-conflict error and the form
     * should stay open so the user can read the toast and retry.
     */
    const submitAll = async (values: ResourcesMoveFormValues, force: boolean): Promise<boolean> => {
        const results = await Promise.all(
            files.map((file) => move(file.path, { destination: resolveDestination(file, values, isMulti) }, force)),
        );

        return results.every(Boolean);
    };

    const handleSave = form.handleSubmit(async (values) => {
        const ok = await submitAll(values, false);

        if (ok) {
            onClose();
        }
    });

    const handleSaveWithOverwrite = form.handleSubmit(async (values) => {
        const ok = await submitAll(values, true);

        if (ok) {
            onClose();
        }
    });

    // After the user picks "Replace" in the conflict dialog the hook retries every
    // failed move with `force = true`. Close the form so it doesn't leave a stale
    // modal — the resolveConflicts promise resolves once every retry has settled.
    const handleResolveConflicts = async () => {
        await resolveConflicts();
        onClose();
    };

    const isSubmitDisabled = !form.formState.isValid;
    const titleText = isMulti
        ? `Move ${files.length} items`
        : files[0].isDir
          ? 'Move directory'
          : 'Rename or move resource';
    const overwriteCtaLabel = isMulti ? `Move ${files.length} with overwrite` : 'Move with overwrite';

    return (
        <DialogContent>
            <DialogHeader>
                <DialogTitle className="flex items-center gap-2">
                    <FolderInput className="size-4" />
                    {titleText}
                </DialogTitle>
                <DialogDescription>
                    {isMulti ? (
                        <>Move every selected item into the destination directory.</>
                    ) : (
                        <>
                            Update the path of <code>{files[0].path}</code>.
                        </>
                    )}
                </DialogDescription>
            </DialogHeader>

            <Form {...form}>
                <form
                    className="flex flex-col gap-4"
                    onSubmit={handleSave}
                >
                    <FormField
                        control={form.control}
                        name="destination"
                        render={({ field }) => (
                            <FormItem>
                                <FormLabel>{isMulti ? 'Destination directory' : 'New path'}</FormLabel>
                                <FormControl>
                                    <Input
                                        {...field}
                                        autoComplete="off"
                                        autoFocus
                                        disabled={isMoving}
                                        placeholder={isMulti ? 'Leave empty to move into the library root' : undefined}
                                    />
                                </FormControl>
                                <FormDescription>
                                    {isMulti ? (
                                        <>
                                            Relative directory inside your library. Leave empty for the root. Each item
                                            keeps its current filename.
                                        </>
                                    ) : (
                                        <>
                                            Relative path inside your library. End with <code>/</code> to drop the entry
                                            into that directory.
                                        </>
                                    )}
                                </FormDescription>
                                <FormMessage />
                            </FormItem>
                        )}
                    />

                    <div className="flex flex-wrap justify-end gap-2">
                        <Button
                            disabled={isMoving}
                            onClick={onClose}
                            type="button"
                            variant="outline"
                        >
                            Cancel
                        </Button>
                        <OverwriteCtaButtons
                            isDisabled={isSubmitDisabled}
                            isProcessing={isMoving}
                            onOverwrite={() => {
                                void handleSaveWithOverwrite();
                            }}
                            overwriteLabel={overwriteCtaLabel}
                            primaryIcon={FolderInput}
                            primaryLabel="Move"
                            primaryType="submit"
                        />
                    </div>
                </form>
            </Form>

            <ResourcesConflictDialog
                conflicts={pendingConflicts}
                onCancel={cancelConflicts}
                onReplaceAll={handleResolveConflicts}
            />
        </DialogContent>
    );
};

export const ResourcesMoveDialog = ({ files, onClose }: ResourcesMoveDialogProps) => {
    const handleDialogOpenChange = (nextOpen: boolean) => {
        if (!nextOpen) {
            onClose();
        }
    };

    // Narrow to a non-empty tuple so the inner form can index `files[0]` without
    // optional-chain noise. The Dialog only mounts when this guard passes.
    const nonEmptyFiles = files && files.length > 0 ? (files as [FileNode, ...FileNode[]]) : null;

    return (
        <Dialog
            onOpenChange={handleDialogOpenChange}
            open={!!nonEmptyFiles}
        >
            {nonEmptyFiles && (
                <ResourcesMoveDialogForm
                    files={nonEmptyFiles}
                    onClose={onClose}
                />
            )}
        </Dialog>
    );
};
