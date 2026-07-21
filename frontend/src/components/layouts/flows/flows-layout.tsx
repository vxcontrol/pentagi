import { Outlet } from 'react-router-dom';

import { FlowsProvider } from '@/providers/flows-provider';

function FlowsLayout() {
    return (
        <FlowsProvider>
            <Outlet />
        </FlowsProvider>
    );
}

export default FlowsLayout;
