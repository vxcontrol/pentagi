import { zodResolver } from '@hookform/resolvers/zod';
import { BookmarkPlus, Loader2 } from 'lucide-react';
import { useEffect, useMemo } from 'react';
import { useForm } from 'react-hook-form';

import type { FileNode } from '@/components/file-manager';

import { Button } from '@/components/ui/button';
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from '@/components/ui/dialog';
import { Form, FormControl, FormDescription, FormField, FormItem, FormLabel, FormMessage } from '@/components/ui/form';
import { Input } from '@/components/ui/input';
import { Switch } from '@/components/ui/switch';

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

const buildSingleDefaultDestination = (file: FileNode) => stripFlowRootPrefix(file.path) || file.name;

/**
 * Default destination directory for a multi-file promote: strip the synthetic
 * root-group prefix (`uploads/`, `container/`, `resources/`) from the file's
 * parent path so the user sees a "library-relative" suggestion. When selections
 * span different parents, fall back to the library root so we don't pick one
 * arbitrarily.
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

const FlowFilesPromoteDialogForm = ({ files, flowId, onClose }: FlowFilesPromoteDialogFormProps) => {
    const { isPromoting, promote } = useFlowFilesPromote({ flowId });
    const isMulti = files.length > 1;

    const defaultDestination = useMemo(() => {
        if (isMulti) {
            return computeMultiDefaultDestination(files);
        }

        return buildSingleDefaultDestination(files[0]);
    }, [files, isMulti]);

    const form = useForm<FlowFilesPromoteFormValues>({
        defaultValues: {
            destination: defaultDestination,
            shouldOverwrite: false,
        },
        mode: 'onChange',
        resolver: zodResolver(flowFilesPromoteFormSchema),
    });

    useEffect(() => {
        form.reset({
            destination: defaultDestination,
            shouldOverwrite: false,
        });
    }, [defaultDestination, form]);

    const handleSubmit = form.handleSubmit(async (values) => {
        if (isMulti) {
            // Multi-file promote: `destination` is the *target directory* inside
            // the resource library, every file keeps its current name. Strip a
            // trailing `/` for consistency with the single-file branch.
            const targetDir = values.destination.trim().replace(/\/+$/, '');
            const results = await Promise.all(
                files.map((file) => {
                    const destination = targetDir ? `${targetDir}/${file.name}` : file.name;

                    return promote(file.path, { destination, shouldOverwrite: values.shouldOverwrite });
                }),
            );

            if (results.every(Boolean)) {
                onClose();
            }

            return;
        }

        const wasPromoted = await promote(files[0].path, values);

        if (wasPromoted) {
            onClose();
        }
    });

    const isSubmitDisabled = !form.formState.isValid || isPromoting;
    const titleText = isMulti ? `Save ${files.length} items as resources` : 'Save as resource';

    return (
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
                            Promote <code>{files[0].path}</code> from this flow into your global resource library so you
                            can reuse it in other flows.
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
                                        disabled={isPromoting}
                                        placeholder={
                                            isMulti ? 'Leave empty to save into the library root' : 'results/scan.txt'
                                        }
                                    />
                                </FormControl>
                                <FormDescription>
                                    {isMulti ? (
                                        <>
                                            Relative directory inside your resource library. Leave empty for the root.
                                            Each item keeps its current filename.
                                        </>
                                    ) : (
                                        <>
                                            Relative path inside your resource library. Use <code>/</code> to nest into
                                            subdirectories.
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
                                        disabled={isPromoting}
                                        onCheckedChange={field.onChange}
                                    />
                                </FormControl>
                                <FormLabel className="cursor-pointer font-normal">
                                    {isMulti
                                        ? 'Overwrite if a resource already exists at any destination'
                                        : 'Overwrite if a resource already exists at this path'}
                                </FormLabel>
                                <FormDescription className="sr-only">
                                    Replace the existing resource entry when one already exists.
                                </FormDescription>
                            </FormItem>
                        )}
                    />

                    <div className="flex justify-end gap-2">
                        <Button
                            disabled={isPromoting}
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
                            {isPromoting ? <Loader2 className="animate-spin" /> : <BookmarkPlus />}
                            Save
                        </Button>
                    </div>
                </form>
            </Form>
        </DialogContent>
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
