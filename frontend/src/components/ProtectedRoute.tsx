import * as React from 'react';
import { Navigate, useLocation } from 'react-router-dom';

import { useUser } from '@/providers/UserProvider';

const ProtectedRoute = ({ children }: { children: React.ReactNode }) => {
    const location = useLocation();
    const { isAuthenticated } = useUser();

    if (!isAuthenticated()) {
        // Save current path for redirect after login
        const currentPath = location.pathname;
        // Only save if it's not the default route
        const returnParam = currentPath !== '/flows/new' ? `?returnUrl=${encodeURIComponent(currentPath)}` : '';
        return (
            <Navigate
                to={`/login${returnParam}`}
                replace
            />
        );
    }

    return children;
};

export default ProtectedRoute;
