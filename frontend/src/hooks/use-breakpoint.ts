import { useEffect, useState } from 'react';

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

    useEffect(() => {
        if (typeof window === 'undefined') {
            return;
        }

        const updateBreakpoint = () => {
            setBreakpoint(getBreakpoint(window.innerWidth));
        };

        window.addEventListener('resize', updateBreakpoint);
        updateBreakpoint();

        return () => window.removeEventListener('resize', updateBreakpoint);
    }, []);

    return {
        breakpoint,
        isMobile: breakpoint === BreakpointName.mobile,
        isTablet: breakpoint === BreakpointName.tablet,
        isDesktop: breakpoint === BreakpointName.desktop,
    };
};
