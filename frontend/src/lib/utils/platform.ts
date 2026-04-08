export function isMac(): boolean {
    if (typeof navigator === 'undefined') {
        return false;
    }

    const platform =
        'userAgentData' in navigator
            ? ((navigator as { userAgentData?: { platform?: string } }).userAgentData?.platform ?? '')
            : (navigator.platform ?? '');

    return /Mac|iPhone|iPad/i.test(platform) || /Mac/i.test(navigator.userAgent);
}
