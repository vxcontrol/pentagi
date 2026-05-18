/**
 * Converts a camelCase prompt key (e.g. "agentSelector") into a display
 * label ("Agent Selector"). Shared between the prompt detail page header
 * and the route handle title.
 */
export const formatPromptId = (key: string): string =>
    key.replaceAll(/([A-Z])/g, ' $1').replace(/^./, (str) => str.toUpperCase());
