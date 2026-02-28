import { cn } from '@/lib/utils';

interface KimiProps extends React.SVGProps<SVGSVGElement> {
    className?: string;
}

const Kimi = ({ className, ...props }: KimiProps) => {
    return (
        <svg
            className={cn(className)}
            fill="currentColor"
            fillRule="evenodd"
            viewBox="0 0 24 24"
            {...props}
        >
            <title>Kimi</title>
            <path
                clipRule="evenodd"
                d="M5 3h3v7.5L14.5 3H19L12 10.5 19.5 21H15L9.5 13 8 14.5V21H5V3z"
            />
        </svg>
    );
};

export default Kimi;
