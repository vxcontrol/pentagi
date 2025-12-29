import { Outlet } from 'react-router-dom';

import { FlowsProvider } from '@/providers/flows-provider';

const FlowsLayout = () => {
    return (
        <FlowsProvider>
            <Outlet />
        </FlowsProvider>
    );
};

export default FlowsLayout;
