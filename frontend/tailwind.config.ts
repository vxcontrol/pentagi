import typographyPlugin from '@tailwindcss/typography';
import type { Config } from 'tailwindcss';
import animatePlugin from 'tailwindcss-animate';

/** @type {Config} */
export default {
    darkMode: ['class'],
    content: ['./index.html', './src/**/*.{ts,tsx}'],
    prefix: '',
    theme: {
    	container: {
    		center: true,
    		padding: '2rem',
    		screens: {
    			'2xl': '1400px'
    		}
    	},
    	extend: {
    		fontFamily: {
    			inter: [
    				'Inter',
    				'sans-serif'
    			]
    		},
    		colors: {
    			border: 'hsl(var(--border))',
    			input: 'hsl(var(--input))',
    			ring: 'hsl(var(--ring))',
    			background: 'hsl(var(--background))',
    			foreground: 'hsl(var(--foreground))',
    			primary: {
    				DEFAULT: 'hsl(var(--primary))',
    				foreground: 'hsl(var(--primary-foreground))'
    			},
    			secondary: {
    				DEFAULT: 'hsl(var(--secondary))',
    				foreground: 'hsl(var(--secondary-foreground))'
    			},
    			destructive: {
    				DEFAULT: 'hsl(var(--destructive))',
    				foreground: 'hsl(var(--destructive-foreground))'
    			},
    			muted: {
    				DEFAULT: 'hsl(var(--muted))',
    				foreground: 'hsl(var(--muted-foreground))'
    			},
    			accent: {
    				DEFAULT: 'hsl(var(--accent))',
    				foreground: 'hsl(var(--accent-foreground))'
    			},
    			popover: {
    				DEFAULT: 'hsl(var(--popover))',
    				foreground: 'hsl(var(--popover-foreground))'
    			},
    			card: {
    				DEFAULT: 'hsl(var(--card))',
    				foreground: 'hsl(var(--card-foreground))'
    			},
    			chart: {
    				'1': 'hsl(var(--chart-1))',
    				'2': 'hsl(var(--chart-2))',
    				'3': 'hsl(var(--chart-3))',
    				'4': 'hsl(var(--chart-4))',
    				'5': 'hsl(var(--chart-5))'
    			},
    			sidebar: {
    				DEFAULT: 'hsl(var(--sidebar-background))',
    				foreground: 'hsl(var(--sidebar-foreground))',
    				primary: 'hsl(var(--sidebar-primary))',
    				'primary-foreground': 'hsl(var(--sidebar-primary-foreground))',
    				accent: 'hsl(var(--sidebar-accent))',
    				'accent-foreground': 'hsl(var(--sidebar-accent-foreground))',
    				border: 'hsl(var(--sidebar-border))',
    				ring: 'hsl(var(--sidebar-ring))'
    			}
    		},
    		borderRadius: {
    			lg: 'var(--radius)',
    			md: 'calc(var(--radius) - 2px)',
    			sm: 'calc(var(--radius) - 4px)'
    		},
    		keyframes: {
    			'accordion-down': {
    				from: {
    					height: '0'
    				},
    				to: {
    					height: 'var(--radix-accordion-content-height)'
    				}
    			},
    			'accordion-up': {
    				from: {
    					height: 'var(--radix-accordion-content-height)'
    				},
    				to: {
    					height: '0'
    				}
    			},
    			'caret-blink': {
    				'0%,70%,100%': {
    					opacity: '1'
    				},
    				'20%,50%': {
    					opacity: '0'
    				}
    			},
    			'roll-reveal': {
    				from: {
    					transform: 'rotate(12deg) scale(0)',
    					opacity: '0'
    				},
    				to: {
    					transform: 'rotate(0deg) scale(1)',
    					opacity: '1'
    				}
    			},
    			'slide-left': {
    				from: {
    					transform: 'translateX(20px)',
    					opacity: '0'
    				},
    				to: {
    					transform: 'translateX(0px)',
    					opacity: '1'
    				}
    			},
    			'slide-top': {
    				from: {
    					transform: 'translateY(20px)',
    					opacity: '0'
    				},
    				to: {
    					transform: 'translateY(0px)',
    					opacity: '1'
    				}
    			},
    			'logo-spin': {
    				'0%': {
    					transform: 'rotate(0deg)'
    				},
    				'30%': {
    					transform: 'rotate(360deg)'
    				},
    				'100%': {
    					transform: 'rotate(360deg)'
    				}
    			}
    		},
    		animation: {
    			'accordion-down': 'accordion-down 0.2s ease-out',
    			'accordion-up': 'accordion-up 0.2s ease-out',
    			'caret-blink': 'caret-blink 1.25s ease-out infinite',
    			'roll-reveal': 'roll-reveal 0.4s cubic-bezier(.22,1.28,.54,.99)',
    			'slide-left': 'slide-left 0.3s ease-out',
    			'slide-top': 'slide-top 0.3s ease-out',
    			'logo-spin': 'logo-spin 10s cubic-bezier(0.5, -0.5, 0.5, 1.25) infinite'
    		},
    		transitionDelay: {
    			'10000': '10000ms'
    		}
    	}
    },
    plugins: [animatePlugin, typographyPlugin, require('tailwindcss-animate')],
} satisfies Config;
