import { zodResolver } from '@hookform/resolvers/zod';
import { BookmarkPlus, Loader2 } from 'lucide-react';
import { useEffect } from 'react';
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

interface FlowFilesPromoteDialogFormProps {
    file: FileNode;
    flowId: null | string;
    onClose: () => void;
}

interface FlowFilesPromoteDialogProps {
    file: FileNode | null;
    flowId: null | string;
    onClose: () => void;
}

const buildDefaultDestination = (file: FileNode) => stripFlowRootPrefix(file.path) || file.name;

const FlowFilesPromoteDialogForm = ({ file, flowId, onClose }: FlowFilesPromoteDialogFormProps) => {
    const { isPromoting, promote } = useFlowFilesPromote({ flowId });

    const form = useForm<FlowFilesPromoteFormValues>({
        defaultValues: {
            destination: buildDefaultDestination(file),
            shouldOverwrite: false,
        },
        mode: 'onChange',
        resolver: zodResolver(flowFilesPromoteFormSchema),
    });

    useEffect(() => {
        form.reset({
            destination: buildDefaultDestination(file),
            shouldOverwrite: false,
        });
    }, [file, form]);

    const handleSubmit = form.handleSubmit(async (values) => {
        const wasPromoted = await promote(file.path, values);

        if (wasPromoted) {
            onClose();
        }
    });

    const isSubmitDisabled = !form.formState.isValid || isPromoting;

    return (
        <DialogContent>
            <DialogHeader>
                <DialogTitle className="flex items-center gap-2">
                    <BookmarkPlus className="size-4" />
                    Save as resource
                </DialogTitle>
                <DialogDescription>
                    Promote <code>{file.path}</code> from this flow into your global resource library so you can reuse
                    it in other flows.
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
                                <FormLabel>Destination path</FormLabel>
                                <FormControl>
                                    <Input
                                        {...field}
                                        autoComplete="off"
                                        autoFocus
                                        disabled={isPromoting}
                                        placeholder="results/scan.txt"
                                    />
                                </FormControl>
                                <FormDescription>
                                    Relative path inside your resource library. Use <code>/</code> to nest into
                                    subdirectories.
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
                                    Overwrite if a resource already exists at this path
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

export const FlowFilesPromoteDialog = ({ file, flowId, onClose }: FlowFilesPromoteDialogProps) => {
    const handleDialogOpenChange = (nextOpen: boolean) => {
        if (!nextOpen) {
            onClose();
        }
    };

    return (
        <Dialog
            onOpenChange={handleDialogOpenChange}
            open={!!file}
        >
            {file && (
                <FlowFilesPromoteDialogForm
                    file={file}
                    flowId={flowId}
                    onClose={onClose}
                />
            )}
        </Dialog>
    );
};
