import { describe, expect, it } from 'vitest';

import { formatPromptId } from './format-prompt-id';

describe('formatPromptId', () => {
    it.each([
        ['assistant', 'Assistant'],
        ['primaryAgent', 'Primary Agent'],
        ['toolCallFixer', 'Tool Call Fixer'],
        ['getShortExecutionContext', 'Get Short Execution Context'],
        ['detectToolCallIdPattern', 'Detect Tool Call Id Pattern'],
    ])('turns the prompt key %s into %s', (key, expected) => {
        expect(formatPromptId(key)).toBe(expected);
    });

    it.each(['99999', '__typename', 'Assistant', 'primary_agent', ''])(
        'falls back to a generic label for %j, which is not a prompt key',
        (key) => {
            expect(formatPromptId(key)).toBe('Prompt');
        },
    );
});
