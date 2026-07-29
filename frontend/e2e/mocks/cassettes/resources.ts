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

export const UPLOADED_RESOURCE = makeRestResource(4, 'uploaded-notes.txt', 42, false);

/** What the server broadcasts once the upload lands — the list only grows through this frame. */
export const UPLOADED_FRAGMENT: UserResourceFragmentFragment = entity('UserResource', {
    createdAt: '2026-01-15T12:00:00Z',
    id: '4',
    isDir: false,
    name: UPLOADED_RESOURCE.name,
    path: UPLOADED_RESOURCE.path,
    size: UPLOADED_RESOURCE.size,
    updatedAt: '2026-01-15T12:00:00Z',
    userId: '1',
});

const seededList: RestResourceList = { items: [FOLDER_RESOURCE, FILE_RESOURCE], total: 2 };

export const emptyResourcesCassette = (): Cassette => ({
    queries: baseQueries(),
    rest: baseRest(),
});

export const RENAMED_PATH = 'renamed-notes.txt';
export const COPY_DESTINATION = 'archive/notes.txt';

/** The write verbs differ per action — move is a PUT, copy a POST, delete a DELETE with the paths in
 *  the query string. A method mismatch does not 404 here: it falls through to the SPA and answers 200
 *  with HTML, so each entry pins the method as well as the payload. */
export const resourceWrites = (): Cassette['rest'] => ({
    'DELETE /api/v1/resources/': [
        { body: { data: {}, status: 'success' }, querySubset: { 'paths[]': FILE_RESOURCE.path } },
    ],
    'POST /api/v1/resources/copy': [
        {
            body: { data: {}, status: 'success' },
            bodySubset: { destination: COPY_DESTINATION, sources: [FILE_RESOURCE.path] },
        },
    ],
    'PUT /api/v1/resources/move': [
        {
            body: { data: {}, status: 'success' },
            bodySubset: { destination: RENAMED_PATH, sources: [FILE_RESOURCE.path] },
        },
    ],
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
