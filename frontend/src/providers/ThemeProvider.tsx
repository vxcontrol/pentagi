import { ThemeProvider as NextThemesProvider } from 'next-themes';
import { useEffect, type ReactNode } from 'react';

const ThemeProvider = ({ children }: { children: ReactNode }) => {
    useEffect(() => {
        const updateDataMode = () => {
            const hasDark = document.documentElement.classList.contains('dark');
            const hasLight = document.documentElement.classList.contains('light');

            if (!hasDark && !hasLight) {
                document.documentElement.classList.add('dark');
            }

            const isDark = document.documentElement.classList.contains('dark');
            const theme = isDark ? 'dark' : 'light';
            document.documentElement.dataset.mode = theme;
        };

        const observer = new MutationObserver(updateDataMode);

        observer.observe(document.documentElement, {
            attributes: true,
            attributeFilter: ['class'],
        });

        updateDataMode();

        return () => observer.disconnect();
    }, []);

    return (
        <NextThemesProvider
            attribute="class"
            defaultTheme="dark"
            enableSystem={false}
            disableTransitionOnChange
            storageKey="theme"
        >
            {children}
        </NextThemesProvider>
    );
};

export { useTheme } from 'next-themes';
export default ThemeProvider;
