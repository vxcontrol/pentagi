import { useCallback, useEffect, useRef, useState } from 'react';

export enum BreakpointName {
    mobile = 'mobile',
    tablet = 'tablet',
    desktop = 'desktop',
}

export const breakpoints = {
    [BreakpointName.mobile]: 768,
    [BreakpointName.tablet]: 1200,
    [BreakpointName.desktop]: Infinity,
} as const;

const breakpointRules: { name: BreakpointName; maxWidth: number }[] = [
    { name: BreakpointName.mobile, maxWidth: breakpoints.mobile },
    { name: BreakpointName.tablet, maxWidth: breakpoints.tablet },
    { name: BreakpointName.desktop, maxWidth: breakpoints.desktop },
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
        isMobile: breakpoint === BreakpointName.mobile,
        isTablet: breakpoint === BreakpointName.tablet,
        isDesktop: breakpoint === BreakpointName.desktop,
    };
};
