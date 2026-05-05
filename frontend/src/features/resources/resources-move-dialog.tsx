import { zodResolver } from '@hookform/resolvers/zod';
import { FolderInput, Loader2 } from 'lucide-react';
import { useEffect, useMemo } from 'react';
import { useForm } from 'react-hook-form';

import type { FileNode } from '@/components/file-manager';

import { Button } from '@/components/ui/button';
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from '@/components/ui/dialog';
import { Form, FormControl, FormDescription, FormField, FormItem, FormLabel, FormMessage } from '@/components/ui/form';
import { Input } from '@/components/ui/input';
import { Switch } from '@/components/ui/switch';

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
        defaultValues: {
            destination: defaultDestination,
            shouldOverwrite: false,
        },
        mode: 'onChange',
        resolver: zodResolver(resourcesMoveFormSchema),
    });

    useEffect(() => {
        form.reset({
            destination: defaultDestination,
            shouldOverwrite: false,
        });
    }, [defaultDestination, form]);

    const handleSubmit = form.handleSubmit(async (values) => {
        if (isMulti) {
            // Multi-file move: `destination` is the *target directory*, every
            // file keeps its name. Strip a trailing `/` for consistency with the
            // single-file branch (which never has one). Empty string = root.
            const targetDir = values.destination.trim().replace(/\/+$/, '');
            const results = await Promise.all(
                files.map((file) => {
                    const destination = targetDir ? `${targetDir}/${file.name}` : file.name;

                    return move(file.path, { destination, shouldOverwrite: values.shouldOverwrite });
                }),
            );

            // Close only when nothing is left over (every move either succeeded
            // or surfaced a conflict that the conflict dialog will pick up).
            // If ANY move actually completed, the bulk bar's selection paths
            // are stale anyway — the host clears them after onSelect resolves.
            if (results.every(Boolean)) {
                onClose();
            }

            return;
        }

        const wasMoved = await move(files[0].path, values);

        if (wasMoved) {
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

    const isSubmitDisabled = !form.formState.isValid || isMoving;

    const titleText = isMulti
        ? `Move ${files.length} items`
        : files[0].isDir
          ? 'Move directory'
          : 'Rename or move resource';

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
                    onSubmit={handleSubmit}
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

                    <FormField
                        control={form.control}
                        name="shouldOverwrite"
                        render={({ field }) => (
                            <FormItem className="flex flex-row items-center gap-2">
                                <FormControl>
                                    <Switch
                                        checked={field.value}
                                        disabled={isMoving}
                                        onCheckedChange={field.onChange}
                                    />
                                </FormControl>
                                <FormLabel className="cursor-pointer font-normal">
                                    {isMulti
                                        ? 'Overwrite if a resource already exists at any destination'
                                        : 'Overwrite if a resource already exists at the destination'}
                                </FormLabel>
                            </FormItem>
                        )}
                    />

                    <div className="flex justify-end gap-2">
                        <Button
                            disabled={isMoving}
                            onClick={onClose}
                            type="button"
                            variant="outline"
                        >
                            Cancel
                        </Button>
                        <Button
                            disabled={isSubmitDisabled}
                            type="submit"
                        >
                            {isMoving ? <Loader2 className="animate-spin" /> : <FolderInput />}
                            Move
                        </Button>
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
