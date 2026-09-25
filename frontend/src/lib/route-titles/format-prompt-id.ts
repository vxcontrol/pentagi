export const formatPromptId = (key: string): string => {
    if (!/^[a-z][a-zA-Z]*$/.test(key)) {
        return 'Prompt';
    }

    return key.replaceAll(/([A-Z])/g, ' $1').replace(/^./, (str) => str.toUpperCase());
};
