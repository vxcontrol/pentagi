import * as React from 'react';
import { Navigate, useLocation } from 'react-router-dom';

import { isAuthenticated } from '@/lib/auth';

const ProtectedRoute = ({ children }: { children: React.ReactNode }) => {
    const location = useLocation();

    if (!isAuthenticated()) {
        // Save current path for redirect after login
        const currentPath = location.pathname;
        // Only save if it's not the default route
        const returnParam = currentPath !== '/flows/new' ? `?returnUrl=${encodeURIComponent(currentPath)}` : '';
        return <Navigate to={`/login${returnParam}`} />;
    }

    return children;
};

export default ProtectedRoute;
