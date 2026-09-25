import * as React from 'react';

import { cn } from '@/lib/utils';

interface ScrollableRegionProps extends React.ComponentProps<'div'> {
    label: string;
}

export function ScrollableRegion({ children, className, label, ...props }: ScrollableRegionProps) {
    const containerRef = React.useRef<HTMLDivElement>(null);
    const [isScrollable, setIsScrollable] = React.useState(false);

    React.useEffect(() => {
        const container = containerRef.current;

        if (!container) {
            return;
        }

        const measure = () => setIsScrollable(container.scrollWidth > container.clientWidth);

        measure();

        const observer = new ResizeObserver(measure);

        observer.observe(container);

        const content = container.firstElementChild;

        if (content) {
            observer.observe(content);
        }

        return () => observer.disconnect();
    }, []);

    return (
        <div
            aria-label={isScrollable ? label : undefined}
            className={cn(
                'focus-visible:ring-ring relative w-full overflow-auto focus-visible:ring-1 focus-visible:outline-none',
                className,
            )}
            ref={containerRef}
            role={isScrollable ? 'region' : undefined}
            tabIndex={isScrollable ? 0 : undefined}
            {...props}
        >
            {children}
        </div>
    );
}
