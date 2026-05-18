/**
 * Converts a camelCase prompt key (e.g. "agentSelector") into a display
 * label ("Agent Selector"). Used in the prompt detail page and route
 * handle title to keep the two in sync.
 */
export const formatPromptId = (key: string): string =>
    key.replaceAll(/([A-Z])/g, ' $1').replace(/^./, (str) => str.toUpperCase());
