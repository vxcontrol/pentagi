import { cn } from '@/lib/utils';

interface GLMProps extends React.SVGProps<SVGSVGElement> {
    className?: string;
}

const GLM = ({ className, ...props }: GLMProps) => {
    return (
        <svg
            className={cn(className)}
            fill="currentColor"
            fillRule="evenodd"
            viewBox="0 0 24 24"
            {...props}
        >
            <title>GLM</title>
            <path
                clipRule="evenodd"
                d="M3 4h18v3.5L8.5 17H21v3H3v-3.5L15.5 7H3V4z"
            />
        </svg>
    );
};

export default GLM;
