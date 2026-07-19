import type { RestResourceEntry, RestResourceList } from '@/features/resources/resources-rest';
import type { UserResourceFragmentFragment } from '@/graphql/types';

import type { Cassette } from '../cassette.ts';

import { entity, mergeCassettes } from '../cassette.ts';
import { baseQueries, baseRest } from './base.ts';

const T = '2026-01-15T09:00:00Z';

const makeRestResource = (id: number, name: string, size: number, isDir: boolean): RestResourceEntry => ({
    created_at: T,
    id,
    is_dir: isDir,
    name,
    path: name,
    size,
    updated_at: T,
    user_id: 1,
});

export const FOLDER_RESOURCE = makeRestResource(1, 'reports', 0, true);
export const FILE_RESOURCE = makeRestResource(2, 'notes.txt', 1234, false);

export const CREATED_FOLDER: UserResourceFragmentFragment = entity('UserResource', {
    createdAt: '2026-01-15T12:00:00Z',
    id: '3',
    isDir: true,
    name: 'new-folder',
    path: 'new-folder',
    size: 0,
    updatedAt: '2026-01-15T12:00:00Z',
    userId: '1',
});

const seededList: RestResourceList = { items: [FOLDER_RESOURCE, FILE_RESOURCE], total: 2 };

export const emptyResourcesCassette = (): Cassette => ({
    queries: baseQueries(),
    rest: baseRest(),
});

export const resourcesCassette = (override: Cassette = {}): Cassette =>
    mergeCassettes(
        {
            queries: baseQueries(),
            rest: {
                ...baseRest(),
                'GET /api/v1/resources/': [{ body: { data: seededList, status: 'success' } }],
            },
        },
        override,
    );
