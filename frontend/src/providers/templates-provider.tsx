import { skipToken, useMutation, useQuery, useSubscription } from '@apollo/client/react';
import { createContext, type ReactNode, useCallback, useContext, useMemo } from 'react';
import { toast } from 'sonner';

import {
    CreateFlowTemplateDocument,
    DeleteFlowTemplateDocument,
    FlowTemplateCreatedDocument,
    FlowTemplateDeletedDocument,
    FlowTemplatesDocument,
    FlowTemplateUpdatedDocument,
    UpdateFlowTemplateDocument,
} from '@/graphql/types';
import { Log } from '@/lib/log';
import { useUser } from '@/providers/user-provider';

export interface Template {
    createdAt: Date;
    id: string;
    text: string;
    title: string;
    updatedAt: Date;
    userId: string;
}

interface TemplatesContextValue {
    createTemplate: (title: string, text: string) => Promise<void>;
    deleteTemplate: (id: string) => Promise<void>;
    getTemplate: (id: string) => Template | undefined;
    isLoading: boolean;
    templates: Template[];
    updateTemplate: (id: string, payload: { text: string; title: string }) => Promise<void>;
}

interface TemplatesProviderProps {
    children: ReactNode;
}

const TemplatesContext = createContext<TemplatesContextValue | undefined>(undefined);

export function TemplatesProvider({ children }: TemplatesProviderProps) {
    const { authInfo, isAuthenticated } = useUser();

    const shouldFetchTemplates = Boolean(authInfo && authInfo.type !== 'guest' && isAuthenticated());

    const { data: templatesData, loading: isLoadingTemplates } = useQuery(
        FlowTemplatesDocument,
        shouldFetchTemplates ? { fetchPolicy: 'cache-and-network' } : skipToken,
    );

    const [createTemplateMutation] = useMutation(CreateFlowTemplateDocument);
    const [updateTemplateMutation] = useMutation(UpdateFlowTemplateDocument);
    const [deleteTemplateMutation] = useMutation(DeleteFlowTemplateDocument);

    useSubscription(FlowTemplateCreatedDocument, {
        skip: !shouldFetchTemplates,
    });

    useSubscription(FlowTemplateUpdatedDocument, {
        skip: !shouldFetchTemplates,
    });

    useSubscription(FlowTemplateDeletedDocument, {
        skip: !shouldFetchTemplates,
    });

    const templates = useMemo(() => {
        const rawTemplates = templatesData?.flowTemplates ?? [];

        return rawTemplates.map((t) => ({
            createdAt: new Date(t.createdAt),
            id: t.id,
            text: t.text,
            title: t.title,
            updatedAt: new Date(t.updatedAt),
            userId: t.userId,
        }));
    }, [templatesData?.flowTemplates]);

    const getTemplate = useCallback(
        (id: string): Template | undefined => {
            return templates.find((t) => t.id === id);
        },
        [templates],
    );

    const createTemplate = useCallback(
        async (title: string, text: string) => {
            try {
                await createTemplateMutation({
                    variables: {
                        input: {
                            text,
                            title,
                        },
                    },
                });
            } catch (error) {
                const errorMessage = error instanceof Error ? error.message : 'Failed to create template';
                toast.error('Failed to create template', {
                    description: errorMessage,
                });
                Log.error('Error creating template:', error);
                throw error;
            }
        },
        [createTemplateMutation],
    );

    const updateTemplate = useCallback(
        async (id: string, payload: { text: string; title: string }) => {
            try {
                await updateTemplateMutation({
                    variables: {
                        input: {
                            text: payload.text,
                            title: payload.title,
                        },
                        templateId: id,
                    },
                });
            } catch (error) {
                const errorMessage = error instanceof Error ? error.message : 'Failed to update template';
                toast.error('Failed to update template', {
                    description: errorMessage,
                });
                Log.error('Error updating template:', error);
                throw error;
            }
        },
        [updateTemplateMutation],
    );

    const deleteTemplate = useCallback(
        async (id: string) => {
            try {
                await deleteTemplateMutation({
                    variables: {
                        templateId: id,
                    },
                });
            } catch (error) {
                const errorMessage = error instanceof Error ? error.message : 'Failed to delete template';
                toast.error('Failed to delete template', {
                    description: errorMessage,
                });
                Log.error('Error deleting template:', error);
                throw error;
            }
        },
        [deleteTemplateMutation],
    );

    const value = useMemo(
        () => ({
            createTemplate,
            deleteTemplate,
            getTemplate,
            isLoading: isLoadingTemplates,
            templates,
            updateTemplate,
        }),
        [createTemplate, deleteTemplate, getTemplate, isLoadingTemplates, templates, updateTemplate],
    );

    return <TemplatesContext.Provider value={value}>{children}</TemplatesContext.Provider>;
}

export function useTemplates() {
    const context = useContext(TemplatesContext);

    if (context === undefined) {
        throw new Error('useTemplates must be used within TemplatesProvider');
    }

    return context;
}
