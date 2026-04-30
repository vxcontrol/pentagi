import { zodResolver } from '@hookform/resolvers/zod';
import { FolderInput, Loader2 } from 'lucide-react';
import { useEffect } from 'react';
import { useForm } from 'react-hook-form';

import type { FileNode } from '@/components/file-manager';

import { Button } from '@/components/ui/button';
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from '@/components/ui/dialog';
import { Form, FormControl, FormDescription, FormField, FormItem, FormLabel, FormMessage } from '@/components/ui/form';
import { Input } from '@/components/ui/input';
import { Switch } from '@/components/ui/switch';

import { ResourcesConflictDialog } from './resources-conflict-dialog';
import { resourcesMoveFormSchema, type ResourcesMoveFormValues, useResourcesMove } from './use-resources-move';

interface ResourcesMoveDialogFormProps {
    file: FileNode;
    onClose: () => void;
}

interface ResourcesMoveDialogProps {
    file: FileNode | null;
    onClose: () => void;
}

const ResourcesMoveDialogForm = ({ file, onClose }: ResourcesMoveDialogFormProps) => {
    const { cancelConflicts, isMoving, move, pendingConflicts, resolveConflicts } = useResourcesMove();

    const form = useForm<ResourcesMoveFormValues>({
        defaultValues: {
            destination: file.path,
            shouldOverwrite: false,
        },
        mode: 'onChange',
        resolver: zodResolver(resourcesMoveFormSchema),
    });

    useEffect(() => {
        form.reset({
            destination: file.path,
            shouldOverwrite: false,
        });
    }, [file, form]);

    const handleSubmit = form.handleSubmit(async (values) => {
        const wasMoved = await move(file.path, values);

        if (wasMoved) {
            onClose();
        }
    });

    // After the user picks "Replace" in the conflict dialog the hook retries the move
    // with `force = true`. Close the form so it doesn't leave a stale modal — the
    // resolveConflicts promise resolves once every retry has settled.
    const handleResolveConflicts = async () => {
        await resolveConflicts();
        onClose();
    };

    const isSubmitDisabled = !form.formState.isValid || isMoving;

    return (
        <DialogContent>
            <DialogHeader>
                <DialogTitle className="flex items-center gap-2">
                    <FolderInput className="size-4" />
                    {file.isDir ? 'Move directory' : 'Rename or move resource'}
                </DialogTitle>
                <DialogDescription>
                    Update the path of <code>{file.path}</code>.
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
                                <FormLabel>New path</FormLabel>
                                <FormControl>
                                    <Input
                                        {...field}
                                        autoComplete="off"
                                        autoFocus
                                        disabled={isMoving}
                                    />
                                </FormControl>
                                <FormDescription>
                                    Relative path inside your library. End with <code>/</code> to drop the entry into
                                    that directory.
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
                                    Overwrite if a resource already exists at the destination
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

export const ResourcesMoveDialog = ({ file, onClose }: ResourcesMoveDialogProps) => {
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
                <ResourcesMoveDialogForm
                    file={file}
                    onClose={onClose}
                />
            )}
        </Dialog>
    );
};
