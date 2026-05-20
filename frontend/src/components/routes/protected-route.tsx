import * as React from 'react';
import { Navigate, useLocation } from 'react-router-dom';

import { getReturnUrlParam } from '@/lib/utils/auth';
import { useUser } from '@/providers/user-provider';

function ProtectedRoute({ children }: { children: React.ReactNode }) {
    const location = useLocation();
    const { isAuthenticated, isLoading } = useUser();

    if (isLoading) {
        return null;
    }

    if (!isAuthenticated()) {
        const returnParam = getReturnUrlParam(location.pathname);

        return (
            <Navigate
                replace
                to={`/login${returnParam}`}
            />
        );
    }

    return children;
}

export default ProtectedRoute;
