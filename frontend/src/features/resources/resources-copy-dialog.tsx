import { zodResolver } from '@hookform/resolvers/zod';
import { Copy, Loader2 } from 'lucide-react';
import { useEffect, useMemo } from 'react';
import { useForm } from 'react-hook-form';

import type { FileNode } from '@/components/file-manager';

import { Button } from '@/components/ui/button';
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from '@/components/ui/dialog';
import { Form, FormControl, FormDescription, FormField, FormItem, FormLabel, FormMessage } from '@/components/ui/form';
import { Input } from '@/components/ui/input';
import { Switch } from '@/components/ui/switch';

import { ResourcesConflictDialog } from './resources-conflict-dialog';
import { resourcesCopyFormSchema, type ResourcesCopyFormValues, useResourcesCopy } from './use-resources-copy';

/** Guaranteed non-empty by `ResourcesCopyDialog` (which gates rendering on `files.length > 0`). */
interface ResourcesCopyDialogFormProps {
    files: readonly [FileNode, ...FileNode[]];
    onClose: () => void;
}

interface ResourcesCopyDialogProps {
    /**
     * One or more files to copy. Single-element arrays render the "duplicate to a
     * new path" UI (full path edit with a `-copy` suffix), multi-element arrays
     * render the "copy N items into…" UI (destination directory only — every file
     * keeps its current name).
     *
     * Use `null` or an empty array to close the dialog.
     */
    files: FileNode[] | null;
    onClose: () => void;
}

/** Parent directory of a virtual path; `''` for root. */
const getParentDir = (path: string): string => {
    const idx = path.lastIndexOf('/');

    return idx === -1 ? '' : path.slice(0, idx);
};

/**
 * Build the single-file copy default destination. Inserts a `-copy` suffix
 * before the extension so the user can submit immediately without manual
 * editing — same convention as Finder's "Duplicate".
 */
const buildSingleDefaultDestination = (file: FileNode): string => {
    const segments = file.path.split('/');
    const lastSegment = segments.at(-1) ?? file.name;
    const parent = segments.slice(0, -1).join('/');
    const dotIndex = lastSegment.lastIndexOf('.');
    const baseName = file.isDir || dotIndex === -1 ? lastSegment : lastSegment.slice(0, dotIndex);
    const extension = file.isDir || dotIndex === -1 ? '' : lastSegment.slice(dotIndex);
    const candidateName = `${baseName}-copy${extension}`;

    return parent ? `${parent}/${candidateName}` : candidateName;
};

/**
 * Default destination directory for a multi-file copy: the common parent if
 * every selection lives under the same one, otherwise the library root.
 * Mirrors `ResourcesMoveDialog`'s logic.
 */
const computeCommonParent = (files: readonly [FileNode, ...FileNode[]]): string => {
    const first = getParentDir(files[0].path);

    return files.every((file) => getParentDir(file.path) === first) ? first : '';
};

const ResourcesCopyDialogForm = ({ files, onClose }: ResourcesCopyDialogFormProps) => {
    const { cancelConflicts, copy, isCopying, pendingConflicts, resolveConflicts } = useResourcesCopy();
    const isMulti = files.length > 1;

    const defaultDestination = useMemo(() => {
        if (isMulti) {
            return computeCommonParent(files);
        }

        return buildSingleDefaultDestination(files[0]);
    }, [files, isMulti]);

    const form = useForm<ResourcesCopyFormValues>({
        defaultValues: {
            destination: defaultDestination,
            shouldOverwrite: false,
        },
        mode: 'onChange',
        resolver: zodResolver(resourcesCopyFormSchema),
    });

    useEffect(() => {
        form.reset({
            destination: defaultDestination,
            shouldOverwrite: false,
        });
    }, [defaultDestination, form]);

    const handleSubmit = form.handleSubmit(async (values) => {
        if (isMulti) {
            const targetDir = values.destination.trim().replace(/\/+$/, '');
            const results = await Promise.all(
                files.map((file) => {
                    const destination = targetDir ? `${targetDir}/${file.name}` : file.name;

                    return copy(file.path, { destination, shouldOverwrite: values.shouldOverwrite });
                }),
            );

            if (results.every(Boolean)) {
                onClose();
            }

            return;
        }

        const wasCopied = await copy(files[0].path, values);

        if (wasCopied) {
            onClose();
        }
    });

    // After the user picks "Replace" in the conflict dialog the hook retries every
    // failed copy with `force = true`. Close the form so it doesn't leave a stale
    // modal — the resolveConflicts promise resolves once every retry has settled.
    const handleResolveConflicts = async () => {
        await resolveConflicts();
        onClose();
    };

    const isSubmitDisabled = !form.formState.isValid || isCopying;

    const titleText = isMulti ? `Copy ${files.length} items` : files[0].isDir ? 'Copy directory' : 'Copy resource';

    return (
        <DialogContent>
            <DialogHeader>
                <DialogTitle className="flex items-center gap-2">
                    <Copy className="size-4" />
                    {titleText}
                </DialogTitle>
                <DialogDescription>
                    {isMulti ? (
                        <>Duplicate every selected item into the destination directory.</>
                    ) : (
                        <>
                            Duplicate <code>{files[0].path}</code> to a new path.
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
                                <FormLabel>{isMulti ? 'Destination directory' : 'Destination path'}</FormLabel>
                                <FormControl>
                                    <Input
                                        {...field}
                                        autoComplete="off"
                                        autoFocus
                                        disabled={isCopying}
                                        placeholder={isMulti ? 'Leave empty to copy into the library root' : undefined}
                                    />
                                </FormControl>
                                <FormDescription>
                                    {isMulti ? (
                                        <>
                                            Relative directory inside your library. Leave empty for the root. Each item
                                            keeps its current filename.
                                        </>
                                    ) : (
                                        <>Relative path inside your library.</>
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
                                        disabled={isCopying}
                                        onCheckedChange={field.onChange}
                                    />
                                </FormControl>
                                <FormLabel className="cursor-pointer font-normal">
                                    {isMulti
                                        ? 'Overwrite or merge if any destination already exists'
                                        : 'Overwrite or merge if the destination already exists'}
                                </FormLabel>
                            </FormItem>
                        )}
                    />

                    <div className="flex justify-end gap-2">
                        <Button
                            disabled={isCopying}
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
                            {isCopying ? <Loader2 className="animate-spin" /> : <Copy />}
                            Copy
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

export const ResourcesCopyDialog = ({ files, onClose }: ResourcesCopyDialogProps) => {
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
                <ResourcesCopyDialogForm
                    files={nonEmptyFiles}
                    onClose={onClose}
                />
            )}
        </Dialog>
    );
};
