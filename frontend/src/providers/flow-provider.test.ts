import { describe, expect, it } from 'vitest';

import { deriveFlowMissing } from './flow-provider';

const flow = { flow: { id: '1' } };
const notFound = new Error('flow not found');
const partialNotFound = new Error('assistant not found'); // a nested resolver under errorPolicy:'all'

describe('deriveFlowMissing', () => {
    it('is true when the query settled with the flow genuinely absent', () => {
        expect(deriveFlowMissing({ flow: null }, undefined)).toBe(true);
    });

    it('is true on a not-found error with no flow loaded', () => {
        expect(deriveFlowMissing(undefined, notFound)).toBe(true);
    });

    // Reverting the `!flowData?.flow` gate turns this red — it is the regression guard.
    it('is false when a not-found error rides alongside a loaded flow', () => {
        expect(deriveFlowMissing(flow, partialNotFound)).toBe(false);
    });

    it('is false on a real load failure (kept behind Retry, not redirected)', () => {
        expect(deriveFlowMissing(undefined, new Error('network error'))).toBe(false);
    });

    it('is false while loading (no data, no error yet)', () => {
        expect(deriveFlowMissing(undefined, undefined)).toBe(false);
    });
});
