import { cn } from '@/lib/utils';

interface AimlapiProps extends React.SVGProps<SVGSVGElement> {
    className?: string;
}

// Aggregator mark: one hub routing to several upstream vendors. Drawn in
// currentColor like every other provider icon here so it inherits the theme.
function Aimlapi({ className, ...props }: AimlapiProps) {
    return (
        <svg
            className={cn(className)}
            fill="currentColor"
            fillRule="evenodd"
            viewBox="0 0 24 24"
            {...props}
        >
            <title>aimlapi.com</title>
            <path d="M12 9.75a2.25 2.25 0 100 4.5 2.25 2.25 0 000-4.5zM12 8.25c.4 0 .784.07 1.14.198l2.61-3.916a.75.75 0 011.248.832l-2.61 3.916c.63.552 1.05 1.34 1.13 2.22h4.732a.75.75 0 010 1.5h-4.732c-.08.88-.5 1.668-1.13 2.22l2.61 3.916a.75.75 0 01-1.248.832l-2.61-3.916A3.75 3.75 0 0112 15.75c-.4 0-.784-.07-1.14-.198l-2.61 3.916a.75.75 0 01-1.248-.832l2.61-3.916c-.63-.552-1.05-1.34-1.13-2.22H3.75a.75.75 0 010-1.5h4.732c.08-.88.5-1.668 1.13-2.22l-2.61-3.916a.75.75 0 111.248-.832l2.61 3.916c.356-.128.74-.198 1.14-.198z" />
            <path d="M4.5 3a1.5 1.5 0 100 3 1.5 1.5 0 000-3zM19.5 3a1.5 1.5 0 100 3 1.5 1.5 0 000-3zM4.5 18a1.5 1.5 0 100 3 1.5 1.5 0 000-3zM19.5 18a1.5 1.5 0 100 3 1.5 1.5 0 000-3z" />
        </svg>
    );
}

export default Aimlapi;
