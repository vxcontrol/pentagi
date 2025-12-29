/**
 * Generates return URL parameter for login redirect
 * @param currentPath - Current pathname to return to after login
 * @returns URL parameter string (empty string or ?returnUrl=...)
 */
export const getReturnUrlParam = (currentPath: string): string => {
    // Don't save default route as return URL
    if (currentPath === '/flows/new' || currentPath === '/login') {
        return '';
    }

    return `?returnUrl=${encodeURIComponent(currentPath)}`;
};
