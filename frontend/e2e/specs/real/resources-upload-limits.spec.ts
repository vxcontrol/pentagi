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
    // The browser-side guard is one DevTools click from gone, so every rule it enforces has to hold
    // when the request is built by hand.
    test('refuses what the client-side guard would have refused', async ({ request }) => {
        const stamp = Date.now();
        const many = Array.from({ length: 1001 }, (_, index) => ({
            body: 'x',
            name: `e2e-many-${stamp}-${index}.txt`,
        }));

        const tooMany = await upload(request, many);

        expect(tooMany.status(), 'one file past the 1000-file cap').toBe(400);

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

        await request.delete('/api/v1/resources/', {
            params: Object.fromEntries(many.slice(0, 1000).map(({ name }) => ['paths[]', name])),
        });
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
