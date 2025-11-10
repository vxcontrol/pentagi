import * as React from 'react';
import { Navigate } from 'react-router-dom';

import { useUser } from '@/providers/UserProvider';

const PublicRoute = ({ children }: { children: React.ReactNode }) => {
    const { isAuthenticated, isLoading } = useUser();

    // Wait for initial auth check to complete
    if (isLoading) {
        return null;
    }

    if (isAuthenticated()) {
        return (
            <Navigate
                to="/flows/new"
                replace
            />
        );
    }

    return children;
};

export default PublicRoute;
