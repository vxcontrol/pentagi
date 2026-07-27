import { expect, test } from '@playwright/test';

const CONTENT = 'e2e download payload\nsecond line\n';
const ZIP_MAGIC = 'PK';
const END_OF_CENTRAL_DIRECTORY = Buffer.from([0x50, 0x4b, 0x05, 0x06]);
const CENTRAL_DIRECTORY_ENTRY = 0x02014b50;

/** Entry names straight out of the central directory — an empty archive is a valid zip too. */
const zipEntryNames = (archive: Buffer): string[] => {
    const end = archive.lastIndexOf(END_OF_CENTRAL_DIRECTORY);

    if (end === -1) {
        throw new Error(`not a zip archive: ${archive.subarray(0, 16).toString('hex')}`);
    }

    const names: string[] = [];
    let cursor = archive.readUInt32LE(end + 16);

    for (let left = archive.readUInt16LE(end + 10); left > 0; left -= 1) {
        expect(archive.readUInt32LE(cursor), 'central directory entry header').toBe(CENTRAL_DIRECTORY_ENTRY);

        const nameLength = archive.readUInt16LE(cursor + 28);

        names.push(archive.subarray(cursor + 46, cursor + 46 + nameLength).toString());
        cursor += 46 + nameLength + archive.readUInt16LE(cursor + 30) + archive.readUInt16LE(cursor + 32);
    }

    return names;
};

test.describe('resources download at the endpoint', { tag: '@real' }, () => {
    // The browser performs an `<a download>` transfer outside the page context, so the mock tier can
    // only see the URL it asked for. Whether those bytes come back is decided here.
    test('returns the stored bytes, and a zip when several paths are asked for', async ({ request }) => {
        const stamp = Date.now();
        const name = `e2e-download-${stamp}.txt`;
        const second = `e2e-download-${stamp}-b.txt`;

        for (const file of [name, second]) {
            const upload = await request.post('/api/v1/resources/', {
                multipart: { files: { buffer: Buffer.from(CONTENT), mimeType: 'text/plain', name: file } },
            });

            expect(upload.status(), `seed ${file}`).toBe(200);
        }

        try {
            const single = await request.get('/api/v1/resources/download', { params: { 'paths[]': name } });

            expect(single.status()).toBe(200);
            expect(await single.text(), 'the file comes back byte for byte').toBe(CONTENT);

            // Several paths cannot be one file, so the server has to archive them; a plain 200 with the
            // first file's bytes would look like success to any caller that only checks the status.
            const archive = await request.get(`/api/v1/resources/download?paths[]=${name}&paths[]=${second}`);

            expect(archive.status()).toBe(200);

            const body = await archive.body();

            expect(body.subarray(0, 2).toString(), 'a zip, not a bare file').toBe(ZIP_MAGIC);
            expect(zipEntryNames(body).sort(), 'both files are inside it').toEqual([name, second].sort());
        } finally {
            await request.delete(`/api/v1/resources/?paths[]=${name}&paths[]=${second}`);
        }
    });
});
