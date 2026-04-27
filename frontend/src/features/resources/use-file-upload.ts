import { useCallback, useState } from 'react';
import { toast } from 'sonner';

import { useResources } from '@/providers/resources-provider';

import type { ResourceItem } from './types';

import {
    getFileBaseName,
    getFileExtension,
    isFileTypeAccepted,
    MAX_FILE_SIZE_MB,
    MAX_FILE_SIZE_MB_BY_EXTENSION,
} from './constants';

export type UploadOption = 'keep-both' | 'replace';

interface PendingFile {
    existingResource: null | ResourceItem;
    file: File;
}

interface UseFileUploadOptions {
    onResourceAttach?: (resource: ResourceItem) => void;
}

const getNextAvailableFileName = (originalFileName: string, usedNames: Set<string>): string => {
    const lastDotIndex = originalFileName.lastIndexOf('.');
    const hasExtension = lastDotIndex !== -1;

    const baseName = hasExtension ? originalFileName.slice(0, lastDotIndex) : originalFileName;
    const extension = hasExtension ? originalFileName.slice(lastDotIndex) : '';

    let counter = 1;
    let nextName = `${baseName} (${counter})${extension}`;

    while (usedNames.has(nextName)) {
        counter += 1;
        nextName = `${baseName} (${counter})${extension}`;
    }

    return nextName;
};

const buildFullName = (item: { format: string; name: string }) => {
    if (!item.format) {
        return item.name;
    }

    return `${item.name}.${item.format}`;
};

export function useFileUpload(options?: UseFileUploadOptions) {
    const { resources, uploadResource } = useResources();
    const onResourceAttach = options?.onResourceAttach;

    const [pendingFiles, setPendingFiles] = useState<PendingFile[]>([]);
    const [isUploadOptionsDialogOpen, setIsUploadOptionsDialogOpen] = useState(false);

    const processFiles = useCallback(
        async (files: PendingFile[], option: UploadOption) => {
            const usedFileNames = new Set<string>(resources.map(buildFullName));

            const filesToUpload = files.map(({ existingResource, file }) => {
                let fileName: string;
                let knowledgeId: string | undefined;

                if (option === 'replace' && existingResource) {
                    knowledgeId = existingResource.id;
                    fileName = file.name;
                } else if (option === 'keep-both' && usedFileNames.has(file.name)) {
                    fileName = getNextAvailableFileName(file.name, usedFileNames);
                    usedFileNames.add(fileName);
                } else {
                    fileName = file.name;
                    usedFileNames.add(fileName);
                }

                return { file, fileName, knowledgeId };
            });

            await Promise.allSettled(
                filesToUpload.map(({ file, fileName, knowledgeId }) =>
                    uploadResource(file, { fileName, knowledgeId, onStart: onResourceAttach }),
                ),
            );
        },
        [resources, uploadResource, onResourceAttach],
    );

    const handleFileChange = useCallback(
        async (selectedFiles: FileList | null, fileInputRef?: React.RefObject<HTMLInputElement | null>) => {
            const resetInput = () => {
                if (!fileInputRef?.current) {
                    return;
                }

                requestAnimationFrame(() => {
                    if (fileInputRef.current) {
                        fileInputRef.current.value = '';
                    }
                });
            };

            if (!selectedFiles?.length) {
                return;
            }

            const filesToProcess: PendingFile[] = [];

            for (const file of Array.from(selectedFiles)) {
                if (!file.size) {
                    toast.error(`File "${file.name}" is empty`, {
                        description: 'Please upload a file with content.',
                    });

                    continue;
                }

                if (!isFileTypeAccepted(file.name)) {
                    toast.error(`File "${file.name}" has unsupported format`, {
                        description: 'Please upload a supported file type.',
                    });

                    continue;
                }

                const extension = getFileExtension(file.name);
                const maxFileSize = MAX_FILE_SIZE_MB_BY_EXTENSION[extension] ?? MAX_FILE_SIZE_MB;

                if (file.size > maxFileSize * 1024 * 1024) {
                    toast.error(`File "${file.name}" is too large`, {
                        description: `Maximum file size is ${maxFileSize}MB.`,
                    });

                    continue;
                }

                const baseName = getFileBaseName(file.name);
                const existingResource =
                    resources.find((item) => item.name === baseName && item.format === extension) ?? null;

                filesToProcess.push({ existingResource, file });
            }

            if (!filesToProcess.length) {
                resetInput();

                return;
            }

            const hasDuplicates = filesToProcess.some((entry) => entry.existingResource !== null);

            if (hasDuplicates) {
                setPendingFiles(filesToProcess);
                setIsUploadOptionsDialogOpen(true);
            } else {
                await processFiles(filesToProcess, 'keep-both');
            }

            resetInput();
        },
        [processFiles, resources],
    );

    const handleUploadOptionsConfirm = useCallback(
        async (option: UploadOption) => {
            setIsUploadOptionsDialogOpen(false);
            const files = pendingFiles;
            setPendingFiles([]);
            await processFiles(files, option);
        },
        [pendingFiles, processFiles],
    );

    const handleUploadOptionsCancel = useCallback(() => {
        setIsUploadOptionsDialogOpen(false);
        setPendingFiles([]);
    }, []);

    return {
        handleFileChange,
        handleUploadOptionsCancel,
        handleUploadOptionsConfirm,
        isUploadOptionsDialogOpen,
        pendingFiles,
    };
}
