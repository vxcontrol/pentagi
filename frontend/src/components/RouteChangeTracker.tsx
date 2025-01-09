import { useEffect } from 'react';
import { useLocation, useNavigate } from 'react-router-dom';

const RouteChangeTracker = ({ onRouteChange }) => {
    const navigate = useNavigate();
    const location = useLocation();

    useEffect(() => {
        onRouteChange({ navigate, location });
    }, [navigate, location, onRouteChange]);

    return null;
};

export default RouteChangeTracker;
