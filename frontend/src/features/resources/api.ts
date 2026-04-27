import type { ResourcePreviewModel, UploadUrlResponse } from './types';

import { getFileBaseName, getFileExtension } from './constants';
import { mockStore } from './mock-store';

const NETWORK_LATENCY_MIN_MS = 200;
const NETWORK_LATENCY_MAX_MS = 500;
const UPLOAD_TICK_INTERVAL_MS = 120;
const UPLOAD_TICK_PROGRESS_STEP = 10;

const randomDelay = () =>
    new Promise<void>((resolve) => {
        const delay = NETWORK_LATENCY_MIN_MS + Math.random() * (NETWORK_LATENCY_MAX_MS - NETWORK_LATENCY_MIN_MS);
        setTimeout(resolve, delay);
    });

const generateId = () => {
    if (typeof crypto !== 'undefined' && 'randomUUID' in crypto) {
        return `res_${crypto.randomUUID()}`;
    }

    return `res_${Date.now().toString(36)}_${Math.random().toString(36).slice(2, 10)}`;
};

export async function deleteResource(id: string): Promise<void> {
    await randomDelay();
    mockStore.removeById(id);
}

export async function getResources(): Promise<ResourcePreviewModel[]> {
    await randomDelay();

    return mockStore.list();
}

export async function getUploadUrl({
    fileName,
    knowledgeId,
}: {
    fileName: string;
    knowledgeId?: string;
}): Promise<UploadUrlResponse> {
    await randomDelay();

    const id = knowledgeId ?? generateId();
    const item: ResourcePreviewModel = {
        format: getFileExtension(fileName),
        id,
        name: getFileBaseName(fileName),
        uploadedAt: new Date().toISOString(),
    };

    if (knowledgeId) {
        const updated = mockStore.update(knowledgeId, item);

        if (!updated) {
            mockStore.insert(item);
        }
    } else {
        mockStore.insert(item);
    }

    return {
        id,
        item,
        uploadUrl: `mock://upload/${id}`,
    };
}

export async function uploadResource({
    file,
    onProgress,
}: {
    file: File;
    onProgress?: (progress: number) => void;
    uploadUrl: string;
}): Promise<void> {
    if (!file.size) {
        throw new Error('Cannot upload an empty file');
    }

    return new Promise<void>((resolve) => {
        let progress = 0;

        const tick = () => {
            progress = Math.min(100, progress + UPLOAD_TICK_PROGRESS_STEP);
            onProgress?.(progress);

            if (progress >= 100) {
                resolve();

                return;
            }

            setTimeout(tick, UPLOAD_TICK_INTERVAL_MS);
        };

        setTimeout(tick, UPLOAD_TICK_INTERVAL_MS);
    });
}
