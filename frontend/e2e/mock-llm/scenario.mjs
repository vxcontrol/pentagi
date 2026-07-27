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
    // An empty operations list means "plan unchanged".
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
    // The transcript's one real sandbox exec — the only Tier-2 leg that
    // exercises exec → terminal-log → WS streaming into the UI. The primary
    // agent has no terminal tool (only delegate tools like `pentester`), so
    // the chain is: primary delegates → the subagent (whose toolset has
    // `terminal` + its `hack_result` barrier) runs the command → reports with
    // the marker → the marker in each caller's transcript settles it. Order
    // matters: each rule below fires only after the ones above stopped
    // matching that agent's turn.
    {
        label: 'subagent-terminal-report',
        match: /^(?=[\s\S]*"hack_result")(?=[\s\S]*E2E_TERMINAL_OK)/,
        toolCalls: [
            {
                args: {
                    message: 'Terminal check passed, greeting delivered',
                    result: 'Terminal check E2E_TERMINAL_OK passed; greeted the user: Hello from the e2e mock LLM!',
                },
                name: 'hack_result',
            },
        ],
    },
    {
        label: 'agent-terminal-done',
        match: /E2E_TERMINAL_OK/,
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
    },
    {
        label: 'agent-terminal',
        match: /"terminal"/,
        toolCalls: [
            {
                args: {
                    cwd: '/',
                    detach: false,
                    input: 'uname -a && echo E2E_TERMINAL_OK',
                    message: 'Checking the sandbox kernel',
                    timeout: 0,
                },
                name: 'terminal',
            },
        ],
    },
    {
        label: 'agent-delegate',
        match: /"pentester"/,
        toolCalls: [
            {
                args: {
                    message: 'Delegating the greeting task',
                    question: 'Say hello to the user from the sandbox.',
                },
                name: 'pentester',
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
