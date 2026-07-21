import { TriangleAlert } from 'lucide-react';
import { useEffect } from 'react';
import { useRouteError } from 'react-router-dom';

import { Button } from '@/components/ui/button';
import { StatusCard } from '@/components/ui/status-card';
import { isChunkLoadError, isDomDesyncError, reloadOnce } from '@/lib/chunk-reload';

/**
 * Root `errorElement` for the data router — replaces React Router's built-in
 * dev crash screen. Two production failures land here: a code-split chunk that
 * 404s after a redeploy (auto-reload once, unless `vite:preloadError` in
 * `main.tsx` already did), and any other render/commit crash (e.g. a `removeChild`
 * desync from browser auto-translation) — both shown as a recoverable card.
 */
function RouteErrorBoundary() {
    const error = useRouteError();
    const isChunk = isChunkLoadError(error);
    const isDesync = isDomDesyncError(error);

    useEffect(() => {
        if (isChunk || isDesync) {
            reloadOnce();
        }
    }, [isChunk, isDesync]);

    return (
        <div
            className="grid min-h-svh w-full place-items-center p-4"
            role="alert"
        >
            <StatusCard
                action={
                    <Button
                        onClick={() => window.location.reload()}
                        variant="secondary"
                    >
                        Reload
                    </Button>
                }
                description={
                    isChunk
                        ? 'A new version was likely just deployed. Reloading will load the latest one.'
                        : isDesync
                          ? 'The page hit a display glitch. Reloading usually clears it.'
                          : 'The page ran into an unexpected error. Reloading usually clears it.'
                }
                icon={
                    <TriangleAlert
                        aria-hidden
                        className="text-muted-foreground size-10"
                    />
                }
                title="Something went wrong"
            />
        </div>
    );
}

export default RouteErrorBoundary;
