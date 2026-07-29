import type { ResultOf } from '@graphql-typed-document-node/core';

import type {
    AgentConfigFragmentFragment,
    AgentsConfigFragmentFragment,
    AgentTestResultFragmentFragment,
    ModelConfigFragmentFragment,
    ProviderConfigFragmentFragment,
    ProviderTestResultFragmentFragment,
    SettingsProvidersDocument,
    TestResultFragmentFragment,
} from '@/graphql/types';

import { ProviderType } from '@/graphql/types';

import type { Cassette } from '../cassette.ts';

import { entity, mergeCassettes } from '../cassette.ts';
import { baseQueries, baseRest } from './base.ts';

const T = '2026-01-15T09:00:00Z';

/** `n: 1` and `json: true` are the shipped simple_json defaults; they are pinned in the create payload
 *  because dropping exactly these on the way to the wire is a regression this repo has already had. */
const agentConfig = (model = 'e2e-model'): AgentConfigFragmentFragment =>
    entity('AgentConfig', {
        extraBody: null,
        frequencyPenalty: null,
        json: true,
        maxLength: null,
        maxTokens: 4096,
        minLength: null,
        minP: null,
        model,
        n: 1,
        presencePenalty: null,
        price: null,
        reasoning: null,
        repetitionPenalty: null,
        responseMimeType: null,
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

const modelConfig = (name: string): ModelConfigFragmentFragment =>
    entity('ModelConfig', { name, price: null, reasoning: null, thinking: null });

// The backend fills a default config AND a model catalog for EVERY type, keyed or not (verified live:
// gemini/deepseek/ollama all return a non-null default and a populated catalog with zero keys). A
// cassette leaving them null/[] encodes a wire state the resolver never emits and leaves the
// create-form's model-dropdown seeding path structurally unreachable.
const allDefaults = () =>
    entity('DefaultProvidersConfig', {
        anthropic: defaultConfig(ProviderType.Anthropic),
        bedrock: defaultConfig(ProviderType.Bedrock),
        custom: defaultConfig(ProviderType.Custom),
        deepseek: defaultConfig(ProviderType.Deepseek),
        gemini: defaultConfig(ProviderType.Gemini),
        glm: defaultConfig(ProviderType.Glm),
        kimi: defaultConfig(ProviderType.Kimi),
        minimax: defaultConfig(ProviderType.Minimax),
        ollama: defaultConfig(ProviderType.Ollama),
        openai: defaultConfig(ProviderType.Openai),
        qwen: defaultConfig(ProviderType.Qwen),
    });

const catalog = (type: ProviderType) => [modelConfig(`e2e-${type}-model`), modelConfig(`e2e-${type}-model-mini`)];

const allModels = () =>
    entity('ProvidersModelsList', {
        anthropic: catalog(ProviderType.Anthropic),
        bedrock: catalog(ProviderType.Bedrock),
        custom: catalog(ProviderType.Custom),
        deepseek: catalog(ProviderType.Deepseek),
        gemini: catalog(ProviderType.Gemini),
        glm: catalog(ProviderType.Glm),
        kimi: catalog(ProviderType.Kimi),
        minimax: catalog(ProviderType.Minimax),
        ollama: catalog(ProviderType.Ollama),
        openai: catalog(ProviderType.Openai),
        qwen: catalog(ProviderType.Qwen),
    });

const noProviders: ResultOf<typeof SettingsProvidersDocument> = {
    settingsProviders: entity('ProvidersConfig', {
        default: allDefaults(),
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
        models: allModels(),
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

export const SEEDED_PROVIDER = {
    id: 'custom-1',
    name: 'My Custom Endpoint',
    type: ProviderType.Custom,
};

export const seededProviderRow = (): ProviderConfigFragmentFragment =>
    entity('ProviderConfig', {
        agents: agentsConfig(),
        createdAt: T,
        id: SEEDED_PROVIDER.id,
        name: SEEDED_PROVIDER.name,
        type: SEEDED_PROVIDER.type,
        updatedAt: T,
    });

/** The whole settingsProviders answer with a chosen set of user-defined rows — a delete asserts against
 *  the follow-up answer, so the list has to be expressible without the deleted row. */
export const providersList = (...userDefined: ProviderConfigFragmentFragment[]) => ({
    settingsProviders: entity('ProvidersConfig', {
        default: allDefaults(),
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
        models: allModels(),
        userDefined,
    }),
});

/** A second row so a delete spec can operate a row that is not index 0, and prove the table survived. */
export const OTHER_PROVIDER = {
    id: 'custom-2',
    name: 'Second Endpoint',
};

export const OTHER_PROVIDER_ROW = (): ProviderConfigFragmentFragment =>
    entity('ProviderConfig', {
        agents: agentsConfig(),
        createdAt: T,
        id: OTHER_PROVIDER.id,
        name: OTHER_PROVIDER.name,
        type: ProviderType.Custom,
        updatedAt: T,
    });

/** updateProvider answers with ProviderConfig, and Apollo writes it straight onto the row the list
 *  renders — returning the pre-edit name would make the list contradict the save. */
export const renamedProviderRow = (name: string): ProviderConfigFragmentFragment =>
    entity('ProviderConfig', {
        agents: agentsConfig(),
        createdAt: T,
        id: SEEDED_PROVIDER.id,
        name,
        type: SEEDED_PROVIDER.type,
        updatedAt: T,
    });

const populatedProviders: ResultOf<typeof SettingsProvidersDocument> = providersList(seededProviderRow());

const testResult = (name: string, result: boolean): TestResultFragmentFragment =>
    entity('TestResult', {
        error: result ? null : 'e2e simulated failure',
        latency: 42,
        name,
        reasoning: false,
        result,
        streaming: false,
        type: 'chat',
    });

export const agentTestResult = (result = true): AgentTestResultFragmentFragment =>
    entity('AgentTestResult', { tests: [testResult('chat completion', result)] });

/** One agent fails on purpose: a dialog that renders only the passing rows would still look right. */
export const providerTestResult = (): ProviderTestResultFragmentFragment =>
    entity('ProviderTestResult', {
        adviser: agentTestResult(),
        assistant: agentTestResult(),
        coder: agentTestResult(),
        enricher: agentTestResult(),
        generator: agentTestResult(),
        installer: agentTestResult(),
        pentester: agentTestResult(false),
        primaryAgent: agentTestResult(),
        refiner: agentTestResult(),
        reflector: agentTestResult(),
        searcher: agentTestResult(),
        simple: agentTestResult(),
        simpleJson: agentTestResult(),
    });

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
