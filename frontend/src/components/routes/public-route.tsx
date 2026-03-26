import * as React from 'react';
import { Navigate, useSearchParams } from 'react-router-dom';

import { getSafeReturnUrl } from '@/lib/utils/auth';
import { useUser } from '@/providers/user-provider';

const PublicRoute = ({ children }: { children: React.ReactNode }) => {
    const [searchParams] = useSearchParams();
    const { authInfo, isAuthenticated, isLoading } = useUser();

    // Wait for initial auth check to complete
    if (isLoading) {
        return null;
    }

    if (isAuthenticated()) {
        // Only show password change form if the user is ACTUALLY authenticated
        // with a valid, non-expired session. Do NOT rely solely on authInfo presence in
        // memory, because clearAuth() is async and during race conditions (e.g., when
        // session expires and user refreshes the page) the old authInfo may still be in
        // state while localStorage is already cleared.
        //
        // Additional safety check: verify that authInfo.type is 'user', not 'guest'.
        // If server returned guest status, we should NOT show password change form.
        if (
            authInfo?.user?.password_change_required &&
            authInfo?.type === 'user' &&
            authInfo?.user?.type === 'local' // Only local users have password_change_required
        ) {
            return children;
        }

        const returnUrl = getSafeReturnUrl(searchParams.get('returnUrl'), '/flows/new');

        return (
            <Navigate
                replace
                to={returnUrl}
            />
        );
    }

    return children;
};

export default PublicRoute;
