import { fireEvent, render, screen } from '@testing-library/react';
import { useState } from 'react';
import { createBrowserRouter, RouterProvider, useLocation, useSearchParams } from 'react-router-dom';
import { describe, expect, it } from 'vitest';

import { useSearchParamTab } from './use-search-param-tab';

function Panel() {
    const [graphTab] = useSearchParamTab({
        fallback: 'overview',
        hasVolatileValues: true,
        key: 'graph',
        values: ['overview'],
    });

    return <span>{graphTab}</span>;
}

function Subject() {
    const [searchParams, setSearchParams] = useSearchParams();
    const [manualTab, setManualTab] = useState<null | string>(null);
    const tab = manualTab ?? searchParams.get('tab');

    return (
        <>
            <span data-slot="probe-search">{useLocation().search}</span>
            <button
                onClick={() => {
                    // Drop this and both writers no longer land in one commit: the gate stops failing.
                    setManualTab('dashboard');
                    setSearchParams((previous) => {
                        const params = new URLSearchParams(previous);

                        params.set('tab', 'dashboard');

                        return params;
                    });
                }}
                type="button"
            >
                dashboard
            </button>
            {tab === 'dashboard' && <Panel />}
        </>
    );
}

// A memory router keeps its location off `window`, so the live read has nothing to find and this
// gate passes on broken code.
const renderAt = (entry: string) => {
    window.history.replaceState({}, '', entry);

    return render(<RouterProvider router={createBrowserRouter([{ element: <Subject />, path: '/flows/:flowId' }])} />);
};

describe('two hooks writing the query string in one commit', () => {
    it('keeps the tab a click just asked for while the panel it opened pins its own value', () => {
        renderAt('/flows/1?tab=assistant&view=analytics');

        fireEvent.click(screen.getByRole('button', { name: 'dashboard' }));

        expect(screen.getByTestId('probe-search').textContent).toContain('graph=overview');
        expect(screen.getByTestId('probe-search').textContent).toContain('tab=dashboard');
    });

    it('writes the resolved tab into the URL when it names none', () => {
        renderAt('/flows/1?tab=dashboard&view=analytics');

        expect(screen.getByTestId('probe-search').textContent).toContain('graph=overview');
    });
});
