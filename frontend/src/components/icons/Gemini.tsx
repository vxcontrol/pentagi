import { cn } from '@/lib/utils';

interface GeminiProps extends React.SVGProps<SVGSVGElement> {
    className?: string;
}

const Gemini = ({ className, ...props }: GeminiProps) => {
    return (
        <svg
            fill="none"
            fillRule="evenodd"
            height="1em"
            width="1em"
            viewBox="0 0 16 16"
            className={cn(className)}
            {...props}
        >
            <title>Gemini</title>
            <path
                d="M16 8.016A8.522 8.522 0 008.016 16h-.032A8.521 8.521 0 000 8.016v-.032A8.521 8.521 0 007.984 0h.032A8.522 8.522 0 0016 7.984v.032z"
                fill="url(#prefix__paint0_radial)"
            />
            <defs>
                <radialGradient
                    id="prefix__paint0_radial"
                    cx="0"
                    cy="0"
                    r="1"
                    gradientUnits="userSpaceOnUse"
                    gradientTransform="matrix(16.1326 5.4553 -43.70045 129.2322 1.588 6.503)"
                >
                    <stop offset=".067" stop-color="#9168C0" />
                    <stop offset=".343" stop-color="#5684D1" />
                    <stop offset=".672" stop-color="#1BA1E3" />
                </radialGradient>
            </defs>
        </svg>
    );
};

export default Gemini;
