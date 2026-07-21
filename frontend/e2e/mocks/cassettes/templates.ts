import type { ResultOf } from '@graphql-typed-document-node/core';

import type { FlowTemplateFragmentFragment, FlowTemplatesDocument } from '@/graphql/types';

import type { Cassette } from '../cassette.ts';

import { entity, mergeCassettes } from '../cassette.ts';
import { baseQueries, baseRest } from './base.ts';

const T = '2026-01-15T11:30:00Z';

export const makeTemplate = (id: string, title: string, text: string): FlowTemplateFragmentFragment =>
    entity('FlowTemplate', {
        createdAt: T,
        id,
        text,
        title,
        updatedAt: T,
        userId: '1',
    });

export const TEMPLATE_SEED = makeTemplate('11', 'E2E Seed Template', 'Scan the target and report findings');

const flowTemplates: ResultOf<typeof FlowTemplatesDocument> = { flowTemplates: [TEMPLATE_SEED] };

export const templatesCassette = (override: Cassette = {}): Cassette =>
    mergeCassettes(
        {
            queries: {
                ...baseQueries(),
                flowTemplates: [{ data: flowTemplates }],
            },
            rest: baseRest(),
        },
        override,
    );
