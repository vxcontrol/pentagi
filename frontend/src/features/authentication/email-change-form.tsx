import { useState } from 'react';
import { toast } from 'sonner';
import * as z from 'zod';

import { Button } from '@/components/ui/button';
import { Form, FormControl, FormField, FormItem, FormLabel, FormMessage } from '@/components/ui/form';
import { FormSubmitButton } from '@/components/ui/form-submit-button';
import { Input } from '@/components/ui/input';
import { InputPassword } from '@/components/ui/input-password';
import { useAppForm } from '@/hooks/use-app-form';
import { api, resolveApiErrorMessage } from '@/lib/axios';
import { useUser } from '@/providers/user-provider';

const emailChangeSchema = z.object({
    currentPassword: z.string().min(1, { message: 'Current password is required' }),
    newEmail: z
        .string()
        .trim()
        .toLowerCase()
        .min(1, { message: 'Email is required' })
        .email({ message: 'Invalid email address' })
        .max(50, { message: 'Email must not exceed 50 characters' }),
});

const ERROR_BY_CODE: Record<string, string> = {
    'Users.ChangeEmailCurrentUser.EmailAlreadyExists': 'Email address is already in use',
    'Users.ChangeEmailCurrentUser.InvalidCurrentPassword': 'Current password is incorrect',
    'Users.ChangeEmailCurrentUser.InvalidEmail': 'New email does not meet requirements',
    'Users.NotFound': 'User not found',
};

interface EmailChangeFormProps {
    onCancel?: () => void;
    onSuccess?: () => void;
}

type EmailChangeFormValues = z.infer<typeof emailChangeSchema>;

export function EmailChangeForm({ onCancel, onSuccess }: EmailChangeFormProps) {
    const [error, setError] = useState<null | string>(null);
    const { patchUser, refreshAuthInfo } = useUser();

    const form = useAppForm<EmailChangeFormValues>({
        defaultValues: {
            currentPassword: '',
            newEmail: '',
        },
        schema: emailChangeSchema,
    });

    const handleSubmit = async (values: EmailChangeFormValues) => {
        setError(null);

        try {
            await api.put('/user/email', {
                current_password: values.currentPassword,
                mail: values.newEmail,
            });

            form.reset();
            toast.success('Email successfully updated');

            patchUser({ mail: values.newEmail });
            await refreshAuthInfo();

            onSuccess?.();
        } catch (err: unknown) {
            setError(resolveApiErrorMessage(err, ERROR_BY_CODE, 'Failed to update email'));
        }
    };

    return (
        <Form {...form}>
            {/* noValidate: the type="email" field would otherwise fire the browser's native (locale-styled)
                validation popup on submit, pre-empting our zod message. Validation runs through zod instead. */}
            <form
                className="flex flex-col gap-4"
                noValidate
                onSubmit={form.handleSubmit(handleSubmit)}
            >
                <FormField
                    control={form.control}
                    name="currentPassword"
                    render={({ field }) => (
                        <FormItem>
                            <FormLabel>Current Password</FormLabel>
                            <FormControl>
                                <InputPassword
                                    {...field}
                                    placeholder="Enter your current password"
                                />
                            </FormControl>
                            <FormMessage />
                        </FormItem>
                    )}
                />

                <FormField
                    control={form.control}
                    name="newEmail"
                    render={({ field }) => (
                        <FormItem>
                            <FormLabel>New Email</FormLabel>
                            <FormControl>
                                <Input
                                    {...field}
                                    placeholder="Enter your new email address"
                                    type="email"
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
                        <span>Update Email</span>
                    </FormSubmitButton>
                </div>
            </form>
        </Form>
    );
}
