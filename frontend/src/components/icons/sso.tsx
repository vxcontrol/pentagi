import { cn } from '@/lib/utils';

interface SsoProps extends React.SVGProps<SVGSVGElement> {
    className?: string;
}

function Sso({ className, ...props }: SsoProps) {
    return (
        <svg
            className={cn(className)}
            fill="none"
            stroke="currentColor"
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth="2"
            viewBox="0 0 24 24"
            xmlns="http://www.w3.org/2000/svg"
            {...props}
        >
            <path d="M12 2 4 5.5v6c0 5 3.4 9.2 8 10.5 4.6-1.3 8-5.5 8-10.5v-6z" />
            <path d="M15 11a3 3 0 1 0-4.6 2.5V17h2.1v-2h1.6v-1.5h-1.6v-.6A3 3 0 0 0 15 11" />
        </svg>
    );
}

export default Sso;
