import { zodResolver } from '@hookform/resolvers/zod';
import { Eye, EyeOff, Loader2 } from 'lucide-react';
import { useState } from 'react';
import { useForm } from 'react-hook-form';
import { toast } from 'sonner';
import * as z from 'zod';

import { Button } from '@/components/ui/button';
import { Form, FormControl, FormDescription, FormField, FormItem, FormLabel, FormMessage } from '@/components/ui/form';
import { Input } from '@/components/ui/input';
import { axios } from '@/lib/axios';

interface AxiosErrorResponse {
    message?: string;
    response?: {
        data?: ErrorResponse;
    };
}

interface ErrorResponse {
    code?: string;
    error?: string;
    msg?: string;
    status?: string;
}

const passwordChangeSchema = z
    .object({
        confirmPassword: z.string().min(1, { message: 'Confirm your password' }),
        currentPassword: z.string().min(1, { message: 'Current password is required' }),
        newPassword: z
            .string()
            .min(8, { message: 'Password must be at least 8 characters' })
            .max(100, { message: 'Password must not exceed 100 characters' })
            .refine(
                (password) => {
                    if (password.length > 15) {
                        return true;
                    }

                    return (
                        password.length >= 8 &&
                        /[0-9]/.test(password) &&
                        /[a-z]/.test(password) &&
                        /[A-Z]/.test(password) &&
                        /[!@#$&*]/.test(password)
                    );
                },
                {
                    message:
                        'Password must be either longer than 15 characters, or at least 8 characters with a number, lowercase, uppercase, and special character (!@#$&*)',
                },
            ),
    })
    .refine((data) => data.newPassword === data.confirmPassword, {
        message: "Passwords don't match",
        path: ['confirmPassword'],
    })
    .refine((data) => data.currentPassword !== data.newPassword, {
        message: 'New password must be different from current password',
        path: ['newPassword'],
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
    const [showCurrentPassword, setShowCurrentPassword] = useState(false);
    const [showNewPassword, setShowNewPassword] = useState(false);
    const [showConfirmPassword, setShowConfirmPassword] = useState(false);

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

            form.reset();
            setShowCurrentPassword(false);
            setShowNewPassword(false);
            setShowConfirmPassword(false);

            toast.success('Password successfully changed');

            if (onSuccess) {
                onSuccess();
            }
        } catch (err: unknown) {
            const error = err as AxiosErrorResponse;
            const responseData = error.response?.data;

            let errorMessage = 'Failed to change password';

            // Always prefer the msg from the response if available
            if (responseData?.msg) {
                errorMessage = responseData.msg;
            } else if (responseData?.code) {
                // Fallback to code-based messages
                switch (responseData.code) {
                    case 'AuthRequired':
                        errorMessage = 'Authentication required';
                        break;
                    case 'Users.ChangePasswordCurrentUser.InvalidCurrentPassword':
                        errorMessage = 'Current password is incorrect';
                        break;
                    case 'Users.ChangePasswordCurrentUser.InvalidNewPassword':
                        errorMessage = 'New password does not meet requirements';
                        break;
                    case 'Users.ChangePasswordCurrentUser.InvalidPassword':
                        errorMessage = 'Password validation failed';
                        break;
                    case 'Users.NotFound':
                        errorMessage = 'User not found';
                        break;
                    default:
                        errorMessage = responseData.msg || error.message || 'Failed to change password';
                }
            } else if (error.message) {
                errorMessage = error.message;
            }

            setError(errorMessage);
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
                                <div className="relative">
                                    <Input
                                        {...field}
                                        placeholder="Enter your current password"
                                        type={showCurrentPassword ? 'text' : 'password'}
                                    />
                                    <Button
                                        className="absolute top-0 right-0 h-full px-3 py-2 hover:bg-transparent"
                                        onClick={() => setShowCurrentPassword(!showCurrentPassword)}
                                        size="sm"
                                        tabIndex={-1}
                                        type="button"
                                        variant="ghost"
                                    >
                                        {showCurrentPassword ? (
                                            <EyeOff className="text-muted-foreground size-4" />
                                        ) : (
                                            <Eye className="text-muted-foreground size-4" />
                                        )}
                                    </Button>
                                </div>
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
                                <div className="relative">
                                    <Input
                                        {...field}
                                        placeholder="Enter new password"
                                        type={showNewPassword ? 'text' : 'password'}
                                    />
                                    <Button
                                        className="absolute top-0 right-0 h-full px-3 py-2 hover:bg-transparent"
                                        onClick={() => setShowNewPassword(!showNewPassword)}
                                        size="sm"
                                        tabIndex={-1}
                                        type="button"
                                        variant="ghost"
                                    >
                                        {showNewPassword ? (
                                            <EyeOff className="text-muted-foreground size-4" />
                                        ) : (
                                            <Eye className="text-muted-foreground size-4" />
                                        )}
                                    </Button>
                                </div>
                            </FormControl>
                            <FormDescription className="text-xs">
                                Must be 16+ characters, or 8+ with number, lowercase, uppercase, and special character
                                (!@#$&*)
                            </FormDescription>
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
                                <div className="relative">
                                    <Input
                                        {...field}
                                        placeholder="Confirm new password"
                                        type={showConfirmPassword ? 'text' : 'password'}
                                    />
                                    <Button
                                        className="absolute top-0 right-0 h-full px-3 py-2 hover:bg-transparent"
                                        onClick={() => setShowConfirmPassword(!showConfirmPassword)}
                                        size="sm"
                                        tabIndex={-1}
                                        type="button"
                                        variant="ghost"
                                    >
                                        {showConfirmPassword ? (
                                            <EyeOff className="text-muted-foreground size-4" />
                                        ) : (
                                            <Eye className="text-muted-foreground size-4" />
                                        )}
                                    </Button>
                                </div>
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
