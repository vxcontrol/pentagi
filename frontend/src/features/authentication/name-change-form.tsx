import { useState } from 'react';
import { toast } from 'sonner';
import * as z from 'zod';

import { Button } from '@/components/ui/button';
import { Form, FormControl, FormField, FormItem, FormLabel, FormMessage } from '@/components/ui/form';
import { FormSubmitButton } from '@/components/ui/form-submit-button';
import { Input } from '@/components/ui/input';
import { useAppForm } from '@/hooks/use-app-form';
import { api, resolveApiErrorMessage } from '@/lib/axios';
import { useUser } from '@/providers/user-provider';

const nameChangeSchema = z.object({
    name: z
        .string()
        .trim()
        .min(1, { message: 'Name is required' })
        .max(70, { message: 'Name must not exceed 70 characters' }),
});

const ERROR_BY_CODE: Record<string, string> = {
    'Users.ChangeNameCurrentUser.InvalidName': 'New name does not meet requirements',
    'Users.NotFound': 'User not found',
};

interface NameChangeFormProps {
    onCancel?: () => void;
    onSuccess?: () => void;
}

type NameChangeFormValues = z.infer<typeof nameChangeSchema>;

export function NameChangeForm({ onCancel, onSuccess }: NameChangeFormProps) {
    const [error, setError] = useState<null | string>(null);
    const { authInfo, patchUser, refreshAuthInfo } = useUser();

    const form = useAppForm<NameChangeFormValues>({
        defaultValues: {
            name: authInfo?.user?.name ?? '',
        },
        schema: nameChangeSchema,
    });

    const handleSubmit = async (values: NameChangeFormValues) => {
        setError(null);

        try {
            await api.put('/user/name', { name: values.name });

            toast.success('Name successfully updated');

            patchUser({ name: values.name });
            await refreshAuthInfo();

            onSuccess?.();
        } catch (err: unknown) {
            setError(resolveApiErrorMessage(err, ERROR_BY_CODE, 'Failed to update name'));
        }
    };

    return (
        <Form {...form}>
            <form
                className="flex flex-col gap-4"
                noValidate
                onSubmit={form.handleSubmit(handleSubmit)}
            >
                <FormField
                    control={form.control}
                    name="name"
                    render={({ field }) => (
                        <FormItem>
                            <FormLabel>Display name</FormLabel>
                            <FormControl>
                                <Input
                                    {...field}
                                    placeholder="Enter your display name"
                                />
                            </FormControl>
                            <FormMessage />
                        </FormItem>
                    )}
                />

                {error && <div className="text-destructive text-sm">{error}</div>}

                <div className="flex justify-end gap-2 pt-2">
                    {onCancel && (
                        <Button
                            onClick={onCancel}
                            size="sm"
                            type="button"
                            variant="outline"
                        >
                            Cancel
                        </Button>
                    )}
                    <FormSubmitButton size="sm">
                        <span>Update Name</span>
                    </FormSubmitButton>
                </div>
            </form>
        </Form>
    );
}
