import { useCallback, useEffect, useRef, useState } from 'react';
import { toast } from 'sonner';

import type { UserResourceFragmentFragment } from '@/graphql/types';

import { api, getApiErrorMessage, unwrapApiResponse } from '@/lib/axios';
import { validateUploadBatch } from '@/lib/upload-validation';

import {
    MAX_FILE_SIZE_MB,
    MAX_UPLOAD_FILES_PER_REQUEST,
    MAX_UPLOAD_TOTAL_SIZE_MB,
    RESOURCES_API_PATH,
} from './resources-constants';
import { restResourceEntryToFragment, type RestResourceList } from './resources-rest';
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
    /**
     * Default virtual directory used when `uploadFiles` is called without an explicit
     * `options.dir`. Drives the toolbar/file-picker / external-DnD flows that don't
     * have a per-call target — empty string keeps uploads at the library root.
     *
     * Stored in a ref internally so changes do NOT invalidate `uploadFiles`,
     * keeping memoized consumers (e.g. `useFilesDragAndDrop`) reference-stable.
     */
    defaultDir?: string;
    onSuccess?: (uploaded: UploadResponse) => void;
}

interface UseResourcesUploadResult {
    fileInputKey: number;
    fileInputProps: {
        onChange: (event: React.ChangeEvent<HTMLInputElement>) => void;
        ref: React.RefObject<HTMLInputElement | null>;
    };
    isUploading: boolean;
    /**
     * Opens the native file picker. The next picked batch uploads to the
     * hook's `defaultDir` (typically the focused folder in the manager, or
     * the library root if nothing is focused). Wired to toolbar buttons,
     * the empty-state CTA, the sidebar quick-upload and external DnD.
     */
    openFilePicker: () => void;
    /**
     * Same picker, but the next picked batch is force-targeted at the
     * supplied directory regardless of `defaultDir`. Drives row-level
     * "Upload here" actions / context-menu items where the user explicitly
     * names the destination folder.
     */
    openFilePickerForDir: (targetDir: string) => void;
    uploadFiles: (selectedFiles: File[], options?: UploadOptions) => Promise<null | UploadResponse>;
}

const UPLOAD_OVERWRITE_HINT = 'Resource already exists — please rename or remove the existing entry';

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
export function useResourcesUpload({ defaultDir, onSuccess }: UseResourcesUploadParams = {}): UseResourcesUploadResult {
    const inputRef = useRef<HTMLInputElement | null>(null);
    const [isUploading, setIsUploading] = useState(false);
    const [fileInputKey, setFileInputKey] = useState(0);

    // Stash the default target directory in a ref so every call to `uploadFiles`
    // sees the latest value without invalidating its memoization. Without this,
    // wrapping consumers (e.g. `useFilesDragAndDrop`) would re-create their
    // handlers on every focus change in the file tree.
    const defaultDirRef = useRef(defaultDir);

    useEffect(() => {
        defaultDirRef.current = defaultDir;
    }, [defaultDir]);

    // One-shot directory override consumed on the next file-picker selection.
    // Lets row-level "Upload here" actions target a specific directory without
    // mutating the hook's default. Both picker entry points reset this ref
    // before opening the dialog, so a stashed override can never leak into a
    // subsequent toolbar / sidebar invocation — even if the user cancels the
    // picker (`change` only fires on actual selection).
    const pendingDirRef = useRef<string | undefined>(undefined);

    const openFilePicker = useCallback(() => {
        pendingDirRef.current = undefined;
        inputRef.current?.click();
    }, []);

    const openFilePickerForDir = useCallback((targetDir: string) => {
        pendingDirRef.current = targetDir;
        inputRef.current?.click();
    }, []);

    const uploadFiles = useCallback(
        async (selectedFiles: File[], options?: UploadOptions): Promise<null | UploadResponse> => {
            if (selectedFiles.length === 0) {
                return null;
            }

            const validationError = validateUploadBatch(selectedFiles, {
                maxFiles: MAX_UPLOAD_FILES_PER_REQUEST,
                maxFileSizeMb: MAX_FILE_SIZE_MB,
                maxTotalSizeMb: MAX_UPLOAD_TOTAL_SIZE_MB,
            });

            if (validationError) {
                toast.error('Upload failed', { description: validationError });

                return null;
            }

            const formData = new FormData();

            selectedFiles.forEach((file) => formData.append('files', file));

            // Per-call `options.dir` wins over the hook-level default so callers
            // can target a specific directory ad-hoc (e.g. row-level "Upload here"
            // actions) without mutating the global default.
            const targetDir = options?.dir ?? defaultDirRef.current;

            setIsUploading(true);

            try {
                const response = await api.post<RestResourceList, FormData>(RESOURCES_API_PATH, formData, {
                    headers: { 'Content-Type': undefined },
                    params: targetDir ? { dir: targetDir } : undefined,
                    timeout: 0,
                });
                // Backend sends `models.ResourceList` (snake_case, numeric IDs).
                // Convert each entry through the canonical bridge so the upload
                // result mirrors what `resources-provider` writes into Apollo on
                // initial hydration and what the `resourceAdded` subscription
                // emits — keeping the cache shape internally consistent.
                const raw = unwrapApiResponse(response);
                const data: UploadResponse = {
                    items: (raw.items ?? []).map(restResourceEntryToFragment),
                    total: raw.total ?? 0,
                };
                const uploadedCount = data.items.length;
                const message = buildUploadSuccessMessage(uploadedCount, targetDir);

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
            // Drain the one-shot override so a subsequent toolbar pick (no arg)
            // resolves through `defaultDirRef` again. Forwarding `undefined`
            // would override `defaultDir` with `undefined`, sending uploads to
            // the root instead of the currently focused folder.
            const pendingDir = pendingDirRef.current;

            pendingDirRef.current = undefined;

            try {
                await uploadFiles(selectedFiles, pendingDir !== undefined ? { dir: pendingDir } : undefined);
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
        openFilePickerForDir,
        uploadFiles,
    };
}
