import { zodResolver } from '@hookform/resolvers/zod';
import { Copy } from 'lucide-react';
import { useEffect, useMemo } from 'react';
import { useForm } from 'react-hook-form';

import type { FileNode } from '@/components/shared/file-manager';

import { OverwriteCtaButtons } from '@/components/shared/overwrite-cta-buttons';
import { Button } from '@/components/ui/button';
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from '@/components/ui/dialog';
import { Form, FormControl, FormDescription, FormField, FormItem, FormLabel, FormMessage } from '@/components/ui/form';
import { Input } from '@/components/ui/input';

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

/**
 * Resolve the per-file destination from the form value:
 *   - single copy → use the typed path verbatim,
 *   - multi-file batch → derive `<targetDir>/<file.name>`, root when empty.
 */
const resolveDestination = (
    file: FileNode,
    values: ResourcesCopyFormValues,
    isMulti: boolean,
): string => {
    if (!isMulti) {
        return values.destination;
    }

    const targetDir = values.destination.trim().replace(/\/+$/, '');

    return targetDir ? `${targetDir}/${file.name}` : file.name;
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
        defaultValues: { destination: defaultDestination },
        mode: 'onChange',
        resolver: zodResolver(resourcesCopyFormSchema),
    });

    useEffect(() => {
        form.reset({ destination: defaultDestination });
    }, [defaultDestination, form]);

    /**
     * Run every copy in parallel with the given `force` flag. Returns `true`
     * when every operation either succeeded or surfaced as a 409 conflict
     * (the latter feeds into `pendingConflicts` for the aggregated dialog).
     * `false` means at least one entry hit a non-conflict error and the form
     * should stay open so the user can read the toast and retry.
     */
    const submitAll = async (values: ResourcesCopyFormValues, force: boolean): Promise<boolean> => {
        const results = await Promise.all(
            files.map((file) => copy(file.path, { destination: resolveDestination(file, values, isMulti) }, force)),
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
    // failed copy with `force = true`. Close the form so it doesn't leave a stale
    // modal — the resolveConflicts promise resolves once every retry has settled.
    const handleResolveConflicts = async () => {
        await resolveConflicts();
        onClose();
    };

    const isSubmitDisabled = !form.formState.isValid;
    const titleText = isMulti ? `Copy ${files.length} items` : files[0].isDir ? 'Copy directory' : 'Copy resource';
    const overwriteCtaLabel = isMulti ? `Copy ${files.length} with overwrite` : 'Copy with overwrite';

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
                    onSubmit={handleSave}
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

                    <div className="flex flex-wrap justify-end gap-2">
                        <Button
                            disabled={isCopying}
                            onClick={onClose}
                            type="button"
                            variant="outline"
                        >
                            Cancel
                        </Button>
                        <OverwriteCtaButtons
                            isDisabled={isSubmitDisabled}
                            isProcessing={isCopying}
                            onOverwrite={() => {
                                void handleSaveWithOverwrite();
                            }}
                            overwriteLabel={overwriteCtaLabel}
                            primaryIcon={Copy}
                            primaryLabel="Copy"
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
