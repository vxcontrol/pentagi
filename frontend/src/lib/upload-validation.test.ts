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

    it('rejects 0-byte files by default', () => {
        const files = [makeFile('ok.txt', 1 * MB), makeFile('empty.txt', 0)];

        expect(validateUploadBatch(files, DEFAULT_LIMITS)).toBe('File "empty.txt" is empty');
    });

    it('lets 0-byte files through when rejectEmpty is disabled', () => {
        const files = [makeFile('ok.txt', 1 * MB), makeFile('empty.txt', 0)];

        expect(validateUploadBatch(files, { ...DEFAULT_LIMITS, rejectEmpty: false })).toBeNull();
    });

    it('reports the first violation when multiple rules would fail', () => {
        // Both "too many files" AND a too-big file are present — the count
        // check runs first by contract.
        const files = Array.from({ length: 6 }, (_, i) => makeFile(`f-${i}.bin`, 400 * MB));

        expect(validateUploadBatch(files, { ...DEFAULT_LIMITS, maxFiles: 5 })).toBe('Too many files: max 5 per upload');
    });
});
