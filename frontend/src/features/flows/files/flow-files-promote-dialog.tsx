import { BookmarkPlus } from 'lucide-react';
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

import { stripFlowRootPrefix } from './flow-files-utils';
import {
    flowFilesPromoteFormSchema,
    type FlowFilesPromoteFormValues,
    useFlowFilesPromote,
} from './use-flow-files-promote';

/** Guaranteed non-empty by `FlowFilesPromoteDialog` (which gates rendering on `files.length > 0`). */
interface FlowFilesPromoteDialogFormProps {
    files: readonly [FileNode, ...FileNode[]];
    flowId: null | string;
    onClose: () => void;
}

interface FlowFilesPromoteDialogProps {
    /**
     * One or more files to promote into the user's resource library. A single-element
     * array renders the "save N to a specific path" UI; multi-element arrays render
     * the "save N items into directory" UI (each file keeps its current name).
     *
     * Use `null` or an empty array to close the dialog.
     */
    files: FileNode[] | null;
    flowId: null | string;
    onClose: () => void;
}

interface PromotePlan {
    /** Destination string sent to the backend (exact path or base directory). */
    destination: string;
    /** Source paths inside the flow cache (sent as `sources[]`). */
    sources: readonly string[];
    /** Pre-computed `(destination, destinationName)` pairs for client-side preflight + 409 fallback. */
    targets: OverwriteConflict[];
}

const buildSingleDefaultDestination = (file: FileNode): string => stripFlowRootPrefix(file.path) || file.name;

/**
 * Default destination directory for a multi-file promote: strip the synthetic
 * root-group prefix (`uploads/`, `container/`, `resources/`) from the file's
 * parent path. When selections span different parents, fall back to the
 * library root so we don't pick one arbitrarily.
 */
const computeMultiDefaultDestination = (files: readonly [FileNode, ...FileNode[]]): string => {
    const stripParent = (path: string): string => {
        const stripped = stripFlowRootPrefix(path);
        const idx = stripped.lastIndexOf('/');

        return idx === -1 ? '' : stripped.slice(0, idx);
    };

    const first = stripParent(files[0].path);

    return files.every((file) => stripParent(file.path) === first) ? first : '';
};

/**
 * Pre-compute the per-file destinations the backend will write to. Mirrors the
 * server's resolution rules:
 *   - 1 source                  → destination is the exact target path
 *   - 2+ sources                → destination is a base directory; each source
 *                                 lands at `<dir>/<file.name>`.
 */
const computeTargets = (files: readonly [FileNode, ...FileNode[]], destination: string): OverwriteConflict[] => {
    const trimmed = destination.trim();

    if (files.length > 1) {
        const baseDir = trimmed.replace(/\/+$/, '');

        return files.map((file) => {
            const dest = baseDir ? `${baseDir}/${file.name}` : file.name;

            return { destination: dest, destinationName: file.name };
        });
    }

    return [
        {
            destination: trimmed,
            destinationName: trimmed.split('/').pop() ?? trimmed,
        },
    ];
};

const buildPromotePlan = (
    files: readonly [FileNode, ...FileNode[]],
    values: FlowFilesPromoteFormValues,
): PromotePlan => ({
    destination: values.destination.trim(),
    sources: files.map((file) => file.path),
    targets: computeTargets(files, values.destination),
});

export function FlowFilesPromoteDialog({ files, flowId, onClose }: FlowFilesPromoteDialogProps) {
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
                <FlowFilesPromoteDialogForm
                    files={nonEmptyFiles}
                    flowId={flowId}
                    onClose={onClose}
                />
            )}
        </Dialog>
    );
}

function FlowFilesPromoteDialogForm({ files, flowId, onClose }: FlowFilesPromoteDialogFormProps) {
    const { isPromoting, promote } = useFlowFilesPromote({ flowId });
    const { resources } = useResources();
    const isMulti = files.length > 1;

    const defaultDestination = useMemo(() => {
        if (isMulti) {
            return computeMultiDefaultDestination(files);
        }

        return buildSingleDefaultDestination(files[0]);
    }, [files, isMulti]);

    const form = useAppForm<FlowFilesPromoteFormValues>({
        defaultValues: { destination: defaultDestination },
        schema: flowFilesPromoteFormSchema,
    });

    useEffect(() => {
        form.reset({ destination: defaultDestination });
    }, [defaultDestination, form]);

    const resourcePaths = useMemo(() => new Set(resources.map((resource) => resource.path)), [resources]);

    /**
     * Drive the canonical "Save / Save with overwrite / Replace all" workflow
     * with a single atomic batch request. Backend handles `sources[]` in one
     * DB transaction (all-or-nothing) — no per-source aggregation needed here.
     */
    const overwriteAction = useOverwrite<PromotePlan>({
        execute: (plan, force) => promote(plan.sources, plan.destination, force),
        findConflicts: (plan) => plan.targets.filter((t) => resourcePaths.has(t.destination)),
        onSuccess: onClose,
        // Race-fallback: backend doesn't return per-path conflict descriptors
        // on a 409, so we synthesize them from the plan we just submitted.
        synthesizeFallbackConflicts: (plan) => plan.targets,
    });

    const handleSave = form.handleSubmit(async (values) => {
        await overwriteAction.primaryExecute(buildPromotePlan(files, values));
    });

    const handleSaveWithOverwrite = form.handleSubmit(async (values) => {
        await overwriteAction.forceExecute(buildPromotePlan(files, values));
    });

    // Convention: stay enabled until the first submit, then reflect validity (so an invalid submit surfaces
    // errors instead of a silently-dead button). Mirrors FormSubmitButton's requireValid gate.
    const isSubmitDisabled = form.formState.isSubmitted && !form.formState.isValid;
    const titleText = isMulti ? `Save ${files.length} items as resources` : 'Save as resource';
    const overwriteCtaLabel = isMulti ? `Save ${files.length} with overwrite` : 'Save with overwrite';

    return (
        <>
            <DialogContent>
                <DialogHeader>
                    <DialogTitle className="flex items-center gap-2">
                        <BookmarkPlus className="size-4" />
                        {titleText}
                    </DialogTitle>
                    <DialogDescription>
                        {isMulti ? (
                            <>
                                Promote every selected entry from this flow into your global resource library so you can
                                reuse them in other flows.
                            </>
                        ) : (
                            <>
                                Promote <code>{files[0].path}</code> from this flow into your global resource library so
                                you can reuse it in other flows.
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
                                            disabled={isPromoting}
                                            placeholder={
                                                isMulti
                                                    ? 'Leave empty to save into the library root'
                                                    : 'results/scan.txt'
                                            }
                                        />
                                    </FormControl>
                                    <FormDescription>
                                        {isMulti ? (
                                            <>
                                                Relative directory inside your resource library. Leave empty for the
                                                root. Each item keeps its current filename.
                                            </>
                                        ) : (
                                            <>
                                                Relative path inside your resource library. Use <code>/</code> to nest
                                                into subdirectories.
                                            </>
                                        )}
                                    </FormDescription>
                                    <FormMessage />
                                </FormItem>
                            )}
                        />

                        <div className="flex flex-wrap justify-end gap-2">
                            <Button
                                disabled={isPromoting}
                                onClick={onClose}
                                type="button"
                                variant="outline"
                            >
                                Cancel
                            </Button>
                            <OverwriteButtons
                                isDisabled={isSubmitDisabled}
                                isProcessing={isPromoting}
                                onOverwrite={() => {
                                    void handleSaveWithOverwrite();
                                }}
                                overwriteLabel={overwriteCtaLabel}
                                primaryIcon={BookmarkPlus}
                                primaryLabel="Save"
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
