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

const settingsPrompts: ResultOf<typeof SettingsPromptsDocument> = {
    settingsPrompts: entity('PromptsConfig', {
        default: entity('DefaultPrompts', {
            agents: entity('AgentsPrompts', {
                adviser: agentPromptPair(PromptType.Adviser, PromptType.QuestionAdviser),
                assistant: agentPrompt(PromptType.Assistant),
                coder: agentPromptPair(PromptType.Coder, PromptType.QuestionCoder),
                enricher: agentPromptPair(PromptType.Enricher, PromptType.QuestionEnricher),
                generator: agentPromptPair(PromptType.SubtasksGenerator, PromptType.Generator),
                installer: agentPromptPair(PromptType.Installer, PromptType.QuestionInstaller),
                memorist: agentPromptPair(PromptType.Memorist, PromptType.QuestionMemorist),
                pentester: agentPromptPair(PromptType.Pentester, PromptType.QuestionPentester),
                primaryAgent: agentPrompt(PromptType.PrimaryAgent),
                refiner: agentPromptPair(PromptType.SubtasksRefiner, PromptType.Refiner),
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
};

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
