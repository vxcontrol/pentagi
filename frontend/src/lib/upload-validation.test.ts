import { describe, expect, it } from 'vitest';

import { type UploadValidationLimits, validateUploadBatch } from './upload-validation';

const MB = 1024 * 1024;

const DEFAULT_LIMITS: UploadValidationLimits = {
    maxFiles: 1000,
    maxFileSizeMb: 300,
    maxTotalSizeMb: 2 * 1024,
};

/**
 * Build a File whose `.size` is forced to `bytes` without actually allocating
 * a buffer of that size. Lets us exercise the multi-GB total-size branch in a
 * unit test without OOMing the worker.
 */
const makeFile = (name: string, bytes: number): File => {
    const file = new File([], name);

    Object.defineProperty(file, 'size', { configurable: true, value: bytes });

    return file;
};

describe('validateUploadBatch', () => {
    it('returns null for an empty batch', () => {
        expect(validateUploadBatch([], DEFAULT_LIMITS)).toBeNull();
    });

    it('returns null when every file is within limits', () => {
        const files = [makeFile('a.txt', 100 * MB), makeFile('b.zip', 250 * MB)];

        expect(validateUploadBatch(files, DEFAULT_LIMITS)).toBeNull();
    });

    // Every rule below is also exercised one unit UNDER its cap: a limit tested only from the
    // rejecting side cannot tell an off-by-one from a correct boundary.
    it('accepts a batch sitting exactly on every cap', () => {
        const limits = { maxFiles: 3, maxFileSizeMb: 10, maxTotalSizeMb: 30 };
        const files = Array.from({ length: 3 }, (_, i) => makeFile(`f-${i}.bin`, 10 * MB));

        expect(validateUploadBatch(files, limits)).toBeNull();
    });

    it('rejects one file past the per-file cap and accepts the file exactly on it', () => {
        expect(validateUploadBatch([makeFile('at.bin', 300 * MB)], DEFAULT_LIMITS)).toBeNull();
        expect(validateUploadBatch([makeFile('over.bin', 300 * MB + 1)], DEFAULT_LIMITS)).toBe(
            'File "over.bin" is larger than 300 MB',
        );
    });

    it('rejects one file past the count cap and accepts the batch exactly on it', () => {
        const limits = { ...DEFAULT_LIMITS, maxFiles: 4 };

        expect(
            validateUploadBatch(
                Array.from({ length: 4 }, (_, i) => makeFile(`f-${i}.txt`, 1)),
                limits,
            ),
        ).toBeNull();
        expect(
            validateUploadBatch(
                Array.from({ length: 5 }, (_, i) => makeFile(`f-${i}.txt`, 1)),
                limits,
            ),
        ).toBe('Too many files: max 4 per upload');
    });

    it('rejects one byte past the total cap and accepts the batch exactly on it', () => {
        const limits = { ...DEFAULT_LIMITS, maxTotalSizeMb: 20 };

        expect(validateUploadBatch([makeFile('a.bin', 10 * MB), makeFile('b.bin', 10 * MB)], limits)).toBeNull();
        expect(validateUploadBatch([makeFile('a.bin', 10 * MB), makeFile('b.bin', 10 * MB + 1)], limits)).toBe(
            'Total upload size exceeds the 20 MB limit',
        );
    });

    it('rejects batches exceeding the file count cap', () => {
        const files = Array.from({ length: 5 }, (_, i) => makeFile(`f-${i}.txt`, 1));

        expect(validateUploadBatch(files, { ...DEFAULT_LIMITS, maxFiles: 4 })).toBe('Too many files: max 4 per upload');
    });

    it('rejects a single file that is larger than the per-file cap', () => {
        const files = [makeFile('ok.txt', 1 * MB), makeFile('huge.bin', 350 * MB)];

        expect(validateUploadBatch(files, DEFAULT_LIMITS)).toBe('File "huge.bin" is larger than 300 MB');
    });

    it('rejects batches whose combined size exceeds the total cap', () => {
        // Each file fits under maxFileSizeMb, but the sum is over the batch limit.
        const files = [makeFile('a.bin', 250 * MB), makeFile('b.bin', 250 * MB), makeFile('c.bin', 250 * MB)];

        expect(validateUploadBatch(files, { ...DEFAULT_LIMITS, maxTotalSizeMb: 600 })).toBe(
            'Total upload size exceeds the 600 MB limit',
        );
    });

    it('accepts 0-byte files alongside other files', () => {
        const files = [makeFile('ok.txt', 1 * MB), makeFile('.gitkeep', 0)];

        expect(validateUploadBatch(files, DEFAULT_LIMITS)).toBeNull();
    });

    it('accepts a batch made only of 0-byte files', () => {
        const files = [makeFile('.gitkeep', 0), makeFile('__init__.py', 0)];

        expect(validateUploadBatch(files, DEFAULT_LIMITS)).toBeNull();
    });

    it('reports the first violation when multiple rules would fail', () => {
        // Both "too many files" AND a too-big file are present — the count
        // check runs first by contract.
        const files = Array.from({ length: 6 }, (_, i) => makeFile(`f-${i}.bin`, 400 * MB));

        expect(validateUploadBatch(files, { ...DEFAULT_LIMITS, maxFiles: 5 })).toBe('Too many files: max 5 per upload');
    });
});
