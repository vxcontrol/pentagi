export const UPLOAD_COMPLETE_ANIMATION_DELAY = 200;

export const MAX_FILE_SIZE_MB = 5;

export const MAX_FILE_SIZE_MB_BY_EXTENSION: Record<string, number> = {
    ppt: 20,
    pptx: 20,
};

export const ACCEPTED_FILE_EXTENSIONS = [
    'txt',
    'md',
    'html',
    'doc',
    'docx',
    'csv',
    'tsv',
    'xls',
    'xlsx',
    'pdf',
    'ppt',
    'pptx',
];

export const ACCEPTED_FILE_TYPES = ACCEPTED_FILE_EXTENSIONS.map((extension) => `.${extension}`).join(',');

export function getFileBaseName(fileName: string): string {
    const lastDotIndex = fileName.lastIndexOf('.');

    return lastDotIndex === -1 ? fileName : fileName.slice(0, lastDotIndex);
}

export function getFileExtension(fileName: string): string {
    return fileName.split('.').pop()?.toLowerCase() ?? '';
}

export function isFileTypeAccepted(fileName: string): boolean {
    const extension = fileName.split('.').pop()?.toLowerCase();

    if (!extension) {
        return false;
    }

    return ACCEPTED_FILE_EXTENSIONS.includes(extension);
}
