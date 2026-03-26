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
            <path d="M12.105 2L9.927 4.953H.653L2.83 2h9.276zM23.254 19.048L21.078 22h-9.242l2.174-2.952h9.244zM24 2L9.264 22H0L14.736 2H24z" />
        </svg>
    );
};

export default GLM;
