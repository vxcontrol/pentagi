import { cn } from '@/lib/utils';

interface MistralProps extends React.SVGProps<SVGSVGElement> {
    className?: string;
}

// Placeholder letter mark, not the vendor's logo: swap in the official asset
// when it is available, keeping the title and the 24x24 box.
function Mistral({ className, ...props }: MistralProps) {
    return (
        <svg
            className={cn(className)}
            fill="currentColor"
            fillRule="evenodd"
            viewBox="0 0 24 24"
            {...props}
        >
            <title>Mistral</title>
            <path d="M3 20V4h3.4l5.6 9.3L17.6 4H21v16h-3V9.7l-4.7 7.8h-2.6L6 9.7V20H3z" />
        </svg>
    );
}

export default Mistral;
