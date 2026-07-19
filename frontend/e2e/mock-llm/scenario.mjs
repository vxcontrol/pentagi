// Deterministic transcript for one "say hello" flow. Rules are matched (first
// hit wins) against the JSON of the whole request — tool definitions live in
// `tools`, not `messages`, so markers may come from either.
export const RULES = [
    {
        content: 'debian:stable-slim',
        label: 'image-chooser',
        match: /Docker Image Selector/,
    },
    {
        content: 'English',
        label: 'language-chooser',
        match: /Language Detector/,
    },
    {
        content: 'Say Hello Flow',
        label: 'flow-title',
        match: /Flow Title Generator/,
    },
    {
        content: 'Say hello to the user',
        label: 'task-title',
        match: /Task Title Generator/,
    },
    // The pattern detector must FAIL cleanly (plain text, no tool call): the
    // backend then derives the tool-call-ID template from the sampled ids —
    // this rule must sit above get_number since detector prompts quote samples.
    {
        content: 'call_000000',
        label: 'toolcall-pattern-detector',
        match: /submit_pattern/,
    },
    {
        label: 'toolcall-id-sample',
        match: /get_number/,
        toolCalls: [{ args: { value: 1 }, name: 'get_number' }],
    },
    // The task reporter summarizes the finished task and only accepts the
    // report_result tool.
    {
        label: 'task-report',
        match: /report_result|TASK EXECUTION EVALUATOR AND REPORTER/,
        toolCalls: [
            {
                args: {
                    message: 'Task finished successfully',
                    result: 'Hello from the e2e mock LLM!',
                    success: true,
                },
                name: 'report_result',
            },
        ],
    },
    // The refiner re-plans between subtasks and only accepts subtask_patch;
    // an empty operations list means "plan unchanged".
    {
        label: 'subtask-refine',
        match: /subtask_patch/,
        toolCalls: [{ args: { operations: [] }, name: 'subtask_patch' }],
    },
    {
        label: 'subtask-plan',
        match: /subtask_list/,
        toolCalls: [
            {
                args: {
                    message: 'Plan ready',
                    subtasks: [{ description: 'Say hello to the user', title: 'Greet' }],
                },
                name: 'subtask_list',
            },
        ],
    },
];

// Any agent turn not matched above settles its subtask via the `done` barrier
// tool — plain text answers would loop through the reflector instead.
export const FALLBACK = {
    label: 'agent-done',
    toolCalls: [
        {
            args: {
                message: 'Greeted the user',
                result: 'Hello from the e2e mock LLM!',
                success: true,
            },
            name: 'done',
        },
    ],
};
