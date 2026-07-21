import { Copy } from 'lucide-react';
import { useEffect, useMemo } from 'react';

import type { FileNode } from '@/components/shared/file-manager';
import type { OverwriteConflict } from '@/components/shared/overwrite';

import { OverwriteButtons, OverwriteDialog, useOverwrite } from '@/components/shared/overwrite';
import { Button } from '@/components/ui/button';
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from '@/components/ui/dialog';
import { Form, FormControl, FormDescription, FormField, FormItem, FormLabel, FormMessage } from '@/components/ui/form';
import { Input } from '@/components/ui/input';
import { useAppForm } from '@/hooks/use-app-form';
import { useResources } from '@/providers/resources-provider';

import { resourcesCopyFormSchema, type ResourcesCopyFormValues, useResourcesCopy } from './use-resources-copy';

interface CopyPlan {
    /** Final destination string sent to the backend (exact path or base directory). */
    destination: string;
    /** Resource paths being copied (sent as `sources[]`). */
    sources: readonly string[];
    /** Pre-computed `(destination, destinationName)` pairs for client-side preflight + 409 fallback. */
    targets: OverwriteConflict[];
}

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

const splitName = (path: string): string => path.split('/').pop() ?? path;

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
 * Pre-compute the per-file destinations the backend will write to. Mirrors the
 * server's resolution rules so the client can preflight and so the conflict
 * dialog can name the exact items at risk. See {@link computeTargets} in the
 * move dialog for the full rule table.
 */
const computeTargets = (files: readonly [FileNode, ...FileNode[]], destination: string): OverwriteConflict[] => {
    const trimmed = destination.trim();
    const treatAsDir = files.length > 1 || (trimmed.length > 1 && trimmed.endsWith('/'));

    if (treatAsDir) {
        const baseDir = trimmed.replace(/\/+$/, '');

        return files.map((file) => {
            const dest = baseDir ? `${baseDir}/${file.name}` : file.name;

            return { destination: dest, destinationName: file.name };
        });
    }

    return [{ destination: trimmed, destinationName: splitName(trimmed) }];
};

const buildCopyPlan = (files: readonly [FileNode, ...FileNode[]], values: ResourcesCopyFormValues): CopyPlan => ({
    destination: values.destination.trim(),
    sources: files.map((file) => file.path),
    targets: computeTargets(files, values.destination),
});

export function ResourcesCopyDialog({ files, onClose }: ResourcesCopyDialogProps) {
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
}

function ResourcesCopyDialogForm({ files, onClose }: ResourcesCopyDialogFormProps) {
    const { copy, isCopying } = useResourcesCopy();
    const { resources } = useResources();
    const isMulti = files.length > 1;

    const defaultDestination = useMemo(() => {
        if (isMulti) {
            return computeCommonParent(files);
        }

        return buildSingleDefaultDestination(files[0]);
    }, [files, isMulti]);

    const form = useAppForm<ResourcesCopyFormValues>({
        defaultValues: { destination: defaultDestination },
        schema: resourcesCopyFormSchema,
    });

    useEffect(() => {
        form.reset({ destination: defaultDestination });
    }, [defaultDestination, form]);

    const resourcePaths = useMemo(() => new Set(resources.map((resource) => resource.path)), [resources]);

    /**
     * Drive the canonical "Copy / Copy with overwrite / Replace all" workflow
     * with a single atomic batch request. Backend handles `sources[]` in one
     * DB transaction (all-or-nothing).
     */
    const overwriteAction = useOverwrite<CopyPlan>({
        execute: (plan, force) => copy(plan.sources, plan.destination, force),
        // Copy never deletes the sources, so collisions with sources are real
        // conflicts (unlike move). Just intersect targets with existing paths.
        findConflicts: (plan) => plan.targets.filter((t) => resourcePaths.has(t.destination)),
        onSuccess: onClose,
        synthesizeFallbackConflicts: (plan) => plan.targets,
    });

    const handleSave = form.handleSubmit(async (values) => {
        await overwriteAction.primaryExecute(buildCopyPlan(files, values));
    });

    const handleSaveWithOverwrite = form.handleSubmit(async (values) => {
        await overwriteAction.forceExecute(buildCopyPlan(files, values));
    });

    // Convention: stay enabled until the first submit, then reflect validity (so an invalid submit surfaces
    // errors instead of a silently-dead button). Mirrors FormSubmitButton's requireValid gate.
    const isSubmitDisabled = form.formState.isSubmitted && !form.formState.isValid;
    const titleText = isMulti ? `Copy ${files.length} items` : files[0].isDir ? 'Copy directory' : 'Copy resource';
    const overwriteCtaLabel = isMulti ? `Copy ${files.length} with overwrite` : 'Copy with overwrite';

    return (
        <>
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
                        noValidate
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
                                            placeholder={
                                                isMulti ? 'Leave empty to copy into the library root' : undefined
                                            }
                                        />
                                    </FormControl>
                                    <FormDescription>
                                        {isMulti ? (
                                            <>
                                                Relative directory inside your library. Leave empty for the root. Each
                                                item keeps its current filename.
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
                            <OverwriteButtons
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
            </DialogContent>

            <OverwriteDialog
                conflicts={overwriteAction.conflicts}
                onCancel={overwriteAction.resetConflicts}
                onReplaceAll={overwriteAction.handleReplaceAll}
            />
        </>
    );
}
