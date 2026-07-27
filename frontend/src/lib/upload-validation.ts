/**
 * Shared client-side preflight for multipart uploads. Run before constructing
 * the FormData so the user gets an instant toast instead of waiting for the
 * network round-trip and a generic 4xx error.
 *
 * Mirrors the per-file / per-batch limits enforced by the backend
 * (`pkg/resources/resources.go` for the resources library and
 * `pkg/flowfiles/files.go` for the flow files cache). Both backends do not
 * whitelist file extensions and both accept 0-byte files, so neither does this
 * validator — it only checks size and count.
 */

export interface UploadValidationLimits {
    maxFiles: number;
    maxFileSizeMb: number;
    maxTotalSizeMb: number;
}

const MEGABYTE = 1024 * 1024;

/**
 * Validate a batch of `File` objects against the supplied limits. Returns
 * `null` when the batch is acceptable or the user-facing error message for
 * the **first** violation otherwise — callers typically forward this string
 * directly into a `toast.error('Upload failed', { description })`.
 *
 * Empty batches are treated as a no-op (`null`); callers usually short-circuit
 * before reaching the validator anyway.
 */
export const validateUploadBatch = (files: readonly File[], limits: UploadValidationLimits): null | string => {
    if (files.length > limits.maxFiles) {
        return `Too many files: max ${limits.maxFiles} per upload`;
    }

    const maxBytesPerFile = limits.maxFileSizeMb * MEGABYTE;
    const maxTotalBytes = limits.maxTotalSizeMb * MEGABYTE;
    let totalBytes = 0;

    for (const file of files) {
        if (file.size > maxBytesPerFile) {
            return `File "${file.name}" is larger than ${limits.maxFileSizeMb} MB`;
        }

        totalBytes += file.size;
    }

    if (totalBytes > maxTotalBytes) {
        return `Total upload size exceeds the ${limits.maxTotalSizeMb} MB limit`;
    }

    return null;
};
