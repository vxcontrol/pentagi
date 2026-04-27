import { Log } from '@/lib/log';

import type { ResourcePreviewModel } from './types';

const STORAGE_KEY = 'pentagi:resources';

const SEED_RESOURCES: ResourcePreviewModel[] = [
    {
        format: 'pdf',
        id: 'res_seed_1',
        name: 'web-app-pentest-checklist',
        uploadedAt: new Date(Date.now() - 1000 * 60 * 60 * 2).toISOString(),
    },
    {
        format: 'md',
        id: 'res_seed_2',
        name: 'recon-tools-cheatsheet',
        uploadedAt: new Date(Date.now() - 1000 * 60 * 60 * 24 * 3).toISOString(),
    },
    {
        format: 'docx',
        id: 'res_seed_3',
        name: 'incident-response-playbook',
        uploadedAt: new Date(Date.now() - 1000 * 60 * 60 * 24 * 10).toISOString(),
    },
    {
        format: 'csv',
        id: 'res_seed_4',
        name: 'cve-export-2025-q4',
        uploadedAt: new Date(Date.now() - 1000 * 60 * 60 * 24 * 32).toISOString(),
    },
    {
        format: 'txt',
        id: 'res_seed_5',
        name: 'wordlist-rockyou-top-1000',
        uploadedAt: new Date(Date.now() - 1000 * 60 * 60 * 24 * 60).toISOString(),
    },
];

const isBrowser = typeof window !== 'undefined';

const readFromStorage = (): null | ResourcePreviewModel[] => {
    if (!isBrowser) {
        return null;
    }

    try {
        const raw = window.localStorage.getItem(STORAGE_KEY);

        if (!raw) {
            return null;
        }

        const parsed = JSON.parse(raw) as unknown;

        if (!Array.isArray(parsed)) {
            return null;
        }

        return parsed as ResourcePreviewModel[];
    } catch (error) {
        Log.warn('Failed to parse resources from localStorage');
        Log.error('mock-store readFromStorage error:', error);

        return null;
    }
};

const writeToStorage = (items: ResourcePreviewModel[]): void => {
    if (!isBrowser) {
        return;
    }

    try {
        window.localStorage.setItem(STORAGE_KEY, JSON.stringify(items));
    } catch (error) {
        Log.warn('Failed to persist resources to localStorage');
        Log.error('mock-store writeToStorage error:', error);
    }
};

const ensureSeed = (): ResourcePreviewModel[] => {
    const existing = readFromStorage();

    if (existing) {
        return existing;
    }

    writeToStorage(SEED_RESOURCES);

    return [...SEED_RESOURCES];
};

export const mockStore = {
    insert(item: ResourcePreviewModel): ResourcePreviewModel {
        const items = ensureSeed();
        const next = [item, ...items.filter((existing) => existing.id !== item.id)];
        writeToStorage(next);

        return item;
    },
    list(): ResourcePreviewModel[] {
        return ensureSeed();
    },
    removeById(id: string): boolean {
        const items = ensureSeed();
        const next = items.filter((item) => item.id !== id);

        if (next.length === items.length) {
            return false;
        }

        writeToStorage(next);

        return true;
    },
    update(id: string, patch: Partial<ResourcePreviewModel>): null | ResourcePreviewModel {
        const items = ensureSeed();
        const index = items.findIndex((item) => item.id === id);

        if (index === -1) {
            return null;
        }

        const updated: ResourcePreviewModel = { ...items[index]!, ...patch, id };
        const next = [...items];
        next[index] = updated;
        writeToStorage(next);

        return updated;
    },
};
