// Authoring-time protocol allowlists for the toolbar's Link and Image popovers. They keep javascript:/data:text
// (and, for images, script-capable data:image/svg+xml) out of the persisted document so fidelity never depends on
// every future render path sanitizing them. Load/paste bypass these — tiptap Link's isAllowedUri and the
// read-only viewer sanitize protocols on render; an <img src> is inert regardless.

const SAFE_LINK_PROTOCOLS = new Set(['http:', 'https:', 'mailto:', 'tel:']);

const KNOWN_LINK_SCHEME = /^(?:https?:\/\/|mailto:|tel:)/i;

// Make a scheme-less value absolute so the URL constructor reads it as one, without laundering a root-relative
// path into a bogus host. `//host/x` (protocol-relative) just needs a scheme; a single leading slash is a
// root-relative path this field never supports (`/settings` must NOT become `https://settings`) → reject it.
// A scheme this field does not allow must be REJECTED, not prefixed: `https://` + `ftp://host/file` parses as
// host `ftp`, so the popover showed no error and persisted a dead link. Matched on `scheme://` rather than a
// bare `scheme:` so a scheme-less `localhost:3000` or `example.com:8080/path` still reads as host:port, and so
// the protocol-relative `//cdn…` form below is untouched.
const FOREIGN_SCHEME = /^[a-z][a-z0-9+.-]*:\/\//i;

const toAbsoluteCandidate = (value: string, hasScheme: RegExp): null | string => {
    if (hasScheme.test(value)) {
        return value;
    }

    if (FOREIGN_SCHEME.test(value)) {
        return null;
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

    // `new URL()` accepts an inner space, so a raw candidate reached the document as
    // `[hello](https://example.com/my page)` — which reloads with the link gone and the brackets as text.
    // Rejected rather than percent-encoded: routing the value through `.href` would also rewrite a trailing
    // slash, a `{{VAR}}` placeholder and a non-ASCII path, all of which this editor must keep byte-identical.
    if (!url || /\s/.test(url)) {
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

    if (!src || /\s/.test(src)) {
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
