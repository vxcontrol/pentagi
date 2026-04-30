import { zodResolver } from '@hookform/resolvers/zod';
import { ArrowDownToLine, Loader2 } from 'lucide-react';
import { useForm } from 'react-hook-form';

import { Button } from '@/components/ui/button';
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from '@/components/ui/dialog';
import { Form, FormControl, FormDescription, FormField, FormItem, FormLabel, FormMessage } from '@/components/ui/form';
import { Input } from '@/components/ui/input';
import { Switch } from '@/components/ui/switch';

import { flowFilesPullFormSchema, type FlowFilesPullFormValues, useFlowFilesPull } from './use-flow-files-pull';

interface FlowFilesPullDialogFormProps {
    flowId: null | string;
    onClose: () => void;
    onSuccess: () => void;
}

interface FlowFilesPullDialogProps {
    flowId: null | string;
    isOpen: boolean;
    onClose: () => void;
    onSuccess: () => void;
}

const DEFAULT_PULL_FORM_VALUES: FlowFilesPullFormValues = {
    containerPath: '',
    shouldOverwrite: false,
};

/**
 * Inner component holding the form state. Mounted only while the dialog is open
 * so closing the dialog discards every transient field without an imperative reset.
 */
const FlowFilesPullDialogForm = ({ flowId, onClose, onSuccess }: FlowFilesPullDialogFormProps) => {
    const { isPulling, pull } = useFlowFilesPull({ flowId, onSuccess });

    const form = useForm<FlowFilesPullFormValues>({
        defaultValues: DEFAULT_PULL_FORM_VALUES,
        mode: 'onChange',
        resolver: zodResolver(flowFilesPullFormSchema),
    });

    const handleSubmit = form.handleSubmit(async (values) => {
        const wasPulled = await pull(values);

        if (wasPulled) {
            onClose();
        }
    });

    const isSubmitDisabled = !form.formState.isValid || isPulling;

    return (
        <DialogContent>
            <DialogHeader>
                <DialogTitle className="flex items-center gap-2">
                    <ArrowDownToLine className="size-4" />
                    Pull from container
                </DialogTitle>
                <DialogDescription>
                    Enter a path inside the running container. The file or directory will be synced to the local cache
                    under <code>container/</code>.
                </DialogDescription>
            </DialogHeader>

            <Form {...form}>
                <form
                    className="flex flex-col gap-4"
                    onSubmit={handleSubmit}
                >
                    <FormField
                        control={form.control}
                        name="containerPath"
                        render={({ field }) => (
                            <FormItem>
                                <FormLabel>Container path</FormLabel>
                                <FormControl>
                                    <Input
                                        {...field}
                                        autoComplete="off"
                                        autoFocus
                                        disabled={isPulling}
                                        placeholder="/etc/nginx/conf"
                                    />
                                </FormControl>
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
                                        disabled={isPulling}
                                        onCheckedChange={field.onChange}
                                    />
                                </FormControl>
                                <FormLabel className="cursor-pointer font-normal">
                                    Overwrite if already cached
                                </FormLabel>
                                <FormDescription className="sr-only">
                                    Replace the cached entry when it already exists.
                                </FormDescription>
                            </FormItem>
                        )}
                    />

                    <div className="flex justify-end gap-2">
                        <Button
                            disabled={isPulling}
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
                            {isPulling ? <Loader2 className="animate-spin" /> : <ArrowDownToLine />}
                            Pull
                        </Button>
                    </div>
                </form>
            </Form>
        </DialogContent>
    );
};

export const FlowFilesPullDialog = ({ flowId, isOpen, onClose, onSuccess }: FlowFilesPullDialogProps) => {
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
                <FlowFilesPullDialogForm
                    flowId={flowId}
                    onClose={onClose}
                    onSuccess={onSuccess}
                />
            )}
        </Dialog>
    );
};
