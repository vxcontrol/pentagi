import { zodResolver } from '@hookform/resolvers/zod';
import { BookmarkPlus } from 'lucide-react';
import { useEffect, useMemo } from 'react';
import { useForm } from 'react-hook-form';

import type { FileNode } from '@/components/shared/file-manager';
import type { OverwriteConflict } from '@/components/shared/overwrite-confirm-dialog';
import type { OverwriteOutcome } from '@/components/shared/use-overwrite-action';

import { OverwriteConfirmDialog } from '@/components/shared/overwrite-confirm-dialog';
import { OverwriteCtaButtons } from '@/components/shared/overwrite-cta-buttons';
import { useOverwriteAction } from '@/components/shared/use-overwrite-action';
import { Button } from '@/components/ui/button';
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from '@/components/ui/dialog';
import { Form, FormControl, FormDescription, FormField, FormItem, FormLabel, FormMessage } from '@/components/ui/form';
import { Input } from '@/components/ui/input';
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
    /** Final virtual path inside the user's library. */
    destination: string;
    /** Display name extracted from the destination for the conflict dialog. */
    destinationName: string;
    /** Source path inside the flow cache (e.g. `uploads/result.md`). */
    source: string;
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
 * Build the final set of `(source, destination)` promotion plans from the form
 * value and the picked files. For a single file the user types the full path;
 * for a batch the input is a directory and every entry keeps its name.
 */
const buildPromotePlans = (
    files: readonly [FileNode, ...FileNode[]],
    values: FlowFilesPromoteFormValues,
): PromotePlan[] => {
    if (files.length > 1) {
        const targetDir = values.destination.trim().replace(/\/+$/, '');

        return files.map((file) => ({
            destination: targetDir ? `${targetDir}/${file.name}` : file.name,
            destinationName: file.name,
            source: file.path,
        }));
    }

    const [single] = files;
    const destination = values.destination.trim();

    return [
        {
            destination,
            destinationName: destination.split('/').pop() ?? destination,
            source: single.path,
        },
    ];
};

const planToConflict = ({ destination, destinationName }: PromotePlan): OverwriteConflict => ({
    destination,
    destinationName,
});

const FlowFilesPromoteDialogForm = ({ files, flowId, onClose }: FlowFilesPromoteDialogFormProps) => {
    const { isPromoting, promote } = useFlowFilesPromote({ flowId });
    const { resources } = useResources();
    const isMulti = files.length > 1;

    const defaultDestination = useMemo(() => {
        if (isMulti) {
            return computeMultiDefaultDestination(files);
        }

        return buildSingleDefaultDestination(files[0]);
    }, [files, isMulti]);

    const form = useForm<FlowFilesPromoteFormValues>({
        defaultValues: { destination: defaultDestination },
        mode: 'onChange',
        resolver: zodResolver(flowFilesPromoteFormSchema),
    });

    useEffect(() => {
        form.reset({ destination: defaultDestination });
    }, [defaultDestination, form]);

    const resourcePaths = useMemo(() => new Set(resources.map((resource) => resource.path)), [resources]);

    /**
     * Drive the canonical "Save / Save with overwrite / Replace all" workflow
     * from the shared hook. The plan is the array of promote operations
     * derived from the form on submit; per-file outcomes are aggregated into
     * a single `OverwriteOutcome` so the hook can branch uniformly.
     */
    const overwriteAction = useOverwriteAction<readonly PromotePlan[]>({
        execute: async (plans, force): Promise<OverwriteOutcome> => {
            const outcomes = await Promise.all(
                plans.map((plan) => promote(plan.source, { destination: plan.destination }, force)),
            );

            // Surface per-file conflict descriptors so the dialog names exactly
            // which destinations are taken — far more useful for multi-promote
            // than a count-based fallback.
            const conflicts: OverwriteConflict[] = [];

            outcomes.forEach((outcome, index) => {
                if (outcome.kind === 'conflict') {
                    const plan = plans[index];

                    if (plan) {
                        conflicts.push(planToConflict(plan));
                    }
                }
            });

            if (conflicts.length > 0) {
                return { conflicts, kind: 'conflict' };
            }

            if (outcomes.some((outcome) => outcome.kind === 'error')) {
                return { kind: 'error' };
            }

            return { kind: 'ok' };
        },
        findConflicts: (plans) =>
            plans.filter((plan) => resourcePaths.has(plan.destination)).map(planToConflict),
        onSuccess: onClose,
    });

    const handleSave = form.handleSubmit(async (values) => {
        await overwriteAction.primaryExecute(buildPromotePlans(files, values));
    });

    const handleSaveWithOverwrite = form.handleSubmit(async (values) => {
        await overwriteAction.forceExecute(buildPromotePlans(files, values));
    });

    const isSubmitDisabled = !form.formState.isValid;
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
                            <OverwriteCtaButtons
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

            <OverwriteConfirmDialog
                conflicts={overwriteAction.conflicts}
                onCancel={overwriteAction.resetConflicts}
                onReplaceAll={overwriteAction.handleReplaceAll}
            />
        </>
    );
};

export const FlowFilesPromoteDialog = ({ files, flowId, onClose }: FlowFilesPromoteDialogProps) => {
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
};
