import { cn } from '@/lib/utils';

interface GeminiProps extends React.SVGProps<SVGSVGElement> {
    className?: string;
}

const Gemini = ({ className, ...props }: GeminiProps) => {
    return (
        <svg
            fill="currentColor"
            fillRule="evenodd"
            viewBox="0 0 16 16"
            className={cn(className)}
            {...props}
        >
            <title>Gemini</title>
            <path d="M16 8.016A8.522 8.522 0 008.016 16h-.032A8.521 8.521 0 000 8.016v-.032A8.521 8.521 0 007.984 0h.032A8.522 8.522 0 0016 7.984v.032z" />
        </svg>
    );
};

export default Gemini;
