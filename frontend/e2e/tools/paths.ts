import { existsSync } from 'node:fs';
import { dirname, join, resolve } from 'node:path';

const MARKER = join('backend', 'go.mod');

export const repoRoot = (from: string = process.cwd()): string => {
    let directory = resolve(from);

    while (!existsSync(join(directory, MARKER))) {
        const parent = dirname(directory);

        if (parent === directory) {
            throw new Error(`no repository root above ${from}: nothing contains ${MARKER}`);
        }

        directory = parent;
    }

    return directory;
};

export const repoFile = (...parts: string[]): string => join(repoRoot(), ...parts);
