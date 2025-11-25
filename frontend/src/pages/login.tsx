import { Loader2 } from 'lucide-react';
import { useLocation, useSearchParams } from 'react-router-dom';

import Logo from '@/components/icons/logo';
import LoginForm from '@/features/authentication/login-form';
import { useUser } from '@/providers/user-provider';

const Login = () => {
    const [searchParams] = useSearchParams();
    const location = useLocation();
    const { authInfo, isLoading } = useUser();
    const authProviders = authInfo?.providers || [];

    // Extract the return URL from either location state or query parameters
    const returnUrl = location.state?.from || searchParams.get('returnUrl') || '/flows/new';

    return (
        <div className="flex h-dvh w-full items-center justify-center">
            <div className="h-dvh w-full lg:grid lg:grid-cols-2">
                <div className="flex items-center justify-center px-4 py-12">
                    {!isLoading ? (
                        <LoginForm
                            providers={authProviders}
                            returnUrl={returnUrl}
                        />
                    ) : (
                        <Loader2 className="size-16 animate-spin" />
                    )}
                </div>
                <div className="hidden bg-gradient-to-r from-slate-800 to-slate-950 lg:flex">
                    <Logo className="m-auto size-32 animate-logo-spin text-white delay-10000" />
                </div>
            </div>
        </div>
    );
};

export default Login;
