import { useCallback, useRef, useState } from 'react';
import { toast } from 'sonner';

import { api, getApiErrorMessage, unwrapApiResponse } from '@/lib/axios';

import { FLOW_FILES_API_PATH, UPLOADS_TARGET_DIRECTORY } from './flow-files-constants';
import { type FlowFilesResponse } from './flow-files-utils';

interface UseFlowFilesUploadParams {
    flowId: null | string;
    refetchFiles: () => Promise<unknown>;
}

interface UseFlowFilesUploadResult {
    fileInputKey: number;
    fileInputProps: {
        onChange: (event: React.ChangeEvent<HTMLInputElement>) => void;
        ref: React.RefObject<HTMLInputElement | null>;
    };
    isUploading: boolean;
    openFilePicker: () => void;
    uploadFiles: (selectedFiles: File[]) => Promise<void>;
}

const buildUploadSuccessMessage = (uploadedCount: number, firstFileName?: string) => {
    if (uploadedCount === 1) {
        return {
            description: `Available at ${UPLOADS_TARGET_DIRECTORY}/${firstFileName ?? ''}`,
            title: 'File uploaded',
        };
    }

    return {
        description: `${uploadedCount} files are now available under ${UPLOADS_TARGET_DIRECTORY}`,
        title: `${uploadedCount} files uploaded`,
    };
};

/**
 * Encapsulates the entire upload flow:
 *   * the hidden file input (consumer just spreads `fileInputProps` into the element),
 *   * the imperative `openFilePicker` action,
 *   * the actual `uploadFiles(File[])` call used by both the picker and drag-and-drop.
 *
 * The `key` value is bumped after every upload so React remounts the `<input>` —
 * this clears its native value declaratively without mutating the DOM directly.
 */
export const useFlowFilesUpload = ({ flowId, refetchFiles }: UseFlowFilesUploadParams): UseFlowFilesUploadResult => {
    const inputRef = useRef<HTMLInputElement | null>(null);
    const [isUploading, setIsUploading] = useState(false);
    const [fileInputKey, setFileInputKey] = useState(0);

    const openFilePicker = useCallback(() => {
        inputRef.current?.click();
    }, []);

    const uploadFiles = useCallback(
        async (selectedFiles: File[]) => {
            if (!flowId || selectedFiles.length === 0) {
                return;
            }

            const formData = new FormData();

            selectedFiles.forEach((file) => formData.append('files', file));

            setIsUploading(true);

            try {
                const response = await api.post<FlowFilesResponse, FormData>(FLOW_FILES_API_PATH(flowId), formData, {
                    // Browser sets the multipart boundary automatically when Content-Type is unset.
                    headers: { 'Content-Type': undefined },
                    // Uploads can take longer than the default 30s — disable timeout for this call.
                    timeout: 0,
                });
                const data = unwrapApiResponse(response);
                const uploadedCount = data.files?.length ?? selectedFiles.length;
                const successMessage = buildUploadSuccessMessage(uploadedCount, data.files?.[0]?.name);

                toast.success(successMessage.title, { description: successMessage.description });

                await refetchFiles();
            } catch (error) {
                const description = getApiErrorMessage(error, 'Failed to upload files');

                toast.error('Upload failed', { description });
            } finally {
                setIsUploading(false);
            }
        },
        [flowId, refetchFiles],
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
