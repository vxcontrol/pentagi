import type { ResultOf } from '@graphql-typed-document-node/core';

import type {
    DefaultPromptFragmentFragment,
    SettingsPromptsDocument,
    UserPromptFragmentFragment,
} from '@/graphql/types';

import { PromptType } from '@/graphql/types';

import type { Cassette } from '../cassette.ts';

import { entity, mergeCassettes } from '../cassette.ts';
import { baseQueries, baseRest } from './base.ts';

export const promptTemplate = (type: PromptType): string => `E2E template for ${type}`;

const prompt = (type: PromptType): DefaultPromptFragmentFragment =>
    entity('DefaultPrompt', { template: promptTemplate(type), type, variables: ['Input'] });

const agentPrompt = (system: PromptType) => entity('AgentPrompt', { system: prompt(system) });

const agentPromptPair = (system: PromptType, human: PromptType) =>
    entity('AgentPrompts', { human: prompt(human), system: prompt(system) });

/** No list-nested fence on purpose: that editor defect is tracked separately, and including it
 * here would make the spec assert a known bug. */
export const RICH_PROMPT_TEMPLATE = [
    '# Pentester',
    '',
    'You are assessing **{{.Target}}** within {{.Scope}}.',
    '',
    '## Rules',
    '',
    '- stay inside the agreed scope',
    '- report every finding with evidence',
    '',
    '```bash',
    'nmap -sV {{.Target}} --script <default>',
    '```',
    '',
    '| Field | Value |',
    '| --- | --- |',
    '| Scope | {{.Scope}} |',
].join('\n');

const richPentesterSystem = entity('DefaultPrompt', {
    template: RICH_PROMPT_TEMPLATE,
    type: PromptType.Pentester,
    variables: ['Target', 'Scope'],
});

const makeSettingsPrompts = (
    pentesterSystem: DefaultPromptFragmentFragment = prompt(PromptType.Pentester),
    userDefined: UserPromptFragmentFragment[] = [],
): ResultOf<typeof SettingsPromptsDocument> => ({
    settingsPrompts: entity('PromptsConfig', {
        default: entity('DefaultPrompts', {
            agents: entity('AgentsPrompts', {
                adviser: agentPromptPair(PromptType.Adviser, PromptType.QuestionAdviser),
                assistant: agentPrompt(PromptType.Assistant),
                coder: agentPromptPair(PromptType.Coder, PromptType.QuestionCoder),
                enricher: agentPromptPair(PromptType.Enricher, PromptType.QuestionEnricher),
                generator: agentPromptPair(PromptType.Generator, PromptType.SubtasksGenerator),
                installer: agentPromptPair(PromptType.Installer, PromptType.QuestionInstaller),
                memorist: agentPromptPair(PromptType.Memorist, PromptType.QuestionMemorist),
                pentester: entity('AgentPrompts', {
                    human: prompt(PromptType.QuestionPentester),
                    system: pentesterSystem,
                }),
                primaryAgent: agentPrompt(PromptType.PrimaryAgent),
                refiner: agentPromptPair(PromptType.Refiner, PromptType.SubtasksRefiner),
                reflector: agentPromptPair(PromptType.Reflector, PromptType.QuestionReflector),
                reporter: agentPromptPair(PromptType.Reporter, PromptType.TaskReporter),
                searcher: agentPromptPair(PromptType.Searcher, PromptType.QuestionSearcher),
                summarizer: agentPrompt(PromptType.Summarizer),
                toolCallFixer: agentPromptPair(PromptType.ToolcallFixer, PromptType.InputToolcallFixer),
            }),
            tools: entity('ToolsPrompts', {
                chooseDockerImage: prompt(PromptType.ImageChooser),
                chooseUserLanguage: prompt(PromptType.LanguageChooser),
                collectToolCallId: prompt(PromptType.ToolCallIdCollector),
                detectToolCallIdPattern: prompt(PromptType.ToolCallIdDetector),
                getExecutionLogs: prompt(PromptType.ExecutionLogs),
                getFlowDescription: prompt(PromptType.FlowDescriptor),
                getFullExecutionContext: prompt(PromptType.FullExecutionContext),
                getShortExecutionContext: prompt(PromptType.ShortExecutionContext),
                getTaskDescription: prompt(PromptType.TaskDescriptor),
                monitorAgentExecution: prompt(PromptType.QuestionExecutionMonitor),
                planAgentTask: prompt(PromptType.QuestionTaskPlanner),
                wrapAgentTask: prompt(PromptType.TaskAssignmentWrapper),
            }),
        }),
        userDefined,
    }),
});

const settingsPrompts = makeSettingsPrompts();

export const settingsPromptsCassette = (override: Cassette = {}): Cassette =>
    mergeCassettes(
        {
            queries: {
                ...baseQueries(),
                settingsPrompts: [{ data: settingsPrompts }],
            },
            rest: baseRest(),
        },
        override,
    );

/** The agent key the prompt detail route resolves from its `:promptId` param. */
export const PROMPT_DETAIL_AGENT = 'pentester';

/** The operator's own copy of the pentester system prompt: its presence is what switches the editor
 *  from the create branch to the update branch, and its id is what a save has to carry. */
export const PROMPT_OVERRIDE = {
    id: 'user-prompt-1',
    // Shares no line with the default on purpose: an override that merely appends to it makes
    // "the default came back" indistinguishable from "nothing happened".
    template: ['# Operator override', '', 'This copy replaces the shipped prompt for {{.Target}}.'].join('\n'),
};

const promptOverride = (): UserPromptFragmentFragment =>
    entity('UserPrompt', {
        createdAt: '2026-01-15T09:00:00Z',
        id: PROMPT_OVERRIDE.id,
        template: PROMPT_OVERRIDE.template,
        type: PromptType.Pentester,
        updatedAt: '2026-01-15T09:00:00Z',
    });

/** Raised by a spec's deletePrompt entry so the refetch can answer without the override. */
export const PROMPT_RESET_FLAG = 'prompt-override-deleted';

/** The same agent, but already overridden — the branch where Save updates instead of creating. The
 *  second answer models the row actually being gone, so a reset has a real post-state to observe. */
export const promptOverrideCassette = (override: Cassette = {}): Cassette =>
    mergeCassettes(
        {
            queries: {
                ...baseQueries(),
                settingsPrompts: [
                    { data: makeSettingsPrompts(richPentesterSystem, [promptOverride()]) },
                    { data: makeSettingsPrompts(richPentesterSystem), whenFlag: PROMPT_RESET_FLAG },
                ],
            },
            rest: baseRest(),
        },
        override,
    );

/** Same config, but that agent's system prompt carries RICH_PROMPT_TEMPLATE. */
export const promptDetailCassette = (override: Cassette = {}): Cassette =>
    mergeCassettes(
        {
            queries: {
                ...baseQueries(),
                settingsPrompts: [{ data: makeSettingsPrompts(richPentesterSystem) }],
            },
            rest: baseRest(),
        },
        override,
    );
