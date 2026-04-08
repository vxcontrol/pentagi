import { type DependencyList, type EffectCallback, useEffect, useRef } from 'react';

export function useEffectAfterMount(effect: EffectCallback, dependencies: DependencyList): void {
    const mountedReference = useRef(false);

    useEffect(() => {
        if (!mountedReference.current) {
            mountedReference.current = true;

            return;
        }

        return effect();
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, dependencies);
}
