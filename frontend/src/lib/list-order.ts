export type ListOrder = 'chronological' | 'newestFirst' | 'unmirrored';

// Each entry restates that collection's ORDER BY, never how a screen chooses to show it.
export const listOrder = {
    agentLogs: 'chronological',
    apiTokens: 'newestFirst',
    assistantLogs: 'chronological',
    assistants: 'newestFirst',
    flowFiles: 'unmirrored',
    flows: 'newestFirst',
    flowTemplates: 'newestFirst',
    knowledgeDocuments: 'unmirrored',
    messageLogs: 'chronological',
    resources: 'unmirrored',
    screenshots: 'chronological',
    searchLogs: 'chronological',
    tasks: 'chronological',
    terminalLogs: 'chronological',
    vectorStoreLogs: 'chronological',
} as const satisfies Record<string, ListOrder>;

export const orderOf = (field: string): ListOrder => listOrder[field as keyof typeof listOrder] ?? 'unmirrored';

export const placeMissedRows = <T>(field: string, arrived: readonly T[], missed: readonly T[]): T[] =>
    orderOf(field) === 'newestFirst' ? [...missed, ...arrived] : [...arrived, ...missed];
