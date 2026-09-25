// Must stay longer than the server's own deadline, so the user sees the
// server's answer rather than an abort of ours on a request still being served.
export const REQUEST_DEADLINE_MS = 45_000;

// The server leaves these unbounded on purpose: each is a model run the operator
// asked for, and taking minutes is its answer. A deadline of ours would cut off
// the very result they are waiting for.
export const UNBOUNDED_OPERATIONS = new Set(['testAgent', 'testProvider']);

export const isUnboundedOperation = (body: unknown): boolean => {
    if (typeof body !== 'string') {
        return false;
    }

    try {
        const { operationName } = JSON.parse(body) as { operationName?: unknown };

        return typeof operationName === 'string' && UNBOUNDED_OPERATIONS.has(operationName);
    } catch {
        return false;
    }
};

export const fetchWithDeadline: typeof fetch = (input, init) => {
    if (isUnboundedOperation(init?.body)) {
        return fetch(input, init);
    }

    const deadline = AbortSignal.timeout(REQUEST_DEADLINE_MS);
    const signal = init?.signal ? AbortSignal.any([init.signal, deadline]) : deadline;

    return fetch(input, { ...init, signal });
};
