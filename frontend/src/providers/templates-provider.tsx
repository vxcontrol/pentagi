import {
    createContext,
    type ReactNode,
    useCallback,
    useContext,
    useEffect,
    useMemo,
    useState,
} from 'react';

import { Log } from '@/lib/log';
import { useUser } from '@/providers/user-provider';

export interface Template {
    createdAt: number;
    id: string;
    text: string;
    title: string;
}

interface TemplatesContextValue {
    createTemplate: (title: string, text: string) => string;
    deleteTemplate: (id: string) => void;
    getTemplate: (id: string) => Template | undefined;
    templates: Template[];
    updateTemplate: (id: string, payload: { text: string; title: string }) => void;
}

interface TemplatesProviderProps {
    children: ReactNode;
}

interface TemplatesStorage {
    [userId: string]: Template[];
}

const TemplatesContext = createContext<TemplatesContextValue | undefined>(undefined);

const TEMPLATES_STORAGE_KEY = 'templates';

const loadTemplates = (): TemplatesStorage => {
    try {
        const stored = localStorage.getItem(TEMPLATES_STORAGE_KEY);

        if (stored) {
            const parsed = JSON.parse(stored);

            return typeof parsed === 'object' && parsed !== null ? parsed : {};
        }
    } catch (error) {
        Log.error('Error loading templates from storage:', error);
    }

    return {};
};

const saveTemplates = (storage: TemplatesStorage): void => {
    try {
        localStorage.setItem(TEMPLATES_STORAGE_KEY, JSON.stringify(storage));
    } catch (error) {
        Log.error('Error saving templates to storage:', error);
    }
};

const generateId = (): string => {
    return `${Date.now()}-${Math.random().toString(36).slice(2, 11)}`;
};

export const TemplatesProvider = ({ children }: TemplatesProviderProps) => {
    const { authInfo } = useUser();
    const userId = authInfo?.user?.id?.toString() ?? 'guest';

    const [storage, setStorage] = useState<TemplatesStorage>(() => loadTemplates());

    const templates = useMemo(() => {
        const list = storage[userId] ?? [];

        return [...list].sort((a, b) => b.createdAt - a.createdAt);
    }, [storage, userId]);

    useEffect(() => {
        saveTemplates(storage);
    }, [storage]);

    const getTemplate = useCallback(
        (id: string): Template | undefined => {
            return storage[userId]?.find((t) => t.id === id);
        },
        [storage, userId],
    );

    const createTemplate = useCallback(
        (title: string, text: string): string => {
            const id = generateId();
            const template: Template = {
                createdAt: Date.now(),
                id,
                text,
                title,
            };

            setStorage((previous) => {
                const list = previous[userId] ?? [];

                return {
                    ...previous,
                    [userId]: [...list, template],
                };
            });

            return id;
        },
        [userId],
    );

    const updateTemplate = useCallback(
        (id: string, payload: { text: string; title: string }) => {
            setStorage((previous) => {
                const list = previous[userId] ?? [];
                const index = list.findIndex((t) => t.id === id);

                if (index < 0) {
                    return previous;
                }

                const existing = list[index];

                if (!existing) {
                    return previous;
                }

                const updated = [...list];
                updated[index] = {
                    createdAt: existing.createdAt,
                    id: existing.id,
                    text: payload.text,
                    title: payload.title,
                };

                return {
                    ...previous,
                    [userId]: updated,
                };
            });
        },
        [userId],
    );

    const deleteTemplate = useCallback(
        (id: string) => {
            setStorage((previous) => {
                const list = previous[userId] ?? [];
                const filtered = list.filter((t) => t.id !== id);

                return {
                    ...previous,
                    [userId]: filtered,
                };
            });
        },
        [userId],
    );

    const value = useMemo(
        () => ({
            createTemplate,
            deleteTemplate,
            getTemplate,
            templates,
            updateTemplate,
        }),
        [createTemplate, deleteTemplate, getTemplate, templates, updateTemplate],
    );

    return (
        <TemplatesContext.Provider value={value}>{children}</TemplatesContext.Provider>
    );
};

export const useTemplates = () => {
    const context = useContext(TemplatesContext);

    if (context === undefined) {
        throw new Error('useTemplates must be used within TemplatesProvider');
    }

    return context;
};
