import { readFileSync } from 'node:fs';
import { join } from 'node:path';
import { describe, expect, it } from 'vitest';

import { listOrder } from '@/lib/list-order';

const MODELS = join(__dirname, '..', '..', 'backend', 'sqlc', 'models');

const servedBy = {
    agentLogs: ['agentlogs.sql', ['GetFlowAgentLogs']],
    apiTokens: ['api_tokens.sql', ['GetAPITokens', 'GetUserAPITokens']],
    assistantLogs: ['assistantlogs.sql', ['GetFlowAssistantLogs']],
    assistants: ['assistants.sql', ['GetFlowAssistants']],
    flows: ['flows.sql', ['GetFlows', 'GetUserFlows']],
    flowTemplates: ['flow_templates.sql', ['GetFlowTemplatesByUserID']],
    messageLogs: ['msglogs.sql', ['GetFlowMsgLogs']],
    screenshots: ['screenshots.sql', ['GetFlowScreenshots']],
    searchLogs: ['searchlogs.sql', ['GetFlowSearchLogs']],
    tasks: ['tasks.sql', ['GetFlowTasks']],
    terminalLogs: ['termlogs.sql', ['GetFlowTermLogs']],
    vectorStoreLogs: ['vecstorelogs.sql', ['GetFlowVectorStoreLogs']],
} as const satisfies Record<string, readonly [string, readonly string[]]>;

const sqlDirection = { chronological: 'ASC', newestFirst: 'DESC' } as const;

const mirrored = Object.entries(listOrder).filter(([, order]) => order !== 'unmirrored');

const firstSortKey = (body: string): string => {
    const clause = /ORDER BY([^;]*)/i.exec(body)?.[1];

    return clause ? (clause.split(',')[0] ?? '').trim().replace(/\s+/g, ' ').toLowerCase() : 'no order by';
};

const directionIn = (sql: string, query: string): string => {
    const start = sql.indexOf(`-- name: ${query} `);

    if (start === -1) {
        return `no query named ${query}`;
    }

    const next = sql.indexOf('-- name: ', start + 1);
    const [column = '', direction] = firstSortKey(sql.slice(start, next === -1 ? undefined : next)).split(' ');

    if (column.replace(/^\w+\./, '') !== 'created_at') {
        return `${query} does not order by created_at first`;
    }

    return direction === 'desc' ? 'DESC' : 'ASC';
};

const directionsOf = ([file, queries]: readonly [string, readonly string[]]) => {
    const sql = readFileSync(join(MODELS, file), 'utf8');

    return queries.map((query) => ({ found: directionIn(sql, query), query }));
};

// GetTaskPlannedSubtasks feeds PopSubtask, which takes subtasks[0]: its id order is behavioural.
const ID_ORDERED = new Set(['GetTaskCompletedSubtasks', 'GetTaskPlannedSubtasks']);

describe.each(['screenshots.sql', 'subtasks.sql'])('%s serves one direction', (file) => {
    it('sorts every list it serves by created_at ASC first, or by id where the agent needs that', () => {
        const sql = readFileSync(join(MODELS, file), 'utf8');
        const lists = [...sql.matchAll(/-- name: (\w+) :many[^\n]*\n((?:(?!-- name: )[\s\S])*)/g)];

        expect(lists.length, `no :many queries parsed out of ${file}`).toBeGreaterThan(0);

        const wrong = lists
            .map(([, name, body]) => ({ name: name ?? '', sort: firstSortKey(body ?? '') }))
            .filter(({ name, sort }) => !(ID_ORDERED.has(name) ? /\bid asc$/ : /\bcreated_at asc$/).test(sort));

        expect(wrong).toEqual([]);
    });
});

describe('listOrder mirrors the order its queries actually serve', () => {
    it('reads the served query alone, not whichever ORDER BY its file happens to hold', () => {
        const file = [
            '-- name: GetSibling :many',
            'SELECT * FROM t ORDER BY created_at ASC;',
            '-- name: GetServed :many',
            'SELECT * FROM t ORDER BY t.id DESC;',
        ].join('\n');

        expect(directionIn(file, 'GetServed')).toBe('GetServed does not order by created_at first');
        expect(directionIn(file, 'GetSibling')).toBe('ASC');
        expect(directionIn(file, 'GetGone')).toBe('no query named GetGone');
    });

    it('reads the first sort key alone, so a secondary created_at cannot vouch for the query', () => {
        const file = [
            '-- name: GetLeadingId :many',
            'SELECT * FROM t ORDER BY t.id DESC, t.created_at ASC;',
            '-- name: GetAliased :many',
            'SELECT * FROM t ORDER BY t.created_at DESC;',
            '-- name: GetImplicit :many',
            'SELECT * FROM t ORDER BY created_at LIMIT $1;',
        ].join('\n');

        expect(directionIn(file, 'GetLeadingId')).toBe('GetLeadingId does not order by created_at first');
        expect(directionIn(file, 'GetAliased')).toBe('DESC');
        expect(directionIn(file, 'GetImplicit')).toBe('ASC');
    });

    it('names a source query for every mirrored collection, and only those', () => {
        expect(mirrored.map(([field]) => field).sort()).toEqual(Object.keys(servedBy).sort());
    });

    it('declares the direction each collection is really ordered by', () => {
        const disagreements = mirrored.flatMap(([field, order]) =>
            directionsOf(servedBy[field as keyof typeof servedBy])
                .filter(({ found }) => found !== sqlDirection[order as keyof typeof sqlDirection])
                .map(({ found, query }) => ({ declared: order, field, found, query })),
        );

        expect(disagreements).toEqual([]);
    });
});
