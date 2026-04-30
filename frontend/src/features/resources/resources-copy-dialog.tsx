import { zodResolver } from '@hookform/resolvers/zod';
import { Copy, Loader2 } from 'lucide-react';
import { useEffect } from 'react';
import { useForm } from 'react-hook-form';

import type { FileNode } from '@/components/file-manager';

import { Button } from '@/components/ui/button';
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from '@/components/ui/dialog';
import { Form, FormControl, FormDescription, FormField, FormItem, FormLabel, FormMessage } from '@/components/ui/form';
import { Input } from '@/components/ui/input';
import { Switch } from '@/components/ui/switch';

import { ResourcesConflictDialog } from './resources-conflict-dialog';
import { resourcesCopyFormSchema, type ResourcesCopyFormValues, useResourcesCopy } from './use-resources-copy';

interface ResourcesCopyDialogFormProps {
    file: FileNode;
    onClose: () => void;
}

interface ResourcesCopyDialogProps {
    file: FileNode | null;
    onClose: () => void;
}

const buildDefaultDestination = (file: FileNode): string => {
    const segments = file.path.split('/');
    const lastSegment = segments.at(-1) ?? file.name;
    const parent = segments.slice(0, -1).join('/');
    const dotIndex = lastSegment.lastIndexOf('.');
    const baseName = file.isDir || dotIndex === -1 ? lastSegment : lastSegment.slice(0, dotIndex);
    const extension = file.isDir || dotIndex === -1 ? '' : lastSegment.slice(dotIndex);
    const candidateName = `${baseName}-copy${extension}`;

    return parent ? `${parent}/${candidateName}` : candidateName;
};

const ResourcesCopyDialogForm = ({ file, onClose }: ResourcesCopyDialogFormProps) => {
    const { cancelConflicts, copy, isCopying, pendingConflicts, resolveConflicts } = useResourcesCopy();

    const form = useForm<ResourcesCopyFormValues>({
        defaultValues: {
            destination: buildDefaultDestination(file),
            shouldOverwrite: false,
        },
        mode: 'onChange',
        resolver: zodResolver(resourcesCopyFormSchema),
    });

    useEffect(() => {
        form.reset({
            destination: buildDefaultDestination(file),
            shouldOverwrite: false,
        });
    }, [file, form]);

    const handleSubmit = form.handleSubmit(async (values) => {
        const wasCopied = await copy(file.path, values);

        if (wasCopied) {
            onClose();
        }
    });

    // After the user picks "Replace" in the conflict dialog the hook retries the copy
    // with `force = true`. Close the form so it doesn't leave a stale modal — the
    // resolveConflicts promise resolves once every retry has settled.
    const handleResolveConflicts = async () => {
        await resolveConflicts();
        onClose();
    };

    const isSubmitDisabled = !form.formState.isValid || isCopying;

    return (
        <DialogContent>
            <DialogHeader>
                <DialogTitle className="flex items-center gap-2">
                    <Copy className="size-4" />
                    {file.isDir ? 'Copy directory' : 'Copy resource'}
                </DialogTitle>
                <DialogDescription>
                    Duplicate <code>{file.path}</code> to a new path.
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
                                        disabled={isCopying}
                                    />
                                </FormControl>
                                <FormDescription>Relative path inside your library.</FormDescription>
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
                                    Overwrite or merge if the destination already exists
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

export const ResourcesCopyDialog = ({ file, onClose }: ResourcesCopyDialogProps) => {
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
                <ResourcesCopyDialogForm
                    file={file}
                    onClose={onClose}
                />
            )}
        </Dialog>
    );
};
