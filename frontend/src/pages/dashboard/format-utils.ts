export const formatTokenCount = (count: number): string => {
    if (count >= 1_000_000_000) {
        return `${(count / 1_000_000_000).toFixed(1)}B`;
    }

    if (count >= 1_000_000) {
        return `${(count / 1_000_000).toFixed(1)}M`;
    }

    if (count >= 1_000) {
        return `${(count / 1_000).toFixed(1)}K`;
    }

    return count.toString();
};

export const formatCost = (cost: number): string => {
    if (!cost) {
        return '$0';
    }

    if (cost >= 1) {
        return `$${cost.toFixed(2)}`;
    }

    if (cost >= 0.01) {
        return `$${cost.toFixed(3)}`;
    }

    return `$${cost.toFixed(4)}`;
};

export const formatDuration = (seconds: number): string => {
    if (seconds >= 3600) {
        const hours = Math.floor(seconds / 3600);
        const minutes = Math.floor((seconds % 3600) / 60);

        return `${hours}h ${minutes}m`;
    }

    if (seconds >= 60) {
        const minutes = Math.floor(seconds / 60);
        const remainingSeconds = Math.floor(seconds % 60);

        return `${minutes}m ${remainingSeconds}s`;
    }

    if (seconds >= 1) {
        return `${seconds.toFixed(1)}s`;
    }

    return `${(seconds * 1000).toFixed(0)}ms`;
};

export const formatNumber = (value: number): string => {
    return new Intl.NumberFormat('en-US').format(value);
};
