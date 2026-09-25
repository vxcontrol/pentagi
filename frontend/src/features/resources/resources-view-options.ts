import { migrateLegacyViewOptions } from '@/lib/view-options-storage';

export type ResourcesViewOptionKey = keyof ResourcesViewOptions;

export interface ResourcesViewOptions {
    isFoldersFirst: boolean;
    isModifiedRelative: boolean;
    isModifiedVisible: boolean;
    isSizeVisible: boolean;
}

export const RESOURCES_PATH = '/resources';

export const defaultResourcesViewOptions: ResourcesViewOptions = {
    isFoldersFirst: true,
    isModifiedRelative: true,
    isModifiedVisible: true,
    isSizeVisible: true,
};

export const seedResourcesViewOptions = (storageKey: string): ResourcesViewOptions => {
    const stored = migrateLegacyViewOptions(RESOURCES_PATH, storageKey);

    return {
        isFoldersFirst: stored.isFoldersFirst ?? stored.foldersFirst ?? defaultResourcesViewOptions.isFoldersFirst,
        isModifiedRelative: stored.isModifiedRelative ?? defaultResourcesViewOptions.isModifiedRelative,
        isModifiedVisible: stored.isModifiedVisible ?? stored.modified ?? defaultResourcesViewOptions.isModifiedVisible,
        isSizeVisible: stored.isSizeVisible ?? stored.size ?? defaultResourcesViewOptions.isSizeVisible,
    };
};
