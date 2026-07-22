import type { ResultOf } from '@graphql-typed-document-node/core';

import type { DefaultPromptFragmentFragment, SettingsPromptsDocument } from '@/graphql/types';

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

/**
 * A realistic Go `text/template` for the prompt detail route: the backend parses these with
 * text/template, so the `{{.Var}}` atoms and the fenced command must survive the editor's
 * parse/serialize round-trip. Deliberately has no list-nested fence — that is a separate,
 * already-tracked editor defect and would make this spec assert a known bug.
 */
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
    'nmap -sV {{.Target}}',
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
        userDefined: [],
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
