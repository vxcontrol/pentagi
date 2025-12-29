import { zodResolver } from '@hookform/resolvers/zod';
import { Loader2 } from 'lucide-react';
import { useState } from 'react';
import { useForm } from 'react-hook-form';
import * as z from 'zod';

import { Button } from '@/components/ui/button';
import { Form, FormControl, FormField, FormItem, FormLabel, FormMessage } from '@/components/ui/form';
import { Input } from '@/components/ui/input';
import { axios } from '@/lib/axios';

const passwordChangeSchema = z
    .object({
        confirmPassword: z.string().min(1, { message: 'Confirm your password' }),
        currentPassword: z.string().min(1, { message: 'Current password is required' }),
        newPassword: z.string().min(8, { message: 'Password must be at least 8 characters' }),
    })
    .refine((data) => data.newPassword === data.confirmPassword, {
        message: "Passwords don't match",
        path: ['confirmPassword'],
    });

interface PasswordChangeFormProps {
    isModal?: boolean;
    onCancel?: () => void;
    onSkip?: () => void;
    onSuccess?: () => void;
    showSkip?: boolean;
}

type PasswordChangeFormValues = z.infer<typeof passwordChangeSchema>;

export function PasswordChangeForm({
    isModal = true,
    onCancel,
    onSkip,
    onSuccess,
    showSkip = false,
}: PasswordChangeFormProps) {
    const [isSubmitting, setIsSubmitting] = useState(false);
    const [error, setError] = useState<null | string>(null);

    const form = useForm<PasswordChangeFormValues>({
        defaultValues: {
            confirmPassword: '',
            currentPassword: '',
            newPassword: '',
        },
        resolver: zodResolver(passwordChangeSchema),
    });

    const handleSubmit = async (values: PasswordChangeFormValues) => {
        setIsSubmitting(true);
        setError(null);

        try {
            await axios.put('/user/password', {
                confirm_password: values.confirmPassword,
                current_password: values.currentPassword,
                password: values.newPassword,
            });

            if (onSuccess) {
                onSuccess();
            }
        } catch (err: any) {
            setError(err.message || 'Failed to change password');
        } finally {
            setIsSubmitting(false);
        }
    };

    return (
        <Form {...form}>
            <form
                className="flex flex-col gap-4"
                onSubmit={form.handleSubmit(handleSubmit)}
            >
                <FormField
                    control={form.control}
                    name="currentPassword"
                    render={({ field }) => (
                        <FormItem>
                            <FormLabel>Current Password</FormLabel>
                            <FormControl>
                                <Input
                                    {...field}
                                    placeholder="Enter your current password"
                                    type="password"
                                />
                            </FormControl>
                            <FormMessage />
                        </FormItem>
                    )}
                />

                <FormField
                    control={form.control}
                    name="newPassword"
                    render={({ field }) => (
                        <FormItem>
                            <FormLabel>New Password</FormLabel>
                            <FormControl>
                                <Input
                                    {...field}
                                    placeholder="Enter new password"
                                    type="password"
                                />
                            </FormControl>
                            <FormMessage />
                        </FormItem>
                    )}
                />

                <FormField
                    control={form.control}
                    name="confirmPassword"
                    render={({ field }) => (
                        <FormItem>
                            <FormLabel>Confirm New Password</FormLabel>
                            <FormControl>
                                <Input
                                    {...field}
                                    placeholder="Confirm new password"
                                    type="password"
                                />
                            </FormControl>
                            <FormMessage />
                        </FormItem>
                    )}
                />

                {error && <div className="text-destructive text-sm">{error}</div>}

                <div className="flex justify-end gap-2 pt-2">
                    {showSkip && (
                        <Button
                            className="text-muted-foreground"
                            onClick={onSkip}
                            type="button"
                            variant="ghost"
                        >
                            Skip for now
                        </Button>
                    )}
                    {isModal && (
                        <Button
                            onClick={onCancel}
                            type="button"
                            variant="outline"
                        >
                            Cancel
                        </Button>
                    )}
                    <Button
                        disabled={isSubmitting || (!form.formState.isValid && form.formState.isSubmitted)}
                        type="submit"
                    >
                        {isSubmitting && <Loader2 className="mr-2 size-4 animate-spin" />}
                        <span>Update Password</span>
                    </Button>
                </div>
            </form>
        </Form>
    );
}
