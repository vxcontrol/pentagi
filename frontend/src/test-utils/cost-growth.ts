export const GROWTH_IF_QUADRATIC = 16;

// One measurement per size makes a busy machine report quadratic growth on linear code.
export const slowdownWhenInputQuadruples = <TInput>(
    build: (size: number) => TInput,
    scan: (input: TInput) => unknown,
    size: number,
    limit = GROWTH_IF_QUADRATIC / 2,
): number => {
    const elapsedMs = (length: number) => {
        const input = build(length);
        const started = performance.now();

        scan(input);

        return performance.now() - started;
    };

    let smallest = Infinity;

    for (let pair = 0; pair < 7 && smallest >= limit; pair++) {
        smallest = Math.min(smallest, elapsedMs(size) / elapsedMs(size / 4));
    }

    return smallest;
};
