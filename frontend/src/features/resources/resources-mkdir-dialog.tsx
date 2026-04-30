import { zodResolver } from '@hookform/resolvers/zod';
import { FolderPlus, Loader2 } from 'lucide-react';
import { useEffect } from 'react';
import { useForm } from 'react-hook-form';

import { Button } from '@/components/ui/button';
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from '@/components/ui/dialog';
import { Form, FormControl, FormDescription, FormField, FormItem, FormLabel, FormMessage } from '@/components/ui/form';
import { Input } from '@/components/ui/input';

import { resourcesMkdirFormSchema, type ResourcesMkdirFormValues, useResourcesMkdir } from './use-resources-mkdir';

interface ResourcesMkdirDialogFormProps {
    defaultParentPath: string;
    onClose: () => void;
}

interface ResourcesMkdirDialogProps {
    /** Pre-filled parent path (without leading "/"). Empty string targets the root. */
    defaultParentPath?: string;
    isOpen: boolean;
    onClose: () => void;
}

const buildDefaultPath = (defaultParentPath: string): string =>
    defaultParentPath ? `${defaultParentPath.replace(/\/+$/u, '')}/new-folder` : 'new-folder';

const ResourcesMkdirDialogForm = ({ defaultParentPath, onClose }: ResourcesMkdirDialogFormProps) => {
    const { isCreating, mkdir } = useResourcesMkdir();

    const form = useForm<ResourcesMkdirFormValues>({
        defaultValues: { path: buildDefaultPath(defaultParentPath) },
        mode: 'onChange',
        resolver: zodResolver(resourcesMkdirFormSchema),
    });

    useEffect(() => {
        form.reset({ path: buildDefaultPath(defaultParentPath) });
    }, [defaultParentPath, form]);

    const handleSubmit = form.handleSubmit(async (values) => {
        const wasCreated = await mkdir(values);

        if (wasCreated) {
            onClose();
        }
    });

    const isSubmitDisabled = !form.formState.isValid || isCreating;

    return (
        <DialogContent>
            <DialogHeader>
                <DialogTitle className="flex items-center gap-2">
                    <FolderPlus className="size-4" />
                    Create directory
                </DialogTitle>
                <DialogDescription>
                    Create a virtual directory inside your resource library. Existing parent directories are reused
                    automatically.
                </DialogDescription>
            </DialogHeader>

            <Form {...form}>
                <form
                    className="flex flex-col gap-4"
                    onSubmit={handleSubmit}
                >
                    <FormField
                        control={form.control}
                        name="path"
                        render={({ field }) => (
                            <FormItem>
                                <FormLabel>Path</FormLabel>
                                <FormControl>
                                    <Input
                                        {...field}
                                        autoComplete="off"
                                        autoFocus
                                        disabled={isCreating}
                                        placeholder="reports/2025"
                                    />
                                </FormControl>
                                <FormDescription>
                                    Relative path. Use <code>/</code> to nest into subdirectories.
                                </FormDescription>
                                <FormMessage />
                            </FormItem>
                        )}
                    />

                    <div className="flex justify-end gap-2">
                        <Button
                            disabled={isCreating}
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
                            {isCreating ? <Loader2 className="animate-spin" /> : <FolderPlus />}
                            Create
                        </Button>
                    </div>
                </form>
            </Form>
        </DialogContent>
    );
};

export const ResourcesMkdirDialog = ({ defaultParentPath = '', isOpen, onClose }: ResourcesMkdirDialogProps) => {
    const handleDialogOpenChange = (nextOpen: boolean) => {
        if (!nextOpen) {
            onClose();
        }
    };

    return (
        <Dialog
            onOpenChange={handleDialogOpenChange}
            open={isOpen}
        >
            {isOpen && (
                <ResourcesMkdirDialogForm
                    defaultParentPath={defaultParentPath}
                    onClose={onClose}
                />
            )}
        </Dialog>
    );
};
