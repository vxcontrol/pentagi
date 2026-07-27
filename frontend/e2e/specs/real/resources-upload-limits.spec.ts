import type { APIRequestContext } from '@playwright/test';

import { expect, test } from '@playwright/test';

const BOUNDARY = 'e2e-upload-boundary';

/** Builds the multipart body by hand: the point is to send what the client refuses to build. */
const multipart = (files: Array<{ body: string; name: string }>): Buffer => {
    const parts = files
        .map(
            ({ body, name }) =>
                `--${BOUNDARY}\r\nContent-Disposition: form-data; name="files"; filename="${name}"\r\n` +
                `Content-Type: application/octet-stream\r\n\r\n${body}\r\n`,
        )
        .join('');

    return Buffer.from(`${parts}--${BOUNDARY}--\r\n`);
};

const upload = (request: APIRequestContext, files: Array<{ body: string; name: string }>) =>
    request.post('/api/v1/resources/', {
        data: multipart(files),
        headers: { 'content-type': `multipart/form-data; boundary=${BOUNDARY}` },
    });

test.describe('resources upload limits at the endpoint', { tag: '@real' }, () => {
    // The browser-side guard is one DevTools click from gone, so what matters is the set of rules the
    // endpoint itself keeps — a smaller set than the client's, as the empty-file case below records.
    // The size caps (300 MB per file, 2 GB per batch) are left to the backend's own tests: proving
    // them here means pushing gigabytes through the stand.
    test('caps a batch at 1000 files and writes nothing from an over-cap batch', async ({ request }) => {
        const stamp = Date.now();
        const many = Array.from({ length: 1001 }, (_, index) => ({
            body: 'x',
            name: `e2e-many-${stamp}-${index}.txt`,
        }));

        const tooMany = await upload(request, many);

        // 400, but not from the handler's own count check: Go's multipart reader caps a form at 1000
        // parts and answers `multipart: message too large` first (confirmed in the backend log), which
        // makes `resources.MaxUploadFiles` unreachable for this endpoint. What this pins is the
        // user-visible contract — 1001 is refused and nothing is written.
        expect(tooMany.status(), 'a batch past 1000 files is refused').toBe(400);

        const afterReject = await request.get('/api/v1/resources/', { params: { recursive: 'true' } });
        const rejectedPaths = ((await afterReject.json()).data.items ?? []).map((item: { path: string }) => item.path);

        expect(
            rejectedPaths.filter((path: string) => path.startsWith(`e2e-many-${stamp}`)),
            'a rejected batch writes nothing at all, not a partial prefix',
        ).toEqual([]);

        // Exactly on the cap is the other half of the boundary: a server that rejected 1000 as well
        // would look just as "safe" and be wrong.
        const atCap = await upload(request, many.slice(0, 1000));

        expect(atCap.status(), 'a batch exactly on the cap').toBe(200);

        // Repeated keys, not an object: `Object.fromEntries` keeps only the last of 1000 identical
        // `paths[]` entries and would delete one file while leaving 999 on the stand.
        const cleanup = new URLSearchParams();

        many.slice(0, 1000).forEach(({ name }) => cleanup.append('paths[]', name));

        const deleted = await request.delete(`/api/v1/resources/?${cleanup}`);

        expect(deleted.status(), 'the seeded batch is removed again').toBe(200);
    });

    // What licenses the client to send a 0-byte file at all: the endpoint stores it. A client-side
    // refusal here would only have hidden an upload the API accepts.
    test('stores an empty file rather than refusing it', async ({ request }) => {
        const name = `e2e-hollow-${Date.now()}.txt`;
        const accepted = await upload(request, [{ body: '', name }]);

        expect(accepted.status(), 'a 0-byte part is not a server-side error').toBe(200);

        const [stored] = (await accepted.json()).data.items ?? [];

        expect(stored?.path, 'it is created, not silently dropped').toBe(name);
        expect(stored?.size, 'as a zero-byte resource').toBe(0);

        await request.delete('/api/v1/resources/', { params: { 'paths[]': name } });
    });

    test('refuses a batch that carries no file part at all', async ({ request }) => {
        const nothing = await upload(request, []);

        expect(nothing.status(), 'at least one file is required').toBe(400);
    });

    // The server sanitises rather than rejects, which is a legitimate choice — what must hold is that
    // nothing lands outside the library. Asserting a 400 here would pin an implementation the server
    // does not have; asserting the stored path is what actually protects the filesystem.
    test('strips a traversal out of the file name instead of honouring it', async ({ request }) => {
        const name = `e2e-escape-${Date.now()}.txt`;
        const escaped = await upload(request, [{ body: 'owned', name: `../${name}` }]);

        expect(escaped.status(), 'the request is answered, not crashed').toBeLessThan(500);

        const listing = await request.get('/api/v1/resources/', { params: { recursive: 'true' } });
        const paths: string[] = ((await listing.json()).data.items ?? []).map((item: { path: string }) => item.path);

        expect(
            paths.filter((path) => path.includes('..')),
            'no stored path may climb out',
        ).toEqual([]);
        expect(
            paths.filter((path) => path.endsWith(name)),
            'it landed under the sanitised name',
        ).toEqual([name]);

        await request.delete('/api/v1/resources/', { params: { 'paths[]': name } });
    });
});
