import { Loader2 } from 'lucide-react';
import { useEffect, useState } from 'react';
import { useLocation, useSearchParams } from 'react-router-dom';

import Logo from '@/components/icons/Logo';
import LoginForm from '@/features/authentication/LoginForm';
import { axios } from '@/lib/axios';
import type { AuthInfoResponse } from '@/models/Info';

const Login = () => {
    const [providers, setProviders] = useState<string[]>([]);
    const [isLoading, setIsLoading] = useState(true);
    const [searchParams] = useSearchParams();
    const location = useLocation();

    // Extract the return URL from either location state or query parameters
    const returnUrl = location.state?.from || searchParams.get('returnUrl') || '/chat/new';

    useEffect(() => {
        const getInfo = async () => {
            try {
                const info: AuthInfoResponse = await axios.get('/info');
                if (info?.status === 'success' && info?.data?.providers?.length) {
                    // OAuth providers are returned as string[] from API
                    setProviders(info.data.providers);
                }
            } catch {
                // ignore
            } finally {
                setIsLoading(false);
            }
        };

        getInfo();
    }, []);

    return (
        <div className="flex h-dvh w-full items-center justify-center">
            <div className="h-dvh w-full lg:grid lg:grid-cols-2">
                <div className="flex items-center justify-center px-4 py-12">
                    {!isLoading ? <LoginForm providers={providers} returnUrl={returnUrl} /> : <Loader2 className="size-16 animate-spin" />}
                </div>
                <div className="hidden bg-gradient-to-r from-slate-800 to-slate-950 lg:flex">
                    <Logo className="m-auto size-32 animate-logo-spin text-white delay-10000" />
                </div>
            </div>
        </div>
    );
};

export default Login;
