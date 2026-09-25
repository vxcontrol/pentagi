export const getParentDir = (path: string): string => {
    const idx = path.lastIndexOf('/');

    return idx === -1 ? '' : path.slice(0, idx);
};

export const getBaseName = (path: string): string => path.split('/').pop() ?? path;
