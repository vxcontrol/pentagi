import * as React from 'react';
import { Navigate, useSearchParams } from 'react-router-dom';

import { routes } from '@/lib/routes';
import { getSafeReturnUrl } from '@/lib/utils/auth';
import { useUser } from '@/providers/user-provider';

function PublicRoute({ children }: { children: React.ReactNode }) {
    const [searchParams] = useSearchParams();
    const { authInfo, isAuthenticated, isLoading } = useUser();

    if (isLoading) {
        return null;
    }

    if (isAuthenticated()) {
        if (
            authInfo?.user?.password_change_required &&
            authInfo?.type === 'user' &&
            authInfo?.user?.type === 'local' // Only local users have password_change_required
        ) {
            return children;
        }

        const returnUrl = getSafeReturnUrl(searchParams.get('returnUrl'), routes.newFlow);

        return (
            <Navigate
                replace
                to={returnUrl}
            />
        );
    }

    return children;
}

export default PublicRoute;
