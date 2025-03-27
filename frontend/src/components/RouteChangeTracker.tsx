import { useEffect } from 'react';
import type { Location, NavigateFunction } from 'react-router-dom';
import { useLocation, useNavigate } from 'react-router-dom';

interface RouteChangeTrackerProps {
    onRouteChange: (params: { navigate: NavigateFunction; location: Location }) => void;
}

const RouteChangeTracker = ({ onRouteChange }: RouteChangeTrackerProps) => {
    const navigate = useNavigate();
    const location = useLocation();

    useEffect(() => {
        onRouteChange({ navigate, location });
    }, [navigate, location, onRouteChange]);

    return null;
};

export default RouteChangeTracker;
