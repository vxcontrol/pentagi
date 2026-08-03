import * as React from 'react';
import { Navigate, useLocation } from 'react-router-dom';

import { routes } from '@/lib/routes';
import { useUser } from '@/providers/user-provider';

function ProtectedRoute({ children }: { children: React.ReactNode }) {
    const location = useLocation();
    const { isAuthenticated, isLoading } = useUser();

    if (isLoading) {
        return null;
    }

    if (!isAuthenticated()) {
        return (
            <Navigate
                replace
                to={routes.login(location.pathname)}
            />
        );
    }

    return children;
}

export default ProtectedRoute;
