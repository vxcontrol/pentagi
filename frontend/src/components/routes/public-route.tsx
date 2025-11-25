import * as React from 'react';
import { Navigate } from 'react-router-dom';

import { useUser } from '@/providers/user-provider';

const PublicRoute = ({ children }: { children: React.ReactNode }) => {
    const { isAuthenticated, isLoading } = useUser();

    // Wait for initial auth check to complete
    if (isLoading) {
        return null;
    }

    if (isAuthenticated()) {
        return (
            <Navigate
                replace
                to="/flows/new"
            />
        );
    }

    return children;
};

export default PublicRoute;
