import { createContext, type ReactNode, useCallback, useContext, useEffect, useMemo, useState } from 'react';
import { toast } from 'sonner';

import type { ResourceItem, ResourcePreviewModel } from '@/features/resources/types';

import {
    deleteResource as deleteResourceApi,
    getResources,
    getUploadUrl,
    uploadResource as uploadResourceApi,
} from '@/features/resources/api';
import { UPLOAD_COMPLETE_ANIMATION_DELAY } from '@/features/resources/constants';
import { Log } from '@/lib/log';
import { useUser } from '@/providers/user-provider';

interface ResourcesContextValue {
    deleteResource: (id: string) => Promise<void>;
    getResource: (id: string) => ResourceItem | undefined;
    isLoading: boolean;
    resources: ResourceItem[];
    uploadResource: (file: File, options?: UploadResourceOptions) => Promise<ResourceItem>;
}

interface ResourcesProviderProps {
    children: ReactNode;
}

interface UploadResourceOptions {
    fileName?: string;
    knowledgeId?: string;
    onStart?: (resource: ResourceItem) => void;
}

const ResourcesContext = createContext<ResourcesContextValue | undefined>(undefined);

const toResourceItem = (item: ResourcePreviewModel): ResourceItem => ({
    ...item,
    isUploaded: true,
    progress: 100,
});

export const ResourcesProvider = ({ children }: ResourcesProviderProps) => {
    const { authInfo, isAuthenticated } = useUser();

    const shouldFetchResources = Boolean(authInfo && authInfo.type !== 'guest' && isAuthenticated());

    const [resources, setResources] = useState<ResourceItem[]>([]);
    const [isLoading, setIsLoading] = useState(false);
    const [hasLoaded, setHasLoaded] = useState(false);

    useEffect(() => {
        if (!shouldFetchResources || hasLoaded) {
            return;
        }

        let isCancelled = false;
        setIsLoading(true);

        getResources()
            .then((items) => {
                if (isCancelled) {
                    return;
                }

                setResources(items.map(toResourceItem));
                setHasLoaded(true);
            })
            .catch((error: unknown) => {
                if (isCancelled) {
                    return;
                }

                const description = error instanceof Error ? error.message : 'Failed to load resources';
                toast.error('Failed to load resources', {
                    description,
                });
                Log.error('Error loading resources:', error);
            })
            .finally(() => {
                if (!isCancelled) {
                    setIsLoading(false);
                }
            });

        return () => {
            isCancelled = true;
        };
    }, [shouldFetchResources, hasLoaded]);

    const getResource = useCallback(
        (id: string): ResourceItem | undefined => resources.find((item) => item.id === id),
        [resources],
    );

    const upsertResource = useCallback((next: ResourceItem) => {
        setResources((current) => {
            const exists = current.some((item) => item.id === next.id);

            if (exists) {
                return current.map((item) => (item.id === next.id ? next : item));
            }

            return [next, ...current];
        });
    }, []);

    const updateResource = useCallback((id: string, patch: Partial<ResourceItem>) => {
        setResources((current) => current.map((item) => (item.id === id ? { ...item, ...patch } : item)));
    }, []);

    const removeResourceFromState = useCallback((id: string) => {
        setResources((current) => current.filter((item) => item.id !== id));
    }, []);

    const uploadResource = useCallback(
        async (file: File, options?: UploadResourceOptions): Promise<ResourceItem> => {
            try {
                const fileName = options?.fileName ?? file.name;
                const { id, item, uploadUrl } = await getUploadUrl({
                    fileName,
                    knowledgeId: options?.knowledgeId,
                });

                const optimisticItem: ResourceItem = {
                    ...item,
                    isUploaded: false,
                    progress: 0,
                };
                upsertResource(optimisticItem);
                options?.onStart?.(optimisticItem);

                try {
                    await uploadResourceApi({
                        file,
                        onProgress: (progress) => {
                            updateResource(id, { progress });
                        },
                        uploadUrl,
                    });
                } catch (uploadError) {
                    removeResourceFromState(id);
                    throw uploadError;
                }

                const finalItem: ResourceItem = {
                    ...item,
                    isUploaded: false,
                    progress: 100,
                };
                updateResource(id, finalItem);

                setTimeout(() => {
                    updateResource(id, { isUploaded: true, progress: 100 });
                }, UPLOAD_COMPLETE_ANIMATION_DELAY);

                return { ...finalItem, isUploaded: true };
            } catch (error) {
                const description = error instanceof Error ? error.message : 'Failed to upload file';
                toast.error('Failed to upload file', {
                    description,
                });
                Log.error('Error uploading resource:', error);
                throw error;
            }
        },
        [removeResourceFromState, updateResource, upsertResource],
    );

    const deleteResource = useCallback(
        async (id: string) => {
            try {
                await deleteResourceApi(id);
                removeResourceFromState(id);
            } catch (error) {
                const description = error instanceof Error ? error.message : 'Failed to delete resource';
                toast.error('Failed to delete resource', {
                    description,
                });
                Log.error('Error deleting resource:', error);
                throw error;
            }
        },
        [removeResourceFromState],
    );

    const value = useMemo<ResourcesContextValue>(
        () => ({
            deleteResource,
            getResource,
            isLoading,
            resources,
            uploadResource,
        }),
        [deleteResource, getResource, isLoading, resources, uploadResource],
    );

    return <ResourcesContext.Provider value={value}>{children}</ResourcesContext.Provider>;
};

export const useResources = () => {
    const context = useContext(ResourcesContext);

    if (context === undefined) {
        throw new Error('useResources must be used within ResourcesProvider');
    }

    return context;
};
