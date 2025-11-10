import { zodResolver } from '@hookform/resolvers/zod';
import { Loader2 } from 'lucide-react';
import { useState } from 'react';
import { useForm } from 'react-hook-form';
import { useNavigate } from 'react-router-dom';
import { z } from 'zod';

import Github from '@/components/icons/Github';
import Google from '@/components/icons/Google';
import { Button } from '@/components/ui/button';
import { Form, FormControl, FormField, FormItem, FormLabel, FormMessage } from '@/components/ui/form';
import { Input } from '@/components/ui/input';
import { type OAuthProvider, useUser } from '@/providers/UserProvider';

import { PasswordChangeForm } from './PasswordChangeForm';

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
    id: OAuthProvider;
    name: string;
    icon: React.ReactNode;
}

const providerActions: AuthProviderAction[] = [
    {
        id: 'google',
        name: 'Continue with Google',
        icon: <Google className="size-5" />,
    },
    {
        id: 'github',
        name: 'Continue with GitHub',
        icon: <Github className="size-5" />,
    },
];

interface LoginFormProps {
    providers: string[]; // OAuth providers: ['google', 'github']
    returnUrl?: string;
}

const LoginForm = ({ providers, returnUrl = '/flows/new' }: LoginFormProps) => {
    const form = useForm<z.infer<typeof formSchema>>({
        resolver: zodResolver(formSchema),
        defaultValues: {
            mail: '',
            password: '',
        },
    });
    const [isSubmitting, setIsSubmitting] = useState(false);
    const [error, setError] = useState<string | null>(null);
    const [passwordChangeRequired, setPasswordChangeRequired] = useState(false);
    const navigate = useNavigate();
    const { authInfo, setAuth, login, loginWithOAuth } = useUser();

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

    // If password change is required, show password change form
    if (passwordChangeRequired && authInfo?.user?.type === 'local') {
        return (
            <div className="mx-auto w-[350px] space-y-6">
                <h1 className="text-center text-3xl font-bold">Update Password</h1>
                <p className="text-center text-sm text-muted-foreground">
                    You need to change your password before continuing.
                </p>
                <PasswordChangeForm
                    onSuccess={handlePasswordChangeSuccess}
                    showSkip={true}
                    onSkip={handleSkipPasswordChange}
                    isModal={false}
                />
            </div>
        );
    }

    return (
        <Form {...form}>
            <form
                onSubmit={form.handleSubmit(handleSubmit)}
                className="mx-auto grid w-[350px] gap-8"
            >
                <h1 className="text-center text-3xl font-bold">PentAGI</h1>

                {providers?.length > 0 && (
                    <>
                        <div className="flex flex-col gap-4">
                            {providerActions
                                .filter((provider) => providers.includes(provider.id))
                                .map((provider) => (
                                    <Button
                                        key={provider.id}
                                        type="button"
                                        variant="secondary"
                                        onClick={() => handleProviderLogin(provider.id)}
                                        disabled={isSubmitting}
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

                <div className="grid gap-4">
                    <FormField
                        control={form.control}
                        name="mail"
                        render={({ field }) => (
                            <FormItem>
                                <FormLabel>Login</FormLabel>
                                <FormControl>
                                    <Input
                                        {...field}
                                        placeholder="Enter your email"
                                        autoFocus
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
                                        type="password"
                                        placeholder="Enter your password"
                                    />
                                </FormControl>
                                <FormMessage />
                            </FormItem>
                        )}
                    />

                    <Button
                        type="submit"
                        className="w-full"
                        disabled={isSubmitting || (!form.formState.isValid && form.formState.isSubmitted)}
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
