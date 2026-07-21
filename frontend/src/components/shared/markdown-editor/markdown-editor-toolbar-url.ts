// Authoring-time protocol allowlists for the toolbar's Link and Image popovers. They keep javascript:/data:text
// (and, for images, script-capable data:image/svg+xml) out of the persisted document so fidelity never depends on
// every future render path sanitizing them. Load/paste bypass these — tiptap Link's isAllowedUri and the
// read-only viewer sanitize protocols on render; an <img src> is inert regardless.

const SAFE_LINK_PROTOCOLS = new Set(['http:', 'https:', 'mailto:', 'tel:']);

const KNOWN_LINK_SCHEME = /^(?:https?:\/\/|mailto:|tel:)/i;

// Make a scheme-less value absolute so the URL constructor reads it as one, without laundering a root-relative
// path into a bogus host. `//host/x` (protocol-relative) just needs a scheme; a single leading slash is a
// root-relative path this field never supports (`/settings` must NOT become `https://settings`) → reject it.
const toAbsoluteCandidate = (value: string, hasScheme: RegExp): null | string => {
    if (hasScheme.test(value)) {
        return value;
    }

    if (/^\/\/[^/]/.test(value)) {
        return `https:${value}`;
    }

    if (value.startsWith('/')) {
        return null;
    }

    return `https://${value}`;
};

// Returns the normalized href, or null for unsafe/malformed/relative input (javascript:/data:/file:, a
// root-relative /path, empty). Validate with NO base URL — passing window.location as base would launder a
// relative value into https:. Neither tiptap nor its UI normalize manual entry — both persist the raw href
// verbatim — so we do it here.
export const normalizeLinkUrl = (raw: string): null | string => {
    const url = raw.trim();

    if (!url) {
        return null;
    }

    const candidate = toAbsoluteCandidate(url, KNOWN_LINK_SCHEME);

    if (candidate === null) {
        return null;
    }

    try {
        return SAFE_LINK_PROTOCOLS.has(new URL(candidate).protocol) ? candidate : null;
    } catch {
        return null;
    }
};

const RASTER_IMAGE_DATA = /^data:image\/(?:png|jpe?g|gif|webp|bmp);base64,/i;

const KNOWN_IMAGE_SCHEME = /^(?:https?:\/\/|data:)/i;

// Image analog of normalizeLinkUrl. Only http(s) and base64 raster data: URLs survive — data:image/svg+xml is
// rejected because SVG can carry script.
export const normalizeImageSrc = (raw: string): null | string => {
    const src = raw.trim();

    if (!src) {
        return null;
    }

    const candidate = toAbsoluteCandidate(src, KNOWN_IMAGE_SCHEME);

    if (candidate === null) {
        return null;
    }

    if (/^data:/i.test(candidate)) {
        return RASTER_IMAGE_DATA.test(candidate) ? candidate : null;
    }

    try {
        const { protocol } = new URL(candidate);

        return protocol === 'http:' || protocol === 'https:' ? candidate : null;
    } catch {
        return null;
    }
};
