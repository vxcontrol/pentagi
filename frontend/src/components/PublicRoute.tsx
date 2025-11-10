import * as React from 'react';
import { Navigate } from 'react-router-dom';

import { useUser } from '@/providers/UserProvider';

const PublicRoute = ({ children }: { children: React.ReactNode }) => {
    const { isAuthenticated } = useUser();

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
