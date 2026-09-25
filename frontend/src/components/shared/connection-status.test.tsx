import { act, render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';

import { ConnectionStatus } from './connection-status';

const fire = (name: string) => act(() => void window.dispatchEvent(new Event(name)));

describe('the reader learns their live data stopped', () => {
    it('says nothing while the socket is up', () => {
        render(<ConnectionStatus />);

        expect(screen.queryByRole('status')).toBeNull();
    });

    it('speaks up when the socket drops and goes quiet again once it is back', () => {
        render(<ConnectionStatus />);

        fire('ws:disconnected');

        expect(screen.getByRole('status')).toHaveTextContent(/not receiving updates/i);

        fire('ws:connected');

        expect(screen.queryByRole('status')).toBeNull();
    });
});
