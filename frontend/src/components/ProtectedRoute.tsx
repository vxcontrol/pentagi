import * as React from 'react';
import { Navigate } from 'react-router-dom';

import { isAuthenticated } from '@/lib/auth';

const ProtectedRoute = ({ children }: { children: React.ReactNode }) => {
    if (!isAuthenticated()) {
        return <Navigate to="/login" />;
    }

    return children;
};

export default ProtectedRoute;
