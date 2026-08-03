import { FolderInput } from 'lucide-react';
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

import { resourcesMoveFormSchema, type ResourcesMoveFormValues, useResourcesMove } from './use-resources-move';

interface MovePlan {
    /** Final destination string sent to the backend (exact path or base directory). */
    destination: string;
    /** Resource paths being moved (sent as `sources[]`). */
    sources: readonly string[];
    /** Pre-computed `(destination, destinationName)` pairs for client-side preflight + 409 fallback. */
    targets: OverwriteConflict[];
}

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

const splitName = (path: string): string => path.split('/').pop() ?? path;

/**
 * Pre-compute the per-file destinations the backend will write to. This mirrors
 * the server's resolution rules so the client can preflight against the local
 * snapshot and so the conflict dialog can name the exact items at risk:
 *
 *   - 1 source, no trailing `/` → destination is the exact target path
 *   - 1 source, trailing `/`    → backend treats destination as a dir;
 *                                 target = `<dir>/<file.name>`
 *   - 2+ sources                → destination is always a base directory;
 *                                 each source lands at `<dir>/<file.name>`.
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

const buildMovePlan = (files: readonly [FileNode, ...FileNode[]], values: ResourcesMoveFormValues): MovePlan => ({
    destination: values.destination.trim(),
    sources: files.map((file) => file.path),
    targets: computeTargets(files, values.destination),
});

export function ResourcesMoveDialog({ files, onClose }: ResourcesMoveDialogProps) {
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
}

function ResourcesMoveDialogForm({ files, onClose }: ResourcesMoveDialogFormProps) {
    const { isMoving, move } = useResourcesMove();
    const { resources } = useResources();
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

    const form = useAppForm<ResourcesMoveFormValues>({
        defaultValues: { destination: defaultDestination },
        schema: resourcesMoveFormSchema,
    });

    useEffect(() => {
        form.reset({ destination: defaultDestination });
    }, [defaultDestination, form]);

    const resourcePaths = useMemo(() => new Set(resources.map((resource) => resource.path)), [resources]);
    const sourcePaths = useMemo(() => new Set(files.map((file) => file.path)), [files]);

    /**
     * Drive the canonical "Move / Move with overwrite / Replace all" workflow
     * with a single atomic batch request. Backend handles `sources[]` in one
     * DB transaction (all-or-nothing) — no per-source aggregation needed here.
     */
    const overwriteAction = useOverwrite<MovePlan>({
        execute: (plan, force) => move(plan.sources, plan.destination, force),
        // Local preflight: filter out targets that match an item we're moving
        // (those are no-ops, not conflicts) and keep the ones already taken
        // by some other resource.
        findConflicts: (plan) =>
            plan.targets.filter((t) => !sourcePaths.has(t.destination) && resourcePaths.has(t.destination)),
        onSuccess: onClose,
        // Race-fallback: backend doesn't return per-path conflict descriptors
        // on a 409, so we synthesize them from the plan we just submitted.
        synthesizeFallbackConflicts: (plan) => plan.targets,
    });

    const handleSave = form.handleSubmit(async (values) => {
        await overwriteAction.primaryExecute(buildMovePlan(files, values));
    });

    const handleSaveWithOverwrite = form.handleSubmit(async (values) => {
        await overwriteAction.forceExecute(buildMovePlan(files, values));
    });

    // Convention: stay enabled until the first submit, then reflect validity (so an invalid submit surfaces
    // errors instead of a silently-dead button). Mirrors FormSubmitButton's requireValid gate.
    const isSubmitDisabled = form.formState.isSubmitted && !form.formState.isValid;
    const titleText = isMulti
        ? `Move ${files.length} items`
        : files[0].isDir
          ? 'Move directory'
          : 'Rename or move resource';
    const overwriteCtaLabel = isMulti ? `Move ${files.length} with overwrite` : 'Move with overwrite';

    return (
        <>
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
                        noValidate
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
                                            placeholder={
                                                isMulti ? 'Leave empty to move into the library root' : undefined
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
                                            <>
                                                Relative path inside your library. End with <code>/</code> to drop the
                                                entry into that directory.
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
                            <OverwriteButtons
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
            </DialogContent>

            <OverwriteDialog
                conflicts={overwriteAction.conflicts}
                onCancel={overwriteAction.resetConflicts}
                onReplaceAll={overwriteAction.handleReplaceAll}
            />
        </>
    );
}
