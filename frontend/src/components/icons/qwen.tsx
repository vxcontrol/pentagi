import { cn } from '@/lib/utils';

interface QwenProps extends React.SVGProps<SVGSVGElement> {
    className?: string;
}

const Qwen = ({ className, ...props }: QwenProps) => {
    return (
        <svg
            className={cn(className)}
            fill="currentColor"
            fillRule="evenodd"
            viewBox="0 0 24 24"
            {...props}
        >
            <title>Qwen</title>
            <path
                clipRule="evenodd"
                d="M12 2C6.477 2 2 6.477 2 12s4.477 10 10 10c2.136 0 4.116-.671 5.743-1.813l1.964 1.813 2.121-2.121-1.957-1.807C20.567 16.49 21 14.299 21 12c0-5.523-4.477-10-9-10zm0 3c3.866 0 7 3.134 7 7 0 1.676-.589 3.215-1.565 4.42l-2.142-1.977-2.121 2.121 2.126 1.962A6.963 6.963 0 0112 19a7 7 0 110-14z"
            />
        </svg>
    );
};

export default Qwen;
