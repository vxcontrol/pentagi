import type { ResultOf } from '@graphql-typed-document-node/core';

import type {
    AgentConfigFragmentFragment,
    AgentsConfigFragmentFragment,
    ProviderConfigFragmentFragment,
    SettingsProvidersDocument,
} from '@/graphql/types';

import { ProviderType } from '@/graphql/types';

import type { Cassette } from '../cassette.ts';

import { entity, mergeCassettes } from '../cassette.ts';
import { baseQueries, baseRest } from './base.ts';

const T = '2026-01-15T09:00:00Z';

const agentConfig = (): AgentConfigFragmentFragment =>
    entity('AgentConfig', {
        frequencyPenalty: null,
        maxLength: null,
        maxTokens: null,
        minLength: null,
        model: 'e2e-model',
        presencePenalty: null,
        price: null,
        reasoning: null,
        repetitionPenalty: null,
        temperature: null,
        topK: null,
        topP: null,
    });

const agentsConfig = (): AgentsConfigFragmentFragment =>
    entity('AgentsConfig', {
        adviser: agentConfig(),
        assistant: agentConfig(),
        coder: agentConfig(),
        enricher: agentConfig(),
        generator: agentConfig(),
        installer: agentConfig(),
        pentester: agentConfig(),
        primaryAgent: agentConfig(),
        refiner: agentConfig(),
        reflector: agentConfig(),
        searcher: agentConfig(),
        simple: agentConfig(),
        simpleJson: agentConfig(),
    });

const defaultConfig = (id: string, type: ProviderType): ProviderConfigFragmentFragment =>
    entity('ProviderConfig', {
        agents: agentsConfig(),
        createdAt: T,
        id,
        name: `default-${type}`,
        type,
        updatedAt: T,
    });

const noProviders: ResultOf<typeof SettingsProvidersDocument> = {
    settingsProviders: entity('ProvidersConfig', {
        default: entity('DefaultProvidersConfig', {
            anthropic: defaultConfig('default-anthropic', ProviderType.Anthropic),
            bedrock: null,
            custom: null,
            deepseek: null,
            gemini: null,
            glm: null,
            kimi: null,
            minimax: null,
            ollama: null,
            openai: defaultConfig('default-openai', ProviderType.Openai),
            qwen: null,
        }),
        enabled: entity('ProvidersReadinessStatus', {
            anthropic: false,
            bedrock: false,
            custom: false,
            deepseek: false,
            gemini: false,
            glm: false,
            kimi: false,
            minimax: false,
            ollama: false,
            openai: false,
            qwen: false,
        }),
        models: entity('ProvidersModelsList', {
            anthropic: [],
            bedrock: [],
            custom: [],
            deepseek: [],
            gemini: [],
            glm: [],
            kimi: [],
            minimax: [],
            ollama: [],
            openai: [],
            qwen: [],
        }),
        userDefined: [],
    }),
};

export const settingsProvidersCassette = (override: Cassette = {}): Cassette =>
    mergeCassettes(
        {
            queries: {
                ...baseQueries(),
                settingsProviders: [{ data: noProviders }],
            },
            rest: baseRest(),
        },
        override,
    );

const populatedProviders: ResultOf<typeof SettingsProvidersDocument> = {
    settingsProviders: entity('ProvidersConfig', {
        default: entity('DefaultProvidersConfig', {
            anthropic: defaultConfig('default-anthropic', ProviderType.Anthropic),
            bedrock: null,
            custom: null,
            deepseek: null,
            gemini: null,
            glm: null,
            kimi: null,
            minimax: null,
            ollama: null,
            openai: defaultConfig('default-openai', ProviderType.Openai),
            qwen: null,
        }),
        enabled: entity('ProvidersReadinessStatus', {
            anthropic: true,
            bedrock: false,
            custom: false,
            deepseek: false,
            gemini: false,
            glm: false,
            kimi: false,
            minimax: false,
            ollama: false,
            openai: true,
            qwen: false,
        }),
        models: entity('ProvidersModelsList', {
            anthropic: [],
            bedrock: [],
            custom: [],
            deepseek: [],
            gemini: [],
            glm: [],
            kimi: [],
            minimax: [],
            ollama: [],
            openai: [],
            qwen: [],
        }),
        userDefined: [
            entity('ProviderConfig', {
                agents: agentsConfig(),
                createdAt: T,
                id: 'custom-1',
                name: 'My Custom Endpoint',
                type: ProviderType.Custom,
                updatedAt: T,
            }),
        ],
    }),
};

export const populatedSettingsProvidersCassette = (override: Cassette = {}): Cassette =>
    mergeCassettes(
        {
            queries: {
                ...baseQueries(),
                settingsProviders: [{ data: populatedProviders }],
            },
            rest: baseRest(),
        },
        override,
    );
