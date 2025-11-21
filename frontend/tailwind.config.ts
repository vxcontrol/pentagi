import type { Config } from 'tailwindcss';

import typographyPlugin from '@tailwindcss/typography';
import animatePlugin from 'tailwindcss-animate';

/** @type {Config} */
export default {
    content: ['./index.html', './src/**/*.{ts,tsx}'],
    darkMode: ['class'],
    plugins: [animatePlugin, typographyPlugin],
    prefix: '',
    theme: {
        container: {
            center: true,
            padding: '2rem',
            screens: {
                '2xl': '1400px',
            },
        },
        extend: {
            animation: {
                'accordion-down': 'accordion-down 0.2s ease-out',
                'accordion-up': 'accordion-up 0.2s ease-out',
                'caret-blink': 'caret-blink 1.25s ease-out infinite',
                'logo-spin': 'logo-spin 10s cubic-bezier(0.5, -0.5, 0.5, 1.25) infinite',
                'roll-reveal': 'roll-reveal 0.4s cubic-bezier(.22,1.28,.54,.99)',
                'slide-left': 'slide-left 0.3s ease-out',
                'slide-top': 'slide-top 0.3s ease-out',
            },
            borderRadius: {
                lg: 'var(--radius)',
                md: 'calc(var(--radius) - 2px)',
                sm: 'calc(var(--radius) - 4px)',
            },
            colors: {
                accent: {
                    DEFAULT: 'hsl(var(--accent))',
                    foreground: 'hsl(var(--accent-foreground))',
                },
                background: 'hsl(var(--background))',
                border: 'hsl(var(--border))',
                card: {
                    DEFAULT: 'hsl(var(--card))',
                    foreground: 'hsl(var(--card-foreground))',
                },
                chart: {
                    '1': 'hsl(var(--chart-1))',
                    '2': 'hsl(var(--chart-2))',
                    '3': 'hsl(var(--chart-3))',
                    '4': 'hsl(var(--chart-4))',
                    '5': 'hsl(var(--chart-5))',
                },
                destructive: {
                    DEFAULT: 'hsl(var(--destructive))',
                    foreground: 'hsl(var(--destructive-foreground))',
                },
                foreground: 'hsl(var(--foreground))',
                input: 'hsl(var(--input))',
                muted: {
                    DEFAULT: 'hsl(var(--muted))',
                    foreground: 'hsl(var(--muted-foreground))',
                },
                popover: {
                    DEFAULT: 'hsl(var(--popover))',
                    foreground: 'hsl(var(--popover-foreground))',
                },
                primary: {
                    DEFAULT: 'hsl(var(--primary))',
                    foreground: 'hsl(var(--primary-foreground))',
                },
                ring: 'hsl(var(--ring))',
                secondary: {
                    DEFAULT: 'hsl(var(--secondary))',
                    foreground: 'hsl(var(--secondary-foreground))',
                },
                sidebar: {
                    accent: 'hsl(var(--sidebar-accent))',
                    'accent-foreground': 'hsl(var(--sidebar-accent-foreground))',
                    border: 'hsl(var(--sidebar-border))',
                    DEFAULT: 'hsl(var(--sidebar-background))',
                    foreground: 'hsl(var(--sidebar-foreground))',
                    primary: 'hsl(var(--sidebar-primary))',
                    'primary-foreground': 'hsl(var(--sidebar-primary-foreground))',
                    ring: 'hsl(var(--sidebar-ring))',
                },
            },
            fontFamily: {
                inter: ['Inter', 'sans-serif'],
            },
            keyframes: {
                'accordion-down': {
                    from: {
                        height: '0',
                    },
                    to: {
                        height: 'var(--radix-accordion-content-height)',
                    },
                },
                'accordion-up': {
                    from: {
                        height: 'var(--radix-accordion-content-height)',
                    },
                    to: {
                        height: '0',
                    },
                },
                'caret-blink': {
                    '0%,70%,100%': {
                        opacity: '1',
                    },
                    '20%,50%': {
                        opacity: '0',
                    },
                },
                'logo-spin': {
                    '0%': {
                        transform: 'rotate(0deg)',
                    },
                    '30%': {
                        transform: 'rotate(360deg)',
                    },
                    '100%': {
                        transform: 'rotate(360deg)',
                    },
                },
                'roll-reveal': {
                    from: {
                        opacity: '0',
                        transform: 'rotate(12deg) scale(0)',
                    },
                    to: {
                        opacity: '1',
                        transform: 'rotate(0deg) scale(1)',
                    },
                },
                'slide-left': {
                    from: {
                        opacity: '0',
                        transform: 'translateX(20px)',
                    },
                    to: {
                        opacity: '1',
                        transform: 'translateX(0px)',
                    },
                },
                'slide-top': {
                    from: {
                        opacity: '0',
                        transform: 'translateY(20px)',
                    },
                    to: {
                        opacity: '1',
                        transform: 'translateY(0px)',
                    },
                },
            },
            transitionDelay: {
                '10000': '10000ms',
            },
        },
    },
} satisfies Config;
