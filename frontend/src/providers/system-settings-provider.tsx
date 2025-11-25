import type { ReactNode } from 'react';

import { createContext, use } from 'react';

import type { SettingsFragmentFragment } from '@/graphql/types';

import { useSettingsQuery } from '@/graphql/types';
import { useUser } from '@/providers/user-provider';

interface SettingsContextType {
    isLoading: boolean;
    settings: null | SettingsFragmentFragment;
}

const SystemSettingsContext = createContext<SettingsContextType | undefined>(undefined);

export const SystemSettingsProvider = ({ children }: { children: ReactNode }) => {
    const { isAuthenticated } = useUser();

    // Load settings via GraphQL query only when user is authenticated
    const { data: settingsData, loading } = useSettingsQuery({
        skip: !isAuthenticated(),
    });

    return (
        <SystemSettingsContext
            value={{
                isLoading: loading,
                settings: settingsData?.settings ?? null,
            }}
        >
            {children}
        </SystemSettingsContext>
    );
};

export const useSystemSettings = () => {
    const context = use(SystemSettingsContext);

    if (context === undefined) {
        throw new Error('useSystemSettings must be used within a SystemSettingsProvider');
    }

    return context;
};
