import { cn } from '@/lib/utils';

interface DeepSeekProps extends React.SVGProps<SVGSVGElement> {
    className?: string;
}

const DeepSeek = ({ className, ...props }: DeepSeekProps) => {
    return (
        <svg
            className={cn(className)}
            fill="currentColor"
            fillRule="evenodd"
            viewBox="0 0 24 24"
            {...props}
        >
            <title>DeepSeek</title>
            <path
                clipRule="evenodd"
                d="M4 3h8c5.523 0 9 3.477 9 9s-3.477 9-9 9H4V3zm3 3v12h5c3.866 0 6-2.686 6-6s-2.134-6-6-6H7z"
            />
        </svg>
    );
};

export default DeepSeek;
