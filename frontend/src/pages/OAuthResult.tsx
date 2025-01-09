import { useEffect } from 'react';

import Logo from '@/components/icons/Logo';
import ThemeProvider from '@/providers/ThemeProvider';

const OAuthResult = () => {
    useEffect(() => {
        const params = new URLSearchParams(window.location.search);
        const status = params.get('status');
        const error = params.get('error');

        if (window.opener) {
            window.opener.postMessage(
                {
                    type: 'oauth-result',
                    status,
                    error,
                },
                window.location.origin,
            );

            const closeWindow = new Promise<void>((resolve) => {
                const timer = setTimeout(() => {
                    if (window && !window.closed) {
                        window.close();
                    }
                    resolve();
                }, 10000);

                return () => clearTimeout(timer);
            });

            closeWindow.catch(() => window.close());
        }
    }, []);

    return (
        <ThemeProvider>
            <div className="flex h-screen w-full items-center justify-center bg-gradient-to-r from-slate-800 to-slate-950">
                <Logo className="m-auto size-32 animate-logo-spin text-white delay-10000" />
            </div>
        </ThemeProvider>
    );
};

export default OAuthResult;
