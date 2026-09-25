import { cn } from '@/lib/utils';

interface XAIProps extends React.SVGProps<SVGSVGElement> {
    className?: string;
}

// Placeholder letter mark, not the vendor's logo: swap in the official asset
// when it is available, keeping the title and the 24x24 box.
function XAI({ className, ...props }: XAIProps) {
    return (
        <svg
            className={cn(className)}
            fill="currentColor"
            fillRule="evenodd"
            viewBox="0 0 24 24"
            {...props}
        >
            <title>xAI</title>
            <path d="M3.2 4h4.1l4.7 6.4L16.7 4h4.1l-6.8 9.2L21 20h-4.1l-4.9-6.7L7.1 20H3l7-9.4L3.2 4z" />
        </svg>
    );
}

export default XAI;
