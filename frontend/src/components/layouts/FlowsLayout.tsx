import { Outlet } from 'react-router-dom';

import { FlowsProvider } from '@/providers/FlowsProvider';

const FlowsLayout = () => {
    return (
        <FlowsProvider>
            <Outlet />
        </FlowsProvider>
    );
};

export default FlowsLayout;
