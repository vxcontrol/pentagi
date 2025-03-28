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
import { axios } from '@/lib/axios';
import { baseUrl } from '@/models/Api';
import type { AuthInfoResponse, AuthLoginResponse } from '@/models/Info';
import type { User } from '@/models/User';

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

const returnOAuthUri = '/oauth/result';
const errorMessage = 'Invalid login or password';
const errorProviderMessage = 'Authentication failed';

type Provider = 'google' | 'github';

interface AuthProviderAction {
    id: Provider;
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
    providers: string[];
    returnUrl?: string;
}

const LoginForm = ({ providers, returnUrl = '/chat/new' }: LoginFormProps) => {
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
    const [user, setUser] = useState<User | null>(null);
    const navigate = useNavigate();

    const providerLoginCheckInterval = 500;
    const providerLoginTimeout = 300000;

    const handleSubmit = async (values: z.infer<typeof formSchema>) => {
        setError(null);
        setIsSubmitting(true);

        try {
            const data: AuthLoginResponse = await axios.post('/auth/login', values);

            if (data?.status !== 'success') {
                setError(errorMessage);
                return;
            }

            const info: AuthInfoResponse = await axios.get('/info');

            if (info?.status !== 'success') {
                setError(errorMessage);
                return;
            }

            localStorage.setItem('auth', JSON.stringify(info.data));

            // Check if password change is required for local users
            if (info.data && info.data?.user?.type === 'local' && info.data?.user?.password_change_required) {
                setUser(info.data?.user as unknown as User);
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

    const handleProviderLoginPopupOpen = async (provider: Provider): Promise<AuthInfoResponse> => {
        const width = 500;
        const height = 600;
        const left = window.screenX + (window.outerWidth - width) / 2;
        const top = window.screenY + (window.outerHeight - height) / 2;

        const popup = window.open(
            `${baseUrl}/auth/authorize?provider=${provider}&return_uri=${returnOAuthUri}`,
            `${provider} Sign In`,
            `width=${width},height=${height},left=${left},top=${top}`,
        );

        if (!popup) {
            return {
                status: 'error',
                error: 'Popup blocked. Please allow popups for this site.',
            };
        }

        return new Promise<AuthInfoResponse>((resolve) => {
            const messageHandler = async (event: MessageEvent) => {
                if (event.origin !== window.location.origin || event.data?.type !== 'oauth-result') {
                    return;
                }

                window.removeEventListener('message', messageHandler);

                const handlerResolve = (value: AuthInfoResponse) => {
                    if (popup && !popup.closed) {
                        popup.close();
                    }

                    resolve(value);
                };

                if (event.data.status === 'success') {
                    try {
                        const info: AuthInfoResponse = await axios.get('/info');

                        if (info?.status === 'success' && info.data?.type === 'user') {
                            handlerResolve(info);
                            return;
                        }
                    } catch {
                        // In case of error, fall through to common handling below
                    }
                }

                handlerResolve({
                    status: 'error',
                    error: event.data.error || errorProviderMessage,
                });
            };

            window.addEventListener('message', messageHandler);

            const popupCheck = setInterval(() => {
                if (popup?.closed) {
                    clearInterval(popupCheck);
                    window.removeEventListener('message', messageHandler);
                    resolve({
                        status: 'error',
                        error: 'Authentication cancelled',
                    });
                }
            }, providerLoginCheckInterval);

            setTimeout(() => {
                clearInterval(popupCheck);
                window.removeEventListener('message', messageHandler);

                if (popup && !popup.closed) {
                    popup.close();
                }

                resolve({
                    status: 'error',
                    error: 'Authentication timeout',
                });
            }, providerLoginTimeout);
        });
    };

    const handleProviderLogin = async (provider: Provider) => {
        setError(null);
        setIsSubmitting(true);

        try {
            const info = await handleProviderLoginPopupOpen(provider);

            if (info?.status !== 'success') {
                setError(info.error || errorProviderMessage);
                return;
            }

            localStorage.setItem('auth', JSON.stringify(info.data));

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
        if (user) {
            // Update stored user info with password_change_required set to false
            const updatedUser = {
                ...user,
                password_change_required: false,
            };
            localStorage.setItem('auth', JSON.stringify(updatedUser));
            navigate(returnUrl);
        }
    };

    // If password change is required, show password change form
    if (passwordChangeRequired && user?.type === 'local') {
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
