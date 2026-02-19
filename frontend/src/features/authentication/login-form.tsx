import { zodResolver } from '@hookform/resolvers/zod';
import { Loader2 } from 'lucide-react';
import { useState } from 'react';
import { useForm } from 'react-hook-form';
import { useNavigate } from 'react-router-dom';
import { z } from 'zod';

import type { OAuthProvider } from '@/providers/user-provider';

import Github from '@/components/icons/github';
import Google from '@/components/icons/google';
import { Button } from '@/components/ui/button';
import { Form, FormControl, FormField, FormItem, FormLabel, FormMessage } from '@/components/ui/form';
import { Input } from '@/components/ui/input';
import { useUser } from '@/providers/user-provider';

import { PasswordChangeForm } from './password-change-form';

const formSchema = z.object({
    mail: z
        .string()
        .min(1, {
            message: 'Login is required',
        })
        .refine(
            (value) => z.string().email().safeParse(value).success || ['admin', 'demo'].includes(value.toLowerCase()),
            {
                message: 'Invalid login',
            },
        ),
    password: z.string().min(1, {
        message: 'Password is required',
    }),
});

const errorMessage = 'Invalid login or password';
const errorProviderMessage = 'Authentication failed';

interface AuthProviderAction {
    icon: React.ReactNode;
    id: OAuthProvider;
    name: string;
}

const providerActions: AuthProviderAction[] = [
    {
        icon: <Google className="size-5" />,
        id: 'google',
        name: 'Continue with Google',
    },
    {
        icon: <Github className="size-5" />,
        id: 'github',
        name: 'Continue with GitHub',
    },
];

interface LoginFormProps {
    providers: string[]; // OAuth providers: ['google', 'github']
    returnUrl?: string;
}

const LoginForm = ({ providers, returnUrl = '/flows/new' }: LoginFormProps) => {
    const form = useForm<z.infer<typeof formSchema>>({
        defaultValues: {
            mail: '',
            password: '',
        },
        resolver: zodResolver(formSchema),
    });
    const [isSubmitting, setIsSubmitting] = useState(false);
    const [error, setError] = useState<null | string>(null);
    const [passwordChangeRequired, setPasswordChangeRequired] = useState(false);
    const navigate = useNavigate();
    const { authInfo, isAuthenticated, login, loginWithOAuth, setAuth } = useUser();

    const handleSubmit = async (values: z.infer<typeof formSchema>) => {
        setError(null);
        setIsSubmitting(true);

        try {
            const result = await login(values);

            if (!result.success) {
                setError(result.error || errorMessage);

                return;
            }

            if (result.passwordChangeRequired) {
                setPasswordChangeRequired(true);

                return;
            }

            navigate(returnUrl);
        } catch {
            setError(errorMessage);
        } finally {
            setIsSubmitting(false);
        }
    };

    const handleProviderLogin = async (provider: OAuthProvider) => {
        setError(null);
        setIsSubmitting(true);

        try {
            const result = await loginWithOAuth(provider);

            if (!result.success) {
                setError(result.error || errorProviderMessage);

                return;
            }

            navigate(returnUrl);
        } catch (error) {
            setError(error instanceof Error ? error.message : errorMessage);
        } finally {
            setIsSubmitting(false);
        }
    };

    const handleSkipPasswordChange = () => {
        navigate(returnUrl);
    };

    const handlePasswordChangeSuccess = () => {
        if (authInfo?.user) {
            // Update auth info with password_change_required set to false
            const updatedAuthData = {
                ...authInfo,
                user: {
                    ...authInfo.user,
                    password_change_required: false,
                },
            };

            setAuth(updatedAuthData);
            navigate(returnUrl);
        }
    };

    // If password change is required, show password change form.
    // Also check isAuthenticated() to ensure the user has a valid session.
    // If the session expired and user refreshed the page, the old authInfo may still
    // be in memory (race condition between clearAuth() and navigate()), but we must
    // NOT show the password change form because:
    //   1. The API endpoint /user/password requires authentication (returns 403 if not)
    //   2. The user must first re-login to establish a new valid session
    // Also check authInfo directly to handle page refresh scenarios where passwordChangeRequired
    // local state is lost but authInfo.user.password_change_required is still true.
    const shouldShowPasswordChange =
        (passwordChangeRequired || authInfo?.user?.password_change_required) &&
        authInfo?.user?.type === 'local' &&
        isAuthenticated();

    if (shouldShowPasswordChange) {
        return (
            <div className="mx-auto flex w-[350px] flex-col gap-6">
                <h1 className="text-center text-3xl font-bold">Update Password</h1>
                <p className="text-muted-foreground text-center text-sm">
                    You need to change your password before continuing.
                </p>
                <PasswordChangeForm
                    isModal={false}
                    onSkip={handleSkipPasswordChange}
                    onSuccess={handlePasswordChangeSuccess}
                    showSkip={true}
                />
            </div>
        );
    }

    return (
        <Form {...form}>
            <form
                className="mx-auto grid w-[350px] gap-8"
                onSubmit={form.handleSubmit(handleSubmit)}
            >
                <h1 className="text-center text-3xl font-bold">PentAGI</h1>

                {providers?.length > 0 && (
                    <>
                        <div className="flex flex-col gap-4">
                            {providerActions
                                .filter((provider) => providers.includes(provider.id))
                                .map((provider) => (
                                    <Button
                                        disabled={isSubmitting}
                                        key={provider.id}
                                        onClick={() => handleProviderLogin(provider.id)}
                                        type="button"
                                        variant="secondary"
                                    >
                                        {provider.icon}
                                        {provider.name}
                                    </Button>
                                ))}
                        </div>

                        <div className="relative -mb-4">
                            <div className="absolute inset-0 flex items-center">
                                <div className="w-full border-t border-gray-300" />
                            </div>
                            <div className="relative flex justify-center text-sm">
                                <span className="bg-background px-2">or</span>
                            </div>
                        </div>
                    </>
                )}

                <div className="flex flex-col gap-4">
                    <FormField
                        control={form.control}
                        name="mail"
                        render={({ field }) => (
                            <FormItem>
                                <FormLabel>Login</FormLabel>
                                <FormControl>
                                    <Input
                                        {...field}
                                        autoFocus
                                        placeholder="Enter your email"
                                    />
                                </FormControl>
                                <FormMessage />
                            </FormItem>
                        )}
                    />

                    <FormField
                        control={form.control}
                        name="password"
                        render={({ field }) => (
                            <FormItem>
                                <FormLabel>Password</FormLabel>
                                <FormControl>
                                    <Input
                                        {...field}
                                        placeholder="Enter your password"
                                        type="password"
                                    />
                                </FormControl>
                                <FormMessage />
                            </FormItem>
                        )}
                    />

                    <Button
                        className="w-full"
                        disabled={isSubmitting || (!form.formState.isValid && form.formState.isSubmitted)}
                        type="submit"
                    >
                        {isSubmitting && <Loader2 className="animate-spin" />}
                        <span>Sign in</span>
                    </Button>

                    {error && <FormMessage>{error}</FormMessage>}
                </div>
            </form>
        </Form>
    );
};

export default LoginForm;
