import { readdirSync, readFileSync } from 'node:fs';
import { join, relative } from 'node:path';
import { describe, expect, it } from 'vitest';

const SRC = join(__dirname, '..', 'src');

const sources = (dir = SRC): { file: string; text: string }[] =>
    readdirSync(dir, { withFileTypes: true }).flatMap((entry) => {
        const path = join(dir, entry.name);

        if (entry.isDirectory()) {
            return sources(path);
        }

        if (!/\.tsx?$/.test(entry.name) || /\.test\.tsx?$/.test(entry.name) || entry.name === 'types.ts') {
            return [];
        }

        return [{ file: relative(SRC, path), text: readFileSync(path, 'utf8') }];
    });

describe('no request is sent twice on its own', () => {
    it('wires no retry link into the GraphQL chain', () => {
        const offenders = sources()
            .filter(({ text }) => /RetryLink|@apollo\/client\/link\/retry/.test(text))
            .map(({ file }) => file);

        expect(offenders).toEqual([]);
    });

    it('installs no axios request interceptor', () => {
        const offenders = sources()
            .filter(({ text }) => /interceptors\s*\.\s*request/.test(text))
            .map(({ file }) => file);

        expect(offenders).toEqual([]);
    });

    it('reads the sources it is meant to guard', () => {
        const files = sources().map(({ file }) => file);

        expect(files).toContain(join('lib', 'apollo.ts'));
        expect(files).toContain(join('lib', 'axios.ts'));
    });
});
