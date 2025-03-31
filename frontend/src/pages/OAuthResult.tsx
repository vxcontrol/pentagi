import { useEffect, useLayoutEffect, useRef, useState } from 'react';

import Logo from '@/components/icons/Logo';
import ThemeProvider from '@/providers/ThemeProvider';

const OAuthResult = () => {
    const [statusMessage, setStatusMessage] = useState('Authentication in progress...');
    const messageRef = useRef(statusMessage);
    const prevMessageRef = useRef(statusMessage);

    // Success delay is short, error delay is longer to allow reading
    const successDelay = 2000;
    const errorDelay = 5000;

    // Synchronize state with ref without problematic dependencies
    useLayoutEffect(() => {
        if (prevMessageRef.current !== messageRef.current) {
            setStatusMessage(messageRef.current);
            prevMessageRef.current = messageRef.current;
        }
    }, []);

    useEffect(() => {
        const params = new URLSearchParams(window.location.search);
        const status = params.get('status');
        const error = params.get('error');

        // This is used to track all timeouts that need to be cleared on cleanup
        let redirectTimer: NodeJS.Timeout | null = null;
        let cleanupTimer: NodeJS.Timeout | null = null;
        let closeTimer: NodeJS.Timeout | null = null;

        const updateMessage = (message: string) => {
            messageRef.current = message;
        };

        // Handle window close safely
        const handleClose = (delay: number) => {
            closeTimer = setTimeout(() => {
                try {
                    if (window && !window.closed) {
                        window.close();
                    }
                } catch (e) {
                    console.error('Delayed window close failed:', e);
                }
            }, delay);
        };

        // Handle redirection if needed
        const handleRedirect = (url: string, delay: number) => {
            redirectTimer = setTimeout(() => {
                try {
                    window.location.href = url;
                } catch (e) {
                    console.error('Redirection failed:', e);
                }
            }, delay);

            // Ensure redirect timer gets cleaned up
            cleanupTimer = setTimeout(() => {
                if (redirectTimer) {
                    clearTimeout(redirectTimer);
                    redirectTimer = null;
                }
            }, delay + 100);
        };

        if (window.opener) {
            try {
                window.opener.postMessage(
                    {
                        type: 'oauth-result',
                        status,
                        error,
                    },
                    window.location.origin,
                );

                // Success handling
                updateMessage('Authentication complete, closing window...');
                handleClose(successDelay);
            } catch (e) {
                console.error('Failed to send message to opener:', e);
                updateMessage('Error communicating with parent window. Closing in a few seconds...');
                handleClose(errorDelay);
            }
        } else {
            // If no opener, redirect to login
            updateMessage('Authentication window opened directly. Redirecting to login page...');
            handleRedirect('/login', errorDelay / 2);
            handleClose(errorDelay);
        }

        // Cleanup function for useEffect
        return () => {
            // Explicitly clear all timeouts
            if (redirectTimer) clearTimeout(redirectTimer);
            if (cleanupTimer) clearTimeout(cleanupTimer);
            if (closeTimer) clearTimeout(closeTimer);
        };
    }, [successDelay, errorDelay]);

    return (
        <ThemeProvider>
            <div className="flex h-screen w-full items-center justify-center bg-gradient-to-r from-slate-800 to-slate-950">
                <Logo className="m-auto size-32 animate-logo-spin text-white delay-10000" />
                <div className="fixed bottom-4 text-sm text-white">{statusMessage}</div>
            </div>
        </ThemeProvider>
    );
};

export default OAuthResult;
