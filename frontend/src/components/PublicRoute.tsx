import * as React from 'react';
import { Navigate } from 'react-router-dom';

import { isAuthenticated } from '@/lib/auth';

const PublicRoute = ({ children }: { children: React.ReactNode }) => {
    if (isAuthenticated()) {
        return <Navigate to="/" />;
    }

    return children;
};

export default PublicRoute;
