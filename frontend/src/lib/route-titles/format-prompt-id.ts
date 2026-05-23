/**
 * Converts a camelCase prompt key (e.g. "agentSelector") into a display
 * label ("Agent Selector"). Shared between the prompt detail page header
 * and the route handle title.
 *
 * Returns a generic "Prompt" fallback when the input does not look like
 * a camelCase identifier — e.g. when the URL contains an invalid id such
 * as `/settings/prompts/99999`. Without this guard the route title would
 * render the raw id (`"99999 — PentAGI"`).
 */
export const formatPromptId = (key: string): string => {
    if (!/^[a-z][a-zA-Z]*$/.test(key)) {
        return 'Prompt';
    }

    return key.replaceAll(/([A-Z])/g, ' $1').replace(/^./, (str) => str.toUpperCase());
};
