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

const agentConfig = (model = 'e2e-model'): AgentConfigFragmentFragment =>
    entity('AgentConfig', {
        extraBody: null,
        frequencyPenalty: null,
        maxLength: null,
        maxTokens: null,
        minLength: null,
        model,
        presencePenalty: null,
        price: null,
        reasoning: null,
        repetitionPenalty: null,
        temperature: null,
        topK: null,
        topP: null,
    });

const agentsConfig = (model?: string): AgentsConfigFragmentFragment =>
    entity('AgentsConfig', {
        adviser: agentConfig(model),
        assistant: agentConfig(model),
        coder: agentConfig(model),
        enricher: agentConfig(model),
        generator: agentConfig(model),
        installer: agentConfig(model),
        pentester: agentConfig(model),
        primaryAgent: agentConfig(model),
        refiner: agentConfig(model),
        reflector: agentConfig(model),
        searcher: agentConfig(model),
        simple: agentConfig(model),
        simpleJson: agentConfig(model),
    });

/**
 * Defaults are not database rows: the resolver builds them without an ID, so the wire carries the
 * bare number `0` for every one. `ProviderConfig.keyFields`'s id-0 branch exists for that shape, and
 * a distinct string id here leaves it unexercised.
 */
const defaultConfig = (type: ProviderType): ProviderConfigFragmentFragment =>
    entity('ProviderConfig', {
        // Distinct per type, so two defaults collapsing onto one cache entry is observable.
        agents: agentsConfig(`e2e-${type}-model`),
        createdAt: T,
        id: 0 as unknown as string,
        name: type,
        type,
        updatedAt: T,
    });

const noProviders: ResultOf<typeof SettingsProvidersDocument> = {
    settingsProviders: entity('ProvidersConfig', {
        default: entity('DefaultProvidersConfig', {
            anthropic: defaultConfig(ProviderType.Anthropic),
            bedrock: null,
            custom: null,
            deepseek: null,
            gemini: null,
            glm: null,
            kimi: null,
            minimax: null,
            ollama: null,
            openai: defaultConfig(ProviderType.Openai),
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
            anthropic: defaultConfig(ProviderType.Anthropic),
            bedrock: null,
            custom: null,
            deepseek: null,
            gemini: null,
            glm: null,
            kimi: null,
            minimax: null,
            ollama: null,
            openai: defaultConfig(ProviderType.Openai),
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
