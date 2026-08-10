import { TriangleAlert } from 'lucide-react';
import { useEffect } from 'react';
import { useRouteError } from 'react-router-dom';

import { Button } from '@/components/ui/button';
import { Empty, EmptyContent, EmptyDescription, EmptyHeader, EmptyMedia, EmptyTitle } from '@/components/ui/empty';
import { useI18n } from '@/hooks/use-i18n';
import { isChunkLoadError, isDomDesyncError, reloadOnce } from '@/lib/chunk-reload';

/**
 * Root `errorElement` for the data router — replaces React Router's built-in
 * dev crash screen. Two production failures land here: a code-split chunk that
 * 404s after a redeploy (auto-reload once, unless `vite:preloadError` in
 * `main.tsx` already did), and any other render/commit crash (e.g. a `removeChild`
 * desync from browser auto-translation) — both shown as a recoverable card.
 */
function RouteErrorBoundary() {
    const { t } = useI18n();
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
            <Empty>
                <EmptyHeader>
                    <EmptyMedia variant="icon">
                        <TriangleAlert />
                    </EmptyMedia>
                    <EmptyTitle>{t('errors.somethingWentWrong')}</EmptyTitle>
                    <EmptyDescription>
                        {isChunk
                            ? t('errors.chunkLoadError')
                            : isDesync
                              ? t('errors.domDesyncError')
                              : t('errors.unexpectedError')}
                    </EmptyDescription>
                </EmptyHeader>
                <EmptyContent>
                    <Button
                        onClick={() => window.location.reload()}
                        variant="secondary"
                    >
                        {t('common.reload')}
                    </Button>
                </EmptyContent>
            </Empty>
        </div>
    );
}

export default RouteErrorBoundary;
