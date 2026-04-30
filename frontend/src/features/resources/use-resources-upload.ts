import { useCallback, useRef, useState } from 'react';
import { toast } from 'sonner';

import type { UserResourceFragmentFragment } from '@/graphql/types';

import { api, getApiErrorMessage, unwrapApiResponse } from '@/lib/axios';

import {
    MAX_FILE_SIZE_MB,
    MAX_UPLOAD_FILES_PER_REQUEST,
    MAX_UPLOAD_TOTAL_SIZE_MB,
    RESOURCES_API_PATH,
} from './resources-constants';
import { pluralizeItems } from './resources-utils';

interface UploadOptions {
    /** Virtual directory path inside the user's library. Empty/undefined uploads to root. */
    dir?: string;
}

interface UploadResponse {
    items: UserResourceFragmentFragment[];
    total: number;
}

interface UseResourcesUploadParams {
    onSuccess?: (uploaded: UploadResponse) => void;
}

interface UseResourcesUploadResult {
    fileInputKey: number;
    fileInputProps: {
        onChange: (event: React.ChangeEvent<HTMLInputElement>) => void;
        ref: React.RefObject<HTMLInputElement | null>;
    };
    isUploading: boolean;
    openFilePicker: () => void;
    uploadFiles: (selectedFiles: File[], options?: UploadOptions) => Promise<null | UploadResponse>;
}

const UPLOAD_OVERWRITE_HINT = 'Resource already exists — please rename or remove the existing entry';

/**
 * Mirrors the per-file / per-batch limits enforced by the backend
 * (`pkg/resources/resources.go`). The backend itself does not whitelist file
 * extensions, so neither does the client.
 */
const validateUploadBatch = (files: File[]): null | string => {
    if (files.length > MAX_UPLOAD_FILES_PER_REQUEST) {
        return `Too many files: max ${MAX_UPLOAD_FILES_PER_REQUEST} per upload`;
    }

    const maxBytesPerFile = MAX_FILE_SIZE_MB * 1024 * 1024;
    const maxTotalBytes = MAX_UPLOAD_TOTAL_SIZE_MB * 1024 * 1024;
    let totalBytes = 0;

    for (const file of files) {
        if (file.size === 0) {
            return `File "${file.name}" is empty`;
        }

        if (file.size > maxBytesPerFile) {
            return `File "${file.name}" is larger than ${MAX_FILE_SIZE_MB} MB`;
        }

        totalBytes += file.size;
    }

    if (totalBytes > maxTotalBytes) {
        return `Total upload size exceeds the ${MAX_UPLOAD_TOTAL_SIZE_MB} MB limit`;
    }

    return null;
};

const buildUploadSuccessMessage = (uploadedCount: number, dir?: string) => {
    const target = dir ? `to /${dir}` : 'to your library';

    if (uploadedCount === 1) {
        return { description: `Uploaded ${target}`, title: 'File uploaded' };
    }

    return {
        description: `${uploadedCount} files uploaded ${target}`,
        title: `${uploadedCount} ${pluralizeItems(uploadedCount)} uploaded`,
    };
};

/**
 * Wraps the multipart `POST /resources/` REST call. Bumps `fileInputKey` after
 * each upload so the consumer can remount the hidden `<input>` and clear it
 * declaratively.
 */
export const useResourcesUpload = ({ onSuccess }: UseResourcesUploadParams = {}): UseResourcesUploadResult => {
    const inputRef = useRef<HTMLInputElement | null>(null);
    const [isUploading, setIsUploading] = useState(false);
    const [fileInputKey, setFileInputKey] = useState(0);

    const openFilePicker = useCallback(() => {
        inputRef.current?.click();
    }, []);

    const uploadFiles = useCallback(
        async (selectedFiles: File[], options?: UploadOptions): Promise<null | UploadResponse> => {
            if (selectedFiles.length === 0) {
                return null;
            }

            const validationError = validateUploadBatch(selectedFiles);

            if (validationError) {
                toast.error('Upload failed', { description: validationError });

                return null;
            }

            const formData = new FormData();

            selectedFiles.forEach((file) => formData.append('files', file));

            setIsUploading(true);

            try {
                const response = await api.post<UploadResponse, FormData>(RESOURCES_API_PATH, formData, {
                    headers: { 'Content-Type': undefined },
                    params: options?.dir ? { dir: options.dir } : undefined,
                    timeout: 0,
                });
                const raw = unwrapApiResponse(response);
                // TODO(backend): drop the `String(item.id)` / `String(item.userId)`
                // coercion once `POST /resources/` returns `id`/`userId` as strings.
                // The GraphQL `ID` scalar is typed as `string` everywhere; REST sends
                // numbers, so we normalize here to keep downstream consumers (forms,
                // Apollo cache, etc.) type-consistent.
                const data: UploadResponse = {
                    ...raw,
                    items: (raw.items ?? []).map((item) => ({
                        ...item,
                        id: String(item.id),
                        userId: String(item.userId),
                    })),
                };
                const uploadedCount = data.items.length;
                const message = buildUploadSuccessMessage(uploadedCount, options?.dir);

                toast.success(message.title, { description: message.description });

                onSuccess?.(data);

                return data;
            } catch (error) {
                const description = getApiErrorMessage(error, 'Failed to upload files', {
                    409: UPLOAD_OVERWRITE_HINT,
                });

                toast.error('Upload failed', { description });

                return null;
            } finally {
                setIsUploading(false);
            }
        },
        [onSuccess],
    );

    const handleFileSelection = useCallback(
        async (event: React.ChangeEvent<HTMLInputElement>) => {
            const selectedFiles = Array.from(event.target.files ?? []);

            try {
                await uploadFiles(selectedFiles);
            } finally {
                setFileInputKey((previousKey) => previousKey + 1);
            }
        },
        [uploadFiles],
    );

    return {
        fileInputKey,
        fileInputProps: {
            onChange: handleFileSelection,
            ref: inputRef,
        },
        isUploading,
        openFilePicker,
        uploadFiles,
    };
};
