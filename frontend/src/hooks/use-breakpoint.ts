import { useCallback, useEffect, useRef, useState } from 'react';

export enum BreakpointName {
    desktop = 'desktop',
    mobile = 'mobile',
    tablet = 'tablet',
}

export const breakpoints = {
    [BreakpointName.desktop]: Infinity,
    [BreakpointName.mobile]: 768,
    [BreakpointName.tablet]: 1200,
} as const;

const breakpointRules: { maxWidth: number; name: BreakpointName }[] = [
    { maxWidth: breakpoints.mobile, name: BreakpointName.mobile },
    { maxWidth: breakpoints.tablet, name: BreakpointName.tablet },
    { maxWidth: breakpoints.desktop, name: BreakpointName.desktop },
];

const getBreakpoint = (width: number): BreakpointName => {
    const breakpoint = breakpointRules.find((rule) => width < rule.maxWidth);

    return breakpoint?.name ?? BreakpointName.desktop;
};

export const useBreakpoint = () => {
    const [breakpoint, setBreakpoint] = useState<BreakpointName>(() => {
        if (typeof window === 'undefined') {
            return BreakpointName.desktop;
        }

        return getBreakpoint(window.innerWidth);
    });

    const prevWidthRef = useRef<number>(typeof window !== 'undefined' ? window.innerWidth : 0);
    const breakpointRef = useRef<BreakpointName>(breakpoint);

    // Move state update logic outside of useEffect
    const updateBreakpointState = useCallback((newBreakpoint: BreakpointName) => {
        if (breakpointRef.current !== newBreakpoint) {
            breakpointRef.current = newBreakpoint;
            setBreakpoint(newBreakpoint);
        }
    }, []);

    useEffect(() => {
        if (typeof window === 'undefined') {
            return;
        }

        const handleResize = () => {
            const currentWidth = window.innerWidth;
            const newBreakpoint = getBreakpoint(currentWidth);

            if (currentWidth !== prevWidthRef.current) {
                prevWidthRef.current = currentWidth;
                updateBreakpointState(newBreakpoint);
            }
        };

        window.addEventListener('resize', handleResize);
        handleResize(); // Check on mount

        return () => window.removeEventListener('resize', handleResize);
    }, [updateBreakpointState]);

    return {
        breakpoint,
        isDesktop: breakpoint === BreakpointName.desktop,
        isMobile: breakpoint === BreakpointName.mobile,
        isTablet: breakpoint === BreakpointName.tablet,
    };
};
