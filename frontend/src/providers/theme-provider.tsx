import { createContext, useEffect, useState } from 'react';

const themes = ['dark', 'light', 'system'] as const;

export type Theme = (typeof themes)[number];

const isThemeValid = (value: unknown): value is Theme => typeof value === 'string' && themes.includes(value as Theme);

interface ThemeProviderProps {
    children: React.ReactNode;
    defaultTheme?: Theme;
    storageKey?: string;
}

interface ThemeProviderState {
    setTheme: (theme: Theme) => void;
    theme: Theme;
}

const initialState: ThemeProviderState = {
    setTheme: () => null,
    theme: 'system',
};

export const ThemeProviderContext = createContext<ThemeProviderState>(initialState);

export const ThemeProvider = ({
    children,
    defaultTheme = 'dark',
    storageKey = 'theme',
    ...props
}: ThemeProviderProps) => {
    const [theme, setTheme] = useState<Theme>(() => {
        const storedTheme = localStorage.getItem(storageKey);

        return isThemeValid(storedTheme) ? storedTheme : defaultTheme;
    });

    useEffect(() => {
        const root = window.document.documentElement;

        root.classList.remove('light', 'dark');

        if (theme === 'system') {
            const systemTheme = window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light';

            root.classList.add(systemTheme);

            return;
        }

        root.classList.add(theme);
    }, [theme]);

    const value = {
        setTheme: (theme: Theme) => {
            localStorage.setItem(storageKey, theme);
            setTheme(theme);
        },
        theme,
    };

    return (
        <ThemeProviderContext.Provider
            {...props}
            value={value}
        >
            {children}
        </ThemeProviderContext.Provider>
    );
};
