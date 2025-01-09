import type * as React from 'react';
import { useEffect } from 'react';

import { useThemeStore } from '@/store/theme-store';

const ThemeProvider = ({ children }: { children: React.ReactNode }) => {
    const theme = useThemeStore((store) => store.theme);

    useEffect(() => {
        document.documentElement.dataset.mode = theme;
        document.documentElement.className = theme;

        localStorage.setItem(
            'theme',
            JSON.stringify({
                state: {
                    theme,
                },
            }),
        );
    }, [theme]);

    return children;
};

export default ThemeProvider;
